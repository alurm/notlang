/*
Minimal syntax:
- [] -- function definition
- 1 2 3 (top level) -- function call, 2 and 3 applied to 1
- $[] (not top level) -- function call
*/
package parse

import "git.sr.ht/~alurm/notlang/stack2/code"

type (
	Token  interface{ token() }
	Dollar struct{}
	Open   struct{}
	Close  struct{}
	String string

	Group   []Token
	Command []Token
)

func (Dollar) token()  {}
func (Open) token()    {}
func (Close) token()   {}
func (String) token()  {}
func (Group) token()   {}
func (Command) token() {}

func Tokenize(in chan byte) chan Token {
	out := make(chan Token)
	go func() {
		var (
			peeked bool
			b      byte
			ok     bool
		)
		for {
			if !peeked {
				b, ok = <-in
				if !ok {
					close(out)
					return
				}
			}
			peeked = false
			switch b {
			case '$':
				out <- Dollar{}
			case '[':
				out <- Open{}
			case ']':
				out <- Close{}
			case ' ', '\t', '\n':
			default:
				var s string
			Name:
				for {
					s += string(b)
					b, ok = <-in
					if !ok {
						out <- String(s)
						close(out)
						return
					}
					switch b {
					case '$', '[', ']', ';', '\n', ' ', '\t':
						out <- String(s)
						peeked = true
						break Name
					}
				}
			}
		}
	}()
	return out
}

func Wrap[A, B any](
	f func(chan A) chan B,
	in []A,
) (out []B) {
	inChan := make(chan A, len(in))
	go func() {
		for _, v := range in {
			inChan <- v
		}
		close(inChan)
	}()

	c := f(inChan)

	for v := range c {
		out = append(out, v)
	}
	return out
}

func GroupThem(in chan Token) chan Token {
	out := make(chan Token)
	go func() {
		for t := range in {
			switch t := t.(type) {
			case Open:
				c := GroupThem(in)
				var o Group
				for v := range c {
					o = append(o, v)
				}
				out <- o
			case Close:
				close(out)
				return
			default:
				out <- t
			}
		}
		close(out)
	}()
	return out
}

func DollarThem(in chan Token) chan Token {
	out := make(chan Token)
	go func() {
		for t := range in {
			switch t := t.(type) {
			case Group:
				out <- Group(Wrap(DollarThem, t))
			case Dollar:
				next := <-in
				switch next := next.(type) {
				case Group:
					out <- Command(Wrap(DollarThem, next))
				case String:
					out <- Command([]Token{
						String("get"),
						next,
					})
				default:
					panic(nil)
				}
			default:
				out <- t
			}
		}
		close(out)
	}()
	return out
}

func Parse(in chan Token) chan code.Code {
	out := make(chan code.Code)
	go func() {
		for t := range in {
			switch t := t.(type) {
			case Group:
				out <- code.Closure(Wrap(Parse, t))
			case Command:
				out <- code.Command(Wrap(Parse, t))
			case String:
				out <- code.String(t)
			default:
				panic(nil)
			}
		}
		close(out)
	}()
	return out
}
