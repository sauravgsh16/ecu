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
func GenerateHash(ip []byte) string {
	sha512 := sha512.New512_256()
	sha512.Write(ip)
	return string(sha512.Sum(nil))
}

// NextCounter returns a next counter
func NextCounter() int64 {
	return atomic.AddInt64(&counter, 1)
}

// GetHandlerName returns a handler name
func GetHandlerName(handle, id string) string {
	return fmt.Sprintf("%s.%s", handle, id)
}
