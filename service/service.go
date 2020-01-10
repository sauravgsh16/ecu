package service

import (
	"bytes"
	"fmt"
	"log"
	"sync"
	"time"

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
	appKey      = "ApplicationID"
	contentType = "ContentType"
)

// Leader interface
type Leader interface {
	AnnounceSn() error
	AnnounceVin() error
	AnnounceNonce() error
	SendSn(string)
}

// Member interface
type Member interface {
	AnnounceRekey() error
	AnnounceNonce() error
	SendJoin(id string)
}

// ECU interface
type ECU interface {
	AnnounceNonce() error
}

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
	domain         *domain.Ecu
	broadcasters   map[string]handler.Sender
	senders        map[string]handler.Sender
	subscribers    map[string]*listenerch
	receivers      map[string]*listenerch
	certs          map[string][]byte
	certLoaded     bool
	incoming       chan *incoming
	p2pincoming    chan *incoming
	done           chan interface{}
	certMux        sync.Mutex
	mux            sync.RWMutex
	processMux     sync.Mutex
	formingNetwork bool
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

	// Register subscribers
	for _, h := range c.subscribers {
		s, err := h()
		if err != nil {
			return nil, err
		}
		e.subscribers[s.GetName()] = &listenerch{
			h:    s,
			done: make(chan interface{}),
			name: s.GetName(),
		}
	}

	// Register receivers
	for _, h := range c.receivers {
		r, err := h()
		if err != nil {
			return nil, err
		}
		e.receivers[r.GetName()] = &listenerch{
			h:    r,
			done: make(chan interface{}),
			name: r.GetName(),
		}
	}

	if err := e.startconsumers(); err != nil {
		return nil, err
	}

	e.multiplexlisteners()

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

func (e *ecuService) startconsumers() error {
	var err error

	if len(e.subscribers) == 0 {
		return nil
	}

	for _, s := range e.subscribers {
		ch, er := s.h.StartConsumer(s.done)
		if er != nil {
			err = multierror.Append(err, er)
		}
		s.ch = ch
		s.init = true
	}

	if len(e.receivers) == 0 {
		return nil
	}

	for _, r := range e.receivers {
		ch, er := r.h.StartConsumer(r.done)
		if er != nil {
			err = multierror.Append(err, er)
		}
		r.ch = ch
		r.init = true
	}

	return nil
}

func (e *ecuService) multiplexlisteners() {
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

// TODO
func (e *ecuService) Close() {}

func (e *ecuService) listen() {
	go func() {
		for {
			for i := range e.incoming {
				switch i.name {
				case config.Sn:
					e.domain.ClearNonceTable()
					go e.handleAnnounceSn(i.msg)

				case config.Vin:
					go e.handleAnnounceVin(i.msg)

				case config.Rekey:
					if e.getnetworkformationflag() {
						continue
					}
					e.domain.ClearNonceTable()
					go e.AnnounceSn()

				case config.Nonce:
					go e.handleReceiveNonce(i.msg)

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

func (e *ecuService) setnetworkformationflag(b bool) {
	e.processMux.Lock()
	defer e.processMux.Unlock()

	e.formingNetwork = b
}

func (e *ecuService) getnetworkformationflag() bool {
	e.processMux.Lock()
	defer e.processMux.Unlock()

	return e.formingNetwork
}

func (e *ecuService) AnnounceNonce() error {
	nonce := e.domain.GetNonce()
	msg := e.generateMessage(nonce)
	msg.Metadata[contentType] = "nonce"
	msg.Metadata[appKey] = e.domain.ID

	h, ok := e.broadcasters[config.Nonce]
	if !ok {
		return fmt.Errorf("announce nonce handler not found")
	}

	log.Printf("(%s) Broadcasting nonce: Type - %d\n", time.Now().Format("2006-01-02T15:04:05.999999-07:00"), e.domain.Kind)

	if err := h.Send(msg); err != nil {
		return err
	}
	return nil
}

func (e *ecuService) handleReceiveNonce(msg *client.Message) error {

	log.Printf("(%s )Received Nonce :- Type - %d\n", time.Now().Format("2006-01-02T15:04:05.999999-07:00"), e.domain.Kind)

	ctype, err := msg.Metadata.Verify(contentType)
	if err != nil && ctype != "nonce" {
		return fmt.Errorf("invalid content type")
	}

	appID, err := msg.Metadata.Verify(appKey)
	if err != nil {
		return fmt.Errorf("application id not found")
	}

	return e.domain.AddToNonceTable(appID, msg.Payload)
}

// NewLeader returns a new leader ecu
func NewLeader() (Leader, error) {
	return newService(leaderConfig())
}

// NewMember returns a new member ecu
func NewMember() (Member, error) {
	return newService(memberConfig())
}
