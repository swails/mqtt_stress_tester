package mqtt

import (
	"crypto/tls"
	"testing"
	"time"
)

const (
	HOSTNAME = "localhost"
	USERNAME = "stresstest"
	PASSWORD = "stressmeout"

	TCP_PORT = 1883
	TLS_PORT = 8883
)

// Tests MQTT connection without TLS encryption. Must have broker running with
// port 1883 open for unencrypted connections on localhost (e.g., w/ mosquitto)
func TestMqttConnectionNoTLS(t *testing.T) {

	var cfg *tls.Config
	client := NewMqttClient(HOSTNAME, USERNAME, PASSWORD, TCP_PORT, cfg)

	// Now try to connect with no TLS
	if err := client.Connect(1 * time.Second); err != nil {
		t.Error("unexpected error connecting without TLS to localhost: " + err.Error())
	}

	if !client.IsConnected() {
		t.Error("expected connection to broker")
	}
}
