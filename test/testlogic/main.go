package main

import (
	"fmt"
	"io"
	"sync"
)

const (
	hexchars = "0123456789abcdef"
)

func fromHexChar(ch byte) (byte, bool) {
	fmt.Printf("%#v\n", ch)
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

func encode(src byte) []byte {
	dst := make([]byte, 2)
	dst[0] = hexchars[src>>4]
	dst[1] = hexchars[src&0x0F]
	return dst
}

func getFrames(l int) uint8 {
	if l <= 8 {
		return uint8(1)
	}

	var chunk int = 7
	var frames uint8
	for l > 0 {
		if l%chunk != 0 && l < chunk {
			chunk = l % chunk
		}
		l -= chunk
		frames++
	}
	return frames
}

func main() {
	/*
		p := []byte{0x1c, 0xeb, 0xf7, 0xe8, 0x5, 0x31, 0x2e, 0x30, 0x31, 0xa2, 0xff, 0xbf, 0x5, 0x31, 0x2e, 0x30, 0x31, 0xa2, 0xff, 0xbf}

		e := &encoder{}
		if _, err := e.Write(p); err != nil {
			log.Fatalf(err.Error())
		}
	*/
	p := make([]byte, 182)
	// p := []byte{0x5, 0x31, 0x2e, 0x30, 0x31, 0xa2, 0xff, 0xbf, 0x5, 0x31, 0x2e, 0x30, 0x31, 0xa2, 0xff, 0xbf, 0x5, 0x31, 0x2e, 0x30, 0x31, 0xa2, 0xff, 0xbf, 0xff, 0xbf}
	fmt.Printf("%#v", getFrames(len(p)))
}

type encoder struct {
	w   io.Writer
	err error
	f   chan bool
	mu  *sync.RWMutex
}

func (e *encoder) encodeAndWrite(p []byte, chunk int, s string) (int, error) {
	var i int
	for i = 0; i < chunk-1; i++ {
		s += string(encode(p[i]))
		s += " "
	}
	s += string(encode(p[i]))
	s += "\n"

	fmt.Printf("writing: %s", s)

	e.mu.RLock()
	w, err := e.w.Write([]byte(s))
	e.mu.RUnlock()

	return w, err
}

func (e *encoder) Write(p []byte) (n int, err error) {
	if len(p) <= 0 {
		return
	}

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
			for _, b := range p[:4] {
				common += string(encode(b))
			}
			common += " "
			c++
			p = p[4:]
		}

	}

	var chunk int = 8
	var s string = common

	switch {
	case len(p) <= chunk:
		s += fmt.Sprintf("%02d ", chunk)
		w, err := e.encodeAndWrite(p, chunk, s)
		if err != nil {
			e.err = err
		}
		n += w / 2

	case len(p) > chunk:
		var counter uint8 = 1
		chunk = chunk - 1

		for len(p) > 0 && e.err == nil {
			if len(p)%chunk != 0 && len(p) < chunk {
				chunk = len(p) % chunk
			}

			s += fmt.Sprintf("%02d %02d ", chunk+1, counter)

			w, err := e.encodeAndWrite(p, chunk, s)
			if err != nil {
				e.err = err
			}
			n += w / 2

			p = p[chunk:]
			s = common
			counter++
		}
	}

	e.f <- true

	return n, e.err
}
