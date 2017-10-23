package main

import (
	"fmt"
	"math/rand"
	"time"

	"code.cloudfoundry.org/clock/fakeclock"
)

var done = func() {}

// Thread-Safe Operation
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

// Worker Pools
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

// polling
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
