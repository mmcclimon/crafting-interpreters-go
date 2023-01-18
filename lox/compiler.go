package lox

import (
	"fmt"
	"math"
	"os"
	"runtime/debug"
	"strconv"
)

// types
type Compiler struct {
	chunk *Chunk
}

type Parser struct {
	current   Token
	previous  Token
	hadError  bool
	panicMode bool
}

type ParseRule struct {
	prefix     ParseFn
	infix      ParseFn
	precedence Precedence
}

type Precedence int
type ParseFn func()

// Variables
var scanner *Scanner
var parser *Parser
var compiler *Compiler
var compilingChunk *Chunk
var rules map[TokenType]ParseRule

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

func init() {
	rules = map[TokenType]ParseRule{
		TOKEN_LEFT_PAREN:    {grouping, nil, PREC_NONE},
		TOKEN_RIGHT_PAREN:   {nil, nil, PREC_NONE},
		TOKEN_LEFT_BRACE:    {nil, nil, PREC_NONE},
		TOKEN_RIGHT_BRACE:   {nil, nil, PREC_NONE},
		TOKEN_COMMA:         {nil, nil, PREC_NONE},
		TOKEN_DOT:           {nil, nil, PREC_NONE},
		TOKEN_MINUS:         {unary, binary, PREC_TERM},
		TOKEN_PLUS:          {nil, binary, PREC_TERM},
		TOKEN_SEMICOLON:     {nil, nil, PREC_NONE},
		TOKEN_SLASH:         {nil, binary, PREC_FACTOR},
		TOKEN_STAR:          {nil, binary, PREC_FACTOR},
		TOKEN_BANG:          {nil, nil, PREC_NONE},
		TOKEN_BANG_EQUAL:    {nil, nil, PREC_NONE},
		TOKEN_EQUAL:         {nil, nil, PREC_NONE},
		TOKEN_EQUAL_EQUAL:   {nil, nil, PREC_NONE},
		TOKEN_GREATER:       {nil, nil, PREC_NONE},
		TOKEN_GREATER_EQUAL: {nil, nil, PREC_NONE},
		TOKEN_LESS:          {nil, nil, PREC_NONE},
		TOKEN_LESS_EQUAL:    {nil, nil, PREC_NONE},
		TOKEN_IDENTIFIER:    {nil, nil, PREC_NONE},
		TOKEN_STRING:        {nil, nil, PREC_NONE},
		TOKEN_NUMBER:        {number, nil, PREC_NONE},
		TOKEN_AND:           {nil, nil, PREC_NONE},
		TOKEN_CLASS:         {nil, nil, PREC_NONE},
		TOKEN_ELSE:          {nil, nil, PREC_NONE},
		TOKEN_FALSE:         {nil, nil, PREC_NONE},
		TOKEN_FOR:           {nil, nil, PREC_NONE},
		TOKEN_FUN:           {nil, nil, PREC_NONE},
		TOKEN_IF:            {nil, nil, PREC_NONE},
		TOKEN_NIL:           {nil, nil, PREC_NONE},
		TOKEN_OR:            {nil, nil, PREC_NONE},
		TOKEN_PRINT:         {nil, nil, PREC_NONE},
		TOKEN_RETURN:        {nil, nil, PREC_NONE},
		TOKEN_SUPER:         {nil, nil, PREC_NONE},
		TOKEN_THIS:          {nil, nil, PREC_NONE},
		TOKEN_TRUE:          {nil, nil, PREC_NONE},
		TOKEN_VAR:           {nil, nil, PREC_NONE},
		TOKEN_WHILE:         {nil, nil, PREC_NONE},
		TOKEN_ERROR:         {nil, nil, PREC_NONE},
		TOKEN_EOF:           {nil, nil, PREC_NONE},
	}
}

func Compile(source string, chunk *Chunk) bool {
	compilingChunk = chunk

	scanner = NewScanner(source)
	parser = &Parser{
		hadError:  false,
		panicMode: false,
	}

	advance()

	expression()
	consume(TOKEN_EOF, "Expect end of expression")

	endCompiler()

	return !parser.hadError
}

func advance() {
	parser.previous = parser.current

	for {
		parser.current = scanner.ScanToken()

		if parser.current.kind != TOKEN_ERROR {
			break
		}

		parser.errorAtCurrent(parser.current.lexeme)
	}
}

func consume(tt TokenType, msg string) {
	if parser.current.kind == tt {
		advance()
		return
	}

	parser.errorAtCurrent(msg)
}

func expression() {
	parsePrecedence(PREC_ASSIGNMENT)
}

func number() {
	value, err := strconv.ParseFloat(parser.previous.lexeme, 64)
	if err != nil {
		panic("strconv.ParseFloat failed somehow")
	}

	emitConstant(Value(value))
}

func grouping() {
	expression()
	consume(TOKEN_RIGHT_PAREN, "Expect ')' after expression.")
}

func unary() {
	op := parser.previous.kind
	parsePrecedence(PREC_UNARY)

	switch op {
	case TOKEN_MINUS:
		emitByte(byte(OP_NEGATE))
	default:
		panic("unreachable")
	}
}

func binary() {
	op := parser.previous.kind
	rule := getRule(op)
	parsePrecedence(rule.precedence + 1)

	switch op {
	case TOKEN_PLUS:
		emitByte(byte(OP_ADD))
	case TOKEN_MINUS:
		emitByte(byte(OP_SUBTRACT))
	case TOKEN_STAR:
		emitByte(byte(OP_MULTIPLY))
	case TOKEN_SLASH:
		emitByte(byte(OP_DIVIDE))
	default:
		panic("unreachable")
	}
}

func getRule(op TokenType) ParseRule {
	return rules[op]
}

func parsePrecedence(precedence Precedence) {
	advance()
	prefixRule := getRule(parser.previous.kind).prefix
	if prefixRule == nil {
		parser.error("Expect expression.")
		return
	}

	prefixRule()

	for precedence <= getRule(parser.current.kind).precedence {
		advance()
		infixRule := getRule(parser.previous.kind).infix
		infixRule()
	}
}

func emitByte(item byte) {
	currentChunk().Write(item, parser.previous.line)
}

func emitBytes(item1 byte, item2 byte) {
	emitByte(item1)
	emitByte(item2)
}

func emitConstant(value Value) {
	emitBytes(byte(OP_CONSTANT), makeConstant(value))
}

func makeConstant(value Value) byte {
	constant := currentChunk().AddConstant(value)
	if constant > math.MaxUint8 {
		parser.error("Too many constants in one chunk")
		return 0
	}

	return byte(constant)
}

func currentChunk() *Chunk {
	return compilingChunk
}

func endCompiler() {
	emitReturn()

	if DEBUG_PRINT_CODE && !parser.hadError {
		currentChunk().Disassemble("code")
	}
}

func emitReturn() {
	emitByte(byte(OP_RETURN))
}

// parser stuff
func (p *Parser) error(message string) {
	p.errorAt(parser.previous, message)
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

	debug.PrintStack()
}
