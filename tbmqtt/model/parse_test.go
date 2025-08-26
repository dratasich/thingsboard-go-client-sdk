package model

import (
	"encoding/json"
	"testing"

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
	// TODO: add a better way to cast the params to a custom type
	// map[string]interface{} would fit better, but didn't manage to cast it without extra packages
	// https://github.com/mitchellh/mapstructure
	var params CustomParameters
	paramsData, err := json.Marshal(req.Params)
	if err != nil {
		t.Fatalf("Failed to marshal params: %v", err)
	}
	if err := json.Unmarshal(paramsData, &params); err != nil {
		t.Fatalf("Failed to unmarshal params: %v", err)
	}

	// assert
	assert.Equal(t, "setGPIO", req.Method)
	assert.Equal(t, 4, params.Pin)
	assert.Equal(t, 1, params.Value)
}
