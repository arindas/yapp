package lexer

import (
	"fmt"
	"io"
	"unicode"
	"unicode/utf8"
)

// EOF denotes the end of file
const EOF rune = -1

// TokenType denotes the type of Token
type TokenType interface{}

// Token represents the atomic semantic entity in a text corpus
type Token struct {
	Lexeme string
	Type   TokenType
}

// ErrorToken is a Token to denote error
var ErrorToken = Token{"error", 0}

// EOLexToken is Token to denote end of lexing
var EOLexToken = Token{"eolex", 0}

func (t Token) String() string {
	if len(t.Lexeme) > 10 {
		return fmt.Sprintf("%10q...", t.Lexeme)
	}

	return fmt.Sprintf("%q", t.Lexeme)
}

// Lexer reperesents a lexing machine reading runes from a Reader
type Lexer struct {
	reader io.Reader
	buffer []byte
	length int // number of bytes last read from input
	pos    int // current position in buffer
	start  int // start of current item being read
	width  int // width of the last read rune
}

// Ignore ignores the last read rune
func (l *Lexer) Ignore() {
	l.start = l.pos
}

// Backup unreads the last read rune
func (l *Lexer) Backup() {
	l.pos -= l.width
}

// Peek peeks at the next rune in the buffer
func (l *Lexer) Peek() rune {
	r := l.Next()

	l.Backup()
	return r
}

// Reset resets the lexer state
func (l *Lexer) Reset() { l.pos, l.start, l.width = 0, 0, 0 }

// Read reads runes from the Reader instance into a private buffer
func (l *Lexer) read() rune {
	l.Reset()

	l.length, _ = l.reader.Read(l.buffer)

	if l.length == 0 {
		return EOF
	}

	return l.Next()
}

// Next reads the next rune present in the private buffer
func (l *Lexer) Next() (r rune) {
	if l.pos >= l.length {
		return l.read()
	}

	r, l.width = utf8.DecodeRune(l.buffer[l.pos:])

	l.pos += l.width

	if unicode.IsSpace(r) && r != '\n' {
		l.Ignore()
		return l.Next()
	}

	return r
}

// Emit emits a token of the given TokenType
func (l *Lexer) Emit(t TokenType) Token {
	token := Token{
		Lexeme: string(l.buffer[l.start:l.pos]),
		Type:   t,
	}
	l.start = l.pos
	return token
}

// NewLexer creates a new instance of Lexer from a
// Reader instance with a given buffer capacity.
func NewLexer(reader io.Reader, capacity int) *Lexer {
	l := &Lexer{
		reader: reader,
		buffer: make([]byte, capacity),
	}

	return l
}
