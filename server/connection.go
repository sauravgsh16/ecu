package server

import (
	"bufio"
	"io"
	"net"
	"sync"

	"github.com/sauravgsh16/ecu/config"

	"github.com/sauravgsh16/ecu/service"
	"github.com/sauravgsh16/message-server/qclient"
	"github.com/sauravgsh16/message-server/qserver/server"
)

// Connection struct
type Connection struct {
	network       net.Conn
	msgServerConn *qclient.Connection
	mux           sync.Mutex
	status        int
	writer        io.Writer
	service       service.EcuService
	channels      map[string]*server.Channel
}

// NewConnection returns a new connection
func NewConnection(n net.Conn) (*Connection, error) {
	conn, err := qclient.Dial(config.MessageServerHost)
	if err != nil {
		return nil, err
	}
	return &Connection{
		msgServerConn: conn,
		writer:        bufio.NewWriter(n),
		channels:      make(map[string]*server.Channel, 0),
	}, nil
}
