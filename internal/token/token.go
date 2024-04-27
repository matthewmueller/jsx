package token

import (
	"strconv"
	"strings"
)

type Type string

type Token struct {
	Type  Type
	Text  string
	Start int
	Line  int
}

func (t *Token) String() string {
	s := new(strings.Builder)
	s.WriteString(string(t.Type))
	if t.Text != "" && t.Text != string(t.Type) {
		s.WriteString(":")
		s.WriteString(strconv.Quote(t.Text))
	}
	return s.String()
}

const (
	EOF              Type = "eof"
	Error            Type = "error"
	LessThan         Type = "<"          // <
	GreaterThan      Type = ">"          // >
	Slash            Type = "/"          // /
	LessThanSlash    Type = "</"         // </
	SlashGreaterThan Type = "/>"         // />
	BackSlash        Type = "\\"         // \
	Identifier       Type = "identifier" // Any identifier
	Equal            Type = "="          // =
	Text             Type = "text"       // Raw text
	Expr             Type = "expr"       // { ... }
	OpenCurly        Type = "{"          // {
	CloseCurly       Type = "}"          // }
	String           Type = "string"     // "..." or '...'
	Space            Type = "space"      // Any whitespace

	Comment Type = "comment" // <!-- ... -->
)
