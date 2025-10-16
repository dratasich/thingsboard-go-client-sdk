package events

// Device provisioning request
//
// see also
// - api: https://thingsboard.io/docs/reference/mqtt-api/#device-provisioning
// - structure: https://thingsboard.io/docs/user-guide/rpc/#server-side-rpc-structure
type DeviceProvisioningRequest struct {
	DeviceName            string  `json:"deviceName"`
	ProvisionDeviceKey    string  `json:"provisionDeviceKey"`
	ProvisionDeviceSecret string  `json:"provisionDeviceSecret"`
	CredentialsType       *string `json:"credentialsType,omitempty"` // optional
	Token                 *string `json:"token,omitempty"`           // optional
	CertificateHash       *string `json:"hash,omitempty"`            // optional
}

// Device provisioning response
type DeviceProvisioningResponse struct {
	DeviceId         string `json:"deviceId"`
	CredentialsType  string `json:"credentialsType"`
	CredentialsId    string `json:"credentialsId"`
	CredentialsValue string `json:"credentialsValue"`
	Status           string `json:"status"`

	// on error
	ErrorMsg string `json:"errorMsg"`
}

// Type of device credentials
const (
	CredentialsTypeAccessToken string = "ACCESS_TOKEN"
	CredentialsTypeX509        string = "X509_CERTIFICATE"
)

// Provisioning response status
const (
	ProvisionStatusSuccess string = "SUCCESS"
	ProvisionStatusFailure string = "FAILURE"
)
