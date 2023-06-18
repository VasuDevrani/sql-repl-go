package pck

import (
	"io"
	"fmt"
	"os"
	"strings"

	"github.com/chzyer/readline"
	"github.com/olekukonko/tablewriter"
)

func doSelect(mb Backend, slct *SelectStatement) error {
	results, err := mb.Select(slct)
	if err != nil {
		return err
	}

	if len(results.Rows) == 0 {
		fmt.Println("(no results)")
		return nil
	}

	table := tablewriter.NewWriter(os.Stdout)
	header := []string{}
	for _, col := range results.Columns {
		header = append(header, col.Name)
	}
	table.SetHeader(header)
	table.SetAutoFormatHeaders(false)

	rows := [][]string{}
	for _, result := range results.Rows {
		row := []string{}
		for i, cell := range result {
			typ := results.Columns[i].Type
			r := ""
			switch typ {
			case IntType:
				i := cell.AsInt()
				if &i != nil {
					r = fmt.Sprintf("%d", i)
				}
			case TextType:
				s := cell.AsText()
				if &s != nil {
					r = s
				}
			case BoolType:
				b := cell.AsBool()
				if &b != nil {
					r = "t"
					if !b {
						r = "f"
					}
				}
			}

			row = append(row, r)
		}

		rows = append(rows, row)
	}

	table.SetBorder(false)
	table.AppendBulk(rows)
	table.Render()

	if len(rows) == 1 {
		fmt.Println("(1 result)")
	} else {
		fmt.Printf("(%d results)\n", len(rows))
	}

	return nil
}

func RunRepl(b Backend) {
	l, err := readline.NewEx(&readline.Config{
		Prompt:          "# ",
		HistoryFile:     "/tmp/tmp",
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",
	})
	if err != nil {
		panic(err)
	}
	defer l.Close()

	fmt.Println("Welcome to gosql.")
repl:
	for {
		fmt.Print("# ")
		line, err := l.Readline()
		if err == readline.ErrInterrupt {
			if len(line) == 0 {
				break
			} else {
				continue repl
			}
		} else if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Println("Error while reading line:", err)
			continue repl
		}

		trimmed := strings.TrimSpace(line)
		if trimmed == "quit" || trimmed == "exit" || trimmed == "\\q" {
			break
		}
		
		ast, err := Parse(line)
		if err != nil {
			fmt.Println("Error while parsing:", err)
			continue repl
		}

		for _, stmt := range ast.Statements {
			switch stmt.Kind {
			case CreateTableKind:
				err = b.CreateTable(ast.Statements[0].CreateTableStatement)
				if err != nil {
					fmt.Println("Error creating table:", err)
					continue repl
				}
			case InsertKind:
				err = b.Insert(stmt.InsertStatement)
				if err != nil {
					fmt.Println("Error inserting values:", err)
					continue repl
				}
			case SelectKind:
				err := doSelect(b, stmt.SelectStatement)
				if err != nil {
					fmt.Println("Error selecting values:", err)
					continue repl
				}
			}
		}
		fmt.Println("ok")
	}
}