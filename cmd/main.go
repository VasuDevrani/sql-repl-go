package main

import (
	pck "github.com/vasudevrani/sql-repl-go/package"
)

func main() {
	mb := pck.NewMemoryBackend()

	pck.RunRepl(mb)
}