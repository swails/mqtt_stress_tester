package mqtt

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
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

	if err := client.Connect(1 * time.Second); err != nil {
		t.Error("unexpected error connecting without TLS to localhost: " + err.Error())
	}

	if !client.IsConnected() {
		t.Error("expected connection to broker")
	}

	// Now try what should be illegal credentials
	client = NewMqttClient(HOSTNAME, USERNAME, "badpassword", TCP_PORT, cfg)
	if err := client.Connect(1 * time.Second); err == nil {
		t.Error("expected error connecting with bad password")
	}

	client = NewMqttClient(HOSTNAME, "badusername", PASSWORD, TCP_PORT, cfg)
	if err := client.Connect(1 * time.Second); err == nil {
		t.Error("expected error connecting with bad username")
	}

	client = NewMqttClient(HOSTNAME, USERNAME, PASSWORD, TLS_PORT, cfg)
	if err := client.Connect(1 * time.Second); err == nil {
		t.Error("expected error connecting with TLS port over TCP")
	}

}

// Tests MQTT connection with TLS encryption. Must have broker running with port
// 8883 open for encrypted connections on localhost (e.g., w/ mosquitto)
func TestMqttConnectionWithTLS(t *testing.T) {
	// Generate the certificate authority token
	caCert, err := ioutil.ReadFile("files/ca.crt")
	if err != nil {
		t.Error("Unexpected failure reading CA cert")
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	cfg := &tls.Config{
		MinVersion: tls.VersionTLS12,
		ClientAuth: tls.NoClientCert,
		RootCAs:    caCertPool,
	}

	client := NewMqttClient(HOSTNAME, USERNAME, PASSWORD, TLS_PORT, cfg)

	if err := client.Connect(1 * time.Second); err != nil {
		t.Error("Unexpected error connected without TLS to localhost: " + err.Error())
	}
}
