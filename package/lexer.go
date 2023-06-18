package pck

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
	EqSymbol         symbol = "="
	NeqSymbol        symbol = "<>"
	NeqSymbol2       symbol = "!="
	ConcatSymbol     symbol = "||"
	PlusSymbol       symbol = "+"
	LtSymbol         symbol = "<"
	LteSymbol        symbol = "<="
	GtSymbol         symbol = ">"
	GteSymbol        symbol = ">="
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

type Token struct {
	value string
	kind  tokenKind
	loc   location
}

type cursor struct {
	pointer uint
	loc     location
}

func (t *Token) equals(other *Token) bool {
	return t.value == other.value && t.kind == other.kind
}

func (t Token) bindingPower() uint {
	switch t.kind {
	case keywordKind:
		switch keyword(t.value) {
		case AndKeyword:
			fallthrough
		case OrKeyword:
			return 1
		}
	case symbolKind:
		switch symbol(t.value) {
		case EqSymbol:
			fallthrough
		case NeqSymbol:
			return 2

		case LtSymbol:
			fallthrough
		case GtSymbol:
			return 3

		case LteSymbol:
			fallthrough
		case GteSymbol:
			return 4

		case ConcatSymbol:
			fallthrough
		case PlusSymbol:
			return 5
		}
	}

	return 0
}

type lexer func(string, cursor) (*Token, cursor, bool)

func lex(source string) ([]*Token, error) {
	var tokens []*Token
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
		// for _, t := range tokens {
		// 	fmt.Println(t.value)
		// }
		if cur.pointer == (uint(len(source)) - 1) {
			break
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

func lexIdentifier(source string, ic cursor) (*Token, cursor, bool) {
	// Handle separately if is a double-quoted identifier
	if token, newCursor, ok := lexCharacterDelimited(source, ic, '"'); ok {
		// Overwrite from stringkind to identifierkind
		token.kind = identifierKind
		return token, newCursor, true
	}

	cur := ic

	c := source[cur.pointer]
	// Other characters count too, big ignoring non-ascii for now
	isAlphabetical := (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z')
	if !isAlphabetical {
		return nil, ic, false
	}
	cur.pointer++
	cur.loc.col++

	value := []byte{c}
	for ; cur.pointer < uint(len(source)); cur.pointer++ {
		c = source[cur.pointer]

		// Other characters count too, big ignoring non-ascii for now
		isAlphabetical := (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z')
		isNumeric := c >= '0' && c <= '9'
		if isAlphabetical || isNumeric || c == '$' || c == '_' {
			value = append(value, c)
			cur.loc.col++
			continue
		}

		break
	}

	return &Token{
		// Unquoted identifiers are case-insensitive
		value: strings.ToLower(string(value)),
		loc:   ic.loc,
		kind:  identifierKind,
	}, cur, true
}

func lexKeyword(source string, ic cursor) (*Token, cursor, bool) {
	cur := ic
	Keywords := []keyword{
		SelectKeyword,
		InsertKeyword,
		ValuesKeyword,
		TableKeyword,
		CreateKeyword,
		DropKeyword,
		WhereKeyword,
		FromKeyword,
		IntoKeyword,
		TextKeyword,
		BoolKeyword,
		IntKeyword,
		AndKeyword,
		OrKeyword,
		AsKeyword,
		TrueKeyword,
		FalseKeyword,
		UniqueKeyword,
		IndexKeyword,
		OnKeyword,
		PrimarykeyKeyword,
		NullKeyword,
		LimitKeyword,
		OffsetKeyword,
	}

	var options []string
	for _, k := range Keywords {
		options = append(options, string(k))
	}

	match := longestMatch(source, ic, options)
	if match == "" {
		return nil, ic, false
	}

	cur.pointer = ic.pointer + uint(len(match))
	cur.loc.col = ic.loc.col + uint(len(match))

	Kind := keywordKind
	if match == string(TrueKeyword) || match == string(FalseKeyword) {
		Kind = boolKind
	}

	if match == string(NullKeyword) {
		Kind = nullKind
	}

	return &Token{
		value: match,
		kind:  Kind,
		loc:   ic.loc,
	}, cur, true
}

func lexNumeric(source string, ic cursor) (*Token, cursor, bool) {
	cur := ic

	periodFound := false
	expMarkerFound := false

	for ; cur.pointer < uint(len(source)); cur.pointer++ {
		c := source[cur.pointer]
		cur.loc.col++

		isDigit := c >= '0' && c <= '9'
		isPeriod := c == '.'
		isExpMarker := c == 'e'

		// Must start with a digit or period
		if cur.pointer == ic.pointer {
			if !isDigit && !isPeriod {
				return nil, ic, false
			}

			periodFound = isPeriod
			continue
		}

		if isPeriod {
			if periodFound {
				return nil, ic, false
			}

			periodFound = true
			continue
		}

		if isExpMarker {
			if expMarkerFound {
				return nil, ic, false
			}

			// No periods allowed after expMarker
			periodFound = true
			expMarkerFound = true

			// expMarker must be followed by digits
			if cur.pointer == uint(len(source)-1) {
				return nil, ic, false
			}

			cNext := source[cur.pointer+1]
			if cNext == '-' || cNext == '+' {
				cur.pointer++
				cur.loc.col++
			}
			continue
		}

		if !isDigit {
			break
		}
	}

	// No characters accumulated
	if cur.pointer == ic.pointer {
		return nil, ic, false
	}

	return &Token{
		value: source[ic.pointer:cur.pointer],
		loc:   ic.loc,
		kind:  numericKind,
	}, cur, true
}

func lexCharacterDelimited(source string, ic cursor, delimiter byte) (*Token, cursor, bool) {
	cur := ic

	if len(source[cur.pointer:]) == 0 {
		return nil, ic, false
	}

	if source[cur.pointer] != delimiter {
		return nil, ic, false
	}

	cur.loc.col++
	cur.pointer++

	var value []byte
	for ; cur.pointer < uint(len(source)); cur.pointer++ {
		c := source[cur.pointer]

		if c == delimiter {
			// SQL escapes are via double characters, not backslash.
			if cur.pointer+1 >= uint(len(source)) || source[cur.pointer+1] != delimiter {
				cur.pointer++
				cur.loc.col++
				return &Token{
					value: string(value),
					loc:   ic.loc,
					kind:  stringKind,
				}, cur, true
			}
			value = append(value, delimiter)
			cur.pointer++
			cur.loc.col++
		}

		value = append(value, c)
		cur.loc.col++
	}

	return nil, ic, false
}

func lexString(source string, ic cursor) (*Token, cursor, bool) {
	return lexCharacterDelimited(source, ic, '\'')
}

func lexSymbol(source string, ic cursor) (*Token, cursor, bool) {
	c := source[ic.pointer]
	cur := ic
	// Will get overwritten later if not an ignored syntax
	cur.pointer++
	cur.loc.col++

	switch c {
	// Syntax that should be thrown away
	case '\n':
		cur.loc.line++
		cur.loc.col = 0
		fallthrough
	case '\t':
		fallthrough
	case ' ':
		return nil, cur, true
	}

	// Syntax that should be kept
	Symbols := []symbol{
		EqSymbol,
		NeqSymbol,
		NeqSymbol2,
		LtSymbol,
		LteSymbol,
		GtSymbol,
		GteSymbol,
		ConcatSymbol,
		PlusSymbol,
		commaSymbol,
		leftparenSymbol,
		rightparenSymbol,
		semicolonSymbol,
		asteriskSymbol,
	}

	var options []string
	for _, s := range Symbols {
		options = append(options, string(s))
	}

	// Use `ic`, not `cur`
	match := longestMatch(source, ic, options)
	// Unknown character
	if match == "" {
		return nil, ic, false
	}

	cur.pointer = ic.pointer + uint(len(match))
	cur.loc.col = ic.loc.col + uint(len(match))

	if match == string(NeqSymbol2) {
		match = string(NeqSymbol)
	}

	return &Token{
		value: match,
		loc:   ic.loc,
		kind:  symbolKind,
	}, cur, true
}
