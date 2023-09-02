package parse

import "git.sr.ht/~alurm/notlang/stack/token"

type (
	Tree     interface{ tree() }
	String   string    // echo
	Command  []Tree    // echo hi
	Block    []Command // $[echo hi; echo lol]
	Function []Command // [echo hi; echo lol]
)

func (String) tree()   {}
func (Command) tree()  {}
func (Block) tree()    {}
func (Function) tree() {}

func GroupTop(in chan token.Token) chan token.Token {
	out := make(chan token.Token)
	go func() {
		for t := range in {
			switch t := t.(type) {
			case token.Close:
				panic(nil)
			case token.Open:
				out <- Group(in)
			default:
				out <- t
			}
		}
		close(out)
	}()
	return out
}

func Group(in chan token.Token) (out token.Group) {
	for t := range in {
		switch t := t.(type) {
		case token.Close:
			return out
		case token.Open:
			out = append(out, Group(in))
		default:
			out = append(out, t)
		}
	}
	panic(nil)
}
