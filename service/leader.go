package service

import (
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"

	"github.com/sauravgsh16/ecu/client"
	"github.com/sauravgsh16/ecu/config"
	"github.com/sauravgsh16/ecu/handler"
	"github.com/sauravgsh16/ecu/util"
)

func newLeader(c *ecuConfig) (*LeaderEcu, error) {
	l := new(LeaderEcu)
	l.initializeFields()
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

// CreateUnicastHandlers creates handler which leaders require
func (l *LeaderEcu) CreateUnicastHandlers(idCh chan string, errCh chan error) {
	go func() {
	loop:
		for {
			select {
			case id := <-idCh:
				go l.createHandlers(id)
			case err := <-errCh:
				log.Printf(err.Error())
				break loop
			}
		}
	}()
}

func (l *LeaderEcu) createHandlers(id string) {
	l.createJoinReceiver(id)
	l.createSnSender(id)
}

func (l *LeaderEcu) createSnSender(id string) {
	if err := l.createSender(id, config.SendSn, handler.NewSendSnSender); err != nil {
		log.Fatalf(err.Error())
	}
}

func (l *LeaderEcu) createJoinReceiver(id string) {
	if err := l.createReceiver(id, handler.NewJoinReceiver); err != nil {
		log.Fatalf(err.Error())
	}
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

					log.Printf("Received Rekey From appID: - %s\n", i.msg.Metadata.Get(appKey))

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
				name := l.unicastPattern.FindStringSubmatch(i.name)[1]
				switch name {

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

	log.Printf("Broadcasting Sn from AppID: - %s\n", l.domain.ID)

	hashSn := util.GenerateHash([]byte(l.domain.GetSn()))
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
	msg.Metadata.Set(contentType, "vinCert")

	log.Printf("Broadcasting VIN from AppID: - %s\n", msg.Metadata.Get(appKey))

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
	// register the send sn register
	appID, err := msg.Metadata.Verify(appKey)
	if err != nil {
		// TODO: improve error handling
		// TODO: panicing for now
		panic(fmt.Sprintf("%s: %s", appKey, err.Error()))
	}

	log.Printf("Received Join from AppID: %s\n", msg.Metadata.Get(appKey))

	go l.SendSn(appID)
}

// SendSn sends Sn to memeber ECUs
func (l *LeaderEcu) SendSn(id string) {
	l.mux.RLock()
	defer l.mux.RUnlock()

	sender, ok := l.senders[util.JoinString(config.SendSn, id)]
	if !ok {
		panic("sender not found")
	}
	// TODO: process Sn
	// TODO: ITK logic to be added

	log.Printf("Send Sn to AppID: - %s\n\n\n\n\n", id)

	sn := l.domain.GetSn()
	if err := sender.Send(l.generateMessage([]byte(sn))); err != nil {
		// TODO: Better Error Handling
		// Add logger
		fmt.Printf("Error while sending message: %s", err.Error())
	}
	// Set network formation flag false - so that any rekey message received
	// during n/w formation will be start network formation again.
	l.setnetworkformationflag(false)
}
