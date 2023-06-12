package gosql

func lexNumeric(source string, ic cursor) (*token, cursor, bool) {
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

    return &token{
        value: source[ic.pointer:cur.pointer],
        loc:   ic.loc,
        kind:  numericKind,
    }, cur, true
}