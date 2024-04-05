package jsx

import (
	"fmt"

	"github.com/matthewmueller/jsx/ast"
	"github.com/matthewmueller/jsx/internal/lexer"
	"github.com/matthewmueller/jsx/internal/parser"
)

// Parse a .jsx or .tsx file and return an AST
func Parse(path, input string) (*ast.Script, error) {
	l := lexer.New(input)
	p := parser.New(path, l)
	ast, err := p.Parse()
	if err != nil {
		return nil, fmt.Errorf("jsx: %w", err)
	}
	return ast, nil
}
