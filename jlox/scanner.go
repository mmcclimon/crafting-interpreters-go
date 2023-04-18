package jlox

import (
	"fmt"
	"strconv"

	"github.com/mmcclimon/glox/jlox/token"
)

var reservedWords map[string]token.Type

type Scanner struct {
	source  []rune
	tokens  []token.Token
	start   int
	current int
	line    int
}

func NewScanner(source string) *Scanner {
	return &Scanner{
		source: []rune(source),
		tokens: make([]token.Token, 0, 10),
		line:   1,
	}
}

func (s *Scanner) Tokens() []token.Token {
	for !s.isAtEnd() {
		s.start = s.current
		s.scanToken()
	}

	s.tokens = append(s.tokens, token.New(token.EOF, "", nil, s.line))
	return s.tokens
}

func (s *Scanner) isAtEnd() bool {
	return s.current >= len(s.source)
}

func (s *Scanner) advance() rune {
	ret := s.source[s.current]
	s.current++
	return ret
}

func isDigit(c rune) bool {
	return c >= '0' && c <= '9'
}

func isAlpha(c rune) bool {
	return (c >= 'a' && c <= 'z') ||
		(c >= 'A' && c <= 'Z') ||
		c == '_'
}

func isAlphaNumeric(c rune) bool {
	return isDigit(c) || isAlpha(c)
}

func (s *Scanner) addToken(kind token.Type) {
	text := string(s.source[s.start:s.current])
	s.tokens = append(s.tokens, token.New(kind, text, nil, s.line))
}

func (s *Scanner) addTokenWithValue(kind token.Type, value any) {
	text := string(s.source[s.start:s.current])
	s.tokens = append(s.tokens, token.New(kind, text, value, s.line))
}

func (s *Scanner) scanToken() {
	c := s.advance()
	switch c {
	case ' ', '\t', '\r':
		// nothing
	case '\n':
		s.line++
	case '(':
		s.addToken(token.LeftParen)
	case ')':
		s.addToken(token.RightParen)
	case '{':
		s.addToken(token.LeftBrace)
	case '}':
		s.addToken(token.RightBrace)
	case ',':
		s.addToken(token.Comma)
	case '.':
		s.addToken(token.Dot)
	case '-':
		s.addToken(token.Minus)
	case '+':
		s.addToken(token.Plus)
	case ';':
		s.addToken(token.Semicolon)
	case '*':
		s.addToken(token.Star)
	case '!':
		tok := token.Bang
		if s.match('=') {
			tok = token.BangEqual
		}
		s.addToken(tok)
	case '=':
		tok := token.Equal
		if s.match('=') {
			tok = token.EqualEqual
		}
		s.addToken(tok)
	case '<':
		tok := token.Less
		if s.match('=') {
			tok = token.LessEqual
		}
		s.addToken(tok)
	case '>':
		tok := token.Greater
		if s.match('=') {
			tok = token.GreaterEqual
		}
		s.addToken(tok)

	case '"':
		s.scanString()

	case '/':
		if s.match('/') {
			// comment; skip until the end of the line
			for s.peek() != '\n' && !s.isAtEnd() {
				s.advance()
			}
			break
		}

		s.addToken(token.Slash)

	default:
		switch {
		case isDigit(c):
			s.scanNumber()
		case isAlpha(c):
			s.scanIdentifier()
		default:
			Error(s.line, fmt.Sprintf("unexpected character %c", c))
		}
	}
}

func (s *Scanner) scanString() {
	for s.peek() != '"' && !s.isAtEnd() {
		if s.peek() == '\n' {
			s.line++
		}
		s.advance()
	}

	if s.isAtEnd() {
		Error(s.line, "Unterminated string.")
		return
	}

	s.advance() // eat closing quote
	val := s.source[s.start+1 : s.current-1]
	s.addTokenWithValue(token.String, string(val))
}

func (s *Scanner) scanNumber() {
	for isDigit(s.peek()) {
		s.advance()
	}

	// Look for a fractional part.
	if s.peek() == '.' && isDigit(s.peekNext()) {
		// Consume the "."
		s.advance()

		for isDigit(s.peek()) {
			s.advance()
		}
	}

	val, err := strconv.ParseFloat(string(s.source[s.start:s.current]), 64)
	if err != nil {
		panic(err)
	}

	s.addTokenWithValue(token.Number, val)
}

func (s *Scanner) scanIdentifier() {
	for isAlphaNumeric(s.peek()) {
		s.advance()
	}

	text := string(s.source[s.start:s.current])
	kind := token.Identifier
	if keyword, ok := reservedWords[text]; ok {
		kind = keyword
	}

	s.addToken(kind)
}

func (s *Scanner) match(expected rune) bool {
	if s.isAtEnd() || s.source[s.current] != expected {
		return false
	}

	s.current++
	return true
}

func (s *Scanner) peek() rune {
	if s.isAtEnd() {
		return 0
	}

	return s.source[s.current]
}

func (s *Scanner) peekNext() rune {
	if s.current+1 >= len(s.source) {
		return 0
	}

	return s.source[s.current+1]
}

// fill in reserved words
func init() {
	reservedWords = map[string]token.Type{
		"and":    token.And,
		"class":  token.Class,
		"else":   token.Else,
		"false":  token.False,
		"for":    token.For,
		"fun":    token.Fun,
		"if":     token.If,
		"nil":    token.Nil,
		"or":     token.Or,
		"print":  token.Print,
		"return": token.Return,
		"super":  token.Super,
		"this":   token.This,
		"true":   token.True,
		"var":    token.Var,
		"while":  token.While,
	}
}
