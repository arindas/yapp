package lexer

import (
	list "container/list"
	"fmt"
)

// RuneMatcher is a component to keep track of matchable runes
type RuneMatcher struct {
	stack    *list.List    // stack for keeping record of matched characters
	matchmap map[rune]rune // map to record character match mappings
}

// NewRuneMatcher creates a new instance of RuneMatcher
func NewRuneMatcher() *RuneMatcher {
	return &RuneMatcher{
		stack:    list.New(),
		matchmap: make(map[rune]rune),
	}
}

// EnlistMapping enlists a new mapping in this RuneMatcher
func (m *RuneMatcher) EnlistMapping(r1, r2 rune) {
	m.matchmap[r1], m.matchmap[r2] = r2, r1
}

// Store stores a rune to matched against in this RuneMatcher
func (m *RuneMatcher) Store(r rune) bool {
	var wasStored bool
	if _, wasStored = m.matchmap[r]; wasStored {
		m.stack.PushBack(r)
	}
	return wasStored
}

// Match matches the given rune against the rune at TOS of matcher's stack
func (m *RuneMatcher) Match(r rune) bool {
	tos := m.stack.Back()
	R, registered := m.matchmap[r]
	return tos != nil && registered &&
		R == m.stack.Remove(tos).(rune)
}

// Reset resets this matcher
func (m *RuneMatcher) Reset() { m.stack.Init() }

// IsMatched checks whether the RuneMatcher is in
// matched state if its stack is completely empty
func (m *RuneMatcher) IsMatched() bool { return m.stack.Back() == nil }

// Path denotes a link or input to traverse
// from the current state to the next state
type Path rune

// DefaultPath is the path to be taken
// in the event of unexpected or no input
const DefaultPath Path = 0

// LexStateType denotes the type of the state for a LexMachine
type LexStateType int

const (
	// Buffer pushes back input after consuming it
	Buffer LexStateType = 0
	// Feeder simply feeds on input
	Feeder LexStateType = 1
	// Storer stores runes to be mathced
	Storer LexStateType = 2
	// Matcher matches runes
	Matcher LexStateType = 3
)

// String (LexStateType) produces a string representation of a LexStateType
func (l LexStateType) String() string {
	stateNames := []string{"Buffer",
		"Feeder", "Storer", "Matcher"}
	return stateNames[int(l)]
}

// LexState denotes an individual
// state in a pushdown automata
type LexState struct {
	emitter        bool
	assocTokenType TokenType
	next           map[Path]*LexState
	stateType      LexStateType
}

// NewLexState creates a new instance of LexState
func NewLexState(tt TokenType, typ LexStateType, isEmitter bool) *LexState {
	return &LexState{isEmitter, tt, make(map[Path]*LexState), typ}
}

// AddDest adds a new LexState as a neighbour to this LexState
func (l *LexState) AddDest(p Path, dest *LexState) { l.next[p] = dest }

// String (LexState) produces a string representation of LexState
func (l LexState) String() string {
	return fmt.Sprintf("%v:[%v]",
		l.assocTokenType, l.stateType)
}

// Bounds represents the bounding states of lex machine
type Bounds struct {
	start, end *LexState
}

// Machine is a PDA which represents a lexing
// machine attached to the "tape reader"
type Machine struct {
	lexer        *Lexer       // lexer for reading: tape reader
	bounds       Bounds       // bounds of machine to ensure finiteness
	tokens       chan Token   // channel to deliver tokens
	currState    *LexState    // current state of Lexing machine
	lastReadRune rune         // path leading to current state
	matcher      *RuneMatcher // stack associated with states of this machine
}

// NewMachine creates a new instance of lexer.Machine
func NewMachine(lexer *Lexer, bounds Bounds, matcher *RuneMatcher) *Machine {
	return &Machine{lexer, bounds, make(chan Token, 2), bounds.start, 0, matcher}
}

// Reset resets all the components of this machine.
func (m *Machine) Reset() {
	// reset lexer: "tape reader"
	m.lexer.Reset()
	// close previous channel and open another
	close(m.tokens)
	m.tokens = make(chan Token, 2)
	// reset current state to start
	m.currState = m.bounds.start
	// reset last read rune to default value
	m.lastReadRune = 0
	// empty PDA stack
	m.matcher.Reset()
}

// AttachLexer attaches new lexer to this machine and resets it.
func (m *Machine) AttachLexer(l *Lexer) { m.lexer = l; m.Reset() }

// CanStep states whether this machine can step over to a new state
func (m *Machine) CanStep() bool {
	return m.currState != nil &&
		m.currState != m.bounds.end
}

// Finished states whether this machine has reached the finishing state.
func (m *Machine) Finished() bool { return m.currState == m.bounds.end }

// Step steps one LexState of the machine
func (m *Machine) Step() error {

	path := Path(m.lexer.Next())

	var nextState *LexState
	var pathExists bool

	if nextState, pathExists = m.currState.next[path]; !pathExists {
		nextState, pathExists = m.currState.next[DefaultPath]
	}

	if !pathExists {
		return fmt.Errorf("Invalid state: %v, no path with input: %v",
			m.currState, rune(path))
	}

	switch m.currState.stateType {
	case Buffer:
		m.lexer.Backup()
	case Storer:
		if m.lastReadRune != 0 && !m.matcher.Store(m.lastReadRune) {
			return fmt.Errorf("unregistered rune: %v", m.lastReadRune)
		}
	case Matcher:
		if !m.matcher.Match(m.lastReadRune) {
			return fmt.Errorf("unmatched rune: %v", m.lastReadRune)
		}
	}

	if m.currState.emitter {
		if m.currState.stateType == Buffer {
			return fmt.Errorf(
				"unsuitable state: %v for emitting token",
				m.currState)
		}
		m.tokens <- m.lexer.Emit(m.currState.assocTokenType)
	}

	m.currState, m.lastReadRune = nextState, rune(path)

	return nil

}

// NextToken returns the next Token from the channel
func (m *Machine) NextToken() (Token, error) {
	for m.CanStep() {
		select {
		case token := <-m.tokens:
			return token, nil
		default:
			if err := m.Step(); err != nil {
				return EOLexToken, err
			}
		}
	}

	return EOLexToken, nil
}
