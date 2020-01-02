package main

import (
	"fmt"
	"sync"
)

func main() {
	in := make(chan int)
	var wg sync.WaitGroup

	go test(in, &wg)

	for i := 1; i <= 6; i++ {
		if i%3 == 0 {
			wg.Add(1)
		}
		in <- i
	}

	wg.Wait()
}

func test(in chan int, wg *sync.WaitGroup) {
	buf := make([]int, 0, 3)
	for {
		select {
		case i := <-in:
			buf = append(buf, i)
			fmt.Printf("buf :%+v\n", buf)
			if len(buf) == 3 {
				for _, i := range buf {
					fmt.Printf("Doing something with %d\n", i)
				}
				buf = nil
				fmt.Printf("Buf now %+v\n", buf)
				wg.Done()
			}
		}
	}
}
