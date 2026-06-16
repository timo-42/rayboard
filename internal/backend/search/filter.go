package search

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/timo-42/rayboard/internal/backend/authz"
)

type wherePart struct {
	SQL  string
	Args []any
}

type compiledFilter struct {
	Parts []wherePart
}

var ticketFilterFields = map[string]string{
	"project":          "p.key",
	"project_id":       "t.project_id",
	"key":              "t.key",
	"title":            "t.title",
	"status":           "t.status",
	"priority":         "t.priority",
	"type":             "t.type",
	"reporter_id":      "t.reporter_id",
	"assignee_id":      "t.assignee_id",
	"parent_ticket_id": "t.parent_ticket_id",
	"sprint_id":        "t.sprint_id",
	"component_id":     "t.component_id",
	"version_id":       "t.version_id",
}

func compileTicketFilter(input string, principal authz.Principal) (compiledFilter, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return compiledFilter{}, nil
	}
	if len(input) > maxFilterLength {
		return compiledFilter{}, fmt.Errorf("must be %d characters or fewer", maxFilterLength)
	}

	parser := newFilterParser(input, principal)
	return parser.parse()
}

type tokenKind int

const (
	tokenEOF tokenKind = iota
	tokenIdent
	tokenString
	tokenEqual
	tokenNotEqual
	tokenAnd
	tokenLParen
	tokenRParen
	tokenInvalid
)

type token struct {
	kind    tokenKind
	literal string
	message string
}

type filterParser struct {
	lexer     *filterLexer
	current   token
	principal authz.Principal
}

func newFilterParser(input string, principal authz.Principal) *filterParser {
	parser := &filterParser{
		lexer:     &filterLexer{input: input},
		principal: principal,
	}
	parser.advance()
	return parser
}

func (p *filterParser) parse() (compiledFilter, error) {
	var filter compiledFilter
	for {
		if p.current.kind == tokenEOF {
			if len(filter.Parts) == 0 {
				return compiledFilter{}, fmt.Errorf("expected filter expression")
			}
			return filter, nil
		}

		part, err := p.parseComparison()
		if err != nil {
			return compiledFilter{}, err
		}
		filter.Parts = append(filter.Parts, part)

		switch p.current.kind {
		case tokenEOF:
			return filter, nil
		case tokenAnd:
			p.advance()
		case tokenInvalid:
			return compiledFilter{}, fmt.Errorf("%s", p.current.message)
		default:
			return compiledFilter{}, fmt.Errorf("expected && between filter clauses")
		}
	}
}

func (p *filterParser) parseComparison() (wherePart, error) {
	fieldToken := p.current
	if fieldToken.kind == tokenInvalid {
		return wherePart{}, fmt.Errorf("%s", fieldToken.message)
	}
	if fieldToken.kind != tokenIdent {
		return wherePart{}, fmt.Errorf("expected field name")
	}
	field := strings.ToLower(fieldToken.literal)
	sqlField, ok := ticketFilterFields[field]
	if !ok {
		return wherePart{}, fmt.Errorf("unsupported field %q", fieldToken.literal)
	}
	p.advance()

	operator := ""
	switch p.current.kind {
	case tokenEqual:
		operator = "="
	case tokenNotEqual:
		operator = "!="
	case tokenInvalid:
		return wherePart{}, fmt.Errorf("%s", p.current.message)
	default:
		return wherePart{}, fmt.Errorf("expected == or !=")
	}
	p.advance()

	value, err := p.parseValue()
	if err != nil {
		return wherePart{}, err
	}
	value = normalizeFilterValue(field, value)

	return wherePart{
		SQL:  sqlField + " " + operator + " ?",
		Args: []any{value},
	}, nil
}

func (p *filterParser) parseValue() (string, error) {
	switch p.current.kind {
	case tokenString:
		value := p.current.literal
		p.advance()
		return value, nil
	case tokenIdent:
		name := p.current.literal
		p.advance()
		if p.current.kind != tokenLParen {
			return "", fmt.Errorf("expected string literal or approved function")
		}
		p.advance()
		if p.current.kind != tokenRParen {
			return "", fmt.Errorf("function %q does not accept arguments", name)
		}
		p.advance()
		switch name {
		case "currentUser":
			if p.principal.UserID == "" {
				return "", fmt.Errorf("currentUser() requires an authenticated user")
			}
			return p.principal.UserID, nil
		default:
			return "", fmt.Errorf("unsupported function %q", name)
		}
	case tokenInvalid:
		return "", fmt.Errorf("%s", p.current.message)
	default:
		return "", fmt.Errorf("expected string literal or approved function")
	}
}

func (p *filterParser) advance() {
	p.current = p.lexer.next()
}

type filterLexer struct {
	input string
	pos   int
}

func (l *filterLexer) next() token {
	l.skipSpace()
	if l.pos >= len(l.input) {
		return token{kind: tokenEOF}
	}

	switch {
	case strings.HasPrefix(l.input[l.pos:], "&&"):
		l.pos += 2
		return token{kind: tokenAnd, literal: "&&"}
	case strings.HasPrefix(l.input[l.pos:], "=="):
		l.pos += 2
		return token{kind: tokenEqual, literal: "=="}
	case strings.HasPrefix(l.input[l.pos:], "!="):
		l.pos += 2
		return token{kind: tokenNotEqual, literal: "!="}
	}

	ch := rune(l.input[l.pos])
	switch ch {
	case '(':
		l.pos++
		return token{kind: tokenLParen, literal: "("}
	case ')':
		l.pos++
		return token{kind: tokenRParen, literal: ")"}
	case '"':
		return l.scanString()
	}

	if isIdentStart(ch) {
		return l.scanIdent()
	}

	return token{kind: tokenInvalid, message: fmt.Sprintf("unexpected character %q", ch)}
}

func (l *filterLexer) skipSpace() {
	for l.pos < len(l.input) {
		r, size := runeAt(l.input, l.pos)
		if !unicode.IsSpace(r) {
			return
		}
		l.pos += size
	}
}

func (l *filterLexer) scanIdent() token {
	start := l.pos
	for l.pos < len(l.input) {
		r, size := runeAt(l.input, l.pos)
		if !isIdentPart(r) {
			break
		}
		l.pos += size
	}
	return token{kind: tokenIdent, literal: l.input[start:l.pos]}
}

func (l *filterLexer) scanString() token {
	start := l.pos
	l.pos++
	escaped := false
	for l.pos < len(l.input) {
		ch := l.input[l.pos]
		l.pos++
		if escaped {
			escaped = false
			continue
		}
		switch ch {
		case '\\':
			escaped = true
		case '"':
			value, err := strconv.Unquote(l.input[start:l.pos])
			if err != nil {
				return token{kind: tokenInvalid, message: "invalid string literal"}
			}
			return token{kind: tokenString, literal: value}
		}
	}
	return token{kind: tokenInvalid, message: "unterminated string literal"}
}

func isIdentStart(r rune) bool {
	return r == '_' || unicode.IsLetter(r)
}

func isIdentPart(r rune) bool {
	return r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r)
}

func runeAt(value string, pos int) (rune, int) {
	r, size := utf8.DecodeRuneInString(value[pos:])
	if r == utf8.RuneError && size == 0 {
		return rune(value[pos]), 1
	}
	return r, size
}

func normalizeFilterValue(field string, value string) string {
	value = strings.TrimSpace(value)
	switch field {
	case "project", "key":
		return strings.ToUpper(value)
	case "status", "priority", "type":
		return strings.ToLower(value)
	default:
		return value
	}
}
