//+build !test
package flooding

import (
	"killswitch"
	"mqtt"
	"mqtt/randomcreds"
	"sync"
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
		fl, err := NewPublishFlooder(client, 10, 0.005, 50, 5, 0, topics[i])
		if err != nil {
			t.Errorf("Unexpected error creating publish flooder: %v", err)
		}
		flooders[i] = fl
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

	// Launch all of the publishers in separate goroutines
	ks := killswitch.NewKillswitch()

	// Now go through and set up listeners for each of the subscription channels
	// so we process the messages we're receiving.
	rcvd_msgs := make([][][]byte, 10)
	for i := 0; i < 10; i++ {
		rcvd_msgs[i] = make([][]byte, 0)
		go func(i int) {
		mainLoop:
			for {
				select {
				case msg := <-subChans[i]:
					rcvd_msgs[i] = append(rcvd_msgs[i], msg)
				case <-ks.Done():
					break mainLoop
				}
			}
		}(i)
	}
	go func() {
		time.Sleep(3 * time.Second)
		ks.Trigger()
	}()
	n_messages := make(chan int, 10)
	for _, fld := range flooders {
		tmp := fld
		go func() {
			n_messages <- tmp.Publish(ks, nil)
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
		pf, sf, err := NewPubSubFlooder(client, 10, 0.005, 50, 5, 0, topic)
		if err != nil {
			t.Errorf("Unexpected error creating pub/sub flooders: %v", err)
		}
		pubflooders[i] = pf
		subflooders[i] = sf
	}

	ks := killswitch.NewKillswitch()

	// Now go through and set up listeners for each of the subscription channels
	// so we process the messages we're receiving.
	rcvd_msgs := make([][][]byte, 10)
	for i := 0; i < 10; i++ {
		rcvd_msgs[i] = make([][]byte, 0)
		go func(i int) {
		mainLoop:
			for {
				select {
				case msg := <-subflooders[i].SubChan:
					rcvd_msgs[i] = append(rcvd_msgs[i], msg)
				case <-ks.Done():
					break mainLoop
				}
			}
		}(i)
	}
	// Launch all of the publishers in separate goroutines
	go func() {
		time.Sleep(3 * time.Second)
		ks.Trigger()
	}()
	n_messages := make(chan int, 10)
	for i, fld := range pubflooders {
		go func(i int, fld *PublishFlooder) {
			n_messages <- fld.Publish(ks, nil)
		}(i, fld)
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

// Tests the closing of the subscription flood channel appropriately
func TestSubFloodClose(t *testing.T) {
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
		pf, sf, err := NewPubSubFlooder(client, 10, 0.005, 50, 5, 0, topic)
		if err != nil {
			t.Errorf("Unexpected error creating pub/sub flooders: %v", err)
		}
		pubflooders[i] = pf
		subflooders[i] = sf
	}

	// A sync point
	var wg sync.WaitGroup
	// Launch all of the publishers in separate goroutines
	ks := killswitch.NewKillswitch()

	// Now go through and set up listeners for each of the subscription channels
	// so we process the messages we're receiving.
	rcvd_msgs := make([][][]byte, 10)
	for i := 0; i < 10; i++ {
		rcvd_msgs[i] = make([][]byte, 0)
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
		mainLoop:
			for {
				select {
				case msg := <-subflooders[i].SubChan:
					rcvd_msgs[i] = append(rcvd_msgs[i], msg)
				case <-ks.Done():
					break mainLoop
				}
			}
		}(i)
	}
	go func() {
		time.Sleep(3 * time.Second)
		ks.Trigger()
	}()
	var n_messages []int = make([]int, 10)
	for i, fld := range pubflooders {
		wg.Add(1)
		go func(i int, fld *PublishFlooder) {
			defer wg.Done()
			n_messages[i] = fld.Publish(ks, nil)
		}(i, fld)
	}

	// Make sure all subscription channels were closed and the publishers
	// properly ended
	wg.Wait()

	// Now collect the number of messages printed
	var total_pub int = 0
	var total_sub int = 0
	for i := 0; i < 10; i++ {
		nmsg := n_messages[i]
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

func TestFloodCollection(t *testing.T) {
	ks := killswitch.NewKillswitch()
	coll := NewFlooderCollection(HOSTNAME, USERNAME, PASSWORD, TCP_PORT, nil, 100,
		1*time.Millisecond, 10*time.Millisecond, ks, 10, 0.001, 100, 20, 0)
	// After 10 ms, we should have ~10 attempted connections
	time.Sleep(10 * time.Millisecond)
	nAttempted := coll.NumAttempted()
	if nAttempted < 8 || nAttempted > 12 {
		t.Errorf("Expected between 8 and 12 attempted connections. Got %d", nAttempted)
	}
	time.Sleep(5 * time.Second)
	ks.Trigger()
	if coll.NumAttempted() != 100 {
		t.Errorf("Should have attempted 100 connections by now. Got %d", coll.NumAttempted())
	}
	if coll.NumFailed() > 0 {
		// Make sure none failed
		t.Errorf("Should not have failed any connections (had %d)", coll.NumFailed())
	}
	ks.Wait()
	// Look at the message stats
	if coll.NumSentMessages() < 3000 {
		t.Errorf("Expected at least 3K messages. Got %d", coll.NumSentMessages())
	}
	rm := len(coll.MessageTimings())
	sm := coll.NumSentMessages()
	if rm > sm {
		t.Errorf("Received more messages than sent (%d vs %d)", rm, sm)
	}
	if sm-rm > 5 {
		t.Errorf("Received messages (%d) not close enough to # of sent messages (%d)", rm, sm)
	}
}
