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
	return newBroadcastSender(snExName, config.Sn)
}

// NewVinAnnouncer returns a new Vin broadcaster
func NewVinAnnouncer() (Sender, error) {
	return newBroadcastSender(vinExName, config.Vin)
}

// NewRekeyAnnouncer returns a new start Rekey broadcaster
func NewRekeyAnnouncer() (Sender, error) {
	return newBroadcastSender(rkExName, config.Rekey)
}

// NewNonceAnnouncer returns a new Nonce broadcaster
func NewNonceAnnouncer() (Sender, error) {
	return newBroadcastSender(nonceExName, config.Nonce)
}

// NewSnSubscriber returns a new sn subscriber
func NewSnSubscriber(id string) (Receiver, error) {
	return newBroadcastReceiver(snExName, util.JoinString(snQName, id), snConsumerName, config.Sn)
}

// NewVinSubscriber returns a new Vin subscriber
func NewVinSubscriber(id string) (Receiver, error) {
	return newBroadcastReceiver(vinExName, util.JoinString(vinQName, id), vinConsumerName, config.Vin)
}

// NewRekeySubscriber returns a new start rekey subscriber
func NewRekeySubscriber(id string) (Receiver, error) {
	return newBroadcastReceiver(rkExName, util.JoinString(rkQName, id), rkConsumerName, config.Rekey)
}

// NewNonceSubscriber returns a new nonce subscriber
func NewNonceSubscriber(id string) (Receiver, error) {
	return newBroadcastReceiver(nonceExName, util.JoinString(nonceQName, id), nonceConsumerName, config.Nonce)
}

// NewSendSnSender returns a new 'send sn' sender
func NewSendSnSender(appID string) (Sender, error) {
	return newPeerSender(sendSnQName, util.JoinString(config.SendSn, appID))
}

// NewJoinSender returns a new 'join' sender
func NewJoinSender(appID string) (Sender, error) {
	return newPeerSender(joinQName, util.JoinString(config.Join, appID))
}

// NewSendSnReceiver returns a new 'send sn' receiver
func NewSendSnReceiver() (Receiver, error) {
	return newPeerReceiver(sendSnQName, sendSnConsumerName, config.SendSn)
}

// NewJoinReceiver returns a new 'join' receiver
func NewJoinReceiver() (Receiver, error) {
	return newPeerReceiver(joinQName, joinConsumerName, config.Join)
}

func newBroadcastSender(exName, senderName string) (Sender, error) {
	bp := client.BroadCastPublish{
		ExchangeName:   exName,
		ExchangeNoWait: pubNoWait,
		Immediate:      immediate,
	}
	config := bp.Marshal()

	publisher, err := client.NewPublisher(config)
	if err != nil {
		return nil, err
	}

	return &send{p: publisher, name: senderName}, nil
}

func newBroadcastReceiver(exName, qName, consumerName, receiverName string) (Receiver, error) {
	bs := client.BroadcastSubscribe{
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
	return &receive{
		s:    sub,
		ctx:  ctx,
		name: receiverName,
	}, nil
}

func newPeerSender(qName, senderName string) (Sender, error) {
	ps := client.PeerSend{
		QueueName:   qName,
		QueueNoWait: queueNoWait,
		Immediate:   immediate,
	}

	config := ps.Marshal()

	publisher, err := client.NewPublisher(config)
	if err != nil {
		return nil, err
	}

	return &send{p: publisher, name: senderName}, nil
}

func newPeerReceiver(qName, consumerName, receiverName string) (Receiver, error) {
	pr := client.PeerReceive{
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
	return &receive{
		s:    receiver,
		ctx:  ctx,
		name: receiverName,
	}, nil
}
