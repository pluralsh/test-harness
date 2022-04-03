package utils

import (
	"math/rand"
	"time"
)

const letters = "abcdefghijklmnopqrstuvwxyz"

func RandomStr(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func init() {
	rand.Seed(time.Now().UnixNano())
}
