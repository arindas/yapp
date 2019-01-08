package lexer

import (
	// "io"
	// "fmt"
	"os"
	"testing"
	"unicode/utf8"
)

const (
	ifilePath = "test/test_input.txt"
	ofilePath = "test/test_output.txt"
)

func actionNecessary(err error, t *testing.T) bool {
	if err != nil {
		t.Errorf("%v", err)
		return true
	}
	return false
}

func GetTestableLexer(t *testing.T) (*Lexer, *os.File) {
	reader, err := os.Open(ifilePath)
	if actionNecessary(err, t) {
		return nil, nil
	}
	capacity := 256
	return NewLexer(reader, capacity), reader
}

func TestNewLexer(t *testing.T) {
	_, file := GetTestableLexer(t)
	actionNecessary(file.Close(), t)
}

func TestNext(t *testing.T) {
	lexr, ifile := GetTestableLexer(t)

	ofile, err := os.OpenFile(ofilePath,
		os.O_WRONLY|os.O_CREATE, 0755)

	if actionNecessary(err, t) {
		return
	}

	position, capacity := 0, 256
	buffer := make([]byte, capacity)

	for r := lexr.Next(); r != EOF; r = lexr.Next() {
		// t.Logf("%v", r)

		if position < capacity {
			width := utf8.EncodeRune(buffer[position:], r)
			position += width
		} else {
			position = 0
			_, err = ofile.Write(buffer)
			if actionNecessary(err, t) {
				break
			}
		}
	}

	if position < capacity {
		_, err = ofile.Write(buffer)
		actionNecessary(err, t)
	}

	actionNecessary(ifile.Close(), t)
	actionNecessary(ofile.Close(), t)
}

func TestEmit(t *testing.T) {
	lexr, ifile := GetTestableLexer(t)

	for r := lexr.Next(); r != EOF; r = lexr.Next() {
		if r == '\n' {
			token := lexr.Emit("line")
			t.Logf("%v\n", token)
		}
	}

	actionNecessary(ifile.Close(), t)
}
