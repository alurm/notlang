package top

import (
	"bufio"
	"fmt"
	"os"
	"strconv"

	"git.sr.ht/~alurm/notlang/v2/parse"
	"git.sr.ht/~alurm/notlang/v2/token"
	"git.sr.ht/~alurm/notlang/v2/value"
)

func Chan[T any](slice []T) chan T {
	out := make(chan T, len(slice))
	for _, v := range slice {
		out <- v
	}
	close(out)
	return out
}

func In(in string) chan byte {
	return Chan([]byte(in))
}

func Shell(mode string) {
	in := bufio.NewReader(os.Stdin)
	out := make(chan byte)
	tokens := token.Tokenize(out)
	go func() {
		for {
			b, err := in.ReadByte()
			if err != nil {
				close(out)
				break
			}
			out <- b
		}
	}()

	pipe := func(tokens chan token.Token, filters ...func(chan token.Token) chan token.Token) chan token.Token {
		for _, filter := range filters {
			tokens = filter(tokens)
		}
		return tokens
	}

	stages := []func(chan token.Token) chan token.Token{
		parse.GroupTop,
		parse.CommandTop,
		parse.DollarStringAsGetCommandGroup,
		parse.ApplicationTop,
		parse.SpaceTop,
	}
	min := func(i, j int) int {
		if i > j {
			return j
		}
		return i
	}
	if i, err := strconv.Atoi(mode); err == nil {
		stages = stages[:min(i, len(stages))]
	}
	tokens = pipe(
		tokens,
		stages...,
	)

	/*for t := range tokens {
		fmt.Printf("%#v\n", t)
	}*/

	if mode == "values" {
		tree := parse.Parse(tokens)

		values := value.Shell(tree)

		for v := range values {
			if v != nil {
				fmt.Printf("%s\n", value.Print(v)) // Fix me: want a real prompt.
			}
			fmt.Println()
		}
	} else {
		for v := range tokens {
			fmt.Printf("%#v\n\n", v) // Fix me: want a real prompt.
		}
	}
}
