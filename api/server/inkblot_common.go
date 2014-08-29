package main

import (
	"log"
	"math/rand"
)

var alpha = "abcdefghijkmnpqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ23456789"

// generates a random string of fixed size
func srand(size int) string {
	buf := make([]byte, size)
	for i := 0; i < size; i++ {
		buf[i] = alpha[rand.Intn(len(alpha))]
	}
	return string(buf)
}

// GetNewId returns random 5 character string.
// Random characters are [a-zA-Z0-9] excluding those with visible similarity = l,1,O,0.
// This gives 550 731 776 unique id's.
func GetNewId() string {
	return srand(5)
}

func info(template string, values ...interface{}) {
	log.Printf("[inkblot] "+template+"\n", values...)
}
