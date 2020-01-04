package config

var (
	MessageServerHost            = "tcp://192.168.1.3:9000"
	DefaultCertificateLocation   = "cert"
	DefaultCertificates          = []string{"vin1.der", "vin1_dh.key", "vin2.der", "vin2_dh.key"}
	DefaultBroadCastCertificates = []string{"vin_ca.der", "inter_ca.der", "root_ca.der"}
	BroadCastType                = "fanout"
)
