package main

import (
	"fmt"
	"time"
)

// 1st pattern

func chanOwner() <-chan int {
	result := make(chan int, 5)

	go func() {
		defer close(result)
		for i := 0; i <= 5; i++ {
			result <- i
		}
	}()
	return result
}

func chanOwnerUsage() {
	stream := chanOwner()
	for result := range stream {
		fmt.Printf("Received: %d", result)
	}
}

// 2nd Pattern

func doWork(done <-chan interface{}, strings <-chan string) <-chan interface{} {
	terminate := make(chan interface{})
	go func() {
		defer close(terminate)

		for {
			select {
			case s := <-strings:
				fmt.Printf("%s\n", s)
			case <-done:
				return
			}
		}
	}()
	return terminate
}

func doWorkUsage() {
	done := make(chan interface{})
	terminate := doWork(done, nil)

	go func() {
		time.Sleep(1 * time.Second)
		fmt.Println("Cancelling doWork goroutine ....")
		close(done)
	}()

	<-terminate // Gets data when close(terminate) is called inside doWork
	fmt.Println("Done")
}

// Pipeline

// FanIn

// Fanout

// HeartBeat
