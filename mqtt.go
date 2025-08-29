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

// MQTT configuration for ThingsBoard
type Config struct {
	ServerURL string `env:"SERVER_URL"` // MQTT server URL
	// set username = tb access token (and leave password empty)
	Username string `env:"USERNAME"` // MQTT Username to use when connecting to server
	Password string `env:"PASSWORD"` // MQTT Password to use when connecting to server

	KeepAlive uint16 `env:"KEEP_ALIVE,default=60"` // seconds between keepalive packets
}

type TBMQTT struct {
	config      Config
	client      *autopaho.ConnectionManager
	isConnected bool

	// queue of RPC requests
	RpcQueue chan *events.RequestRPC
}

const (
	rpcTopic = "v1/devices/me/rpc/request/+" // topic to subscribe for RPC requests
	rpcQos   = byte(1)                       // qos to utilise when publishing

	requestTopicRpc  = "v1/devices/me/rpc/request/"
	responseTopicRpc = "v1/devices/me/rpc/response/"

	//keepaliveTopic = "v1/devices/me/attributes" // Keepalive topic to send to Thingsboard
)

func NewClient(cfg Config) *TBMQTT {
	tbmqtt := &TBMQTT{
		config:      cfg,
		isConnected: false,
		RpcQueue:    make(chan *events.RequestRPC, 100),
	}
	return tbmqtt
}

func (tbmqtt *TBMQTT) Connect(ctx context.Context) {
	parsedURL, err := url.Parse(tbmqtt.config.ServerURL)
	if err != nil {
		log.Fatal().Msgf("Failed to parse server URL (%s): %s", tbmqtt.config.ServerURL, err)
	}

	var subscriptions = []paho.SubscribeOptions{
		// listen to RPC commands
		{
			Topic: rpcTopic,
			QoS:   byte(rpcQos),
		},
	}

	handler := func(msg *paho.Publish) {
		// handle RPCs
		if rpcId, found := strings.CutPrefix(msg.Topic, requestTopicRpc); found {
			log.Info().Msgf("RPC Request received with id #%s", rpcId)
			var rpc = events.RequestRPC{
				RpcRequestId: rpcId,
			}
			// check if RPC parsable
			err := json.Unmarshal(msg.Payload, &rpc)
			if err != nil {
				log.Error().Msgf("Message could not be parsed: %s. Payload: %s", err, msg.Payload)
			} else {
				// push to a queue
				log.Debug().Msgf("Pushing RPC request to queue: %s", rpc)
				tbmqtt.RpcQueue <- &rpc
			}
		} else {
			log.Error().Msgf("RPC request id could not be extracted from %s", msg.Topic)
		}
	}

	cliCfg := autopaho.ClientConfig{
		BrokerUrls:                    []*url.URL{parsedURL},
		KeepAlive:                     tbmqtt.config.KeepAlive,
		CleanStartOnInitialConnection: true,
		OnConnectionUp: func(cm *autopaho.ConnectionManager, connAck *paho.Connack) {
			log.Info().Msg("MQTT connection up")
			tbmqtt.isConnected = true
			if _, err := cm.Subscribe(context.Background(), &paho.Subscribe{
				Subscriptions: subscriptions,
			}); err != nil {
				log.Error().Msgf("Failed to subscribe: %s", err)
				return
			}
			log.Info().Msg("MQTT subscription made")
		},

		OnConnectError: func(err error) {
			log.Error().Msgf("Error whilst attempting connection: %s", err)
		},

		ClientConfig: paho.ClientConfig{
			Router: paho.NewStandardRouterWithDefault(handler),
			OnClientError: func(err error) {
				log.Error().Msgf("Client error: %s", err)
			},
			OnServerDisconnect: func(d *paho.Disconnect) {
				if d.Properties != nil {
					log.Error().Msgf("Server requested disconnect: %s", d.Properties.ReasonString)
				} else {
					log.Error().Msgf("Server requested disconnect with reason code: %d", d.ReasonCode)
				}
			},
		},
	}

	if tbmqtt.config.Username != "" {
		cliCfg.ConnectUsername = tbmqtt.config.Username
		cliCfg.ConnectPassword = []byte(tbmqtt.config.Password)
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

// Wait for MQTT connection is up
func (tbmqtt *TBMQTT) AwaitConnection() {
	for {
		if tbmqtt.isConnected {
			return
		}
		log.Info().Msg("Waiting for MQTT Connection ...")
		time.Sleep(5 * time.Second)
	}
}

func (tbmqtt *TBMQTT) Disconnect(ctx context.Context) {
	if tbmqtt.client != nil {
		err := tbmqtt.client.Disconnect(ctx)
		if err != nil {
			log.Error().Msgf("Failed to disconnect: %s", err)
		}
	}
	tbmqtt.isConnected = false
	log.Info().Msg("Disconnected from Thingsboard MQTT")
}

func (tbmqtt *TBMQTT) ReplyRPC(rpcRequestId string, payload_json []byte) {
	log.Debug().Msgf("Sending RPC reply: \n%s\n", payload_json)

	responseTopic := responseTopicRpc + rpcRequestId
	responseMsg := &paho.Publish{
		QoS:     rpcQos,
		Topic:   responseTopic,
		Payload: payload_json,
	}

	tbmqtt.AwaitConnection()

	_, err := tbmqtt.client.Publish(context.Background(), responseMsg)
	log.Info().Msgf("Published RPC reply for %s: %s", rpcRequestId, payload_json)
	if err != nil {
		log.Error().Msgf("Failed to publish RPC reply: %s", err)
	}
}
