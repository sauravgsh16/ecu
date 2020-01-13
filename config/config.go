package config

var (
	DefaultCertificates          = []string{"vin1.der", "vin1_dh.key", "vin2.der", "vin2_dh.key"}
	DefaultBroadCastCertificates = []string{"vin_ca.der", "inter_ca.der", "root_ca.der"}
)

const (
	MessageServerHost = "tcp://localhost:9000"
	SupervisorHost    = "localhost:9090"
	// TODO: Temp -- need to find better way to define directory location
	DefaultCertificateLocation = "cert"

	// Broadcaster and receiver names

	Sn     = "SN"
	Vin    = "VIN"
	Rekey  = "REKEY"
	Nonce  = "NONCE"
	SendSn = "SENDSN"
	Join   = "JOIN"

	// Broadcast
	SnExName    = "announcesn"
	VinExName   = "announceVin"
	RkExName    = "announceRk"
	NonceExName = "annonceNonce"

	SnQName    = "snQueue"
	VinQName   = "vinQueue"
	RkQName    = "rkQueue"
	NonceQName = "nonceQueue"

	SnConsumerName    = "snConsumer"
	VinConsumerName   = "vinConsumer"
	RkConsumerName    = "rkConsumer"
	NonceConsumerName = "nonceConsumer"

	// P2P
	SendSnQName        = "sendSnQueue"
	JoinQName          = "joinQueue"
	SendSnConsumerName = "sendSnConsumer"
	JoinConsumerName   = "joinConsumer"

	// Common
	PubNoWait      = false
	SubNoWait      = false
	QueueNoWait    = false
	BindNoWait     = false
	ConsumerNoWait = false
	ConsumerNoAck  = true
	Immediate      = false
	RoutingKey     = ""
)
