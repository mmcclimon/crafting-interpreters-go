package main

import (
	"github.com/mmcclimon/glox/lox"
)

func main() {
	chunk := lox.NewChunk()

	constant := chunk.AddConstant(1.2)
	chunk.Write(byte(lox.OP_CONSTANT), 123)
	chunk.Write(byte(constant), 123)

	constant = chunk.AddConstant(3.4)
	chunk.Write(byte(lox.OP_CONSTANT), 123)
	chunk.Write(byte(constant), 123)

	chunk.Write(byte(lox.OP_ADD), 123)

	constant = chunk.AddConstant(5.6)
	chunk.Write(byte(lox.OP_CONSTANT), 124)
	chunk.Write(byte(constant), 124)

	chunk.Write(byte(lox.OP_DIVIDE), 124)
	chunk.Write(byte(lox.OP_NEGATE), 124)

	chunk.Write(byte(lox.OP_RETURN), 124)
	// chunk.Disassemble("test chunk")

	vm := lox.NewVM()
	vm.Interpret(chunk)
}
