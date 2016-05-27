package config

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/go-yaml/yaml"
)

var yamlOptions struct {
	Hostname   string
	Passwd     string `yaml:"passwd-file"`
	Username   string
	Password   string
	Port       int
	Num        int `yaml:"num-publishers"`
	MsgPerSec  int `yaml:"messages-per-second"`
	Duration   int
	MsgSize    int     `yaml:"message-size"`
	MsgRateVar float64 `yaml:"msg-rate-variance"`
	MsgSizeVar float64 `yaml:"msg-size-variance"`
	CnctIntvl  float64 `yaml:"interval"`
	TopicPfx   string  `yaml:"topic-prefix"`
	CA         string  `yaml:"ca-file"`
	Output     string
}

func processYaml(fname string) error {
	contents, err := ioutil.ReadFile(fname)
	if err != nil {
		panic("Error reading YAML file: " + err.Error())
	}
	fmt.Fprintf(os.Stderr, "Processing YAML input file %s\n", fname)
	// Load the contents of the YAML file into the struct
	return yaml.Unmarshal(contents, &yamlOptions)
}
