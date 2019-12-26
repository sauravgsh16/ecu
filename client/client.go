package client

import (
	"errors"
	"fmt"
	"sync"

	"github.com/sauravgsh16/ecu/config"
	"github.com/sauravgsh16/ecu/util"
	"github.com/sauravgsh16/message-server/proto"
	"github.com/sauravgsh16/message-server/qclient"
)

var (
	conn *qclient.Connection
)

var (
	errExchangeNotFound = errors.New("no exchange found")
	errQueueNotFound    = errors.New("no queue found")
)

// MessageServer interface
type MessageServer interface {
	DeclareExchange(string, string, bool) (int64, error)
	DeclareQueue(string, bool) (string, error)
	BindQueue(string, int64, bool) error
	BroadCast(int64, bool, []byte) error
	Send(string, bool, []byte) error
	Receive(string, string, bool, bool) (<-chan qclient.Delivery, error)
	Close(bool, bool) error
}

type messageServer struct {
	conn      *qclient.Connection
	ch        *qclient.Channel
	exchanges map[int64]string
	queues    map[string]*proto.QueueDeclareOk
	mux       sync.RWMutex
}

// NewMessageServerClient returns a message server client for a leader
func NewMessageServerClient() (MessageServer, error) {
	ms := &messageServer{
		conn:      conn,
		exchanges: make(map[int64]string, 0),
		queues:    make(map[string]*proto.QueueDeclareOk),
	}

	var err error
	ms.ch, err = ms.conn.Channel()
	if err != nil {
		return nil, err
	}

	return ms, nil
}

func init() {
	var err error

	conn, err = qclient.Dial(config.MessageServerHost)
	if err != nil {
		panic(fmt.Sprintf("failed to connect to message server, %s", err))
	}
}

func (ms *messageServer) DeclareExchange(name, extype string, noWait bool) (int64, error) {
	if err := ms.ch.ExchangeDeclare(name, extype, noWait); err != nil {
		return 0, err
	}

	ms.mux.Lock()
	defer ms.mux.Unlock()

	id := util.NextCounter()
	ms.exchanges[id] = name

	return id, nil
}

func (ms *messageServer) DeclareQueue(name string, noWait bool) (string, error) {
	qok, err := ms.ch.QueueDeclare(name, noWait)
	if err != nil {
		return "", err
	}

	ms.mux.Lock()
	defer ms.mux.Unlock()

	ms.queues[qok.Queue] = qok
	return qok.Queue, nil
}

func (ms *messageServer) BindQueue(qName string, exID int64, noWait bool) error {
	q, ok := ms.queues[qName]
	if !ok {
		return errQueueNotFound
	}
	exName, ok := ms.exchanges[exID]
	if !ok {
		return errExchangeNotFound
	}

	rKey := ""
	if err := ms.ch.QueueBind(q.Queue, exName, rKey, noWait); err != nil {
		return err
	}
	return nil
}

func (ms *messageServer) BroadCast(exID int64, immediate bool, body []byte) error {
	exName, ok := ms.exchanges[exID]
	if !ok {
		return errExchangeNotFound
	}

	rKey := ""
	if err := ms.ch.Publish(exName, rKey, immediate, body); err != nil {
		return err
	}
	return nil
}

func (ms *messageServer) Send(qName string, immediate bool, body []byte) error {
	q, ok := ms.queues[qName]
	if !ok {
		return errQueueNotFound
	}

	exName := ""
	if err := ms.ch.Publish(exName, q.Queue, immediate, body); err != nil {
		return err
	}
	return nil
}

func (ms *messageServer) Receive(qName, consumerName string, noAck, noWait bool) (<-chan qclient.Delivery, error) {
	q, ok := ms.queues[qName]
	if !ok {
		return nil, errQueueNotFound
	}

	c, err := ms.ch.Consume(q.Queue, consumerName, noAck, noWait)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (ms *messageServer) closeCH(isUsed, noWait bool) error {
	// TODO: check if queue unbind is required

	// Delete queue
	for q := range ms.queues {
		if _, err := ms.ch.QueueDelete(q, isUsed, true, noWait); err != nil {
			return err
		}
	}

	for _, ex := range ms.exchanges {
		if err := ms.ch.ExchangeDelete(ex, isUsed, false); err != nil {
			return err
		}
	}

	// TODO: empty maps

	return ms.ch.Close()
}

func (ms *messageServer) closeConn(isUsed, noWait bool) error {
	if err := ms.closeCH(isUsed, noWait); err != nil {
		return err
	}

	if ms.conn.IsClosed() {
		return nil
	}
	return ms.conn.Close()
}

func (ms *messageServer) Close(isUsed, noWait bool) error {
	return ms.closeConn(isUsed, noWait)
}
