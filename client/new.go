package client

import (
	"log"
	"time"

	"github.com/sauravgsh16/ecu/config"
	"github.com/sauravgsh16/message-server/qclient"
)

// Message struct
type Message struct{}

// MessageSession struct
type MessageSession struct {
	conn      *qclient.Connection
	ch        *qclient.Channel
	connected chan *qclient.Connection
	Publish   chan Message
	Consumer  chan *qclient.Delivery
}

// Close tears down the MessageSession
func (ms *MessageSession) Close() error {
	if ms.conn == nil {
		return nil
	}
	return ms.conn.Close()
}

func initConn(connected chan<- *qclient.Connection) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		conn, err := qclient.Dial(config.MessageServerHost)
		if err != nil {
			log.Printf("error connection to message server, %v", err)
			log.Println("retry after 5 seconds")
			continue
		}
		connected <- conn
		return
	}
}
