package lexer

import (
	// io "io"
	strings "strings"
	testing "testing"
)

// grammar to be tested
// A <- aAb|nil
func getPDA() Bounds {
	a_ := NewLexState("char", Storer, true)
	b := NewLexState("char", Matcher, true)

	a_.AddDest('a', a_)
	a_.AddDest('b', b)

	b.AddDest('b', b)
	b.AddDest(Path(EOF), nil)

	return Bounds{a_, nil}
}

const (
	bufSize = 256
)

type testCase struct {
	input string
	valid bool
}

func getMachineToLex(s string) *Machine {
	lexer := NewLexer(strings.NewReader(s), bufSize)

	matcher := NewRuneMatcher()
	matcher.EnlistMapping('a', 'b')

	return NewMachine(lexer, getPDA(), matcher)
}

func testTestCase(m *Machine, tc testCase, t *testing.T) bool {
	lexer := NewLexer(strings.NewReader(tc.input), bufSize)
	m.AttachLexer(lexer)
	token, hasTokens := m.NextToken()

	for ; hasTokens; token, hasTokens = m.NextToken() {
		t.Logf("token read: %v\n", token)
	}

	return m.matcher.IsMatched()
}

func TestMatcher(t *testing.T) {

	// testCases := []testCase{
	// 	 {"asd", false},
	//	 {"ab", true},
	//	 {"aabb", true},
	//	 {"abab", false},
	//	 {"baba", false},
	//	 {"abba", false},
	//	 {"baab", false},
	// }
}
