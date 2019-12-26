package service

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"sync"

	"github.com/sauravgsh16/ecu/config"
	"github.com/sauravgsh16/ecu/domain"
	"github.com/sauravgsh16/ecu/request"
	"github.com/sauravgsh16/ecu/util"
)

const (
	leader = iota
	member

	initialized = 1
	broadcast   = "fanout"
	peer        = "direct"
)

// Leader interface ....
type Leader interface {
	AnnounceSN() error
	AnnounceVinCert() error
}

// Member interface ....
type Member interface{}

type ecuService struct {
	domain   *domain.Ecu
	SN       []byte
	mux      sync.RWMutex
	emitter  *broadcastSender
	consumer *broadcastConsumer
	certs    map[string][]byte
	// TODO: Add senders and receivers
}

func (e *ecuService) init() error {
	if err := e.domain.GenerateSn(); err != nil {
		return err
	}

	e.mux.Lock()
	defer e.mux.Unlock()

	e.SN = []byte(e.domain.GetSn())

	return nil
}

func (e *ecuService) AnnounceSN() error {
	hashSn := util.GenerateHash(e.SN)

	if err := e.emitter.broadcast([]byte(hashSn)); err != nil {
		return err
	}

	return nil
}

func (e *ecuService) AnnounceVinCert() error {
	if err := e.loadCerts(); err != nil {
		return err
	}

	for _, v := range e.certs {
		if err := e.emitter.broadcast(v); err != nil {
			return err
		}
	}
	return nil
}

func (e *ecuService) loadCerts() error {
	e.mux.Lock()
	defer e.mux.Unlock()

	for _, c := range config.DefaultBroadCastCertificates {
		b, err := ioutil.ReadFile(filepath.Join(config.DefaultCertificateLocation, c))
		if err != nil {
			return fmt.Errorf("failed to load certificate: %s", err.Error())
		}
		e.certs[c] = b
	}
	return nil
}

func (e *ecuService) HandleJoin(req *request.JoinRequest) error {

	return nil
}

// TODO : FIND WAY TO INITIALIZE HANDLERS
// Need to decide how many handler initializations will be required

// NewMember returns a new Member ECU
func NewMember() (Member, error) {
	d, err := domain.NewEcu(member)
	if err != nil {
		return nil, err
	}

	return &ecuService{
		domain: d,
		certs:  make(map[string][]byte, 0),
	}, nil
}

// NewLeader returns a new Leader ECU
func NewLeader() (Leader, error) {
	d, err := domain.NewEcu(leader)
	if err != nil {
		return nil, err
	}
	l := &ecuService{
		domain: d,
		certs:  make(map[string][]byte, 0),
	}

	if err := l.init(); err != nil {
		return nil, err
	}
	return l, nil
}
