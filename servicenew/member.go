package servicenew

import (
	"bytes"
	"fmt"

	"github.com/sauravgsh16/ecu/client"
	"github.com/sauravgsh16/ecu/handler"
)

func (e *ecuService) SendJoin(id string) {
	e.mux.RLock()
	defer e.mux.RUnlock()

	sender, _ := e.senders[id]

	payload, err := e.aggregateCertNone()
	if err != nil {
		// TODO: Better Error Handling
		// Add logger
		fmt.Printf("Error while creating payload: %s", err.Error())
	}

	if err := sender.Send(e.generateMessage(payload)); err != nil {
		// TODO: Better Error Handling
		// Add logger
		fmt.Printf("Error while sending message: %s", err.Error())
	}
}

func (e *ecuService) AnnounceRekey() {}

// This needs to be added to leader interface too
func (e *ecuService) AnnounceNonce() {}

func (e *ecuService) handleAnnounceVin() {

	// TODO : ******************

	fmt.Println("Not sure what needs to done here")

	// TODO : ******************

}

func (e *ecuService) handleSn(msg *client.Message) {
	if bytes.Equal(msg.Payload, []byte(e.domain.GetSn())) {
		return
	}

	if err := e.domain.SetSn(msg.Payload); err != nil {
		// TODO: Better Error Handling
		// Add logger
		fmt.Printf("error setting network key Sn: %s", err.Error())
	}
}

func (e *ecuService) handleAnnounceSn(msg *client.Message) {
	// In case of member - Sn represents hash(Sn)
	if bytes.Equal(msg.Payload, []byte(e.domain.GetSn())) {
		return
	}

	appID, err := msg.Metadata.Verify(appKey)
	if err != nil {
		// TODO: improve error handling
		// panicing for now
		panic(fmt.Sprintf("%s: %s", appKey, err.Error()))
	}

	if err := e.registerJoinSender(appID); err != nil {
		// TODO: improve error handling
		// panicing for now
		panic(fmt.Sprintf("%s: %s", errJoinRegister, err.Error()))
	}
	go e.SendJoin(appID)
}

func (e *ecuService) registerJoinSender(appID string) error {
	e.mux.Lock()
	defer e.mux.Unlock()

	_, ok := e.senders[appID]
	if ok {
		return nil
	}

	h, err := handler.NewJoinSender(appID)
	if err != nil {
		return err
	}

	e.senders[h.GetName()] = h
	return nil
}
