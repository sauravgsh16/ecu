package domain

import (
	"bytes"
	"crypto/rand"
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"sync"

	"github.com/gofrs/uuid"

	"github.com/sauravgsh16/ecu/config"
)

const (
	// leader ECU type
	leader = iota
	// Member ECU type
	member
)

const (
	snsize    = 16
	noncesize = 16
	encKey    = 64
	mackey    = 64
)

var (
	errInvalidEcuKind           = errors.New("invalid Ecu kind")
	errInvalidSn                = errors.New("invalid Sn")
	errUnauthorizedSnGeneration = errors.New("Ecu unauthorized to generate Sn")
	errResetSn                  = errors.New("non-zero Sn tried to reset it's value")
	errUnAuthorizedSnSet        = errors.New("ecu unauthorized to set Sn")
	errSnSizeInvalid            = errors.New("sn size invalid")
	errNonceSizeInvalid         = errors.New("size of nonce to be add to table is invalid")
)

// Ecu struct
type Ecu struct {
	ID       string
	VIN      string            // Vehicle identification number
	sn       []byte            // Network key
	nonce    []byte            // nonce - unique identification
	encKey   []byte            // EncKey -Sv-ENC - Ecu key for encryption
	macKey   []byte            // Sv-MAC - Ecu key for generation MAC (also known as tag)
	Kind     int               // Ecu Kind: leader or member
	CertLoc  string            // Location where certificates are present
	nonceAll map[string][]byte // nonceAll stores all the nonces reveived from all ecus in the n/w
	Certs    map[string][]byte // Cert bytes
	mux      sync.Mutex
	nonceMux sync.Mutex
}

// NewEcu returns a new Ecu with sn = 0
func NewEcu(kind int) (*Ecu, error) {
	uuid := fmt.Sprintf("%s", uuid.Must(uuid.NewV4()))
	e := &Ecu{
		ID:       uuid,
		encKey:   make([]byte, encKey),
		macKey:   make([]byte, mackey),
		sn:       make([]byte, snsize),
		nonce:    make([]byte, noncesize),
		nonceAll: make(map[string][]byte, 0),
		CertLoc:  config.DefaultCertificateLocation,
		Certs:    make(map[string][]byte, 0),
	}
	if _, err := rand.Read(e.encKey); err != nil {
		return nil, err
	}
	if _, err := rand.Read(e.macKey); err != nil {
		return nil, err
	}

	switch kind {
	case leader:
		e.Kind = leader
	case member:
		e.Kind = member
	default:
		return nil, errInvalidEcuKind
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

	e.nonce = buf.Bytes()
	return nil
}

// ClearNonceTable clears the nonce table
func (e *Ecu) ClearNonceTable() {
	e.nonceMux.Lock()
	defer e.nonceMux.Unlock()

	e.nonceAll = make(map[string][]byte)
}

// AddToNonceTable adds entry in the nonce map
func (e *Ecu) AddToNonceTable(id string, nonce []byte) error {
	if len(nonce) > noncesize {
		return errNonceSizeInvalid
	}

	e.nonceMux.Lock()
	defer e.nonceMux.Unlock()

	e.nonceAll[id] = nonce
	return nil
}

// GenerateSn generates and sets the Ecu Sn
func (e *Ecu) GenerateSn() error {
	if e.Kind != leader {
		return errUnauthorizedSnGeneration
	}

	if !checkZero(e.sn) {
		return errResetSn
	}

	sn, err := GenerateRandom(snsize)
	if err != nil {
		return err
	}

	copy(e.sn, []byte(sn))
	return nil
}

// GetNonce returns the ecu nonce
func (e *Ecu) GetNonce() []byte {
	return e.nonce
}

// GetSn returns the Ecu Sn string
func (e *Ecu) GetSn() string {
	if checkZero(e.sn) {
		return "0"
	}
	return string(e.sn)
}

// SetSn sets the sn
func (e *Ecu) SetSn(sn []byte) error {
	if e.Kind != member {
		return errUnAuthorizedSnSet
	}

	if len(sn) > snsize {
		return errSnSizeInvalid
	}

	e.mux.Lock()
	defer e.mux.Unlock()

	e.sn = sn
	return nil
}

func (e *Ecu) loadCerts() error {
	e.mux.Lock()
	defer e.mux.Unlock()

	for _, c := range config.DefaultCertificates {
		b, err := ioutil.ReadFile(filepath.Join(e.CertLoc, c))
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

/*
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
*/
