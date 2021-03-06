package controller

import (
	"context"
	"errors"
	"io"
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
	Wait(done chan interface{}) (chan string, chan error)
	CloseClient()
	StartReceiverRoutines(w chan string, errCh chan error)
}

type controller struct {
	service        service.ECU
	supervisorConn *grpc.ClientConn
	client         supervisor.SuperviseClient
	ctx            context.Context
	cancel         context.CancelFunc
	ctype          int
	initCh         chan bool
}

// New returns a new controller
func New(kind int, sim bool) (Controller, error) {
	var err error
	c := &controller{
		ctype:  kind,
		initCh: make(chan bool),
	}

	c.ctx, c.cancel = context.WithTimeout(context.Background(), timeout*time.Second)
	c.service, err = service.NewEcu(kind, sim, c.initCh)
	if err != nil {
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

	c.supervisorConn, ok = <-connected
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
	c.client = supervisor.NewSuperviseClient(c.supervisorConn)
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
	resp, err := c.client.Register(c.ctx, req)
	if err != nil {
		return nil, err
	}

	// Set Ecu ID
	c.service.SetID(resp.Id)

	select {
	case <-c.initCh:
		// Start the listeners associated with the ECU
		service.StartListeners(c.service)
	}

	return resp, nil
}

func (c *controller) Wait(done chan interface{}) (chan string, chan error) {
	ctx := context.Background()
	idCh := make(chan string)
	errch := make(chan error)

	switch c.ctype {
	case leader:
		waitReq := &supervisor.MemberStatusRequest{
			Id: c.service.GetID(),
		}
		respStream, err := c.client.WatchMember(ctx, waitReq)
		if err != nil {
			log.Fatalf(err.Error())
		}

		go func(done chan interface{}) {
			for {
				resp, err := respStream.Recv()
				if err == io.EOF {
					break
				}
				if err != nil {
					errch <- err
					break
				}
				select {
				case idCh <- resp.DependentID:
				case <-done:
					break
				}
			}
		}(done)

	case member:
		waitReq := &supervisor.LeaderStatusRequest{
			Id: c.service.GetID(),
		}
		go func(done chan interface{}) {
			resp, err := c.client.WatchLeader(ctx, waitReq)
			if err != nil {
				errch <- err
				return
			}
			select {
			case idCh <- resp.DependentID:
			case <-done:
				return
			}
		}(done)
	}
	return idCh, errch
}

func (c *controller) StartReceiverRoutines(w chan string, errCh chan error) {
	c.service.CreateUnicastHandlers(w, errCh, c.ctype)
}

func (c *controller) CloseClient() {
	c.cancel()
	c.supervisorConn.Close()
}

func (c *controller) Initiate() error {
	switch t := c.service.(type) {
	case *service.LeaderEcu:
		go t.AnnounceRekey()
		return t.AnnounceSn()

	case *service.MemberEcu:
		return t.AnnounceRekey()

	}
	return nil
}
