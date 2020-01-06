package domain

import (
	"bytes"
	"crypto/rand"
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"sync"

	"github.com/sauravgsh16/ecu/config"
)

const (
	// Leader ECU type
	Leader = iota
	// Member ECU type
	Member
	// SnSize byte size
	SnSize = 16
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
	// Nonce - unique identification
	Nonce string
	// Network key
	Sn []byte
	// EncKey -Sv-ENC - Ecu key for encryption
	EncKey []byte
	// Sv-MAC - Ecu key for generation MAC (also known as tag)
	MacKey []byte
	// Ecu Kind: leader or member
	Kind int
	// Location where certificates are present
	CertLoc string
	// Cert bytes
	Certs map[string][]byte
	mux   sync.Mutex
}

// NewEcu returns a new Ecu with Sn = 0
func NewEcu(kind int) (*Ecu, error) {
	e := &Ecu{
		EncKey:  make([]byte, 64),
		MacKey:  make([]byte, 64),
		Sn:      make([]byte, 16),
		CertLoc: config.DefaultCertificateLocation,
		Certs:   make(map[string][]byte, 0),
	}
	if _, err := rand.Read(e.EncKey); err != nil {
		return nil, err
	}
	if _, err := rand.Read(e.MacKey); err != nil {
		return nil, err
	}

	switch kind {
	case Leader:
		e.Kind = Leader
	case Member:
		e.Kind = Member
	default:
		return nil, ErrInvalidEcuKind
	}

	if err := e.loadCerts(); err != nil {
		return nil, err
	}

	if err := e.generateNonce(); err != nil {
		return nil, err
	}

	return e, nil
}

func (e *Ecu) generateNonce() error {
	b := make([]byte, 16)
	buf := bytes.NewBuffer(b)

	if _, err := rand.Read(buf.Bytes()); err != nil {
		return err
	}

	e.Nonce = string(buf.Bytes())
	return nil
}

// GenerateSn generates and sets the Ecu Sn
func (e *Ecu) GenerateSn() error {
	if e.Kind != Leader {
		return ErrUnauthorizedSnGeneration
	}

	if !checkZero(e.Sn) {
		return ErrResetSn
	}

	sn, err := GenerateRandom(SnSize)
	if err != nil {
		return err
	}

	copy(e.Sn, []byte(sn))
	return nil
}

// GetSn returns the Ecu Sn string
func (e *Ecu) GetSn() string {
	if checkZero(e.Sn) {
		return "0"
	}
	return string(e.Sn)
}

func (e *Ecu) loadCerts() error {
	e.mux.Lock()
	defer e.mux.Unlock()

	for _, c := range config.DefaultCertificates {
		b, err := ioutil.ReadFile(filepath.Join(config.DefaultCertificateLocation, c))
		if err != nil {
			return fmt.Errorf("failed to load certificates, %s", err.Error())
		}
		e.Certs[c] = b
	}
	return nil
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

func (m *Message) CreateNonce() {}
