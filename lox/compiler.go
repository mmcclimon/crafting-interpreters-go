package lox

import "fmt"

func Compile(source string) {
	scanner := NewScanner(source)

	line := -1

	for {
		token := scanner.ScanToken()
		if token.line != line {
			fmt.Printf("%4d ", token.line)
			line = token.line
		} else {
			fmt.Printf("   | ")
		}
		fmt.Printf("%2d '%s'\n", token.kind, token.lexeme)

		if token.kind == TOKEN_EOF {
			break
		}
	}
}
