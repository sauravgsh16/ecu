package serv

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"regexp"
	"sync"
	"time"

	"github.com/gofrs/uuid"
	"github.com/hashicorp/go-multierror"
	"github.com/sauravgsh16/ecu/client"
	"github.com/sauravgsh16/ecu/config"
	"github.com/sauravgsh16/ecu/domain"
	"github.com/sauravgsh16/ecu/handler"
	"github.com/sauravgsh16/ecu/util"
)

const (
	// Key Names
	appKey       = "ApplicationID"
	contentType  = "ContentType"
	rekeytimeout = 15
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
	Close()
	GetDomainID() string
}

type rekeytimer struct {
	*time.Timer
	end time.Time
}

func newRekeyTimer(t time.Duration) *rekeytimer {
	return &rekeytimer{time.NewTimer(t), time.Now().Add(t)}
}

func (r *rekeytimer) Reset(t time.Duration) {
	r.Reset(t)
	r.end = time.Now().Add(t)
}

func (r *rekeytimer) Running() bool {
	return r.end.After(time.Now())
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
	rktimer        *rekeytimer
	flagMux        sync.Mutex
	mux            sync.RWMutex
	unicastRe      *regexp.Regexp
	wg             sync.WaitGroup
	formingNetwork bool
	first          bool
	repeat         bool
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

func init() {
	rand.Seed(time.Now().UnixNano())
}

func (e *ecuService) initializeFields() {
	e.broadcasters = make(map[string]handler.Sender)
	e.subscribers = make(map[string]*listenerch)
	e.senders = make(map[string]handler.Sender)
	e.receivers = make(map[string]*listenerch)
	e.incoming = make(chan *incoming)
	e.unicastCh = make(chan *incoming)
	e.done = make(chan interface{})
	e.unicastRe = regexp.MustCompile(`^(?P<type>.*?)\.`)
	e.first = true
}

func (e *ecuService) closeReceivers(wg *sync.WaitGroup) {
	wg.Add(len(e.receivers))
	wg.Add(len(e.subscribers))

	go func() {
		for _, r := range e.receivers {
			select {
			case r.done <- true:
			default:
				log.Printf("done channel blocking - for %s\n", r.name)
			}
			wg.Done()
		}
	}()

	go func() {
		for _, s := range e.subscribers {
			select {
			case s.done <- true:
			default:
				log.Printf("done channel blocking - for %s\n", s.name)
			}
			wg.Done()
		}
	}()
}

func (e *ecuService) closeSenders(wg *sync.WaitGroup) {

	wg.Add(len(e.broadcasters))
	wg.Add(len(e.senders))

	go func() {
		for _, b := range e.broadcasters {
			b.Close()
			wg.Done()
		}
	}()

	go func() {
		for _, s := range e.senders {
			s.Close()
			wg.Done()
		}
	}()
}

// Close calls the appropriate means of closing all the senders and receivers
// attached to the ecu
func (e *ecuService) Close() {
	var wg *sync.WaitGroup

	go e.closeSenders(wg)
	go e.closeReceivers(wg)

	wg.Wait()
}

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

func (e *ecuService) generateNonce() ([]byte, error) {
	b := make([]byte, 16)
	buf := bytes.NewBuffer(b)

	if _, err := rand.Read(buf.Bytes()); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (e *ecuService) aggregateCert() ([]byte, error) {
	var err error
	var buf bytes.Buffer

	for _, c := range e.domain.Certs {
		if _, wErr := buf.Write(c); wErr != nil {
			err = multierror.Append(err, wErr)
		}
	}

	return buf.Bytes(), err
}

func (e *ecuService) createSender(id string, name string, h func(string) (handler.Sender, error)) error {
	e.mux.Lock()
	defer e.mux.Unlock()

	_, ok := e.senders[util.JoinString(name, id)]
	if ok {
		return nil
	}

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

	e.wg.Add(1)
	go e.multiplex(e.receivers[r.GetName()])
	return nil
}

func (e *ecuService) GetDomainID() string {
	return e.domain.ID
}

func (e *ecuService) handleRekey(msg *client.Message) {
	log.Printf("Received rekey from AppID: %s\n", msg.Metadata.Get(appKey))

	e.flagMux.Lock()
	defer e.flagMux.Unlock()

	e.first = false
	e.repeat = true

	if e.rktimer.Running() {
		e.rktimer.Reset(rekeytimeout * time.Second)
	}
}

func (e *ecuService) handleReceiveNonce(msg *client.Message) error {
	log.Printf("Received Nonce :- From - %s\n", msg.Metadata.Get(appKey).(string))

	ctype, err := msg.Metadata.Verify(contentType)
	if err != nil && ctype != "nonce" {
		return errInvalidContentType
	}

	appID, err := msg.Metadata.Verify(appKey)
	if err != nil {
		return errAppIDNotFound
	}

	e.domain.AddToNonceTable(appID, msg.Payload)

	if e.rktimer.Running() {
		if e.repeat {
			if err := e.annouceNonce(); err != nil {
				return err
			}
			return nil
		}
		e.rktimer.Reset(rekeytimeout * time.Second)
		return nil
	}
	return e.AnnounceRekey()
}

// AnnounceRekey announces rekey to the networks
func (e *ecuService) AnnounceRekey() error {
	if e.first {
		/*

			if e.getnetworkformationflag() {
				// Signifies that it has already taken part in n/w formation
				return nil
			}
		*/

		h, ok := e.broadcasters[config.Rekey]
		if !ok {
			return fmt.Errorf("announce rekey handler not found")
		}

		empty := make([]byte, 8)
		rkmsg := e.generateMessage(empty)
		rkmsg.Metadata.Set(contentType, "rekey")

		log.Printf("Sending Rekey From AppID - %s\n", e.domain.ID)

		if err := h.Send(rkmsg); err != nil {
			return err
		}
	}

	return e.annouceNonce()
}

func (e *ecuService) annouceNonce() error {
	if !e.first && !e.rktimer.Running() {
		// This logic is added here to handle case :-
		// when we receive "my_nonce" for other ecu, and the rekey timer
		// for this ecu has stopped. We thus need to generate a new nonce
		// and use this newly created nonce for sending.
		if err := e.domain.ResetNonce(); err != nil {
			log.Printf("error resetting my_nonce: %s", err.Error())
		}
	}

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

	e.flagMux.Lock()
	defer e.flagMux.Unlock()

	e.repeat = false

	if e.rktimer.Running() {
		// Case when - if we reach this code block
		// from 'repeat nonce', we know that we have
		// rekey timeout goroutine running, we just need
		// to reset the time
		e.rktimer.Reset(rekeytimeout * time.Second)
		return nil
	}

	e.startTimerToGatherNonce()
	return nil
}

func (e *ecuService) startTimerToGatherNonce() {
	if e.rktimer != nil {
		e.rktimer.Stop()
	}
	e.rktimer = newRekeyTimer(rekeytimeout * time.Second)
	e.listenTimer()
}

func (e *ecuService) listenTimer() {
	go func() {
		select {
		case <-e.rktimer.C:
			go e.handleSvCreation()
			return
		}
	}()
}

func (e *ecuService) handleSvCreation() {
	e.flagMux.Lock()
	defer e.flagMux.Unlock()

	e.first = true
	e.repeat = false

	e.domain.CalculateSv()
	e.domain.ClearNonceTable()
}

func (e *ecuService) send(s handler.Sender, b []byte, content string) error {
	msg := e.generateMessage(b)
	msg.Metadata.Set(contentType, content)

	if err := s.Send(msg); err != nil {
		return err
	}
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
