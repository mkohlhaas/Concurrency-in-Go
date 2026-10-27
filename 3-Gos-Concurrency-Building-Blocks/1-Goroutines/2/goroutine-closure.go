package main

import (
	"fmt"
	"sync"
)

func main() {
	var wg sync.WaitGroup
	salutation := "hello"

	wg.Go(func() {
		salutation = "welcome"
	})

	wg.Wait()

	fmt.Println(salutation)
}
