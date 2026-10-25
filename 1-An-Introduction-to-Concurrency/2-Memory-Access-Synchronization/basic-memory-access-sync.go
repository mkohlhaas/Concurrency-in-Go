package main

import (
	"fmt"
	"sync"
)

func main() {
	var memoryAccess sync.Mutex
	var value int

	go func() {
		memoryAccess.Lock()
		value++
		memoryAccess.Unlock()
	}()

	memoryAccess.Lock()
	if value == 0 {
		fmt.Printf("value: %d\n", value)
	} else {
		fmt.Printf("value: %d\n", value)
	}
	memoryAccess.Unlock()
}
