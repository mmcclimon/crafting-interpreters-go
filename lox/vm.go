package lox

import (
	"errors"
	"fmt"
)

var InterpretCompileError = errors.New("compile error")
var InterpretRuntimeError = errors.New("runtime error")

type VM struct {
	chunk *Chunk
	ip    int
}

func NewVM() *VM {
	return &VM{}
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
		}

		instruction := vm.readByte()

		switch instruction {
		case byte(OP_CONSTANT):
			constant := vm.readConstant()
			constant.Print()
			fmt.Printf("\n")

		case byte(OP_RETURN):
			return nil
		}
	}

	return nil
}

func (vm *VM) readByte() byte {
	ret := vm.chunk.code[vm.ip]
	vm.ip += 1
	return ret
}

func (vm *VM) readConstant() Value {
	return vm.chunk.constantAt(vm.readByte())
}
