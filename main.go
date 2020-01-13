package main

import (
	"fmt"
	"regexp"
)

func main() {
	str := "Join.abcd-efhecdsfks-sdfsdf2q312"
	pat := regexp.MustCompile(`^(?P<type>.*?)\.`)
	match := pat.FindStringSubmatch(str)

	fmt.Printf("%v: %s\n", pat.SubexpNames(), match)
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
