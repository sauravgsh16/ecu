package clientnew

import (
	"fmt"
	"reflect"

	multierror "github.com/hashicorp/go-multierror"

	"github.com/sauravgsh16/message-server/qclient"
)

// Marshaller interface
type Marshaller interface {
	Marshal(*Message) (qclient.MetaDataWithBody, error)
	Unmarshal(*qclient.Delivery) (*Message, error)
}

type marshall struct{}

// NewMarshaller returns a new marshaller
func NewMarshaller() Marshaller {
	return marshall{}
}

// Marshall message into qclient.MetaDataWithBody
func (m marshall) Marshal(msg *Message) (qclient.MetaDataWithBody, error) {
	md := qclient.MetaDataWithBody{
		Body: msg.Payload,
	}

	var err error

	for k, v := range msg.Metadata {
		switch k {
		case "ContentType":
			md.ContentType = v.(string)
		case "MessageID":
			md.MessageID = v.(string)
		case "UserID":
			md.UserID = v.(string)
		case "ApplicationID":
			md.ApplicationID = v.(string)
		default:
			err = multierror.Append(err, fmt.Errorf("invalid metadata: %s", k))
		}
	}

	return md, err
}

// Unmarshal delivery struct into Message
func (m marshall) Unmarshal(d *qclient.Delivery) (*Message, error) {
	if len(d.Body) == 0 {
		return nil, fmt.Errorf("missing data")
	}

	msg := NewMessage(d.Body)

	v := reflect.ValueOf(*d)
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {

		switch t.Field(i).Name {
		case "Body":
		case "DeliveryTag":
			msg.Metadata[t.Field(i).Name] = v.Field(i).Interface().(uint64)
		default:
			msg.Metadata[t.Field(i).Name] = v.Field(i).Interface().(string)
		}
	}

	return msg, nil
}
