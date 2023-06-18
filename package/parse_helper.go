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

func parseToken(tokens []*Token, initialCursor uint, t Token) (*Token, uint, bool) {
	cursor := initialCursor

	if cursor >= uint(len(tokens)) {
		return nil, initialCursor, false
	}

	if p := tokens[cursor]; t.equals(p) {
		return p, cursor + 1, true
	}

	return nil, initialCursor, false
}

func parseTokenKind(tokens []*Token, initialCursor uint, kind tokenKind) (*Token, uint, bool) {
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

func parseLiteralExpression(tokens []*Token, initialCursor uint) (*expression, uint, bool) {
	cursor := initialCursor

	kinds := []tokenKind{identifierKind, numericKind, stringKind, boolKind, nullKind}
	for _, kind := range kinds {
		t, newCursor, ok := parseTokenKind(tokens, cursor, kind)
		if ok {
			return &expression{
				literal: t,
				kind:    literalKind,
			}, newCursor, true
		}
	}

	return nil, initialCursor, false
}

func parseExpressions(tokens []*Token, initialCursor uint, delimiter Token) (*[]*expression, uint, bool) {
	cursor := initialCursor

	var exps []*expression
	for {
		if cursor >= uint(len(tokens)) {
			return nil, initialCursor, false
		}

		current := tokens[cursor]
		if delimiter.equals(current) {
			break
		}

		if len(exps) > 0 {
			var ok bool
			_, cursor, ok = parseToken(tokens, cursor, tokenFromSymbol(commaSymbol))
			if !ok {
				helpMessage(tokens, cursor, "Expected comma")
				return nil, initialCursor, false
			}
		}

		exp, newCursor, ok := parseExpression(tokens, cursor, []Token{tokenFromSymbol(commaSymbol), tokenFromSymbol(rightparenSymbol)}, 0)
		if !ok {
			helpMessage(tokens, cursor, "Expected expression")
			return nil, initialCursor, false
		}
		cursor = newCursor

		exps = append(exps, exp)
	}

	return &exps, cursor, true
}

func parseExpression(tokens []*Token, initialCursor uint, delimiters []Token, minBp uint) (*expression, uint, bool) {
    cursor := initialCursor

    var exp *expression
    _, newCursor, ok := parseToken(tokens, cursor, tokenFromSymbol(leftparenSymbol))
    if ok {
        cursor = newCursor
        rightParenToken := tokenFromSymbol(rightparenSymbol)

        exp, cursor, ok = parseExpression(tokens, cursor, append(delimiters, rightParenToken), minBp)
        if !ok {
            helpMessage(tokens, cursor, "Expected expression after opening paren")
            return nil, initialCursor, false
        }

        _, cursor, ok = parseToken(tokens, cursor, rightParenToken)
        if !ok {
            helpMessage(tokens, cursor, "Expected closing paren")
            return nil, initialCursor, false
        }
    } else {
        exp, cursor, ok = parseLiteralExpression(tokens, cursor)
        if !ok {
            return nil, initialCursor, false
        }
    }

    lastCursor := cursor
outer:
    for cursor < uint(len(tokens)) {
        for _, d := range delimiters {
            _, _, ok = parseToken(tokens, cursor, d)
            if ok {
                break outer
            }
        }

        binOps := []Token{
            tokenFromKeyword(AndKeyword),
            tokenFromKeyword(OrKeyword),
            tokenFromSymbol(EqSymbol),
            tokenFromSymbol(NeqSymbol),
            tokenFromSymbol(ConcatSymbol),
            tokenFromSymbol(PlusSymbol),
        }

        var op *Token = nil
        for _, bo := range binOps {
            var t *Token
            t, cursor, ok = parseToken(tokens, cursor, bo)
            if ok {
                op = t
                break
            }
        }

        if op == nil {
            helpMessage(tokens, cursor, "Expected binary operator")
            return nil, initialCursor, false
        }

        bp := op.bindingPower()
        if bp < minBp {
            cursor = lastCursor
            break
        }

        b, newCursor, ok := parseExpression(tokens, cursor, delimiters, bp)
        if !ok {
            helpMessage(tokens, cursor, "Expected right operand")
            return nil, initialCursor, false
        }
        exp = &expression{
            binary: &binaryExpression{
                *exp,
                *b,
                *op,
            },
            kind: binaryKind,
        }
        cursor = newCursor
        lastCursor = cursor
    }

    return exp, cursor, true
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
        id, newCursor, ok := parseTokenKind(tokens, cursor, identifierKind)
        if !ok {
            helpMessage(tokens, cursor, "Expected column name")
            return nil, initialCursor, false
        }
        cursor = newCursor

        // Look for a column type
        ty, newCursor, ok := parseTokenKind(tokens, cursor, keywordKind)
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

func parseSelectItem(tokens []*Token, initialCursor uint, delimiters []Token) (*[]*SelectItem, uint, bool) {
	cursor := initialCursor

	var s []*SelectItem
outer:
	for {
		if cursor >= uint(len(tokens)) {
			return nil, initialCursor, false
		}

		current := tokens[cursor]
		for _, delimiter := range delimiters {
			if delimiter.equals(current) {
				break outer
			}
		}

		var ok bool
		if len(s) > 0 {
			_, cursor, ok = parseToken(tokens, cursor, tokenFromSymbol(commaSymbol))
			if !ok {
				helpMessage(tokens, cursor, "Expected comma")
				return nil, initialCursor, false
			}
		}

		var si SelectItem
		_, cursor, ok = parseToken(tokens, cursor, tokenFromSymbol(asteriskSymbol))
		if ok {
			si = SelectItem{Asterisk: true}
		} else {
			asToken := tokenFromKeyword(AsKeyword)
			delimiters := append(delimiters, tokenFromSymbol(commaSymbol), asToken)
			exp, newCursor, ok := parseExpression(tokens, cursor, delimiters, 0)
			if !ok {
				helpMessage(tokens, cursor, "Expected expression")
				return nil, initialCursor, false
			}

			cursor = newCursor
			si.Exp = exp

			_, cursor, ok = parseToken(tokens, cursor, asToken)
			if ok {
				id, newCursor, ok := parseTokenKind(tokens, cursor, identifierKind)
				if !ok {
					helpMessage(tokens, cursor, "Expected identifier after AS")
					return nil, initialCursor, false
				}

				cursor = newCursor
				si.As = id
			}
		}

		s = append(s, &si)
	}

	return &s, cursor, true
}