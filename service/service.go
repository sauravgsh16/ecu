package service

import (
	"sync"

	"github.com/sauravgsh16/ecu/util"

	"github.com/sauravgsh16/ecu/domain"
)

// EcuService interface ....
type EcuService interface{}

type ecuService struct {
	domain *domain.Ecu
	SN     []byte
	mux    sync.RWMutex
}

// NewEcuService returns a new vehicle service
func NewEcuService(kind int) (EcuService, error) {
	d, err := domain.NewEcu(kind)
	if err != nil {
		return nil, err
	}
	vs := &ecuService{domain: d}
	if err = vs.init(); err != nil {
		// ECU is member type
		if err == domain.ErrUnauthorizedSnGeneration && kind == domain.Member {
			return vs, nil
		}
		return nil, err
	}
	return vs, nil
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

func (e *ecuService) AnnounceSn() error {
	_ = util.GenerateHash(e.SN)
	return nil
}
