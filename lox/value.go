package lox

import "fmt"

type Value float64
type ValueArray []Value

func (v Value) Print() {
	fmt.Printf("%g", v)
}

func NewValueArray() *ValueArray {
	va := make(ValueArray, 0, 8)
	return &va
}

func (va *ValueArray) Write(item Value) {
	*va = append(*va, item)
}
