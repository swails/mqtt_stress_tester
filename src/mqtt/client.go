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

// Create a new pointer to a MqttClient instance. Setting tlsConfig to nil
// disables TLS encryption
func NewMqttClient(hostname, username, password string, port int, tlsConfig *tls.Config) *MqttClient {
	co := paho.NewClientOptions()
	var broker string
	if tlsConfig == nil {
		broker = "tcp://" + hostname + ":" + fmt.Sprintf("%d", port)
	} else {
		broker = "tcp://" + hostname + ":" + fmt.Sprintf("%d", port)
	}
	fmt.Println("Adding broker: " + broker)
	co.AddBroker(broker)
	co.SetPassword(password)
	co.SetUsername(username)
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
