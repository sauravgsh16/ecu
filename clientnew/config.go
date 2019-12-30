package clientnew

import (
	"errors"
	"reflect"
)

const (
	broadcastExType = "fanout"
	p2pExType       = "direct"
)

// BroadCastPublish struct
type BroadCastPublish struct {
	URI            string
	ExchangeName   string
	QueueName      string
	ExchangeNoWait bool
	QueueNoWait    bool
	Immediate      bool
}

func (b BroadCastPublish) Marshal() Config {
	c := NewBareConfig(b.URI)
	c.Exchange.Type = broadcastExType

	v := reflect.ValueOf(b)
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		switch t.Field(i).Name {
		case "ExchangeName":
			c.Exchange.Name = v.Field(i).Interface().(string)
		case "ExchangeNoWait":
			c.Exchange.NoWait = v.Field(i).Interface().(bool)
		case "QueueName":
			c.QueueDeclare.Queue = v.Field(i).Interface().(string)
		case "QueueNoWait":
			c.QueueDeclare.NoWait = v.Field(i).Interface().(bool)
		case "Immediate":
			c.Publish.Immediate = v.Field(i).Interface().(bool)
		default:
		}
	}
	return c
}

// SendPublish struct
type SendPublish struct {
	URI         string
	QueueName   string
	QueueNoWait bool
	Immediate   bool
}

func (s SendPublish) Marshal() Config {
	c := NewBareConfig(s.URI)

	v := reflect.ValueOf(s)
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		switch t.Field(i).Name {
		case "QueueName":
			c.QueueDeclare.Queue = v.Field(i).Interface().(string)
		case "QueueNoWait":
			c.QueueDeclare.NoWait = v.Field(i).Interface().(bool)
		case "Immediate":
			c.Publish.Immediate = v.Field(i).Interface().(bool)
		default:
		}
	}
	return c
}

// ExchangeConfig struct
type ExchangeConfig struct {
	Name   string
	Type   string
	NoWait bool
}

// PublishConfig struct
type PublishConfig struct {
	Immediate bool
}

// ConsumerConfig struct
type ConsumerConfig struct {
	Queue    string
	Consumer string
	NoWait   bool
	NoAck    bool
}

// QueueDeclareConfig struct
type QueueDeclareConfig struct {
	Queue  string
	NoWait bool
}

// QueueBindConfig struct
type QueueBindConfig struct {
	Queue      string
	Exchange   string
	RoutingKey string
	NoWait     bool
}

// Config struct
type Config struct {
	URI          string
	Exchange     ExchangeConfig
	Publish      PublishConfig
	Consumer     ConsumerConfig
	QueueDeclare QueueDeclareConfig
	QueueBind    QueueBindConfig
	Marshaller   Marshaller
}

// NewBareConfig returns a new confic struct
func NewBareConfig(uri string) Config {
	return Config{
		URI:          uri,
		Exchange:     ExchangeConfig{},
		Publish:      PublishConfig{},
		Consumer:     ConsumerConfig{},
		QueueDeclare: QueueDeclareConfig{},
		QueueBind:    QueueBindConfig{},
		Marshaller:   NewMarshaller(),
	}
}

// NewBroadCastPublisherConfig config
func NewBroadCastPublisherConfig(pub BroadCastPublish) Config {
	return Config{}
}

// NewP2PPublisherConfig struct
func NewP2PPublisherConfig() Config {
	return Config{}
}

// Validate the config
func (c Config) Validate() error {
	if c.URI == "" {
		return errors.New("uri cannot be empty")
	}

	if c.Marshaller == nil {
		return errors.New("missing message marshaller")
	}
	return nil
}
