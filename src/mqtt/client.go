package mqtt

import (
	"crypto/tls"
	"fmt"
	"time"

	paho "github.com/eclipse/paho.mqtt.golang"
)

// Client for connecting to MQTT brokers
type MqttClient struct {
	hostname string
	username string
	password string
	port     int
	co       *paho.ClientOptions
	client   paho.Client
}

// Create a new pointer to a MqttClient instance. Setting tlsConfig to nil disables TLS encryption
func NewMqttClient(hostname, username, password string, port int, tlsConfig *tls.Config) *MqttClient {
	co := paho.NewClientOptions()
	var broker string
	if tlsConfig == nil {
		broker = "tcp://" + hostname + ":" + fmt.Sprintf("%d", port)
	} else {
		broker = "tls://" + hostname + ":" + fmt.Sprintf("%d", port)
	}
	co.AddBroker(broker)
	if len(username) > 0 {
		co.SetUsername(username)
		if len(password) > 0 {
			co.SetPassword(password)
		}
	}
	if tlsConfig != nil {
		co.SetTLSConfig(tlsConfig)
	}
	client := paho.NewClient(co)
	return &MqttClient{hostname, username, password, port, co, client}
}

// Connect to the broker, waiting a specified timeout (seconds)
func (c *MqttClient) Connect(timeout time.Duration) error {
	if conn := c.client.Connect(); conn.WaitTimeout(timeout) && conn.Error() != nil {
		return fmt.Errorf("connection error: %v", conn.Error())
	}
	return nil
}

// Return whether or not the client is connected to its server
func (c *MqttClient) IsConnected() bool {
	if c.client == nil {
		return false
	}
	return c.client.IsConnected()
}

// Disconnects from the broker
func (c *MqttClient) Disconnect() {
	if c.client.IsConnected() {
		c.client.Disconnect(250)
	}
}

// Subscribes to the broker and sends messages it receives back on a channel
// with a buffer for 10 messages. You must make sure to pull messages off this
// channel to avoid blocking the receiver and losing messages.
func (c *MqttClient) Subscribe(topic string, qos int) (<-chan []byte, error) {
	if !c.client.IsConnected() {
		return nil, fmt.Errorf("not connected to subscribing broker")
	}
	subChan := make(chan []byte, 10)

	// The callback puts the received message on the subChan
	c.client.Subscribe(topic, byte(qos), func(c paho.Client, m paho.Message) {
		payload := m.Payload()
		subChan <- payload
	})
	return subChan, nil
}

// Publishes a message to a particular topic on a broker
func (c *MqttClient) Publish(topic string, qos int, payload []byte) error {
	if !c.client.IsConnected() {
		return fmt.Errorf("not connected to publishing broker")
	}
	token := c.client.Publish(topic, byte(qos), true, payload)
	if ret := token.Wait(); ret && token.Error() != nil {
		return token.Error()
	}
	return nil
}
