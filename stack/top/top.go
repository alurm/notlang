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
	for t := range parse.SpaceTop(parse.GroupTop(tokens)) {
		fmt.Printf("%#v\n", t)
	}
}
