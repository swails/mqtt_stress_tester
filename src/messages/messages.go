// Creates random messages conforming to a specific size with a specified spread
// in actual message size

package messages

import (
	"math"
	"math/rand"
)

const MESSAGE_CHARACTERS = `ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789_!@#$%^&*(),./;'"`

// Generates a random message that is a particular size
func GenerateRandomMessage(size int) []byte {
	msg := make([]byte, size)
	for i := 0; i < size; i++ {
		idx := rand.Intn(len(MESSAGE_CHARACTERS))
		msg[i] = byte(MESSAGE_CHARACTERS[idx])
	}
	return msg
}

// Generate a list of "num" random messages, with an average length of "size"
// and a variance in the sizes of "variance"
func GenerateRandomMessages(num, size int, variance float64) [][]byte {
	msgs := make([][]byte, num)
	fac := math.Sqrt(variance * float64(num) / (float64(num) - 1))
	for i := 0; i < num; i++ {
		msg_size := int(math.Floor(rand.NormFloat64()*fac)) + size
		msgs[i] = GenerateRandomMessage(msg_size)
	}
	return msgs
}
