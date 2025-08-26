package model

import (
	"encoding/json"
	"testing"

	"github.com/mitchellh/mapstructure"
	"github.com/stretchr/testify/assert"
)

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
