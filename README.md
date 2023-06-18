# sql-repl-go
🚧 golang sql implementation with REPL

<img width="320" alt="image" src="https://github.com/VasuDevrani/sql-repl-go/assets/101383635/cca2b9e6-438d-43c6-b215-f93a591ba667">

## Current support:

- [x] REPL
- [x] Create table 
- [x] Insert into table
- [x] Select from table
- [x] binary expression and filters
- [x] database driver support
- [ ] Indexing

## Archiecture
- [cmd/main.go](https://github.com/VasuDevrani/sql-repl-go/blob/master/cmd/main.go) </br>
  Dataflow is: user input -> lexer -> parser -> in-memory backend
- [lexer.go](https://github.com/VasuDevrani/sql-repl-go/blob/master/package/lexer.go) </br>
  Tokenization using lexing functions to break the SQL queries into separate tokens for AST tree
- [parser.go](https://github.com/VasuDevrani/sql-repl-go/blob/master/package/parser.go) </br>
  Matches a list of tokens into an AST or fails if the user input is not a valid program
- [memory.go](https://github.com/VasuDevrani/sql-repl-go/blob/master/package/memory.go) </br>
  in-memory backend function for producing results
- [repl.go](https://github.com/VasuDevrani/sql-repl-go/blob/master/package/repl.go) </br>

## Basic driver usage
```go
package main

import (
    "database/sql"
    "fmt"

    _ "github.com/vasudevrani/sql-repl-go/package"
)

func main() {
    db, err := sql.Open("postgres", "")
    if err != nil {
        panic(err)
    }
    defer db.Close()

    _, err = db.Query("CREATE TABLE users (name TEXT, age INT);")
    if err != nil {
        panic(err)
    }

    _, err = db.Query("INSERT INTO users VALUES ('Terry', 45);")
    if err != nil {
        panic(err)
    }

    _, err = db.Query("INSERT INTO users VALUES ('Anette', 57);")
    if err != nil {
        panic(err)
    }

    rows, err := db.Query("SELECT name, age FROM users;")
    if err != nil {
        panic(err)
    }

    var name string
    var age uint64
    defer rows.Close()
    for rows.Next() {
        err := rows.Scan(&name, &age)
        if err != nil {
            panic(err)
        }

        fmt.Printf("Name: %s, Age: %d\n", name, age)
    }

    if err = rows.Err(); err != nil {
        panic(err)
    }
}
```
