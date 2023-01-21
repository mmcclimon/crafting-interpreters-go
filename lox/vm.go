package lox

import (
	"errors"
	"fmt"
	"os"

	"github.com/mmcclimon/glox/lox/op"
)

const FRAMES_MAX = 64
const STACK_MAX = FRAMES_MAX * UINT8_COUNT

var InterpretCompileError = errors.New("compile error")
var InterpretRuntimeError = errors.New("runtime error")

type CallFrame struct {
	function *ValueFunction
	ip       int
	slots    []Value
}

type VM struct {
	frames     [FRAMES_MAX]CallFrame
	frameCount int
	stack      [STACK_MAX]Value
	sp         int
	globals    map[string]Value
}

func NewVM() *VM {
	return &VM{
		globals: make(map[string]Value),
	}
}

func (vm *VM) InterpretString(source string) error {
	compiler := NewCompiler(source, TYPE_SCRIPT)
	function, err := compiler.Compile()

	if err != nil {
		return InterpretCompileError
	}

	vm.push(function)
	vm.frameCount++
	frame := vm.currentFrame()
	frame.function = function
	frame.ip = 0
	frame.slots = vm.stack[0:]

	return vm.run()
}

/*
func (vm *VM) Interpret(chunk *Chunk) error {
	vm.chunk = chunk
	vm.ip = 0
	return vm.run()
}
*/

func (vm *VM) run() error {
	frame := vm.currentFrame()

	for {
		if DEBUG_TRACE_EXECUTION {
			frame.function.chunk.DisassembleInstruction(frame.ip)
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
			_, aIsStr := vm.peek(0).(ValueString)
			_, bIsStr := vm.peek(1).(ValueString)

			_, aIsNum := vm.peek(0).(ValueNumber)
			_, bIsNum := vm.peek(1).(ValueNumber)

			if aIsStr && bIsStr {
				vm.concatenate()
			} else if aIsNum && bIsNum {
				if err := vm.binaryOp(op.Plus); err != nil {
					return err
				}
			} else {
				return vm.RuntimeError("Operands must be numbers or strings.")
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
		case OP_POP:
			vm.pop()

		case OP_GET_LOCAL:
			slot := vm.readByte()
			vm.push(frame.slots[slot])

		case OP_SET_LOCAL:
			slot := vm.readByte()
			frame.slots[slot] = vm.peek(0)

		case OP_DEFINE_GLOBAL:
			name := vm.readConstant().(ValueString)
			vm.globals[string(name)] = vm.peek(0)
			vm.pop()

		case OP_GET_GLOBAL:
			name := vm.readConstant().(ValueString)
			value, ok := vm.globals[string(name)]
			if !ok {
				return vm.RuntimeError("Undefined variable '%s'.", name)
			}
			vm.push(value)

		case OP_SET_GLOBAL:
			name := string(vm.readConstant().(ValueString))

			if _, exists := vm.globals[name]; !exists {
				return vm.RuntimeError("Undefined variable '%s'.", name)
			}

			vm.globals[name] = vm.peek(0)

		case OP_EQUAL:
			b := vm.pop()
			a := vm.pop()
			vm.push(ValueBool(a.Equals(b)))

		case OP_PRINT:
			PrintValue(vm.pop())
			fmt.Printf("\n")

		case OP_JUMP:
			offset := vm.readShort()
			frame.ip += offset

		case OP_JUMP_IF_FALSE:
			offset := vm.readShort()
			if IsFalsy(vm.peek(0)) {
				frame.ip += offset
			}

		case OP_JUMP_IF_TRUE:
			offset := vm.readShort()
			if !IsFalsy(vm.peek(0)) {
				frame.ip += offset
			}

		case OP_LOOP:
			offset := vm.readShort()
			frame.ip -= offset

		case OP_RETURN:
			return nil
		}
	}
}

func (vm *VM) currentFrame() *CallFrame {
	return &vm.frames[vm.frameCount-1]
}

func (vm *VM) readByte() byte {
	frame := vm.currentFrame()
	frame.ip++
	return frame.function.chunk.code[frame.ip-1]
}

func (vm *VM) readConstant() Value {
	frame := vm.currentFrame()
	return frame.function.chunk.constantAt(vm.readByte())
}

func (vm *VM) readShort() int {
	frame := vm.currentFrame()
	frame.ip += 2
	code := frame.function.chunk.code
	return int(code[frame.ip-2]<<8 | code[frame.ip-1])
}

func (vm *VM) RuntimeError(format string, args ...any) error {
	fmt.Fprintf(os.Stderr, format, args...)
	fmt.Fprintf(os.Stderr, "\n")

	frame := vm.currentFrame()
	line := frame.function.chunk.GetLine(frame.ip)
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
	return vm.stack[vm.sp-1-dist]
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

func (vm *VM) concatenate() {
	b := vm.pop().(ValueString)
	a := vm.pop().(ValueString)
	vm.push(ValueString(a + b))
}
