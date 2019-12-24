package domain

import (
	"crypto/rand"
)

const (
	alphabets = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
)

func generateRandomBytes(n int) ([]byte, error) {
	buf := make([]byte, n)
	_, err := rand.Read(buf)
	if err != nil {
		return nil, err
	}

	return buf, nil
}

// GenerateRandom returns of random strings
func GenerateRandom(n int) (string, error) {
	bytes, err := generateRandomBytes(n)
	if err != nil {
		return "", err
	}

	for i, b := range bytes {
		bytes[i] = alphabets[b%byte(len(alphabets))]
	}

	return string(bytes), nil
}
