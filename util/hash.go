package util

import (
	"crypto/sha512"
)

// GenerateHash returns a sha512/256 hashes string
func GenerateHash(ip []byte) string {
	sha512 := sha512.New512_256()
	sha512.Write(ip)
	return string(sha512.Sum(nil))
}
