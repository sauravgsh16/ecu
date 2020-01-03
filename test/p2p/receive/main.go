package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/sauravgsh16/ecu/handler"
)

func main() {
	r, err := handler.NewSendSnReceiver()
	if err != nil {
		panic(err)
	}

	done := make(chan interface{})
	// stop := make(chan bool)

	out, err := r.StartReceiver(done)
	var counter int

	var wg sync.WaitGroup

	wg.Add(10000)

	now := time.Now()
	fmt.Println(now.Format("2006-01-02-15:04:05"))

	/*

		loop:
			for {
				select {
				case <-out:
					// fmt.Printf("%+v\n", string(m.Payload))
					counter++
					wg.Done()
				case <-stop:
					done <- true
					now := time.Now()
					fmt.Println(now.Format("2006-01-02-15:04:05"))
					break loop
				}
			}

	*/
	go func() {
		wg.Wait()
		now = time.Now()
		fmt.Println(now.Format("2006-01-02-15:04:05"))
		close(out)
	}()

	for range out {
		counter++
		wg.Done()
	}

	fmt.Printf("Received: %d messages\n", counter)
}
