package service

import (
	"fmt"

	"github.com/sauravgsh16/ecu/can"
	"github.com/sauravgsh16/ecu/config"
)

type hwService struct {
	ID  string
	can *can.Can
}

func (hw *hwService) getID() string {
	return hw.ID
}

// LeaderEcuHW is a leader interface with h/w
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

				case config.Rekey:
					if i.msg.Metadata.Get(appKey) == l.s.(*swService).domain.ID {
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
	m.s = &hwService{}
	initEcu(m.s, c)

	if err := m.init(c); err != nil {
		return nil, err
	}
	return m, nil
}
