package can

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"sync"
)

var (
	vpsAddr = "tcp://localhost:19000"
)

// Can struct
type Can struct {
	r            *reader
	w            *writer
	conn         *Connection
	currTP       *TP
	In           chan *Message
	Out          chan *Message
	Done         chan bool
	wg           sync.WaitGroup
	ServIncoming chan *TP
}

// New returns a pointer to Can
func New(in chan *TP) (*Can, error) {
	conn, err := Dial(vpsAddr)
	if err != nil {
		return nil, err
	}
	buf := bufio.NewReader(conn.Conn)

	c := &Can{
		conn:         conn,
		r:            newReader(buf),
		w:            newWriter(conn.writer.w),
		In:           make(chan *Message),
		Out:          make(chan *Message),
		Done:         make(chan bool),
		ServIncoming: in,
	}
	return c, nil
}

// Init can goroutines
func (c *Can) Init() {
	go c.handleIncoming()
	go c.processIncoming()
	go c.handleOutgoing()
	go c.handleClose()
}

func (c *Can) handleClose() {
	select {
	case <-c.Done:
		c.conn.close()
	}
}

func (c *Can) handleIncoming() {
	for {
		if c.conn.isClosed() {
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
	close(c.In)
}

func (c *Can) processIncoming() {
	var fc int
	handle := make(chan *TP)

	go func() {
	loop:
		for {
			select {

			case msg, ok := <-c.In:
				if !ok {
					break loop
				}

				switch {
				case isPrefix(msg.PGN, "EC"):
					if c.currTP != nil {
						c.wg.Wait()
					}

					c.currTP = newTp(msg)
					c.wg.Add(1)

				case isPrefix(msg.PGN, "EB"):
					if c.currTP == nil {
						log.Fatalf("invalid tp pgn '%s' received, before receving tp initial tp info", msg.PGN)
					}

					if fc < c.currTP.frames {
						c.currTP.append(msg.Data[1:])
						fc++
					}

					if fc >= c.currTP.frames {
						// handle <- c.currTP
						fmt.Printf("%#v\n, len:%d\n", c.currTP, len(c.currTP.Data))

						c.currTP = nil
						fc = 0
						c.wg.Done()
					}
				default:
					fmt.Println("Here")
				}
			}
		}
		close(handle)
	}()
	go c.handleMessage(handle)
}

func (c *Can) handleMessage(h chan *TP) {
	for {
		select {
		case tp, ok := <-h:
			if !ok {
				break
			}

			c.ServIncoming <- tp
		}
	}
}

func (c *Can) handleOutgoing() {
loop:
	for {
		select {
		case m := <-c.Out:
			if c.conn.isClosed() {
				break loop
			}
			go c.write(m)
		}
	}
	close(c.Out)
}

func (c *Can) write(m *Message) error {
	if err := c.w.writeMessage(m); err != nil {
		log.Printf(err.Error())
		return err
	}
	return nil
}

func isPrefix(s, substr string) bool {
	if len(s) < len(substr) {
		return false
	}

	i := 0
	s = strings.ToUpper(s)
	for i < len(substr) {
		if s[i] != substr[i] {
			return false
		}
		i++
	}
	return true
}
