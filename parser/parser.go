package parser

import (
	list "container/list"
	lexer "scratch.io/lexer"
)

type ElementType interface{}

// Type to denote a grammatical element.
type Element struct {
	Type     ElementType    // type of the element
	Tokens   []*lexer.Token // tokens associated with this element
	Children *list.List     // list of sub elements
}

// Component to keep track of matchable runes
type RuneMatcher struct {
	stack    *list.List    // stack for keeping record of matched characters
	matchmap map[rune]rune // map to record character match mappings
}

// Enlists a new mapping in this RuneMatcher
func (m *RuneMatcher) EnlistMapping(r1, r2 rune) {
	m.matchmap[r1], m.matchmap[r2] = r2, r1
}

// Stores a rune to matched against in this RuneMatcher
func (m *RuneMatcher) Store(r rune) bool {
	var wasStored bool
	if _, wasStored = m.matchmap[r]; wasStored {
		m.stack.PushBack(r)
	}
	return wasStored
}

func (m *RuneMatcher) Match(r rune) bool {
	tos := m.stack.Back()
	R, registered := m.matchmap[r]
	return tos != nil && registered &&
		r == R && m.stack.Remove(tos) != nil
}
