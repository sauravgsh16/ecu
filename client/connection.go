package client

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/sauravgsh16/ecu/config"

	"github.com/sauravgsh16/message-server/qclient"
)

const (
	tickerInterval  = 5
	timeoutInterval = 30
)

var (
	c    conn
	once sync.Once
)

type connectionWrapper struct {
	conn      *qclient.Connection
	mux       sync.Mutex
	closed    bool
	pubWG     sync.WaitGroup
	subWG     sync.WaitGroup
	connected bool
}

type conn struct {
	client      *qclient.Connection
	initialized bool
}

func (c *conn) initialize() error {
	return c.connect()
}

func newConnection() (*connectionWrapper, error) {
	var err error
	once.Do(func() {
		err = c.initialize()
		c.initialized = true
	})
	if err != nil {
		return nil, err
	}
	if !c.initialized {
		return nil, errors.New("uninitialized connected")
	}

	return &connectionWrapper{conn: c.client, connected: true}, nil
}

func (c *conn) connect() error {
	connected := make(chan *qclient.Connection)
	ticker := time.NewTicker(tickerInterval * time.Second)

	go func() {
		timeout := time.After(timeoutInterval * time.Second)
		for {
			select {
			case <-ticker.C:
				conn, err := qclient.Dial(config.MessageServerHost)
				if err != nil {
					fmt.Printf("Error connecting to message server: %s", err.Error())
					continue
				}
				connected <- conn
				return
			case <-timeout:
				close(connected)
				return
			}
		}
	}()

	var ok bool
	c.client, ok = <-connected
	ticker.Stop()
	if !ok {
		return errors.New("failed to connect to message server")
	}

	close(connected)
	return nil
}

func (cw *connectionWrapper) Close() error {
	if cw.closed {
		return nil
	}

	// Wait for publisher close
	cw.pubWG.Wait()

	if err := cw.conn.Close(); err != nil {
		return fmt.Errorf("error closing connection: %s", err.Error())
	}

	// Wait for subscriber to close
	cw.subWG.Wait()

	cw.mux.Lock()
	defer cw.mux.Unlock()

	cw.closed = true
	cw.connected = false
	return nil
}
