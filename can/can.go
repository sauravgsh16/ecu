package can

import (
	"bufio"
	"encoding/hex"
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

type tp struct {
	pgn    string
	size   int
	frames int
	data   []byte
	mux    sync.Mutex
}

func newTp(m *Message) *tp {
	// TODO: @Gourab: needs to get index of bytes which form PGN

	return &tp{
		pgn:    hex.EncodeToString(m.Data[6:]),
		size:   int(m.Data[1]),
		frames: int(m.Data[3]),
		data:   make([]byte, 0, int(m.Data[1])),
	}
}

func (t *tp) currSize() int {
	return len(t.data)
}

func (t *tp) append(d []byte) {
	t.mux.Lock()
	defer t.mux.Unlock()

	t.data = append(t.data, d...)
}

func (t *tp) isValid(pgn string) bool {
	return t.pgn == pgn
}

// Can struct
type Can struct {
	r      *reader
	w      *writer
	conn   *Connection
	currTP *tp
	In     chan *Message
	Out    chan *Message
	Done   chan bool
	wg     sync.WaitGroup
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
		Done: make(chan bool),
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
	handle := make(chan *tp)

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
						fmt.Printf("%#v\n", c.currTP)
						// TODO: process currTp before setting to nil

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

func (c *Can) handleMessage(h chan *tp) {
	for {
		select {
		case tp, ok := <-h:
			if !ok {
				break
			}

			switch tp.pgn {

			// announce Sn
			case "B000":

			// announce VIN
			case "B100":

			// send Join
			case "B200":

			// send Sn
			case "B300":

			// start rekey
			case "B400":

			// send nonce
			case "B500":
			}
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
