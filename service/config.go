package service

import (
	"github.com/sauravgsh16/ecu/handler"
)

const (
	leader = iota
	member
)

type ecuConfig struct {
	ecuType      int
	broadcasters []func() (handler.Sender, error)
	subscribers  []func(id string) (handler.Receiver, error)
	senders      []func(id string) (handler.Sender, error)
	receivers    []func() (handler.Receiver, error)
}

func leaderConfig() *ecuConfig {
	return &ecuConfig{
		ecuType: leader,
		broadcasters: []func() (handler.Sender, error){
			handler.NewSnAnnouncer,
			handler.NewVinAnnouncer,
			handler.NewNonceAnnouncer,
		},
		subscribers: []func(id string) (handler.Receiver, error){
			handler.NewRekeySubscriber,
			handler.NewNonceSubscriber,
		},
		senders: []func(id string) (handler.Sender, error){
			handler.NewSendSnSender,
		},
		receivers: []func() (handler.Receiver, error){
			handler.NewJoinReceiver,
		},
	}
}

func memberConfig() *ecuConfig {
	return &ecuConfig{
		ecuType: member,
		broadcasters: []func() (handler.Sender, error){
			handler.NewNonceAnnouncer,
			handler.NewRekeyAnnouncer,
		},
		subscribers: []func(id string) (handler.Receiver, error){
			handler.NewNonceSubscriber,
			handler.NewSnSubscriber,
			handler.NewVinSubscriber,
		},
		senders: []func(id string) (handler.Sender, error){
			handler.NewJoinSender,
		},
		receivers: []func() (handler.Receiver, error){
			handler.NewSendSnReceiver,
		},
	}
}
