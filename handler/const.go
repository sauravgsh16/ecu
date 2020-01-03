package handler

const (
	// Broadcast
	snExName    = "announcesn"
	vinExName   = "announceVin"
	rkExName    = "announceRk"
	nonceExName = "annonceNonce"

	snQName    = "snQueue"
	vinQName   = "vinQueue"
	rkQName    = "rkQueue"
	nonceQName = "nonceQueue"

	snConsumerName    = "snConsumer"
	vinConsumerName   = "vinConsumer"
	rkConsumerName    = "rkConsumer"
	nonceConsumerName = "nonceConsumer"

	// P2P
	sendSnQName        = "sendSnQueue"
	joinQName          = "joinQueue"
	sendSnConsumerName = "sendSnConsumer"
	joinConsumerName   = "joinConsumer"

	// Common
	pubNoWait      = false
	subNoWait      = false
	queueNoWait    = false
	bindNoWait     = false
	consumerNoWait = false
	consumerNoAck  = true
	immediate      = false
	routingKey     = ""
)
