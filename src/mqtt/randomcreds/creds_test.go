package randomcreds

import (
	"strings"
	"testing"
)

func TestRandomUsername(t *testing.T) {
	username := RandomUsername()

	if len(username) <= 0 {
		t.Errorf("Invalid username. Expected non-zero length")
	}
}

func TestRandomPassword(t *testing.T) {
	username := RandomPassword()

	if len(username) <= 0 {
		t.Errorf("Invalid password. Expected non-zero length")
	}
}

func TestRandomTopic(t *testing.T) {
	topic := RandomTopic("")

	if len(topic) <= 0 {
		t.Errorf("Invalid topic. Expected non-zero length")
	}

	for _, c := range topic {
		if c == '#' {
			t.Errorf("# not allowed in any topic")
		}
	}

	topic2 := RandomTopic("test/")

	if len(topic) != len(topic2) {
		t.Errorf("topics should always be the same length")
	}

	if strings.Index(topic2, "test/") != 0 {
		t.Errorf("Expected topic (%s) to start with test/", topic)
	}
}
