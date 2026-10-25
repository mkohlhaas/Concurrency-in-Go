package main

import (
	"fmt"
)

func multiply[T ~int | ~int64 | ~float64](value, multiplier T) T {
	return value * multiplier
}

func add[T ~int | ~int64 | ~float64](value, additive T) T {
	return value + additive
}

func main() {
	ints := []int{1, 2, 3, 4}
	for _, v := range ints {
		fmt.Println(multiply(add(multiply(v, 2), 1), 2))
	}
}
