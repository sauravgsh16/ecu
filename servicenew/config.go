package servicenew

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
	subscribers  []func() (handler.Receiver, error)
	senders      []func(id string) (handler.Sender, error)
	receivers    []func() (handler.Receiver, error)
	leader       bool
}

func leaderConfig() *ecuConfig {
	return &ecuConfig{
		ecuType: leader,
		broadcasters: []func() (handler.Sender, error){
			handler.NewSnAnnouncer,
			handler.NewVinAnnouncer,
			handler.NewNonceAnnouncer,
		},
		subscribers: []func() (handler.Receiver, error){
			handler.NewRekeyReceiver,
			handler.NewNonceReceiver,
		},
		senders: []func(id string) (handler.Sender, error){
			handler.NewSendSnSender,
		},
		receivers: []func() (handler.Receiver, error){
			handler.NewJoinReceiver,
		},
		leader: true,
	}
}

func memberConfig() *ecuConfig {
	return &ecuConfig{
		ecuType: member,
		broadcasters: []func() (handler.Sender, error){
			handler.NewNonceAnnouncer,
			handler.NewRekeyAnnouncer,
		},
		subscribers: []func() (handler.Receiver, error){
			handler.NewNonceReceiver,
			handler.NewSnReceiver,
			handler.NewVinReceiver,
		},
		senders: []func(id string) (handler.Sender, error){
			handler.NewJoinSender,
		},
		receivers: []func() (handler.Receiver, error){
			handler.NewSendSnReceiver,
		},
	}
}
