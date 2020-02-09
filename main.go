package main

import (
	"encoding/hex"
	"fmt"
	"log"
	"time"

	"github.com/sauravgsh16/ecu/can"
)

func main() {
	forever := make(chan interface{})
	done := make(chan interface{})
	c, err := can.New()
	if err != nil {
		log.Fatalf(err.Error())
	}
	c.Init()

	go func() {
		timer := time.NewTimer(5 * time.Second)
		ticker := time.NewTicker(1 * time.Millisecond)
		var count int
	loop:
		for {
			select {
			case <-timer.C:
				break loop
			case <-ticker.C:
				c.Out <- &can.Message{
					ArbitrationID: []uint8{0x1c, 0xeb, 0xf7, 0xe8},
					Priority:      0x00,
					PGN:           hex.EncodeToString([]uint8{0xf7, 0x0}),
					Src:           0xe8, Dst: 0xf7,
					Size: 0x8,
					Data: []uint8{0x5, 0x31, 0x2e, 0x30, 0x31, 0xa2, 0xff, 0xbf},
				}
				count++
			}
		}
		done <- count
	}()

	count := <-done
	fmt.Printf("%d\n", count)
	c.Done <- true

	<-forever
}
