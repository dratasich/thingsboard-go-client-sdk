package mqtt

import (
	"context"
	"testing"
	"time"

	"github.com/dratasich/thingsboard-go-client-sdk/events"
	"github.com/stretchr/testify/assert"
)

func fixture(t *testing.T) *TBMQTT {
	cfg := Config{
		// free to use, see https://test.mosquitto.org/
		ServerURL: "mqtt://test.mosquitto.org:1883",
		// 1883/8883 is unauthenticated
		Username:  "",
		Password:  "",
		KeepAlive: 60,
	}
	return NewClient(cfg)
}

func TestConnect(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping TB connection test")
	}

	tbmqtt := fixture(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tbmqtt.Connect(ctx)

	assert.NotNil(t, tbmqtt, "Failed to create TBMQTT instance")
	tbmqtt.Disconnect(ctx)
	assert.False(t, tbmqtt.isConnected, "TBMQTT should be disconnected")
}

func TestAttributeUpdates(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping TB integration test")
	}

	tbmqtt := fixture(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	tbmqtt.Connect(ctx)

	// collect received attributes
	var attrs []*events.Attributes
	go func() {
		for {
			attr := <-tbmqtt.AttributesQueue
			attrs = append(attrs, attr)
		}
	}()

	// publish client attributes
	attr := events.Attributes{
		"serial":  "123",
		"timeout": 60,
	}
	tbmqtt.PublishAttributes(attr)

	// await send and receive
	time.Sleep(500 * time.Millisecond)
	assert.Equal(t, 1, len(attrs), "Expected 1 attribute update to be received")
	assert.Equal(t, "123", (*attrs[0])["serial"], "Expected serial attribute to match")

	// cleanup
	tbmqtt.Disconnect(ctx)
	assert.False(t, tbmqtt.isConnected, "TBMQTT should be disconnected")
}
