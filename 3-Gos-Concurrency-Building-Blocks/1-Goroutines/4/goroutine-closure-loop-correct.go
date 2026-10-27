package main

import (
	"fmt"
	"sync"
)

func main() {
	var wg sync.WaitGroup

	for _, salutation := range []string{"hello", "greetings", "good day"} {
		wg.Go(func() {
			fmt.Println(salutation)
		})
	}

	wg.Wait()
}
