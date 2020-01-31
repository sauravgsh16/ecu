package main

import (
	"bufio"
	"os"

	"github.deere.com/sg30983/ecu/client"
	"github.deere.com/sg30983/ecu/handler"
)

func main() {
	s, err := handler.NewSendSnSender()
	if err != nil {
		panic(err)
	}

	r := bufio.NewReader(os.Stdin)
	for {
		l, err := r.ReadBytes('\n')
		if err != nil {
			panic(err)
		}

		payload := client.Payload([]byte(l))
		msg := client.NewMessage(payload)

		s.Send(msg)
	}
}
