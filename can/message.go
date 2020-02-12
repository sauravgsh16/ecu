package can

import (
	"bytes"
	"encoding/hex"
	"errors"
	"strings"
	"sync"
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
	PGN           string
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

	var pgn []byte
	if int(m.ArbitrationID[2]) > broadcastThreshold {
		pgn = []byte{m.ArbitrationID[1], 0x00}
	} else {
		pgn = m.ArbitrationID[1:3]
	}

	m.PGN = strings.ToUpper(hex.EncodeToString(pgn))
	return nil
}

// TP Can Message
type TP struct {
	Pgn    string
	size   int
	frames int
	Data   []byte
	mux    sync.Mutex
}

func newTp(m *Message) *TP {
	return &TP{
		Pgn: hex.EncodeToString([]byte{m.Data[5], m.Data[6]}),
		// TODO: change package format
		size:   int(m.Data[2] + m.Data[1]),
		frames: int(m.Data[3]),
		Data:   make([]byte, 0, int(m.Data[1])),
	}
}

func (t *TP) currSize() int {
	return len(t.Data)
}

func (t *TP) append(d []byte) {
	t.mux.Lock()
	defer t.mux.Unlock()

	t.Data = append(t.Data, d...)
}

func (t *TP) isValid(pgn string) bool {
	return t.Pgn == pgn
}
