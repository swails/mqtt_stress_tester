// The main package for driving MQTT stress testing
package main

import (
	"config"
	"crypto/tls"
	"flooding"
	"fmt"
	"messages"
	"mqtt"
	"mqtt/randomcreds"
	"os"
	"time"
)

// Processes a message, returning how long between publish and reception and the
// size in bytes of the message
func processMessage(msg []byte) (time.Duration, int) {
	now := time.Duration(time.Now().UnixNano())
	nbytes := len(msg)
	msgTime := messages.ExtractTimeFromMessage(msg)
	return (now - msgTime), nbytes
}

func main() {
	_, err := config.Parser.Parse()
	if err != nil {
		os.Exit(1)
	}

	// Get the output file and print the input parameters to it
	var output *os.File
	if config.OutputFile() == "stdout" {
		output = os.Stdout
	} else {
		output, err = os.Create(config.OutputFile())
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	}
	config.Echo(output)

	// Set up the publish/subscribe pool with an MqttConnection
	publishers := make([]*flooding.PublishFlooder, config.NumPublishers())
	subscribers := make([]*flooding.SubscribeFlooder, config.NumPublishers())
	for i := 0; i < config.NumPublishers(); i++ {
		topic := randomcreds.RandomTopic(config.TopicPrefix())
		var cfg *tls.Config
		if len(config.CertificateAuthority()) > 0 {
			cfg, err = mqtt.NewTLSAnonymousConfig(config.CertificateAuthority())
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v", err)
				os.Exit(1)
			}
		}
		c := mqtt.NewMqttClient(config.Hostname(), config.Username(), config.Password(), config.Port(), cfg)
		pf, sf := flooding.NewPubSubFlooder(c, config.MessagesPerSecond(), config.MessageRateVariance(),
			config.MessageSize(), config.MessageSizeVariance(), 0, topic)
		publishers[i] = pf
		subscribers[i] = sf
	}

	// To keep track of statistics for each publish flooder. But only keep running stats to avoid running out of memory
	// if we do *lots* of pubs and subs. Do both measurements and squares of measurements so we can get variances
	msgTimes := make([]float64, config.NumPublishers())
	msgTimesSquared := make([]float64, config.NumPublishers())
	msgSizes := make([]int, config.NumPublishers())
	msgSizesSquared := make([]int, config.NumPublishers())
	msgCnt := make([]int, config.NumPublishers())

	// To check that the publish and subscribe count match, and/or report differences
	msgCntPub := make([]int, config.NumPublishers())

	// Spin up goroutines for each subcription channel to process all of the messages as they come in
	for i, sub := range subscribers {
		ch := sub.SubChan
		// TODO: add synchronization
		go func(i int) {
			for msg := range ch {
				elapsed, size := processMessage(msg)
				tm := float64(elapsed.Nanoseconds()) * 1e-9
				msgTimes[i] += tm
				msgSizes[i] += size
				msgTimesSquared[i] += tm * tm
				msgSizesSquared[i] += size * size
				msgCnt[i]++
			}
		}(i)
	}

	// Now spin up goroutines to have each publish flooder start flooding messages
	for i, pub := range publishers {
		go func(i int, pub *flooding.PublishFlooder) {
			msgCntPub[i] = pub.PublishFor(time.Duration(int64(config.MessagesPerSecond())) * time.Second)
		}(i, pub)
	}
}
