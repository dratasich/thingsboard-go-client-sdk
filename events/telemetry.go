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
