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
	forever := make(chan interface{})
	c, err := can.New()
	if err != nil {
		log.Fatalf(err.Error())
	}
	go c.Read()

	i := 0
	for i < 10 {
		c.Out <- &can.Message{
			ArbitrationID: []uint8{0x1c, 0xeb, 0xf7, 0xe8},
			Priority:      0x00,
			PGN:           []uint8{0xf7, 0x0},
			Src:           0xe8, Dst: 0xf7,
			Size: 0x8,
			Data: []uint8{0x5, 0x31, 0x2e, 0x30, 0x31, 0xa2, 0xff, 0xbf},
		}
		i++
	}

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
	/*
		//b := []byte{0x16, 0x10, 0x53, 0x69, 0x65, 0x49, 0xDE, 0x4C, 0xA0, 0x26, 0x5D, 0x77, 0xEB, 0x62, 0x8A, 0x36, 0xE4, 0x1C, 0x96, 0x60, 0xB0, 0x47, 0x14, 0xCD, 0xF7, 0x24, 0x9F, 0xF, 0xD6, 0xE6, 0xC5, 0x55}
		b := []byte("xtd 02 1CECF7E8 08 10 23 05 05 FF 00 CB 00")

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
	*/
	// xtd 02 1CEBF7E8 08 05 31 2E 30 31 A2 FF BF\n
	/*
			m := &can.Message{
				ArbitrationID: []uint8{0x1c, 0xeb, 0xf7, 0xe8},
				Priority:      0x00,
				PGN:           []uint8{0xf7, 0x0},
				Src:           0xe8, Dst: 0xf7,
				Size: 0x8,
				Data: []uint8{0x5, 0x31, 0x2e, 0x30, 0x31, 0xa2, 0xff, 0xbf},
			}
			b := make([]byte, 0)
			b = append(b, m.ArbitrationID...)
			b = append(b, m.Data...)


		b := []byte{0x1c, 0xeb, 0xf7, 0xe8, 0x5, 0x31, 0x2e, 0x30, 0x31, 0xa2, 0xff, 0xbf, 0x5, 0x31, 0x2e, 0x30, 0x31, 0xa2, 0xff, 0xbf}

		var c int
		var common string
		for c < 3 {
			switch {
			case c < 1:
				common += "xtd "
				c++
			case c < 2:
				common += string(encode(0x02))
				common += " "
				c++
			case c < 3:
				for _, p := range b[:4] {
					common += string(encode(p))
				}
				c++
				common += " "
				b = b[4:]
			}

		}
		// []byte{0x1c, 0xeb, 0xf7, 0xe8, 0x8, 0x5, 0x31, 0x2e, 0x30, 0x31, 0xa2, 0xff, 0xbf}
		var s string = common
		var chunk int = 8

		fmt.Printf("%#v\n", b)

		for len(b) > 0 {
			if len(b)%chunk != 0 && len(b) < chunk {
				chunk = len(b) % chunk
			}

			s += fmt.Sprintf("%02d ", chunk)

			for i := 0; i < chunk; i++ {
				s += string(encode(b[i]))
				s += " "
			}
			s += "\n"
			/*
				w, err := e.Write([]byte(s))
				if err != nil {
					e.err = err
				}
				n += w / 2

			fmt.Printf("%s\n", s)

			b = b[chunk:]
			s = common
		}
	*/
}

const hexchars = "0123456789abcdef"

func encode(src byte) []byte {
	dst := make([]byte, 2)
	dst[0] = hexchars[src>>4]
	dst[1] = hexchars[src&0x0F]
	return dst
}
