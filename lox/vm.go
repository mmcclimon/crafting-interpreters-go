package lox

import (
	"errors"
	"fmt"

	"github.com/mmcclimon/glox/lox/op"
)

const STACK_MAX = 256

var InterpretCompileError = errors.New("compile error")
var InterpretRuntimeError = errors.New("runtime error")

type VM struct {
	chunk *Chunk
	ip    int
	stack [STACK_MAX]Value
	sp    int
}

func NewVM() *VM {
	return &VM{}
}

func (vm *VM) InterpretString(source string) error {
	chunk := NewChunk()

	if !Compile(source, chunk) {
		return InterpretCompileError
	}

	vm.chunk = chunk
	vm.ip = 0

	return vm.run()
}

func (vm *VM) Interpret(chunk *Chunk) error {
	vm.chunk = chunk
	vm.ip = 0
	return vm.run()
}

func (vm *VM) run() error {
	for {
		if DEBUG_TRACE_EXECUTION {
			vm.chunk.DisassembleInstruction(vm.ip)
			fmt.Printf("          ")
			for i := 0; i < vm.sp; i++ {
				fmt.Printf("[ ")
				vm.stack[i].Print()
				fmt.Printf(" ]")
			}
			fmt.Printf("\n")
		}

		instruction := vm.readByte()

		switch instruction {
		case byte(OP_CONSTANT):
			constant := vm.readConstant()
			vm.push(constant)

		case byte(OP_ADD):
			vm.binaryOp(op.Plus)

		case byte(OP_SUBTRACT):
			vm.binaryOp(op.Minus)

		case byte(OP_MULTIPLY):
			vm.binaryOp(op.Mul)

		case byte(OP_DIVIDE):
			vm.binaryOp(op.Div)

		case byte(OP_NEGATE):
			vm.push(-vm.pop())

		case byte(OP_RETURN):
			vm.pop().Print()
			fmt.Printf("\n")
			return nil
		}
	}
}

func (vm *VM) readByte() byte {
	ret := vm.chunk.code[vm.ip]
	vm.ip++
	return ret
}

func (vm *VM) readConstant() Value {
	return vm.chunk.constantAt(vm.readByte())
}

// stack manipulation
func (vm *VM) push(value Value) {
	vm.stack[vm.sp] = value
	vm.sp++
}

func (vm *VM) pop() Value {
	vm.sp--
	return vm.stack[vm.sp]
}

func (vm *VM) binaryOp(oper op.BinaryOp) {
	b := vm.pop()
	a := vm.pop()

	var res Value
	switch oper {
	case op.Plus:
		res = a + b
	case op.Minus:
		res = a - b
	case op.Mul:
		res = a * b
	case op.Div:
		res = a / b
	}

	vm.push(res)
}
