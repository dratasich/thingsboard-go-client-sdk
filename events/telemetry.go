package events

// Telemetry with timestamp
//
// example:
// `{"ts": 1756742602000, "values": {"temperature": 22.5, "humidity": 60}}`
type Telemetry struct {
	// Unix timestamp in milliseconds
	Timestamp int64 `json:"ts"`
	// Key value pairs of telemetry data measured at the corresponding timestamp
	Values map[string]any `json:"values"`
}

// Batch of telemetry data of multiple devices
//
// Used by the gateway API to send multiple telemetries of multiple devices to TB.
//
// example:
// `{"Device A":[{"ts":1483228800000,"values":{"temperature":42,"humidity":80}},{"ts":1483228801000,"values":{"temperature":43,"humidity":82}}],"Device B":[{"ts":1483228800000,"values":{"temperature":42,"humidity":80}}]}`
type TelemetryBatch map[string][]Telemetry
