package client

import (
	"errors"

	"github.com/hashicorp/go-multierror"

	"github.com/sauravgsh16/message-server/qclient"
)

var (
	errPublishOnClosedConn = errors.New("publish on closed connection")
	errPublishFailure      = errors.New("failed to publish message")
)

// Publisher interface
type Publisher interface {
	Publish(msg ...*Message) error
	Close() error
}

type publisher struct {
	cw     *connectionWrapper
	ch     *qclient.Channel
	config Config
}

// NewPublisher returns a new publisher
func NewPublisher(c Config) (Publisher, error) {
	if err := c.Validate(); err != nil {
		return nil, err
	}

	cw, err := newConnection()
	if err != nil {
		return nil, err
	}

	p := &publisher{cw: cw, config: c}
	if err := p.setChannel(); err != nil {
		return nil, err
	}

	return p, nil
}

func (p *publisher) setChannel() error {
	var err error

	p.ch, err = p.cw.conn.Channel()
	if err != nil {
		return err
	}
	return nil
}

func (p *publisher) Close() error {

	// TODO: Better implmentation
	// ChannelClose should wait for ChannelCloseOk message

	if err := p.ch.Close(); err != nil {
		return err
	}
	return nil
}

func (p *publisher) Publish(msgs ...*Message) error {
	if p.cw.closed {
		return errPublishOnClosedConn
	}

	p.cw.pubWG.Add(1)
	defer p.cw.pubWG.Done()

	switch p.config.Type {
	case broadcastExType:
		if err := p.declareExchange(); err != nil {
			return err
		}
	case p2pExType:
		if err := p.declareQueue(); err != nil {
			return err
		}
	default:
		panic("Unknown exchange type")
	}

	var err error

	for _, msg := range msgs {
		if pubErr := p.publishMessage(msg); pubErr != nil {
			err = multierror.Append(err, pubErr)
		}
	}
	return err
}

func (p *publisher) publishMessage(msg *Message) error {
	servMsg, err := p.config.Marshaller.Marshal(msg)
	if err != nil {
		return err
	}

	ex := p.config.Exchange.Name
	rk := p.config.QueueDeclare.Queue
	imme := p.config.Publish.Immediate

	if err := p.ch.Publish(ex, rk, imme, servMsg); err != nil {
		return errPublishFailure
	}
	return nil
}

func (p *publisher) declareExchange() error {
	ex := p.config.Exchange.Name
	exType := p.config.Exchange.Type
	noWait := p.config.Exchange.NoWait

	if err := p.ch.ExchangeDeclare(ex, exType, noWait); err != nil {
		return err
	}
	return nil
}

func (p *publisher) declareQueue() error {
	q := p.config.QueueDeclare.Queue
	noWait := p.config.QueueDeclare.NoWait

	_, err := p.ch.QueueDeclare(q, noWait)
	if err != nil {
		return err
	}
	return nil
}
