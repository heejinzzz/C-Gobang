package sessionIdGenerator

import (
	"math/rand"
	"time"
)

const dictionary = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"

func NewSessionId(length int) string {
	rand.Seed(time.Now().UnixNano())
	sessionId := make([]byte, length)
	for i := range sessionId {
		sessionId[i] = dictionary[rand.Intn(len(dictionary))]
	}
	return string(sessionId)
}
