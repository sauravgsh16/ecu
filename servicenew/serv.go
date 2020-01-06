package servicenew

import (
	"fmt"
	"sync"

	"github.com/sauravgsh16/ecu/config"

	"github.com/hashicorp/go-multierror"

	"github.com/sauravgsh16/ecu/client"

	"github.com/sauravgsh16/ecu/domain"
	"github.com/sauravgsh16/ecu/handler"
)

// Leader interface
type Leader interface{}

// Member interface
type Member interface{}

type listenerch struct {
	name string
	ch   chan *client.Message
	done chan interface{}
	h    handler.Receiver
	init bool
}

type incoming struct {
	name string
	msg  *client.Message
}

type ecuService struct {
	domain    *domain.Ecu
	senders   map[string]handler.Sender
	receivers map[string]*listenerch
	mux       sync.RWMutex
	incoming  chan *incoming
}

func newService(c *ecuConfig) (*ecuService, error) {
	d, err := domain.NewEcu(c.ecuType)
	if err != nil {
		return nil, err
	}

	e := &ecuService{
		domain:    d,
		senders:   make(map[string]handler.Sender),
		receivers: make(map[string]*listenerch),
		incoming:  make(chan *incoming),
	}

	e.mux.Lock()
	defer e.mux.Unlock()

	// Register senders
	for _, h := range c.senders {
		sender, err := h()
		if err != nil {
			return nil, err
		}
		e.senders[sender.GetName()] = sender
	}

	// Register receivers
	for _, h := range c.receivers {
		receiver, err := h()
		if err != nil {
			return nil, err
		}
		e.receivers[receiver.GetName()] = &listenerch{
			h:    receiver,
			done: make(chan interface{}),
			name: receiver.GetName(),
		}
	}

	return e, nil
}

func (e *ecuService) startreceivers() error {
	var err error

	if len(e.receivers) == 0 {
		return nil
	}

	for _, r := range e.receivers {
		ch, er := r.h.StartReceiver(r.done)
		if er != nil {
			err = multierror.Append(err, er)
		}
		r.ch = ch
		r.init = true
	}
	return nil
}

func (e *ecuService) startlisteners() {
	var wg sync.WaitGroup

	wg.Add(len(e.receivers))

	multiplex := func(l *listenerch) {
		defer wg.Done()
	loop:
		for {
			select {
			case <-l.done:
				break loop
			case msg := <-l.ch:
				e.incoming <- &incoming{l.name, msg}
			}
		}
	}

	for _, r := range e.receivers {
		go multiplex(r)
	}

	go func() {
		wg.Wait()
		close(e.incoming)
	}()
}

// TODO
func (e *ecuService) closeReceivers() {}

func (e *ecuService) listen() {
	go func() {
		for {
			for m := range e.incoming {
				switch m.name {
				case config.Sn:
					fmt.Printf("Received Sn: %s", m.msg)
				case config.Vin:
					fmt.Printf("Received Vin: %s", m.msg)
				case config.Rekey:
					fmt.Printf("Received Rekey: %s", m.msg)
				case config.Nonce:
					fmt.Printf("Received Nonce: %s", m.msg)
				case config.SendSn:
					fmt.Printf("Received SendSn: %s", m.msg)
				case config.Join:
					fmt.Printf("Received Join: %s", m.msg)
				default:
					// TODO: handle normal message communication
				}
			}
		}
	}()
}

// NewLeader returns a new leader ecu
func NewLeader() (Leader, error) {
	return newService(leaderConfig())
}

// NewMember returns a new member ecu
func NewMember() (Member, error) {
	return newService(memberConfig())
}
