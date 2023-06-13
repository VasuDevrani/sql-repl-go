package sqlgo

func parseSelectStatement(tokens []*Token, initialCursor uint, delimiter Token) (*SelectStatement, uint, bool) {
    cursor := initialCursor
    if !expectToken(tokens, cursor, tokenFromKeyword(SelectKeyword)) {
        return nil, initialCursor, false
    }
    cursor++

    slct := SelectStatement{}

    exps, newCursor, ok := parseExpressions(tokens, cursor, []Token{tokenFromKeyword(FromKeyword), delimiter})
    if !ok {
        return nil, initialCursor, false
    }

    slct.item = *exps
    cursor = newCursor

    if expectToken(tokens, cursor, tokenFromKeyword(FromKeyword)) {
        cursor++

        from, newCursor, ok := parseToken(tokens, cursor, identifierKind)
        if !ok {
            helpMessage(tokens, cursor, "Expected FROM token")
            return nil, initialCursor, false
        }

        slct.from = *from
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
    table, newCursor, ok := parseToken(tokens, cursor, identifierKind)
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
    values, newCursor, ok := parseExpressions(tokens, cursor, []Token{tokenFromSymbol(rightparenSymbol)})
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

    name, newCursor, ok := parseToken(tokens, cursor, identifierKind)
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