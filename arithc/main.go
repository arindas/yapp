package main

import (
	bufio "bufio"
	fmt "fmt"
	os "os"
	// parser "scratch.io/parser"
	lexer "scratch.io/lexer"
)

func main() {
	if len(os.Args) > 1 {
		return
	}
	reader := bufio.NewReader(os.Stdin)
	capacity := 256
	l := lexer.NewLexer(reader, capacity)

	for r := l.Next(); r != lexer.EOF; r = l.Next() {
		fmt.Printf("rune: %q\n", r)
	}
}
