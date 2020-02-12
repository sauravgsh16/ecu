package main

import (
	"fmt"
	"io"
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
		"Xtd 01 1CECFF12 08 20 01 01 25 FF 03 FF 01\n",
		"Xtd 01 1CEBFF12 08 01 53 43 46 30 4D 79 20\n",
		"Xtd 01 1CEBFF12 08 02 45 43 55 00 00 00 00\n",
		"Xtd 01 1CEBFF12 08 03 00 00 00 00 00 00 00\n",
		"Xtd 01 1CEBFF12 08 04 00 00 00 00 00 00 00\n",
		"Xtd 01 1CEBFF12 08 05 00 00 00 00 00 00 00\n",
		"Xtd 01 1CEBFF12 08 06 00 00 00 00 00 00 00\n",
		"Xtd 01 1CEBFF12 08 07 00 00 00 00 00 00 00\n",
		"Xtd 01 1CEBFF12 08 08 00 00 00 00 00 00 00\n",
		"Xtd 01 1CEBFF12 08 09 00 00 00 00 00 00 00\n",
		"Xtd 01 1CEBFF12 08 0A 00 00 00 00 00 00 52\n",
		"Xtd 01 1CEBFF12 08 0B 6F 6F 74 00 00 00 00\n",
		"Xtd 01 1CEBFF12 08 0C 00 00 00 00 00 00 00\n",
		"Xtd 01 1CEBFF12 08 0D 00 00 00 00 00 00 00\n",
		"Xtd 01 1CEBFF12 08 0E 00 00 00 00 00 00 00\n",
		"Xtd 01 1CEBFF12 08 0F 00 00 00 00 00 00 00\n",
		"Xtd 01 1CEBFF12 08 10 00 00 00 00 00 00 00\n",
		"Xtd 01 1CEBFF12 08 11 00 00 00 00 00 00 00\n",
		"Xtd 01 1CEBFF12 08 12 00 00 00 00 00 00 00\n",
		"Xtd 01 1CEBFF12 08 13 00 00 00 00 00 00 00\n",
		"Xtd 01 1CEBFF12 08 14 00 30 31 32 33 34 35\n",
		"Xtd 01 1CEBFF12 08 15 36 37 38 39 00 A7 62\n",
		"Xtd 01 1CEBFF12 08 16 E8 A1 BC EA C3 7F D0\n",
		"Xtd 01 1CEBFF12 08 17 25 D9 5C E4 FE 07 C2\n",
		"Xtd 01 1CEBFF12 08 18 7C 14 79 A9 5F B5 E1\n",
		"Xtd 01 1CEBFF12 08 19 87 B7 16 43 74 49 EB\n",
		"Xtd 01 1CEBFF12 08 1A C2 A1 18 2A 75 6B FC\n",
		"Xtd 01 1CEBFF12 08 1B 55 50 76 21 3A CA E7\n",
		"Xtd 01 1CEBFF12 08 1C 9A D9 AB 77 E6 BD 9D\n",
		"Xtd 01 1CEBFF12 08 1D F0 D4 1D A4 27 C3 A0\n",
		"Xtd 01 1CEBFF12 08 1E E4 53 0E 15 51 7A D9\n",
		"Xtd 01 1CEBFF12 08 1F BC 64 32 3F 9E 2F 3A\n",
		"Xtd 01 1CEBFF12 08 20 41 15 89 AC C5 F5 97\n",
		"Xtd 01 1CEBFF12 08 21 E7 56 23 B1 F9 3C 3D\n",
		"Xtd 01 1CEBFF12 08 22 A4 8B 43 1B 10 B1 89\n",
		"Xtd 01 1CEBFF12 08 23 EE 46 00 C0 51 93 68\n",
		"Xtd 01 1CEBFF12 08 24 C4 5C 46 09 E9 38 FF\n",
		"Xtd 01 1CEBFF12 08 25 37 8C 77 E1 2B FF FF\n",
	}

	tp := []string{
		"Xtd 01 1CECFF12 08 20 10 00 03 FF 02 FF 01\n",
		"Xtd 01 1CEBFF12 08 01 8B 45 00 24 B5 9D D9\n",
		"Xtd 01 1CEBFF12 08 02 68 D0 91 5F 05 B9 1D\n",
		"Xtd 01 1CEBFF12 08 03 5B B3 FF FF FF FF FF\n",
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Fatalf(err.Error())
		}
		go read(conn)

		for _, t := range tp {
			//conn.Write([]byte("Xtd 02 0CCBF782 08 13 00 86 00 B8 0B 00 00\n"))
			conn.Write([]byte(t)) // 13 00 86 00 B8 0B 00 00\n"))
		}
	}

}

func read(c net.Conn) {
	for {
		b := make([]byte, 43)
		if _, err := io.ReadFull(c, b[:43]); err != nil {
			if err == io.EOF {
				break
			}
			if netErr, ok := err.(net.Error); ok {
				log.Printf(netErr.Error())
				break
			}
		}
		fmt.Printf("%s", string(b))
	}
}
