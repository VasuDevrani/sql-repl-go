package gosql

func lexCharacterDelimited(source string, ic cursor, delimiter byte) (*token, cursor, bool) {
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
                return &token{
                    value: string(value),
                    loc:   ic.loc,
                    kind:  stringKind,
                }, cur, true
            } else {
                value = append(value, delimiter)
                cur.pointer++
                cur.loc.col++
            }
        }

        value = append(value, c)
        cur.loc.col++
    }

    return nil, ic, false
}

func lexString(source string, ic cursor) (*token, cursor, bool) {
    return lexCharacterDelimited(source, ic, '\'')
}