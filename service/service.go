package service

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"regexp"
	"sync"
	"time"

	"github.com/hashicorp/go-multierror"

	"github.com/sauravgsh16/can-interface"
	"github.com/sauravgsh16/ecu/client"
	"github.com/sauravgsh16/ecu/config"
	"github.com/sauravgsh16/ecu/domain"
	"github.com/sauravgsh16/ecu/handler"
	"github.com/sauravgsh16/ecu/util"
)

const (
	// Key Names
	appKey        = "ApplicationID"
	contentType   = "ContentType"
	rekeytimeout  = 15
	defaultvbsURL = "tcp://localhost:19000"
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
	GetID() string
	SetID(string)
	CreateUnicastHandlers(chan string, chan error, int)
}

type rekeytimer struct {
	timer *time.Timer
	end   time.Time
}

func newRekeyTimer(t time.Duration) *rekeytimer {
	return &rekeytimer{
		timer: time.NewTimer(t),
		end:   time.Now().Add(t),
	}
}

func (r *rekeytimer) Reset(t time.Duration) {
	r.timer.Reset(t)
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

type service interface {
	getID() string
	setID(string)
}

type swService struct {
	domain  *domain.Ecu
	rktimer *rekeytimer
	first   bool
	repeat  bool
	joinCh  chan bool
	idCh    chan bool
}

func (sw *swService) getID() string {
	return sw.domain.ID
}

func (sw *swService) setID(id string) {
	sw.domain.SetID(id)

	select {
	case sw.idCh <- true:
	}
}

type ecuService struct {
	s            service
	broadcasters map[string]handler.Sender
	senders      map[string]handler.Sender
	subscribers  map[string]*listenerch
	receivers    map[string]*listenerch
	incoming     chan *incoming
	unicastCh    chan *incoming
	done         chan interface{}
	idCh         chan bool
	flagMux      sync.RWMutex
	mux          sync.RWMutex
	unicastRe    *regexp.Regexp
	wg           sync.WaitGroup
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func initEcu(s service, c *ecuConfig) error {
	var err error

	switch t := s.(type) {
	case *swService:
		if t.domain, err = domain.New(c.ecuType); err != nil {
			return err
		}
		t.first = true

	case *hwService:
		t.Incoming = make(chan can.DataHolder)
		if t.can, err = can.New(defaultvbsURL, t.Incoming); err != nil {
			return err
		}

	default:
		return errors.New("unknow type - service")
	}

	return nil
}

func (e *ecuService) initializeFields() {
	e.broadcasters = make(map[string]handler.Sender)
	e.subscribers = make(map[string]*listenerch)
	e.senders = make(map[string]handler.Sender)
	e.receivers = make(map[string]*listenerch)
	e.incoming = make(chan *incoming)
	e.unicastCh = make(chan *incoming)
	e.done = make(chan interface{})
	e.idCh = make(chan bool)
	e.unicastRe = regexp.MustCompile(`^(?P<type>.*?)\.`)
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

func (e *ecuService) init(c *ecuConfig, done chan error) {
	var err error

	go func(er error) {
		select {
		case <-e.idCh:
			// Register broadcasters
			for _, h := range c.broadcasters {
				b, err := h()
				if err != nil {
					er = err
				}
				e.broadcasters[b.GetName()] = b
			}

			// Register subscribers
			for _, h := range c.subscribers {
				s, err := h(e.s.getID())
				if err != nil {
					er = err
				}
				e.subscribers[s.GetName()] = &listenerch{
					h:    s,
					done: make(chan interface{}),
					name: s.GetName(),
				}
			}

			if err = e.startconsumers(); err != nil {
				er = err
			}

			e.multiplexlisteners()
			done <- er
		}
		close(done)
	}(err)
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

func (e *ecuService) CreateUnicastHandlers(watcher chan string, errCh chan error, kind int) {
	switch kind {
	case leader:
		go func() {
		loop:
			for {
				select {
				case id := <-watcher:
					if err := e.createSender(id, config.SendSn, handler.NewSendSnSender); err != nil {
						log.Fatalf(err.Error())
					}

					if err := e.createReceiver(id, handler.NewJoinReceiver); err != nil {
						log.Fatalf(err.Error())
					}

				case err := <-errCh:
					log.Printf(err.Error())
					break loop
				}
			}
		}()
	case member:
		go func() {
			select {
			case <-watcher:
				if err := e.createReceiver(e.s.getID(), handler.NewSendSnReceiver); err != nil {
					log.Fatalf(err.Error())
				}
				if err := e.createSender(e.s.getID(), config.Join, handler.NewJoinSender); err != nil {
					log.Fatalf(err.Error())
				}

				switch t := e.s.(type) {
				case *swService:
					t.joinCh <- true
				case *hwService:
					t.joinCh <- true
				}

			case err := <-errCh:
				log.Fatalf(err.Error())
			}
		}()
	}
}

func (e *ecuService) generateMessage(payload []byte) *client.Message {
	return &client.Message{
		Payload: client.Payload(payload),
		Metadata: client.Metadata(map[string]interface{}{
			"ApplicationID": e.s.getID(),
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

	for _, c := range e.s.(*swService).domain.Certs {
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

func (e *ecuService) GetID() string {
	return e.s.getID()
}

func (e *ecuService) SetID(id string) {
	e.s.setID(id)
}

func (e *ecuService) handleRekey(msg *client.Message) {

	log.Printf("Received rekey from AppID:\n")
	fmt.Printf("%#v\n", msg.Metadata.Get("ApplicationID"))
	fmt.Printf("%#v\n", msg.Metadata.Get("Exchange"))
	fmt.Printf("%#v\n\n\n\n", msg.Payload)

	e.flagMux.Lock()
	defer e.flagMux.Unlock()

	e.s.(*swService).first = false
	e.s.(*swService).repeat = true

	if e.s.(*swService).rktimer != nil && e.s.(*swService).rktimer.Running() {
		e.s.(*swService).rktimer.Reset(rekeytimeout * time.Second)
		return
	}
	e.startTimerToGatherNonce()
}

func (e *ecuService) handleReceiveNonce(msg *client.Message) error {
	t := e.s.(*swService)

	log.Printf("Received Nonce From:\n")
	fmt.Printf("%#v\n", msg.Metadata.Get("ApplicationID"))
	fmt.Printf("%#v\n", msg.Metadata.Get("Exchange"))
	fmt.Printf("%#v\n\n\n\n", msg.Payload)

	ctype, err := msg.Metadata.Verify(contentType)
	if err != nil && ctype != "nonce" {
		return errInvalidContentType
	}

	appID, err := msg.Metadata.Verify(appKey)
	if err != nil {
		return errAppIDNotFound
	}

	t.domain.AddToNonceTable(appID, msg.Payload)

	if t.rktimer != nil && t.rktimer.Running() {
		if t.repeat {
			if err := e.annouceNonce(); err != nil {
				return err
			}
			return nil
		}
		t.rktimer.Reset(rekeytimeout * time.Second)
		return nil
	}
	return e.AnnounceRekey()
}

// AnnounceRekey announces rekey to the networks
func (e *ecuService) AnnounceRekey() error {
	t := e.s.(*swService)
	if t.first {
		h, ok := e.broadcasters[config.Rekey]
		if !ok {
			return fmt.Errorf("announce rekey handler not found")
		}

		empty := make([]byte, 8)
		rkmsg := e.generateMessage(empty)
		rkmsg.Metadata.Set(contentType, "rekey")

		log.Printf("Sending Rekey From AppID - %s\n", t.domain.ID)

		if err := h.Send(rkmsg); err != nil {
			return err
		}
	}

	return e.annouceNonce()
}

func (e *ecuService) annouceNonce() error {
	t := e.s.(*swService)
	if !t.first && !t.rktimer.Running() {
		// This logic is added here to handle case :-
		// when we receive "my_nonce" for other ecu, and the rekey timer
		// for this ecu has stopped. We thus need to generate a new nonce
		// and use this newly created nonce for sending.
		if err := t.domain.ResetNonce(); err != nil {
			log.Printf("error resetting my_nonce: %s", err.Error())
		}
	}

	nonce := t.domain.GetNonce()
	msg := e.generateMessage(nonce)

	// Set headers
	msg.Metadata.Set(contentType, "nonce")
	msg.Metadata.Set(appKey, t.domain.ID)

	h, ok := e.broadcasters[config.Nonce]
	if !ok {
		return fmt.Errorf("announce nonce handler not found")
	}

	// TODO: proper logging
	log.Printf("Broadcasting nonce from AppID - %s\n", t.domain.ID)

	if err := h.Send(msg); err != nil {
		return err
	}

	e.flagMux.Lock()
	defer e.flagMux.Unlock()

	t.repeat = false

	if t.rktimer != nil && t.rktimer.Running() {
		// Case when - if we reach this code block
		// from 'repeat nonce', we know that we have
		// rekey timeout goroutine running, we just need
		// to reset the time
		t.rktimer.Reset(rekeytimeout * time.Second)
		return nil
	}

	e.startTimerToGatherNonce()
	return nil
}

func (e *ecuService) startTimerToGatherNonce() {
	t := e.s.(*swService)
	if t.rktimer != nil && t.rktimer.Running() {
		t.rktimer.timer.Stop()
	}
	t.rktimer = newRekeyTimer(rekeytimeout * time.Second)
	e.listenTimer()
}

func (e *ecuService) listenTimer() {
	t := e.s.(*swService)
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		counter := 0
	loop:
		for {
			select {
			case <-t.rktimer.timer.C:
				go e.handleSvCreation()
				break loop
			case <-ticker.C:
				counter++
				fmt.Printf("ticking: %d\n", counter)
			}
		}
		ticker.Stop()
	}()
}

func (e *ecuService) handleSvCreation() {
	e.flagMux.Lock()
	defer e.flagMux.Unlock()

	e.s.(*swService).first = true
	e.s.(*swService).repeat = false

	nonceAll := e.s.(*swService).domain.GetNonceAll()
	fmt.Printf("NONCE ALL: %#v\n", nonceAll)

	e.s.(*swService).domain.CalculateSv()
	e.s.(*swService).domain.ClearNonceTable()
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
func NewEcu(kind int, sim bool, initCh chan bool) (ECU, error) {
	switch kind {
	case leader:
		if sim {
			return newLeader(leaderConfig(), initCh)
		}
		return newLeaderHW(leaderConfig(), initCh)
	case member:
		if sim {
			return newMember(memberConfig(), initCh)
		}
		return newMemberHW(memberConfig(), initCh)
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
	case *LeaderEcuHW:
		t.StartListeners()
	case *MemberEcuHW:
		t.StartListeners()
	}
}
