package lox

import "fmt"

// this all feels fairly kludgey to me...
type Value interface {
	Equals(Value) bool
}
type ValueArray []Value

type ValueBool bool
type ValueNil int8
type ValueNumber float64
type ValueString string

type ValueFunction struct {
	arity        int
	upvalueCount int
	chunk        *Chunk
	name         string
}

type ValueClosure struct {
	function ValueFunction
}

type NativeFn func(argCount int, args []Value) Value

type ValueNative struct {
	function NativeFn
}

func PrintValue(v Value) {
	switch v.(type) {
	case ValueBool:
		val := v.(ValueBool)
		fmt.Printf("%v", val)
	case ValueNumber:
		fmt.Printf("%g", v)
	case ValueNil:
		fmt.Printf("nil")
	case ValueString:
		fmt.Print(v)
	case ValueFunction:
		function := v.(ValueFunction)
		name := function.name
		if name == "" {
			name = "<script>"
		}
		fmt.Printf("<fn %s>", name)
	case ValueClosure:
		PrintValue(v.(ValueClosure).function)
	case ValueNative:
		fmt.Printf("<native fn>")
	default:
		fmt.Printf("wat? %T", v)
	}
}

func IsFalsy(v Value) bool {
	switch v.(type) {
	case ValueBool:
		val := v.(ValueBool)
		return val == false
	case ValueNil:
		return true
	default:
		return false
	}
}

func NewValueArray() *ValueArray {
	va := make(ValueArray, 0, 8)
	return &va
}

func (va *ValueArray) Write(item Value) {
	*va = append(*va, item)
}

func NewFunction() *ValueFunction {
	return &ValueFunction{
		chunk: NewChunk(),
	}
}

func (v ValueBool) Equals(other Value) bool {
	x, isBool := other.(ValueBool)
	return isBool && v == x
}

func (v ValueNumber) Equals(other Value) bool {
	x, isNum := other.(ValueNumber)
	return isNum && v == x
}

func (v ValueNil) Equals(other Value) bool {
	_, isNil := other.(ValueNil)
	return isNil
}

func (v ValueString) Equals(other Value) bool {
	x, isStr := other.(ValueString)
	return isStr && v == x
}

func (v ValueFunction) Equals(other Value) bool {
	return false
}

func (v ValueNative) Equals(other Value) bool {
	return false
}

func (v ValueClosure) Equals(other Value) bool {
	return false
}
