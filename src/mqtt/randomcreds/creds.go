package randomcreds

import (
	"math/rand"
)

const ALPHABET = `ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789!@$%^&*()_=+,./#`
const TOPIC_ALPHABET = `ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789!@$%^&*()_=+,./`

func RandomUsername() string {
	return generateRandomString(ALPHABET, 20)
}

func RandomPassword() string {
	return generateRandomString(ALPHABET, 30)
}

func RandomTopic() string {
	return generateRandomString(TOPIC_ALPHABET, 40)
}

func generateRandomString(chars string, size int) string {
	msg := make([]byte, size)
	for i := 0; i < size; i++ {
		idx := rand.Intn(len(chars))
		msg[i] = byte(chars[idx])
	}
	return string(msg)
}
