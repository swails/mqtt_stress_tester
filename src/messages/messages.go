// Creates random messages conforming to a specific size with a specified spread
// in actual message size

package messages

import (
	"encoding/binary"
	"fmt"
	"killswitch"
	"math"
	"math/rand"
	"time"
)

const MESSAGE_CHARACTERS = `ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789_!@#$%^&*(),./;'"`

func init() {
	// Set a random seed to avoid duplicating messages
	rand.Seed(time.Now().UnixNano())
}

// Generates a random message that is a particular size with the time at which
// the message was generated encoded into the first 8 bytes (as long as
// size > 8, otherwise no time is encoded)
func GenerateRandomMessage(size int) []byte {
	msg := make([]byte, size)
	var start int = 0
	for i := start; i < size; i++ {
		idx := rand.Intn(len(MESSAGE_CHARACTERS))
		msg[i] = byte(MESSAGE_CHARACTERS[idx])
	}
	if size >= 8 {
		start = 8
		int64ToBytes(time.Now().UnixNano(), msg)
	}
	return msg
}

// Generate a list of "num" random messages, with an average length of "size"
// and a variance in the sizes of "variance"
func GenerateRandomMessages(ks *killswitch.Killswitch, size int, variance float64) <-chan []byte {
	msgs := make(chan []byte)
	fac := math.Sqrt(variance)
	// Fire off a goroutine to populate the messages channel. Make it blocking so
	// that the time is always up-to-date.
	go func() {
	mainLoop:
		for {
			msg_size := int(math.Floor(rand.NormFloat64()*fac)) + size
			if size >= 8 {
				msg_size = int(math.Max(float64(msg_size), 8))
			}
			select {
			case <-ks.Done():
				break mainLoop
			default:
				msgs <- GenerateRandomMessage(msg_size)
			}
		}
		close(msgs)
	}()
	return msgs
}

// Extracts the time from the message's first 8 bytes. If there are fewer than 8
// bytes, a Duration of 0 is returned.
func ExtractTimeFromMessage(msg []byte) time.Duration {
	return time.Duration(bytesToInt64(msg)) * time.Nanosecond
}

// Convert a 64-bit integer to bytes (unsigned int!)
func int64ToBytes(integer int64, inp []byte) error {
	if cap(inp) < 8 {
		return fmt.Errorf("byte array too small to hold int64")
	}
	bytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(bytes, uint64(integer))
	for i := 0; i < 8; i++ {
		inp[i] = bytes[i]
	}
	return nil
}

func bytesToInt64(inp []byte) int64 {
	if len(inp) < 8 {
		return 0
	}
	bytes := inp[:8]
	return int64(binary.LittleEndian.Uint64(bytes))
}
