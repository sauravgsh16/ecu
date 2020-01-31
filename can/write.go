package can

import (
	"bufio"
	"io"
)

// "Xtd 02 0CCBF782 08 13 00 86 00 B8 0B 00 00\n"

type writer struct {
	w io.Writer
}

func (w writer) writeMessage(m *Message) error {
	if err := m.write(w.w); err != nil {
		return err
	}

	if buf, ok := w.w.(*bufio.Writer); ok {
		if err := buf.Flush(); err != nil {
			return err
		}
	}

	return nil
}

func writeStringWithSpace(w io.Writer, s string) error {
	s = s + " "
	return writeString(w, s)
}

func writeString(w io.Writer, s string) error {
	p := []byte(s)

	if _, err := w.Write(p); err != nil {
		return err
	}
	return nil
}
