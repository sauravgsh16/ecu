package can

import (
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
)

type decoder struct {
	r   io.Reader
	err error
	in  []byte
	out [1024]byte
}

func newDecoder(r io.Reader) *decoder {
	return &decoder{
		r: r,
	}
}

type invalidByte byte

func (i invalidByte) Error() string {
	return fmt.Sprintf("encoding error: invalid byte: %#U", rune(i))
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

func decode(dst, src []byte) (int, int, error) {
	fmt.Printf("%#v\n", src)
	i, j, s := 0, 0, 0
	for i < len(src) && j < len(dst) {
		fmt.Printf("%02x, %02x\n", src[i], src[i+1])
		if src[i] == ' ' {
			i++
			s++
			continue
		}
		a, ok := fromHexChar(src[i])
		if !ok {
			fmt.Println("Exiting here 1")
			return 0, j, invalidByte(src[i])
		}
		b, ok := fromHexChar(src[i+1])
		if !ok {
			fmt.Println("Exiting here 2")
			return 0, j, invalidByte(src[i+1])
		}

		dst[j] = (a << 4) | b
		i += 2
		j++

		fmt.Printf("i: %d, j:%d\n", i, j)
	}
	return j, s, nil
}

// "Xtd 02 0CCBF782 08 13 00 86 00 B8 0B 00 00\n"

func (d *decoder) Read(p []byte) (n int, err error) {
	if len(d.in) < 2 && d.err == nil {
		var numRead int
		numRead, d.err = d.r.Read(d.out[0 : len(p)*2+1])
		d.in = d.out[:numRead]
		if d.err == io.EOF && len(d.in)%2 != 0 {
			if _, ok := fromHexChar(d.in[len(d.in)-1]); !ok {
				d.err = invalidByte(d.in[len(d.in)-1])
			}
		}
	}

	// After reading data, we decode it
	if numAvail := len(d.in) / 2; len(p) > numAvail {
		p = p[:numAvail]
	}

	numdec, numspace, err := decode(p, d.in)
	d.in = d.in[(numdec*2)+numspace:]
	if len(d.in) > 0 {
		return numdec, errors.New("could not read full")
	}

	return numdec, nil
}

type reader struct {
	r io.Reader
}

func (r reader) read() (*Message, error) {
	m := newMsg()

	if err := m.read(r.r); err != nil {
		return nil, err
	}

	return m, nil
}

func (r reader) readMessage() (*Message, error) {
	m := newMsg()
	incoming := make([]byte, 0, 7)

	if _, err := io.ReadFull(r.r, incoming[:7]); err != nil {
		return nil, err
	}
	// 0CCBF782 08 13 00 86 00 B8 0B 00 00\n
	d := newDecoder(r.r)
	b := make([]byte, 4)

	if _, err := d.Read(b); err != nil {
		return nil, err
	}

	fmt.Printf("%08X\n", b)
	/*
				size, err := getIntValFromHex(string(incoming[16:18]))
				if err != nil {
					return nil, err
				}
		/*

		d := hex.NewDecoder(r.r)
		b := make([]byte, 4)

		i, err := d.Read(b)
		if err != nil {
			log.Println(err.Error())
		}

		fmt.Printf("%d : %08X\n", i, b)
	*/
	return m, nil
}

/*
func (r reader) readMessage() (*Message, error) {
	var err error
	m := newMsg()

	incoming := make([]byte, 19)

	if _, err = io.ReadFull(r.r, incoming[:19]); err != nil {
		return nil, err
	}
	m.Priority, m.PGN, m.Dst, m.Src, err = r.parseArbitrationID(incoming[7:15])
	if err != nil {
		return nil, err
	}

	size, err := getIntValFromHex(string(incoming[16:18]))
	if err != nil {
		return nil, err
	}

	m.Data, err = r.readData(size)
	if err != nil {
		return nil, err
	}

	return m, nil
}

func (r reader) parseArbitrationID(b []byte) (int64, string, int64, int64, error) {
	priority, err := getIntValFromHex(string(b[:2]))
	if err != nil {
		return 0, "", 0, 0, err
	}

	dst, err := getIntValFromHex(string(b[4:6]))
	if err != nil {
		return 0, "", 0, 0, err
	}

	var pgn string
	if dst > broadcastThreshold {
		pgn = string(b[2:4]) + "00"
	} else {
		pgn = string(b[2:6])
	}

	src, err := getIntValFromHex(string(b[6:]))
	if err != nil {
		return 0, "", 0, 0, err
	}

	return priority, pgn, dst, src, nil
}

func (r reader) readData(size int64) ([]byte, error) {
	// Data size count: .
	// no of space :- size - 1
	// data - 2 bytes each.
	// Total space to allocate := 2 * size + (size-1) + 1 (new line char)

	body := make([]byte, 2*size+(size-1)+1)

	if err := binary.Read(r.r, binary.BigEndian, &body); err != nil {
		return nil, err
	}

	return body, nil
}
*/
func getIntValFromHex(s string) (int64, error) {
	s = strings.Replace(s, "0x", "", -1)
	i, err := strconv.ParseUint(s, 16, 64)
	if err != nil {
		return 0, err
	}
	return int64(i), nil
}
