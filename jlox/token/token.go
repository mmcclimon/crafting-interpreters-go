package token

import "fmt"

type Token struct {
	kind    Type
	lexeme  string
	literal any
	line    int
}

func New(kind Type, lexeme string, literal any, line int) Token {
	return Token{
		kind, lexeme, literal, line,
	}
}

func (t Token) String() string {
	return fmt.Sprintf("%4d %-16s %-10s", t.line, t.kind, t.lexeme)
}
