package util

import (
	"crypto/sha512"
	"fmt"
	"math/rand"
	"sync/atomic"
	"time"
)

var (
	counter int64
)

func init() {
	rand.Seed(time.Now().UnixNano())
	counter = time.Now().UnixNano()
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
