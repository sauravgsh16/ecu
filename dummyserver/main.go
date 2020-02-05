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
	_ = []string{
		"xtd 02 1CECF7E8 08 10 23 05 05 FF 00 CB 00\n",
		"xtd 02 1CEBF7E8 08 01 61 44 56 43 00 00 14\n",
		"xtd 02 1CEBF7E8 08 02 41 45 46 5F 53 50 52\n",
		"xtd 02 1CEBF7E8 08 03 41 59 45 52 28 53 43\n",
		"xtd 02 1CEBF7E8 08 04 33 41 29 2D 35 53 04\n",
		"xtd 02 1CEBF7E8 08 05 31 2E 30 31 A2 FF BF\n",
	}

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
