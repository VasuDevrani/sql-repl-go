package gosql

import (
	"fmt"
	"strings"
)

// location of the token in source code
type location struct {
	line uint
	col  uint
}

type keyword string

const (
	SelectKeyword     keyword = "select"
	FromKeyword       keyword = "from"
	AsKeyword         keyword = "as"
	TableKeyword      keyword = "table"
	CreateKeyword     keyword = "create"
	DropKeyword       keyword = "drop"
	InsertKeyword     keyword = "insert"
	IntoKeyword       keyword = "into"
	ValuesKeyword     keyword = "values"
	IntKeyword        keyword = "int"
	TextKeyword       keyword = "text"
	BoolKeyword       keyword = "boolean"
	WhereKeyword      keyword = "where"
	AndKeyword        keyword = "and"
	OrKeyword         keyword = "or"
	TrueKeyword       keyword = "true"
	FalseKeyword      keyword = "false"
	UniqueKeyword     keyword = "unique"
	IndexKeyword      keyword = "index"
	OnKeyword         keyword = "on"
	PrimarykeyKeyword keyword = "primary key"
	NullKeyword       keyword = "null"
	LimitKeyword      keyword = "limit"
	OffsetKeyword     keyword = "offset"
)

type symbol string

const (
    semicolonSymbol  symbol = ";"
    asteriskSymbol   symbol = "*"
    commaSymbol      symbol = ","
    leftparenSymbol  symbol = "("
    rightparenSymbol symbol = ")"
)

type tokenKind uint

const (
    keywordKind tokenKind = iota
	symbolKind
	identifierKind
	stringKind
	numericKind
	boolKind
	nullKind
)

type token struct {
    value string
    kind  tokenKind
    loc   location
}

type cursor struct {
    pointer uint
    loc     location
}

func (t *token) equals(other *token) bool {
    return t.value == other.value && t.kind == other.kind
}

type lexer func(string, cursor) (*token, cursor, bool)

func lex(source string) ([]*token, error) {
    tokens := []*token{}
    cur := cursor{}

lex:
    for cur.pointer < uint(len(source)) {
        lexers := []lexer{lexKeyword, lexSymbol, lexString, lexNumeric, lexIdentifier}
        for _, l := range lexers {
            if token, newCursor, ok := l(source, cur); ok {
                cur = newCursor

                // Omit nil tokens for valid, but empty syntax like newlines
                if token != nil {
                    tokens = append(tokens, token)
                }

                continue lex
            }
        }

        hint := ""
        if len(tokens) > 0 {
            hint = " after " + tokens[len(tokens)-1].value
        }
        return nil, fmt.Errorf("Unable to lex token%s, at %d:%d", hint, cur.loc.line, cur.loc.col)
    }

    return tokens, nil
}

// longestMatch iterates through a source string starting at the given
// cursor to find the longest matching substring among the provided
// options
func longestMatch(source string, ic cursor, options []string) string {
    var value []byte
    var skipList []int
    var match string

    cur := ic

    for cur.pointer < uint(len(source)) {

        value = append(value, strings.ToLower(string(source[cur.pointer]))...)
        cur.pointer++

    match:
        for i, option := range options {
            for _, skip := range skipList {
                if i == skip {
                    continue match
                }
            }

            // Deal with cases like INT vs INTO
            if option == string(value) {
                skipList = append(skipList, i)
                if len(option) > len(match) {
                    match = option
                }

                continue
            }

            sharesPrefix := string(value) == option[:cur.pointer-ic.pointer]
            tooLong := len(value) > len(option)
            if tooLong || !sharesPrefix {
                skipList = append(skipList, i)
            }
        }

        if len(skipList) == len(options) {
            break
        }
    }

    return match
}