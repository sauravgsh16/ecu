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
)
