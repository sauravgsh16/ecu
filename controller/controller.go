package controller

import (
	"context"
	"errors"
	"log"
	"time"

	"google.golang.org/grpc"

	"github.com/sauravgsh16/ecu/config"
	"github.com/sauravgsh16/ecu/service"
	supervisor "github.com/sauravgsh16/supervisor/server"
)

const (
	leader int = iota
	member

	// supervisor grpc timeout
	timeout = 600
)

// Controller interface
type Controller interface {
	Initiate() error
	Register() (*supervisor.RegisterNodeResponse, error)
	Wait(id int64) (*supervisor.NodeStatusResponse, error)
	CloseClient()
}

type controller struct {
	service        service.ECU
	superVisorConn *grpc.ClientConn
	client         supervisor.SuperviseClient
	ctx            context.Context
	cancel         context.CancelFunc
	ctype          int
}

// New returns a new controller
func New(kind int) (Controller, error) {
	var err error
	c := &controller{
		ctype: kind,
	}

	c.ctx, c.cancel = context.WithTimeout(context.Background(), timeout*time.Second)
	c.service, err = service.NewEcu(kind)
	if err == nil {
		return nil, err
	}
	return c, c.connect()
}

func (c *controller) connect() error {
	connected := make(chan *grpc.ClientConn)
	ticker := time.NewTicker(1 * time.Second)

	go func() {
		timeout := time.After(30 * time.Second)
		for {
			select {
			case <-timeout:
				close(connected)
				return
			case <-ticker.C:
				conn, err := grpc.Dial(config.SupervisorHost, grpc.WithInsecure())
				if err != nil {
					log.Printf(err.Error())
					continue
				}
				connected <- conn
				return
			}
		}
	}()

	var ok bool

	c.superVisorConn, ok = <-connected
	ticker.Stop()
	if !ok {
		return errors.New("failed to connect to supervisor server")
	}
	log.Println("Successfully connected to supervisor server")
	close(connected)
	c.registerClient()
	return nil
}

func (c *controller) registerClient() {
	c.client = supervisor.NewSuperviseClient(c.superVisorConn)
}

func (c *controller) Register() (*supervisor.RegisterNodeResponse, error) {
	req := new(supervisor.RegisterNodeRequest)

	switch c.ctype {
	case leader:
		req = &supervisor.RegisterNodeRequest{
			Node: &supervisor.Node{
				Type: supervisor.Node_Leader,
			},
		}
	case member:
		req = &supervisor.RegisterNodeRequest{
			Node: &supervisor.Node{
				Type: supervisor.Node_Member,
			},
		}
	default:
		return nil, errors.New("unknown service type")
	}
	return c.client.Register(c.ctx, req)
}

func (c *controller) Wait(id int64) (*supervisor.NodeStatusResponse, error) {
	req := &supervisor.NodeStatusRequest{
		Id: id,
	}
	return c.client.Watch(c.ctx, req)
}

func (c *controller) CloseClient() {
	c.cancel()
	c.superVisorConn.Close()
}

func (c *controller) Initiate() error {
	switch t := c.service.(type) {
	case *service.LeaderEcu:
		return t.AnnounceSn()
	case *service.MemberEcu:
		return t.AnnounceRekey()
	}
	return nil
}
