package flooding

import (
	"fmt"
	"math"
	"math/rand"
	"time"

	"messages"
	"mqtt"
)

// Generic flooder type common to both subscription and publishing
type Flooder struct {
	// The topic we are listening and/or talking on
	Topic string
	// Quality of Service
	QoS int
	// The MQTT client we are communicating over
	client *mqtt.MqttClient
}

type SubscribeFlooder struct {
	// The subscription channel where messages come in on
	SubChan <-chan []byte
	// Attributes from the Flooder type
	Flooder
}

// Used to flood a particular topic with published messages
type PublishFlooder struct {
	// How many messages to send on average each second
	MessagesPerSecond int
	// The variance in the waiting times (which will be normally distributed)
	MessageRateVariance float64
	// The average size of the message in bytes
	MessageSize int
	// The variance of the message size in bytes
	MessageSizeVariance float64
	// Attributes from the Flooder type
	Flooder
}

// Creates a new PublishFlooder and returns a pointer. If the MQTT client is not
// connected, this creates that connection
func NewPublishFlooder(c *mqtt.MqttClient, mps int, mrv float64, ms int, msv float64, qos int, topic string) *PublishFlooder {
	if !c.IsConnected() {
		c.Connect(10 * time.Second) // Allow 10 seconds to connect
	}
	return &PublishFlooder{mps, mrv, ms, msv, Flooder{topic, qos, c}}
}

// Creates a new SubscribeFlooder and returns a pointer. if the MQTT client is
// not connected, this creates that connection and starts the listening
func NewSubscribeFlooder(c *mqtt.MqttClient, qos int, topic string) *SubscribeFlooder {
	if !c.IsConnected() {
		c.Connect(10 * time.Second) // Allow 10 seconds to connect
	}
	ch, err := c.Subscribe(topic, qos)
	if err != nil {
		panic(fmt.Sprintf("error subscribing to topic %s with QoS %d: %v", topic, qos, err))
	}
	return &SubscribeFlooder{ch, Flooder{topic, qos, c}}
}

// Creates a new SubscribeFlooder and PublishFlooder from the same MQTT client
func NewPubSubFlooder(c *mqtt.MqttClient, mps int, mrv float64, ms int, msv float64, qos int, topic string) (*PublishFlooder, *SubscribeFlooder) {
	return NewPublishFlooder(c, mps, mrv, ms, msv, qos, topic), NewSubscribeFlooder(c, qos, topic)
}

// Publishes on the MQTT channel continuously for the given duration with the
// average rate and variance
func (p *PublishFlooder) PublishFor(dur time.Duration, callback func()) int {
	waitTime := 0 * time.Microsecond
	var numMessages int = 0
	nmsg := int(float64(int(dur.Seconds())*p.MessagesPerSecond) * 1.5)
	msgChan := messages.GenerateRandomMessages(nmsg, p.MessageSize, p.MessageSizeVariance)
	// Make the done channel here as close as possible to where it will be used
	doneChan := time.After(dur)
	// Store some variables for later to determine how many ns to wait between pubs
	fac := math.Sqrt(p.MessageRateVariance) * 1e9
	avgWait := float64(1.0e9 / float64(p.MessagesPerSecond))
	var n_ns float64
	for {
		select {
		case <-time.After(waitTime):
			// Publish a random message
			err := p.client.Publish(p.Topic, 0, <-msgChan)
			if err == nil {
				numMessages += 1
			}
		case <-doneChan:
			if callback != nil {
				callback()
			}
			return numMessages
		}
		// Figure out how much time to wait until the next message in ns
		n_ns = math.Floor(rand.NormFloat64()*fac) + avgWait
		waitTime = time.Duration(int64(math.Max(n_ns, 100))) // always wait at least 100 ns
	}
	return -1
}

// Closes the subscription channel for a subscription flooder
func (s *SubscribeFlooder) Complete() {
	s.client.CloseSubchannel()
}
