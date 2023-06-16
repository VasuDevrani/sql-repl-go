package sqlgo

func parseStatement(tokens []*Token, initialCursor uint, delimiter Token) (*Statement, uint, bool) {
    cursor := initialCursor

    // Look for a SELECT statement
    semicolonToken := tokenFromSymbol(semicolonSymbol)
    slct, newCursor, ok := parseSelectStatement(tokens, cursor, semicolonToken)
    if ok {
        return &Statement{
            Kind:            SelectKind,
            SelectStatement: slct,
        }, newCursor, true
    }

    // Look for a INSERT statement
    inst, newCursor, ok := parseInsertStatement(tokens, cursor, semicolonToken)
    if ok {
        return &Statement{
            Kind:            InsertKind,
            InsertStatement: inst,
        }, newCursor, true
    }

    // Look for a CREATE statement
    crtTbl, newCursor, ok := parseCreateTableStatement(tokens, cursor, semicolonToken)
    if ok {
        return &Statement{
            Kind:                 CreateTableKind,
            CreateTableStatement: crtTbl,
        }, newCursor, true
    }

    return nil, initialCursor, false
}

func parseToken(tokens []*Token, initialCursor uint, kind tokenKind) (*Token, uint, bool) {
    cursor := initialCursor

    if cursor >= uint(len(tokens)) {
        return nil, initialCursor, false
    }

    current := tokens[cursor]

    if current.kind == kind {
        return current, cursor + 1, true
    }

    return nil, initialCursor, false
}

func parseExpressions(tokens []*Token, initialCursor uint, delimiters []Token) (*[]*expression, uint, bool) {
    cursor := initialCursor

    exps := []*expression{}
outer:
    for {
        if cursor >= uint(len(tokens)) {
            break;
        }

        // Look for delimiter
        current := tokens[cursor]
        for _, delimiter := range delimiters {
            if delimiter.equals(current) {
                break outer
            }
        }

        // Look for comma
        if len(exps) > 0 {
            if !expectToken(tokens, cursor, tokenFromSymbol(commaSymbol)) {
                helpMessage(tokens, cursor, "Expected comma")
                return nil, initialCursor, false
            }

            cursor++
        }

        // Look for expression
        exp, newCursor, ok := parseExpression(tokens, cursor, tokenFromSymbol(commaSymbol))
        if !ok {
            helpMessage(tokens, cursor, "Expected expression")
            return nil, initialCursor, false
        }
        cursor = newCursor

        exps = append(exps, exp)
    }

    return &exps, cursor, true
}

func parseExpression(tokens []*Token, initialCursor uint, _ Token) (*expression, uint, bool) {
    cursor := initialCursor

    kinds := []tokenKind{identifierKind, numericKind, stringKind}
    for _, kind := range kinds {
        t, newCursor, ok := parseToken(tokens, cursor, kind)
        if ok {
            return &expression{
                literal: t,
                kind:    literalKind,
            }, newCursor, true
        }
    }

    return nil, initialCursor, false
}

func parseColumnDefinitions(tokens []*Token, initialCursor uint, delimiter Token) (*[]*columnDefinition, uint, bool) {
    cursor := initialCursor

    cds := []*columnDefinition{}
    for {
        if cursor >= uint(len(tokens)) {
            return nil, initialCursor, false
        }

        // Look for a delimiter
        current := tokens[cursor]
        if delimiter.equals(current) {
            break
        }

        // Look for a comma
        if len(cds) > 0 {
            if !expectToken(tokens, cursor, tokenFromSymbol(commaSymbol)) {
                helpMessage(tokens, cursor, "Expected comma")
                return nil, initialCursor, false
            }

            cursor++
        }

        // Look for a column name
        id, newCursor, ok := parseToken(tokens, cursor, identifierKind)
        if !ok {
            helpMessage(tokens, cursor, "Expected column name")
            return nil, initialCursor, false
        }
        cursor = newCursor

        // Look for a column type
        ty, newCursor, ok := parseToken(tokens, cursor, keywordKind)
        if !ok {
            helpMessage(tokens, cursor, "Expected column type")
            return nil, initialCursor, false
        }
        cursor = newCursor

        cds = append(cds, &columnDefinition{
            name:     *id,
            datatype: *ty,
        })
    }

    return &cds, cursor, true
}