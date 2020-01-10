package service

import (
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"time"

	"github.com/gofrs/uuid"

	"github.com/sauravgsh16/ecu/client"
	"github.com/sauravgsh16/ecu/config"
	"github.com/sauravgsh16/ecu/handler"
	"github.com/sauravgsh16/ecu/util"
)

func (e *ecuService) AnnounceSn() error {
	// Set network formation flag true - so that any rekey message
	// during n/w formation will be ignored.
	e.setnetworkformationflag(true)

	e.mux.RLock()
	defer e.mux.RUnlock()

	h, ok := e.broadcasters[config.Sn]
	if !ok {
		return fmt.Errorf("announce Sn handler not found")
	}

	hashSn := util.GenerateHash([]byte(e.domain.GetSn()))

	log.Printf("(%s) : Broadcasting Sn : Type - %d\n", time.Now().Format("2006-01-02T15:04:05.999999-07:00"), e.domain.Kind)

	if err := h.Send(e.generateMessage([]byte(hashSn))); err != nil {
		return err
	}
	return e.AnnounceVin()
}

func (e *ecuService) AnnounceVin() error {
	e.mux.RLock()
	defer e.mux.RUnlock()

	h, ok := e.broadcasters[config.Vin]
	if !ok {
		return fmt.Errorf("announce vim handler not found")
	}

	if !e.certLoaded || len(e.certs) == 0 {
		return fmt.Errorf("vim certs not loaded")
	}

	payload, err := e.aggregateCertNone()
	if err != nil {
		return err
	}

	// TODO: standardize message format for all message types
	msg := e.generateMessage(payload)
	msg.Metadata[contentType] = "vinCert"

	log.Printf("(%s)Broadcasting VIN : Type - %d\n", time.Now().Format("2006-01-02T15:04:05.999999-07:00"), e.domain.Kind)

	if err := h.Send(msg); err != nil {
		return err
	}
	return nil
}

// TODO: CHECK WHICH CERTS ARE REQUIRED TO BE LOADED
func (e *ecuService) loadCerts() error {
	if e.certLoaded {
		return nil
	}

	e.certMux.Lock()
	defer e.certMux.Unlock()

	e.certs = make(map[string][]byte)

	for _, c := range config.DefaultBroadCastCertificates {
		b, err := ioutil.ReadFile(filepath.Join(e.domain.CertLoc, c))
		if err != nil {
			return fmt.Errorf("failed to load certificate: %s", err.Error())
		}
		e.certs[c] = b
	}
	e.certLoaded = true
	return nil
}

func (e *ecuService) handleJoin(msg *client.Message) {

	log.Printf("(%s) Received Join : Type - %d\n", time.Now().Format("2006-01-02T15:04:05.999999-07:00"), e.domain.Kind)

	// register the send sn register
	appID, err := msg.Metadata.Verify(appKey)
	if err != nil {
		// TODO: improve error handling
		// panicing for now
		panic(fmt.Sprintf("%s: %s", appKey, err.Error()))
	}

	if err := e.registerSnSender(appID); err != nil {
		// TODO: improve error handling
		// panicing for now
		panic(fmt.Sprintf("%s: %s", errSendSnRegister, err.Error()))
	}
	go e.SendSn(appID)
}

func (e *ecuService) registerSnSender(appID string) error {
	e.mux.Lock()
	defer e.mux.Unlock()

	_, ok := e.senders[appID]
	if ok {
		return nil
	}

	h, err := handler.NewSendSnSender(appID)
	if err != nil {
		return err
	}

	e.senders[h.GetName()] = h
	return nil
}

func (e *ecuService) SendSn(id string) {
	e.mux.RLock()
	defer e.mux.RUnlock()

	sender, _ := e.senders[util.JoinString(config.SendSn, id)]
	sn := e.domain.GetSn()

	// TODO: process Sn
	// TODO: ITK logic to be added

	log.Printf("(%s) Send Sn : Type - %d\n", time.Now().Format("2006-01-02T15:04:05.999999-07:00"), e.domain.Kind)

	if err := sender.Send(e.generateMessage([]byte(sn))); err != nil {
		// TODO: Better Error Handling
		// Add logger
		fmt.Printf("Error while sending message: %s", err.Error())
	}
	// Set network formation flag false - so that any rekey message received
	// during n/w formation will be start network formation again.
	e.setnetworkformationflag(false)

}

func (e *ecuService) generateMessage(payload []byte) *client.Message {
	uuid := fmt.Sprintf("%s", uuid.Must(uuid.NewV4()))

	return &client.Message{
		UUID:    uuid,
		Payload: client.Payload(payload),
		Metadata: client.Metadata(map[string]interface{}{
			"ApplicationID": e.domain.ID,
		}),
	}
}
