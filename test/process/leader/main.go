package main

import (
	"fmt"
	"log"
	"time"

	"github.com/sauravgsh16/ecu/controller"
)

func main() {
	c, err := controller.New(0)
	if err != nil {
		log.Fatalf(err.Error())
	}

	fmt.Printf("Started %T\n", c)

	_, err = c.Register()
	if err != nil {
		log.Fatalf(err.Error())
	}

	forever := make(chan interface{})

	fmt.Printf("Registered leader\n\n\n\n")

	d := make(chan interface{})
	idCh, errch := c.Wait(d)

	c.StartReceiverRoutines(idCh, errch)

	time.Sleep(5 * time.Second)

	c.Initiate()

	<-forever
}
