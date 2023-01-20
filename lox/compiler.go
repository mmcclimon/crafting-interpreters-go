package lox

import (
	"fmt"
	"math"
	"os"
	"strconv"
)

// types
type Compiler struct {
	chunk *Chunk
	rules map[TokenType]ParseRule
	Parser
}

type Parser struct {
	scanner   *Scanner
	current   Token
	previous  Token
	hadError  bool
	panicMode bool
}

// This is nullary because it gets bound to the compiler instance on create,
// so that we can call things as methods
type parseFn func()
type Precedence int

type ParseRule struct {
	prefix     parseFn
	infix      parseFn
	precedence Precedence
}

const (
	PREC_NONE       Precedence = iota
	PREC_ASSIGNMENT            // =
	PREC_OR                    // or
	PREC_AND                   // and
	PREC_EQUALITY              // == !=
	PREC_COMPARISON            // < > <= >=
	PREC_TERM                  // + -
	PREC_FACTOR                // * /
	PREC_UNARY                 // ! -
	PREC_CALL                  // . ()
	PREC_PRIMARY
)

func NewCompiler(source string) *Compiler {
	c := &Compiler{
		chunk: NewChunk(),
		Parser: Parser{
			scanner:   NewScanner(source),
			hadError:  false,
			panicMode: false,
		},
	}

	c.initRules()

	return c
}

func (c *Compiler) Compile() bool {
	c.advance()
	c.expression()
	c.consume(TOKEN_EOF, "Expect end of expression")
	c.end()

	return !c.hadError
}

func (c *Compiler) initRules() {
	c.rules = map[TokenType]ParseRule{
		TOKEN_LEFT_PAREN:    {c.grouping, nil, PREC_NONE},
		TOKEN_RIGHT_PAREN:   {nil, nil, PREC_NONE},
		TOKEN_LEFT_BRACE:    {nil, nil, PREC_NONE},
		TOKEN_RIGHT_BRACE:   {nil, nil, PREC_NONE},
		TOKEN_COMMA:         {nil, nil, PREC_NONE},
		TOKEN_DOT:           {nil, nil, PREC_NONE},
		TOKEN_MINUS:         {c.unary, c.binary, PREC_TERM},
		TOKEN_PLUS:          {nil, c.binary, PREC_TERM},
		TOKEN_SEMICOLON:     {nil, nil, PREC_NONE},
		TOKEN_SLASH:         {nil, c.binary, PREC_FACTOR},
		TOKEN_STAR:          {nil, c.binary, PREC_FACTOR},
		TOKEN_BANG:          {c.unary, nil, PREC_NONE},
		TOKEN_BANG_EQUAL:    {nil, c.binary, PREC_EQUALITY},
		TOKEN_EQUAL:         {nil, nil, PREC_NONE},
		TOKEN_EQUAL_EQUAL:   {nil, c.binary, PREC_EQUALITY},
		TOKEN_GREATER:       {nil, c.binary, PREC_COMPARISON},
		TOKEN_GREATER_EQUAL: {nil, c.binary, PREC_COMPARISON},
		TOKEN_LESS:          {nil, c.binary, PREC_COMPARISON},
		TOKEN_LESS_EQUAL:    {nil, c.binary, PREC_COMPARISON},
		TOKEN_IDENTIFIER:    {nil, nil, PREC_NONE},
		TOKEN_STRING:        {c.string, nil, PREC_NONE},
		TOKEN_NUMBER:        {c.number, nil, PREC_NONE},
		TOKEN_AND:           {nil, nil, PREC_NONE},
		TOKEN_CLASS:         {nil, nil, PREC_NONE},
		TOKEN_ELSE:          {nil, nil, PREC_NONE},
		TOKEN_FALSE:         {c.literal, nil, PREC_NONE},
		TOKEN_FOR:           {nil, nil, PREC_NONE},
		TOKEN_FUN:           {nil, nil, PREC_NONE},
		TOKEN_IF:            {nil, nil, PREC_NONE},
		TOKEN_NIL:           {c.literal, nil, PREC_NONE},
		TOKEN_OR:            {nil, nil, PREC_NONE},
		TOKEN_PRINT:         {nil, nil, PREC_NONE},
		TOKEN_RETURN:        {nil, nil, PREC_NONE},
		TOKEN_SUPER:         {nil, nil, PREC_NONE},
		TOKEN_THIS:          {nil, nil, PREC_NONE},
		TOKEN_TRUE:          {c.literal, nil, PREC_NONE},
		TOKEN_VAR:           {nil, nil, PREC_NONE},
		TOKEN_WHILE:         {nil, nil, PREC_NONE},
		TOKEN_ERROR:         {nil, nil, PREC_NONE},
		TOKEN_EOF:           {nil, nil, PREC_NONE},
	}
}

func (c *Compiler) expression() {
	c.parsePrecedence(PREC_ASSIGNMENT)
}

func (c *Compiler) number() {
	n, err := strconv.ParseFloat(c.previous.lexeme, 64)
	if err != nil {
		panic("strconv.ParseFloat failed somehow")
	}

	c.emitConstant(ValueNumber(n))
}

func (c *Compiler) string() {
	s := c.previous.lexeme
	c.emitConstant(ValueString(s[1 : len(s)-1]))
}

func (c *Compiler) grouping() {
	c.expression()
	c.consume(TOKEN_RIGHT_PAREN, "Expect ')' after expression.")
}

func (c *Compiler) unary() {
	op := c.previous.kind
	c.parsePrecedence(PREC_UNARY)

	switch op {
	case TOKEN_BANG:
		c.emitByte(byte(OP_NOT))
	case TOKEN_MINUS:
		c.emitByte(byte(OP_NEGATE))
	default:
		panic("unreachable")
	}
}

func (c *Compiler) binary() {
	op := c.previous.kind
	rule := c.getRule(op)
	c.parsePrecedence(rule.precedence + 1)

	switch op {
	case TOKEN_BANG_EQUAL:
		c.emitBytes(byte(OP_EQUAL), byte(OP_NOT))
	case TOKEN_EQUAL_EQUAL:
		c.emitByte(byte(OP_EQUAL))
	case TOKEN_GREATER:
		c.emitByte(byte(OP_GREATER))
	case TOKEN_GREATER_EQUAL:
		c.emitBytes(byte(OP_LESS), byte(OP_NOT))
	case TOKEN_LESS:
		c.emitByte(byte(OP_LESS))
	case TOKEN_LESS_EQUAL:
		c.emitBytes(byte(OP_GREATER), byte(OP_NOT))
	case TOKEN_PLUS:
		c.emitByte(byte(OP_ADD))
	case TOKEN_MINUS:
		c.emitByte(byte(OP_SUBTRACT))
	case TOKEN_STAR:
		c.emitByte(byte(OP_MULTIPLY))
	case TOKEN_SLASH:
		c.emitByte(byte(OP_DIVIDE))
	default:
		panic("unreachable")
	}
}

func (c *Compiler) literal() {
	switch c.previous.kind {
	case TOKEN_FALSE:
		c.emitByte(byte(OP_FALSE))
	case TOKEN_NIL:
		c.emitByte(byte(OP_NIL))
	case TOKEN_TRUE:
		c.emitByte(byte(OP_TRUE))
	default:
		panic("unreachable")
	}
}

func (c *Compiler) getRule(op TokenType) ParseRule {
	return c.rules[op]
}

func (c *Compiler) parsePrecedence(precedence Precedence) {
	c.advance()
	prefixRule := c.getRule(c.previous.kind).prefix
	if prefixRule == nil {
		c.error("Expect expression.")
		return
	}

	prefixRule()

	for precedence <= c.getRule(c.current.kind).precedence {
		c.advance()
		infixRule := c.getRule(c.previous.kind).infix
		infixRule()
	}
}

// these functions all write to our chunk

func (c *Compiler) emitByte(item byte) {
	c.chunk.Write(item, c.previous.line)
}

func (c *Compiler) emitBytes(item1 byte, item2 byte) {
	c.emitByte(item1)
	c.emitByte(item2)
}

func (c *Compiler) emitConstant(value Value) {
	c.emitBytes(byte(OP_CONSTANT), c.makeConstant(value))
}

func (c *Compiler) emitReturn() {
	c.emitByte(byte(OP_RETURN))
}

func (c *Compiler) makeConstant(value Value) byte {
	constant := c.chunk.AddConstant(value)
	if constant > math.MaxUint8 {
		c.error("Too many constants in one chunk")
		return 0
	}

	return byte(constant)
}

func (c *Compiler) end() {
	c.emitReturn()

	if DEBUG_PRINT_CODE && !c.hadError {
		c.chunk.Disassemble("code")
	}
}

/*
 * Parsing functions
 */
func (p *Parser) advance() {
	p.previous = p.current

	for {
		p.current = p.scanner.ScanToken()

		if p.current.kind != TOKEN_ERROR {
			break
		}

		p.errorAtCurrent(p.current.lexeme)
	}
}

func (p *Parser) consume(tt TokenType, msg string) {
	if p.current.kind == tt {
		p.advance()
		return
	}

	p.errorAtCurrent(msg)
}

func (p *Parser) error(message string) {
	p.errorAt(p.previous, message)
}

func (p *Parser) errorAtCurrent(message string) {
	p.errorAt(p.current, message)
}

func (p *Parser) errorAt(tok Token, message string) {
	if p.panicMode {
		return
	}

	p.panicMode = true

	fmt.Fprintf(os.Stderr, "[line %d] Error", tok.line)

	switch tok.kind {
	case TOKEN_EOF:
		fmt.Fprintf(os.Stderr, " at end")
	case TOKEN_ERROR:
		// do nothing
	default:
		fmt.Fprintf(os.Stderr, " at '%s'", tok.lexeme)
	}

	fmt.Fprintf(os.Stderr, ": %s\n", message)
	p.hadError = true

	// debug.PrintStack()
}
