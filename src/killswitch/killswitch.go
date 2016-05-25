// Implements a killswitch that you can pass to goroutines to trigger the end of the program
package killswitch

import (
	"sync"
)

// Killswitch to tell goroutines when program execution should stop
type Killswitch struct {
	mux       *sync.Mutex
	triggered bool
	done      chan struct{}
}

// Creates a pointer to a new Killswitch
func NewKillswitch() *Killswitch {
	return &Killswitch{&sync.Mutex{}, false, make(chan struct{})}
}

// Returns a read-only channel that is closed when the killswitch is triggered
func (k *Killswitch) Done() <-chan struct{} {
	return k.done
}

// Triggers the killswitch to close the channel returned by Done. If it has already been triggered, nothing happens
func (k *Killswitch) Trigger() {
	k.mux.Lock()
	defer k.mux.Unlock()
	if !k.triggered {
		close(k.done)
		k.triggered = true
	}
}
