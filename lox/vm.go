package lox

import (
	"errors"
	"fmt"
	"os"

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
	compiler := NewCompiler(source)

	if !compiler.Compile() {
		return InterpretCompileError
	}

	vm.chunk = compiler.chunk
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
				PrintValue(vm.stack[i])
				fmt.Printf(" ]")
			}
			fmt.Printf("\n")
		}

		instruction := vm.readByte()

		switch OpCode(instruction) {
		case OP_CONSTANT:
			constant := vm.readConstant()
			vm.push(constant)

		case OP_ADD:
			if err := vm.binaryOp(op.Plus); err != nil {
				return err
			}

		case OP_SUBTRACT:
			if err := vm.binaryOp(op.Minus); err != nil {
				return err
			}

		case OP_MULTIPLY:
			if err := vm.binaryOp(op.Mul); err != nil {
				return err
			}

		case OP_DIVIDE:
			if err := vm.binaryOp(op.Div); err != nil {
				return err
			}

		case OP_GREATER:
			if err := vm.binaryOp(op.Greater); err != nil {
				return err
			}

		case OP_LESS:
			if err := vm.binaryOp(op.Less); err != nil {
				return err
			}

		case OP_NOT:
			vm.push(ValueBool(IsFalsy(vm.pop())))

		case OP_NEGATE:
			if _, isNum := vm.peek(0).(ValueNumber); !isNum {
				return vm.RuntimeError("Operand must be a number.")
			}

			val := vm.pop().(ValueNumber)
			vm.push(ValueNumber(-val))

		case OP_TRUE:
			vm.push(ValueBool(true))
		case OP_FALSE:
			vm.push(ValueBool(false))
		case OP_NIL:
			vm.push(ValueNil(0))

		case OP_EQUAL:
			b := vm.pop()
			a := vm.pop()
			vm.push(ValueBool(a.Equals(b)))

		case OP_RETURN:
			PrintValue(vm.pop())
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

func (vm *VM) RuntimeError(format string, args ...any) error {
	fmt.Fprintf(os.Stderr, format, args...)
	fmt.Fprintf(os.Stderr, "\n")

	line := vm.chunk.GetLine(vm.ip)
	fmt.Fprintf(os.Stderr, "[line %d] in script\n", line)

	return InterpretRuntimeError
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

func (vm *VM) peek(dist int) Value {
	return vm.stack[vm.sp-dist]
}

func (vm *VM) binaryOp(oper op.BinaryOp) error {
	bval := vm.pop()
	aval := vm.pop()

	a, aIsNum := aval.(ValueNumber)
	b, bIsNum := bval.(ValueNumber)

	if !aIsNum || !bIsNum {
		return vm.RuntimeError("Operand must be a number.")
	}

	var res Value
	switch oper {
	case op.Plus:
		res = ValueNumber(a + b)
	case op.Minus:
		res = ValueNumber(a - b)
	case op.Mul:
		res = ValueNumber(a * b)
	case op.Div:
		res = ValueNumber(a / b)
	case op.Greater:
		res = ValueBool(a > b)
	case op.Less:
		res = ValueBool(a < b)
	}

	vm.push(res)
	return nil
}
