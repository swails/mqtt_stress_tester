package mqtt

import (
	"crypto/tls"
	"fmt"
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
		t.Error("expected error connecting to TLS port with TCP")
	}

	client.Disconnect()
	if client.IsConnected() {
		t.Error("Expected client disconnect")
	}
}

// Tests MQTT connection with TLS encryption. Must have broker running with port
// 8883 open for encrypted connections on localhost (e.g., w/ mosquitto)
func TestMqttConnectionWithTLS(t *testing.T) {
	cfg, err := NewTLSAnonymousConfig("files/ca.crt")
	if err != nil {
		t.Error("unexpected error in generating TLS config: %v", err)
	}

	client := NewMqttClient(HOSTNAME, USERNAME, PASSWORD, TLS_PORT, cfg)

	if err := client.Connect(1 * time.Second); err != nil {
		t.Error("Unexpected error connected without TLS to localhost: " + err.Error())
	}

	// Now try what should be illegal credentials
	client = NewMqttClient(HOSTNAME, USERNAME, "badpassword", TLS_PORT, cfg)
	if err := client.Connect(1 * time.Second); err == nil {
		t.Error("expected error connecting with bad password")
	}

	client = NewMqttClient(HOSTNAME, "badusername", PASSWORD, TLS_PORT, cfg)
	if err := client.Connect(1 * time.Second); err == nil {
		t.Error("expected error connecting with bad username")
	}

	client = NewMqttClient(HOSTNAME, USERNAME, PASSWORD, TCP_PORT, cfg)
	if err := client.Connect(1 * time.Second); err == nil {
		t.Error("expected error connecting to TCP port with TLS")
	}

	client.Disconnect()
	if client.IsConnected() {
		t.Error("Expected client disconnect")
	}
}

// Tests pub/sub over TCP connection. Must have broker running with port 1883
// open for encrypted connections on localhost (e.g., w/ mosquitto)
func TestMqttTCPPubSub(t *testing.T) {
	pubclient := NewMqttClient(HOSTNAME, USERNAME, PASSWORD, TCP_PORT, nil)
	subclient := NewMqttClient(HOSTNAME, USERNAME, PASSWORD, TCP_PORT, nil)

	doPubSubTests(pubclient, subclient, t)
}

func TestMqttTLSPubSub(t *testing.T) {
	cfg, err := NewTLSAnonymousConfig("files/ca.crt")
	if err != nil {
		t.Errorf("Unexpected error getting TLS configuration: %v", err)
	}
	pubclient := NewMqttClient(HOSTNAME, USERNAME, PASSWORD, TLS_PORT, cfg)
	subclient := NewMqttClient(HOSTNAME, USERNAME, PASSWORD, TLS_PORT, cfg)

	doPubSubTests(pubclient, subclient, t)
}

func doPubSubTests(pubclient, subclient *MqttClient, t *testing.T) {

	if err := pubclient.Connect(1 * time.Second); err != nil {
		t.Error("unexpected problem connecting pubclient to broker")
	}
	if err := subclient.Connect(1 * time.Second); err != nil {
		t.Error("unexpected problem connecting subclient to broker")
	}

	subChan, err := subclient.Subscribe("test/topic", 0)
	if err != nil {
		t.Errorf("unexpected error subscribing to test/topic: %v", err)
	}

	err = pubclient.Publish("test/topic", 0, []byte("test message"))

	if err != nil {
		t.Errorf("unexpected error publishing to test/topic: %v", err)
	}

	x := string(<-subChan)

	if x != "test message" {
		t.Errorf("Expected subscription to receive '%s'. Got '%s' instead.", "test message", x)
	}

	// Publish again

	err = pubclient.Publish("test/topic", 0, []byte("test message 2"))
	if err != nil {
		t.Errorf("unexpected error publishing second time: %v", err)
	}

	x = string(<-subChan)

	if x != "test message 2" {
		t.Errorf("Expected subscription to receive '%s'. Got '%s' instead.", "test message 2", x)
	}
	// Multi-publish

	var num_messages int = 100

	go func() {
		messages := make([]string, num_messages)
		for i := 0; i < num_messages; i++ {
			messages = append(messages, string(<-subChan))
		}

		// Now check that I got what I expected
		for i := 0; i < num_messages; i++ {
			if messages[i] != fmt.Sprintf("test message swarm %d", i) {
				t.Errorf("Expected subscription to receive '%s'. Got '%s' instead.",
					fmt.Sprintf("test message swarm %d", i), messages[i])
			}
		}
	}()

	for i := 0; i < num_messages; i++ {
		err = pubclient.Publish("test/topic", 0, []byte(fmt.Sprintf("test message swarm %d", i)))
		if err != nil {
			t.Errorf("Unexpected error publishing swarm number %d", i)
		}
	}

}
