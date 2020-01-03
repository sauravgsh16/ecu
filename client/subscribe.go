package client

import (
	"context"
	"errors"
	"fmt"

	"github.com/sauravgsh16/message-server/qclient"
)

var (
	errSubscribeOnCloseConn = errors.New("subscriber on closed connection")
)

// Subscriber interface
type Subscriber interface {
	Subscribe(ctx context.Context) (chan *Message, error)
	Close() error
}

type subscriber struct {
	cw     *connectionWrapper
	ch     *qclient.Channel
	config Config
	errCh  chan interface{}
	out    chan *Message
}

// NewSubscriber returns a new subscriber
func NewSubscriber(c Config) (Subscriber, error) {
	if err := c.Validate(); err != nil {
		return nil, err
	}

	cw, err := newConnection(c)
	if err != nil {
		return nil, err
	}

	s := &subscriber{
		cw:     cw,
		config: c,
		errCh:  make(chan interface{}),
	}
	if err := s.setChannel(); err != nil {
		return nil, err
	}

	return s, nil
}

func (s *subscriber) setChannel() error {
	var err error

	s.ch, err = s.cw.conn.Channel()
	if err != nil {
		return err
	}
	return nil
}

func (s *subscriber) Close() error {
	if err := s.ch.Close(); err != nil {
		return err
	}
	return nil
}

func (s *subscriber) Subscribe(ctx context.Context) (chan *Message, error) {
	if s.cw.closed {
		return nil, errSubscribeOnCloseConn
	}

	s.out = make(chan *Message)

	switch s.config.Type {
	case broadcastExType:
		if err := s.prepareBroadcastConsumer(); err != nil {
			return nil, err
		}
	case p2pExType:
		if err := s.prepareDirectConsumer(); err != nil {
			return nil, err
		}
	default:
		panic("Unknown exchange type")
	}

	initiate := make(chan struct{})

	s.cw.subWG.Add(1)
	// Goroutine to close out channel once subscription context is cancelled
	go func(ctx context.Context) {
		defer func() {
			close(s.out)
			s.cw.subWG.Done()
		}()
	loop:
		for {
			select {
			case <-ctx.Done():
				break loop
			case <-initiate:
				s.run(ctx)
			}
		}
	}(ctx)

	// TODO : See if this works. Need to test throughly

	go func() {
		for {
			select {
			case <-s.errCh:
				close(s.out)
			}
		}
	}()

	close(initiate)
	return s.out, nil
}

func (s *subscriber) run(ctx context.Context) {
	qName := s.config.Consumer.Queue
	cName := s.config.Consumer.Consumer
	noAck := s.config.Consumer.NoAck
	noWait := s.config.Consumer.NoWait

	delivery, err := s.ch.Consume(qName, cName, noAck, noWait)
	if err != nil {
		s.errCh <- err
	}

	for {
		select {
		case msg := <-delivery:
			if err := s.processMessage(ctx, msg); err != nil {
				fmt.Printf("error: %s", err.Error())
				break
			}
		case <-ctx.Done():
			fmt.Printf("close from context received")
			break
		}
	}
}

func (s *subscriber) processMessage(ctx context.Context, dmsg qclient.Delivery) error {
	msg, err := s.config.Marshaller.Unmarshal(&dmsg)
	if err != nil {
		return err
	}

	ctx, cancelFunc := context.WithCancel(ctx)
	defer cancelFunc()

	msg.ctx = ctx

	select {
	case s.out <- msg:
	}

	return nil
}

func (s *subscriber) prepareBroadcastConsumer() error {
	if err := s.declareExchange(); err != nil {
		return err
	}

	if err := s.declareQueue(); err != nil {
		return err
	}

	if err := s.bindQueue(); err != nil {
		return err
	}
	return nil
}

func (s *subscriber) prepareDirectConsumer() error {
	return s.declareQueue()
}

func (s *subscriber) declareExchange() error {
	ex := s.config.Exchange.Name
	exType := s.config.Exchange.Type
	noWait := s.config.Exchange.NoWait

	return s.ch.ExchangeDeclare(ex, exType, noWait)
}

func (s *subscriber) declareQueue() error {
	q := s.config.QueueDeclare.Queue
	noWait := s.config.QueueDeclare.NoWait

	qOk, err := s.ch.QueueDeclare(q, noWait)
	if err != nil {
		return err
	}
	if qOk.Queue != q {
		return errors.New("Queue name mismatch")
	}
	return nil
}

func (s *subscriber) bindQueue() error {
	q := s.config.QueueBind.Queue
	ex := s.config.QueueBind.Exchange
	rk := s.config.QueueBind.RoutingKey
	noWait := s.config.QueueBind.NoWait

	return s.ch.QueueBind(q, ex, rk, noWait)
}
