package handler

import (
	"context"

	"github.com/sauravgsh16/ecu/client"
	"github.com/sauravgsh16/ecu/config"
	"github.com/sauravgsh16/ecu/util"
)

// Broadcaster interface
type Broadcaster interface {
	Broadcast() bool
}

// SenderReceiver interface
type SenderReceiver interface {
	Peer() bool
}

type broadcast struct {
	Sender
	Receiver
}

func (b broadcast) Broadcast() bool { return true }

type peer struct {
	Sender
	Receiver
}

func (sr peer) Peer() bool { return true }

// NewSnAnnouncer returns a new Sn broadcaster
func NewSnAnnouncer() (Sender, error) {
	return newBroadcastSender(config.SnExName, config.Sn)
}

// NewVinAnnouncer returns a new Vin broadcaster
func NewVinAnnouncer() (Sender, error) {
	return newBroadcastSender(config.VinExName, config.Vin)
}

// NewRekeyAnnouncer returns a new start Rekey broadcaster
func NewRekeyAnnouncer() (Sender, error) {
	return newBroadcastSender(config.RkExName, config.Rekey)
}

// NewNonceAnnouncer returns a new Nonce broadcaster
func NewNonceAnnouncer() (Sender, error) {
	return newBroadcastSender(config.NonceExName, config.Nonce)
}

// NewSnSubscriber returns a new sn subscriber
func NewSnSubscriber(id string) (Receiver, error) {
	return newBroadcastReceiver(
		config.SnExName,
		util.JoinString(config.SnQName, id),
		util.JoinString(config.SnConsumerName, id),
		config.Sn,
	)
}

// NewVinSubscriber returns a new Vin subscriber
func NewVinSubscriber(id string) (Receiver, error) {
	return newBroadcastReceiver(
		config.VinExName,
		util.JoinString(config.VinQName, id),
		util.JoinString(config.VinConsumerName, id),
		config.Vin,
	)
}

// NewRekeySubscriber returns a new start rekey subscriber
func NewRekeySubscriber(id string) (Receiver, error) {
	return newBroadcastReceiver(
		config.RkExName,
		util.JoinString(config.RkQName, id),
		util.JoinString(config.RkConsumerName, id),
		config.Rekey,
	)
}

// NewNonceSubscriber returns a new nonce subscriber
func NewNonceSubscriber(id string) (Receiver, error) {
	return newBroadcastReceiver(
		config.NonceExName,
		util.JoinString(config.NonceQName, id),
		util.JoinString(config.NonceConsumerName, id),
		config.Nonce,
	)
}

// NewSendSnSender returns a new 'send sn' sender
func NewSendSnSender(id string) (Sender, error) {
	return newPeerSender(util.JoinString(config.SendSn, id))
}

// NewJoinSender returns a new 'join' sender
func NewJoinSender(id string) (Sender, error) {
	return newPeerSender(util.JoinString(config.Join, id))
}

// NewSendSnReceiver returns a new 'send sn' receiver
func NewSendSnReceiver(id string) (Receiver, error) {
	return newPeerReceiver(util.JoinString(config.SendSn, id), util.JoinString(config.SendSnConsumerName, id))
}

// NewJoinReceiver returns a new 'join' receiver
func NewJoinReceiver(id string) (Receiver, error) {
	return newPeerReceiver(util.JoinString(config.Join, id), util.JoinString(config.JoinConsumerName, id))
}

func newBroadcastSender(exName, name string) (Sender, error) {
	bp := client.BroadCastPublish{
		ExchangeName:   exName,
		ExchangeNoWait: config.PubNoWait,
		Immediate:      config.Immediate,
	}
	config := bp.Marshal()

	publisher, err := client.NewPublisher(config)
	if err != nil {
		return nil, err
	}

	return &send{p: publisher, name: name}, nil
}

func newBroadcastReceiver(exName, qName, consumerName, name string) (Receiver, error) {
	bs := client.BroadcastSubscribe{
		ExchangeName:   exName,
		ExchangeNoWait: config.SubNoWait,
		QueueName:      qName,
		ConsumerName:   consumerName,
		RoutingKey:     config.RoutingKey,
		QueueNoWait:    config.QueueNoWait,
		BindNoWait:     config.BindNoWait,
		ConsumerNoAck:  config.ConsumerNoAck,
		ConsumerNoWait: config.ConsumerNoWait,
	}

	config := bs.Marshal()

	sub, err := client.NewSubscriber(config)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	return &receive{
		s:    sub,
		ctx:  ctx,
		name: name,
	}, nil
}

func newPeerSender(name string) (Sender, error) {
	ps := client.PeerSend{
		QueueName:   name,
		QueueNoWait: config.QueueNoWait,
		Immediate:   config.Immediate,
	}

	config := ps.Marshal()

	publisher, err := client.NewPublisher(config)
	if err != nil {
		return nil, err
	}

	return &send{p: publisher, name: name}, nil
}

func newPeerReceiver(name, consumerName string) (Receiver, error) {
	pr := client.PeerReceive{
		QueueName:      name,
		ConsumerName:   consumerName,
		ConsumerNoAck:  config.ConsumerNoAck,
		ConsumerNoWait: config.ConsumerNoWait,
	}
	config := pr.Marshal()

	receiver, err := client.NewSubscriber(config)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	return &receive{
		s:    receiver,
		ctx:  ctx,
		name: name,
	}, nil
}
