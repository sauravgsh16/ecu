package main

import (
	"fmt"

	"github.com/sauravgsh16/ecu/service"
)

func main() {
	l, err := service.NewLeader()
	if err != nil {
		panic(err)
	}

	fmt.Printf("%+v", l)
}
