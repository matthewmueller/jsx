package lexer

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/matthewmueller/jsx/internal/token"
)

type state = func(l *Lexer) token.Type

func New(input string) *Lexer {
	l := &Lexer{
		input:  input,
		states: []state{initialState},
		line:   1,
	}
	l.step()
	return l
}

func Lex(input string) []token.Token {
	l := New(input)
	var tokens []token.Token
	for l.Next() {
		tokens = append(tokens, l.Token)
	}
	return tokens
}

// Print the input as tokens
func Print(input string) string {
	tokens := Lex(input)
	stoken := make([]string, len(tokens))
	for i, token := range tokens {
		stoken[i] = token.String()
	}
	return strings.Join(stoken, " ")
}

type Lexer struct {
	Token token.Token // Current token
	input string      // Input string
	start int         // Index to the start of the current token
	end   int         // Index to the end of the current token
	cp    rune        // Code point being considered
	next  int         // Index to the next rune to be considered
	line  int         // Line number
	err   string      // Error message for an error token
	prev  rune        // Previous rune

	states []state // Stack of states
	peaked []token.Token
}

func (l *Lexer) nextToken() token.Token {
	l.start = l.end
	tokenType := l.states[len(l.states)-1](l)
	text := l.input[l.start:l.end]
	t := token.Token{
		Type:  tokenType,
		Start: l.start,
		Text:  text,
		Line:  l.line,
	}
	// update newlines
	for _, ch := range text {
		if ch == '\n' {
			l.line++
		}
	}
	if tokenType == token.Error {
		t.Text = l.err
		l.err = ""
	}
	return t
}

func (l *Lexer) Next() bool {
	if len(l.peaked) > 0 {
		l.Token = l.peaked[0]
		l.peaked = l.peaked[1:]
	} else {
		l.Token = l.nextToken()
	}
	return l.Token.Type != token.EOF
}

func (l *Lexer) Peak(nth int) token.Token {
	if len(l.peaked) >= nth {
		return l.peaked[nth-1]
	}
	for i := len(l.peaked); i < nth; i++ {
		l.peaked = append(l.peaked, l.nextToken())
	}
	return l.peaked[nth-1]
}

// TODO: replace with an errorf that creates a nice error message
func (l *Lexer) Latest() token.Token {
	if len(l.peaked) > 0 {
		return l.peaked[len(l.peaked)-1]
	}
	return l.Token
}

// Use -1 to indicate the end of the file
const eof = -1

// Step advances the lexer to the next token
func (l *Lexer) step() {
	codePoint, width := utf8.DecodeRuneInString(l.input[l.next:])
	if width == 0 {
		codePoint = eof
	}
	l.prev = l.cp
	l.cp = codePoint
	l.end = l.next
	l.next += width
}

func (l *Lexer) pushState(state state) {
	l.states = append(l.states, state)
}

func (l *Lexer) popState() {
	l.states = l.states[:len(l.states)-1]
}

func (l *Lexer) errorf(msg string, args ...any) token.Type {
	l.err = fmt.Sprintf(msg, args...)
	return token.Error
}

func (l *Lexer) unexpected() token.Type {
	if l.cp == eof {
		return l.errorf("unexpected end of input while lexing")
	}
	return l.errorf("unexpected token '%s' while lexing", l.input[l.start:l.end])
}

func initialState(l *Lexer) (t token.Type) {
	switch {
	case l.cp == eof:
		return token.EOF
	case l.cp == '<' && l.isTagStart():
		l.pushState(startOpenTagState)
		l.step()
		return token.LessThan
	default:
		for {
			if !l.skipJS() {
				l.step()
			}
			if l.cp == eof || (l.cp == '<' && l.isTagStart()) {
				break
			}
		}
		return token.Text
	}
}

func childTagState(l *Lexer) (t token.Type) {
	switch {
	case l.cp == eof:
		return token.EOF
	case l.cp == '<':
		l.step()
		switch {
		case l.cp == '/':
			l.step()
			l.popState()
			l.pushState(startCloseTagState)
			return token.LessThanSlash
		case isNameStart(l.cp) || isSpace(l.cp):
			l.pushState(startOpenTagState)
			return token.LessThan
		case l.cp == '>':
			l.pushState(startOpenTagState)
			return token.LessThan
		default:
			for {
				l.step()
				if l.cp == eof || l.cp == '<' || l.cp == '{' {
					break
				}
			}
			return token.Text
		}
	case l.cp == '{':
		l.step()
		l.pushState(expressionState())
		return token.OpenCurly
	default:
		for {
			l.step()
			if l.cp == eof || l.cp == '<' || l.cp == '{' {
				break
			}
		}
		return token.Text
	}
}

func startOpenTagState(l *Lexer) (t token.Type) {
	switch {
	case isSpace(l.cp):
		l.step()
		for isSpace(l.cp) {
			l.step()
		}
		return token.Space
	case l.cp == '>':
		l.step()
		l.popState()
		l.pushState(childTagState)
		return token.GreaterThan
	case isNameStart(l.cp):
		l.step()
		for isNamePart(l.cp) {
			l.step()
		}
		l.popState()
		l.pushState(middleTagState)
		return token.Identifier
	default:
		l.step()
		l.popState()
		return token.Text
	}
}

func middleTagState(l *Lexer) (t token.Type) {
	switch {
	case l.cp == eof:
		l.popState()
		return l.unexpected()
	case l.cp == '>':
		l.step()
		l.popState()
		l.pushState(childTagState)
		return token.GreaterThan
	case strings.HasPrefix(l.input[l.end:], "/*"):
		return commentBlockState(l)
	case l.cp == '/':
		l.step()
		if l.cp == '>' {
			l.step()
			l.popState()
			return token.SlashGreaterThan
		}
		// inline comment support
		if l.cp == '/' {
			l.step()
			return commentInlineState(l)
		}
		return l.unexpected()
	case isNameStart(l.cp):
		l.step()
		for isNamePart(l.cp) {
			l.step()
		}
		return token.Identifier
	case l.cp == '=':
		l.step()
		return token.Equal
	case l.cp == '"':
		l.step()
		return stringState(l, '"')
	case l.cp == '\'':
		l.step()
		return stringState(l, '\'')
	case l.cp == '{':
		l.step()
		l.pushState(expressionState())
		return token.OpenCurly
	case isSpace(l.cp):
		l.step()
		for isSpace(l.cp) {
			l.step()
		}
		return token.Space
	default:
		l.step()
		for l.cp != '<' && l.cp != eof {
			l.step()
		}
		l.popState()
		return token.Text
	}
}

func startCloseTagState(l *Lexer) (t token.Type) {
	switch {
	case isNameStart(l.cp) || isDash(l.cp):
		l.step()
		for isNamePart(l.cp) {
			l.step()
		}
		return token.Identifier
	case l.cp == '>':
		l.step()
		l.popState()
		return token.GreaterThan
	case isSpace(l.cp):
		l.step()
		for isSpace(l.cp) {
			l.step()
		}
		return token.Space
	default:
		l.step()
		return l.unexpected()
	}
}

// stringState lexes an attribute string. JSX attribute strings have no escapes
// and may span lines.
func stringState(l *Lexer, quote rune) (t token.Type) {
	for {
		switch {
		case l.cp == eof:
			l.popState()
			return l.unexpected()
		case l.cp == quote:
			l.step()
			return token.String
		default:
			l.step()
		}
	}
}

// skipJS skips over a JS string, template literal, comment or regex at the
// current position, so their contents aren't mistaken for JSX or braces.
func (l *Lexer) skipJS() bool {
	rest := l.input[l.end:]
	switch {
	case l.cp == '"' || l.cp == '\'':
		quote := l.cp
		l.step()
		for l.cp != quote && l.cp != eof && l.cp != '\n' {
			if l.cp == '\\' {
				l.step()
			}
			l.step()
		}
		l.step()
		return true
	case l.cp == '`':
		l.step()
		for l.cp != '`' && l.cp != eof {
			switch {
			case l.cp == '\\':
				l.step()
				l.step()
			case strings.HasPrefix(l.input[l.end:], "${"):
				l.step()
				l.skipBraces()
			default:
				l.step()
			}
		}
		l.step()
		return true
	case strings.HasPrefix(rest, "//"):
		for l.cp != '\n' && l.cp != eof {
			l.step()
		}
		return true
	case strings.HasPrefix(rest, "/*"):
		l.step()
		l.step()
		for l.cp != eof && !strings.HasPrefix(l.input[l.end:], "*/") {
			l.step()
		}
		l.step()
		l.step()
		return true
	case l.cp == '/' && l.isRegexStart():
		inClass := false
		l.step()
		for l.cp != eof && l.cp != '\n' && (inClass || l.cp != '/') {
			switch l.cp {
			case '\\':
				l.step()
			case '[':
				inClass = true
			case ']':
				inClass = false
			}
			l.step()
		}
		l.step()
		return true
	default:
		return false
	}
}

// skipBraces skips from the current '{' to its matching '}'
func (l *Lexer) skipBraces() {
	depth := 0
	for l.cp != eof {
		if l.cp == '{' {
			depth++
		} else if l.cp == '}' {
			depth--
			if depth == 0 {
				l.step()
				return
			}
		}
		if !l.skipJS() {
			l.step()
		}
	}
}

// isRegexStart reports whether the '/' at the current position starts a regex
// rather than a division, i.e. it follows an operator or the start of input.
func (l *Lexer) isRegexStart() bool {
	before := strings.TrimRightFunc(l.input[:l.end], isSpace)
	return before == "" || strings.ContainsRune("(,=:[!&|?{;", rune(before[len(before)-1]))
}

// expressionState returns a new state for lexing a {...} expression. The brace
// depth lives in the closure so it survives the expression being split into
// multiple tokens (e.g. by a nested element or a comment).
func expressionState() state {
	depth := 0
	return func(l *Lexer) (t token.Type) {
		nextRune, _ := utf8.DecodeRuneInString(l.input[l.next:])
		switch {
		case l.cp == eof:
			l.popState()
			return l.unexpected()
		case l.cp == '<' && isBeforeTag(l.prev) && (isNameStart(nextRune) || nextRune == '>'):
			l.pushState(startOpenTagState)
			l.step()
			return token.LessThan
		case l.cp == '}' && depth == 0:
			l.popState()
			l.step()
			return token.CloseCurly
		case strings.HasPrefix(l.input[l.end:], "/*"):
			return commentBlockState(l)
		default:
			for {
				if l.cp == '{' {
					depth++
				} else if l.cp == '}' {
					depth--
				}
				if !l.skipJS() {
					l.step()
				}
				if l.cp == eof || (l.cp == '<' && isBeforeTag(l.prev)) || (l.cp == '}' && depth == 0) || strings.HasPrefix(l.input[l.end:], "/*") {
					break
				}
			}
			return token.Expr
		}
	}
}

func isAlpha(cp rune) bool {
	return (cp >= 'a' && cp <= 'z') || (cp >= 'A' && cp <= 'Z')
}

func isAlphaNumeric(cp rune) bool {
	return isAlpha(cp) || (cp >= '0' && cp <= '9')
}

func isDash(cp rune) bool {
	return cp == '-'
}

// isNameStart reports whether cp can start a tag or attribute name
func isNameStart(cp rune) bool {
	return isAlpha(cp) || cp == '_' || cp == '$'
}

// isNamePart reports whether cp can continue a tag or attribute name
// (e.g. my-el, Foo.Bar, Icon_S, $Comp, svg:rect)
func isNamePart(cp rune) bool {
	return isAlphaNumeric(cp) || isDash(cp) || cp == '.' || cp == '_' || cp == '$' || cp == ':'
}

func isSpace(cp rune) bool {
	return cp == ' ' || cp == '\t' || cp == '\n' || cp == '\r'
}

// isBeforeTag reports whether a tag can start right after cp (e.g. `(<a>`,
// `[<a>`, `=><a>`, `&&<a>`)
func isBeforeTag(cp rune) bool {
	return cp == 0 || isSpace(cp) || strings.ContainsRune("({[,=>?:&|", cp)
}

// isTagStart reports whether the '<' at the current position starts a tag. A
// '<' with whitespace on both sides is a less than sign (e.g. `a < b`).
func (l *Lexer) isTagStart() bool {
	next, _ := utf8.DecodeRuneInString(l.input[l.next:])
	return isBeforeTag(l.prev) && !(isSpace(l.prev) && isSpace(next))
}

func commentBlockState(l *Lexer) token.Type {
	// Skip the opening /* so that /*/ isn't treated as closed
	l.step()
	l.step()
	for {
		switch {
		case l.cp == eof:
			return l.errorf("unclosed comment")
		case strings.HasPrefix(l.input[l.end:], "*/"):
			l.step()
			l.step()
			return token.Comment
		default:
			l.step()
		}
	}
}

func commentInlineState(l *Lexer) token.Type {
	for {
		switch {
		case l.cp == eof:
			return token.Comment
		case l.cp == '\n':
			l.step()
			return token.Comment
		default:
			l.step()
		}
	}
}
