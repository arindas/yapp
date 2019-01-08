package parser

import (
	lexer "scratch.io/lexer"
	"testing"
)

// expr:
// a
// a + expr
//
// line:
// eof
// expr eof

func prototype() *Machine {
	line := GetElement("line", "eof/?")
	expr := GetElement("expr", "operands/operator")

	// states of our machine
	state_1 := NewState(0x0, nil, line)
	state_2 := NewState(0x0, nil, expr)
	state_2_star := NewState(0x3, nil, expr)
	state_1_star := NewState(0x3, nil, line)

	// transition functions
	state_1.AddPath(DefaultPath, state_2)
	state_1.AddPath(lexer.EOF, state_1_star)
	state_2.AddPath('a', state_2_star)
	state_2_star.AddPath('+', state_2)
	state_2_star.AddPath(lexer.EOF, state_1_star)

	return nil
}
