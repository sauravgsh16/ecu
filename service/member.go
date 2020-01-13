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

const (
	sep = "."
)

func newMember(c *ecuConfig) (*MemberEcu, error) {
	m := new(MemberEcu)
	m.initializeFields()
	m.init(c)
	return m, nil
}

// CreateUnicastHandlers creates handler which memebers require
func (m *MemberEcu) CreateUnicastHandlers(idCh chan string, errCh chan error) {
	select {
	case <-idCh:
		go m.createHandlers()
	case err := <-errCh:
		log.Fatalf(err.Error())
	}
}

func (m *MemberEcu) createHandlers() {
	if err := m.createReceiver(m.domain.ID, handler.NewSendSnReceiver); err != nil {
		log.Fatalf(err.Error())
	}
	if err := m.createSender(m.domain.ID, config.Join, handler.NewJoinSender); err != nil {
		log.Fatalf(err.Error())
	}
}

// StartListeners starts the listeners for a member
func (m *MemberEcu) StartListeners() {
	go func() {
		for {
			for i := range m.incoming {
				switch i.name {

				case config.Nonce:
					go m.handleReceiveNonce(i.msg)

				case config.Sn:
					m.domain.ClearNonceTable()
					go m.handleAnnounceSn(i.msg)

				case config.Vin:
					go m.handleAnnounceVin(i.msg)

				case config.Rekey:

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
				name := m.unicastPattern.FindStringSubmatch(i.name)[1]
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

// SendJoin sends join request to the leader
func (m *MemberEcu) SendJoin() {
	m.mux.RLock()
	defer m.mux.RUnlock()

	sender, ok := m.senders[util.JoinString(config.Join, m.domain.ID)]
	if !ok {
		panic("join Handler not found")
	}

	payload, err := m.aggregateCertNone()
	if err != nil {
		// TODO: Better Error Handling
		// Add logger
		fmt.Printf("Error while creating payload: %s", err.Error())
	}

	log.Printf("Sent Join from - AppID: %s\n", m.domain.ID)

	if err := sender.Send(m.generateMessage(payload)); err != nil {
		// TODO: Better Error Handling
		// Add logger
		fmt.Printf("Error while sending message: %s", err.Error())
	}
}

// AnnounceRekey announces rekey to the networks
func (m *MemberEcu) AnnounceRekey() error {
	if m.getnetworkformationflag() {
		// Signifies that it has already taken part in n/w formation
		return nil
	}

	m.domain.ClearNonceTable()

	h, ok := m.broadcasters[config.Rekey]
	if !ok {
		return fmt.Errorf("announce rekey handler not found")
	}

	empty := make([]byte, 8)
	rkmsg := m.generateMessage(empty)
	rkmsg.Metadata.Set(contentType, "rekey")

	log.Printf("Sending Rekey From AppID - %s\n", m.domain.ID)

	if err := h.Send(rkmsg); err != nil {
		return err
	}
	return nil
}

func (m *MemberEcu) handleAnnounceVin(msg *client.Message) {

	// TODO : ******************

	log.Printf("Received VIN From AppID - %s\n", msg.Metadata.Get(appKey))

	fmt.Println("Not sure what needs to done here")
	// fmt.Printf("Printing the received message: %+v\n", msg)

	// TODO : ******************

}

func (m *MemberEcu) handleSn(msg *client.Message) {

	log.Printf("Received Sn From AppID: - %s\n", msg.Metadata.Get(appKey))

	if bytes.Equal(msg.Payload, []byte(m.domain.GetSn())) {
		return
	}

	if err := m.domain.SetSn(msg.Payload); err != nil {
		// TODO: Better Error Handling
		// Add logger
		fmt.Printf("error setting network key Sn: %s", err.Error())
	}
}

func (m *MemberEcu) handleAnnounceSn(msg *client.Message) {

	log.Printf("Received AnnounceSn From AppID: - %s\n", msg.Metadata.Get(appKey))

	// set n/w formation flag true, to ignore any rekey message
	m.setnetworkformationflag(true)

	// In case of member - Sn represents hash(Sn)
	if bytes.Equal(msg.Payload, []byte(m.domain.GetSn())) {
		return
	}

	go m.SendJoin()
}
