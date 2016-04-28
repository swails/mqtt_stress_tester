// Configuration package for reading and processing user input
package config

import (
	"fmt"
	"io"

	flags "github.com/jessevdk/go-flags"
)

var Parser *flags.Parser

var conn struct {
	Host   string `long:"hostname" description:"Address of the broker to connect to" default:"localhost"`
	Passwd string `long:"passwd-file" description:"File with raw-text usernames and passwords"`
	User   string `short:"u" long:"username" description:"Name of the user to connect with. Superceded by --passwd-file if specified"`
	Pass   string `short:"P" long:"password" description:"Password of the user to connect with. Used in tandem with username"`
	Port   int    `short:"p" long:"port" default:"1883" description:"The port to connect through"`
}

var pubsub struct {
	Num        int     `short:"n" long:"num-publishers" default:"1" description:"Number of concurrent publishers"`
	MsgPerSec  int     `short:"m" long:"messages-per-second" default:"10" description:"Average number of messages per second to send from each publisher"`
	Duration   int     `short:"d" long:"duration" default:"5" description:"Number of seconds to run the publishers for"`
	MsgSize    int     `short:"s" long:"message-size" default:"50" description:"Average number of bytes per message. At least 8 needed to collect timing data"`
	MsgRateVar float64 `short:"v" long:"msg-rate-variance" default:"0.005" description:"Variance (seconds squared) of the sample of message rates"`
	MsgSizeVar float64 `short:"V" long:"msg-size-variance" default:"5" description:"Variance (messages squared) of the sample of message sizes"`
	TopicPfx   string  `short:"t" long:"topic-prefix" default:"test/" description:"Prefix to add to all random topic names for each publisher"`
}

var files struct {
	CA     string       `short:"c" long:"ca-file" description:"Certificate authority to enable anonymous TLS connection"`
	Output string       `short:"o" long:"output" default:"stdout" description:"Output file to write detailed pub/sub statistics to"`
	Yaml   func(string) `short:"y" long:"yaml" description:"Input file with command-line parameters in YAML format. CL options appearing before are overridden. Those appearing after override."`
}

func init() {
	files.Yaml = processYaml
	// Add the connection options group
	Parser = flags.NewParser(nil, flags.Default)
	_, err := Parser.AddGroup("Broker Connection Options", "Options controlling how the connection to the broker is made.", &conn)
	if err != nil {
		panic("Error adding group to parser: " + err.Error())
	}
	// Add the publish statistics group
	_, err = Parser.AddGroup("Publish/Subscribe Options", "Options controlling the publish/subscribe load to the broker", &pubsub)
	if err != nil {
		panic("Error adding group to parser: " + err.Error())
	}
	// Add the file options
	_, err = Parser.AddGroup("Input/Output Files", "Options specifying input and output files", &files)
	if err != nil {
		panic("Error adding group to parser: " + err.Error())
	}
}

// Returns the hostname of the broker
func Hostname() string {
	return conn.Host
}

// Returns the name of the passwd file containing raw text usernames and passwords
func Passwd() string {
	return conn.Passwd
}

// Name of a single user (superceded by Passwd() file)
func Username() string {
	return conn.User
}

// The password for the single user (superceded by Passwd() file)
func Password() string {
	return conn.Pass
}

// The port through which the connection will take place
func Port() int {
	return conn.Port
}

// The number of publishers to launch
func NumPublishers() int {
	return pubsub.Num
}

// The number of messages per second to run on average
func MessagesPerSecond() int {
	return pubsub.MsgPerSec
}

// The duration of the publishing barrage in seconds
func PublishDuration() int {
	return pubsub.Duration
}

// The average size of the messages in bytes
func MessageSize() int {
	return pubsub.MsgSize
}

// The variance in the rate of publishing in seconds^2
func MessageRateVariance() float64 {
	return pubsub.MsgRateVar
}

// The variance in the size of published messages in bytes^2
func MessageSizeVariance() float64 {
	return pubsub.MsgRateVar
}

// The prefix for all randomly generated topic names
func TopicPrefix() string {
	return pubsub.TopicPfx
}

// The output file where the statistics are written
func OutputFile() string {
	return files.Output
}

// The certificate authority file to use for TLS encryption
func CertificateAuthority() string {
	return files.CA
}

// Echoes all input variables to the designated file
func Echo(w io.Writer) {
	fmt.Fprintf(w, "Broker Hostname: %s\n", conn.Host)
	if len(conn.Passwd) > 0 {
		fmt.Fprintf(w, "Passwd File:     %s\n", conn.Passwd)
	} else if len(conn.User) > 0 {
		fmt.Fprintf(w, "Username:        %s\n", conn.User)
		if len(conn.Pass) > 0 {
			fmt.Fprintf(w, "Password:        %s\n", conn.User)
		}
	}
	if len(files.CA) > 0 {
		fmt.Fprintf(w, "TLS Certificate: %s\n", files.CA)
	}
	fmt.Fprintf(w, "Port:            %d\n", conn.Port)
	fmt.Fprintf(w, "# of publishers: %d\n", pubsub.Num)
	fmt.Fprintf(w, "Message Rate:    %d per second\n", pubsub.MsgPerSec)
	fmt.Fprintf(w, "Variance (MpS):  %f\n", pubsub.MsgRateVar)
	fmt.Fprintf(w, "Message Size:    %d bytes\n", pubsub.MsgSize)
	fmt.Fprintf(w, "Variance (size): %f\n", pubsub.MsgSizeVar)
	fmt.Fprintf(w, "Topic prefix:    %s\n", pubsub.TopicPfx)
}
