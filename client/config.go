package client

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
	ExchangeName   string
	QueueName      string
	ExchangeNoWait bool
	QueueNoWait    bool
	Immediate      bool
}

// Marshal data to Config struct
func (bp BroadCastPublish) Marshal() Config {
	c := NewBareConfig(broadcastExType)
	c.Exchange.Type = broadcastExType

	v := reflect.ValueOf(bp)
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

// BroadcastSubscribe struct
type BroadcastSubscribe struct {
	ExchangeName   string
	QueueName      string
	ConsumerName   string
	RoutingKey     string
	ExchangeNoWait bool
	QueueNoWait    bool
	BindNoWait     bool
	ConsumerNoAck  bool
	ConsumerNoWait bool
}

// Marshal data to Config struct
func (bs BroadcastSubscribe) Marshal() Config {
	c := NewBareConfig(broadcastExType)
	c.Exchange.Type = broadcastExType

	v := reflect.ValueOf(bs)
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		switch t.Field(i).Name {
		case "ExchangeName":
			c.Exchange.Name = v.Field(i).Interface().(string)
			c.QueueBind.Exchange = v.Field(i).Interface().(string)
		case "ExchangeNoWait":
			c.Exchange.NoWait = v.Field(i).Interface().(bool)
		case "QueueName":
			c.QueueDeclare.Queue = v.Field(i).Interface().(string)
			c.QueueBind.Queue = v.Field(i).Interface().(string)
			c.Consumer.Queue = v.Field(i).Interface().(string)
		case "QueueNoWait":
			c.QueueDeclare.NoWait = v.Field(i).Interface().(bool)
		case "ConsumerName":
			c.Consumer.Consumer = v.Field(i).Interface().(string)
		case "RoutingKey":
			c.QueueBind.RoutingKey = v.Field(i).Interface().(string)
		case "BindNoWait":
			c.QueueBind.NoWait = v.Field(i).Interface().(bool)
		case "ConsumerNoAck":
			c.Consumer.NoAck = v.Field(i).Interface().(bool)
		case "ConsumerNoWait":
			c.Consumer.NoWait = v.Field(i).Interface().(bool)
		default:
		}
	}
	return c
}

// PeerSend struct
type PeerSend struct {
	QueueName   string
	QueueNoWait bool
	Immediate   bool
}

// Marshal data to Config struct
func (s PeerSend) Marshal() Config {
	c := NewBareConfig(p2pExType)

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

// PeerReceive struct
type PeerReceive struct {
	QueueName      string
	ConsumerName   string
	ConsumerNoAck  bool
	ConsumerNoWait bool
}

// Marshal data to Config struct
func (r PeerReceive) Marshal() Config {
	c := NewBareConfig(p2pExType)

	v := reflect.ValueOf(r)
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		switch t.Field(i).Name {
		case "QueueName":
			c.QueueDeclare.Queue = v.Field(i).Interface().(string)
			c.Consumer.Queue = v.Field(i).Interface().(string)
		case "QueueNoWait":
			c.QueueDeclare.NoWait = v.Field(i).Interface().(bool)
		case "ConsumerName":
			c.Consumer.Consumer = v.Field(i).Interface().(string)
		case "ConsumerNoAck":
			c.Consumer.NoAck = v.Field(i).Interface().(bool)
		case "ConsumerNoWait":
			c.Consumer.NoWait = v.Field(i).Interface().(bool)
		default:
		}
	}
	return c
}

// exchangeConfig struct
type exchangeConfig struct {
	Name   string
	Type   string
	NoWait bool
}

// publishConfig struct
type publishConfig struct {
	Immediate bool
}

// consumerConfig struct
type consumerConfig struct {
	Queue    string
	Consumer string
	NoWait   bool
	NoAck    bool
}

// queueDeclareConfig struct
type queueDeclareConfig struct {
	Queue  string
	NoWait bool
}

// queueBindConfig struct
type queueBindConfig struct {
	Queue      string
	Exchange   string
	RoutingKey string
	NoWait     bool
}

// Config struct
type Config struct {
	Type         string // Defines the exchange type
	Exchange     exchangeConfig
	Publish      publishConfig
	Consumer     consumerConfig
	QueueDeclare queueDeclareConfig
	QueueBind    queueBindConfig
	Marshaller   Marshaller
}

// NewBareConfig returns a new confic struct
func NewBareConfig(t string) Config {
	return Config{
		Type:         t,
		Exchange:     exchangeConfig{},
		Publish:      publishConfig{},
		Consumer:     consumerConfig{},
		QueueDeclare: queueDeclareConfig{},
		QueueBind:    queueBindConfig{},
		Marshaller:   NewMarshaller(),
	}
}

// Validate the config
func (c Config) Validate() error {
	if c.Marshaller == nil {
		return errors.New("missing message marshaller")
	}
	return nil
}
