package main

import (
	"crypto/rand"
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
}
