package sqlgo

import (
	"errors"
	"fmt"
)

func tokenFromKeyword(k keyword) Token {
	return Token{
		kind:  keywordKind,
		value: string(k),
	}
}

func tokenFromSymbol(s symbol) Token {
	return Token{
		kind:  symbolKind,
		value: string(s),
	}
}

func expectToken(tokens []*Token, cursor uint, t Token) bool {
	if cursor >= uint(len(tokens)) {
		return false
	}

	return t.equals(tokens[cursor])
}

func helpMessage(tokens []*Token, cursor uint, msg string) {
	var c *Token
	if cursor < uint(len(tokens)) {
		c = tokens[cursor]
	} else {
		c = tokens[cursor-1]
	}

	fmt.Printf("[%d,%d]: %s, got: %s\n", c.loc.line, c.loc.col, msg, c.value)
}

func Parse(source string) (*Ast, error) {
	tokens, err := lex(source)
	if err != nil {
		return nil, err
	}

	a := Ast{}
	cursor := uint(0)
	for cursor < uint(len(tokens)) {
		stmt, newCursor, ok := parseStatement(tokens, cursor, tokenFromSymbol(semicolonSymbol))
		if !ok {
			helpMessage(tokens, cursor, "Expected statement")
			return nil, errors.New("Failed to parse, expected statement")
		}
		cursor = newCursor

		a.Statements = append(a.Statements, stmt)

		atLeastOneSemicolon := false
		for expectToken(tokens, cursor, tokenFromSymbol(semicolonSymbol)) {
			cursor++
			atLeastOneSemicolon = true
		}

		if !atLeastOneSemicolon {
			helpMessage(tokens, cursor, "Expected semi-colon delimiter between statements")
			return nil, errors.New("Missing semi-colon between statements")
		}
	}

	return &a, nil
}
