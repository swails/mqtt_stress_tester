package mqtt

import (
	"crypto/tls"
	"fmt"
	"os"
	"sync"
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
	subchan  chan []byte // a holder so we can close it when necessary
	*sync.RWMutex
}

// Create a new pointer to a MqttClient instance. Setting tlsConfig to nil disables TLS encryption
func NewMqttClient(hostname, username, password string, port int, tlsConfig *tls.Config) *MqttClient {
	co := paho.NewClientOptions().SetAutoReconnect(false)
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
	return &MqttClient{hostname, username, password, port, co, client, nil, &sync.RWMutex{}}
}

// Connect to the broker, waiting a specified timeout (seconds)
func (c *MqttClient) Connect(timeout time.Duration) error {
	c.Lock()
	defer c.Unlock()
	defer func() {
		if p := recover(); p != nil {
			if err, ok := p.(error); ok {
				fmt.Fprintf(os.Stderr, "recovered panic: %v", err)
			} else {
				fmt.Fprintf(os.Stderr, "recovered panic; no error!")
			}
		}
	}()
	if conn := c.client.Connect(); conn.WaitTimeout(timeout) && conn.Error() != nil {
		return fmt.Errorf("connection error: %v", conn.Error())
	}
	return nil
}

// Return whether or not the client is connected to its server
func (c *MqttClient) IsConnected() bool {
	c.RLock()
	defer c.RUnlock()
	if c.client == nil {
		return false
	}
	return c.client.IsConnected()
}

// Disconnects from the broker
func (c *MqttClient) Disconnect() {
	if c.client.IsConnected() {
		c.Lock()
		defer c.Unlock()
		c.client.Disconnect(250)
	}
}

// Subscribes to the broker and sends messages it receives back on a channel
// with a buffer for 10 messages. You must make sure to pull messages off this
// channel to avoid blocking the receiver and losing messages.
func (c *MqttClient) Subscribe(topic string, qos int) (<-chan []byte, error) {
	c.RLock()
	if !c.client.IsConnected() {
		c.RUnlock()
		return nil, fmt.Errorf("not connected to subscribing broker")
	}
	c.RUnlock()
	c.Lock()
	defer c.Unlock()
	subChan := make(chan []byte, 10)

	// The callback puts the received message on the subChan
	c.client.Subscribe(topic, byte(qos), func(c paho.Client, m paho.Message) {
		payload := m.Payload()
		subChan <- payload
	})
	c.subchan = subChan
	return subChan, nil
}

// Publishes a message to a particular topic on a broker
func (c *MqttClient) Publish(topic string, qos int, payload []byte) error {
	if !c.client.IsConnected() {
		return fmt.Errorf("not connected to publishing broker")
	}
	c.Lock()
	defer c.Unlock()
	token := c.client.Publish(topic, byte(qos), true, payload)
	// Try to recover from any panic during publishing
	if p := recover(); p != nil {
		if err, ok := p.(error); ok {
			fmt.Fprintf(os.Stderr, "Recovered from panic: %v", err)
			return err
		}
		fmt.Fprintf(os.Stderr, "Recovered from panic; unknown error")
		return fmt.Errorf("recovered from panic; unknown error")
	}
	if ret := token.Wait(); ret && token.Error() != nil {
		return token.Error()
	}
	return nil
}

/*
// Closes the subscription channel. If it is not open, this is a no-op
func (c *MqttClient) CloseSubchannel() {
	c.Lock()
	defer c.Unlock()
	if c.subchan != nil {
		close(c.subchan)
		c.subchan = nil
	}
}
*/
