package main

import (
	"context"
	"fmt"
	"log"

	"github.com/sauravgsh16/ecu/clientnew"
)

const (
	url = "tcp://localhost:9000"
)

func main() {
	bs := clientnew.BroadcastSubscribe{
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
	s, err := clientnew.NewSubscriber(config)
	if err != nil {
		log.Fatalf(err.Error())
	}
	ctx := context.Background()

	ch, err := s.Subscribe(ctx)

	for msg := range ch {
		fmt.Printf("%#v", msg)
	}

	/*
		timeout := time.After(30 * time.Second)

		for {
			select {
			case msg := <-ch:
				fmt.Printf("%#v", msg)

			case <-timeout:
				ctx.Done()
			}
		}
	*/
}
