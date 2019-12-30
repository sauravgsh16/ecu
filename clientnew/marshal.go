package clientnew

import (
	"github.com/sauravgsh16/message-server/qclient"
)

type Marshaller interface {
	Marshal(*Message) (Payload, error)
	Unmarshal(*qclient.Delivery) (*Message, error)
}
