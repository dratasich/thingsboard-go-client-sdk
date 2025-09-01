package events

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/stretchr/testify/assert"
)

type CustomConfig struct {
	Name    string `json:"name"`
	Timeout int    `json:"timeout"`
}

func TestAttributes(t *testing.T) {
	// arrange
	jsonData := `
	{
		"name": "test device",
		"timeout": 60
	}`

	// act
	var attrs Attributes
	if err := json.Unmarshal([]byte(jsonData), &attrs); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}
	var config CustomConfig
	if err := mapstructure.Decode(attrs, &config); err != nil {
		t.Fatalf("Failed to decode attributes: %v", err)
	}

	// assert
	assert.Equal(t, "test device", config.Name)
	assert.Equal(t, 60, config.Timeout)
}

func TestAttributesResponseSharedOnly(t *testing.T) {
	// arrange
	jsonData := `
	{
		"shared": {
			"name": "test device",
			"timeout": 60
		}
	}`

	// act
	var attr ResponseAttributes
	if err := json.Unmarshal([]byte(jsonData), &attr); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}
	var config CustomConfig
	if err := mapstructure.Decode(attr.SharedAttr, &config); err != nil {
		t.Fatalf("Failed to decode attributes: %v", err)
	}

	// assert
	assert.Equal(t, "test device", config.Name)
	assert.Equal(t, 60, config.Timeout)
}

type CustomParameters struct {
	Pin   int `json:"pin"`
	Value int `json:"value"`
}

func TestRequestRPC(t *testing.T) {
	// arrange
	// https://thingsboard.io/docs/user-guide/rpc/#server-side-rpc-structure
	jsonData := `
	{
		"method": "setGPIO",
		"params": {
			"pin": 4,
			"value": 1
		},
		"timeout": 30000
	}`

	// act
	var req RequestRPC
	if err := json.Unmarshal([]byte(jsonData), &req); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}
	params := CustomParameters{}
	if err := mapstructure.Decode(req.Params, &params); err != nil {
		t.Fatalf("Failed to decode params: %v", err)
	}

	// assert
	assert.Equal(t, "setGPIO", req.Method)
	assert.Equal(t, 4, params.Pin)
	assert.Equal(t, 1, params.Value)
}

type CustomTelemetry struct {
	Temperature float64 `json:"temperature"`
	Humidity    int     `json:"humidity"`
}

func TestTelemetryCustomExample(t *testing.T) {
	// arrange
	data := CustomTelemetry{
		Temperature: 22.5,
		Humidity:    60,
	}
	startTime := time.Now().UnixMilli()

	// act
	// data to telemetry event
	event := Telemetry{
		Timestamp: time.Now().UnixMilli(),
		Values: map[string]any{
			"temperature": data.Temperature,
			"humidity":    data.Humidity,
		},
	}
	jsonData, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("Failed to marshal telemetry: %v", err)
	}

	// assert
	assert.GreaterOrEqual(t, event.Timestamp, startTime)
	assert.Equal(t, 22.5, event.Values["temperature"])
	assert.Equal(t, 60, event.Values["humidity"])
	assert.Contains(t, string(jsonData), `"ts":`)
	assert.Contains(t, string(jsonData), `"values":`)
	assert.Contains(t, string(jsonData), `"temperature":`)
}

func TestTelemetryToJson(t *testing.T) {
	// arrange
	event := Telemetry{
		Timestamp: time.Now().UnixMilli(),
		Values: map[string]any{
			"temperature": 22.5,
			"humidity":    60,
		},
	}

	// act
	jsonData, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("Failed to marshal telemetry: %v", err)
	}

	// assert
	assert.Contains(t, string(jsonData), `"ts":`)
	assert.Contains(t, string(jsonData), `"values":`)
	assert.Contains(t, string(jsonData), `"temperature":`)
}
