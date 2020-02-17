package main

import (
	"fmt"
	"log"
	"time"

	"github.com/sauravgsh16/can-interface"
)

func main() {
	forever := make(chan interface{})
	done := make(chan interface{})

	in := make(chan can.DataHolder)

	test := "Client Header \"0.1\" Name:\"Test Simulator Interface\""
	c, err := can.New("tcp://localhost:19000", in, test)
	if err != nil {
		log.Fatalf(err.Error())
	}
	c.Init()

	go func() {
		timer := time.NewTimer(5 * time.Second)
		ticker := time.NewTicker(3 * time.Second)
		var count int
	loop:
		for {
			select {
			case <-timer.C:
				break loop
			case <-ticker.C:
				c.Out <- &can.Message{
					ArbitrationID: []uint8{0x1c, 0xeb, 0xf7, 0xe8},
					Data:          []uint8{0x5, 0x31, 0x2e, 0x30, 0x31, 0xa2, 0xff, 0xbf},
				}
				count++
			}
		}
		done <- count
	}()

	count := <-done
	fmt.Printf("%d\n", count)
	time.Sleep(1 * time.Second)
	c.Done <- true

	<-forever
}
