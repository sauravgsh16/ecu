package main

import (
	"github.com/ar3s3ru/gobus"
)

func main() {
	b := gobus.NewEventBus()
	b.Subscribe(test)
}

func test(event string) {

}

/*
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
*/
