package lox

type TokenType int

const (
	// Single-character tokens.
	TOKEN_LEFT_PAREN TokenType = iota
	TOKEN_RIGHT_PAREN
	TOKEN_LEFT_BRACE
	TOKEN_RIGHT_BRACE
	TOKEN_COMMA
	TOKEN_DOT
	TOKEN_MINUS
	TOKEN_PLUS
	TOKEN_SEMICOLON
	TOKEN_SLASH
	TOKEN_STAR

	// One/two character tokens
	TOKEN_BANG
	TOKEN_BANG_EQUAL
	TOKEN_EQUAL
	TOKEN_EQUAL_EQUAL
	TOKEN_GREATER
	TOKEN_GREATER_EQUAL
	TOKEN_LESS
	TOKEN_LESS_EQUAL

	// Literals
	TOKEN_IDENTIFIER
	TOKEN_STRING
	TOKEN_NUMBER

	// Keywords
	TOKEN_AND
	TOKEN_CLASS
	TOKEN_ELSE
	TOKEN_FALSE
	TOKEN_FOR
	TOKEN_FUN
	TOKEN_IF
	TOKEN_NIL
	TOKEN_OR
	TOKEN_PRINT
	TOKEN_RETURN
	TOKEN_SUPER
	TOKEN_THIS
	TOKEN_TRUE
	TOKEN_VAR
	TOKEN_WHILE

	TOKEN_ERROR
	TOKEN_EOF
)

type Scanner struct {
	source  string
	start   int
	current int
	line    int
}

type Token struct {
	kind   TokenType
	lexeme string
	line   int
}

var reservedWords map[string]TokenType
var tokenNames map[TokenType]string

func init() {
	reservedWords = map[string]TokenType{
		"and":    TOKEN_AND,
		"class":  TOKEN_CLASS,
		"else":   TOKEN_ELSE,
		"false":  TOKEN_FALSE,
		"for":    TOKEN_FOR,
		"fun":    TOKEN_FUN,
		"if":     TOKEN_IF,
		"nil":    TOKEN_NIL,
		"or":     TOKEN_OR,
		"print":  TOKEN_PRINT,
		"return": TOKEN_RETURN,
		"super":  TOKEN_SUPER,
		"this":   TOKEN_THIS,
		"true":   TOKEN_TRUE,
		"var":    TOKEN_VAR,
		"while":  TOKEN_WHILE,
	}

	tokenNames = map[TokenType]string{
		TOKEN_LEFT_PAREN:    "(",
		TOKEN_RIGHT_PAREN:   ")",
		TOKEN_LEFT_BRACE:    "{",
		TOKEN_RIGHT_BRACE:   "}",
		TOKEN_COMMA:         ",",
		TOKEN_DOT:           ".",
		TOKEN_MINUS:         "-",
		TOKEN_PLUS:          "+",
		TOKEN_SEMICOLON:     ";",
		TOKEN_SLASH:         "/",
		TOKEN_STAR:          "*",
		TOKEN_BANG:          "!",
		TOKEN_BANG_EQUAL:    "!=",
		TOKEN_EQUAL:         "=",
		TOKEN_EQUAL_EQUAL:   "==",
		TOKEN_GREATER:       ">",
		TOKEN_GREATER_EQUAL: ">=",
		TOKEN_LESS:          "<",
		TOKEN_LESS_EQUAL:    "<=",
		TOKEN_IDENTIFIER:    "<identifier>",
		TOKEN_STRING:        "<string>",
		TOKEN_NUMBER:        "<number>",
		TOKEN_AND:           "&&",
		TOKEN_CLASS:         "class",
		TOKEN_ELSE:          "else",
		TOKEN_FALSE:         "false",
		TOKEN_FOR:           "for",
		TOKEN_FUN:           "fun",
		TOKEN_IF:            "if",
		TOKEN_NIL:           "nil",
		TOKEN_OR:            "or",
		TOKEN_PRINT:         "print",
		TOKEN_RETURN:        "return",
		TOKEN_SUPER:         "super",
		TOKEN_THIS:          "this",
		TOKEN_TRUE:          "true",
		TOKEN_VAR:           "var",
		TOKEN_WHILE:         "while",
		TOKEN_ERROR:         "<error>",
		TOKEN_EOF:           "<eof>",
	}
}

func NewScanner(source string) *Scanner {
	return &Scanner{source: source, line: 1}
}

// helpers
func isDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

func isAlpha(c byte) bool {
	return (c >= 'a' && c <= 'z') ||
		(c >= 'A' && c <= 'Z') ||
		c == '_'
}

func (s *Scanner) ScanToken() Token {
	s.skipWhitespace()
	s.start = s.current

	if s.isAtEnd() {
		return s.makeToken(TOKEN_EOF)
	}

	c := s.advance()

	switch {
	case isDigit(c):
		return s.number()
	case isAlpha(c):
		return s.identifier()
	}

	switch c {
	// single chars
	case '(':
		return s.makeToken(TOKEN_LEFT_PAREN)
	case ')':
		return s.makeToken(TOKEN_RIGHT_PAREN)
	case '{':
		return s.makeToken(TOKEN_LEFT_BRACE)
	case '}':
		return s.makeToken(TOKEN_RIGHT_BRACE)
	case ';':
		return s.makeToken(TOKEN_SEMICOLON)
	case ',':
		return s.makeToken(TOKEN_COMMA)
	case '.':
		return s.makeToken(TOKEN_DOT)
	case '-':
		return s.makeToken(TOKEN_MINUS)
	case '+':
		return s.makeToken(TOKEN_PLUS)
	case '/':
		return s.makeToken(TOKEN_SLASH)
	case '*':
		return s.makeToken(TOKEN_STAR)

	// double chars
	case '!':
		if s.match('=') {
			return s.makeToken(TOKEN_BANG_EQUAL)
		}

		return s.makeToken(TOKEN_BANG)

	case '=':
		if s.match('=') {
			return s.makeToken(TOKEN_EQUAL_EQUAL)
		}

		return s.makeToken(TOKEN_EQUAL)

	case '<':
		if s.match('=') {
			return s.makeToken(TOKEN_LESS_EQUAL)
		}

		return s.makeToken(TOKEN_LESS)

	case '>':
		if s.match('=') {
			return s.makeToken(TOKEN_GREATER_EQUAL)
		}

		return s.makeToken(TOKEN_GREATER)

	case '"':
		return s.string()

	}

	return s.errorToken("Unexpected character.")
}

func (s *Scanner) skipWhitespace() {
	for {
		c := s.peek()
		switch c {
		case '\r', '\t', ' ':
			s.advance()
			break
		case '\n':
			s.line++
			s.advance()
			break
		case '/':
			if s.peekNext() == '/' {
				// A comment goes until the end of the line.
				for s.peek() != '\n' && !s.isAtEnd() {
					s.advance()
				}
			} else {
				return
			}
		default:
			return
		}
	}
}

func (s *Scanner) string() Token {
	for s.peek() != '"' && !s.isAtEnd() {
		if s.peek() == '\n' {
			s.line++
		}
		s.advance()
	}

	if s.isAtEnd() {
		return s.errorToken("Unterminated string.")
	}

	s.advance() // closing quote

	return s.makeToken(TOKEN_STRING)
}

func (s *Scanner) number() Token {
	for isDigit(s.peek()) {
		s.advance()
	}

	// decimal part
	if s.peek() == '.' && isDigit(s.peekNext()) {
		s.advance() // eat the .

		for isDigit(s.peek()) {
			s.advance()
		}
	}

	return s.makeToken(TOKEN_NUMBER)
}

func (s *Scanner) identifier() Token {
	for isAlpha(s.peek()) || isDigit(s.peek()) {
		s.advance()
	}

	return s.makeToken(s.identifierType())
}

func (s *Scanner) identifierType() TokenType {
	// Ok look, so I implemented the trie that's in the book, which was fine,
	// but I'm in a higher-level language that has hash maps, and I might as
	// well use them, ok?
	word := s.source[s.start:s.current]
	if kind, ok := reservedWords[word]; ok == true {
		return kind
	}

	return TOKEN_IDENTIFIER
}

func (s *Scanner) isAtEnd() bool {
	return s.current >= len(s.source)
}

func (s *Scanner) makeToken(kind TokenType) Token {
	return Token{
		kind:   kind,
		lexeme: string(s.source[s.start:s.current]),
		line:   s.line,
	}
}

func (s *Scanner) errorToken(msg string) Token {
	return Token{
		kind:   TOKEN_ERROR,
		lexeme: msg,
		line:   s.line,
	}
}

func (s *Scanner) peek() byte {
	if s.isAtEnd() {
		return 0
	}

	return s.source[s.current]
}

func (s *Scanner) peekNext() byte {
	if s.current+1 >= len(s.source) {
		return 0
	}

	return s.source[s.current+1]
}

func (s *Scanner) advance() byte {
	s.current++
	return s.source[s.current-1]
}

func (s *Scanner) match(expected byte) bool {
	if s.isAtEnd() || s.source[s.current] != expected {
		return false
	}

	s.current++
	return true
}

func (tt TokenType) String() string {
	return tokenNames[tt]
}
