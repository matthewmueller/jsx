package ast

import (
	"strconv"
	"strings"
)

type Visitor interface {
	VisitScript(*Script)
	VisitText(*Text)
	VisitField(*Field)
	VisitStringValue(*StringValue)
	VisitExpr(*Expr)
	VisitBoolValue(*BoolValue)
	VisitElement(*Element)
	VisitComment(*Comment)
}

type Node interface {
	String() string
	Visit(Visitor)
}

var (
	_ Node = (*Script)(nil)
)

type Script struct {
	Body []Fragment
}

func (s *Script) String() string {
	sb := new(strings.Builder)
	for _, d := range s.Body {
		sb.WriteString(d.String())
	}
	return sb.String()
}

func (s *Script) Visit(v Visitor) {
	v.VisitScript(s)
}

type Fragment interface {
	Node
	fragment()
}

var (
	_ Fragment = (*Text)(nil)
	_ Fragment = (*Comment)(nil)
	_ Fragment = (*Element)(nil)
	_ Fragment = (*Expr)(nil)
)

type Text struct {
	Value string
}

func (r *Text) fragment() {}

func (r *Text) String() string {
	return r.Value
}

func (r *Text) Visit(v Visitor) {
	v.VisitText(r)
}

type Comment struct {
	Value string
}

func (c *Comment) fragment() {}

func (c *Comment) String() string {
	return "/*" + c.Value + "*/"
}

func (c *Comment) Visit(v Visitor) {
	v.VisitComment(c)
}

type Attr interface {
	Node
	attr()
}

var (
	_ Attr = (*Field)(nil)
	_ Attr = (*Expr)(nil)
)

type Field struct {
	Name  string
	Value Value
}

var _ Attr = (*Field)(nil)

func (f *Field) attr() {}

func (f *Field) String() string {
	if b, ok := f.Value.(*BoolValue); ok {
		if b.Value {
			return f.Name
		}
		return ""
	}
	return f.Name + "=" + f.Value.String()
}

func (f *Field) Visit(v Visitor) {
	v.VisitField(f)
}

type Value interface {
	Node
	value()
}

var (
	_ Value = (*StringValue)(nil)
	_ Value = (*BoolValue)(nil)
	_ Value = (*Expr)(nil)
)

type StringValue struct {
	Value string
	Raw   string
}

func (s *StringValue) value() {}

func (s *StringValue) String() string {
	return s.Raw
}

func (s *StringValue) Visit(v Visitor) {
	v.VisitStringValue(s)
}

type Expr struct {
	Fragments []Fragment
}

func (e *Expr) attr()     {}
func (e *Expr) value()    {}
func (e *Expr) fragment() {}

func (e *Expr) String() string {
	sb := new(strings.Builder)
	sb.WriteString("{")
	for _, f := range e.Fragments {
		sb.WriteString(f.String())
	}
	sb.WriteString("}")
	return sb.String()
}

func (e *Expr) Visit(v Visitor) {
	v.VisitExpr(e)
}

type BoolValue struct {
	Value bool
}

func (b *BoolValue) value() {}

func (b *BoolValue) String() string {
	return strconv.FormatBool(b.Value)
}

func (b *BoolValue) Visit(v Visitor) {
	v.VisitBoolValue(b)
}

type Element struct {
	Name        string
	Attrs       []Attr
	Children    []Fragment
	SelfClosing bool
}

func (e *Element) fragment() {}

func (e *Element) Type() string { return "Element" }

func (e *Element) String() string {
	out := new(strings.Builder)
	out.WriteString("<")
	out.WriteString(e.Name)
	for _, attr := range e.Attrs {
		out.WriteString(" ")
		out.WriteString(attr.String())
	}
	if e.SelfClosing {
		out.WriteString(" />")
		return out.String()
	}
	out.WriteString(">")
	if len(e.Children) > 0 {
		for _, child := range e.Children {
			out.WriteString(child.String())
		}
	}
	out.WriteString("</")
	out.WriteString(e.Name)
	out.WriteString(">")
	return out.String()
}

func (e *Element) Visit(v Visitor) {
	v.VisitElement(e)
}
