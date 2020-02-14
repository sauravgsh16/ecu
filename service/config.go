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
	subscribers  []func(string) (handler.Receiver, error)
	senders      []func(string) (handler.Sender, error)
	receivers    []func(string) (handler.Receiver, error)
}

func leaderConfig() *ecuConfig {
	return &ecuConfig{
		ecuType: leader,
		broadcasters: []func() (handler.Sender, error){
			handler.NewSnAnnouncer,
			handler.NewVinAnnouncer,
			handler.NewNonceAnnouncer,
			handler.NewRekeyAnnouncer,
		},
		subscribers: []func(string) (handler.Receiver, error){
			handler.NewSnSubscriber,
			handler.NewVinSubscriber,
			handler.NewNonceSubscriber,
			handler.NewRekeySubscriber,
		},
		senders: []func(string) (handler.Sender, error){
			handler.NewSendSnSender,
		},
		receivers: []func(string) (handler.Receiver, error){
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
		subscribers: []func(string) (handler.Receiver, error){
			handler.NewSnSubscriber,
			handler.NewVinSubscriber,
			handler.NewNonceSubscriber,
			handler.NewRekeySubscriber,
		},
		senders: []func(string) (handler.Sender, error){
			handler.NewJoinSender,
		},
		receivers: []func(string) (handler.Receiver, error){
			handler.NewSendSnReceiver,
		},
	}
}
