package lexer

import (
	"fmt"
	"strings"
	"unicode"
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

	inScript bool
	inStyle  bool
}

func (l *Lexer) nextToken() token.Token {
	l.start = l.end
	tokenType := l.states[len(l.states)-1](l)
	t := token.Token{
		Type:  tokenType,
		Start: l.start,
		Text:  l.input[l.start:l.end],
		Line:  l.line,
	}
	if tokenType == token.Error {
		t.Error = l.err
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
	if l.cp == '\n' {
		l.line++
	}
}

func (l *Lexer) accept(cp rune, run ...rune) bool {
	// Check the current rune
	if l.cp != cp {
		return false
	}
	str := l.peak(len(run))
	if len(str) != len(run) {
		return false
	}
	for i, r := range str {
		if r != run[i] {
			return false
		}
	}
	for i := 0; i < len(run)+1; i++ {
		l.step()
	}
	return true
}

func (l *Lexer) acceptFold(cp rune, run ...rune) bool {
	// Check the current rune
	if unicode.ToLower(l.cp) != unicode.ToLower(cp) {
		return false
	}
	str := l.peak(len(run))
	if len(str) != len(run) {
		return false
	}
	for i, r := range str {
		if unicode.ToLower(r) != unicode.ToLower(run[i]) {
			return false
		}
	}
	for i := 0; i < len(run)+1; i++ {
		l.step()
	}
	return true
}

func (l *Lexer) text() string {
	return l.input[l.start:l.end]
}

func (l *Lexer) stepUntil(rs ...rune) bool {
	for {
		if l.cp == eof {
			return false
		}
		for _, r := range rs {
			if l.cp == r {
				return true
			}
		}
		l.step()
	}
}

// func (l *Lexer) peak1() rune {
// 	cp, width := utf8.DecodeRuneInString(l.input[l.next:])
// 	if width == 0 {
// 		return eof
// 	}
// 	return cp
// }

func (l *Lexer) peak(n int) string {
	s := new(strings.Builder)
	if n == 0 {
		s.WriteRune(l.cp)
		return s.String()
	}
	next := l.next
	for i := 0; i < n; i++ {
		cp, width := utf8.DecodeRuneInString(l.input[next:])
		if width == 0 {
			cp = eof
			break
		}
		s.WriteRune(cp)
		next += width
	}
	return s.String()
}

func (l *Lexer) pushState(state state) {
	l.states = append(l.states, state)
}

func (l *Lexer) popState() {
	l.states = l.states[:len(l.states)-1]
}

func (l *Lexer) errorf(msg string, args ...interface{}) token.Type {
	l.err = fmt.Sprintf(msg, args...)
	return token.Error
}

func (l *Lexer) unexpected() token.Type {
	if l.cp == eof {
		return l.errorf("unexpected end of input")
	}
	return l.errorf("unexpected tokens '%s'", l.input[l.start:l.end])
}

func initialState(l *Lexer) (t token.Type) {
	for {
		switch l.cp {
		case eof:
			return token.EOF
		case '<':
			if l.prev == 0 || isSpace(l.prev) || l.prev == '(' {
				l.step()
				l.pushState(startOpenTagState)
				return token.LessThan
			}
			l.step()
			continue
		default:
			l.step()
			for l.cp != '<' && l.cp != eof {
				l.step()
			}
			return token.Text
		}
	}
}

func childTagState(l *Lexer) (t token.Type) {
	switch l.cp {
	case eof:
		return token.EOF
	case '<':
		l.step()
		switch {
		case l.cp == '/':
			l.step()
			l.pushState(startCloseTagState)
			return token.LessThanSlash
		case isAlpha(l.cp):
			l.pushState(startOpenTagState)
			return token.LessThan
		default:
			return l.unexpected()
		}
	case '{':
		l.step()
		return exprState(l)
	case ' ', '\t', '\n', '\r':
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
	case isAlpha(l.cp):
		l.step()
		for isAlphaNumeric(l.cp) || isDash(l.cp) {
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
	case l.cp == '/':
		l.step()
		if l.cp == '>' {
			l.step()
			l.popState()
			return token.SlashGreaterThan
		}
		return l.unexpected()
	case isAlpha(l.cp):
		l.step()
		for isAlphaNumeric(l.cp) || isDash(l.cp) {
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
		return exprState(l)
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
	case isAlpha(l.cp) || isDash(l.cp):
		l.step()
		for isAlphaNumeric(l.cp) || isDash(l.cp) {
			l.step()
		}
		return token.Identifier
	case l.cp == '>':
		l.step()
		l.popState()
		return token.GreaterThan
	default:
		l.step()
		return l.unexpected()
	}
}

func stringState(l *Lexer, end rune) (t token.Type) {
	for {
		switch {
		case l.cp == eof:
			l.popState()
			return l.unexpected()
		case l.cp == end:
			l.step()
			return token.String
		case l.cp == '\\':
			l.step()
			if l.cp == end {
				l.step()
			}
		case l.cp == '\n':
			return l.errorf("unexpected newline in string")
		default:
			l.step()
		}
	}
}

func exprState(l *Lexer) (t token.Type) {
	depth := 1
	for {
		switch {
		case l.cp == eof:
			l.popState()
			return l.unexpected()
		case l.cp == '{':
			depth++
			l.step()
		case l.cp == '}':
			l.step()
			depth--
			if depth == 0 {
				return token.Expr
			}
		default:
			l.step()
		}
	}
}

func isIdentifierHead(cp rune) bool {
	return isAlpha(cp) || cp == '_' || cp == '$'
}

func isAlpha(cp rune) bool {
	return (cp >= 'a' && cp <= 'z') || (cp >= 'A' && cp <= 'Z')
}

func isLower(cp rune) bool {
	return cp >= 'a' && cp <= 'z'
}

func isLowerNumeric(cp rune) bool {
	return isLower(cp) || (cp >= '0' && cp <= '9')
}

func isUpper(cp rune) bool {
	return cp >= 'A' && cp <= 'Z'
}

func isAlphaNumeric(cp rune) bool {
	return isAlpha(cp) || (cp >= '0' && cp <= '9')
}

func isDash(cp rune) bool {
	return cp == '-'
}

func isSpace(cp rune) bool {
	return cp == ' ' || cp == '\t' || cp == '\n' || cp == '\r'
}
