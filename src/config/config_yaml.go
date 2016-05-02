package config

import (
	"io/ioutil"

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
	TopicPfx   string  `yaml:"topic-prefix"`
	CA         string  `yaml:"ca-file"`
	Output     string
}

func processYaml(fname string) error {
	contents, err := ioutil.ReadFile(fname)
	if err != nil {
		panic("Error reading YAML file: " + err.Error())
	}
	// Load the contents of the YAML file into the struct
	err = yaml.Unmarshal(contents, &yamlOptions)
	// Transfer YAML input to the structs
	if len(yamlOptions.Hostname) > 0 {
		conn.Host = yamlOptions.Hostname
	}
	if len(yamlOptions.Passwd) > 0 {
		conn.Passwd = yamlOptions.Passwd
	}
	if len(yamlOptions.Username) > 0 {
		conn.User = yamlOptions.Username
	}
	if len(yamlOptions.Password) > 0 {
		conn.Pass = yamlOptions.Password
	}
	if yamlOptions.Port > 0 {
		conn.Port = yamlOptions.Port
	}
	if yamlOptions.Num > 0 {
		pubsub.Num = yamlOptions.Num
	}
	if yamlOptions.MsgPerSec > 0 {
		pubsub.MsgPerSec = yamlOptions.MsgPerSec
	}
	if yamlOptions.MsgSize > 0 {
		pubsub.MsgSize = yamlOptions.MsgSize
	}
	if yamlOptions.MsgRateVar > 0 {
		pubsub.MsgRateVar = yamlOptions.MsgRateVar
	}
	if yamlOptions.MsgSizeVar > 0 {
		pubsub.MsgSizeVar = yamlOptions.MsgSizeVar
	}
	if len(yamlOptions.TopicPfx) > 0 {
		pubsub.TopicPfx = yamlOptions.TopicPfx
	}
	if len(yamlOptions.CA) > 0 {
		files.CA = yamlOptions.CA
	}
	if len(yamlOptions.Output) > 0 {
		files.Output = yamlOptions.Output
	}
	if yamlOptions.Duration > 0 {
		pubsub.Duration = yamlOptions.Duration
	}
	return nil
}
