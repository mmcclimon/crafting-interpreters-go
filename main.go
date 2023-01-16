package main

import (
	"github.com/mmcclimon/glox/lox"
)

func main() {
	chunk := lox.NewChunk()

	constant := chunk.AddConstant(1.2)
	chunk.Write(byte(lox.OP_CONSTANT), 123)
	chunk.Write(byte(constant), 123)

	chunk.Write(byte(lox.OP_RETURN), 123)

	chunk.Disassemble("test chunk")
}
