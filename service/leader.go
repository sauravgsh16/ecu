package service

import (
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"time"

	"github.com/sauravgsh16/ecu/client"
	"github.com/sauravgsh16/ecu/config"
	"github.com/sauravgsh16/ecu/handler"
	"github.com/sauravgsh16/ecu/util"
)

func newLeader(c *ecuConfig) (*LeaderEcu, error) {
	l := new(LeaderEcu)
	l.initCore()
	l.certs = make(map[string][]byte)
	l.init(c)

	if err := l.loadCerts(); err != nil {
		return nil, err
	}

	if err := l.domain.GenerateSn(); err != nil {
		return nil, err
	}
	return l, nil
}

// StartListeners starts the listeners for a leader
func (l *LeaderEcu) StartListeners() {
	go func() {
		for {
			for i := range l.incoming {
				switch i.name {

				case config.Nonce:
					go l.handleReceiveNonce(i.msg)

				case config.Rekey:
					if l.getnetworkformationflag() {
						continue
					}
					l.domain.ClearNonceTable()
					go l.AnnounceSn()

				default:
					l.unicastCh <- i
				}
			}
		}
	}()

	l.handleUnicast()
}

func (l *LeaderEcu) handleUnicast() {
	go func() {
		for {
			select {
			case <-l.done:
				return
			case i := <-l.unicastCh:
				switch i.name {

				case config.Join:
					go l.handleJoin(i.msg)

				default:
					// TODO: HANDLE NORMAL MESSAGE"
					// TODO: Log it. Now just printing to stdout
					fmt.Println(i.msg)
				}
			}
		}
	}()
}

// AnnounceSn accnounces the Sn
func (l *LeaderEcu) AnnounceSn() error {
	// Set network formation flag true - so that any rekey message
	// during n/w formation will be ignored.
	l.setnetworkformationflag(true)

	l.mux.RLock()
	defer l.mux.RUnlock()

	h, ok := l.broadcasters[config.Sn]
	if !ok {
		return fmt.Errorf("announce Sn handler not found")
	}

	hashSn := util.GenerateHash([]byte(l.domain.GetSn()))

	log.Printf("(%s) : Broadcasting Sn : Type - %d\n", time.Now().Format("2006-01-02T15:04:05.999999-07:00"), l.domain.Kind)

	if err := h.Send(l.generateMessage([]byte(hashSn))); err != nil {
		return err
	}
	return l.AnnounceVin()
}

// AnnounceVin announces the vin
func (l *LeaderEcu) AnnounceVin() error {
	l.mux.RLock()
	defer l.mux.RUnlock()

	h, ok := l.broadcasters[config.Vin]
	if !ok {
		return fmt.Errorf("announce vim handler not found")
	}

	if !l.certLoaded || len(l.certs) == 0 {
		return fmt.Errorf("vim certs not loaded")
	}

	payload, err := l.aggregateCertNone()
	if err != nil {
		return err
	}

	// TODO: standardize message format for all message types
	msg := l.generateMessage(payload)
	msg.Metadata[contentType] = "vinCert"

	log.Printf("(%s)Broadcasting VIN : Type - %d\n", time.Now().Format("2006-01-02T15:04:05.999999-07:00"), l.domain.Kind)

	if err := h.Send(msg); err != nil {
		return err
	}
	return nil
}

// TODO: CHECK WHICH CERTS ARE REQUIRED TO BE LOADED
func (l *LeaderEcu) loadCerts() error {
	if l.certLoaded {
		return nil
	}

	l.certMux.Lock()
	defer l.certMux.Unlock()

	l.certs = make(map[string][]byte)

	for _, c := range config.DefaultBroadCastCertificates {
		b, err := ioutil.ReadFile(filepath.Join(l.domain.CertLoc, c))
		if err != nil {
			return fmt.Errorf("failed to load certificate: %s", err.Error())
		}
		l.certs[c] = b
	}
	l.certLoaded = true
	return nil
}

func (l *LeaderEcu) handleJoin(msg *client.Message) {

	log.Printf("(%s) Received Join : Type - %d\n", time.Now().Format("2006-01-02T15:04:05.999999-07:00"), l.domain.Kind)

	// register the send sn register
	appID, err := msg.Metadata.Verify(appKey)
	if err != nil {
		// TODO: improve error handling
		// TODO: panicing for now
		panic(fmt.Sprintf("%s: %s", appKey, err.Error()))
	}

	if err := l.registerSnSender(appID); err != nil {
		// TODO: improve error handling
		// TODO: panicing for now
		panic(fmt.Sprintf("%s: %s", errSendSnRegister, err.Error()))
	}
	go l.SendSn(appID)
}

func (l *LeaderEcu) registerSnSender(appID string) error {
	l.mux.Lock()
	defer l.mux.Unlock()

	_, ok := l.senders[appID]
	if ok {
		return nil
	}

	h, err := handler.NewSendSnSender(appID)
	if err != nil {
		return err
	}

	l.senders[h.GetName()] = h
	return nil
}

// SendSn sends Sn to memeber ECUs
func (l *LeaderEcu) SendSn(id string) {
	l.mux.RLock()
	defer l.mux.RUnlock()

	sender, _ := l.senders[util.JoinString(config.SendSn, id)]
	sn := l.domain.GetSn()

	// TODO: process Sn
	// TODO: ITK logic to be added

	log.Printf("(%s) Send Sn : Type - %d\n", time.Now().Format("2006-01-02T15:04:05.999999-07:00"), l.domain.Kind)

	if err := sender.Send(l.generateMessage([]byte(sn))); err != nil {
		// TODO: Better Error Handling
		// Add logger
		fmt.Printf("Error while sending message: %s", err.Error())
	}
	// Set network formation flag false - so that any rekey message received
	// during n/w formation will be start network formation again.
	l.setnetworkformationflag(false)
}
