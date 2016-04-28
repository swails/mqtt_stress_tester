// The main package for driving MQTT stress testing
package main

import (
	"config"
	"crypto/tls"
	"flooding"
	"fmt"
	"math"
	"messages"
	"mqtt"
	"mqtt/randomcreds"
	"os"
	"sync"
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

	// Set up the publish/subscribe pool with an MqttConnection
	publishers := make([]*flooding.PublishFlooder, config.NumPublishers())
	subscribers := make([]*flooding.SubscribeFlooder, config.NumPublishers())
	for i := 0; i < config.NumPublishers(); i++ {
		topic := randomcreds.RandomTopic(config.TopicPrefix())
		var cfg *tls.Config
		if len(config.CertificateAuthority()) > 0 {
			cfg, err = mqtt.NewTLSAnonymousConfig(config.CertificateAuthority())
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
		}
		c := mqtt.NewMqttClient(config.Hostname(), config.Username(), config.Password(), config.Port(), cfg)
		pf, sf, err := flooding.NewPubSubFlooder(c, config.MessagesPerSecond(), config.MessageRateVariance(),
			config.MessageSize(), config.MessageSizeVariance(), 0, topic)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		publishers[i] = pf
		subscribers[i] = sf
	}

	// We are connected to the broker, go ahead and echo our parameters
	config.Echo(os.Stdout)

	// To keep track of statistics for each publish flooder. But only keep running stats to avoid running out of memory
	// if we do *lots* of pubs and subs. Do both measurements and squares of measurements so we can get variances
	msgTimes := make([]float64, config.NumPublishers())
	msgTimesSquared := make([]float64, config.NumPublishers())
	msgSizes := make([]int, config.NumPublishers())
	msgSizesSquared := make([]int, config.NumPublishers())
	msgCnt := make([]int, config.NumPublishers())

	// To check that the publish and subscribe count match, and/or report differences
	msgCntPub := make([]int, config.NumPublishers())

	// To make sure we let the pub/sub go until it's done
	var wg sync.WaitGroup

	// Spin up goroutines for each subcription channel to process all of the messages as they come in
	for i, sub := range subscribers {
		wg.Add(1)
		go func(i int, ch <-chan []byte) {
			defer wg.Done()
			for msg := range ch {
				elapsed, size := processMessage(msg)
				tm := float64(elapsed.Nanoseconds()) * 1e-9
				msgTimes[i] += tm
				msgSizes[i] += size
				msgTimesSquared[i] += tm * tm
				msgSizesSquared[i] += size * size
				msgCnt[i]++
			}
		}(i, sub.SubChan)
	}

	// Now spin up goroutines to have each publish flooder start flooding messages
	for i, pub := range publishers {
		wg.Add(1)
		go func(i int, pub *flooding.PublishFlooder) {
			defer wg.Done()
			msgCntPub[i] = pub.PublishFor(time.Duration(int64(config.PublishDuration()))*time.Second, func() {
				subscribers[i].Complete(1 * time.Second)
			})
		}(i, pub)
	}

	// Sync point. We can't do anything with the stats we collected until all the flooders are done
	wg.Wait()

	// We have all of our stats. Go ahead and compute averages and stuff now
	fmt.Fprintf(output, "Success rate,Transit Time (s),,Message size,\n")
	for i := 0; i < config.NumPublishers(); i++ {
		subcnt := float64(msgCnt[i])
		fmt.Printf("Messages (Sub, Pub): %d, %d\n", msgCnt[i], msgCntPub[i])
		suc := subcnt / float64(msgCntPub[i]) * 100
		avgTime := msgTimes[i] / subcnt
		stdTime := math.Sqrt(math.Abs(msgTimesSquared[i]/subcnt - avgTime*avgTime))
		avgSize := float64(msgSizes[i]) / subcnt
		stdSize := math.Sqrt(math.Abs(float64(msgSizesSquared[i])/subcnt - avgSize*avgSize))
		fmt.Fprintf(output, "%.2f%%,%g,%g,%g,%g\n", suc, avgTime, stdTime, avgSize, stdSize)
	}
}
