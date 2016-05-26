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

func TestKillswitchWait(t *testing.T) {
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
			ks.Add()
			i++
		case <-ks.Done():
			break mainLoop
		}
	}
	if i < 4 || i > 6 {
		t.Errorf("Incremented %d times, expected between 4 and 6", i)
	}
	var j int
	go func(n int) {
		for i := 0; i < n; i++ {
			time.Sleep(1 * time.Millisecond)
			j += 1
			ks.Subtract()
		}
	}(i)
	ks.Wait()
	if i != j {
		t.Errorf("i (%d) != j (%d)", i, j)
	}
}
