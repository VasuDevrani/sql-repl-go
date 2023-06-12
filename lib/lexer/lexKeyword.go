package gosql

func lexKeyword(source string, ic cursor) (*token, cursor, bool) {
    cur := ic
    keywords := []keyword{
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
    for _, k := range keywords {
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

    return &token{
        value: match,
        kind:  Kind,
        loc:   ic.loc,
    }, cur, true
}