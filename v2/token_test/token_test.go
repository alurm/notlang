package token_test

import (
	"testing"

	"git.sr.ht/~alurm/notlang/stack/token"
	"git.sr.ht/~alurm/notlang/stack/top"
)

func makeChan[T any](slice []T) chan T {
	out := make(chan T, len(slice))
	for _, v := range slice {
		out <- v
	}
	return out
}

func TestSpaces(t *testing.T) {
	c := token.Tokenize(top.In("  \t"))
	s := []token.Token{}
	for v := range c {
		s = append(s, v)
	}
	if len(s) != 1 {
		t.Errorf("want one value")
	}
	if _, ok := s[0].(token.Space); !ok {
		t.Errorf("want space")
	}
}
