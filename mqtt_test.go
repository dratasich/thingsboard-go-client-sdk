package mqtt

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestConnect(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping TB connection test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cfg := Config{
		// free to use, see https://test.mosquitto.org/
		ServerURL: "mqtt://test.mosquitto.org:1883",
		// 1883/8883 is unauthenticated
		Username:  "",
		Password:  "",
		KeepAlive: 60,
	}
	tbmqtt := NewClient(cfg)
	tbmqtt.Connect(ctx)

	assert.NotNil(t, tbmqtt, "Failed to create TBMQTT instance")
	tbmqtt.Disconnect(ctx)
	assert.False(t, tbmqtt.isConnected, "TBMQTT should be disconnected")
}
