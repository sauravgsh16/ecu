package domain

import (
	"bytes"
	"crypto/rand"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/gofrs/uuid"
	"github.com/keep94/appcommon/kdf"

	"github.com/sauravgsh16/ecu/config"
	"github.com/sauravgsh16/ecu/util"
)

const (
	// leader ECU type
	leader = iota
	// Member ECU type
	member
)

const (
	snsize        = 16
	noncesize     = 16
	encKey        = 64
	mackey        = 64
	secocRekeyEnc = 0
	secocRekeyMac = 1
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
	svenc    []byte            // Enc Calculated Key
	svmac    []byte            // Mac Calculated key
	myNonce  []byte            // nonce - unique identification
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

	log.Printf("New ECU registered with ID: %s\n\n", uuid)

	e := &Ecu{
		ID:       uuid,
		encKey:   make([]byte, encKey),
		macKey:   make([]byte, mackey),
		sn:       make([]byte, snsize),
		myNonce:  make([]byte, noncesize),
		nonceAll: make(map[string][]byte, 0),
		Certs:    make(map[string][]byte, 0),
		svenc:    make([]byte, 0),
		svmac:    make([]byte, 0),
	}
	dir, err := os.Getwd()
	if err != nil {
		log.Fatalf(err.Error())
	}
	e.CertLoc = filepath.Join(dir, "../../../", config.DefaultCertificateLocation)

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

	e.myNonce = buf.Bytes()
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

// RemoveFromNonceTable removes key-value pair from map
func (e *Ecu) RemoveFromNonceTable(id string) {
	e.nonceMux.Lock()
	e.nonceMux.Unlock()

	delete(e.nonceAll, id)
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
	e.mux.Lock()
	defer e.mux.Unlock()

	return e.myNonce
}

// ResetNonce resets the ecu's my_nonce
func (e *Ecu) ResetNonce() error {
	e.mux.Lock()
	defer e.mux.Unlock()

	return e.generateNonce()
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

// CalculateSv calculates the sv based on Sn and all the received nonces
func (e *Ecu) CalculateSv() {
	e.nonceMux.Lock()
	e.nonceMux.Unlock()

	all := make([]byte, 0)
	all = append(all, e.myNonce...)

	for _, v := range e.nonceAll {
		all = append(all, v...)
	}

	// TODO - apply KDF algorithm
	nonceAll := []byte(util.GenerateHash(all))
	e.svenc = kdf.KDF(e.sn, nonceAll, secocRekeyEnc)
	e.svmac = kdf.KDF(e.sn, nonceAll, secocRekeyMac)
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
