package top

import (
	"bufio"
	"fmt"
	"os"

	"git.sr.ht/~alurm/notlang/stack/parse"
	"git.sr.ht/~alurm/notlang/stack/token"
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

func Shell() {
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

	tokens = pipe(
		tokens,
		parse.GroupTop,
		parse.CommandTop,
		parse.DollarStringAsGetCommandGroup,
		parse.ApplicationTop,
		parse.SpaceTop,
	)

	for t := range tokens {
		fmt.Printf("%#v\n", t)
	}
}
