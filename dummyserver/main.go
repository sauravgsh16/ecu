package main

import (
	"log"
	"net"
)

func main() {
	ln, err := net.Listen("tcp", ":19000")
	if err != nil {
		log.Fatalf(err.Error())
	}
	log.Println("Listening on :19000")

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Fatalf(err.Error())
		}
		i := 0
		for i < 1 {
			//conn.Write([]byte("Xtd 02 0CCBF782 08 13 00 86 00 B8 0B 00 00\n"))
			conn.Write([]byte("Xtd 02 0CCBF782")) // 13 00 86 00 B8 0B 00 00\n"))
			i++
		}
	}

}
