package lox

import "fmt"

// this all feels fairly kludgey to me...
type Value interface {
	AsNumber() float64
	AsBool() bool
}
type ValueArray []Value

type ValueBool bool
type ValueNil int8
type ValueNumber float64

func PrintValue(v Value) {
	switch v.(type) {
	case ValueBool:
		fmt.Printf("%v", v.AsBool())
	case ValueNumber:
		fmt.Printf("%g", v)
	case ValueNil:
		fmt.Printf("nil")
	}
}

func NewValueArray() *ValueArray {
	va := make(ValueArray, 0, 8)
	return &va
}

func (va *ValueArray) Write(item Value) {
	*va = append(*va, item)
}

func (v ValueBool) AsBool() bool        { return bool(v) }
func (v ValueNumber) AsNumber() float64 { return float64(v) }

func (v ValueBool) AsNumber() float64 { panic("AsNumber called on a ValueBool") }
func (v ValueNil) AsNumber() float64  { panic("AsNumber called on a ValueNil") }
func (v ValueNumber) AsBool() bool    { panic("AsBool called on a ValueNumber") }
func (v ValueNil) AsBool() bool       { panic("AsBool called on a ValueNil") }
