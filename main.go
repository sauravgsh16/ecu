package main

import "fmt"

const (
	hexchars = "0123456789abcdef"
)

func encode(src byte) []byte {
	dst := make([]byte, 2)
	dst[0] = hexchars[src>>4]
	dst[1] = hexchars[src&0x0F]
	return dst
}

func main() {
	sb := encode(byte(255))
	a, ok := fromHexChar(sb[0])
	if !ok {
		fmt.Println("Error")
	}
	b, ok := fromHexChar(sb[1])
	if !ok {
		fmt.Println("Error")
	}

	h := a<<4 | b
	fmt.Printf("%#v\n", h)

}

func fromHexChar(ch byte) (byte, bool) {
	switch {
	case '0' <= ch && ch <= '9':
		return ch - '0', true
	case 'a' <= ch && ch <= 'f':
		return ch - 'a' + 10, true
	case 'A' <= ch && ch <= 'F':
		return ch - 'A' + 10, true
	default:
		return 0, false
	}
}

/*
func main() {
	forever := make(chan interface{})
	done := make(chan interface{})

	in := make(chan can.DataHolder)

	c, err := can.New("tcp://localhost:19000", in)
	if err != nil {
		log.Fatalf(err.Error())
	}
	c.Init()

	go func() {
		timer := time.NewTimer(5 * time.Second)
		ticker := time.NewTicker(3 * time.Second)
		var count int
	loop:
		for {
			select {
			case <-timer.C:
				break loop
			case <-ticker.C:
				c.Out <- &can.Message{
					ArbitrationID: []uint8{0x1c, 0xeb, 0xf7, 0xe8},
					// Priority:      0x00,
					// PGN:           hex.EncodeToString([]uint8{0xf7, 0x00}),
					// Src:           0xe8,
					// Dst: 0xf7,
					// Size: 0x8,
					Data: []uint8{0x5, 0x31, 0x2e, 0x30, 0x31, 0xa2, 0xff, 0xbf},
				}
				count++
			}
		}
		done <- count
	}()

	count := <-done
	fmt.Printf("%d\n", count)
	time.Sleep(1 * time.Second)
	c.Done <- true

	<-forever
}
*/
