package main

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_dispatch1(t *testing.T) {
	nw, nj := 3, 10
	dispatch1(nw, nj)
	time.Sleep(3 * time.Second)
}

func Test_dispatch2(t *testing.T) {
	nw, nj := 3, 10

	var wg sync.WaitGroup
	wg.Add(nj)
	done = func() {
		wg.Done()
	}
	dispatch2(nw, nj)

	wg.Wait()

	// reset
	done = func() {}
}

func Test_opSet1(t *testing.T) {
	go opSet1()
	chSet <- "foo"
	time.Sleep(1 * time.Second)
	assert.True(t, set["foo"])

	// reset
	close(chQuit)
	chQuit = make(chan bool)
}

func Test_opSet2(t *testing.T) {
	chDone := make(chan struct{}, 1)
	done = func() {
		chDone <- struct{}{}
	}

	go opSet2()
	chSet <- "foo"

	<-chDone
	assert.True(t, set["foo"])

	chDelete <- "foo"

	<-chDone
	_, ok := set["foo"]
	assert.False(t, ok)

	// reset
	close(chQuit)
	chQuit = make(chan bool)
}

func Test_polling1Timeout(t *testing.T) {
	pollFn = func() error {
		return fmt.Errorf("err")
	}

	var err error
	go func() {
		err = polling1()
	}()

	time.Sleep(6 * time.Second)

	assert.Equal(t, 4, numOfTick)
	assert.Error(t, err)

	// reset
	numOfTick = 0
	pollFn = func() error { return nil }
}

func Test_polling1Success(t *testing.T) {
	pollFn = func() error {
		if numOfTick == 1 {
			return fmt.Errorf("err")
		}
		return nil
	}

	var err error
	go func() {
		err = polling1()
	}()

	time.Sleep(3 * time.Second)

	assert.Equal(t, 2, numOfTick)
	assert.NoError(t, err)

	// reset
	numOfTick = 0
	pollFn = func() error { return nil }
}

func Test_polling2Timeout(t *testing.T) {
	pollFn = func() error {
		return fmt.Errorf("err")
	}

	chDone := make(chan struct{})
	var err error
	go func() {
		err = polling2()
		close(chDone)
	}()

	fc.WaitForNWatchersAndIncrement(timeout, 2)
	<-chDone

	assert.Error(t, err)

	// reset
	numOfTick = 0
	pollFn = func() error { return nil }
}

func Test_polling2Success(t *testing.T) {
	chTickDone := make(chan struct{}, 1)
	tickDone = func() {
		chTickDone <- struct{}{}
	}

	pollFn = func() error {
		if numOfTick == 1 {
			return fmt.Errorf("err")
		}
		return nil
	}

	chDone := make(chan struct{})
	var err error
	go func() {
		err = polling2()
		close(chDone)
	}()

	fc.WaitForNWatchersAndIncrement(interval, 2)
	<-chTickDone
	assert.Equal(t, 1, numOfTick)
	assert.Nil(t, err)

	fc.WaitForNWatchersAndIncrement(interval, 2)
	<-chTickDone
	assert.Equal(t, 2, numOfTick)

	<-chDone
	assert.NoError(t, err)

	// reset
	numOfTick = 0
	pollFn = func() error { return nil }
}
