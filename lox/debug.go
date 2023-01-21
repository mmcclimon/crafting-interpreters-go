package lox

import "fmt"

const DEBUG_TRACE_EXECUTION = false
const DEBUG_PRINT_CODE = true

func (c *Chunk) Disassemble(name string) {
	fmt.Printf("== %s ==\n", name)

	for offset := 0; offset < len(c.code); {
		offset = c.DisassembleInstruction(offset)
	}
}

func (c *Chunk) DisassembleInstruction(offset int) int {
	fmt.Printf("%04d ", offset)

	lineNum := c.GetLine(offset)
	if offset > 0 && lineNum == c.GetLine(offset-1) {
		fmt.Printf("   | ")
	} else {
		fmt.Printf("%4d ", lineNum)
	}

	instruction := c.code[offset]

	s := OpCode(instruction).String()
	if s == "" {
		fmt.Printf("unknown opcode %d\n", instruction)
		return offset + 1
	}

	switch OpCode(instruction) {
	case OP_CONSTANT, OP_DEFINE_GLOBAL, OP_SET_GLOBAL, OP_GET_GLOBAL:
		return constantInstruction(s, c, offset)
	default:
		return simpleInstruction(s, offset)
	}
}

func simpleInstruction(name string, offset int) int {
	fmt.Printf("%s\n", name)
	return offset + 1
}

func constantInstruction(name string, chunk *Chunk, offset int) int {
	constant := chunk.code[offset+1]
	fmt.Printf("%-16s %4d '", name, constant)
	PrintValue(chunk.constantAt(constant)) // could be improved, probably
	fmt.Printf("'\n")

	return offset + 2
}
