package randomcreds

import (
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
		t.Errorf("Invalid username. Expected non-zero length")
	}
}

func TestRandomTopic(t *testing.T) {
	topic := RandomTopic()

	if len(topic) <= 0 {
		t.Errorf("Invalid username. Expected non-zero length")
	}

	for _, c := range topic {
		if c == '#' {
			t.Errorf("# not allowed in any topic")
		}
	}
}
