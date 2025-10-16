package mqtt

import (
	"strings"
	"testing"

	"github.com/dratasich/thingsboard-go-client-sdk/events"
	"github.com/stretchr/testify/assert"
)

func provisioningClientFixture(t *testing.T) *TBMQTTProvisioningClient {
	cfg := Config{
		// free to use, see https://test.mosquitto.org/
		ServerURL: "mqtt://test.mosquitto.org:1883",
		// 1883/8883 is unauthenticated
		Username: "",
		Password: "",
		// provisioning key/secret (dummy values, not used on public broker)
		ProvisioningKey:    "my_provision_key",
		ProvisioningSecret: "my_provision_secret",
		KeepAlive:          60,
	}
	return NewProvisioningClient(cfg)
}

func TestUsername(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping TB connection test")
	}

	tbmqtt := provisioningClientFixture(t)

	assert.Equal(t, "provision", tbmqtt.TBMQTT.config.Username)
}

func TestResponseCheck(t *testing.T) {
	onelineHash := "MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAz6Xg+F7l+3cH4bO8x0Yp" +
		"AQEAt42wOFRVmKfr4/f/BrrFJgL3eoWjSBgND9u/MGPdVk4c6HMaUdqdcxMdh29J" +
		"yYxcMs3+adIwG+hqRtr2gkqMf0mTKFgVNR2v=="
	success := events.DeviceProvisioningResponse{
		DeviceId:         "3b829220-232f-11eb-9d5c-e9ed3235dff8",
		CredentialsType:  "X509_CERTIFICATE",
		CredentialsId:    "f307a1f717a12b32c27203cf77728d305d29f64694a8311be921070dd1259b3",
		CredentialsValue: onelineHash,
		Status:           "SUCCESS",
	}
	notFound := events.DeviceProvisioningResponse{
		ErrorMsg: "Provision data was not found!",
		Status:   "NOT_FOUND",
	}
	failed := events.DeviceProvisioningResponse{
		ErrorMsg: "Failed to provision device!",
		Status:   "FAILURE",
	}
	emptyCredentials := events.DeviceProvisioningResponse{
		Status:           "SUCCESS",
		CredentialsType:  "ACCESS_TOKEN",
		CredentialsValue: "",
	}
	wrongType := events.DeviceProvisioningResponse{
		Status:           "SUCCESS",
		CredentialsType:  "PSK",
		CredentialsValue: onelineHash,
	}

	assert.True(t, IsProvisionResponseOK(&success))
	assert.False(t, IsProvisionResponseOK(&notFound))
	assert.False(t, IsProvisionResponseOK(&failed))
	assert.False(t, IsProvisionResponseOK(&emptyCredentials))
	assert.False(t, IsProvisionResponseOK(&wrongType))
}

func TestPemToHashAndBack(t *testing.T) {
	pem := `-----BEGIN CERTIFICATE-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAz6Xg+F7l+3cH4bO8x0Yp
AQEAt42wOFRVmKfr4/f/BrrFJgL3eoWjSBgND9u/MGPdVk4c6HMaUdqdcxMdh29J
yYxcMs3+adIwG+hqRtr2gkqMf0mTKFgVNR2v==
-----END CERTIFICATE-----`
	pemWithExtraNewline := pem + "\n"

	onelineHash := "MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAz6Xg+F7l+3cH4bO8x0Yp" +
		"AQEAt42wOFRVmKfr4/f/BrrFJgL3eoWjSBgND9u/MGPdVk4c6HMaUdqdcxMdh29J" +
		"yYxcMs3+adIwG+hqRtr2gkqMf0mTKFgVNR2v=="

	// convert to hash for TB
	// both PEM formats should give the same hash
	hash1 := PemToHash(pem)
	hash2 := PemToHash(pemWithExtraNewline)

	assert.Equal(t, onelineHash, hash1)
	assert.Equal(t, onelineHash, hash2)

	// back to PEM format
	pemBack := PemFromHash(onelineHash)
	assert.Equal(t, pem, pemBack)

	// check format (first and last line, and all are 64 chars max)
	lines := strings.Split(pemBack, "\n")
	assert.Equal(t, lines[0], "-----BEGIN CERTIFICATE-----")
	assert.Equal(t, lines[len(lines)-1], "-----END CERTIFICATE-----")
	for _, line := range lines[1 : len(lines)-1] {
		assert.LessOrEqual(t, len(line), 64)
	}
}

func TestExtraShortPem(t *testing.T) {
	pem := `-----BEGIN CERTIFICATE-----
yYxcMs3+adIwG+hqRtr2gkqMf0mTKFgVNR2v==
-----END CERTIFICATE-----`

	onelineHash := "yYxcMs3+adIwG+hqRtr2gkqMf0mTKFgVNR2v=="

	// convert to hash for TB
	hash := PemToHash(pem)

	assert.Equal(t, onelineHash, hash)

	// back to PEM format
	pemBack := PemFromHash(onelineHash)
	assert.Equal(t, pem, pemBack)
}
