package can

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
)

var (
	vpsAddr = "tcp://localhost:19000"
)

// Can struct
type Can struct {
	r    *reader
	w    *writer
	conn *Connection
	In   chan *Message
	Out  chan *Message
}

// New returns a pointer to Can
func New() (*Can, error) {
	conn, err := Dial(vpsAddr)
	if err != nil {
		return nil, err
	}
	buf := bufio.NewReader(conn.Conn)

	c := &Can{
		conn: conn,
		r:    newReader(buf),
		w:    newWriter(conn.writer.w),
		In:   make(chan *Message),
		Out:  make(chan *Message),
	}

	go c.handleIncoming()
	go c.handleOutgoing()
	return c, nil
}

func (c *Can) handleIncoming() {
	for {
		if c.conn.closed {
			break
		}
		msg, err := c.r.readMessage()
		if err == io.EOF {
			return
		}
		if err != nil {
			if netErr, ok := err.(net.Error); ok {
				log.Printf("closing connection: %s\n", netErr)
				break
			}
			// TODO: better error handling
			log.Printf(err.Error())
			continue
		}
		c.In <- msg
	}
}

// TODO: Call inside go routine
// Extend: to send data to message-server
func (c *Can) Read() {
	for m := range c.In {
		fmt.Printf("%#v\n", m)
	}
}

func (c *Can) handleOutgoing() {
	for {
		if c.conn.closed {
			break
		}
		select {
		case m := <-c.Out:
			go c.write(m)
		}
	}
}

func (c *Can) write(m *Message) error {
	return c.w.writeMessage(m)
}
