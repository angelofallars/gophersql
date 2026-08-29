package main

import (
	"testing"

	"github.com/nalgeon/be"
)

func TestLexer_BasicStatements(t *testing.T) {
	f := func(s string, expected []token) {
		lexer := newLexer(s)
		tokens, err := lexer.lexAll()
		be.Err(t, err, nil)
		be.Equal(t, tokens, expected)
	}

	f(
		"() , ;",
		[]token{
			{typ: tokenTypLParen, literal: "("},
			{typ: tokenTypRParen, literal: ")"},
			{typ: tokenTypComma, literal: ","},
			{typ: tokenTypSemicolon, literal: ";"},
		},
	)

	f(
		"",
		[]token{},
	)

	f(
		"IDENT",
		[]token{
			{typ: tokenTypIdent, literal: "IDENT"},
		},
	)

	f(
		"CREATE TABLE users (id INT, name TEXT);",
		[]token{
			{typ: tokenTypIdent, literal: "CREATE"},
			{typ: tokenTypIdent, literal: "TABLE"},
			{typ: tokenTypIdent, literal: "users"},
			{typ: tokenTypLParen, literal: "("},
			{typ: tokenTypIdent, literal: "id"},
			{typ: tokenTypIdent, literal: "INT"},
			{typ: tokenTypComma, literal: ","},
			{typ: tokenTypIdent, literal: "name"},
			{typ: tokenTypIdent, literal: "TEXT"},
			{typ: tokenTypRParen, literal: ")"},
			{typ: tokenTypSemicolon, literal: ";"},
		},
	)

	f(
		"INSERT INTO users(id, name) VALUES (1, 'Angelo');",
		[]token{
			{typ: tokenTypIdent, literal: "INSERT"},
			{typ: tokenTypIdent, literal: "INTO"},
			{typ: tokenTypIdent, literal: "users"},
			{typ: tokenTypLParen, literal: "("},
			{typ: tokenTypIdent, literal: "id"},
			{typ: tokenTypComma, literal: ","},
			{typ: tokenTypIdent, literal: "name"},
			{typ: tokenTypRParen, literal: ")"},
			{typ: tokenTypIdent, literal: "VALUES"},
			{typ: tokenTypLParen, literal: "("},
			{typ: tokenTypInt, literal: "1"},
			{typ: tokenTypComma, literal: ","},
			{typ: tokenTypString, literal: "Angelo"},
			{typ: tokenTypRParen, literal: ")"},
			{typ: tokenTypSemicolon, literal: ";"},
		},
	)

	f(
		"SELECT id, name FROM users;",
		[]token{
			{typ: tokenTypIdent, literal: "SELECT"},
			{typ: tokenTypIdent, literal: "id"},
			{typ: tokenTypComma, literal: ","},
			{typ: tokenTypIdent, literal: "name"},
			{typ: tokenTypIdent, literal: "FROM"},
			{typ: tokenTypIdent, literal: "users"},
			{typ: tokenTypSemicolon, literal: ";"},
		},
	)
}

func TestParser_BasicStatements(t *testing.T) {
	f := func(s string, expected any) {
		lexer := newLexer(s)
		tokens, err := lexer.lexAll()
		be.Err(t, err, nil)

		parser := newParser(tokens)
		result, err := parser.parse()
		be.Err(t, err, nil)
		be.Equal(t, result, expected)
	}

	f(
		"CREATE TABLE users (id INT);",
		StmtCreateTable{
			tableName: "users",
			columns: []Column{
				{name: "id", kind: ValueKindInt},
			},
		},
	)

	f(
		"CREATE TABLE users (id INT, name TEXT);",
		StmtCreateTable{
			tableName: "users",
			columns: []Column{
				{name: "id", kind: ValueKindInt},
				{name: "name", kind: ValueKindText},
			},
		},
	)

	f(
		"INSERT INTO users(id, name) VALUES (1, 'Angelo');",
		StmtInsert{
			tableName: "users",
			columns:   []string{"id", "name"},
			values: []Value{
				{kind: ValueKindInt, int: 1},
				{kind: ValueKindText, text: "Angelo"},
			},
		},
	)

	f(
		"SELECT id, name FROM users;",
		StmtSelect{
			tableName: "users",
			columns:   []string{"id", "name"},
			clauses:   []Column{},
		},
	)
}
