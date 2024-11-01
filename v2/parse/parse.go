package parse

import (
	"git.sr.ht/~alurm/notlang/v2/token"
)

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

func Parse(in chan token.Token) chan Tree {
	out := make(chan Tree)

	go func() {
		for t := range in {
			switch t := t.(type) {
			case token.String:
				out <- String(t)
			case token.Group:
				var abstraction Abstraction
				trees := wrapTopFunc(Parse, t)
				for _, tree := range trees {
					abstraction = append(abstraction, tree.(Command))
				}
				out <- abstraction
			case token.Application:
				var application Application
				trees := wrapTopFunc(Parse, t)
				for _, tree := range trees {
					application = append(application, tree.(Command))
				}
				out <- application
			case token.Command:
				var command Command
				trees := wrapTopFunc(Parse, t)
				for _, tree := range trees {
					command = append(command, tree)
				}
				out <- command
			default:
				panic(nil)
			}
		}
		close(out)
	}()
	return out
}

// Consumes all token.Dollar tokens.
func ApplicationTop(in chan token.Token) chan token.Token {
	out := make(chan token.Token)
	go func() {
		for t := range in {
			switch t := t.(type) {
			case token.Dollar:
				t2, ok := <-in
				if !ok {
					panic(nil)
				}
				group := t2.(token.Group)
				out <- token.Application(wrapTopFunc(ApplicationTop, group))
			case token.Group:
				out <- token.Group(wrapTopFunc(ApplicationTop, t))
			case token.Command:
				out <- token.Command(wrapTopFunc(ApplicationTop, t))
			default:
				out <- t
			}
		}
		close(out)
	}()
	return out
}

// Rewrites $foo to $[get foo]
func DollarStringAsGetCommandGroup(in chan token.Token) chan token.Token {
	out := make(chan token.Token)
	go func() {
		for t := range in {
			switch t := t.(type) {
			case token.Dollar:
				t2, ok := <-in
				if !ok {
					panic(nil)
				}
				switch t2 := t2.(type) {
				case token.String:
					str := t2
					out <- token.Dollar{}
					out <- token.Group{
						token.Command{
							token.String("get"),
							// Hack.
							token.Space{},
							str,
						},
					}
				case token.Group:
					out <- token.Dollar{}
					out <- token.Group(wrapTopFunc(
						DollarStringAsGetCommandGroup,
						t2,
					))
				case token.Command:
					out <- token.Dollar{}
					out <- token.Command(wrapTopFunc(
						DollarStringAsGetCommandGroup,
						t2,
					))
				default:
					out <- token.Dollar{}
					out <- t2
				}
			case token.Group:
				out <- token.Group(wrapTopFunc(
					DollarStringAsGetCommandGroup,
					t,
				))
			case token.Command:
				out <- token.Command(wrapTopFunc(
					DollarStringAsGetCommandGroup,
					t,
				))
			default:
				out <- t
			}
		}
		close(out)
	}()
	return out
}

// To-do: make this type safe, if possible.
// Given token.Group, return token.Group.
// Requires changes to topFuncs elsewhere probably.
func wrapTopFunc[In, Out any](
	topFunc func(chan In) chan Out,
	in []In,
) []Out {
	return Slice(topFunc(Chan(in)))
}

// Consumes all token.Separator tokens.
func CommandTop(in chan token.Token) chan token.Token {
	out := make(chan token.Token)
	go func() {
		var command token.Command
		for t := range in {
			switch t := t.(type) {
			case token.Separator:
				if command != nil {
					if len(command) == 1 {
						_, ok := command[0].(token.Space)
						if ok {
							command = nil
							continue
						}
					}
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
			if len(command) == 1 {
				_, bad := command[0].(token.Space)
				if !bad {
					out <- command
				}
			} else {
				out <- command
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
				command := token.Command{token.String("paste")}
				command = append(command, (*paste)...)
				out <- token.Application{command}
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
			case token.Application:
				paste = append(paste, token.Application(Slice(SpaceTop(Chan(t)))))
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
