package mqtt

import (
	"context"
	"encoding/json"

	"github.com/dratasich/thingsboard-go-client-sdk/datastructures"
	"github.com/dratasich/thingsboard-go-client-sdk/events"
	"github.com/eclipse/paho.golang/paho"
	"github.com/rs/zerolog/log"
)

type TBMQTTGatewayClient struct {
	// mqtt client
	client           *TBMQTT
	connectedDevices *datastructures.Set[string]

	// counter for attribute request ids
	attributeRequestCounter int32

	// queues of received events from TB
	AttributesQueue         chan *events.Attributes
	AttributesResponseQueue chan *events.ResponseAttributes
	RpcQueue                chan *events.RequestRPC
}

const (
	// --- gateway topics ---

	gatewayConnectTopic    = "v1/gateway/connect"
	gatewayDisconnectTopic = "v1/gateway/disconnect"

	gatewayAttributesTopic         = "v1/gateway/attributes"
	gatewayAttributesRequestTopic  = "v1/gateway/attributes/request"
	gatewayAttributesResponseTopic = "v1/gateway/attributes/response"

	gatewayRpcRequestTopic = "v1/gateway/rpc"

	gatewayTelemetryTopic = "v1/gateway/telemetry"
)

func NewGatewayClient(cfg Config) *TBMQTTGatewayClient {
	gateway := &TBMQTTGatewayClient{
		client:                  NewClient(cfg),
		connectedDevices:        datastructures.NewSet[string](),
		attributeRequestCounter: 0,
		AttributesQueue:         make(chan *events.Attributes, 10),
		AttributesResponseQueue: make(chan *events.ResponseAttributes, 10),
		RpcQueue:                make(chan *events.RequestRPC, 100),
	}
	return gateway
}

func (gateway *TBMQTTGatewayClient) Connect(ctx context.Context) {
	deviceSubs := gateway.client.subscriptions()
	// extend subscriptions with gateway topics
	gatewaySubs := []paho.SubscribeOptions{
		// listen to attribute updates
		{
			Topic: gatewayAttributesTopic,
			QoS:   qos,
		},
		// listen to attribute responses
		{
			Topic: gatewayAttributesResponseTopic,
			QoS:   qos,
		},
		// listen to RPC commands for devices
		{
			Topic: gatewayRpcRequestTopic,
			QoS:   qos,
		},
	}
	subscriptions := append(deviceSubs, gatewaySubs...)

	handler := func(msg *paho.Publish) {
		// handle messages for the gateway itself
		gateway.client.handler(msg)
	}

	gateway.client.connect(ctx, subscriptions, handler)
}

// Wait for MQTT connection is up
func (gateway *TBMQTTGatewayClient) AwaitConnection() {
	gateway.client.AwaitConnection()
}

// Disconnect all devices and the gateway itself
func (gateway *TBMQTTGatewayClient) Disconnect(ctx context.Context) {
	// disconnect all devices
	for device := range gateway.connectedDevices.Iterator() {
		gateway.DisconnectDevice(device)
	}

	// disconnect gateway
	gateway.client.Disconnect(ctx)
	log.Info().Msg("Gateway disconnected")
}

// Connect a device
//
// A new device will be created if it does not yet exist and device provisioning is enabled.
// Attributes updates and new RPCs will be sent to the gateway.
func (gateway *TBMQTTGatewayClient) ConnectDevice(deviceName string, deviceProfile string) {
	msg := events.Connect{
		Device: deviceName,
		Type:   deviceProfile,
	}
	payload, _ := json.Marshal(msg)
	gateway.client.publishRaw(gatewayConnectTopic, payload)

	gateway.connectedDevices.Add(deviceName)
	log.Info().Msgf("Connected device: %s", deviceName)
}

// Disconnect a device
func (gateway *TBMQTTGatewayClient) DisconnectDevice(deviceName string) {
	msg := events.Disconnect{
		Device: deviceName,
	}
	payload, _ := json.Marshal(msg)
	gateway.client.publishRaw(gatewayDisconnectTopic, payload)

	gateway.connectedDevices.Remove(deviceName)
	log.Info().Msgf("Disconnected device: %s", deviceName)
}

// Send single device telemetry
func (gateway *TBMQTTGatewayClient) SendTelemetry(deviceName string, telemetry events.Telemetry) {
	msg := events.TelemetryBatch{
		deviceName: []events.Telemetry{telemetry},
	}
	payload, _ := json.Marshal(msg)
	gateway.client.publishRaw(gatewayTelemetryTopic, payload)
}

// Send telemetry batch
func (gateway *TBMQTTGatewayClient) SendTelemetryBatch(batch events.TelemetryBatch) {
	payload, _ := json.Marshal(batch)
	gateway.client.publishRaw(gatewayTelemetryTopic, payload)
}
