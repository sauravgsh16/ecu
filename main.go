package main

import (
	"fmt"

	"github.com/sauravgsh16/ecu/domain"
	"github.com/sauravgsh16/ecu/util"
)

func main() {
	s, err := domain.GenerateRandom(16)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("%s\n", s)

	h := util.GenerateHash([]byte(s))
	fmt.Printf("%s\n", h)
}
