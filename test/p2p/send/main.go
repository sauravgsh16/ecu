package main

import (
	"time"

	"github.com/sauravgsh16/ecu/client"
	"github.com/sauravgsh16/ecu/handler"
)

func main() {
	payload := client.Payload([]byte("Let's test sending this as Sn!!"))
	msg := client.NewMessage(payload)

	s, err := handler.NewSendSnSender(msg)
	if err != nil {
		panic(err)
	}
	for i := 0; i < 10000; i++ {
		s.Send()
		time.Sleep(5 * time.Millisecond)
	}
}
