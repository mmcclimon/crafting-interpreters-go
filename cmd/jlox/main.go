package main

import (
	"fmt"
	"os"

	"github.com/mmcclimon/glox/jlox"
)

func main() {
	switch len(os.Args) {
	case 1:
		jlox.REPL()
	case 2:
		jlox.RunFile(os.Args[1])
	default:
		fmt.Fprintln(os.Stderr, "usage: lox [path]")
		os.Exit(64)
	}
}
