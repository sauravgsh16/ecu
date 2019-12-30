package clientnew

import (
	"context"
)

// Payload represents actual bytes
type Payload []byte

// Message is a wrapper on the actual payload
type Message struct {
	UUID    string
	Payload Payload
	ctx     context.Context
}

// NewMessage returns a new message
func NewMessage(uuid string, p Payload) *Message {
	return &Message{
		UUID:    uuid,
		Payload: p,
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
