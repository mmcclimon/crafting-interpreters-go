package jlox

import (
	"bufio"
	"fmt"
	"log"
	"os"
)

var hadError bool

func RunFile(filename string) {
	bytes, err := os.ReadFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading file: %s\n", err)
		os.Exit(74)
	}

	run(string(bytes))

	if hadError {
		os.Exit(65)
	}
}

func REPL() {
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Print("> ")
	for scanner.Scan() {
		line := scanner.Text()
		run(line)
		hadError = false

		fmt.Print("> ")
	}

	if err := scanner.Err(); err != nil {
		log.Println(err)
	}
}

func run(source string) {
	scanner := NewScanner(source)
	tokens := scanner.Tokens()

	for _, tok := range tokens {
		fmt.Println(tok)
	}
}

func Error(line int, message string) {
	report(line, "", message)
}

func report(line int, where, message string) {
	fmt.Fprintf(os.Stderr, "[line %d] Error%s: %s\n", line, where, message)
	hadError = true
}
