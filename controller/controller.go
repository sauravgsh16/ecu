package controller

import (
	"github.com/sauravgsh16/ecu/service"
)

const (
	leader int = iota
	member
)

// Controller interface
type Controller interface {
	Initiate() error
}

type controller struct {
	service service.ECU
}

// New returns a new controller
func New(kind int) (Controller, error) {
	var err error
	c := &controller{}

	switch kind {
	case leader:
		c.service, err = service.NewLeader()
		if err != nil {
			return nil, err
		}
	case member:
		c.service, err = service.NewMember()
		if err != nil {
			return nil, err
		}
	default:
		panic("unknown ECU type")
	}
	return c, nil
}

func (c *controller) Initiate() error {
	switch t := c.service.(type) {
	case service.Leader:
		return t.AnnounceSn()
	case service.Member:
		return t.AnnounceRekey()
	}
	return nil
}
