package can

import (
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"
)

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
	incoming := make([]byte, 0, 19)

	if _, err := io.ReadFull(r.r, incoming[:19]); err != nil {
		return nil, err
	}

	size, err := getIntValFromHex(string(incoming[16:18]))
	if err != nil {
		return nil, err
	}

	d := hex.NewDecoder(r.r)
	b := make([]byte, 2*size+(size-1)+1)

	i, err := d.Read(b)
	if err != nil {
		log.Println(err.Error())
	}

	fmt.Printf("%d \n %#v\n", i, b)
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
