package main

import (
	"errors"
	"strconv"
	"strings"
)

// Parser outputs

type StmtCreateTable struct {
	tableName string
	columns   []Column
}

type StmtSelect struct {
	columns   []string
	clauses   []Column
	tableName string
}

type StmtInsert struct {
	tableName string
	columns   []string
	values    []Value
}

type Column struct {
	name string
	kind ValueKind
}

type Clause struct {
	column string
	value  Value
}

type Value struct {
	kind ValueKind
	int  int
	text string
}

func (v Value) String() string {
	switch v.kind {
	case ValueKindInt:
		return strconv.Itoa(v.int)
	case ValueKindText:
		return "'" + string(v.text) + "'"
	default:
		return "UNKNOWN"
	}
}

//go:generate stringer -type=ValueKind
type ValueKind int

const (
	ValueKindInt ValueKind = iota
	ValueKindText
)

// Parser internals

func isWhitespace(c byte) bool {
	return c == ' ' || c == '\t' || c == '\n'
}

func isIdentCharStart(c byte) bool {
	return (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z')
}

func isIdentChar(c byte) bool {
	return (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || c == '_'
}

func isDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

type token struct {
	typ     tokenTyp
	literal string
}

//go:generate stringer -type=tokenTyp
type tokenTyp int

const (
	tokenTypUnknown tokenTyp = iota
	tokenTypIdent
	tokenTypInt
	tokenTypLParen
	tokenTypRParen
	tokenTypComma
	tokenTypString
	tokenTypSemicolon
	tokenTypEOF
)

type lexer struct {
	s   string
	pos int
	c   byte
}

func newLexer(s string) lexer {
	l := lexer{
		s:   s,
		pos: -1,
		c:   '\x00',
	}

	l.nextChar()

	return l
}

func (l *lexer) nextChar() {
	l.pos++

	if l.pos >= len(l.s) {
		l.c = '\x00'
		return
	}

	l.c = l.s[l.pos]
}

func (l *lexer) skipWhitespace() {
	for isWhitespace(l.c) {
		l.nextChar()
	}
}

func (l *lexer) isEnd() bool {
	return l.c == '\x00'
}

func (l *lexer) parseIdent() string {
	startPos := l.pos

	for isIdentChar(l.c) {
		l.nextChar()
	}

	endPos := l.pos

	return l.s[startPos:endPos]
}

func (l *lexer) parseNumber() string {
	startPos := l.pos

	for isDigit(l.c) {
		l.nextChar()
	}

	endPos := l.pos

	return l.s[startPos:endPos]
}

func (l *lexer) parseString() (string, error) {
	l.nextChar()
	startPos := l.pos

	for l.c != '\'' {
		if l.isEnd() {
			return "", errors.New("unterminated string")
		}
		l.nextChar()
	}

	endPos := l.pos

	return l.s[startPos:endPos], nil
}

func (l *lexer) nextToken() (token, error) {
	l.skipWhitespace()

	c := l.c
	if isIdentCharStart(c) {
		ident := l.parseIdent()
		return token{typ: tokenTypIdent, literal: ident}, nil
	}

	if isDigit(c) {
		number := l.parseNumber()
		return token{typ: tokenTypInt, literal: number}, nil
	}

	defer l.nextChar()

	switch c {
	case '\'':
		s, err := l.parseString()
		if err != nil {
			return token{}, err
		}
		return token{typ: tokenTypString, literal: s}, nil
	case '(':
		return token{typ: tokenTypLParen, literal: string(c)}, nil
	case ')':
		return token{typ: tokenTypRParen, literal: string(c)}, nil
	case ',':
		return token{typ: tokenTypComma, literal: string(c)}, nil
	case ';':
		return token{typ: tokenTypSemicolon, literal: string(c)}, nil
	case '\x00':
		return token{typ: tokenTypEOF, literal: ""}, nil
	default:
		return token{typ: tokenTypUnknown, literal: string(c)}, nil
	}
}

// lexAll lexes all tokens, effectively consuming the lexer.
func (l *lexer) lexAll() ([]token, error) {
	tokens := []token{}
	for {
		token, err := l.nextToken()
		if err != nil {
			return nil, err
		}
		if token.typ == tokenTypEOF {
			break
		}
		tokens = append(tokens, token)
	}
	return tokens, nil
}

type Parser struct {
	tokens      []token
	token       token
	pos         int
	lastLiteral string
}

func newParser(tokens []token) Parser {
	p := Parser{
		tokens: tokens,
		pos:    -1,
	}

	p.nextToken()

	return p
}

func (p *Parser) nextToken() {
	p.lastLiteral = p.token.literal
	p.pos++

	if p.pos >= len(p.tokens) {
		p.token = token{typ: tokenTypEOF}
		return
	}

	p.token = p.tokens[p.pos]
}

func (p *Parser) parse() (any, error) {
	switch {
	case p.eatKeyword("CREATE"):
		return p.parseStmtCreate()
	case p.eatKeyword("SELECT"):
		return p.parseStmtSelect()
	case p.eatKeyword("INSERT"):
		return p.parseStmtInsert()
	}
	return nil, nil
}

func (p *Parser) eatKeyword(keyword string) bool {
	ok := p.token.typ == tokenTypIdent && strings.EqualFold(p.token.literal, keyword)
	if ok {
		p.nextToken()
	}
	return ok
}

func (p *Parser) eatTokenTyp(typ tokenTyp) bool {
	ok := p.token.typ == typ
	if ok {
		p.nextToken()
	}
	return ok
}

func (p *Parser) expectValueKind() (ValueKind, error) {
	if p.token.typ != tokenTypIdent {
		return *new(ValueKind), errors.New("expected column type")
	}

	switch strings.ToUpper(p.token.literal) {
	case "INT":
		p.nextToken()
		return ValueKindInt, nil
	case "TEXT":
		p.nextToken()
		return ValueKindText, nil
	default:
		return *new(ValueKind), errors.New("expected column type")
	}
}

func (p *Parser) expectValue() (Value, error) {
	switch p.token.typ {
	case tokenTypInt:
		p.nextToken()
		int, err := strconv.Atoi(p.lastLiteral)
		if err != nil {
			return Value{}, err
		}
		return Value{kind: ValueKindInt, int: int}, nil
	case tokenTypString:
		p.nextToken()
		return Value{kind: ValueKindText, text: p.lastLiteral}, nil
	default:
		return Value{}, errors.New("expected value")
	}
}

func (p *Parser) parseStmtCreate() (StmtCreateTable, error) {
	if !p.eatKeyword("TABLE") {
		return StmtCreateTable{}, errors.New("expected keyword 'TABLE'")
	}

	if !p.eatTokenTyp(tokenTypIdent) {
		return StmtCreateTable{}, errors.New("expected identifier")
	}
	tableName := p.lastLiteral

	if !p.eatTokenTyp(tokenTypLParen) {
		return StmtCreateTable{}, errors.New("expected '('")
	}

	columns := []Column{}
	for {
		if !p.eatTokenTyp(tokenTypIdent) {
			return StmtCreateTable{}, errors.New("expected identifier")
		}
		name := p.lastLiteral

		kind, err := p.expectValueKind()
		if err != nil {
			return StmtCreateTable{}, err
		}

		column := Column{name: name, kind: kind}
		columns = append(columns, column)

		if p.eatTokenTyp(tokenTypComma) {
			continue
		} else if p.eatTokenTyp(tokenTypRParen) {
			break
		} else {
			return StmtCreateTable{}, errors.New("expected ',' or ')'")
		}
	}

	stmt := StmtCreateTable{
		tableName: tableName,
		columns:   columns,
	}

	return stmt, nil
}

func (p *Parser) parseStmtInsert() (StmtInsert, error) {
	if !p.eatKeyword("INTO") {
		return StmtInsert{}, errors.New("expected keyword 'INTO'")
	}

	if !p.eatTokenTyp(tokenTypIdent) {
		return StmtInsert{}, errors.New("expected identifier")
	}
	tableName := p.lastLiteral

	if !p.eatTokenTyp(tokenTypLParen) {
		return StmtInsert{}, errors.New("expected '('")
	}

	columns := []string{}
	for {
		if !p.eatTokenTyp(tokenTypIdent) {
			return StmtInsert{}, errors.New("expected identifier")
		}
		name := p.lastLiteral

		columns = append(columns, name)

		if p.eatTokenTyp(tokenTypComma) {
			continue
		} else if p.eatTokenTyp(tokenTypRParen) {
			break
		} else {
			return StmtInsert{}, errors.New("expected ',' or ')'")
		}
	}

	if !p.eatKeyword("VALUES") {
		return StmtInsert{}, errors.New("expected keyword 'VALUES'")
	}

	if !p.eatTokenTyp(tokenTypLParen) {
		return StmtInsert{}, errors.New("expected '('")
	}

	values := []Value{}
	for {
		value, err := p.expectValue()
		if err != nil {
			return StmtInsert{}, err
		}

		values = append(values, value)

		if p.eatTokenTyp(tokenTypComma) {
			continue
		} else if p.eatTokenTyp(tokenTypRParen) {
			break
		} else {
			return StmtInsert{}, errors.New("expected ',' or ')'")
		}
	}

	stmt := StmtInsert{
		tableName: tableName,
		columns:   columns,
		values:    values,
	}

	return stmt, nil
}

func (p *Parser) parseStmtSelect() (StmtSelect, error) {
	columns := []string{}
	for {
		if !p.eatTokenTyp(tokenTypIdent) {
			return StmtSelect{}, errors.New("expected identifier")
		}
		name := p.lastLiteral

		columns = append(columns, name)

		if p.eatTokenTyp(tokenTypComma) {
			continue
		} else if p.eatKeyword("FROM") {
			break
		} else {
			return StmtSelect{}, errors.New("expected ',' or 'FROM'")
		}
	}

	if !p.eatTokenTyp(tokenTypIdent) {
		return StmtSelect{}, errors.New("expected identifier")
	}
	tableName := p.lastLiteral

	stmt := StmtSelect{
		tableName: tableName,
		columns:   columns,
		// todo: handle clauses
		clauses: []Column{},
	}

	return stmt, nil
}
