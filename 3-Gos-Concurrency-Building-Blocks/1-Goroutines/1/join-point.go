package main

import (
	"fmt"
	"sync"
)

func main() {
	var wg sync.WaitGroup

	sayHello := func() {
		fmt.Println("hello")
	}

	wg.Go(sayHello)

	wg.Wait()
}
