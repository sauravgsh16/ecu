package main

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"log"
	"time"
)

type mytimer struct {
	*time.Timer
	end time.Time
}

func main() {
	b := make([]byte, 16)
	buf := bytes.NewBuffer(b)

	if _, err := rand.Read(buf.Bytes()); err != nil {
		log.Fatalf(err.Error())
	}
	log.Println(buf.Bytes())
}

func selectTimer(t *mytimer, f chan bool) {
	go func() {
		for {
			select {
			case <-t.C:
				fmt.Println("returning after timer expired")
			}
			fmt.Println("exiting go routine")
			break
		}
		f <- true
	}()
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
