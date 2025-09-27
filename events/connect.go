package events

// Connect a device message
//
// see also:
// https://thingsboard.io/docs/reference/gateway-mqtt-api/#device-connect-api
type Connect struct {
	// device name (unique name in TB)
	Device string `json:"device"`
	// device profile
	Type string `json:"type"`
}

// Disconnect a device message
//
// see also:
// https://thingsboard.io/docs/reference/gateway-mqtt-api/#device-disconnect-api
type Disconnect struct {
	Device string `json:"device"`
}
