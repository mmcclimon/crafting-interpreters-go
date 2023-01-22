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
	enclosing  *Compiler
	function   *ValueFunction
	kind       FunctionType
	rules      map[TokenType]ParseRule
	localCount int
	scopeDepth int
	locals     [UINT8_COUNT]Local
}

type Local struct {
	name  Token
	depth int
}

// this is a global, basically, set up when we call Compile()
var parser Parser

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
type FunctionType int

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

const (
	TYPE_FUNCTION FunctionType = iota
	TYPE_SCRIPT
)

func NewCompiler(kind FunctionType, parent *Compiler) *Compiler {
	c := &Compiler{
		enclosing: parent,
		function:  NewFunction(),
		kind:      kind,
	}

	c.initRules()

	if kind != TYPE_SCRIPT {
		c.function.name = parser.previous.lexeme
	}

	local := c.locals[c.localCount]
	local.name.lexeme = ""
	c.localCount++

	return c
}

func Compile(source string) (*ValueFunction, error) {
	parser = Parser{
		scanner:   NewScanner(source),
		hadError:  false,
		panicMode: false,
	}

	c := NewCompiler(TYPE_SCRIPT, nil)

	parser.advance()

	for !parser.match(TOKEN_EOF) {
		c.declaration()
	}

	function := c.end()

	if parser.hadError {
		return nil, errors.New("compilation error")
	}

	return function, nil
}

func (c *Compiler) initRules() {
	c.rules = map[TokenType]ParseRule{
		TOKEN_LEFT_PAREN:    {c.grouping, c.call, PREC_CALL},
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
		TOKEN_AND:           {nil, c.and, PREC_AND},
		TOKEN_CLASS:         {nil, nil, PREC_NONE},
		TOKEN_ELSE:          {nil, nil, PREC_NONE},
		TOKEN_FALSE:         {c.literal, nil, PREC_NONE},
		TOKEN_FOR:           {nil, nil, PREC_NONE},
		TOKEN_FUN:           {nil, nil, PREC_NONE},
		TOKEN_IF:            {nil, nil, PREC_NONE},
		TOKEN_NIL:           {c.literal, nil, PREC_NONE},
		TOKEN_OR:            {nil, c.or, PREC_OR},
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
	if parser.match(TOKEN_FUN) {
		c.funDeclaration()
	} else if parser.match(TOKEN_VAR) {
		c.varDeclaration()
	} else {
		c.statement()
	}

	if parser.panicMode {
		parser.synchronize()
	}
}

func (c *Compiler) funDeclaration() {
	global := c.parseVariable("Expect function name.")
	c.markInitialized()
	c.compileFunction(TYPE_FUNCTION)
	c.defineVariable(global)
}

func (c *Compiler) varDeclaration() {
	global := c.parseVariable("Expect variable name.")

	if parser.match(TOKEN_EQUAL) {
		c.expression()
	} else {
		c.emitOp(OP_NIL)
	}

	parser.consume(TOKEN_SEMICOLON, "Expect ';' after variable declaration")

	c.defineVariable(global)
}

func (c *Compiler) statement() {
	if parser.match(TOKEN_PRINT) {
		c.printStatement()
	} else if parser.match(TOKEN_FOR) {
		c.forStatement()
	} else if parser.match(TOKEN_IF) {
		c.ifStatement()
	} else if parser.match(TOKEN_RETURN) {
		c.returnStatement()
	} else if parser.match(TOKEN_WHILE) {
		c.whileStatement()
	} else if parser.match(TOKEN_LEFT_BRACE) {
		c.beginScope()
		c.block()
		c.endScope()
	} else {
		c.expressionStatement()
	}
}

func (c *Compiler) printStatement() {
	c.expression()
	parser.consume(TOKEN_SEMICOLON, "Expect ';' after value")
	c.emitOp(OP_PRINT)
}

func (c *Compiler) returnStatement() {
	if c.kind == TYPE_SCRIPT {
		parser.error("Can't return from top-level code.")
	}

	if parser.match(TOKEN_SEMICOLON) {
		c.emitReturn()
	} else {
		c.expression()
		parser.consume(TOKEN_SEMICOLON, "Expect ';' after return value")
		c.emitOp(OP_RETURN)
	}
}

func (c *Compiler) expressionStatement() {
	c.expression()
	parser.consume(TOKEN_SEMICOLON, "Expect ';' after expression")
	c.emitOp(OP_POP)
}

func (c *Compiler) forStatement() {
	c.beginScope()
	parser.consume(TOKEN_LEFT_PAREN, "Expect '(' after 'for'.")

	if parser.match(TOKEN_SEMICOLON) {
		// no initializer
	} else if parser.match(TOKEN_VAR) {
		c.varDeclaration()
	} else {
		c.expression()
	}

	loopStart := c.currentChunk().Count()
	exitJump := -1

	if !parser.match(TOKEN_SEMICOLON) {
		c.expression()
		parser.consume(TOKEN_SEMICOLON, "Expect ';' after loop condition.")

		exitJump = c.emitJump(OP_JUMP_IF_FALSE)
		c.emitOp(OP_POP)
	}

	if !parser.match(TOKEN_RIGHT_PAREN) {
		bodyJump := c.emitJump(OP_JUMP)
		incStart := c.currentChunk().Count()

		c.expression()
		c.emitOp(OP_POP)
		parser.consume(TOKEN_RIGHT_PAREN, "Expect ')' after for clauses.")

		c.emitLoop(loopStart)
		loopStart = incStart
		c.patchJump(bodyJump)
	}

	c.statement() // loop body

	c.emitLoop(loopStart)
	if exitJump != -1 {
		c.patchJump(exitJump)
		c.emitOp(OP_POP)
	}

	c.endScope()
}

func (c *Compiler) ifStatement() {
	parser.consume(TOKEN_LEFT_PAREN, "Expect '(' after if.")
	c.expression()
	parser.consume(TOKEN_RIGHT_PAREN, "Expect ')' after condition.")

	thenJump := c.emitJump(OP_JUMP_IF_FALSE)
	c.emitOp(OP_POP) // pop off the condition

	c.statement()

	elseJump := c.emitJump(OP_JUMP)

	c.patchJump(thenJump)
	c.emitOp(OP_POP) // the condition, else case

	if parser.match(TOKEN_ELSE) {
		c.statement()
	}

	c.patchJump(elseJump)
}

func (c *Compiler) whileStatement() {
	loopStart := c.currentChunk().Count()
	parser.consume(TOKEN_LEFT_PAREN, "Expect '(' after while.")
	c.expression()
	parser.consume(TOKEN_RIGHT_PAREN, "Expect ')' after condition.")

	exitJump := c.emitJump(OP_JUMP_IF_FALSE)
	c.emitOp(OP_POP)
	c.statement()

	c.emitLoop(loopStart)

	c.patchJump(exitJump)
	c.emitOp(OP_POP)
}

func (c *Compiler) expression() {
	c.parsePrecedence(PREC_ASSIGNMENT)
}

func (c *Compiler) block() {
	for !parser.check(TOKEN_RIGHT_BRACE) && !parser.check(TOKEN_EOF) {
		c.declaration()
	}

	parser.consume(TOKEN_RIGHT_BRACE, "Expect '}' after block.")
}

func (c *Compiler) compileFunction(kind FunctionType) {
	local := NewCompiler(kind, c)
	local.beginScope()

	parser.consume(TOKEN_LEFT_PAREN, "Expect '(' after function name.")

	if !parser.check(TOKEN_RIGHT_PAREN) {
		for {
			local.function.arity++
			if local.function.arity > 255 {
				parser.errorAtCurrent("Can't have more than 255 parameters, you animal.")
			}

			constant := local.parseVariable("Expect parameter name.")
			local.defineVariable(constant)

			if !parser.match(TOKEN_COMMA) {
				break
			}
		}
	}

	parser.consume(TOKEN_RIGHT_PAREN, "Expect ')' after parameters.")
	parser.consume(TOKEN_LEFT_BRACE, "Expect '{' before function body.")
	local.block()

	function := local.end()
	c.emitOpAndArg(OP_CONSTANT, c.makeConstant(function))
}

func (c *Compiler) number(_ bool) {
	n, err := strconv.ParseFloat(parser.previous.lexeme, 64)
	if err != nil {
		panic("strconv.ParseFloat failed somehow")
	}

	c.emitConstant(ValueNumber(n))
}

func (c *Compiler) string(_ bool) {
	s := parser.previous.lexeme
	c.emitConstant(ValueString(s[1 : len(s)-1]))
}

func (c *Compiler) variable(canAssign bool) {
	c.namedVariable(parser.previous, canAssign)
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

	if canAssign && parser.match(TOKEN_EQUAL) {
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
				parser.error("Can't read local variable in its own initializer")
			}

			return byte(i), nil
		}
	}

	return 0, errors.New("local var not found")
}

func (c *Compiler) grouping(_ bool) {
	c.expression()
	parser.consume(TOKEN_RIGHT_PAREN, "Expect ')' after expression.")
}

func (c *Compiler) unary(_ bool) {
	op := parser.previous.kind
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
	op := parser.previous.kind
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

func (c *Compiler) call(bool) {
	argCount := c.argumentList()
	c.emitOpAndArg(OP_CALL, argCount)
}

func (c *Compiler) literal(_ bool) {
	switch parser.previous.kind {
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

func (c *Compiler) and(bool) {
	endJump := c.emitJump(OP_JUMP_IF_FALSE)

	c.emitOp(OP_POP)
	c.parsePrecedence(PREC_AND)

	c.patchJump(endJump)
}

func (c *Compiler) or(bool) {
	elseJump := c.emitJump(OP_JUMP_IF_TRUE)

	c.emitOp(OP_POP)
	c.parsePrecedence(PREC_OR)

	c.patchJump(elseJump)
}

func (c *Compiler) getRule(op TokenType) ParseRule {
	return c.rules[op]
}

func (c *Compiler) parsePrecedence(precedence Precedence) {
	parser.advance()
	prefixRule := c.getRule(parser.previous.kind).prefix
	if prefixRule == nil {
		parser.error("Expect expression.")
		return
	}

	canAssign := precedence <= PREC_ASSIGNMENT
	prefixRule(canAssign)

	for precedence <= c.getRule(parser.current.kind).precedence {
		parser.advance()
		infixRule := c.getRule(parser.previous.kind).infix
		infixRule(canAssign)
	}
}

func (c *Compiler) parseVariable(errMsg string) byte {
	parser.consume(TOKEN_IDENTIFIER, errMsg)

	c.declareVariable()
	if c.scopeDepth > 0 {
		return 0
	}

	return c.identifierConstant(parser.previous)
}

func (c *Compiler) markInitialized() {
	if c.scopeDepth == 0 {
		return
	}

	c.locals[c.localCount-1].depth = c.scopeDepth
}

func (c *Compiler) defineVariable(global byte) {
	if c.scopeDepth > 0 {
		c.markInitialized()
		return
	}

	c.emitOpAndArg(OP_DEFINE_GLOBAL, global)
}

func (c *Compiler) argumentList() byte {
	argCount := 0

	if !parser.check(TOKEN_RIGHT_PAREN) {
		for {
			c.expression()
			argCount++

			if argCount == 255 {
				parser.error("Can't have more than 255 arguments.")
			}

			if !parser.match(TOKEN_COMMA) {
				break
			}
		}
	}

	parser.consume(TOKEN_RIGHT_PAREN, "Expect ')' after arguments")
	return byte(argCount)
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

	name := parser.previous

	for i := c.localCount - 1; i >= 0; i-- {
		local := c.locals[i]
		if local.depth != -1 && local.depth < c.scopeDepth {
			break
		}

		if identifiersEqual(name, local.name) {
			parser.error("Already a variable with this name in this scope")
		}
	}

	c.addLocal(name)
}

func (c *Compiler) addLocal(name Token) {
	if c.localCount == UINT8_COUNT {
		parser.error("Too many local variables in scope")
		return
	}

	c.locals[c.localCount] = Local{name, -1}
	c.localCount++
}

// these functions all write to our chunk
func (c *Compiler) currentChunk() *Chunk {
	return c.function.chunk
}

func (c *Compiler) emitOp(op OpCode) {
	c.emitByte(byte(op))
}

func (c *Compiler) emitOpAndArg(op OpCode, arg byte) {
	c.emitBytes(byte(op), arg)
}

func (c *Compiler) emitByte(item byte) {
	c.currentChunk().Write(item, parser.previous.line)
}

func (c *Compiler) emitBytes(item1 byte, item2 byte) {
	c.emitByte(item1)
	c.emitByte(item2)
}

func (c *Compiler) emitJump(op OpCode) int {
	c.emitOp(op)
	c.emitBytes(0xff, 0xff)
	return c.currentChunk().Count() - 2
}

func (c *Compiler) emitLoop(start int) {
	c.emitOp(OP_LOOP)

	offset := c.currentChunk().Count() - start + 2
	if offset > math.MaxUint16 {
		parser.error("Loop body too large.")
	}

	c.emitByte(byte((offset >> 8) & 0xff))
	c.emitByte(byte(offset & 0xff))
}

func (c *Compiler) patchJump(offset int) {
	// -2 adjusts for bytecode of the jump op itself
	jump := c.currentChunk().Count() - offset - 2

	if jump > math.MaxUint16 {
		parser.error("Too much code to jump over")
	}

	c.currentChunk().code[offset] = byte((jump >> 8) & 0xff)
	c.currentChunk().code[offset+1] = byte(jump & 0xff)
}

func (c *Compiler) emitConstant(value Value) {
	c.emitOpAndArg(OP_CONSTANT, c.makeConstant(value))
}

func (c *Compiler) emitReturn() {
	c.emitOp(OP_NIL)
	c.emitOp(OP_RETURN)
}

func (c *Compiler) makeConstant(value Value) byte {
	constant := c.currentChunk().AddConstant(value)
	if constant > math.MaxUint8 {
		parser.error("Too many constants in one chunk")
		return 0
	}

	return byte(constant)
}

func (c *Compiler) end() *ValueFunction {
	c.emitReturn()

	function := c.function

	if DEBUG_PRINT_CODE && !parser.hadError {
		name := function.name

		if function.name == "" {
			name = "<script>"
		}

		c.currentChunk().Disassemble(name)
	}

	return c.function
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
