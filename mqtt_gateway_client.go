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
	// embedded mqtt client
	*TBMQTT

	connectedDevices *datastructures.Set[string]

	// counter for attribute request ids
	gatewayAttributeRequestCounter int32

	// queues of received events from TB
	GatewayAttributesQueue         chan *events.Attributes
	GatewayAttributesResponseQueue chan *events.ResponseAttributes
	GatewayRpcQueue                chan *events.RequestRPC
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

// Create a new gateway MQTT client
func NewGatewayClient(cfg Config) *TBMQTTGatewayClient {
	gateway := &TBMQTTGatewayClient{
		TBMQTT:                         NewClient(cfg),
		connectedDevices:               datastructures.NewSet[string](),
		gatewayAttributeRequestCounter: 0,
		GatewayAttributesQueue:         make(chan *events.Attributes, 10),
		GatewayAttributesResponseQueue: make(chan *events.ResponseAttributes, 10),
		GatewayRpcQueue:                make(chan *events.RequestRPC, 100),
	}
	return gateway
}

func (gateway *TBMQTTGatewayClient) Connect(ctx context.Context) {
	// get base subscriptions of the embedded client
	deviceSubs := gateway.subscriptions()
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
		gateway.handler(msg)
	}

	gateway.connect(ctx, subscriptions, handler)
}

// Disconnect all devices and the gateway itself
func (gateway *TBMQTTGatewayClient) Disconnect(ctx context.Context) {
	// disconnect all devices
	for device := range gateway.connectedDevices.Iterator() {
		gateway.DisconnectDevice(device)
	}

	// disconnect gateway itself
	gateway.TBMQTT.Disconnect(ctx)
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
	gateway.publishRaw(gatewayConnectTopic, payload)

	gateway.connectedDevices.Add(deviceName)
	log.Info().Msgf("Connected device: %s", deviceName)
}

// Disconnect a device
func (gateway *TBMQTTGatewayClient) DisconnectDevice(deviceName string) {
	msg := events.Disconnect{
		Device: deviceName,
	}
	payload, _ := json.Marshal(msg)
	gateway.publishRaw(gatewayDisconnectTopic, payload)

	gateway.connectedDevices.Remove(deviceName)
	log.Info().Msgf("Disconnected device: %s", deviceName)
}

// Send single device telemetry
func (gateway *TBMQTTGatewayClient) SendTelemetry(deviceName string, telemetry events.Telemetry) {
	msg := events.TelemetryBatch{
		deviceName: []events.Telemetry{telemetry},
	}
	payload, _ := json.Marshal(msg)
	gateway.publishRaw(gatewayTelemetryTopic, payload)
	log.Info().Msgf("Published telemetry data: %s", payload)
}

// Send telemetry batch
func (gateway *TBMQTTGatewayClient) SendTelemetryBatch(batch events.TelemetryBatch) {
	payload, _ := json.Marshal(batch)
	gateway.publishRaw(gatewayTelemetryTopic, payload)
	log.Info().Msgf("Published telemetry data batch: %s", payload)
}
