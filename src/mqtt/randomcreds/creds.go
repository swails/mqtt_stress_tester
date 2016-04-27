package randomcreds

import (
	"math/rand"
)

const ALPHABET = `ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789_`

// Generates a random user name with 20 characters
func RandomUsername() string {
	return generateRandomString(ALPHABET, 20)
}

// Generates a random password with 30 characters
func RandomPassword() string {
	return generateRandomString(ALPHABET, 30)
}

// Generates a random topic starting with prefix with a total of 40 characters.
// If prefix is more than 40 characters, it panics
func RandomTopic(prefix string) string {
	if len(prefix) > 40 {
		panic("prefix must be <= 40 characters")
	}
	return (prefix + generateRandomString(ALPHABET, 40-len(prefix)))
}

// Private helper function to build random strings
func generateRandomString(chars string, size int) string {
	msg := make([]byte, size)
	for i := 0; i < size; i++ {
		idx := rand.Intn(len(chars))
		msg[i] = byte(chars[idx])
	}
	return string(msg)
}
