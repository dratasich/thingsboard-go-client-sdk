package model

// RPC request
//
// see also
// - api: https://thingsboard.io/docs/reference/mqtt-api/#server-side-rpc
// - structure: https://thingsboard.io/docs/user-guide/rpc/#server-side-rpc-structure
type RequestRPC struct {
	// Unique ID of the request
	//
	// derived from the topic name
	RpcRequestId string

	// rest is parsed from the payload

	Method string                 `json:"method"`
	Params map[string]interface{} `json:"params"`
}

// RPC response may be any json so we don't specify a model here
