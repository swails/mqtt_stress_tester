// +build test
package config

import "testing"

// Test the command-line parser
func TestCommandLineParser(t *testing.T) {
	args := []string{
		"--hostname", "somehost",
		"--passwd-file", "passwd",
		"--username", "username",
		"--password", "password",
		"--port", "1985",
		"--num-publishers", "100",
		"--messages-per-second", "100",
		"--msg-rate-variance", "1.0",
		"--message-size", "100",
		"--msg-size-variance", "10",
		"--duration", "120",
		"--topic-prefix", "sometopic/",
		"--ca-file", "some.crt",
		"--output", "some.csv",
	}
	_, err := Parser.ParseArgs(args)
	if err != nil {
		t.Errorf("Unexpected error parsing arguments: %v", err)
	}
	doCheck(t)
}

// Check short arguments
func TestCommandLineParserShortArgs(t *testing.T) {
	args := []string{
		"--hostname", "somehost",
		"--passwd-file", "passwd",
		"-u", "username",
		"-P", "password",
		"-p", "1985",
		"-n", "100",
		"-m", "100",
		"-v", "1.0",
		"-s", "100",
		"-V", "10",
		"-d", "120",
		"-t", "sometopic/",
		"-c", "some.crt",
		"-o", "some.csv",
	}
	_, err := Parser.ParseArgs(args)
	if err != nil {
		t.Errorf("Unexpected error parsing arguments: %v", err)
	}
	doCheck(t)
}

func doCheck(t *testing.T) {
	if conn.Host != "somehost" {
		t.Errorf("Expected hostname to be somehost, not %s", conn.Host)
	}
	if conn.Passwd != "passwd" {
		t.Errorf("Expected passwd to be passwd, not %s", conn.Passwd)
	}
	if conn.User != "username" {
		t.Errorf("Expected username to be username, not %s", conn.User)
	}
	if conn.Pass != "password" {
		t.Errorf("Expected password to be password, not %s", conn.Pass)
	}
	if conn.Port != 1985 {
		t.Errorf("Expected port to be 1985, not %d", conn.Port)
	}
	if pubsub.Num != 100 {
		t.Errorf("Expected # of publishers to be 100, not %d", pubsub.Num)
	}
	if pubsub.MsgPerSec != 100 {
		t.Errorf("Expected messages per second to be 100, not %d", pubsub.MsgPerSec)
	}
	if pubsub.Duration != 120 {
		t.Errorf("Expected duration to be 120, not %d", pubsub.Duration)
	}
	if pubsub.MsgSize != 100 {
		t.Errorf("Expected message size to be 100, not %d", pubsub.MsgSize)
	}
	if pubsub.MsgRateVar != 1 {
		t.Errorf("Expected message rate variance to be 1, not %g", pubsub.MsgRateVar)
	}
	if pubsub.MsgSizeVar != 10 {
		t.Errorf("Expected message size variance to be 10, not %g", pubsub.MsgSizeVar)
	}
	if pubsub.TopicPfx != "sometopic/" {
		t.Errorf("Expected topic prefix to be sometopic/, not %s", pubsub.TopicPfx)
	}
	if files.CA != "some.crt" {
		t.Errorf("Expected CA file to be some.crt, not %s", files.CA)
	}
	if files.Output != "some.csv" {
		t.Errorf("Expected output file to be some.csv, not %s", files.Output)
	}
	// Reset the variables for the next test
	conn.Host = ""
	conn.Passwd = ""
	conn.User = ""
	conn.Pass = ""
	conn.Port = 0
	pubsub.Num = 0
	pubsub.MsgPerSec = 0
	pubsub.Duration = 0
	pubsub.MsgSize = 0
	pubsub.MsgRateVar = 0
	pubsub.MsgSizeVar = 0
	pubsub.TopicPfx = ""
	files.CA = ""
	files.Output = ""
}

func TestYaml(t *testing.T) {
	args := []string{
		"--yaml", "files/sample.yaml",
	}
	_, err := Parser.ParseArgs(args)
	if err != nil {
		t.Errorf("Unexpected error parsing arguments: %v", err)
	}
	doCheck(t)
}
