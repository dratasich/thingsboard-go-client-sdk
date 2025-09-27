package events

// Attributes
type Attributes map[string]any

// Attributes request
//
// see also
// - api: https://thingsboard.io/docs/reference/mqtt-api/#request-attribute-values-from-the-server
type RequestAttributes struct {
	ClientKeys string `json:"clientKeys"`
	SharedKeys string `json:"sharedKeys"`
}

// Attributes response
//
// see also
// - api: https://thingsboard.io/docs/reference/mqtt-api/#request-attribute-values-from-the-server
type ResponseAttributes struct {
	// id (references request)
	//
	// extracted from the topic
	RequestId int32

	// rest is payload

	ClientAttr *map[string]any `json:"client"`
	SharedAttr *map[string]any `json:"shared"`
}

// Device attributes sent by gateway
type GatewaySendAttributes map[string]Attributes

// Device attributes updates received by gateway
type GatewayAttributes struct {
	Device string     `json:"device"`
	Data   Attributes `json:"data"`
}

// Gateway attribute request
//
// https://thingsboard.io/docs/reference/gateway-mqtt-api/#request-attribute-values-from-the-server
//
// though a keys list is not documented, it works :)
type GatewayRequestAttributes struct {
	RequestId     int32    `json:"id"`
	Device        string   `json:"device"`
	AreClientKeys bool     `json:"client"`
	Key           string   `json:"key,omitempty"`
	Keys          []string `json:"keys,omitempty"`
}

// Gateway attribute response
type GatewayResponseAttributes struct {
	RequestId int32      `json:"id"`
	Device    string     `json:"device"`
	Value     any        `json:"value,omitempty"`  // single value
	Values    Attributes `json:"values,omitempty"` // multiple values
}
