package sqlgo

func parseSelectStatement(tokens []*Token, initialCursor uint, delimiter Token) (*SelectStatement, uint, bool) {
	var ok bool
	cursor := initialCursor
	_, cursor, ok = parseToken(tokens, cursor, tokenFromKeyword(SelectKeyword))
	if !ok {
		return nil, initialCursor, false
	}

	slct := SelectStatement{}

	fromToken := tokenFromKeyword(FromKeyword)
	item, newCursor, ok := parseSelectItem(tokens, cursor, []Token{fromToken, delimiter})
	if !ok {
		return nil, initialCursor, false
	}

	slct.item = item
	cursor = newCursor

	whereToken := tokenFromKeyword(WhereKeyword)

	_, cursor, ok = parseToken(tokens, cursor, fromToken)
	if ok {
		from, newCursor, ok := parseTokenKind(tokens, cursor, identifierKind)
		if !ok {
			helpMessage(tokens, cursor, "Expected FROM item")
			return nil, initialCursor, false
		}

		slct.from = from
		cursor = newCursor
	}

	limitToken := tokenFromKeyword(LimitKeyword)
	offsetToken := tokenFromKeyword(OffsetKeyword)

	_, cursor, ok = parseToken(tokens, cursor, whereToken)
	if ok {
		where, newCursor, ok := parseExpression(tokens, cursor, []Token{limitToken, offsetToken, delimiter}, 0)
		if !ok {
			helpMessage(tokens, cursor, "Expected WHERE conditionals")
			return nil, initialCursor, false
		}

		slct.where = where
		cursor = newCursor
	}

	return &slct, cursor, true
}

func parseInsertStatement(tokens []*Token, initialCursor uint, delimiter Token) (*InsertStatement, uint, bool) {
	cursor := initialCursor

	// Look for INSERT
	if !expectToken(tokens, cursor, tokenFromKeyword(InsertKeyword)) {
		return nil, initialCursor, false
	}
	cursor++

	// Look for INTO
	if !expectToken(tokens, cursor, tokenFromKeyword(IntoKeyword)) {
		helpMessage(tokens, cursor, "Expected into")
		return nil, initialCursor, false
	}
	cursor++

	// Look for table name
	table, newCursor, ok := parseTokenKind(tokens, cursor, identifierKind)
	if !ok {
		helpMessage(tokens, cursor, "Expected table name")
		return nil, initialCursor, false
	}
	cursor = newCursor

	// Look for VALUES
	if !expectToken(tokens, cursor, tokenFromKeyword(ValuesKeyword)) {
		helpMessage(tokens, cursor, "Expected VALUES")
		return nil, initialCursor, false
	}
	cursor++

	// Look for left paren
	if !expectToken(tokens, cursor, tokenFromSymbol(leftparenSymbol)) {
		helpMessage(tokens, cursor, "Expected left paren")
		return nil, initialCursor, false
	}
	cursor++

	// Look for expression list
	values, newCursor, ok := parseExpressions(tokens, cursor, tokenFromSymbol(rightparenSymbol))
	if !ok {
		return nil, initialCursor, false
	}
	cursor = newCursor

	// Look for right paren
	if !expectToken(tokens, cursor, tokenFromSymbol(rightparenSymbol)) {
		helpMessage(tokens, cursor, "Expected right paren")
		return nil, initialCursor, false
	}
	cursor++

	return &InsertStatement{
		table:  *table,
		values: values,
	}, cursor, true
}

func parseCreateTableStatement(tokens []*Token, initialCursor uint, delimiter Token) (*CreateTableStatement, uint, bool) {
	cursor := initialCursor

	if !expectToken(tokens, cursor, tokenFromKeyword(CreateKeyword)) {
		return nil, initialCursor, false
	}
	cursor++

	if !expectToken(tokens, cursor, tokenFromKeyword(TableKeyword)) {
		return nil, initialCursor, false
	}
	cursor++

	name, newCursor, ok := parseTokenKind(tokens, cursor, identifierKind)
	if !ok {
		helpMessage(tokens, cursor, "Expected table name")
		return nil, initialCursor, false
	}
	cursor = newCursor

	if !expectToken(tokens, cursor, tokenFromSymbol(leftparenSymbol)) {
		helpMessage(tokens, cursor, "Expected left parenthesis")
		return nil, initialCursor, false
	}
	cursor++

	cols, newCursor, ok := parseColumnDefinitions(tokens, cursor, tokenFromSymbol(rightparenSymbol))
	if !ok {
		return nil, initialCursor, false
	}
	cursor = newCursor

	if !expectToken(tokens, cursor, tokenFromSymbol(rightparenSymbol)) {
		helpMessage(tokens, cursor, "Expected right parenthesis")
		return nil, initialCursor, false
	}
	cursor++

	return &CreateTableStatement{
		name: *name,
		cols: cols,
	}, cursor, true
}
