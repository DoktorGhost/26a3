package main

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"time"
)

type RingIntBuffer struct {
	array []int
	pos   int
	size  int
	m     sync.Mutex
}

func NewRingIntBuffer(size int) *RingIntBuffer {
	return &RingIntBuffer{make([]int, size), -1, size, sync.Mutex{}}
}

func (r *RingIntBuffer) Push(el int) {
	r.m.Lock()
	defer r.m.Unlock()
	if r.pos == r.size-1 {
		for i := 1; i <= r.size-1; i++ {
			r.array[i-1] = r.array[i]
		}
		r.array[r.pos] = el
	} else {
		r.pos++
		r.array[r.pos] = el
	}
}

func (r *RingIntBuffer) Get() []int {
	if r.pos <= 0 {
		return nil
	}
	r.m.Lock()
	defer r.m.Unlock()
	var output []int = r.array[:r.pos]

	r.pos = 0
	return output
}

func read(input chan<- int) {
	for {
		var u int
		_, err := fmt.Scanf("%d\n", &u)
		if err != nil {
			fmt.Println("This isn`t a number")
		} else {
			input <- u
		}
	}
}

func removeNegatives(currentChan <-chan int, nextChan chan<- int) {
	for number := range currentChan {
		if number >= 0 {
			nextChan <- number
		}
	}
}

func removeDivThree(currentChan <-chan int, nextChan chan<- int) {
	for number := range currentChan {
		if number%3 != 0 {
			nextChan <- number
		}
	}
}

func writeToBuffer(currentChan <-chan int, r *RingIntBuffer) {
	for number := range currentChan {
		r.Push(number)
	}
}

func writeToConsole(r *RingIntBuffer, t *time.Ticker) {
	for range t.C {
		buffer := r.Get()
		if len(buffer) > 0 {
			fmt.Println("the buffer is", buffer)
		}
	}
}

func main() {
	size := 20

	input := make(chan int)
	go read(input)

	negFilterChannel := make(chan int)
	go removeNegatives(input, negFilterChannel)

	divThreeChannel := make(chan int)
	go removeDivThree(negFilterChannel, divThreeChannel)

	r := NewRingIntBuffer(size)
	go writeToBuffer(divThreeChannel, r)

	delay := 5
	ticker := time.NewTicker(time.Second * time.Duration(delay))
	go writeToConsole(r, ticker)

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt)
	select {
	case sig := <-c:
		fmt.Println("Got %s signal. Aborting ...", sig)
		os.Exit(0)
	}

}
