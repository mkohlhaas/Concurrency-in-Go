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

func add[T ~int | ~int64 | ~float64](values []T, additive T) []T {
	addedValues := make([]T, len(values))
	for i, v := range values {
		addedValues[i] = v + additive
	}
	return addedValues
}

func main() {
	ints := []int{1, 2, 3, 4}
	for _, v := range add(multiply(ints, 2), 1) {
		fmt.Println(v)
	}
}
