package mqtt

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/dratasich/thingsboard-go-client-sdk/events"
	"github.com/eclipse/paho.golang/paho"
	"github.com/rs/zerolog/log"
)

// Provisioning MQTT client
//
// Client connecting to MQTT broker to provision (= get credentials for) a device.
//
// Because it uses different topics we only need when we provision a device,
// we create a separate client for this (to avoid subscribing to provision all the time).
type TBMQTTProvisioningClient struct {
	// embedded mqtt client
	*TBMQTT

	// Provisioning credentials
	provisionDeviceKey    string
	provisionDeviceSecret string

	// Queue for device provisioning responses
	ResponseQueue chan *events.DeviceProvisioningResponse
}

const (
	// --- device provisioning topics ---
	provisioningRequestTopic  = "/provision/request"
	provisioningResponseTopic = "/provision/response"
)

// Create a new MQTT client for device provisioning
func NewProvisioningClient(cfg Config) *TBMQTTProvisioningClient {
	// Provision topics are not secured with a proper username/password
	// the provisioning key/secret are sent in the payload and checked by TB
	cfg.Username = "provision"

	client := &TBMQTTProvisioningClient{
		TBMQTT:                NewClient(cfg),
		provisionDeviceKey:    cfg.ProvisioningKey,
		provisionDeviceSecret: cfg.ProvisioningSecret,
		ResponseQueue:         make(chan *events.DeviceProvisioningResponse, 1),
	}
	return client
}

// Connect to the MQTT broker and subscribe to necessary topics
func (client *TBMQTTProvisioningClient) Connect(ctx context.Context) {
	// extend subscriptions with gateway topics
	subscriptions := []paho.SubscribeOptions{
		// listen to RPC commands for devices
		{
			Topic: provisioningResponseTopic,
			QoS:   qos,
		},
	}

	client.connect(ctx, subscriptions, client.handler)
}

// Handle received messages from all the subscribed topics
func (client *TBMQTTProvisioningClient) handler(msg *paho.Publish) {
	log.Debug().Msgf("Provisioning client received message on topic %s", msg.Topic)

	// device provisioning response
	if msg.Topic == provisioningResponseTopic {
		log.Debug().Msgf("Received device provisioning response: %s", string(msg.Payload))
		// parse payload
		var response events.DeviceProvisioningResponse
		if err := json.Unmarshal([]byte(msg.Payload), &response); err != nil {
			log.Error().Msgf("Message could not be parsed: %s. Payload: %s", err, msg.Payload)
		} else {
			// check if response is OK
			if !IsProvisionResponseOK(&response) {
				log.Error().Msgf("Device provisioning failed (%s): %s", response.Status, response.ErrorMsg)
			} else {
				log.Info().Msgf("Device provisioning successful, got device ID %s", response.DeviceId)
			}
			// push to a queue (also on error so the caller can check and not await a timeout)
			client.ResponseQueue <- &response
		}
		return
	}
}

// Provision a device
func (client *TBMQTTProvisioningClient) Provision(deviceName string) {
	msg := events.DeviceProvisioningRequest{
		DeviceName:            deviceName,
		ProvisionDeviceKey:    client.provisionDeviceKey,
		ProvisionDeviceSecret: client.provisionDeviceSecret,
	}
	payload, _ := json.Marshal(msg)
	client.publishRaw(provisioningRequestTopic, payload)

	log.Info().Msgf("Published device provisioning request for %s", deviceName)
	log.Debug().Msgf("Provisioning device request payload: %s", payload)
}

// Provision a device with access token
func (client *TBMQTTProvisioningClient) ProvisionWithAccessToken(deviceName string, accessToken string) {
	credentialsType := events.CredentialsTypeAccessToken
	msg := events.DeviceProvisioningRequest{
		DeviceName:            deviceName,
		ProvisionDeviceKey:    client.provisionDeviceKey,
		ProvisionDeviceSecret: client.provisionDeviceSecret,
		CredentialsType:       &credentialsType,
		Token:                 &accessToken,
	}
	payload, _ := json.Marshal(msg)
	client.publishRaw(provisioningRequestTopic, payload)

	log.Info().Msgf("Published device provisioning request for %s", deviceName)
	log.Debug().Msgf("Provisioning device request payload: %s", payload)
}

// Provision a device with X.509 certificate
func (client *TBMQTTProvisioningClient) ProvisionWithCertificate(deviceName string, certificatePem string) {
	credentialsType := events.CredentialsTypeX509
	hash := PemToHash(certificatePem)
	msg := events.DeviceProvisioningRequest{
		DeviceName:            deviceName,
		ProvisionDeviceKey:    client.provisionDeviceKey,
		ProvisionDeviceSecret: client.provisionDeviceSecret,
		CredentialsType:       &credentialsType,
		CertificateHash:       &hash,
	}
	payload, _ := json.Marshal(msg)
	client.publishRaw(provisioningRequestTopic, payload)

	log.Info().Msgf("Published device provisioning request for %s", deviceName)
	log.Debug().Msgf("Provisioning device request payload: %s", payload)
}

// Returns true if the provision response indicates success
func IsProvisionResponseOK(resp *events.DeviceProvisioningResponse) bool {
	return resp.ErrorMsg == "" &&
		resp.Status == events.ProvisionStatusSuccess &&
		resp.CredentialsValue != "" &&
		(resp.CredentialsType == events.CredentialsTypeAccessToken ||
			resp.CredentialsType == events.CredentialsTypeX509)
}

// Converts provision response to PEM format for output
func PemFromHash(hash string) string {
	// newline after each 64 chars according to PEM format
	// https://datatracker.ietf.org/doc/html/rfc7468#page-8
	var withNewlines strings.Builder
	for i := 0; i < len(hash); i += 64 {
		end := min(i+64, len(hash))
		withNewlines.WriteString(hash[i:end] + "\n")
	}
	hash = withNewlines.String()
	// add PEM header/footer and return
	return "-----BEGIN CERTIFICATE-----\n" + hash + "-----END CERTIFICATE-----"
}

// Removes PEM formatting from certificate
func PemToHash(pem string) string {
	hash := strings.ReplaceAll(pem, "-----BEGIN CERTIFICATE-----", "")
	hash = strings.ReplaceAll(hash, "-----END CERTIFICATE-----", "")
	hash = strings.ReplaceAll(hash, "\n", "")
	return hash
}
