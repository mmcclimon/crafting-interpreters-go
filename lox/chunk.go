package lox

type OpCode byte

type Chunk struct {
	code      []byte
	constants *ValueArray
	lines     lines
}

const (
	OP_UNKNOWN OpCode = iota
	OP_CONSTANT
	OP_NIL
	OP_TRUE
	OP_FALSE
	OP_POP
	OP_DEFINE_GLOBAL
	OP_GET_GLOBAL
	OP_SET_GLOBAL
	OP_GET_LOCAL
	OP_SET_LOCAL
	OP_EQUAL
	OP_GREATER
	OP_LESS
	OP_ADD
	OP_SUBTRACT
	OP_MULTIPLY
	OP_DIVIDE
	OP_NOT
	OP_NEGATE
	OP_PRINT
	OP_JUMP
	OP_JUMP_IF_FALSE
	OP_JUMP_IF_TRUE
	OP_LOOP
	OP_CALL
	OP_RETURN
)

var opNames map[OpCode]string

func NewChunk() *Chunk {
	code := make([]byte, 0, 8)
	return &Chunk{code: code, constants: NewValueArray()}
}

func (c *Chunk) Write(item byte, line int) {
	c.lines.add(line, len(c.code))
	c.code = append(c.code, item)
}

func (c *Chunk) AddConstant(item Value) int {
	c.constants.Write(item)
	return len(*c.constants) - 1
}

func (c *Chunk) constantAt(offset byte) Value {
	return (*c.constants)[offset]
}

func (c *Chunk) GetLine(offset int) int {
	return c.lines.getForOffset(offset)
}

func (c *Chunk) Count() int {
	return len(c.code)
}

// lines is a run-length encoded array like [line-num, max-offset, ...]
type lines []int

// lines.add adds a new line, incrementing the count or appending as necessary
func (l *lines) add(line, offset int) {
	lineLen := len(*l)

	// If we're empty, or this is a new line, add it onto the end with count 1
	if lineLen == 0 || line != (*l)[lineLen-2] {
		*l = append(*l, line, offset)
		return
	}

	// otherwise, store this offset for this line number
	(*l)[lineLen-1] = offset
}

// lines.getForOffset walks the array, returning the line number for a given
// instruction offset
func (l *lines) getForOffset(offset int) int {
	for idx := 0; true; idx += 2 {
		if (*l)[idx+1] >= offset {
			return (*l)[idx]
		}
	}

	panic("internal error in line accounting")
}

// OpCode stuff

func init() {
	opNames = map[OpCode]string{
		OP_CONSTANT:      "OP_CONSTANT",
		OP_NIL:           "OP_NIL",
		OP_TRUE:          "OP_TRUE",
		OP_FALSE:         "OP_FALSE",
		OP_POP:           "OP_POP",
		OP_DEFINE_GLOBAL: "OP_DEFINE_GLOBAL",
		OP_GET_GLOBAL:    "OP_GET_GLOBAL",
		OP_SET_GLOBAL:    "OP_SET_GLOBAL",
		OP_GET_LOCAL:     "OP_GET_LOCAL",
		OP_SET_LOCAL:     "OP_SET_LOCAL",
		OP_EQUAL:         "OP_EQUAL",
		OP_GREATER:       "OP_GREATER",
		OP_LESS:          "OP_LESS",
		OP_ADD:           "OP_ADD",
		OP_SUBTRACT:      "OP_SUBTRACT",
		OP_MULTIPLY:      "OP_MULTIPLY",
		OP_DIVIDE:        "OP_DIVIDE",
		OP_NOT:           "OP_NOT",
		OP_NEGATE:        "OP_NEGATE",
		OP_PRINT:         "OP_PRINT",
		OP_JUMP:          "OP_JUMP",
		OP_JUMP_IF_FALSE: "OP_JUMP_IF_FALSE",
		OP_JUMP_IF_TRUE:  "OP_JUMP_IF_TRUE",
		OP_LOOP:          "OP_LOOP",
		OP_CALL:          "OP_CALL",
		OP_RETURN:        "OP_RETURN",
	}
}

func (op OpCode) String() string {
	return opNames[op]
}
