package lox

import (
	"errors"
	"fmt"
	"math"
	"os"
	"strconv"
)

// types
const UINT8_COUNT = math.MaxUint8 + 1

type Compiler struct {
	chunk *Chunk
	rules map[TokenType]ParseRule
	Parser
	localCount int
	scopeDepth int
	locals     [UINT8_COUNT]Local
}

type Local struct {
	name  Token
	depth int
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
type parseFn func(bool)
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

	for !c.match(TOKEN_EOF) {
		c.declaration()
	}

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
		TOKEN_IDENTIFIER:    {c.variable, nil, PREC_NONE},
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

func (c *Compiler) declaration() {
	if c.match(TOKEN_VAR) {
		c.varDeclaration()
	} else {
		c.statement()
	}

	if c.panicMode {
		c.synchronize()
	}
}

func (c *Compiler) varDeclaration() {
	global := c.parseVariable("Expect variable name.")

	if c.match(TOKEN_EQUAL) {
		c.expression()
	} else {
		c.emitOp(OP_NIL)
	}

	c.consume(TOKEN_SEMICOLON, "Expect ';' after variable declaration")

	c.defineVariable(global)
}

func (c *Compiler) statement() {
	if c.match(TOKEN_PRINT) {
		c.printStatement()
	} else if c.match(TOKEN_LEFT_BRACE) {
		c.beginScope()
		c.block()
		c.endScope()
	} else {
		c.expressionStatement()
	}
}

func (c *Compiler) printStatement() {
	c.expression()
	c.consume(TOKEN_SEMICOLON, "Expect ';' after value")
	c.emitOp(OP_PRINT)
}

func (c *Compiler) expressionStatement() {
	c.expression()
	c.consume(TOKEN_SEMICOLON, "Expect ';' after expression")
	c.emitOp(OP_POP)
}

func (c *Compiler) expression() {
	c.parsePrecedence(PREC_ASSIGNMENT)
}

func (c *Compiler) block() {
	for !c.check(TOKEN_RIGHT_BRACE) && !c.check(TOKEN_EOF) {
		c.declaration()
	}

	c.consume(TOKEN_RIGHT_BRACE, "Expect '}' after block.")
}

func (c *Compiler) number(_ bool) {
	n, err := strconv.ParseFloat(c.previous.lexeme, 64)
	if err != nil {
		panic("strconv.ParseFloat failed somehow")
	}

	c.emitConstant(ValueNumber(n))
}

func (c *Compiler) string(_ bool) {
	s := c.previous.lexeme
	c.emitConstant(ValueString(s[1 : len(s)-1]))
}

func (c *Compiler) variable(canAssign bool) {
	c.namedVariable(c.previous, canAssign)
}

func (c *Compiler) namedVariable(name Token, canAssign bool) {
	var getOp, setOp OpCode
	arg, err := c.resolveLocal(name)

	if err == nil {
		getOp = OP_GET_LOCAL
		setOp = OP_SET_LOCAL
	} else {
		arg = c.identifierConstant(name)
		getOp = OP_GET_GLOBAL
		setOp = OP_SET_GLOBAL
	}

	if canAssign && c.match(TOKEN_EQUAL) {
		c.expression()
		c.emitOpAndArg(setOp, arg)
	} else {
		c.emitOpAndArg(getOp, arg)
	}
}

func (c *Compiler) resolveLocal(name Token) (byte, error) {
	for i := c.localCount - 1; i >= 0; i-- {
		local := c.locals[i]
		if identifiersEqual(name, local.name) {
			if local.depth == -1 {
				c.error("Can't read local variable in its own initializer")
			}

			return byte(i), nil
		}
	}

	return 0, errors.New("local var not found")
}

func (c *Compiler) grouping(_ bool) {
	c.expression()
	c.consume(TOKEN_RIGHT_PAREN, "Expect ')' after expression.")
}

func (c *Compiler) unary(_ bool) {
	op := c.previous.kind
	c.parsePrecedence(PREC_UNARY)

	switch op {
	case TOKEN_BANG:
		c.emitOp(OP_NOT)
	case TOKEN_MINUS:
		c.emitOp(OP_NEGATE)
	default:
		panic("unreachable")
	}
}

func (c *Compiler) binary(_ bool) {
	op := c.previous.kind
	rule := c.getRule(op)
	c.parsePrecedence(rule.precedence + 1)

	switch op {
	case TOKEN_BANG_EQUAL:
		c.emitOp(OP_EQUAL)
		c.emitOp(OP_NOT)
	case TOKEN_EQUAL_EQUAL:
		c.emitOp(OP_EQUAL)
	case TOKEN_GREATER:
		c.emitOp(OP_GREATER)
	case TOKEN_GREATER_EQUAL:
		c.emitOp(OP_LESS)
		c.emitOp(OP_NOT)
	case TOKEN_LESS:
		c.emitOp(OP_LESS)
	case TOKEN_LESS_EQUAL:
		c.emitOp(OP_GREATER)
		c.emitOp(OP_NOT)
	case TOKEN_PLUS:
		c.emitOp(OP_ADD)
	case TOKEN_MINUS:
		c.emitOp(OP_SUBTRACT)
	case TOKEN_STAR:
		c.emitOp(OP_MULTIPLY)
	case TOKEN_SLASH:
		c.emitOp(OP_DIVIDE)
	default:
		panic("unreachable")
	}
}

func (c *Compiler) literal(_ bool) {
	switch c.previous.kind {
	case TOKEN_FALSE:
		c.emitOp(OP_FALSE)
	case TOKEN_NIL:
		c.emitOp(OP_NIL)
	case TOKEN_TRUE:
		c.emitOp(OP_TRUE)
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

	canAssign := precedence <= PREC_ASSIGNMENT
	prefixRule(canAssign)

	for precedence <= c.getRule(c.current.kind).precedence {
		c.advance()
		infixRule := c.getRule(c.previous.kind).infix
		infixRule(canAssign)
	}
}

func (c *Compiler) parseVariable(errMsg string) byte {
	c.consume(TOKEN_IDENTIFIER, errMsg)

	c.declareVariable()
	if c.scopeDepth > 0 {
		return 0
	}

	return c.identifierConstant(c.previous)
}

func (c *Compiler) markInitialized() {
	c.locals[c.localCount-1].depth = c.scopeDepth
}

func (c *Compiler) defineVariable(global byte) {
	if c.scopeDepth > 0 {
		c.markInitialized()
		return
	}

	c.emitOpAndArg(OP_DEFINE_GLOBAL, global)
}

func (c *Compiler) identifierConstant(name Token) byte {
	return c.makeConstant(ValueString(name.lexeme))
}

func identifiersEqual(a, b Token) bool {
	return a.lexeme == b.lexeme
}

func (c *Compiler) declareVariable() {
	if c.scopeDepth == 0 {
		return
	}

	name := c.previous

	for i := c.localCount - 1; i >= 0; i-- {
		local := c.locals[i]
		if local.depth != -1 && local.depth < c.scopeDepth {
			break
		}

		if identifiersEqual(name, local.name) {
			c.error("Already a variable with this name in this scope")
		}
	}

	c.addLocal(name)
}

func (c *Compiler) addLocal(name Token) {
	if c.localCount == UINT8_COUNT {
		c.error("Too many local variables in scope")
		return
	}

	c.locals[c.localCount] = Local{name, -1}
	c.localCount++
}

// these functions all write to our chunk
func (c *Compiler) emitOp(op OpCode) {
	c.emitByte(byte(op))
}

func (c *Compiler) emitOpAndArg(op OpCode, arg byte) {
	c.emitBytes(byte(op), arg)
}

func (c *Compiler) emitByte(item byte) {
	c.chunk.Write(item, c.previous.line)
}

func (c *Compiler) emitBytes(item1 byte, item2 byte) {
	c.emitByte(item1)
	c.emitByte(item2)
}

func (c *Compiler) emitConstant(value Value) {
	c.emitOpAndArg(OP_CONSTANT, c.makeConstant(value))
}

func (c *Compiler) emitReturn() {
	c.emitOp(OP_RETURN)
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

func (c *Compiler) beginScope() {
	c.scopeDepth++
}

func (c *Compiler) endScope() {
	c.scopeDepth--

	for c.localCount > 0 && c.locals[c.localCount-1].depth > c.scopeDepth {
		c.emitOp(OP_POP)
		c.localCount--
	}

}

/*
 * Parsing functions
 */
func (p *Parser) match(tt TokenType) bool {
	if !p.check(tt) {
		return false
	}

	p.advance()
	return true
}

func (p *Parser) check(tt TokenType) bool {
	return p.current.kind == tt
}

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

func (p *Parser) synchronize() {
	p.panicMode = false

	for p.current.kind != TOKEN_EOF {
		if p.previous.kind == TOKEN_SEMICOLON {
			return
		}

		switch p.current.kind {
		case TOKEN_CLASS,
			TOKEN_FUN,
			TOKEN_VAR,
			TOKEN_FOR,
			TOKEN_IF,
			TOKEN_WHILE,
			TOKEN_PRINT,
			TOKEN_RETURN:
			return

		default:
			// Do nothing.
		}
	}
}
