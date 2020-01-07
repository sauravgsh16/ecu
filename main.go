package main

import (
	"log"

	"github.com/sauravgsh16/ecu/controller"
)

func main() {
	c, err := controller.New(0)
	if err != nil {
		log.Fatalf(err.Error())
	}
	c.Initiate()
}
