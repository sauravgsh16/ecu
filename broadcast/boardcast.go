package broadcast

import (
	"errors"

	"github.com/sauravgsh16/ecu/client"
	"github.com/sauravgsh16/ecu/config"
)

const (
	initialized = true
)

// broadcast message handler
type Broadcaster struct {
	ID     int64
	ExName string
	noWait bool
	status bool
	client client.MessageServer
}

func (bs *Broadcaster) init(ex string) error {
	if bs.status == initialized {
		return errors.New("init on uninitialized broadcaster")
	}

	exID, err := bs.client.DeclareExchange(ex, config.BroadCastType, bs.noWait)
	if err != nil {
		return err
	}

	bs.ID = exID
	bs.ExName = ex
	bs.status = initialized
	return nil
}

// GetExchangeName of the broadcaster
func (bs *Broadcaster) GetExchangeName() string {
	return bs.ExName
}

// New returns a new broadcaster
func New(reqName string, noWait bool, c client.MessageServer) (*Broadcaster, error) {
	bs := &Broadcaster{
		noWait: noWait,
		client: c,
	}

	if err := bs.init(reqName); err != nil {
		return nil, err
	}
	return bs, nil
}
