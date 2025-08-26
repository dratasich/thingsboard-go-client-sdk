package mqtt

import (
	"context"
	"encoding/json"
	"net/url"
	"strings"
	"time"

	"github.com/dratasich/thingsboard-go-client-sdk/events"
	"github.com/eclipse/paho.golang/autopaho"
	"github.com/eclipse/paho.golang/paho"
	"github.com/rs/zerolog/log"
)

type TBMQTT struct {
	client      *autopaho.ConnectionManager
	isConnected bool

	// queue of RPC requests
	RpcQueue chan *events.RequestRPC
}

// MQTT configuration for ThingsBoard
type TBConfig struct {
	ServerURL string `env:"SERVER_URL"` // MQTT server URL
	// set username = tb access token (and leave password empty)
	Username string `env:"USERNAME"` // MQTT Username to use when connecting to server
	Password string `env:"PASSWORD"` // MQTT Password to use when connecting to server

	KeepAlive uint16 `env:"KEEP_ALIVE,default=60"` // seconds between keepalive packets
}

const (
	rpcTopic = "v1/devices/me/rpc/request/+" // topic to subscribe for RPC requests
	rpcQos   = byte(1)                       // qos to utilise when publishing

	requestTopicRpc  = "v1/devices/me/rpc/request/"
	responseTopicRpc = "v1/devices/me/rpc/response/"

	//keepaliveTopic = "v1/devices/me/attributes" // Keepalive topic to send to Thingsboard
)

// Connect via MQTT
// https://github.com/eclipse-paho/paho.golang/tree/master/autopaho
func (tbmqtt *TBMQTT) connect(ctx context.Context, cfg TBConfig, subscriptions []paho.SubscribeOptions, handler func(msg *paho.Publish)) {
	parsedURL, err := url.Parse(cfg.ServerURL)
	if err != nil {
		log.Fatal().Msgf("Failed to parse server URL (%s): %s", cfg.ServerURL, err)
	}

	cliCfg := autopaho.ClientConfig{
		BrokerUrls:                    []*url.URL{parsedURL},
		KeepAlive:                     cfg.KeepAlive,
		CleanStartOnInitialConnection: true,
		OnConnectionUp: func(cm *autopaho.ConnectionManager, connAck *paho.Connack) {
			log.Info().Msg("mqtt connection up")
			tbmqtt.isConnected = true
			if _, err := cm.Subscribe(context.Background(), &paho.Subscribe{
				Subscriptions: subscriptions,
			}); err != nil {
				log.Error().Msgf("failed to subscribe (%s). This is likely to mean no messages will be received.", err)
				return
			}
			log.Info().Msg("mqtt subscription made")
		},

		OnConnectError: func(err error) {
			log.Error().Msgf("error whilst attempting connection: %s", err)
		},

		ClientConfig: paho.ClientConfig{
			Router: paho.NewStandardRouterWithDefault(handler),
			OnClientError: func(err error) {
				log.Error().Msgf("client error: %s\n", err)
			},
			OnServerDisconnect: func(d *paho.Disconnect) {
				if d.Properties != nil {
					log.Error().Msgf("server requested disconnect: %s\n", d.Properties.ReasonString)
				} else {
					log.Error().Msgf("server requested disconnect; reason code: %d\n", d.ReasonCode)
				}
			},
		},
	}

	if cfg.Username != "" {
		cliCfg.ConnectUsername = cfg.Username
		cliCfg.ConnectPassword = []byte(cfg.Password)
	}

	//
	// Connect to the broker
	//
	//ctx, cancel := context.WithCancel(context.Background())
	//defer cancel()
	log.Info().Msg("Connect to Thingsboard MQTT...")
	tbmqtt.client, err = autopaho.NewConnection(ctx, cliCfg)
	if err != nil {
		log.Fatal().Msgf("Failed to connect to Thingsboard MQTT: %s", err)
	}
	// Wait for the connection to come up
	if err = tbmqtt.client.AwaitConnection(ctx); err != nil {
		log.Fatal().Msgf("Failed to connect to Thingsboard MQTT: %s", err)
	}
}

func CreateAndConnect(ctx context.Context, cfg TBConfig) *TBMQTT {
	tbmqtt := new(TBMQTT)
	tbmqtt.isConnected = false
	// buffered channel with 100 slots
	tbmqtt.RpcQueue = make(chan *events.RequestRPC, 100)

	var subscriptions = []paho.SubscribeOptions{
		// listen to RPC commands
		{
			Topic: rpcTopic,
			QoS:   byte(rpcQos),
		},
	}

	handler := func(msg *paho.Publish) {
		// check if RPC Command?
		if rpcId, found := strings.CutPrefix(msg.Topic, requestTopicRpc); found {
			log.Info().Msgf("RPC Request received with id #%s", rpcId)
			var rpc = events.RequestRPC{
				RpcRequestId: rpcId,
			}
			// check if RPC parsable
			err := json.Unmarshal(msg.Payload, &rpc)
			if err != nil {
				log.Error().Msgf("Message could not be parsed (%s): %s", msg.Payload, err)
			} else {
				// push to a queue
				log.Debug().Msgf("Pushing RPC request to queue: %s", rpc)
				tbmqtt.RpcQueue <- &rpc
			}
		} else {
			log.Error().Msgf("RPC Request Id could not be extracted (%s)", msg.Topic)
		}
	}

	tbmqtt.connect(ctx, cfg, subscriptions, handler)

	return tbmqtt
}

func (tbmqtt *TBMQTT) Disconnect(ctx context.Context) {
	if tbmqtt.client != nil {
		err := tbmqtt.client.Disconnect(ctx)
		if err != nil {
			log.Error().Msgf("failed to disconnect: %s", err)
		}
	}
	tbmqtt.isConnected = false
	log.Info().Msg("Disconnected from Thingsboard MQTT")
}

func (tbmqtt *TBMQTT) RPCReply(rpcRequestId string, payload_json []byte) {
	log.Debug().Msgf("sending RPC reply: \n%s\n", payload_json)

	responseTopic := responseTopicRpc + rpcRequestId
	responseMsg := &paho.Publish{
		QoS:     rpcQos,
		Topic:   responseTopic,
		Payload: payload_json,
	}

	// wait for mqtt connection
	for {
		if tbmqtt.isConnected {
			break
		}
		log.Info().Msg("Waiting for MQTT Connection ...")
		time.Sleep(1 * time.Second)
	}

	_, err := tbmqtt.client.Publish(context.Background(), responseMsg)
	if err != nil {
		log.Error().Msgf("failed to publish RPC reply: %s", err)
	}
}
