package parser

import (
	list "container/list"
	fmt "fmt"
	lexer "scratch.io/lexer"
)

// Represents a grammatical element. Every grammar can be composed as a hierarchy of grammatical elements.
// Every grammatical element is associated with a token and the element name as specified in the grammar
type Element struct {
	Name  string      // name of the elementType
	Token lexer.Token // token associated with this elementType
}

func GetElement(name string, tokenType lexer.TokenType) Element {
	return Element{name, lexer.Token{Lexeme: "", Type: tokenType}}
}

type ElementTree struct {
	Elem     Element    // element associated with this tree
	Children *list.List // list of children of this tree
}

type Path rune

const DefaultPath Path = 0

// Represents a matcher for tokens that come in pairs
type RuneMatcher struct {
	matchmap map[rune]rune // map for matching opening closing runes
	stack    *list.List    // stack for matching characters
}

// Creates a new RuneMatcher with an empty match map and stack
func NewRuneMatcher() *RuneMatcher {
	matcher := &RuneMatcher{
		matchmap: make(map[rune]rune),
		stack:    list.New(),
	}

	return matcher
}

func (matcher *RuneMatcher) store(r rune) { matcher.stack.PushBack(r) }

// Registers a pair of matchable characters
func (matcher *RuneMatcher) Register(r1, r2 rune) {
	matchmap := matcher.matchmap
	matchmap[r1], matchmap[r2] = r2, r1
}

func (matcher *RuneMatcher) match(r rune) bool {
	stack, matchmap := matcher.stack, matcher.matchmap

	if matchmap[r] == stack.Back().Value {
		stack.Remove(stack.Back())
		return true
	}

	return false
}

// Collection of bool variables to denote the type of State
type StateType struct {
	IsMatcher, IsBuffer, IsTerminal, IsEmitter bool
}

func _uint8(b bool) uint8 {
	if b {
		return 1
	}

	return 0
}

// Returns the state type repsresented by the 8bit integer
func GetStateType(code uint8) StateType {
	return StateType{
		IsMatcher:  code&0x8 == 0x8,
		IsBuffer:   code&0x4 == 0x4,
		IsTerminal: code&0x2 == 0x2,
		IsEmitter:  code&0x1 == 0x1,
	}
}

func EncodeStateType(t StateType) uint8 {
	return _uint8(t.IsMatcher)<<3 |
		_uint8(t.IsBuffer)<<2 |
		_uint8(t.IsTerminal)<<1 |
		_uint8(t.IsEmitter)
}

// Represents states in an automata. Every state is assocaited with a list of paths leading to other states
// in the automata and a flag denoting whether this is an accepting state is not. Implementations of
// automata should create a new elementTree and store it in the tree stack if this is a non accepting
// state. In the case of an accepting state the tos element should be popped and added to the list of children
// of the new tos element.
type State struct {
	Type              StateType       // type of this state: terminal, buffer or starting
	Matcher           *RuneMatcher    // buffer to be used as a memory unit
	DestStates        map[Path]*State // mapping from runes to states to obtain the next destination state from a rune
	AssociatedElement Element         // element assocaited with this automata state (in its default state)
}

// Constructs a new state with Type obtained from te given code, matcher from the given matcher
// and associates the given element with the newly constructed state
func NewState(code uint8, matcher *RuneMatcher, element Element) *State {
	state := &State{
		Type:              GetStateType(code),
		Matcher:           matcher,
		DestStates:        make(map[Path]*State),
		AssociatedElement: element,
	}

	return state
}

// Add a new path leading to the given dest state from the given state
func (s *State) AddPath(p Path, dest *State) { s.DestStates[p] = dest }

// Terminal State that represents an error
var ErrorState State = State{}

func (s *State) TargetState(p Path) (*State, bool) {
	targetState, pathTaken := s.DestStates[p], true

	if targetState == nil {
		pathTaken = false // the given path was not taken
		targetState = s.DestStates[DefaultPath]
	}

	if targetState == nil {
		targetState = &ErrorState // error state
	}

	return targetState, pathTaken
}

type MachineBounds struct {
	Start, End *State
}

// Represents a deterministic finite automata. It comprises of a graph of State instances interconnected
// with Path instances which act as edges. Each path is traversed as the machine is fed with new input.
// The machine keeps record of three states: the starting state, the final accepting state, and the current
// state that the machine is in. As the machine steps over a new input, it transitions to the state following
// the path for the given input.
type Machine struct {
	Bounds           MachineBounds // bounding states of this machine
	Lexer            *lexer.Lexer  // lexer associated with this
	elementTreeStack *list.List    // stack of element trees associated with this machine
	currentState     *State        // current state of the machine
}

func NewMachine(bounds MachineBounds, l *lexer.Lexer) *Machine {
	machine := &Machine{
		Bounds:           bounds,
		Lexer:            l,
		elementTreeStack: list.New(),
		currentState:     bounds.Start,
	}

	return machine
}

func (p Path) Error() string {
	return fmt.Sprintf("Unexpected rune: %q", rune(p))
}

func (m *Machine) createAndPushElementTree() {
	// create a new element Tree
	elementTree := &ElementTree{
		Elem:     m.currentState.AssociatedElement,
		Children: list.New(),
	}
	// push the elementTree in the elementTree stack
	m.elementTreeStack.PushBack(elementTree)
}

func (m *Machine) insertElementTree() {
	elementTree := m.elementTreeStack.Back().Value.(*ElementTree)
	// pop an elementTree from the stack
	m.elementTreeStack.Remove(m.elementTreeStack.Back())
	// add the popped elementTree as a child of the new top of stack
	m.elementTreeStack.Back().Value.(*ElementTree).Children.PushBack(elementTree)
}

func (m *Machine) emitToken() {
	// peek from the element tree stack
	elementTree := m.elementTreeStack.Back().Value.(*ElementTree)
	// assign the emitted token from the lexer to the element in the element tree
	elementTree.Elem.Token = m.Lexer.Emit(elementTree.Elem.Token.Type)
}

func (m *Machine) manageElementTree(runeToBeMatched rune) {
	state := m.currentState

	// if this is not a terminal state
	if !state.Type.IsTerminal {
		// create a new grammatical element
		// and push it to the element tree stack
		m.createAndPushElementTree()
		// if this is matcher
		if state.Type.IsMatcher {
			// store opening matchable rune
			state.Matcher.store(runeToBeMatched)
		}
	} else if !state.Type.IsBuffer {
		// close the last element and insert
		// to the list of children of its parent
		// grammatical element
		m.insertElementTree()
		// if this a matcher
		if state.Type.IsMatcher {
			// match this mathcable rune with the
			// stored matchable rune
			state.Matcher.match(runeToBeMatched)
		}
	}

	// if this is an emitter state
	if state.Type.IsEmitter {
		// emit the token currently being parsed
		// and store it in the element being read
		m.emitToken()
	}

}

// States whether this machine can step into any more states or not.
// It is meant to be used in conjunction with Step() as follows:
//
// 		func (m *Machine) Run() {
//			for m.CanStep() {
//				m.Step()
//			}
//		}
//
func (m *Machine) CanStep() bool {
	return m.currentState != m.Bounds.End ||
		m.currentState != &ErrorState
}

// Updates the current state of the machine by reading the next rune from the lexer.
// If a this is a starting state, a new element is created, which is stored in an element
// tree. Now the tree is pushed onto the tree stack. If this is a terminal state, the
// last opened grammatical element is closed by popping the latest added element tree
// and adding it as a child of its parent. Buffer states do not make any modifications
// to the element tree. Emitter states emit the token that was being read and store it
// in the element of the element tree currently at the top of element tree stack. Any
// state can be an emitter state.
func (m *Machine) Step() {
	if !m.CanStep() {
		return
	}

	nextRune := m.Lexer.Next()
	m.manageElementTree(nextRune)

	var pathTaken bool
	m.currentState, pathTaken =
		m.currentState.TargetState(Path(nextRune))

	if !pathTaken {
		m.Lexer.Backup()
	}

}

// Returns the parsed element tree if the machine has reached terminal state, nil otherwise.
// This returns the last element-tree in the element-tree stack stored in this machine. The
// element-tree returned contains the entire parsed respresentation of the input.
func (m *Machine) ParsedElementTree() *ElementTree {
	if m.CanStep() {
		return nil
	}

	return m.elementTreeStack.Back().Value.(*ElementTree)
}
