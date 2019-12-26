package main

import (
	"encoding/json"
	"fmt"
	// "log"
)

//"io/ioutil"
// "os"

// "github.com/sauravgsh16/ecu/util"

/*
func main() {
	e, err := domain.NewEcu(domain.Leader)
	if err != nil {
		log.Fatalf("%s", err)
	}
	// e.GenerateNonce()

	fmt.Printf("%+v\n", e)
}
*/

type Request struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

func main() {
	var req Request

	b := []byte(`{"name": "saurav", "email": "test@foo.com"}`)
	fmt.Printf("%+v\n", b)

	if err := json.Unmarshal(b, &req); err != nil {
		fmt.Println(err)
	}

	fmt.Printf("%+v\n", req)
}
