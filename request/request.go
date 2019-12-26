package request

// JoinRequest struct
type JoinRequest struct {
	EcuID    string
	EcuCert  []byte
	EcuNonce []byte
}
