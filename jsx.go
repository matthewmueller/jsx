package jsx

import (
	"github.com/matthewmueller/jsx/ast"
	"github.com/matthewmueller/jsx/internal/lexer"
	"github.com/matthewmueller/jsx/internal/parser"
)

// Parse a .jsx or .tsx file and return an AST
func Parse(path, input string) (*ast.Script, error) {
	l := lexer.New(input)
	p := parser.New(path, l)
	return p.Parse()
}
