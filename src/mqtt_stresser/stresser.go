// The main package for driving MQTT stress testing
package main

import (
	"config"
	"fmt"
	"os"
)

func main() {
	stuff, err := config.Parser.Parse()
	if err != nil {
		os.Exit(1)
	}
	fmt.Println("Hostname is " + config.Host())
	fmt.Printf("And stuff is %s\n", stuff)
	os.Exit(0)
}
