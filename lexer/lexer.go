package lexer

import (
	"fmt"
	"io"
	"unicode"
	"unicode/utf8"
)

// rune to denote EOF
const EOF rune = 0

type TokenType interface{}

// Represents the atomic semantic entity in a text corpus
type Token struct {
	Lexeme string
	Type   TokenType
}

func (t Token) String() string {
	if len(t.Lexeme) > 10 {
		return fmt.Sprintf("%10q...", t.Lexeme)
	}

	return fmt.Sprintf("%q", t.Lexeme)
}

// Reperesents a Lexing machine reading runes from a Reader
type Lexer struct {
	reader io.Reader
	buffer []byte
	length int // number of bytes last read from input
	pos    int // current position in buffer
	start  int // start of current item being read
	width  int // width of the last read rune
}

// Ignores the last read rune
func (l *Lexer) Ignore() {
	l.start = l.pos
}

// Unreads the last read rune
func (l *Lexer) Backup() {
	l.pos -= l.width
}

// Peeks at the next rune in the buffer
func (l *Lexer) Peek() rune {
	r := l.Next()

	l.Backup()
	return r
}

// Reads runes from the Reader instance into a private buffer
func (l *Lexer) read() rune {
	l.pos, l.start, l.width = 0, 0, 0

	l.length, _ = l.reader.Read(l.buffer)

	if l.length == 0 {
		return EOF
	}

	return l.Next()
}

// Reads the next rune present in the private buffer
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

// Emits a token of the given TokenType
func (l *Lexer) Emit(t TokenType) Token {
	token := Token{
		Lexeme: string(l.buffer[l.start:l.pos]),
		Type:   t,
	}
	l.start = l.pos
	return token
}

// Create a new instance of Lexer from a Reader instance
// and the capacity of the buffer to be maintained in the lexer
func NewLexer(reader io.Reader, capacity int) *Lexer {
	l := &Lexer{
		reader: reader,
		buffer: make([]byte, capacity),
	}

	return l
}
