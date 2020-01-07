package main

import (
	"log"
	"time"

	"github.com/sauravgsh16/ecu/controller"
)

func main() {
	c, err := controller.New(0)
	if err != nil {
		log.Fatalf(err.Error())
	}

	after := time.After(5 * time.Second)
	select {
	case <-after:
		c.Initiate()
	}
}
