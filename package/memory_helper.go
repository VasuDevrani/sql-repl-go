package sqlgo

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"strconv"
)

func (mb *MemoryBackend) tokenToCell(t *Token) MemoryCell {
	if t.kind == numericKind {
		buf := new(bytes.Buffer)
		i, err := strconv.Atoi(t.value)
		if err != nil {
			panic(err)
		}

		err = binary.Write(buf, binary.BigEndian, int32(i))
		if err != nil {
			panic(err)
		}
		return MemoryCell(buf.Bytes())
	}

	if t.kind == stringKind {
		return MemoryCell(t.value)
	}

	return nil
}

func literalToMemoryCell(t *Token) MemoryCell {
    if t.kind == numericKind {
        buf := new(bytes.Buffer)
        i, err := strconv.Atoi(t.value)
        if err != nil {
            fmt.Printf("Corrupted data [%s]: %s\n", t.value, err)
            return MemoryCell(nil)
        }

        // TODO: handle bigint
        err = binary.Write(buf, binary.BigEndian, int32(i))
        if err != nil {
            fmt.Printf("Corrupted data [%s]: %s\n", string(buf.Bytes()), err)
            return MemoryCell(nil)
        }
		// type conversion MemoryCell alias for []byte
        return MemoryCell(buf.Bytes())
    }

    if t.kind == stringKind {
        return MemoryCell(t.value)
    }

    if t.kind == boolKind {
        if t.value == "true" {
            return MemoryCell([]byte{1})
        } else {
            return MemoryCell(nil)
        }
    }

    return nil
}

func (t *table) evaluateLiteralCell(rowIndex uint, exp expression) (MemoryCell, string, ColumnType, error) {
    if exp.kind != literalKind {
        return nil, "", 0, ErrInvalidCell
    }

    lit := exp.literal
    if lit.kind == identifierKind {
        for i, tableCol := range t.columns {
            if tableCol == lit.value {
                return t.rows[rowIndex][i], tableCol, t.columnTypes[i], nil
            }
        }

        return nil, "", 0, ErrColumnDoesNotExist
    }

    columnType := IntType
    if lit.kind == stringKind {
        columnType = TextType
    } else if lit.kind == boolKind {
        columnType = BoolType
    }

    return literalToMemoryCell(lit), "?column?", columnType, nil
}

func (t *table) evaluateBinaryCell(rowIndex uint, exp expression) (MemoryCell, string, ColumnType, error) {
    if exp.kind != binaryKind {
        return nil, "", 0, ErrInvalidCell
    }

    bexp := exp.binary

    l, _, lt, err := t.evaluateCell(rowIndex, bexp.a)
    if err != nil {
        return nil, "", 0, err
    }

    r, _, rt, err := t.evaluateCell(rowIndex, bexp.b)
    if err != nil {
        return nil, "", 0, err
    }

    switch bexp.op.kind {
    case symbolKind:
        switch symbol(bexp.op.value) {
        case EqSymbol:
            eq := l.equals(r)
            if lt == TextType && rt == TextType && eq {
                return trueMemoryCell, "?column?", BoolType, nil
            }

            if lt == IntType && rt == IntType && eq {
                return trueMemoryCell, "?column?", BoolType, nil
            }

            if lt == BoolType && rt == BoolType && eq {
                return trueMemoryCell, "?column?", BoolType, nil
            }

            return falseMemoryCell, "?column?", BoolType, nil
        case NeqSymbol:
            if lt != rt || !l.equals(r) {
                return trueMemoryCell, "?column?", BoolType, nil
            }

            return falseMemoryCell, "?column?", BoolType, nil
        case ConcatSymbol:
            if lt != TextType || rt != TextType {
                return nil, "", 0, ErrInvalidOperands
            }

            return literalToMemoryCell(&Token{kind: stringKind, value: l.AsText() + r.AsText()}), "?column?", TextType, nil
        case PlusSymbol:
            if lt != IntType || rt != IntType {
                return nil, "", 0, ErrInvalidOperands
            }

            iValue := int(l.AsInt() + r.AsInt())
            return literalToMemoryCell(&Token{kind: numericKind, value: strconv.Itoa(iValue)}), "?column?", IntType, nil
        default:
            // TODO
            break
        }
    case keywordKind:
        switch keyword(bexp.op.value) {
        case AndKeyword:
            if lt != BoolType || rt != BoolType {
                return nil, "", 0, ErrInvalidOperands
            }

            res := falseMemoryCell
            if l.AsBool() && r.AsBool() {
                res = trueMemoryCell
            }

            return res, "?column?", BoolType, nil
        case OrKeyword:
            if lt != BoolType || rt != BoolType {
                return nil, "", 0, ErrInvalidOperands
            }

            res := falseMemoryCell
            if l.AsBool() || r.AsBool() {
                res = trueMemoryCell
            }

            return res, "?column?", BoolType, nil
        default:
            // TODO
            break
        }
    }

    return nil, "", 0, ErrInvalidCell
}

func (t *table) evaluateCell(rowIndex uint, exp expression) (MemoryCell, string, ColumnType, error) {
    switch exp.kind {
        case literalKind:
            return t.evaluateLiteralCell(rowIndex, exp)
        case binaryKind:
            return t.evaluateBinaryCell(rowIndex, exp)
        default:
            return nil, "", 0, ErrInvalidCell
    }
}