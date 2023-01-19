package lox

type OpCode byte

type Chunk struct {
	code      []byte
	constants *ValueArray
	lines     lines
}

const (
	OP_CONSTANT OpCode = iota
	OP_NIL
	OP_TRUE
	OP_FALSE
	OP_ADD
	OP_SUBTRACT
	OP_MULTIPLY
	OP_DIVIDE
	OP_NEGATE
	OP_RETURN
)

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
