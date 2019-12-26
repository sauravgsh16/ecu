package config

var (
	MessageServerHost            = "tcp://localhost:9000"
	DefaultCertificateLocation   = "cert"
	DefaultCertificates          = []string{"vin1.der", "vin1_dh.key", "vin2.der", "vin2_dh.key"}
	DefaultBroadCastCertificates = []string{"vin_ca.der", "inter_ca.der", "root_ca.der"}
)
