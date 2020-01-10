package main

import (
	"fmt"
	"log"

	"github.com/sauravgsh16/ecu/controller"
)

func main() {
	c, err := controller.New(1)
	if err != nil {
		log.Fatalf(err.Error())
	}

	fmt.Printf("Started %T\n", c)

	resp, err := c.Register()
	if err != nil {
		log.Fatalf(err.Error())
	}

	fmt.Printf("Registered member with id: %d\n\n\n\n", resp.Id)

	wait := make(chan bool)
	forever := make(chan interface{})

	go func() {
		r, err := c.Wait(resp.Id)
		if err != nil {
			close(wait)
			return
		}
		select {
		case wait <- r.Result:
		}
	}()

	_, ok := <-wait
	if !ok {
		log.Fatalf("Could not find valid ECUs")
	}
	close(wait)

	c.Initiate()
	<-forever
}
