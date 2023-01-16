package lox

type OpCode byte
type Chunk struct {
	code      []byte
	constants *ValueArray
	lines     []int
}

const (
	OP_CONSTANT OpCode = iota
	OP_RETURN
)

func NewChunk() *Chunk {
	code := make([]byte, 0, 8)
	return &Chunk{code: code, constants: NewValueArray()}
}

func (c *Chunk) Write(item byte, line int) {
	c.code = append(c.code, item)
	c.lines = append(c.lines, line)
}

func (c *Chunk) AddConstant(item Value) int {
	c.constants.Write(item)
	return len(*c.constants) - 1
}

func (c *Chunk) constantAt(offset byte) Value {
	return (*c.constants)[offset]
}
