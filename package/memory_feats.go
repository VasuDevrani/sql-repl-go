package sqlgo

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