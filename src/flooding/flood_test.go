package flooding

import (
	"testing"
	"time"

	"mqtt"
	"mqtt/randomcreds"
)

const (
	HOSTNAME = "localhost"
	USERNAME = "stresstest"
	PASSWORD = "stressmeout"

	TCP_PORT = 1883
	TLS_PORT = 8883
)

// Tests publish flooding
func TestPubFlood(t *testing.T) {

	cfg, err := mqtt.NewTLSAnonymousConfig("../mqtt/files/ca.crt")
	if err != nil {
		t.Errorf("Unexpected failure creating TLS configuration")
	}

	// First create a set of 10 flooders
	flooders := make([]*PublishFlooder, 10)
	topics := make([]string, 10)
	for i := 0; i < 10; i++ {
		topics[i] = randomcreds.RandomTopic("test/")
		client := mqtt.NewMqttClient(HOSTNAME, USERNAME, PASSWORD, TLS_PORT, cfg)
		flooders[i] = NewPublishFlooder(client, 10, 0.005, 50, 5, 0, topics[i])
		for j := i; j < 10; j++ {
			if i != j {
				if topics[i] == topics[j] {
					t.Errorf("Detected duplicate topics %d and %d", j, i)
				}
			}
		}
	}

	// We just created the publishers. Now we want to create subscriptions to
	// each of the topics
	subClients := make([]*mqtt.MqttClient, 10)
	subChans := make([](<-chan []byte), 10)
	for i := 0; i < 10; i++ {
		subClients[i] = mqtt.NewMqttClient(HOSTNAME, USERNAME, PASSWORD, TLS_PORT, cfg)
		subClients[i].Connect(1 * time.Second)
		ch, err := subClients[i].Subscribe(topics[i], 0)
		if err != nil {
			t.Errorf("Subscription to channel %d unexpectedly failed.", i)
		}
		subChans[i] = ch
	}

	// Now go through and set up listeners for each of the subscription channels
	// so we process the messages we're receiving.
	rcvd_msgs := make([][][]byte, 10)
	for i := 0; i < 10; i++ {
		rcvd_msgs[i] = make([][]byte, 0)
		go func(i int) {
			for msg := range subChans[i] {
				rcvd_msgs[i] = append(rcvd_msgs[i], msg)
			}
		}(i)
	}
	// Launch all of the publishers in separate goroutines
	n_messages := make(chan int, 10)
	for _, fld := range flooders {
		tmp := fld
		go func() {
			n_messages <- tmp.PublishFor(3 * time.Second)
		}()
	}

	// Now collect the number of messages printed
	var total_pub int = 0
	var total_sub int = 0
	for i := 0; i < 10; i++ {
		nmsg := <-n_messages
		if nmsg < 20 || nmsg > 40 {
			t.Errorf("Published %d messages. Expected to publish between 20 and 40", nmsg)
		}
		total_pub += nmsg
		total_sub += len(rcvd_msgs[i])
	}
	if total_pub != total_sub {
		t.Errorf("Published %d total messages. Only received %d", total_pub, total_sub)
	}
}

// Tests pub/sub flooding with the same MQTT client
func TestPubSubFlood(t *testing.T) {
	cfg, err := mqtt.NewTLSAnonymousConfig("../mqtt/files/ca.crt")
	if err != nil {
		t.Errorf("Unexpected failure creating TLS configuration")
	}

	// First create a set of 10 flooders
	pubflooders := make([]*PublishFlooder, 10)
	subflooders := make([]*SubscribeFlooder, 10)
	for i := 0; i < 10; i++ {
		topic := randomcreds.RandomTopic("test/")
		client := mqtt.NewMqttClient(HOSTNAME, USERNAME, PASSWORD, TLS_PORT, cfg)
		pf, sf := NewPubSubFlooder(client, 10, 0.005, 50, 5, 0, topic)
		pubflooders[i] = pf
		subflooders[i] = sf
	}

	// Now go through and set up listeners for each of the subscription channels
	// so we process the messages we're receiving.
	rcvd_msgs := make([][][]byte, 10)
	for i := 0; i < 10; i++ {
		rcvd_msgs[i] = make([][]byte, 0)
		go func(i int) {
			for msg := range subflooders[i].SubChan {
				rcvd_msgs[i] = append(rcvd_msgs[i], msg)
			}
		}(i)
	}
	// Launch all of the publishers in separate goroutines
	n_messages := make(chan int, 10)
	for _, fld := range pubflooders {
		tmp := fld
		go func() {
			n_messages <- tmp.PublishFor(3 * time.Second)
		}()
	}

	// Now collect the number of messages printed
	var total_pub int = 0
	var total_sub int = 0
	for i := 0; i < 10; i++ {
		nmsg := <-n_messages
		if nmsg < 20 || nmsg > 40 {
			t.Errorf("Published %d messages. Expected to publish between 20 and 40", nmsg)
		}
		total_pub += nmsg
		total_sub += len(rcvd_msgs[i])
	}
	if total_pub != total_sub {
		t.Errorf("Published %d total messages. Only received %d", total_pub, total_sub)
	}
}
