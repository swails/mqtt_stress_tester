// The main package for driving MQTT stress testing
package main

import (
	"config"
	"crypto/tls"
	"flooding"
	"fmt"
	"killswitch"
	"messages"
	"mqtt"
	"os"
	"time"

	"github.com/montanaflynn/stats"
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
	err := config.ParseCommandLine(os.Args)
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

	// Echo our parameters before we connect so we can see what the settings were in case we have a problem
	config.Echo(os.Stdout)

	// Set up TLS configuration if we have it
	var cfg *tls.Config
	if len(config.CertificateAuthority()) > 0 {
		cfg, err = mqtt.NewTLSAnonymousConfig(config.CertificateAuthority())
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	}

	// Killswitch to bring down the flooders smoothly
	ks := killswitch.NewKillswitch()

	// Set up the publish/subscribe pool with an MqttConnection
	fc := flooding.NewFlooderCollection(config.Hostname(), config.Username(), config.Password(), config.Port(), cfg,
		config.NumPublishers(), config.ConnectInterval(), 10*time.Second, ks, config.MessagesPerSecond(), config.MessageRateVariance(),
		config.MessageSize(), config.MessageSizeVariance(), 0)

	// Run the connections as long as we asked for
	go func() {
		time.Sleep(time.Duration(config.PublishDuration()) * time.Second)
		ks.Trigger()
	}()

	// block until the trigger is done
	<-ks.Done()
	ks.Wait()

	// We've finished flooding our broker with connections and messages. Collect the stats and check it out
	nAttempted := fc.NumAttempted()
	nFailed := fc.NumFailed()
	nMsg := fc.NumSentMessages()
	msgTimings := fc.MessageTimings()
	// Figure out how many messages we've received based on how many records of messages we have
	nRcv := len(msgTimings)

	fmt.Fprintf(output, "Created a total of %d (out of %d attempted) connections to the MQTT broker\n", nAttempted-nFailed, nAttempted)
	fmt.Fprintf(output, "Sent a total of %d messages.\n", nMsg)
	fmt.Fprintf(output, "Received a total of %d messages.\n", nRcv)
	timings := stats.Float64Data(msgTimings)
	mean, err := timings.Mean()
	if err != nil {
		fmt.Fprintf(output, "ERROR computing average message latency!\n")
	} else {
		fmt.Fprintf(output, "Average latency   = %10.6f ms\n", mean*1e3)
	}
	median, err := timings.Median()
	if err != nil {
		fmt.Fprintf(output, "ERROR computing median message latency!\n")
	} else {
		fmt.Fprintf(output, "Median latency    = %10.6f ms\n", median*1e3)
	}
	std, err := timings.StandardDeviation()
	if err != nil {
		fmt.Fprintf(output, "ERROR computing standard deviation of message latencies!\n")
	} else {
		fmt.Fprintf(output, "Std. Dev. latency = %10.6f ms\n", std)
	}
}
