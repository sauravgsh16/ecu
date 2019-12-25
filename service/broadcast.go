package service

import (
	"fmt"

	"github.com/sauravgsh16/ecu/client"
)

// broadcast message handler
type broadcastSender struct {
	exName string
	noWait bool
	exID   int64
	status int
	client client.MessageServer
}

func (bs *broadcastSender) init() error {
	if bs.status == initialized {
		return nil
	}

	exID, err := bs.client.DeclareExchange(bs.exName, broadcast, bs.noWait)
	if err != nil {
		return err
	}

	bs.exID = exID
	bs.status = initialized
	return nil
}

func (bs *broadcastSender) broadcast(body []byte) error {
	if bs.status != initialized {
		return fmt.Errorf("broadcaster yet to be initialized")
	}

	if err := bs.client.BroadCast(bs.exID, bs.noWait, body); err != nil {
		return err
	}
	return nil
}

func newBroadcastSender(exName string, noWait bool, c client.MessageServer) (*broadcastSender, error) {
	bs := &broadcastSender{
		exName: exName,
		noWait: noWait,
		client: c,
	}

	if err := bs.init(); err != nil {
		return nil, err
	}
	return bs, nil
}

// broadcast receive handler
type broadcastConsumer struct {
	exName string
	qName  string
	noWait bool
	exID   int64
	status int
	client client.MessageServer
}

func (bc *broadcastConsumer) init() error {
	if bc.status == initialized {
		return nil
	}

	exID, err := bc.client.DeclareExchange(bc.exName, broadcast, bc.noWait)
	if err != nil {
		return err
	}
	bc.exID = exID

	q, err := bc.client.DeclareQueue(bc.qName, bc.noWait)
	if err != nil {
		return err
	}

	if q == bc.qName {
		return fmt.Errorf("Queue name mismatch")
	}

	if err := bc.client.BindQueue(bc.qName, bc.exID, bc.noWait); err != nil {
		return err
	}

	bc.status = initialized
	return nil
}

func newBroadcastConsumer(exName, qName string, noWait bool, c client.MessageServer) (*broadcastConsumer, error) {
	bc := &broadcastConsumer{
		exName: exName,
		qName:  qName,
		noWait: noWait,
		client: c,
	}

	if err := bc.init(); err != nil {
		return nil, err
	}
	return bc, nil
}
