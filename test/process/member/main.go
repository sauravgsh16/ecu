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

	fmt.Println("Started member")

	/*
		after := time.After(5 * time.Second)
		select {
		case <-after:
			c.Initiate()
		}
	*/
	fmt.Printf("%#v\n", c)
}
