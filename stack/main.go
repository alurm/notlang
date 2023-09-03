package main

import (
	"os"

	"git.sr.ht/~alurm/notlang/stack/top"
)

func main() {
	mode := "values"
	if len(os.Args) >= 2 {
		mode = os.Args[1]
	}
	top.Shell(mode)
}
