package parser

import (
	"fmt"
	"strings"

	"github.com/matthewmueller/jsx/ast"
	"github.com/matthewmueller/jsx/internal/lexer"
	"github.com/matthewmueller/jsx/internal/token"
)

func Parse(path, input string) (*ast.Script, error) {
	l := lexer.New(input)
	p := New(path, l)
	return p.Parse()
}

func Print(path, input string) string {
	doc, err := Parse(path, input)
	if err != nil {
		return err.Error()
	}
	return doc.String()
}

func New(path string, l *lexer.Lexer) *Parser {
	return &Parser{path, l}
}

type Parser struct {
	path string
	l    *lexer.Lexer
}

func (p *Parser) Parse() (*ast.Script, error) {
	return p.parseScript()
}

func (p *Parser) Text() string {
	return p.l.Token.Text
}

func (p *Parser) Type() token.Type {
	return p.l.Token.Type
}

// Checks that the next token is one of the given types
func (p *Parser) Is(types ...token.Type) bool {
	token := p.l.Peak(1)
	for _, t := range types {
		if token.Type == t {
			return true
		}
	}
	return false
}

// Returns true if all the given tokens are next
func (p *Parser) Check(tokens ...token.Type) bool {
	for i, token := range tokens {
		if p.l.Peak(i+1).Type != token {
			return false
		}
	}
	return true
}

func (p *Parser) More() bool {
	return p.Check(token.EOF)
}

// Returns true if all the given tokens are next
func (p *Parser) Accept(tokens ...token.Type) bool {
	if !p.Check(tokens...) {
		return false
	}
	for i := 0; i < len(tokens); i++ {
		p.l.Next()
	}
	return true
}

func (p *Parser) Expect(tokens ...token.Type) error {
	for i, tok := range tokens {
		peaked := p.l.Peak(i + 1)
		if peaked.Type == token.Error {
			return fmt.Errorf("expected %s, got %s (%d:%d)", tok, peaked.Text, peaked.Line, peaked.Start)
		} else if peaked.Type != tok {
			return fmt.Errorf("expected %s, got %s (%d:%d)", tok, peaked.Type, peaked.Line, peaked.Start)
		}
	}
	for i := 0; i < len(tokens); i++ {
		p.l.Next()
	}
	return nil
}

// TODO: this needs to be updated to better handle peaked tokens
func (p *Parser) unexpected(prefix string) error {
	token := p.l.Latest()
	return fmt.Errorf("%s unexpected token %s (%d:%d)", prefix, token.String(), token.Line, token.Start)
}

func (p *Parser) parseScript() (*ast.Script, error) {
	var body []ast.Fragment
	for !p.Accept(token.EOF) {
		fragment, err := p.parseFragment()
		if err != nil {
			return nil, err
		}
		body = append(body, fragment)
	}
	return &ast.Script{
		Body: body,
	}, nil
}

func (p *Parser) parseFragment() (ast.Fragment, error) {
	switch {
	case p.Accept(token.Text), p.Accept(token.Space):
		return p.parseText()
	case p.Accept(token.LessThan):
		return p.parseElement()
	case p.Accept(token.OpenCurly):
		return p.parseExpr()
	default:
		return nil, p.unexpected("fragment")
	}
}

func (p *Parser) parseText() (*ast.Text, error) {
	return &ast.Text{
		Value: p.Text(),
	}, nil
}

func (p *Parser) parseElement() (ast.Fragment, error) {
	for p.Accept(token.Space) {
	}
	// Sometimes < are false positives (typically when there are generics or less than signs)
	if p.Accept(token.Text) {
		return &ast.Text{
			Value: "<" + p.Text(),
		}, nil
	}
	for p.Accept(token.Space) {
	}
	if p.Accept(token.GreaterThan) {
		return p.parseJSXFragment()
	}
	if err := p.Expect(token.Identifier); err != nil {
		return nil, err
	}
	name := p.Text()
	// Sometimes < identifier is a false positive (typically when there are generics)
	if p.Accept(token.Text) {
		return &ast.Text{
			Value: "<" + name + p.Text(),
		}, nil
	}
	node := &ast.Element{
		Name: name,
	}
	for p.Accept(token.Space) {
	}
	// Handle attributes
	for !p.Check(token.SlashGreaterThan) && !p.Check(token.GreaterThan) {
		attr, err := p.parseAttr()
		if err != nil {
			return nil, err
		}
		node.Attrs = append(node.Attrs, attr)
		for p.Accept(token.Space) {
		}
	}
	if p.Accept(token.SlashGreaterThan) {
		node.SelfClosing = true
		return node, nil
	}
	if err := p.Expect(token.GreaterThan); err != nil {
		return nil, err
	}

	// Children
	for !p.Accept(token.LessThanSlash) {
		child, err := p.parseFragment()
		if err != nil {
			return nil, err
		}
		node.Children = append(node.Children, child)
	}

	// Closing tag
	if err := p.Expect(token.Identifier); err != nil {
		return nil, err
	} else if p.Text() != node.Name {
		token := p.l.Latest()
		return nil, fmt.Errorf("expected closing tag %s, got %s (%d:%d)", node.Name, p.Text(), token.Line, token.Start)
	}
	if err := p.Expect(token.GreaterThan); err != nil {
		return nil, err
	}
	return node, nil
}

func (p *Parser) parseJSXFragment() (*ast.Element, error) {
	node := &ast.Element{
		Name: "",
	}
	// Children
	for !p.Accept(token.LessThanSlash) {
		child, err := p.parseFragment()
		if err != nil {
			return nil, err
		}
		node.Children = append(node.Children, child)
	}
	// Closing tag
	if err := p.Expect(token.GreaterThan); err != nil {
		return nil, err
	}
	return node, nil
}

func (p *Parser) parseAttr() (ast.Attr, error) {
	switch {
	case p.Accept(token.Identifier):
		return p.parseField()
	case p.Accept(token.OpenCurly):
		return p.parseExpr()
	default:
		return nil, p.unexpected("attribute")
	}
}

func (p *Parser) parseField() (*ast.Field, error) {
	name := p.Text()
	for p.Accept(token.Space) {
	}
	// Handle boolean attributes
	if p.Is(token.Identifier, token.GreaterThan, token.SlashGreaterThan, token.OpenCurly) {
		return &ast.Field{
			Name: name,
			Value: &ast.BoolValue{
				Value: true,
			},
		}, nil
	}
	if err := p.Expect(token.Equal); err != nil {
		return nil, err
	}
	for p.Accept(token.Space) {
	}
	value, err := p.parseAttrValue()
	if err != nil {
		return nil, err
	}
	return &ast.Field{
		Name:  name,
		Value: value,
	}, nil
}

func (p *Parser) parseAttrValue() (ast.Value, error) {
	switch {
	case p.Accept(token.String):
		raw := p.Text()
		return &ast.StringValue{
			Value: unquote(raw),
			Raw:   raw,
		}, nil
	case p.Accept(token.OpenCurly):
		return p.parseExpr()
	default:
		return nil, p.unexpected("attr value")
	}
}

// Unquote the string. We can't use Go's strconv.Unquote because it doesn't
// handle single quotes and backslashes the same way that HTML does.
func unquote(s string) string {
	sl := len(s)
	if s[0] == '"' && s[sl-1] == '"' {
		s = s[1 : sl-1]
	} else if s[0] == '\'' && s[sl-1] == '\'' {
		s = s[1 : sl-1]
	}
	return s
}

func (p *Parser) parseExpr() (*ast.Expr, error) {
	var frags []ast.Fragment
	for {
		switch {
		case p.Accept(token.Expr):
			frags = append(frags, &ast.Text{
				Value: p.Text(),
			})
		case p.Accept(token.LessThan):
			element, err := p.parseElement()
			if err != nil {
				return nil, err
			}
			frags = append(frags, element)
		case p.Accept(token.Comment):
			frags = append(frags, &ast.Comment{
				Value: strings.Trim(p.Text(), "/*"),
			})
		case p.Accept(token.CloseCurly):
			return &ast.Expr{
				Fragments: frags,
			}, nil
		default:
			return nil, p.unexpected("expr")
		}
	}
}
