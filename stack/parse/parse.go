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
			case token.Group:
				command = append(command, token.Group(Slice(CommandTop(Chan(t)))))
			default:
				command = append(command, t)
			}
		}
		if command != nil {
			out <- command
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
Spaces separate, commands separate.
Strings stick, dollars stick, groups stick.
*/
func SpaceTop(in chan token.Token) chan token.Token {
	out := make(chan token.Token)
	send := func(out chan token.Token, paste *[]token.Token) {
		if *paste != nil {
			if len(*paste) == 1 {
				out <- (*paste)[0]
			} else {
				out <- token.Dollar{}
				command := token.Command{token.String("paste")}
				command = append(command, (*paste)...)
				out <- token.Group{command}
			}
			*paste = nil
		}
	}
	go func() {
		var paste []token.Token
		for t := range in {
			switch t := t.(type) {
			case token.Space:
				send(out, &paste)
			case token.Group:
				// Ugly? Doesn't leave routine hanging at least?
				paste = append(paste, token.Group(Slice(SpaceTop(Chan(t)))))
			case token.Command:
				send(out, &paste)
				out <- token.Command(Slice(SpaceTop(Chan(t))))
			default:
				paste = append(paste, t)
			}
		}
		send(out, &paste)
		close(out)
	}()
	return out
}
