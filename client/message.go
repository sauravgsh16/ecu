package client

import (
	"context"
	"errors"
)

var (
	errKeyNotFound = errors.New("Not Found")
	errInvalidType = errors.New("Type Invalid")
)

// Payload represents actual bytes
type Payload []byte

// Metadata represents the message metadata
type Metadata map[string]interface{}

// Set metadata
func (m Metadata) Set(key string, value interface{}) {
	m[key] = value
}

// Get metadata
func (m Metadata) Get(key string) interface{} {
	val, ok := m[key]
	if !ok {
		return ""
	}
	return val
}

// Verify key data valid
func (m Metadata) Verify(key string) (string, error) {
	v, ok := m[key]
	if !ok {
		return "", errKeyNotFound
	}

	val, ok := v.(string)
	if !ok {
		return "", errInvalidType
	}
	return val, nil
}

// Message is a wrapper on the actual payload
type Message struct {
	UUID     string
	Payload  Payload
	ctx      context.Context
	Metadata Metadata
}

// NewMessage returns a new message
func NewMessage(p Payload) *Message {
	return &Message{
		Payload:  p,
		Metadata: make(map[string]interface{}),
	}
}

// Context returns the message context if not nil
// Else sets a new context and returns it
func (m *Message) Context() context.Context {
	if m.ctx != nil {
		return m.ctx
	}
	m.ctx = context.Background()
	return m.ctx
}
