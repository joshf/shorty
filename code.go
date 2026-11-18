package main

import "crypto/rand"

// no 0/O or 1/l/I so the codes are easier to read
const alphabet = "abcdefghijkmnpqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ23456789"

func generateCode(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	for i := range b {
		// modulo is a tiny bit biased but doesnt matter for this
		b[i] = alphabet[int(b[i])%len(alphabet)]
	}
	return string(b), nil
}
