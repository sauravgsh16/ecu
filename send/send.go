package service

import (
	"errors"

	"github.com/sauravgsh16/ecu/client"
	"github.com/sauravgsh16/ecu/util"
)

const (
	initialized = true
)

// Sender struct
type Sender struct {
	ID        int64
	QueueName string
	status    bool
	noWait    bool
	client    client.MessageServer
}

func (s *Sender) init(qn string) error {
	if s.status != initialized {
		return errors.New("init on a uninitialized Sender")
	}

	qName, err := s.client.DeclareQueue(qn, s.noWait)
	if err != nil {
		return err
	}

	if qName != qn {
		return errors.New("invalid queue name")
	}

	s.QueueName = qName
	s.status = initialized
	return nil
}

// GetQueueName of the sender
func (s *Sender) GetQueueName() string {
	return s.QueueName
}

// New returns a new sender
func New(reqName string, noWait bool, c client.MessageServer) (*Sender, error) {
	s := &Sender{
		ID:     util.NextCounter(),
		noWait: noWait,
		client: c,
	}

	if err := s.init(reqName); err != nil {
		return nil, err
	}

	return s, nil
}
