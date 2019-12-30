package clientnew

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
	PublishMessage(msg ...*Message) error
	ClosePublisher() error
}

type publisher struct {
	cw     *connectionWrapper
	Ch     *qclient.Channel
	config Config
}

// NewPublisher returns a new publisher
func NewPublisher(c Config) (Publisher, error) {
	if err := c.Validate(); err != nil {
		return nil, err
	}

	cw, err := newConnection(c)
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

	p.Ch, err = p.cw.conn.Channel()
	if err != nil {
		return err
	}
	return nil
}

func (p *publisher) ClosePublisher() error {
	if err := p.Ch.Close(); err != nil {
		return err
	}
	return nil
}

func (p *publisher) PublishMessage(msgs ...*Message) error {
	if p.cw.closed {
		return errPublishOnClosedConn
	}

	switch p.config.Exchange.Type {
	case "fanout":
		if err := p.declareExchange(); err != nil {
			return err
		}
	case "direct":
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

	if err := p.Ch.Publish(ex, rk, imme, servMsg); err != nil {
		return errPublishFailure
	}
	return nil
}

func (p *publisher) declareExchange() error {
	ex := p.config.Exchange.Name
	exType := p.config.Exchange.Type
	noWait := p.config.Exchange.NoWait

	if err := p.Ch.ExchangeDeclare(ex, exType, noWait); err != nil {
		return err
	}
	return nil
}

func (p *publisher) declareQueue() error {
	q := p.config.QueueDeclare.Queue
	noWait := p.config.QueueDeclare.NoWait

	_, err := p.Ch.QueueDeclare(q, noWait)
	if err != nil {
		return err
	}
	return nil
}
