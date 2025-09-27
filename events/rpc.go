package events

// RPC request
//
// see also
// - api: https://thingsboard.io/docs/reference/mqtt-api/#server-side-rpc
// - structure: https://thingsboard.io/docs/user-guide/rpc/#server-side-rpc-structure
type RequestRPC struct {
	// Unique ID of the request
	//
	// derived from the topic name
	RpcRequestId int32 `json:"id"`

	// rest is parsed from the payload

	Method string `json:"method"`
	Params any    `json:"params"`
}

// RPC request for gateway
type GatewayRequestRPC struct {
	// Device name
	Device string `json:"device"`
	// Embedded RPC request for the device
	Data RequestRPC
}

// RPC response may be any json so we don't specify a model here
//type ResponseRPC any

// however, the gateway must follow a structure

// RPC response of the gateway
type GatewayResponseRPC struct {
	Device    string `json:"device"`
	RequestId int32  `json:"id"`
	Data      any    `json:"data"`
}
