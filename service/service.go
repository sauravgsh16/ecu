package service

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"regexp"
	"sync"
	"time"

	"github.com/hashicorp/go-multierror"

	"github.com/gofrs/uuid"

	"github.com/sauravgsh16/ecu/client"
	"github.com/sauravgsh16/ecu/config"
	"github.com/sauravgsh16/ecu/domain"
	"github.com/sauravgsh16/ecu/handler"
	"github.com/sauravgsh16/ecu/util"
)

const (
	// Key Names
	appKey      = "ApplicationID"
	contentType = "ContentType"
)

var (
	errInvalidContentType = errors.New("invalid content type")
	errInvalidJoinRequest = errors.New("join request sent - invalid")
	errSendSnRegister     = errors.New("failed to register Send Sn handler")
	errJoinRegister       = errors.New("failed to register Join handler")
	errAppIDNotFound      = errors.New("application id not found")
)

// ECU interface
type ECU interface {
	AnnouceNonce() error
	Close()
	GetDomainID() string
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
	incoming       chan *incoming
	unicastCh      chan *incoming
	done           chan interface{}
	processMux     sync.Mutex
	mux            sync.RWMutex
	formingNetwork bool
	unicastPattern *regexp.Regexp
	wg             sync.WaitGroup
}

// LeaderEcu struct
type LeaderEcu struct {
	ecuService
	certs      map[string][]byte
	certMux    sync.Mutex
	certLoaded bool
}

// MemberEcu struct
type MemberEcu struct {
	ecuService
}

func (e *ecuService) handleReceiveNonce(msg *client.Message) error {
	// TODO: proper logging
	log.Printf("(%s )Received Nonce :- Type - %d\n", time.Now().Format("2006-01-02T15:04:05.999999-07:00"), e.domain.Kind)

	ctype, err := msg.Metadata.Verify(contentType)
	if err != nil && ctype != "nonce" {
		return errInvalidContentType
	}

	appID, err := msg.Metadata.Verify(appKey)
	if err != nil {
		return errAppIDNotFound
	}

	return e.domain.AddToNonceTable(appID, msg.Payload)
}

func (e *ecuService) initializeFields() {
	e.broadcasters = make(map[string]handler.Sender)
	e.subscribers = make(map[string]*listenerch)
	e.senders = make(map[string]handler.Sender)
	e.receivers = make(map[string]*listenerch)
	e.incoming = make(chan *incoming)
	e.unicastCh = make(chan *incoming)
	e.done = make(chan interface{})
	e.unicastPattern = regexp.MustCompile(`^(?P<type>.*?)\.`)
}

// TODO
func (e *ecuService) closeReceivers() {}

// TODO
func (e *ecuService) Close() {}

func (e *ecuService) init(c *ecuConfig) error {
	var err error

	e.domain, err = domain.NewEcu(c.ecuType)
	if err != nil {
		return err
	}

	// Register broadcasters
	for _, h := range c.broadcasters {
		b, err := h()
		if err != nil {
			return err
		}
		e.broadcasters[b.GetName()] = b
	}

	// Register subscribers
	for _, h := range c.subscribers {
		s, err := h(e.domain.ID)
		if err != nil {
			return err
		}
		e.subscribers[s.GetName()] = &listenerch{
			h:    s,
			done: make(chan interface{}),
			name: s.GetName(),
		}
	}

	if err := e.startconsumers(); err != nil {
		return err
	}

	e.multiplexlisteners()

	return nil
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

	return nil
}

func (e *ecuService) multiplex(l *listenerch) {
	defer e.wg.Done()
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

func (e *ecuService) multiplexlisteners() {
	e.wg.Add(len(e.subscribers))

	for _, s := range e.subscribers {
		go e.multiplex(s)
	}

	go func() {
		e.wg.Wait()
		close(e.incoming)
	}()
}

func (e *ecuService) generateMessage(payload []byte) *client.Message {
	uuid := fmt.Sprintf("%s", uuid.Must(uuid.NewV4()))

	return &client.Message{
		UUID:    uuid,
		Payload: client.Payload(payload),
		Metadata: client.Metadata(map[string]interface{}{
			"ApplicationID": e.domain.ID,
		}),
	}
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

func (e *ecuService) createSender(id string, name string, h func(string) (handler.Sender, error)) error {
	e.mux.Lock()
	defer e.mux.Unlock()

	fmt.Printf("HEREEEEEEEEEEEEEEEE ------->%s\n", util.JoinString(name, id))

	_, ok := e.senders[util.JoinString(name, id)]
	if ok {
		return nil
	}

	fmt.Printf("CREATING SENDER ------->%s\n", util.JoinString(name, id))

	s, err := h(id)
	if err != nil {
		return err
	}
	e.senders[s.GetName()] = s
	return nil
}

func (e *ecuService) createReceiver(id string, h func(string) (handler.Receiver, error)) error {
	e.mux.Lock()
	defer e.mux.Unlock()

	r, err := h(id)
	if err != nil {
		return err
	}

	fmt.Printf("HEREEEEEEEEEEEEEEEE -------> Creating receiver\n")

	done := make(chan interface{})

	ch, err := r.StartConsumer(done)
	if err != nil {
		log.Fatalf(err.Error())
	}

	e.receivers[r.GetName()] = &listenerch{
		ch:   ch,
		h:    r,
		done: done,
		name: r.GetName(),
		init: true,
	}

	fmt.Printf("CREATED RECEIVER ------->%s\n", r.GetName())

	e.wg.Add(1)
	go e.multiplex(e.receivers[r.GetName()])
	return nil
}

// NewEcu returns a new ECU
func NewEcu(kind int) (ECU, error) {
	switch kind {
	case leader:
		return newLeader(leaderConfig())
	case member:
		return newMember(memberConfig())
	default:
		panic("unknown ecu kind")
	}
}

// StartListeners accepts the ecu interface and start the appropriate listeners
func StartListeners(e ECU) {
	switch t := e.(type) {
	case *LeaderEcu:
		t.StartListeners()
	case *MemberEcu:
		t.StartListeners()
	}
}

func (e *ecuService) GetDomainID() string {
	return e.domain.ID
}

func (e *ecuService) AnnouceNonce() error {
	nonce := e.domain.GetNonce()
	msg := e.generateMessage(nonce)

	// Set headers
	msg.Metadata.Set(contentType, "nonce")
	msg.Metadata.Set(appKey, e.domain.ID)

	h, ok := e.broadcasters[config.Nonce]
	if !ok {
		return fmt.Errorf("announce nonce handler not found")
	}

	// TODO: proper logging
	log.Printf("Broadcasting nonce from AppID - %s\n", e.domain.ID)

	if err := h.Send(msg); err != nil {
		return err
	}

	return nil
}
