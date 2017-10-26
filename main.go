package main

import (
	"fmt"
	"math/rand"
	"time"

	"code.cloudfoundry.org/clock/fakeclock"
)

// --- Concurrency 101 ---
func goroutine() {
	// Start a goroutine and execute println concurrently
	go println("goroutine message")
	println("main function message")
}

func chanSimple1() {
	c := make(chan int)
	c <- 42    // write to a channel
	val := <-c // read from a channel
	println(val)
}

func chanSimple2() {
	c := make(chan int)
	go func() {
		c <- 42 // write to a channel
	}()
	val := <-c // read from a channel
	println(val)
}

func sentinel1() {
	// Create a channel to synchronize goroutines
	done := make(chan bool)

	// Execute println in goroutine
	go func() {
		time.Sleep(1 * time.Second)
		println("goroutine message")

		// Tell the main function everything is done.
		// This channel is visible inside this goroutine because
		// it is executed in the same address space.
		done <- true
	}()

	println("main function message")
	<-done // Wait for the goroutine to finish
}

func noBuffer() {
	message := make(chan string) // no buffer
	count := 3

	go func() {
		for i := 1; i <= count; i++ {
			fmt.Println("send message")
			message <- fmt.Sprintf("message %d", i)
		}
	}()

	time.Sleep(time.Second * 3)

	for i := 1; i <= count; i++ {
		fmt.Println(<-message)
	}
}

func withBuffer() {
	message := make(chan string, 2) // with buffer 2
	count := 3

	go func() {
		for i := 1; i <= count; i++ {
			fmt.Println("send message")
			message <- fmt.Sprintf("message %d", i)
		}
	}()

	time.Sleep(time.Second * 3)

	for i := 1; i <= count; i++ {
		fmt.Println(<-message)
	}
}

func rangeChan() {
	message := make(chan string)
	count := 3

	go func() {
		for i := 1; i <= count; i++ {
			message <- fmt.Sprintf("message %d", i)
		}
		close(message)
	}()

	for msg := range message {
		fmt.Println(msg)
	}
}

func closeChanNotBlock() {
	done := make(chan bool)
	close(done)

	// Will not block and will print false twice
	// because it's the default value for bool type
	println(<-done)
	println(<-done)
}

func sentinel2() {
	// Data is irrelevant
	done := make(chan struct{})

	go func() {
		println("goroutine message")

		// Just send a signal "I'm done"
		close(done)
	}()

	println("main function message")
	<-done
}

func selectChan() {
	// For our example we'll select across two channels.
	c1 := make(chan string)
	c2 := make(chan string)

	// Each channel will receive a value after some amount
	// of time, to simulate e.g. blocking RPC operations
	// executing in concurrent goroutines.
	go func() {
		time.Sleep(time.Second * 1)
		c1 <- "one"
	}()
	go func() {
		time.Sleep(time.Second * 2)
		c2 <- "two"
	}()

	// We'll use `select` to await both of these values
	// simultaneously, printing each one as it arrives.
	for i := 0; i < 2; i++ {
		select {
		case msg1 := <-c1:
			fmt.Println("received", msg1)
		case msg2 := <-c2:
			fmt.Println("received", msg2)
		}
	}
}

// --- Thread-Safe Operation ---
var done = func() {}

var (
	chDelete = make(chan string)
	chSet    = make(chan string)
	chQuit   = make(chan bool)
	set      = make(map[string]bool)
)

func opSet1() {
	for {
		select {
		case v := <-chSet:
			fmt.Printf("set %s\n", v)
			set[v] = true
		case v := <-chDelete:
			fmt.Printf("delete %s\n", v)
			delete(set, v)
		case <-chQuit:
			return
		}
	}
}

func opSet2() {
	for {
		select {
		case v := <-chSet:
			fmt.Printf("set %s\n", v)
			set[v] = true
			done()
		case v := <-chDelete:
			fmt.Printf("delete %s\n", v)
			delete(set, v)
			done()
		case <-chQuit:
			return
		}
	}
}

// --- Worker Pools ---
func worker1(id int, jobs <-chan int) {
	for j := range jobs {
		r := rand.Intn(100)
		time.Sleep(time.Duration(r) * time.Millisecond)
		fmt.Printf("finished: worker[%d], job[%d]\n", id, j)
	}
}

func dispatch1(nw, nj int) {
	jobs := make(chan int, 100)

	for w := 1; w <= nw; w++ {
		go worker1(w, jobs)
	}

	for j := 1; j <= nj; j++ {
		jobs <- j
	}
	close(jobs)
}

func worker2(id int, jobs <-chan int, done func()) {
	for j := range jobs {
		r := rand.Intn(100)
		time.Sleep(time.Duration(r) * time.Millisecond)
		done()
		fmt.Printf("finished: worker[%d], job[%d]\n", id, j)
	}
}

func dispatch2(nw, nj int) {
	jobs := make(chan int, 100)

	for w := 1; w <= nw; w++ {
		go worker2(w, jobs, done)
	}

	for j := 1; j <= nj; j++ {
		jobs <- j
	}
	close(jobs)
}

// --- Polling ---
var (
	timeout   = 5 * time.Second
	interval  = 1 * time.Second
	numOfTick = 0
	pollFn    = func() error { return nil }
)

func polling1() error {
	chTo := time.NewTimer(timeout).C
	chTk := time.NewTicker(interval).C
	for {
		select {
		case <-chTo:
			fmt.Println("timeout")
			return fmt.Errorf("timeout")
		case <-chTk:
			numOfTick++
			fmt.Printf("tick %d\n", numOfTick)

			err := pollFn()
			if err != nil {
				continue
			}
			return nil
		}
	}
}

var fc = fakeclock.NewFakeClock(time.Now())
var tickDone = func() {}

func polling2() error {
	chTo := fc.NewTimer(timeout).C()
	chTk := fc.NewTicker(interval).C()

	for {
		select {
		case <-chTo:
			fmt.Println("timeout")
			return fmt.Errorf("timeout")
		case <-chTk:
			numOfTick++
			fmt.Printf("tick %d\n", numOfTick)
			tickDone()

			err := pollFn()
			if err != nil {
				continue
			}
			return nil
		}
	}
}

func main() {
}
