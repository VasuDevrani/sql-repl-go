package sqlgo

type Ast struct {
    Statements []*Statement
}

type AstKind uint

const (
    SelectKind AstKind = iota
    CreateTableKind
    InsertKind
)

type Statement struct {
    SelectStatement      *SelectStatement
    CreateTableStatement *CreateTableStatement
    InsertStatement      *InsertStatement
    Kind                 AstKind
}

type InsertStatement struct {
    table  Token
    values *[]*expression
}

type expressionKind uint

const (
    literalKind expressionKind = iota
    binaryKind
)

type expression struct {
    literal *Token
    binary  *binaryExpression
    kind    expressionKind
}

type columnDefinition struct {
    name     Token
    datatype Token
    primaryKey bool
}

type CreateTableStatement struct {
    name Token
    cols *[]*columnDefinition
}

type SelectItem struct {
	Exp      *expression
	Asterisk  bool // for *
	As       *Token
}

type SelectStatement struct {
	item   *[]*SelectItem
	from   *Token
	where  *expression
}

type binaryExpression struct {
    a  expression
    b  expression
    op Token
}