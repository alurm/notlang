package token

import (
	"strings"
)

/*
Notice,
	echo foo/$[echo bar]
is not the same as
	echo foo/ $[echo bar]
. But
	echo     foo/$[echo bar]
is the same. So space token is needed.
*/

type (
	Token interface {
		token()
	}
	String    string
	Dollar    struct{}
	Space     struct{}
	Separator struct{}
	Open      struct{}
	Close     struct{}

	// Special characters.
	// Quote // '[Hello world] -> "Hello world"
	// Backslash // \' -> "'"

	// Tokenizing must not generate these tokens, but parser will.
	Group []Token // [1 2 3] -> Group{"1", "2", "3"}, consume Opens and Closes
	// 1 2; 3 4 -> Command{"1", "2"} Command{"3", "4"}, consume Separators
	Command []Token
)

func (String) token()    {}
func (Dollar) token()    {}
func (Space) token()     {}
func (Separator) token() {}
func (Open) token()      {}
func (Close) token()     {}

func (Group) token()   {}
func (Command) token() {}

func Tokenize(in chan byte) chan Token {
	// in := bufio.NewReader(r)

	out := make(chan Token)

	go func() {
		var (
			peeked bool
			peek   byte
			b      byte
		)
	Top:
		for {
			if peeked {
				b = peek
			} else {
				var ok bool
				b, ok = <-in
				if !ok {
					close(out)
					break
				}
			}

			peeked = false
			switch b {
			case ' ', '\t':
				out <- Space{}
				for {
					var ok bool
					b, ok = <-in
					if !ok {
						close(out)
						break Top
					}
					if b != ' ' && b != '\t' {
						peeked = true
						peek = b
						break
					}
				}
			case ';', '\n':
				out <- Separator{}
			case '[':
				out <- Open{}
			case ']':
				out <- Close{}
			case '$':
				out <- Dollar{}
			case '\'':
				// Restricted for now.
				// Would be better to reuse and recurse tokenizer.
				// Perhaps by changing its mode somehow.
				var sb strings.Builder
				var ok bool
				var depth int
				b, ok = <-in
				if !ok || b != '[' {
					panic(nil)
				}
				depth++
			Quoted:
				for {
					b, ok = <-in
					if !ok {
						panic(nil)
					}
					switch b {
					case '[':
						depth++
					case ']':
						depth--
						if depth == 0 {
							break Quoted
						}
					case '\\':
						b, ok = <-in
						if !ok {
							panic(nil)
						}
					}
					sb.WriteByte(b)
				}
				out <- String(sb.String())
			case '\\':
				b, ok := <-in
				if !ok {
					// Restricted for now.
					panic(nil)
					//out <- String("\\")
				} else {
					out <- String(b)
				}
			default:
				var sb strings.Builder
			Word:
				for {
					var ok bool
					switch b {
					case ' ', '\t', ';', '\n', '[', ']', '$', '\'', '\\':
						peeked = true
						peek = b
						break Word
					default:
						sb.WriteByte(b)
					}
					b, ok = <-in
					if !ok {
						out <- String(sb.String())
						close(out)
						break Top
					}
				}
				out <- String(sb.String())
			}
		}
	}()

	return out
}
