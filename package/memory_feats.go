package sqlgo

import (
    "fmt"
)

func (mb *MemoryBackend) CreateTable(crt *CreateTableStatement) error {
    t := table{}
    mb.tables[crt.name.value] = &t
    if crt.cols == nil {

        return nil
    }

    for _, col := range *crt.cols {
        t.columns = append(t.columns, col.name.value)

        var dt ColumnType
        switch col.datatype.value {
        case "int":
            dt = IntType
        case "text":
            dt = TextType
        default:
            return ErrInvalidDatatype
        }

        t.columnTypes = append(t.columnTypes, dt)
    }

    return nil
}

func (mb *MemoryBackend) Insert(inst *InsertStatement) error {
    table, ok := mb.tables[inst.table.value]
    if !ok {
        return ErrTableDoesNotExist
    }

    if inst.values == nil {
        return nil
    }

    row := []MemoryCell{}

    if len(*inst.values) != len(table.columns) {
        return ErrMissingValues
    }

    for _, value := range *inst.values {
        if value.kind != literalKind {
            fmt.Println("Skipping non-literal.")
            continue
        }

        row = append(row, mb.tokenToCell(value.literal))
    }

    table.rows = append(table.rows, row)
    return nil
}

func (mb *MemoryBackend) Select(slct *SelectStatement) (*Results, error) {
    table, ok := mb.tables[slct.from.value]
    if !ok {
        return nil, ErrTableDoesNotExist
    }

    results := [][]Cell{}
    columns := []struct {
        Type ColumnType
        Name string
    }{}

    for i, row := range table.rows {
        result := []Cell{}
        isFirstRow := i == 0

        for _, exp := range slct.item {
            if exp.kind != literalKind {
                // Unsupported, doesn't currently exist, ignore.
                fmt.Println("Skipping non-literal expression.")
                continue
            }

            lit := exp.literal
            if lit.kind == identifierKind {
                found := false
                for i, tableCol := range table.columns {
                    if tableCol == lit.value {
                        if isFirstRow {
                            columns = append(columns, struct {
                                Type ColumnType
                                Name string
                            }{
                                Type: table.columnTypes[i],
                                Name: lit.value,
                            })
                        }

                        result = append(result, row[i])
                        found = true
                        break
                    }
                }

                if !found {
                    return nil, ErrColumnDoesNotExist
                }

                continue
            }

            return nil, ErrColumnDoesNotExist
        }

        results = append(results, result)
    }

    return &Results{
        Columns: columns,
        Rows:    results,
    }, nil
}