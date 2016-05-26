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
	waiter    *sync.WaitGroup
}

// Creates a pointer to a new Killswitch
func NewKillswitch() *Killswitch {
	return &Killswitch{&sync.Mutex{}, false, make(chan struct{}), &sync.WaitGroup{}}
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

func (k *Killswitch) Add() {
	k.waiter.Add(1)
}

func (k *Killswitch) Subtract() {
	k.waiter.Done()
}

func (k *Killswitch) Wait() {
	k.waiter.Wait()
}
