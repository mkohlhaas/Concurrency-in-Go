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
		fmt.Printf("The value is %v.\n", value)
	} else {
		fmt.Printf("The value is %v.\n", value)
	}
	memoryAccess.Unlock()
}
