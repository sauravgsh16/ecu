package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"

	"github.com/sauravgsh16/ecu/can"
)

type Message struct {
	ArbitrationID int
	PduFormat     int
	PduSpecific   int
	Priority      int64
	PGN           string
	Src           int64
	Dst           int64
	Data          []byte
	TimeStamp     int64
}

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

func main() {

	c, err := can.Dial("tcp://localhost:19000")
	if err != nil {
		log.Fatalf(err.Error())
	}

	forever := make(chan interface{})

	go c.HandleIncoming(c.Conn)

	<-forever
	/*
		m := &Message{
			Priority: int64(12),
			PGN:      "CB00",
			Dst:      int64(247),
			Src:      int64(130),
			Data:     []uint8{0x31, 0x33, 0x20, 0x30, 0x30, 0x20, 0x38, 0x36, 0x20, 0x30, 0x30, 0x20, 0x42, 0x38, 0x20, 0x30, 0x42, 0x20, 0x30, 0x30, 0x20, 0x30, 0x30, 0xa},
		}

		arb := fmt.Sprintf("%02X%s%02X%02X", m.Priority, m.PGN[:2], m.Dst, m.Src)
		fmt.Printf("%s\n", arb)

		fmt.Printf("%02d\n", len(strings.Split(string(m.Data), " ")))

		s, _ := GenerateRandom(16)

		fmt.Printf("%s\n", s)
		fmt.Printf("%#v\n", []byte(s))

		s = "13 00 86 00 B8 0B 00 00"
		// b := make([]byte, 0)
		s = strings.ReplaceAll(s, " ", "")
		fmt.Printf("%s\n", s)

		h, _ := hex.DecodeString(s)
		fmt.Printf("%#v\n", h)
	*/

	//b := []byte{0x16, 0x10, 0x53, 0x69, 0x65, 0x49, 0xDE, 0x4C, 0xA0, 0x26, 0x5D, 0x77, 0xEB, 0x62, 0x8A, 0x36, 0xE4, 0x1C, 0x96, 0x60, 0xB0, 0x47, 0x14, 0xCD, 0xF7, 0x24, 0x9F, 0xF, 0xD6, 0xE6, 0xC5, 0x55}
	b := []byte("A long test string")

	buf := make([]byte, hex.EncodedLen(len(b)))
	hex.Encode(buf, b)
	fmt.Printf("%+v\n", b)
	fmt.Printf("%+v\n", buf)
	l0 := len(buf)
	l1 := len(fmt.Sprintf("%#v\n", buf))
	l2 := len(fmt.Sprintf("%+v\n", buf))
	fmt.Printf("%d %d %d\n", l0, l1, l2)

	n := make([]byte, len(buf))
	hex.Decode(n, buf)
	fmt.Printf("%s\n", n)

	table := "0123456789abcdef"
	v := 65
	i := v >> 4
	fmt.Printf("%+v\n", table[i])
}
