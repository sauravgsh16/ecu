package service

import (
	"bytes"
	"fmt"
	"log"
	"time"

	"github.com/sauravgsh16/ecu/client"
	"github.com/sauravgsh16/ecu/config"
	"github.com/sauravgsh16/ecu/handler"
	"github.com/sauravgsh16/ecu/util"
)

func newMember(c *ecuConfig) (*MemberEcu, error) {
	m := new(MemberEcu)
	m.initCore()
	m.init(c)
	return m, nil
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
				switch i.name {

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
func (m *MemberEcu) SendJoin(id string) {
	m.mux.RLock()
	defer m.mux.RUnlock()

	sender, ok := m.senders[util.JoinString(config.Join, id)]
	if !ok {
		panic("join Handler not found")
	}

	payload, err := m.aggregateCertNone()
	if err != nil {
		// TODO: Better Error Handling
		// Add logger
		fmt.Printf("Error while creating payload: %s", err.Error())
	}

	log.Printf("(%s) Sent Join : Type - %d\n", time.Now().Format("2006-01-02T15:04:05.999999-07:00"), m.domain.Kind)

	if err := sender.Send(m.generateMessage(payload)); err != nil {
		// TODO: Better Error Handling
		// Add logger
		fmt.Printf("Error while sending message: %s", err.Error())
	}
}

// AnnounceRekey announces rekey to the networks
func (m *MemberEcu) AnnounceRekey() error {
	m.domain.ClearNonceTable()

	empty := make([]byte, 8)

	rkmsg := m.generateMessage(empty)
	rkmsg.Metadata[contentType] = "rekey"

	h, ok := m.broadcasters[config.Rekey]
	if !ok {
		return fmt.Errorf("announce rekey handler not found")
	}

	if err := h.Send(rkmsg); err != nil {
		return err
	}
	return nil
}

func (m *MemberEcu) handleAnnounceVin(msg *client.Message) {

	// TODO : ******************

	log.Printf("(%s) Received VIN : Type - %d\n", time.Now().Format("2006-01-02T15:04:05.999999-07:00"), m.domain.Kind)

	fmt.Println("Not sure what needs to done here")
	// fmt.Printf("Printing the received message: %+v\n", msg)

	// TODO : ******************

}

func (m *MemberEcu) handleSn(msg *client.Message) {

	log.Printf("(%s) Received Sn : Type - %d\n", time.Now().Format("2006-01-02T15:04:05.999999-07:00"), m.domain.Kind)

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

	log.Printf("(%s) Received AnnounceSn : Type - %d\n", time.Now().Format("2006-01-02T15:04:05.999999-07:00"), m.domain.Kind)

	// set n/w formation flag true, to ignore any rekey message
	m.setnetworkformationflag(true)

	// In case of member - Sn represents hash(Sn)
	if bytes.Equal(msg.Payload, []byte(m.domain.GetSn())) {
		return
	}

	appID, err := msg.Metadata.Verify(appKey)
	if err != nil {
		// TODO: improve error handling
		// panicing for now
		panic(fmt.Sprintf("%s: %s", appKey, err.Error()))
	}

	if err := m.registerJoinSender(appID); err != nil {
		// TODO: improve error handling
		// panicing for now
		panic(fmt.Sprintf("%s: %s", errJoinRegister, err.Error()))
	}
	go m.SendJoin(appID)
}

func (m *MemberEcu) registerJoinSender(appID string) error {
	m.mux.Lock()
	defer m.mux.Unlock()

	_, ok := m.senders[appID]
	if ok {
		return nil
	}

	h, err := handler.NewJoinSender(appID)
	if err != nil {
		return err
	}

	m.senders[h.GetName()] = h
	return nil
}
