package main

import (
	"fmt"
	"time"
)

func main() {
	var c <-chan int

	select {
	case <-c:
	case <-time.After(2 * time.Second):
		fmt.Println("Timed out after two seconds.")
	}
}
