package mqtt

import (
	"context"
	"encoding/json"
	"strings"

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
	GatewayAttributesQueue         chan *events.GatewayAttributes
	GatewayAttributesResponseQueue chan *events.ResponseAttributes
	GatewayRpcQueue                chan *events.GatewayRequestRPC
}

const (
	// --- gateway topics ---

	gatewayConnectTopic    = "v1/gateway/connect"
	gatewayDisconnectTopic = "v1/gateway/disconnect"

	gatewayAttributesTopic         = "v1/gateway/attributes"
	gatewayAttributesRequestTopic  = "v1/gateway/attributes/request"
	gatewayAttributesResponseTopic = "v1/gateway/attributes/response"

	// send and receive RPCs for devices
	gatewayRpcTopic = "v1/gateway/rpc"

	gatewayTelemetryTopic = "v1/gateway/telemetry"
)

// Create a new gateway MQTT client
func NewGatewayClient(cfg Config) *TBMQTTGatewayClient {
	gateway := &TBMQTTGatewayClient{
		TBMQTT:                         NewClient(cfg),
		connectedDevices:               datastructures.NewSet[string](),
		gatewayAttributeRequestCounter: 0,
		GatewayAttributesQueue:         make(chan *events.GatewayAttributes, 10),
		GatewayAttributesResponseQueue: make(chan *events.ResponseAttributes, 10),
		GatewayRpcQueue:                make(chan *events.GatewayRequestRPC, 100),
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
			Topic: gatewayRpcTopic,
			QoS:   qos,
		},
	}
	subscriptions := append(deviceSubs, gatewaySubs...)

	gateway.connect(ctx, subscriptions, gateway.handler)
}

// Handle received messages from all the subscribed topics
func (gateway *TBMQTTGatewayClient) handler(msg *paho.Publish) {
	log.Debug().Msgf("Gateway received message on topic %s", msg.Topic)

	// handle messages for the gateway itself
	gateway.TBMQTT.handler(msg)

	// attribute updates
	if msg.Topic == gatewayAttributesTopic {
		log.Info().Msg("Gateway received attribute updates")
		var attrs events.GatewayAttributes
		err := json.Unmarshal(msg.Payload, &attrs)
		if err != nil {
			log.Error().Msgf("Failed to unmarshal attributes: %s", err)
			return
		}
		log.Debug().Msgf("Pushing attributes to queue: %s", attrs)
		gateway.GatewayAttributesQueue <- &attrs
		return
	}
	// attribute response
	if id, found := strings.CutPrefix(msg.Topic, gatewayAttributesResponseTopic); found {
		log.Info().Msgf("Attribute response received with id #%s", id)
		var attrs = events.ResponseAttributes{
			Id: id,
		}
		err := json.Unmarshal(msg.Payload, &attrs)
		if err != nil {
			log.Error().Msgf("Failed to unmarshal attribute response: %s. Payload: %s", err, msg.Payload)
			return
		}
		log.Debug().Msgf("Pushing attribute response to queue: %s", id)
		gateway.GatewayAttributesResponseQueue <- &attrs
		return
	}
	// RPCs
	if msg.Topic == gatewayRpcTopic {
		log.Info().Msg("Received RPC request")
		// parse payload
		var rpc events.GatewayRequestRPC
		if err := json.Unmarshal([]byte(msg.Payload), &rpc); err != nil {
			log.Error().Msgf("Message could not be parsed: %s. Payload: %s", err, msg.Payload)
		} else {
			// push to a queue
			log.Debug().Msgf("Pushing RPC request #%d for device %s to queue: %+v", rpc.Data.RpcRequestId, rpc.Device, rpc)
			gateway.GatewayRpcQueue <- &rpc
		}
		return
	}
	log.Warn().Msgf("No handler for topic %s", msg.Topic)
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

// Reply to an RPC request for a device
func (gateway *TBMQTTGatewayClient) ReplyDeviceRPC(msg events.GatewayResponseRPC) {
	rpc_json, _ := json.Marshal(msg)

	gateway.publishRaw(gatewayRpcTopic, rpc_json)

	log.Info().Msgf("Published RPC reply #%d of device %s: %s", msg.RequestId, msg.Device, rpc_json)
}

// Reply to gateway_ping
func (gateway *TBMQTTGatewayClient) ReplyGatewayPingRPC(requestId int) {
	// https://thingsboard.io/docs/iot-gateway/guides/how-to-use-gateway-rpc-methods/#gateway_ping-rpc-method
	response := map[string]any{
		"code": 200,
		"resp": "pong",
	}
	msg_json, _ := json.Marshal(response)
	gateway.ReplyRPC(requestId, msg_json)
}

// Reply to gateway_devices
func (gateway *TBMQTTGatewayClient) ReplyGatewayDevicesRPC(requestId int) {
	// https://thingsboard.io/docs/iot-gateway/guides/how-to-use-gateway-rpc-methods/#gateway_devices-rpc-method
	devices := make(map[string]string, gateway.connectedDevices.Size())
	for device := range gateway.connectedDevices.Iterator() {
		devices[device] = "default"
	}
	response := map[string]any{
		"code": 200,
		"resp": devices,
	}

	// map to json bytes
	msg_json, _ := json.Marshal(response)

	gateway.ReplyRPC(requestId, msg_json)
}
