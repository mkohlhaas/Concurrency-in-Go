package main

import (
	"fmt"
)

func main() {
	var data int

	go func() {
		data++
	}()

	//	time.Sleep(1 * time.Second) // ⚠️ this is bad!

	if data == 0 {
		fmt.Printf("value: %v\n", data)
	}
}
