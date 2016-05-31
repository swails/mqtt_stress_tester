package flooding

import (
	"crypto/tls"
	"fmt"
	"killswitch"
	"math"
	"math/rand"
	"messages"
	"mqtt"
	"mqtt/randomcreds"
	"os"
	"sync"
	"time"
)

const maxGoRoutines = 100000 // a cap to prevent blowing memory

var connectionTimeout time.Duration = 10 * time.Second

var firstErrorMessage chan error = make(chan error, 1) // store a single error

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
func NewPublishFlooder(c *mqtt.MqttClient, mps int, mrv float64, ms int, msv float64, qos int, topic string) (*PublishFlooder, error) {
	if !c.IsConnected() {
		err := c.Connect(connectionTimeout)
		if err != nil {
			return nil, fmt.Errorf("creating pub flooder: %v", err)
		}
	}
	return &PublishFlooder{mps, mrv, ms, msv, Flooder{topic, qos, c}}, nil
}

// Creates a new SubscribeFlooder and returns a pointer. if the MQTT client is
// not connected, this creates that connection and starts the listening
func NewSubscribeFlooder(c *mqtt.MqttClient, qos int, topic string) (*SubscribeFlooder, error) {
	if !c.IsConnected() {
		err := c.Connect(connectionTimeout)
		if err != nil {
			return nil, fmt.Errorf("creating sub flooder: %v", err)
		}
	}
	ch, err := c.Subscribe(topic, qos)
	if err != nil {
		return nil, fmt.Errorf("subscribing in sub flooder: %v", err)
	}
	return &SubscribeFlooder{ch, Flooder{topic, qos, c}}, nil
}

// Creates a new SubscribeFlooder and PublishFlooder from the same MQTT client
func NewPubSubFlooder(c *mqtt.MqttClient, c2 *mqtt.MqttClient, mps int, mrv float64, ms int, msv float64, qos int, topic string) (*PublishFlooder, *SubscribeFlooder, error) {
	p, err := NewPublishFlooder(c, mps, mrv, ms, msv, qos, topic)
	if err != nil {
		return nil, nil, err
	}
	s, err := NewSubscribeFlooder(c2, qos, topic)
	if err != nil {
		return nil, nil, err
	}
	return p, s, nil
}

// Publishes on the MQTT channel continuously until the killswitch is triggered with the
// average rate and variance set when initializing the publish flooder
func (p *PublishFlooder) Publish(ks *killswitch.Killswitch, callback func()) int {
	defer func() {
		if p := recover(); p != nil {
			if err, ok := p.(error); ok {
				fmt.Fprintf(os.Stderr, "recovered panic: %v", err)
			} else {
				fmt.Fprintf(os.Stderr, "recovered panic; no error!")
			}
		}
	}()
	waitTime := 0 * time.Microsecond
	var numMessages int = 0
	msgChan := messages.GenerateRandomMessages(ks, p.MessageSize, p.MessageSizeVariance)
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
		case <-ks.Done():
			defer p.client.Disconnect()
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

type FlooderCollection struct {
	// The list of all publish flooders
	publishers []*PublishFlooder
	// The list of all subscribe flooders
	subscribers []*SubscribeFlooder
	// The total number of attempted connections
	nAttempted int
	// The total number of failed connections
	nFailed int
	// A count of messages that have been sent
	nMsgSent int
	// A running total of message lengths and how long it's taken to process these messages
	messages []float64
	// A sync mutex
	mux *sync.RWMutex
	// The counting mutex
	cmux *sync.RWMutex
}

// This creates a new FlooderCollection that returns instantly and begins populating its list of publishers and
// flooders. It creates no more than maxFlooders connections to the broker (whose connection settings are defined by
// hostname, username, password, port, and tlsConfig -- see mqtt/MqttClient for more information on these arguments). It
// will attempt new connections at intervals of connDelay (with a connection timeout equal to timeout) to avoid
// overloading the broker with attempted connections. The provided killswitch (k) provides a way to spin down all
// publishers and subscribers. The remaining parameters deal with the size and frequency of messages.
//
// Parameters:
//   mps - average messages per second
//   mrv - message rate variance
//   ms - average message size
//   msv - message size variance
//   qos - Quality of Service
//   topic - Subscription/publishing topic to use
func NewFlooderCollection(hostname, username, password string, port int, tlsConfig *tls.Config, maxFlooders int,
	connDelay, timeout time.Duration, k *killswitch.Killswitch, mps int, mrv float64, ms int, msv float64,
	qos int) *FlooderCollection {

	connectionTimeout = timeout
	fc := &FlooderCollection{
		publishers:  make([]*PublishFlooder, 0, maxFlooders),
		subscribers: make([]*SubscribeFlooder, 0, maxFlooders),
		messages:    make([]float64, 0),
		mux:         &sync.RWMutex{},
		cmux:        &sync.RWMutex{},
	}
	// Lock the message counter so we can't access it until it's released
	fc.cmux.Lock()
	// In case maxFlooders is not evenly divisible by maxGoRoutines, some goroutines will need to launch one more
	// connection than others. Specifically, nExtra goroutines will need to launch one extra connection
	nExtra := maxFlooders - (maxFlooders/maxGoRoutines)*maxGoRoutines
	nConns := maxFlooders / maxGoRoutines
	msgCntChan := make(chan int, maxFlooders)
	msgTimingsChan := make(chan float64, 100) // buffered
	for i := 0; i < mini(maxFlooders, maxGoRoutines); i++ {
		go func(connNumber, numToAttempt int) {
			for i := 0; i < numToAttempt; i++ {
				time.Sleep(time.Duration(connNumber+i*maxGoRoutines) * connDelay)
				pclient := mqtt.NewMqttClient(hostname, username, password, port, tlsConfig)
				sclient := mqtt.NewMqttClient(hostname, username, password, port, tlsConfig)
				err := pclient.Connect(connectionTimeout)
				if err != nil {
					select {
					case firstErrorMessage <- err:
						// nothing to do
					default:
						//nothing to do
					}
				}
				err = sclient.Connect(connectionTimeout)
				if err != nil {
					select {
					case firstErrorMessage <- err:
						// nothing to do
					default:
						//nothing to do
					}
				}
				topic := randomcreds.RandomTopic("test/")
				pub, sub, err := NewPubSubFlooder(pclient, sclient, mps, mrv, ms, msv, qos, topic)
				fc.mux.Lock()
				fc.nAttempted++
				if err != nil || pub == nil || sub == nil {
					fc.nFailed++
					fc.mux.Unlock()
					continue
				}
				k.Add()
				fc.publishers = append(fc.publishers, pub)
				fc.subscribers = append(fc.subscribers, sub)
				//fc.messages = append(fc.messages, msgs)
				fc.mux.Unlock()
				// Now start listening to each of the sub flooders
				go func(sub *SubscribeFlooder) {
				mainLoop:
					for {
						select {
						case msg := <-sub.SubChan:
							elapsed := processMessage(msg)
							msgTimingsChan <- float64(elapsed.Nanoseconds()) * 1e-9
						case <-k.Done():
							defer sub.client.Disconnect()
							break mainLoop
						}
					}
				}(sub)
				// Start the pub flooder publishing
				go func(pub *PublishFlooder) {
					msgCntChan <- pub.Publish(k, nil)
					k.Subtract()
				}(pub)
			}
		}(i, nConns+step(nExtra-i))
	}

	go func() {
		for {
			select {
			case val := <-msgTimingsChan:
				fc.messages = append(fc.messages, val)
			case <-k.Done():
				return
			}
		}
	}()

	go func() {
		// Block until we're done. We guaranteed that our msgCntChan will have enough space for every count, so it will
		// not block. But we won't have filled the channel until our killswitch was triggered
		<-k.Done()
		k.Wait() // wait for everyone to finish
		defer fc.cmux.Unlock()
	mainLoop:
		for i := 0; i < maxFlooders; i++ {
			select {
			case n := <-msgCntChan:
				fc.nMsgSent += n
			default:
				break mainLoop
			}
		}
	}()

	return fc
}

func (fc *FlooderCollection) NumAttempted() int {
	fc.mux.RLock()
	defer fc.mux.RUnlock()
	return fc.nAttempted
}

func (fc *FlooderCollection) NumFailed() int {
	fc.mux.RLock()
	defer fc.mux.RUnlock()
	return fc.nFailed
}

func (fc *FlooderCollection) NumSentMessages() int {
	fc.cmux.RLock()
	defer fc.cmux.RUnlock()
	return fc.nMsgSent
}

func (fc *FlooderCollection) MessageTimings() []float64 {
	fc.mux.RLock()
	defer fc.mux.RUnlock()
	return fc.messages
}

// Returns 1 if i is greater than 0, 0 otherwise
func step(i int) int {
	if i <= 0 {
		return 0
	}
	return 1
}

// returns the smaller of two integers
func mini(i, j int) int {
	if i <= j {
		return i
	}
	return j
}

func processMessage(msg []byte) time.Duration {
	now := time.Duration(time.Now().UnixNano())
	msgTime := messages.ExtractTimeFromMessage(msg)
	return now - msgTime
}

func GetFirstErrorMessage() error {
	select {
	case err := <-firstErrorMessage:
		return err
	default:
		return nil
	}
}
