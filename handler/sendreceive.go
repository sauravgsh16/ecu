package handler

import (
	"context"
	"errors"

	"github.deere.com/sg30983/ecu/client"
)

// Sender interface
type Sender interface {
	Send(msg *client.Message) error
	GetName() string
	Close() error
}

// Receiver interface
type Receiver interface {
	StartConsumer(chan interface{}) (chan *client.Message, error)
	GetName() string
}

type send struct {
	name string
	p    client.Publisher
}

func (s *send) Send(msg *client.Message) error {
	return s.p.Publish(msg)
}

func (s *send) GetName() string {
	return s.name
}

func (s *send) Close() error {
	return s.p.Close()
}

type receive struct {
	name string
	s    client.Subscriber
	ctx  context.Context
	out  chan *client.Message
}

func (r *receive) GetName() string {
	return r.name
}

func (r *receive) StartConsumer(done chan interface{}) (chan *client.Message, error) {
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
