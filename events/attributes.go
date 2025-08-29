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
	Id string

	// rest is payload

	ClientAttr *map[string]any `json:"client"`
	SharedAttr *map[string]any `json:"shared"`
}
