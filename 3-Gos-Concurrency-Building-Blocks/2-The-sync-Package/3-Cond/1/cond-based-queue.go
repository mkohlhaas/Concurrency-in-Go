package main

import (
	"fmt"
	"sync"
	"time"
)

func main() {
	const capacity = 10
	queue := make([]any, 0, capacity)
	c := sync.NewCond(&sync.Mutex{})

	removeFromQueue := func(delay time.Duration) {
		time.Sleep(delay)
		c.L.Lock()
		queue = queue[1:]
		fmt.Println("Removed from queue")
		c.L.Unlock()
		c.Signal()
	}

	for i := 0; i < capacity; i++ {
		c.L.Lock()
		for len(queue) > 2 {
			c.Wait()
		}
		fmt.Println("Adding to queue")
		queue = append(queue, struct{}{})
		go removeFromQueue(1 * time.Second)
		c.L.Unlock()
	}
}
