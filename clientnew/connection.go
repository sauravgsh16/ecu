package clientnew

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/sauravgsh16/message-server/qclient"
)

const (
	tickerInterval  = 5
	timeoutInterval = 30
)

type connectionWrapper struct {
	config Config
	conn   *qclient.Connection
	mux    sync.Mutex
	closed bool
	pubWG  sync.WaitGroup
	subWG  sync.WaitGroup
}

func newConnection(c Config) (*connectionWrapper, error) {
	cw := &connectionWrapper{
		config: c,
	}

	if err := cw.connect(); err != nil {
		return nil, err
	}

	return cw, nil
}

func (cw *connectionWrapper) connect() error {
	connected := make(chan *qclient.Connection)
	ticker := time.NewTicker(tickerInterval * time.Second)

	go func() {
		timeout := time.After(timeoutInterval * time.Second)
		for {
			select {
			case <-ticker.C:
				conn, err := qclient.Dial(cw.config.URI)
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
	conn, ok := <-connected
	if !ok {
		return errors.New("failed to connect to message server")
	}
	close(connected)
	cw.conn = conn
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

	cw.closed = true
	return nil
}
