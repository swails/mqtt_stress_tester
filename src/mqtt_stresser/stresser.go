// The main package for driving MQTT stress testing
package main

import (
	"config"
	// "fmt"
	"os"
)

func main() {
	_, err := config.Parser.Parse()
	if err != nil {
		os.Exit(1)
	}
	config.Echo(os.Stdout)
}
