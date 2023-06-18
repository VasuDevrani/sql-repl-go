package pck

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
	t, ok := mb.tables[inst.table.value]
	if !ok {
		return ErrTableDoesNotExist
	}

	if inst.values == nil {
		return nil
	}

	row := []MemoryCell{}

	if len(*inst.values) != len(t.columns) {
		return ErrMissingValues
	}

	for _, value := range *inst.values {
		if value.kind != literalKind {
			fmt.Println("Skipping non-literal.")
			continue
		}

		emptyTable := &table{}
		value, _, _, err := emptyTable.evaluateCell(0, *value)
		if err != nil {
			return err
		}

		row = append(row, value)
	}

	t.rows = append(t.rows, row)
	return nil
}

func (mb *MemoryBackend) Select(slct *SelectStatement) (*Results, error) {
	t := &table{}

	if slct.from != nil {
		var ok bool
		t, ok = mb.tables[slct.from.value]
		if !ok {
			return nil, ErrTableDoesNotExist
		}
	}

	if slct.item == nil || len(*slct.item) == 0 {
		return &Results{}, nil
	}

	results := [][]Cell{}
	columns := []ResultColumn{}

	if slct.from == nil {
		t := &table{}
		t.rows = [][]MemoryCell{{}}
	}

	// Expand SELECT * at the AST level into a SELECT on all columns
	finalItems := []*SelectItem{}
	for _, item := range *slct.item {
		if item.Asterisk {
			newItems := []*SelectItem{}
			for j := 0; j < len(t.columns); j++ {
				newSelectItem := &SelectItem{
					Exp: &expression{
						literal: &Token{
							value: t.columns[j],
							kind:  identifierKind,
							loc:   location{0, uint(len("SELECT") + 1)},
						},
						binary: nil,
						kind:   literalKind,
					},
					Asterisk: false,
					As:       nil,
				}
				newItems = append(newItems, newSelectItem)
			}
			finalItems = append(finalItems, newItems...)
		} else {
			finalItems = append(finalItems, item)
		}
	}

	rowIndex := -1
	for i := range t.rows {
		result := []Cell{}
		isFirstRow := len(results) == 0

		if slct.where != nil {
			val, _, _, err := t.evaluateCell(uint(i), *slct.where)
			if err != nil {
				return nil, err
			}

			if !val.AsBool() {
				continue
			}
		}

		rowIndex++

		for _, col := range finalItems {
			value, columnName, columnType, err := t.evaluateCell(uint(i), *col.Exp)
			if err != nil {
				return nil, err
			}

			if isFirstRow {
				columns = append(columns, ResultColumn{
					Type: columnType,
					Name: columnName,
				})
			}

			result = append(result, value)
		}

		results = append(results, result)
	}

	var resultColumns []ResultColumn

	for _, col := range columns {
		resultColumns = append(resultColumns, ResultColumn{
			Type: col.Type,
			Name: col.Name,
		})
	}

	return &Results{
		Columns: resultColumns,
		Rows:    results,
	}, nil
}