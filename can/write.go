package can

import (
	"bufio"
	"fmt"
	"io"
	"sync"
)

const (
	hexchars   = "0123456789abcdef"
	bufferSize = 1024
)

func encode(src byte) []byte {
	dst := make([]byte, 2)
	dst[0] = hexchars[src>>4]
	dst[1] = hexchars[src&0x0F]
	return dst
}

type encoder struct {
	w   io.Writer
	err error
	out [1024]byte
	f   chan bool
	mu  *sync.RWMutex
}

// xtd 02 1CECF7E8 08 10 23 05 05 FF 00 CB 00\n
// []byte{0x78, 0x74, 0x64, 0x2, 0x1c, 0xeb, 0xf7, 0xe8, 0x8, 0x5, 0x31, 0x2e, 0x30, 0x31, 0xa2, 0xff, 0xbf}
// []byte{0x1c, 0xeb, 0xf7, 0xe8, 0x8, 0x5, 0x31, 0x2e, 0x30, 0x31, 0xa2, 0xff, 0xbf}

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

	var s string = common
	var chunk int = 8
	var i int

	for len(p) > 0 && e.err == nil {
		if len(p)%chunk != 0 && len(p) < chunk {
			chunk = len(p) % chunk
		}

		s += fmt.Sprintf("%02d ", chunk)

		for i = 0; i < chunk-1; i++ {
			s += string(encode(p[i]))
			s += " "
		}
		s += string(encode(p[i]))
		s += "\n"

		// fmt.Printf("writing: %slen:%d\n", s, len(s))

		e.mu.RLock()
		w, err := e.w.Write([]byte(s))
		if err != nil {
			e.err = err
		}
		e.mu.RUnlock()

		n += w / 2

		p = p[chunk:]
		s = common
	}

	e.f <- true

	return n, e.err
}

func newEncoder(w io.Writer, f chan bool, mu *sync.RWMutex) *encoder {
	return &encoder{
		w:  w,
		f:  f,
		mu: mu,
	}
}

type writer struct {
	w       io.Writer
	e       *encoder
	flushCh chan bool
	done    chan bool
	errCh   chan error
	mu      *sync.RWMutex
}

func newWriter(w io.Writer) *writer {
	var mu sync.RWMutex
	f := make(chan bool)
	e := newEncoder(w, f, &mu)

	wtr := &writer{
		w:       w,
		e:       e,
		flushCh: f,
		errCh:   make(chan error),
		done:    make(chan bool),
		mu:      &mu,
	}
	return wtr
}

func (w *writer) writeMessage(m *Message) error {
	b, err := m.group()
	if err != nil {
		return err
	}

	go w.flush()

	if _, err := w.e.Write(b); err != nil {
		return err
	}
	/*
	   	if buf, ok := w.w.(*bufio.Writer); ok {
	   		if err := buf.Flush(); err != nil {
	   			return err
	   		}
	   	}
	   	return nil
	   }
	*/

	select {
	case err := <-w.errCh:
		if err != nil {
			return err
		}
	}

	return nil
}

func (w *writer) flush() {
	select {
	case <-w.flushCh:
		if buf, ok := w.w.(*bufio.Writer); ok {
			w.mu.Lock()
			if err := buf.Flush(); err != nil {
				w.errCh <- err
			}
			w.mu.Unlock()
		}
		w.errCh <- nil
	}
}
