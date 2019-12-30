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

// Marshal data to Config struct
func (bp BroadCastPublish) Marshal() Config {
	c := NewBareConfig(bp.URI, broadcastExType)
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

// SendPublish struct
type SendPublish struct {
	URI         string
	QueueName   string
	QueueNoWait bool
	Immediate   bool
}

// Marshal data to Config struct
func (s SendPublish) Marshal() Config {
	c := NewBareConfig(s.URI, p2pExType)

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

// BroadcastSubscribe struct
type BroadcastSubscribe struct {
	URI            string
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
	c := NewBareConfig(bs.URI, broadcastExType)
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
	Type         string // Defines the exchange type
	Exchange     ExchangeConfig
	Publish      PublishConfig
	Consumer     ConsumerConfig
	QueueDeclare QueueDeclareConfig
	QueueBind    QueueBindConfig
	Marshaller   Marshaller
}

// NewBareConfig returns a new confic struct
func NewBareConfig(uri, t string) Config {
	return Config{
		URI:          uri,
		Type:         t,
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
