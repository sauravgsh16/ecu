package can

import (
	"bytes"
	"errors"
	"time"
)

const (
	broadcastThreshold int = 239
)

var (
	errorEmptyArbitrationID = errors.New("attempt to parse empty arbitration id")
)

// Message struct
type Message struct {
	ArbitrationID []byte
	Priority      byte
	PGN           []byte
	Src           byte
	Dst           byte
	Size          byte
	Data          []byte
	TimeStamp     int64
}

// "Xtd 02 0CCBF782 08 13 00 86 00 B8 0B 00 00\n"

// New Can message
func newMsg() *Message {
	return &Message{
		ArbitrationID: make([]byte, 4),
		TimeStamp:     time.Now().UnixNano(),
		Data:          make([]byte, 8),
	}
}

func (m *Message) group() ([]byte, error) {
	b := make([]byte, 0)
	buf := bytes.NewBuffer(b)

	if _, err := buf.Write(m.ArbitrationID); err != nil {
		return nil, err
	}

	// writing size not necessary
	// needs to be dynamically written
	/*
		if err := binary.Write(buf, binary.BigEndian, m.Size); err != nil {
			return nil, err
		}
	*/

	if _, err := buf.Write(m.Data); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (m *Message) parseArbitrationID() error {
	if len(m.ArbitrationID) == 0 {
		return errorEmptyArbitrationID
	}

	m.Priority = m.ArbitrationID[0]
	m.Src = m.ArbitrationID[3]
	m.Dst = m.ArbitrationID[2]

	m.PGN = make([]byte, 0, 2)
	if int(m.ArbitrationID[2]) > broadcastThreshold {
		b := []byte{m.ArbitrationID[2], 0x00}
		m.PGN = append(m.PGN, b...)
	} else {
		m.PGN = append(m.PGN, m.ArbitrationID[2:4]...)
	}
	return nil
}
