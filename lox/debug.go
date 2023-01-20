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

	switch OpCode(instruction) {
	case OP_CONSTANT:
		return constantInstruction("OP_CONSTANT", c, offset)
	case OP_ADD:
		return simpleInstruction("OP_ADD", offset)
	case OP_SUBTRACT:
		return simpleInstruction("OP_SUBTRACT", offset)
	case OP_MULTIPLY:
		return simpleInstruction("OP_MULTIPLY", offset)
	case OP_DIVIDE:
		return simpleInstruction("OP_DIVIDE", offset)
	case OP_NOT:
		return simpleInstruction("OP_NOT", offset)
	case OP_NEGATE:
		return simpleInstruction("OP_NEGATE", offset)
	case OP_TRUE:
		return simpleInstruction("OP_TRUE", offset)
	case OP_FALSE:
		return simpleInstruction("OP_FALSE", offset)
	case OP_POP:
		return simpleInstruction("OP_POP", offset)
	case OP_GREATER:
		return simpleInstruction("OP_GREATER", offset)
	case OP_LESS:
		return simpleInstruction("OP_LESS", offset)
	case OP_DEFINE_GLOBAL:
		return constantInstruction("OP_DEFINE_GLOBAL", c, offset)
	case OP_GET_GLOBAL:
		return constantInstruction("OP_GET_GLOBAL", c, offset)
	case OP_EQUAL:
		return simpleInstruction("OP_EQUAL", offset)
	case OP_NIL:
		return simpleInstruction("OP_NIL", offset)
	case OP_PRINT:
		return simpleInstruction("OP_PRINT", offset)
	case OP_RETURN:
		return simpleInstruction("OP_RETURN", offset)

	default:
		fmt.Printf("unknown opcode %d\n", instruction)
		return offset + 1
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
