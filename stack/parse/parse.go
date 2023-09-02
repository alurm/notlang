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

// Consumes all token.Open and token.Close tokens.
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

func Chan[T any](slice []T) chan T {
	out := make(chan T, len(slice))
	for _, v := range slice {
		out <- v
	}
	close(out)
	return out
}

func Slice[T any](channel chan T) (out []T) {
	for v := range channel {
		out = append(out, v)
	}
	return out
}

/*
Consumes all token.Space tokens.
Spaces separate, separators separate.
Strings stick, dollars stick, groups stick.
*/
func SpaceTop(in chan token.Token) chan token.Token {
	out := make(chan token.Token)
	var one token.Token
	var paste token.Paste
	go func() {
		for t := range in {
			switch t := t.(type) {
			case token.Space:
				if one == nil && paste == nil {
					continue
				}
				if one != nil {
					out <- one
					one = nil
					continue
				}
				out <- paste
				paste = nil
			case token.Separator:
				if one == nil && paste == nil {
					out <- token.Separator{}
					continue
				}
				if one != nil {
					out <- one
					one = nil
					out <- token.Separator{}
					continue
				}
				out <- paste
				paste = nil
				out <- token.Separator{}
			case token.Group:
				// Ugly? Doesn't leave routine hanging at least?
				out <- token.Group(Slice(SpaceTop(Chan(t))))
			default:
				if one == nil && paste == nil {
					one = t
					continue
				}
				if one != nil {
					paste = append(paste, one)
					one = nil
					paste = append(paste, t)
					continue
				}
				paste = append(paste, t)
			}
		}
		if one != nil {
			out <- one
		}
		if paste != nil {
			out <- paste
		}
		close(out)
	}()
	return out
}
