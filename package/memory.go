package pck

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

type ColumnType uint

const (
	TextType ColumnType = iota
	IntType
	BoolType
)

type Cell interface {
	AsText() string
	AsInt() int32
	AsBool() bool
}

type Results struct {
	Columns ResultColumns
	Rows    [][]Cell
}

type ResultColumns []ResultColumn

type ResultColumn struct {
	Type ColumnType
	Name string
}

type Backend interface {
	CreateTable(*CreateTableStatement) error
	Insert(*InsertStatement) error
	Select(*SelectStatement) (*Results, error)
}

type MemoryCell []byte

func (mc MemoryCell) AsInt() int32 {
	var i int32
	err := binary.Read(bytes.NewBuffer(mc), binary.BigEndian, &i)
	if err != nil {
		fmt.Printf("Corrupted data [%s]: %s\n", mc, err)
		return 0
	}

	return i
}

func (mc MemoryCell) AsText() string {
	return string(mc)
}

func (mc MemoryCell) AsBool() bool {
	return len(mc) != 0
}

func (mc MemoryCell) equals(b MemoryCell) bool {
	// Seems verbose but need to make sure if one is nil, the
	// comparison still fails quickly
	if mc == nil || b == nil {
		return mc == nil && b == nil
	}

	return bytes.Compare(mc, b) == 0
}

var (
	trueToken  = Token{kind: boolKind, value: "true"}
	falseToken = Token{kind: boolKind, value: "false"}

	trueMemoryCell  = literalToMemoryCell(&trueToken)
	falseMemoryCell = literalToMemoryCell(&falseToken)
)

type table struct {
	columns     []string
	columnTypes []ColumnType
	rows        [][]MemoryCell
}

type MemoryBackend struct {
	tables map[string]*table
}

func NewMemoryBackend() *MemoryBackend {
	return &MemoryBackend{
		tables: map[string]*table{},
	}
}
