package util

import (
	"crypto/sha512"
	"fmt"
	"sync/atomic"
)

var (
	counter int64
)

func init() {
	counter = 40
}

// GenerateHash returns a sha512/256 hashes string
func GenerateHash(input []byte) string {
	sha512 := sha512.New512_256()
	sha512.Write(input)
	return string(sha512.Sum(nil))
}

// NextCounter returns a next counter
func NextCounter() int64 {
	return atomic.AddInt64(&counter, 1)
}

// JoinString returns a handler name
func JoinString(handle, id string) string {
	return fmt.Sprintf("%s.%s", handle, id)
}
