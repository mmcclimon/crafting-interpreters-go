package main

import (
	"bufio"
	"fmt"
	"log"
	"os"

	"github.com/mmcclimon/glox/lox"
)

func main() {
	if len(os.Args) == 1 {
		repl()
	} else if len(os.Args) == 2 {
		runFile(os.Args[1])
	} else {
		fmt.Fprintln(os.Stderr, "usage: lox [path]")
		os.Exit(64)
	}
}

func repl() {
	vm := lox.NewVM()

	scanner := bufio.NewScanner(os.Stdin)

	fmt.Print("> ")
	for scanner.Scan() {
		line := scanner.Text()
		vm.InterpretString(line)

		fmt.Print("> ")
	}

	if err := scanner.Err(); err != nil {
		log.Println(err)
	}
}

func runFile(filename string) {
	bytes, err := os.ReadFile(filename)

	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading file: %s\n", err)
		os.Exit(74)
	}

	vm := lox.NewVM()
	err = vm.InterpretString(string(bytes))

	if err == lox.InterpretCompileError {
		os.Exit(65)
	} else if err == lox.InterpretRuntimeError {
		os.Exit(70)
	}
}
