package lexer

import (
	list "container/list"
	"fmt"
)

// Component to keep track of matchable runes
type RuneMatcher struct {
	stack    *list.List    // stack for keeping record of matched characters
	matchmap map[rune]rune // map to record character match mappings
}

func NewRuneMatcher() *RuneMatcher {
	return &RuneMatcher{
		stack:    list.New(),
		matchmap: make(map[rune]rune),
	}
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

// Match the rune against the rune at TOS of matcher's stack
func (m *RuneMatcher) Match(r rune) bool {
	tos := m.stack.Back()
	R, registered := m.matchmap[r]
	return tos != nil && registered &&
		r == R && m.stack.Remove(tos) != nil
}

// RuneMatcher is in matched state if its stack is completely empty
func (m *RuneMatcher) IsMatched() bool { return m.stack.Back() == nil }

type Path rune

const DefaultPath Path = 0

type LexStateType int

const (
	Buffer  LexStateType = 0
	Feeder  LexStateType = 1
	Storer  LexStateType = 2
	Matcher LexStateType = 3
)

func (l LexStateType) String() string {
	stateNames := []string{"Buffer",
		"Feeder", "Storer", "Matcher"}
	return stateNames[int(l)]
}

// Individual state in a pushdown automata
type LexState struct {
	assocTokenType TokenType
	stateType      LexStateType
	emitter        bool
	next           map[Path]*LexState
}

func NewLexState(tt TokenType, st LexStateType, isEmitter bool) *LexState {
	return &LexState{tt, st, isEmitter, make(map[Path]*LexState)}
}

func (l *LexState) AddDest(p Path, dest *LexState) { l.next[p] = dest }

func (state LexState) String() string {
	return fmt.Sprintf("{%v}:[%v]",
		state.assocTokenType, state.stateType)
}

type Bounds struct {
	start, end *LexState
}

// (PDA) Represents a lexing machine attached to the "tape reader"
type Machine struct {
	lexer        *Lexer       // lexer for reading: tape reader
	bounds       Bounds       // bounds of machine to ensure finiteness
	tokens       chan Token   // channel to deliver tokens
	currState    *LexState    // current state of Lexing machine
	lastReadRune rune         // path leading to current state
	matcher      *RuneMatcher // stack associated with states of this machine
}

func NewMachine(lexer *Lexer, bounds Bounds, matcher *RuneMatcher) *Machine {
	return &Machine{lexer, bounds, make(chan Token, 2), bounds.start, 0, matcher}
}

// Resets all the components of this machine.
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
	m.matcher.stack.Init()
}

// Attaches new lexer to this machine and resets it.
func (m *Machine) AttachLexer(l *Lexer) { m.lexer = l; m.Reset() }

func (m *Machine) CanStep() bool { return m.currState != m.bounds.end }

// States whether this machine has reached the finishing state.
func (m *Machine) Finished() bool { return m.currState == m.bounds.end }

func (m *Machine) Step() error {

	fmt.Printf("currState: %v ", m.currState)

	path := Path(m.lexer.Next())

	var nextState *LexState
	var pathExists bool

	if nextState, pathExists = m.currState.next[path]; !pathExists {
		nextState, pathExists = m.currState.next[DefaultPath]
	}

	if !pathExists || nextState == nil {
		return fmt.Errorf("Invalid state: %v, no path with input: %v",
			m.currState, rune(path))
	}

	fmt.Printf("path: %v\n", path)

	switch m.currState.stateType {
	case Buffer:
		m.lexer.Backup()
	case Storer:
		if !m.matcher.Store(rune(path)) {
			return fmt.Errorf("unregistered rune: %v\n", path)
		}
	case Matcher:
		if !m.matcher.Match(m.lastReadRune) {
			return fmt.Errorf("unmatched rune: %v\n", path)
		}
	}

	if m.currState.emitter {
		if m.currState.stateType == Buffer {
			return fmt.Errorf("unsuitable state: %v for emitting token.\n",
				m.currState)
		} else {
			m.tokens <- m.lexer.Emit(nextState.assocTokenType)
		}
	}

	m.currState, m.lastReadRune = nextState, rune(path)

	return nil

}

func (m *Machine) NextToken() (Token, bool) {
	var err error = nil
	for err != nil {
		select {
		case token := <-m.tokens:
			return token, m.CanStep()
		default:
			err = m.Step()
		}
	}

	panic("Not reached")
}

// func (m *Machine) Run() chan Token {
// 	 go func() {
//	 	 var err error = nil
//		 for ; err == nil &&
//			 m.CanStep(); err = m.Step() {
//		 }
//	 }()
//
//	 return m.Tokens
// }
