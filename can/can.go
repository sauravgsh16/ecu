package can

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"time"
)

const (
	broadcastThreshold int64 = 239
)

// Message struct
type Message struct {
	ArbitrationID []byte
	PduFormat     int
	PduSpecific   int
	Priority      int64
	PGN           string
	Src           int64
	Dst           int64
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

func (m *Message) getArbitrationID() string {
	return fmt.Sprintf("%02X%s%02X%02X", m.Priority, m.PGN[:2], m.Dst, m.Src)
}

func (m *Message) write(w io.Writer) error {
	b := make([]byte, 0, 3+2+8+len(m.Data)+12)
	_ = bytes.NewBuffer(b)

	// Write format xtd - blah
	if err := writeStringWithSpace(w, "Xtd"); err != nil {
		return err
	}

	// Write 02 - god knows what this is
	if err := writeStringWithSpace(w, "02"); err != nil {
		return err
	}

	// Write arbitrationID
	if err := writeStringWithSpace(w, m.getArbitrationID()); err != nil {
		return err
	}

	// Write no of bytes
	l := fmt.Sprintf("%02d", len(strings.Split(string(m.Data), " ")))
	if err := writeStringWithSpace(w, l); err != nil {
		return err
	}
	// Write data

	return nil
}

func getHexFromInt(i int64) (string, error) {
	return "", nil
}

func (m *Message) read(r io.Reader) error {
	return nil
}

// FormMessage creates a CAN message struct from input parameters
func FormMessage() {}

// FromMessage creates a CAN message struct from message read on tcp
func FromMessage() {
}
