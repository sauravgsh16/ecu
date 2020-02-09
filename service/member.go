package service

import (
	"bytes"
	"fmt"
	"log"

	"github.com/sauravgsh16/ecu/client"
	"github.com/sauravgsh16/ecu/config"
	"github.com/sauravgsh16/ecu/handler"
	"github.com/sauravgsh16/ecu/util"
)

// MemberEcu struct
type MemberEcu struct {
	ecuService
	joinCh chan bool
}

func newMember(c *ecuConfig) (*MemberEcu, error) {
	m := new(MemberEcu)
	m.initializeFields()
	m.init(c)

	m.joinCh = make(chan bool)

	return m, nil
}

// CreateUnicastHandlers creates handler which memebers require
func (m *MemberEcu) CreateUnicastHandlers(idCh chan string, errCh chan error) {
	go func() {
		select {
		case <-idCh:
			go m.createHandlers()
		case err := <-errCh:
			log.Fatalf(err.Error())
		}
	}()
}

func (m *MemberEcu) createHandlers() {
	if err := m.createReceiver(m.domain.ID, handler.NewSendSnReceiver); err != nil {
		log.Fatalf(err.Error())
	}
	if err := m.createSender(m.domain.ID, config.Join, handler.NewJoinSender); err != nil {
		log.Fatalf(err.Error())
	}
	select {
	case m.joinCh <- true:
	}
}

// StartListeners starts the listeners for a member
func (m *MemberEcu) StartListeners() {
	go func() {
		for {
			for i := range m.incoming {
				switch i.name {

				case config.Sn:
					go m.handleAnnounceSn(i.msg)

				case config.Vin:
					go m.handleAnnounceVin(i.msg)

				case config.Rekey:
					go m.handleRekey(i.msg)

				case config.Nonce:
					go m.handleReceiveNonce(i.msg)

				default:
					m.unicastCh <- i
				}
			}
		}
	}()

	m.handleUnicast()
}

func (m *MemberEcu) handleUnicast() {
	go func() {
		for {
			select {
			case <-m.done:
				return
			case i := <-m.unicastCh:
				name := m.unicastRe.FindStringSubmatch(i.name)[1]
				switch name {

				case config.SendSn:
					go m.handleSn(i.msg)

				default:
					// TODO: HANDLE NORMAL MESSAGE"
					// TODO: Log it. Now just printing to stdout
					fmt.Println(i.msg)
				}
			}

		}
	}()
}

func (m *MemberEcu) handleAnnounceSn(msg *client.Message) {
	log.Printf("Received AnnounceSn From AppID: - %s\n", msg.Metadata.Get(appKey))

	/*
		if m.getnetworkformationflag() {
			return
		}
	*/

	if bytes.Equal(msg.Payload, []byte(m.domain.GetSn())) {
		log.Println("Received Sn is equal to Sn stored. Returning")
		return
	}

	/*
		// set n/w formation flag true, to ignore any rekey message
		m.setnetworkformationflag(true)
	*/

	go m.SendJoin()
}

func (m *MemberEcu) handleAnnounceVin(msg *client.Message) {

	// TODO : ******************
	log.Printf("Received VIN From AppID - %s\n", msg.Metadata.Get(appKey))
	fmt.Println("Not sure what needs to done here")
	// fmt.Printf("Printing the received message: %+v\n", msg)

	// TODO : ******************

}

// SendJoin sends join request to the leader
func (m *MemberEcu) SendJoin() {
	select {
	case <-m.joinCh:
	}

	sender, ok := m.senders[util.JoinString(config.Join, m.domain.ID)]
	if !ok {
		panic("join Handler not found")
	}

	m.mux.RLock()
	defer m.mux.RUnlock()

	cert, err := m.aggregateCert()
	if err != nil {
		// TODO: Better Error Handling
		// Add logger
		log.Printf("Error while creating payload: %s", err.Error())
	}

	nonce, err := m.generateNonce()
	if err != nil {
		log.Printf("Error while creating payload: %s", err.Error())
	}

	log.Printf("Sent Join from - AppID: %s\n", m.domain.ID)

	if err := m.send(sender, cert, "cert"); err != nil {
		// TODO: Better Error Handling
		// Add logger
		log.Printf("Error while sending message: %s", err.Error())
	}

	if err := m.send(sender, nonce, "nonce"); err != nil {
		log.Printf("Error while sending message: %s", err.Error())
	}

}

func (m *MemberEcu) handleSn(msg *client.Message) {
	log.Printf("Received Sn From AppID: - %s\n", msg.Metadata.Get(appKey))

	if err := m.domain.SetSn(msg.Payload); err != nil {
		// TODO: Better Error Handling
		// Add logger
		log.Printf("error setting network key Sn: %s", err.Error())
	}
	/*

		m.setnetworkformationflag(false)

	*/
}
