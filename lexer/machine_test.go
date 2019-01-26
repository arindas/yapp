package lexer

import (
	// io "io"
	fmt "fmt"
	strings "strings"
	testing "testing"
)

// grammar to be tested
// A <- aAb|nil
func automata() Bounds {
	a := NewLexState("char", Storer, true)
	b := NewLexState("char", Matcher, true)

	a.AddDest('a', a)
	a.AddDest('b', b)

	b.AddDest('b', b)
	b.AddDest(Path(EOF), nil)

	return Bounds{a, nil}
}

func runematcher() *RuneMatcher {
	matcher := NewRuneMatcher()
	matcher.EnlistMapping('a', 'b')
	return matcher
}

const (
	bufSize = 256
)

type testCase struct {
	input string
	valid bool
}

func (t testCase) String() string {
	return fmt.Sprintf("%v => %v", t.input, t.valid)
}

var testCases = []testCase{
	{"asd", false},
	{"ab", true},
	{"aabb", true},
	{"abab", false},
	{"baba", false},
	{"abba", false},
	{"baab", false},
}

func TestMatcher(t *testing.T) {
	var err error
	bounds, matcher := automata(), runematcher()
	for index, testcase := range testCases {
		t.Logf("test case #%v\n", index)
		matcher.Reset()
		lexer := NewLexer(strings.NewReader(testcase.input), bufSize)
		machine := NewMachine(lexer, bounds, matcher)

		var token Token
		for token != EOLexToken {
			token, err = machine.NextToken()
			t.Logf("token: %v\n", token)
		}

		if testcase.valid != (err == nil && matcher.IsMatched()) {
			m := make(map[bool]string)
			m[true], m[false] = "matched", "unmatched"
			t.Errorf("[error: %v] [matcher status-> %v]\n",
				err, m[matcher.IsMatched()])
			t.Errorf("test case %v: (%v) not passed.\n", index, testcase)
		}

	}
}
