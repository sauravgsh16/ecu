package handler

import (
	"context"

	"github.com/sauravgsh16/ecu/client"
	"github.com/sauravgsh16/ecu/config"
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
	return newBroadcastSender(snExName)
}

// NewVinAnnouncer returns a new Vin broadcaster
func NewVinAnnouncer() (Sender, error) {
	return newBroadcastSender(vinExName)
}

// NewRekeyAnnouncer returns a new start Rekey broadcaster
func NewRekeyAnnouncer() (Sender, error) {
	return newBroadcastSender(rkExName)
}

// NewNonceAnnouncer returns a new Nonce broadcaster
func NewNonceAnnouncer() (Sender, error) {
	return newBroadcastSender(nonceExName)
}

// NewSnReceiver returns a new sn receiver
func NewSnReceiver() (Receiver, error) {
	return newBroadcastReceiver(snExName, snQName, snConsumerName)
}

// NewVinReceiver returns a new Vin receiver
func NewVinReceiver() (Receiver, error) {
	return newBroadcastReceiver(vinExName, vinQName, vinConsumerName)
}

// NewRekeyReceiver returns a new start rekey receiver
func NewRekeyReceiver() (Receiver, error) {
	return newBroadcastReceiver(rkExName, rkQName, rkConsumerName)
}

// NewNonceReceiver returns a new nonce receiver
func NewNonceReceiver() (Receiver, error) {
	return newBroadcastReceiver(nonceExName, nonceQName, nonceConsumerName)
}

// NewSendSnSender returns a new 'send sn' sender
func NewSendSnSender() (Sender, error) {
	return newPeerSender(sendSnQName)
}

// NewJoinSender returns a new 'join' sender
func NewJoinSender() (Sender, error) {
	return newPeerSender(joinQName)
}

// NewSendSnReceiver returns a new 'send sn' receiver
func NewSendSnReceiver() (Receiver, error) {
	return newPeerReceiver(sendSnQName, sendSnConsumerName)
}

// NewJoinReceiver returns a new 'join' receiver
func NewJoinReceiver() (Receiver, error) {
	return newPeerReceiver(joinQName, joinConsumerName)
}

func newBroadcastSender(exName string) (Sender, error) {
	bp := client.BroadCastPublish{
		URI:            config.MessageServerHost,
		ExchangeName:   exName,
		ExchangeNoWait: pubNoWait,
		Immediate:      immediate,
	}
	config := bp.Marshal()

	publisher, err := client.NewPublisher(config)
	if err != nil {
		return nil, err
	}

	return &send{p: publisher}, nil
}

func newBroadcastReceiver(exName, qName, consumerName string) (Receiver, error) {
	bs := client.BroadcastSubscribe{
		URI:            config.MessageServerHost,
		ExchangeName:   exName,
		ExchangeNoWait: subNoWait,
		QueueName:      qName,
		ConsumerName:   consumerName,
		RoutingKey:     routingKey,
		QueueNoWait:    queueNoWait,
		BindNoWait:     bindNoWait,
		ConsumerNoAck:  consumerNoAck,
		ConsumerNoWait: consumerNoWait,
	}

	config := bs.Marshal()

	sub, err := client.NewSubscriber(config)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	return &receive{sub, ctx, nil}, nil
}

func newPeerSender(qName string) (Sender, error) {
	ps := client.PeerSend{
		URI:         config.MessageServerHost,
		QueueName:   qName,
		QueueNoWait: queueNoWait,
		Immediate:   immediate,
	}

	config := ps.Marshal()

	publisher, err := client.NewPublisher(config)
	if err != nil {
		return nil, err
	}

	return &send{p: publisher}, nil
}

func newPeerReceiver(qName, consumerName string) (Receiver, error) {
	pr := client.PeerReceive{
		URI:            config.MessageServerHost,
		QueueName:      qName,
		ConsumerName:   consumerName,
		ConsumerNoAck:  consumerNoAck,
		ConsumerNoWait: consumerNoWait,
	}
	config := pr.Marshal()

	receiver, err := client.NewSubscriber(config)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	return &receive{receiver, ctx, nil}, nil
}
