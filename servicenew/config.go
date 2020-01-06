package servicenew

import (
	"github.com/sauravgsh16/ecu/handler"
)

const (
	leader = iota
	member
)

type ecuConfig struct {
	ecuType   int
	senders   []func() (handler.Sender, error)
	receivers []func() (handler.Receiver, error)
}

func leaderConfig() *ecuConfig {
	return &ecuConfig{
		ecuType: leader,
		senders: []func() (handler.Sender, error){
			handler.NewSnAnnouncer,
			handler.NewVinAnnouncer,
			handler.NewSendSnSender,
			handler.NewNonceAnnouncer,
		},
		receivers: []func() (handler.Receiver, error){
			handler.NewRekeyReceiver,
			handler.NewJoinReceiver,
			handler.NewNonceReceiver,
		},
	}
}

func memberConfig() *ecuConfig {
	return &ecuConfig{
		ecuType: member,
		senders: []func() (handler.Sender, error){
			handler.NewNonceAnnouncer,
			handler.NewRekeyAnnouncer,
			handler.NewJoinSender,
		},
		receivers: []func() (handler.Receiver, error){
			handler.NewNonceReceiver,
			handler.NewSendSnReceiver,
			handler.NewSnReceiver,
			handler.NewVinReceiver,
		},
	}
}
