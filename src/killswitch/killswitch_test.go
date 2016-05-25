package killswitch

import (
	"testing"
	"time"
)

func TestKillswitch(t *testing.T) {
	ks := NewKillswitch()
	go func() {
		time.Sleep(5 * time.Millisecond)
		ks.Trigger()
	}()

	i := 0
mainLoop:
	for {
		select {
		case <-time.After(1 * time.Millisecond):
			i++
		case <-ks.Done():
			break mainLoop
		}
	}
	if i < 4 || i > 6 {
		t.Errorf("Incremented %d times, expected between 4 and 6", i)
	}
}
