package servicenew

import (
	"bytes"
	"fmt"
	"sync"

	"github.com/hashicorp/go-multierror"

	"github.com/sauravgsh16/ecu/client"
	"github.com/sauravgsh16/ecu/config"
	"github.com/sauravgsh16/ecu/domain"
	"github.com/sauravgsh16/ecu/handler"
)

const (
	errInvalidJoinRequest = "join request sent - invalid"
	errSendSnRegister     = "failed to register Send Sn handler"
	errJoinRegister       = "failed to register Join handler"

	// Key Names
	appKey = "ApplicationID"
)

// Leader interface
type Leader interface {
	AnnounceSn() error
	AnnounceVin() error
	SendSn(string)
}

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
	domain       *domain.Ecu
	broadcasters map[string]handler.Sender
	subscribers  map[string]*listenerch
	senders      map[string]handler.Sender
	receivers    map[string]*listenerch
	certs        map[string][]byte
	certLoaded   bool
	certMux      sync.Mutex
	mux          sync.RWMutex
	incoming     chan *incoming
	p2pincoming  chan *incoming
	done         chan interface{}
}

func newService(c *ecuConfig) (*ecuService, error) {
	d, err := domain.NewEcu(c.ecuType)
	if err != nil {
		return nil, err
	}

	e := &ecuService{
		domain:       d,
		broadcasters: make(map[string]handler.Sender),
		subscribers:  make(map[string]*listenerch),
		senders:      make(map[string]handler.Sender),
		receivers:    make(map[string]*listenerch),
		incoming:     make(chan *incoming),
		p2pincoming:  make(chan *incoming),
		done:         make(chan interface{}),
	}
	if c.leader {
		if err := e.loadCerts(); err != nil {
			return nil, err
		}

		if err := e.domain.GenerateSn(); err != nil {
			return nil, err
		}
	}

	e.mux.Lock()
	defer e.mux.Unlock()

	// Register broadcasters
	for _, h := range c.broadcasters {
		b, err := h()
		if err != nil {
			return nil, err
		}
		e.broadcasters[b.GetName()] = b
	}

	// Register receivers
	for _, h := range c.subscribers {
		s, err := h()
		if err != nil {
			return nil, err
		}
		e.receivers[s.GetName()] = &listenerch{
			h:    s,
			done: make(chan interface{}),
			name: s.GetName(),
		}
	}

	if err := e.startreceivers(); err != nil {
		return nil, err
	}

	e.startlisteners()

	go e.listen()
	go e.handleIncoming()

	return e, nil
}

func (e *ecuService) handleIncoming() {
	for {
		select {
		case <-e.done:
			return
		case i := <-e.p2pincoming:
			switch i.name {

			case config.Join:
				go e.handleJoin(i.msg)

			case config.SendSn:
				go e.handleSn(i.msg)

			default:
				panic("unknown type")
			}
		}
	}
}

func (e *ecuService) startreceivers() error {
	var err error

	if len(e.receivers) == 0 {
		return nil
	}

	for _, r := range e.subscribers {
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

	wg.Add(len(e.subscribers))

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

	for _, s := range e.subscribers {
		go multiplex(s)
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
			for i := range e.incoming {
				switch i.name {
				case config.Sn:
					fmt.Printf("Received Sn: %+v", i.msg)

				case config.Vin:
					fmt.Printf("Received Vin: %+v", i.msg)

				case config.Rekey:
					fmt.Printf("Received Rekey: %+v", i.msg)

				case config.Nonce:
					fmt.Printf("Received Nonce: %+v", i.msg)

				default:
					e.p2pincoming <- i
				}
			}
		}
	}()
}

func (e *ecuService) aggregateCertNone() ([]byte, error) {
	var err error
	var buf bytes.Buffer

	for _, c := range e.domain.Certs {
		if _, wErr := buf.Write(c); wErr != nil {
			err = multierror.Append(err, wErr)
		}
	}

	if _, wErr := buf.Write(e.domain.GetNonce()); wErr != nil {
		err = multierror.Append(err, wErr)
	}

	return buf.Bytes(), err
}

// NewLeader returns a new leader ecu
func NewLeader() (Leader, error) {
	return newService(leaderConfig())
}

// NewMember returns a new member ecu
func NewMember() (Member, error) {
	return newService(memberConfig())
}
