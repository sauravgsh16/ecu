package service

import (
	"fmt"
	"log"
	"strconv"
	"sync"

	"github.com/sauravgsh16/ecu/util"

	"github.com/sauravgsh16/can-interface"
	"github.com/sauravgsh16/ecu/config"
)

type hwService struct {
	ID       string
	SrcID    string
	can      *can.Can
	joinCh   chan bool
	idCh     chan bool
	mux      sync.Mutex
	Incoming chan can.DataHolder
}

func (hw *hwService) getID() string {
	return hw.SrcID
}

func (hw *hwService) setID(id string) {
	hw.mux.Lock()
	defer hw.mux.Unlock()

	hw.SrcID = id

	select {
	case hw.idCh <- true:
	}
}

// LeaderEcuHW is a leader interface with the h/w
type LeaderEcuHW struct {
	hwEcuService
}

func newLeaderHW(c *ecuConfig, initCh chan bool) (*LeaderEcuHW, error) {
	l := new(LeaderEcuHW)
	l.initializeFields()
	l.s = &hwService{
		idCh: l.idCh,
	}
	l.distinct = make(chan can.DataHolder)

	if err := initEcu(l.s, c); err != nil {
		return nil, err
	}

	done := make(chan error)

	go func() {
		select {
		case err := <-done:
			if err != nil {
				panic(err)
			}
			initCh <- true
		}
		close(initCh)
	}()
	l.init(c, done)
	l.s.(*hwService).can.Init()

	return l, nil
}

// StartListeners starts the listeners for a leader
func (l *LeaderEcuHW) StartListeners() {
	go func() {
		for {
			for i := range l.incoming {
				switch i.name {
				case config.Sn:
					fmt.Println("Received Sn. Irrelevant Context.")

				case config.Vin:
					fmt.Println("Received VIN. Irrelevant Context.")

				case config.Nonce:
					/*
						msgs := prepareCanMsg()
						for _, msg := range msgs {

						}
					*/

				case config.Rekey:
					if i.msg.Metadata.Get(appKey) == l.s.getID() {
						continue
					}

				default:
					l.unicastCh <- i
				}
			}
		}
	}()
	l.handleUnicast()

}

func (l *LeaderEcuHW) handleUnicast() {
	go func() {
		for {
			select {
			case <-l.done:
				return
			case i := <-l.unicastCh:
				name := l.unicastRe.FindStringSubmatch(i.name)[1]
				switch name {

				case config.Join:
					// TODO: need to write Join request

				default:
					// TODO: HANDLE NORMAL MESSAGE"
					// TODO: Send to normal message write

				}
			}
		}
	}()

	l.handleDistinct()
	l.handleCanIncoming()
}

func (l *LeaderEcuHW) handleDistinct() {
	go func() {
		for {
			select {
			case d := <-l.distinct:
				switch d.(*can.TP).Pgn {
				case "0100":
					l.sendSnHW(d)
				default:
					log.Fatalf("incorrect message received")
					break
				}
			}
		}
	}()
}

func (hw *hwEcuService) handleCanIncoming() {
	go func() {
		for msg := range hw.s.(*hwService).Incoming {

			switch t := msg.(type) {

			case *can.TP:
				switch t.Pgn {
				// announce Sn
				case "ff02":
					go hw.announceSnHW(t)
				// announce VIN
				case "ff03":
					go hw.announceVinHW(t)
				// send nonce
				case "ff01":
					go hw.announceNonceHW(t)

				default:
					// Join or SendSn
					hw.distinct <- t
				}

			case *can.Message:
				switch t.PGN {
				// start rekey
				case "ff02":
					go hw.announceRekeyHW(t)

				default:
					fmt.Printf("%#v\n", t)
				}
			}
		}
	}()
}

func (l *LeaderEcuHW) sendSnHW(m can.DataHolder) {
	h, ok := l.senders[util.JoinString(config.SendSn, string(encode(m.GetDst())))]
	if !ok {
		log.Fatalf("send Sn handler not found")
	}

	if err := l.send(h, m.GetData(), "sendSn"); err != nil {
		log.Fatalf(err.Error())
	}
}

// MemberEcuHW is a member interface with h/w
type MemberEcuHW struct {
	hwEcuService
}

func newMemberHW(c *ecuConfig, initCh chan bool) (*MemberEcuHW, error) {
	m := new(MemberEcuHW)
	m.initializeFields()
	m.s = &hwService{
		joinCh: make(chan bool),
		idCh:   m.idCh,
	}
	m.distinct = make(chan can.DataHolder)

	if err := initEcu(m.s, c); err != nil {
		return nil, err
	}

	done := make(chan error)

	go func() {
		select {
		case err := <-done:
			if err != nil {
				panic(err)
			}
			initCh <- true
		}
		close(initCh)
	}()
	m.init(c, done)

	return m, nil
}

// StartListeners starts the listeners for a member
func (m *MemberEcuHW) StartListeners() {
	go func() {
		for {
			for i := range m.incoming {
				switch i.name {

				case config.Sn:

				case config.Vin:

				case config.Rekey:
					if i.msg.Metadata.Get(appKey) == m.s.getID() {
						continue
					}

				case config.Nonce:
					if i.msg.Metadata.Get(appKey) == m.s.getID() {
						continue
					}

				default:
					m.unicastCh <- i
				}
			}
		}
	}()

	m.handleUnicast()
}

func (m *MemberEcuHW) handleUnicast() {
	go func() {
		for {
			select {
			case <-m.done:
				return
			case i := <-m.unicastCh:
				name := m.unicastRe.FindStringSubmatch(i.name)[1]
				switch name {

				case config.SendSn:

				default:
					// TODO: HANDLE NORMAL MESSAGE"
					// TODO: Log it. Now just printing to stdout
					fmt.Println(i.msg)
				}
			}
		}
	}()

	m.handleDistinct()
	m.handleCanIncoming()
}

func (m *MemberEcuHW) handleDistinct() {
	go func() {
		for {
			select {
			case d := <-m.distinct:
				switch d.(*can.TP).Pgn {
				case "B200":
					// TODO: SEND JOIN
				default:
					log.Fatalf("incorrect message received")
					break
				}
			}
		}
	}()
}

type hwEcuService struct {
	ecuService
	distinct chan can.DataHolder
}

func (hw *hwEcuService) announceSnHW(tp can.DataHolder) {
	h, ok := hw.broadcasters[config.Sn]
	if !ok {
		log.Fatalf("announce Sn handler not found")
	}

	if err := hw.send(h, tp.GetData(), "announceSn"); err != nil {
		log.Fatalf(err.Error())
	}
}

func (hw *hwEcuService) announceVinHW(tp can.DataHolder) {
	h, ok := hw.broadcasters[config.Vin]
	if !ok {
		log.Fatalf("announce Sn handler not found")
	}

	if err := hw.send(h, tp.GetData(), "announceVin"); err != nil {
		log.Fatalf(err.Error())
	}
}

func (hw *hwEcuService) announceNonceHW(tp can.DataHolder) {
	h, ok := hw.broadcasters[config.Nonce]
	if !ok {
		log.Fatalf("announce nonce handler not found")
	}

	if err := hw.send(h, tp.GetData(), "nonce"); err != nil {
		log.Fatalf(err.Error())
	}
}

func (hw *hwEcuService) announceRekeyHW(m can.DataHolder) {
	h, ok := hw.broadcasters[config.Rekey]
	if !ok {
		log.Fatalf("announce nonce handler not found")
	}

	if err := hw.send(h, m.GetData(), "rekey"); err != nil {
		log.Fatalf(err.Error())
	}
}

func prepareCanMsg(src, dst byte, data []byte, pgn []byte) ([]*can.Message, error) {
	size, err := getSize(int64(len(data)))
	if err != nil {
		return nil, err
	}
	frames := getFrames(len(data))

	var msgs []*can.Message

	// EC message
	msgs = append(msgs, &can.Message{
		ArbitrationID: []uint8{0x00, 0xec, src, dst},
		Data:          []uint8{0x00, size, 0x00, frames, 0x00, pgn[1], pgn[0], 0x00},
	})

	// EB message
	msgs = append(msgs, &can.Message{
		ArbitrationID: []uint8{0x00, 0xeb, src, dst},
		Data:          data,
	})

	return msgs, nil
}

// TODO: Move to utils
const (
	hexchars = "0123456789abcdef"
)

func encode(src byte) []byte {
	dst := make([]byte, 2)
	dst[0] = hexchars[src>>4]
	dst[1] = hexchars[src&0x0F]
	return dst
}

func fromHexChar(ch byte) (byte, bool) {
	switch {
	case '0' <= ch && ch <= '9':
		return ch - '0', true
	case 'a' <= ch && ch <= 'f':
		return ch - 'a' + 10, true
	case 'A' <= ch && ch <= 'F':
		return ch - 'A' + 10, true
	default:
		return 0, false
	}
}

type invalidByte byte

func (i invalidByte) Error() string {
	return fmt.Sprintf("error: invalid byte: %#U", rune(i))
}

func getSize(l int64) (byte, error) {
	var size byte

	sb := []byte(strconv.FormatInt(l, 16))
	a, ok := fromHexChar(sb[0])
	if !ok {
		return 0, invalidByte(sb[0])
	}
	b, ok := fromHexChar(sb[1])
	if !ok {
		return 0, invalidByte(sb[1])
	}

	size = (a << 4) | b
	return size, nil
}

func getFrames(l int) uint8 {
	if l <= 8 {
		return uint8(l)
	}

	var chunk int = 7
	var frames uint8
	for l > 0 {
		if l%chunk != 0 && l < chunk {
			chunk = l % chunk
		}
		l -= chunk
		frames++
	}
	return frames
}
