package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.deere.com/sg30983/ecu/client"
)

const (
	url = "tcp://localhost:9000"
)

func main() {
	bs := client.BroadcastSubscribe{
		URI:            url,
		ExchangeName:   "test",
		ExchangeNoWait: false,
		QueueName:      "qtest1",
		ConsumerName:   "c1",
		RoutingKey:     "",
		QueueNoWait:    false,
		BindNoWait:     false,
		ConsumerNoAck:  true,
		ConsumerNoWait: false,
	}

	config := bs.Marshal()
	s, err := client.NewSubscriber(config)
	if err != nil {
		log.Fatalf(err.Error())
	}
	ctx := context.Background()

	ch, err := s.Subscribe(ctx)

	timeout := time.After(30 * time.Second)

loop:
	for {
		select {
		case msg := <-ch:
			fmt.Printf("%#v\n", msg)

		case <-timeout:
			ctx.Done()
			break loop
		}
	}

}
