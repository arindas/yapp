package parser

import (
	list "container/list"
	fmt "fmt"
	strings "strings"
	lexer "yapp/lexer"
)

// Element metadata
type ElementType struct {
	Name     string            // Name of the ekement
	TokTypes []lexer.TokenType // types of allowed tokens
	NTokens  int               // number of tokens managed
}

// Type to denote a grammatical element.
type Element struct {
	Type     ElementType    // type of the element
	Tokens   []*lexer.Token // tokens associated with this element
	Children *list.List     // list of sub elements
}

func NewElement(t ElementType) *Element {
	return &Element{
		Type:     t,
		Tokens:   make([]*lexer.Token, t.NTokens),
		Children: list.New(),
	}
}

func (elem *Element) String() string {
	type selem struct {
		elem   *Element
		indent string
	}
	var b strings.Builder

	stack := list.New()
	stack.PushBack(selem{elem, ""})

	for stack.Back() != nil {

		tos := stack.Back()
		stack.Remove(tos)
		selem_ := tos.Value.(selem)

		for child := selem_.elem.Children.Front(); child != nil; child = child.Next() {
			stack.PushBack(selem{child.Value.(*Element),
				strings.Join([]string{selem_.indent, "\t"}, "")})
		}
		b.WriteString(fmt.Sprintf("%q%q %v\n", selem_.indent,
			selem_.elem.Type.Name, selem_.elem.Tokens))
	}

	return b.String()
}
