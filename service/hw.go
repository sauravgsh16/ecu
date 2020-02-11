package service

import (
	"fmt"
	"strconv"

	"github.com/sauravgsh16/ecu/can"
	"github.com/sauravgsh16/ecu/config"
)

type hwService struct {
	ID     string
	can    *can.Can
	joinCh chan bool
}

func (hw *hwService) getID() string {
	return hw.ID
}

// LeaderEcuHW is a leader interface with the h/w
type LeaderEcuHW struct {
	ecuService
}

func newLeaderHW(c *ecuConfig) (*LeaderEcuHW, error) {
	l := new(LeaderEcuHW)
	l.initializeFields()
	l.s = &hwService{}
	initEcu(l.s, c)

	if err := l.init(c); err != nil {
		return nil, err
	}
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
					if i.msg.Metadata.Get(appKey) == l.s.(*hwService).ID {
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

				default:
					// TODO: HANDLE NORMAL MESSAGE"
					// TODO: Send to normal message write

				}
			}
		}
	}()
}

// MemberEcuHW is a member interface with h/w
type MemberEcuHW struct {
	ecuService
}

func newMemberHW(c *ecuConfig) (*MemberEcuHW, error) {
	m := new(MemberEcuHW)
	m.initializeFields()
	m.s = &hwService{
		joinCh: make(chan bool),
	}
	initEcu(m.s, c)

	if err := m.init(c); err != nil {
		return nil, err
	}

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
					if i.msg.Metadata.Get(appKey) == m.s.(*hwService).ID {
						continue
					}

				case config.Nonce:
					if i.msg.Metadata.Get(appKey) == m.s.(*hwService).ID {
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
		Data:          []uint8{0x00, size, 0x00, frames, 0x00, 0x00, pgn[0], pgn[1]},
	})

	// EB message
	msgs = append(msgs, &can.Message{
		ArbitrationID: []uint8{0x00, 0xeb, src, dst},
		Data:          data,
	})

	return msgs, nil
}

// TODO: Move to utils
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
