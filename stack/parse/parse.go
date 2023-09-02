package parse

import "git.sr.ht/~alurm/notlang/stack/token"

type (
	Tree    interface{ tree() }
	String  string // echo
	Command []Tree // echo hi
	// Names from lambda calculus.
	Abstraction []Command // [echo hi; echo lol]
	Application []Command // $[echo hi; echo lol]
)

func (String) tree()      {}
func (Command) tree()     {}
func (Abstraction) tree() {}
func (Application) tree() {}

/*func Parse(in chan token.Token) chan Tree {
	out := make(chan Tree)

	go func() {
		for t := range in {
			switch t := t.(type) {
			case
			}
		}
		close(out)
	}()
	return out
}*/

// Consumes all token.Separator tokens.
func CommandTop(in chan token.Token) chan token.Token {
	out := make(chan token.Token)
	go func() {
		var command token.Command
		for t := range in {
			switch t := t.(type) {
			case token.Separator:
				if command != nil {
					out <- command
					command = nil
				}
			default:
				command = append(command, t)
			}
		}
		close(out)
	}()
	return out
}

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
	go func() {
		var paste token.Paste
		for t := range in {
			switch t := t.(type) {
			case token.Space:
				if paste == nil {
					continue
				}
				if len(paste) == 1 {
					out <- paste[0]
					paste = nil
					continue
				}
				out <- paste
				paste = nil
			case token.Separator:
				if paste == nil {
					out <- token.Separator{}
					continue
				}
				if len(paste) == 1 {
					out <- paste[0]
					paste = nil
					out <- token.Separator{}
					continue
				}
				out <- paste
				paste = nil
				out <- token.Separator{}
			case token.Group:
				// Ugly? Doesn't leave routine hanging at least?
				paste = append(paste, token.Group(Slice(SpaceTop(Chan(t)))))
			default:
				paste = append(paste, t)
			}
		}
		if paste != nil {
			if len(paste) == 1 {
				out <- paste[0]
			} else {
				out <- paste
			}
		}
		close(out)
	}()
	return out
}
