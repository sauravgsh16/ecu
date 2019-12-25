package service

import (
	"sync"

	"github.com/sauravgsh16/ecu/domain"
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
}

func (e *ecuService) AnnounceSN() error {
	hashSn := util.GenerateHash(e.SN)

	if err := e.emitter.broadcast([]byte(hashSn)); err != nil {
		return err
	}

	return nil
}

func (e *ecuService) AnnounceVinCert() error {
	return nil
}

// TODO : FIND WAY TO INITIALIZE HANDLERS
// Need to decide how many handler initialization required

// NewMember returns a new Member ECU
func NewMember() (Member, error) {
	d, err := domain.NewEcu(member)
	if err != nil {
		return nil, err
	}

	return &ecuService{domain: d}, nil
}

// NewLeader returns a new Leader ECU
func NewLeader() (Leader, error) {
	d, err := domain.NewEcu(leader)
	if err != nil {
		return nil, err
	}
	l := &ecuService{
		domain: d,
	}

	if err := l.init(); err != nil {
		return nil, err
	}
	return l, nil
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
