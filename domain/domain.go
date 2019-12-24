package domain

import (
	"crypto/rand"
	"errors"
)

const (
	// SnSize byte size
	SnSize = 16

	Leader = iota
	Member
)

var (
	ErrInvalidEcuKind           = errors.New("invalid Ecu kind")
	ErrInvalidSn                = errors.New("invalid Sn")
	ErrUnauthorizedSnGeneration = errors.New("Ecu unauthorized to generate Sn")
	ErrResetSn                  = errors.New("non-zero Sn tried to reset it's value")
)

// Ecu struct
type Ecu struct {
	VIN string
	// Network key
	Sn []byte
	// EncKey -Sv-ENC - Ecu key for encryption
	EncKey []byte
	// Sv-MAC - Ecu key for generation MAC (also known as tag)
	MacKey []byte
	// Ecu Kind: leader or member
	Kind int
}

// NewEcu returns a new Ecu with Sn = 0
// Complete Sn is a combination of Sn1 and Sn2
func NewEcu(kind int) (*Ecu, error) {
	v := &Ecu{
		EncKey: make([]byte, 64),
		MacKey: make([]byte, 64),
		Sn:     make([]byte, 16),
	}
	if _, err := rand.Read(v.EncKey); err != nil {
		return nil, err
	}
	if _, err := rand.Read(v.MacKey); err != nil {
		return nil, err
	}

	switch kind {
	case Leader:
		v.Kind = Leader
	case Member:
		v.Kind = Member
	default:
		return nil, ErrInvalidEcuKind
	}

	return v, nil
}

// GenerateSn generates and sets the Ecu Sn
func (v *Ecu) GenerateSn() error {
	if v.Kind != Leader {
		return ErrUnauthorizedSnGeneration
	}

	if !checkZero(v.Sn) {
		return ErrResetSn
	}

	sn, err := GenerateRandom(SnSize)
	if err != nil {
		return err
	}

	copy(v.Sn, []byte(sn))
	return nil
}

// GetSn returns the Ecu Sn string
func (v *Ecu) GetSn() string {
	if checkZero(v.Sn) {
		return "0"
	}
	return string(v.Sn)
}

func checkZero(b []byte) bool {
	for _, v := range b {
		if v != 0 {
			return false
		}
	}
	return true
}

// Message struct
type Message struct {
	ID                   int
	FreshnessValue       int32
	EncryptBits          int64
	BlkCount             int64
	PlainText            []byte
	Nonce                []byte
	InitializationVector []byte
}

// NewMessage returns a new Message
func NewMessage() *Message {
	return &Message{
		PlainText:            make([]byte, 16),
		Nonce:                make([]byte, 4),
		InitializationVector: make([]byte, 8),
	}
}

func (m *Message) CreateNonce() {

}
