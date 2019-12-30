package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/sauravgsh16/ecu/clientnew"
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

const (
	url = "tcp://localhost:9000"
)

type Request struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type AnnounceSn struct {
	clientnew.Publisher
	Message *clientnew.Message
}

func main() {
	var req Request

	b := []byte(`{"name": "saurav", "email": "test@foo.com"}`)

	if err := json.Unmarshal(b, &req); err != nil {
		fmt.Println(err)
	}

	bp := clientnew.BroadCastPublish{
		URI:            url,
		ExchangeName:   "test",
		ExchangeNoWait: false,
		Immediate:      false,
	}

	config := bp.Marshal()
	msg := &clientnew.Message{
		// UUID:    fmt.Sprintf("%s", uuid.Must(uuid.NewV4())),
		UUID:    "fuck_that_shit",
		Payload: clientnew.Payload([]byte("a test string")),
		Metadata: clientnew.Metadata(map[string]interface{}{
			"ContentType":   "text/plain",
			"MessageID":     "msgID",
			"UserID":        "userid",
			"ApplicationID": "aapid",
		}),
	}
	p, _ := clientnew.NewPublisher(config)
	asn := AnnounceSn{p, msg}
	defer asn.Close()

	if err := asn.Publish(msg); err != nil {
		fmt.Println(err)
	}

	fmt.Println("Success")

	c := make(chan struct{})

	go test(c)
	select {
	case <-c:
		fmt.Printf("received close\n")
	}
}

func test(c chan struct{}) {
	timeout := time.After(5 * time.Second)
	select {
	case <-timeout:
		close(c)
	}
}
