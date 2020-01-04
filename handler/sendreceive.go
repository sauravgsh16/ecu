package handler

import (
	"context"
	"errors"

	"github.com/sauravgsh16/ecu/client"
)

// Sender interface
type Sender interface {
	Send(msg *client.Message) error
}

// Receiver interface
type Receiver interface {
	StartReceiver(chan interface{}) (chan *client.Message, error)
}

type send struct {
	p client.Publisher
}

func (s *send) Send(msg *client.Message) error {
	return s.p.Publish(msg)
}

type receive struct {
	s   client.Subscriber
	ctx context.Context
	out chan *client.Message
}

func (r *receive) StartReceiver(done chan interface{}) (chan *client.Message, error) {
	if r.out != nil {
		return nil, errors.New("receiver already started")
	}
	var err error

	r.out, err = r.s.Subscribe(r.ctx)
	if err != nil {
		return nil, err
	}

	go func() {
	loop:
		for {
			select {
			case <-done:
				r.ctx.Done()
				break loop
			}
		}
	}()
	return r.out, nil
}
