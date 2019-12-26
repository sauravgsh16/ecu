package consume

import (
	"errors"

	"github.com/sauravgsh16/ecu/client"
	"github.com/sauravgsh16/ecu/util"
	"github.com/sauravgsh16/message-server/qclient"
)

const (
	Receive = iota
	Subscribe
)

// Consumer interface
type Consumer interface{}

// Receiver struct
type receiver struct {
	ID     int64
	Queue  string
	Name   string
	noWait bool
	noAck  bool
	dChan  chan qclient.Delivery
	client client.MessageServer
}

// Subscriber struct
type subscriber struct {
	ID     int64
	Queue  string
	ExID   int64
	Name   string
	noWait bool
	noAck  bool
	dChan  chan qclient.Delivery
	done   chan bool
	client client.MessageServer
}

func (s *subscriber) init() {

}

func (s *subscriber) bindQueue() error {
	if err := s.client.BindQueue(s.Queue, s.ExID, s.noWait); err != nil {
		return err
	}
	return nil
}

func (s *subscriber) Consume() (chan qclient.Delivery, error) {
	return s.dChan, nil
}

// New returns a new consumer
func New(qName, consumerName string, exID int64, noAck, noWait bool, cType int, client client.MessageServer) (Consumer, error) {
	switch cType {
	case Receive:
		return &receiver{
			ID:     util.NextCounter(),
			Queue:  qName,
			Name:   consumerName,
			noWait: noWait,
			noAck:  noAck,
			client: client,
		}, nil
	case Subscribe:
		s := &subscriber{
			ID:     util.NextCounter(),
			Queue:  qName,
			ExID:   exID,
			Name:   consumerName,
			noWait: noWait,
			noAck:  noAck,
			client: client,
		}
		if err := s.bindQueue(); err != nil {
			return nil, err
		}
		return s, nil
	default:
		return nil, errors.New("invalid consumer type")
	}
}
