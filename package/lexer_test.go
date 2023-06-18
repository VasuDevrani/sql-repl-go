package pck

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLexNumeric(t *testing.T) {
	tests := []struct {
		number bool
		value  string
	}{
		{
			number: true,
			value:  "105",
		},
		{
			number: true,
			value:  "105 ",
		},
		{
			number: true,
			value:  "123.",
		},
		// false tests
		{
			number: false,
			value:  "e4",
		},
		{
			number: false,
			value:  "1..",
		},
	}

	for _, test := range tests {
		tok, _, ok := lexNumeric(test.value, cursor{})
		assert.Equal(t, test.number, ok, test.value)
		if ok {
			assert.Equal(t, strings.TrimSpace(test.value), tok.value, test.value)
		}
	}
}

func TestLexString(t *testing.T) {
	tests := []struct {
		string bool
		value  string
	}{
		{
			string: false,
			value:  "a",
		},
		{
			string: true,
			value:  "'abc'",
		},
		// false tests
		{
			string: false,
			value:  "'",
		},
	}

	for _, test := range tests {
		tok, _, ok := lexString(test.value, cursor{})
		assert.Equal(t, test.string, ok, test.value)
		if ok {
			test.value = strings.TrimSpace(test.value)
			assert.Equal(t, test.value[1:len(test.value)-1], tok.value, test.value)
		}
	}
}

func TestLexSymbol(t *testing.T) {
	tests := []struct {
		symbol bool
		value  string
	}{
		{
			symbol: false,
			value:  "= ",
		},
		{
			symbol: false,
			value:  "||",
		},
	}

	for _, test := range tests {
		tok, _, ok := lexSymbol(test.value, cursor{})
		assert.Equal(t, test.symbol, ok, test.value)
		if ok {
			test.value = strings.TrimSpace(test.value)
			assert.Equal(t, test.value, tok.value, test.value)
		}
	}
}

func TestLexIdentifier(t *testing.T) {
	tests := []struct {
		Identifier bool
		input      string
		value      string
	}{
		{
			Identifier: true,
			input:      "a",
			value:      "a",
		},
		{
			Identifier: true,
			input:      "abc",
			value:      "abc",
		},
		{
			Identifier: true,
			input:      `"userName"`,
			value:      "userName",
		},
		// false tests
		{
			Identifier: false,
			input:      `"`,
		},
		{
			Identifier: false,
			input:      "_sadsfa",
		},
	}

	for _, test := range tests {
		tok, _, ok := lexIdentifier(test.input, cursor{})
		assert.Equal(t, test.Identifier, ok, test.input)
		if ok {
			assert.Equal(t, test.value, tok.value, test.input)
		}
	}
}

func TestLexKeyword(t *testing.T) {
	tests := []struct {
		keyword bool
		value   string
	}{
		{
			keyword: true,
			value:   "select ",
		},
		{
			keyword: true,
			value:   "from",
		},
		{
			keyword: true,
			value:   "SELECT",
		},
		{
			keyword: true,
			value:   "into",
		},
		// false tests
		{
			keyword: false,
			value:   " into",
		},
		{
			keyword: false,
			value:   "flubbrety",
		},
	}

	for _, test := range tests {
		tok, _, ok := lexKeyword(test.value, cursor{})
		assert.Equal(t, test.keyword, ok, test.value)
		if ok {
			test.value = strings.TrimSpace(test.value)
			assert.Equal(t, strings.ToLower(test.value), tok.value, test.value)
		}
	}
}

func TestLex(t *testing.T) {
	tests := []struct {
		input  string
		Tokens []Token
		err    error
	}{
		{
			input: "select a",
			Tokens: []Token{
				{
					loc:   location{col: 0, line: 0},
					value: string(SelectKeyword),
					kind:  keywordKind,
				},
				{
					loc:   location{col: 7, line: 0},
					value: "a",
					kind:  identifierKind,
				},
			},
		},
		{
			input: "insert into users Values (105, 233)",
			Tokens: []Token{
				{
					loc:   location{col: 0, line: 0},
					value: string(InsertKeyword),
					kind:  keywordKind,
				},
				{
					loc:   location{col: 7, line: 0},
					value: string(IntoKeyword),
					kind:  keywordKind,
				},
				{
					loc:   location{col: 12, line: 0},
					value: "users",
					kind:  identifierKind,
				},
				{
					loc:   location{col: 18, line: 0},
					value: string(ValuesKeyword),
					kind:  keywordKind,
				},
				{
					loc:   location{col: 25, line: 0},
					value: "(",
					kind:  symbolKind,
				},
				{
					loc:   location{col: 26, line: 0},
					value: "105",
					kind:  numericKind,
				},
				{
					loc:   location{col: 30, line: 0},
					value: ",",
					kind:  symbolKind,
				},
				{
					loc:   location{col: 32, line: 0},
					value: "233",
					kind:  numericKind,
				},
				{
					loc:   location{col: 36, line: 0},
					value: ")",
					kind:  symbolKind,
				},
			},
			err: nil,
		},
	}

	for _, test := range tests {
		tokens, err := lex(test.input)
		assert.Equal(t, test.err, err, test.input)
		assert.Equal(t, len(test.Tokens), len(tokens), test.input)

		for i, tok := range tokens {
			assert.Equal(t, &test.Tokens[i], tok, test.input)
		}
	}
}
