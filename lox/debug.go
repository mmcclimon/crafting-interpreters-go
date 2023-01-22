package lox

import "fmt"

const DEBUG_PRINT_CODE = true
const DEBUG_TRACE_EXECUTION = false

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

	case OP_GET_LOCAL, OP_SET_LOCAL, OP_CALL:
		return byteInstruction(s, c, offset)

	case OP_JUMP, OP_JUMP_IF_FALSE, OP_JUMP_IF_TRUE:
		return jumpInstruction(s, 1, c, offset)

	case OP_LOOP:
		return jumpInstruction(s, -1, c, offset)

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

func byteInstruction(name string, chunk *Chunk, offset int) int {
	slot := chunk.code[offset+1]
	fmt.Printf("%-16s %4d\n", name, slot)
	return offset + 2
}

func jumpInstruction(name string, sign int, chunk *Chunk, offset int) int {
	jump := int(chunk.code[offset+1] << 8)
	jump |= int(chunk.code[offset+2])
	fmt.Printf("%-16s %4d -> %d\n", name, offset, offset+3+sign*jump)

	return offset + 3
}
