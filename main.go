package main

import (
	"fmt"
	"io/ioutil"
	// "os"
	// "github.com/sauravgsh16/ecu/domain"
	// "github.com/sauravgsh16/ecu/util"
)

func main() {
	data, err := ioutil.ReadFile("./cert/vin1.der")
	if err != nil {
		fmt.Println(err)
	}

	fmt.Printf("%+v\n", data)
}
