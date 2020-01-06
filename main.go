package main

import (
	"fmt"

	"github.com/sauravgsh16/ecu/servicenew"
)

func main() {
	l, err := servicenew.NewLeader()
	if err != nil {
		panic(err)
	}

	fmt.Printf("%+v", l)
}
