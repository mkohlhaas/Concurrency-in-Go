package main

import (
	"fmt"
)

func multiply[T ~int | ~int64 | ~float64](values []T, multiplier T) []T {
	multipliedValues := make([]T, len(values))

	for i, v := range values {
		multipliedValues[i] = v * multiplier
	}

	return multipliedValues
}

func add[T ~int | ~int64 | ~float64](values []T, addend T) []T {
	addedValues := make([]T, len(values))

	for i, v := range values {
		addedValues[i] = v + addend
	}

	return addedValues
}

func main() {
	ints := []int{1, 2, 3, 4}

	fmt.Println(ints)
	fmt.Println(multiply(add(multiply(ints, 2), 1), 2)) // *2 -> +1 -> *2
}
