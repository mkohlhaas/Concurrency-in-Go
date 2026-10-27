package main

import (
	"fmt"
	"sync"
	"time"
)

func main() {
	var wg sync.WaitGroup

	wg.Go(func() {
		fmt.Println("1st goroutine sleeping...")
		time.Sleep(1 * time.Second)
		fmt.Println("1st goroutine finished sleeping")
	})

	wg.Go(func() {
		fmt.Println("2nd goroutine sleeping...")
		time.Sleep(2 * time.Second)
		fmt.Println("2nd goroutine finished sleeping")
	})

	wg.Wait()

	fmt.Println("All goroutines complete.")
}
