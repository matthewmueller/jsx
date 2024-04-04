package jsx_test

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/matthewmueller/diff"
	"github.com/matthewmueller/jsx"
	"github.com/matthewmueller/jsx/ast"
)

func ExampleParse() {
	input := `export default () => <h1>hello world</h1>`
	ast, err := jsx.Parse("input.jsx", input)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(ast.String())
	// Output:
	// export default () => <h1>hello world</h1>
}

type Printer struct {
	s strings.Builder
}

var _ ast.Visitor = (*Printer)(nil)

func (p *Printer) VisitScript(s *ast.Script) {
	for _, fragment := range s.Body {
		fragment.Visit(p)
	}
}

func (p *Printer) VisitText(t *ast.Text) {
	p.s.WriteString(t.Value)
}

func (p *Printer) VisitComment(c *ast.Comment) {
	p.s.WriteString(c.String())
}

func (p *Printer) VisitField(f *ast.Field) {
	p.s.WriteString(f.Name)
	p.s.WriteString("=")
	f.Value.Visit(p)
}

func (p *Printer) VisitStringValue(s *ast.StringValue) {
	p.s.WriteString(s.Value)
}

func (p *Printer) VisitExpr(e *ast.Expr) {
	p.s.WriteString("{")
	for _, frag := range e.Fragments {
		frag.Visit(p)
	}
	p.s.WriteString("}")
}

func (p *Printer) VisitBoolValue(b *ast.BoolValue) {
	p.s.WriteString(strconv.Quote(strconv.FormatBool(b.Value)))
}

func (p *Printer) VisitElement(e *ast.Element) {
	p.s.WriteString("<")
	p.s.WriteString(e.Name)
	if len(e.Attrs) > 0 {
		p.s.WriteString(" ")
		for i, attr := range e.Attrs {
			if i > 0 {
				p.s.WriteString(" ")
			}
			attr.Visit(p)
		}
	}
	p.s.WriteString(">")
	for _, child := range e.Children {
		child.Visit(p)
	}
	p.s.WriteString("</")
	p.s.WriteString(e.Name)
	p.s.WriteString(">")
}

func (p *Printer) String() string {
	return p.s.String()
}

func TestVisit(t *testing.T) {
	input := `export default () => <style scoped>{"body { background: blue }"}</style>`
	script, err := jsx.Parse("input.jsx", input)
	if err != nil {
		t.Fatal(err)
	}
	printer := &Printer{}
	script.Visit(printer)
	actual := printer.String()
	expect := `export default () => <style scoped="true">{"body { background: blue }"}</style>`
	diff.TestString(t, actual, expect)
}
