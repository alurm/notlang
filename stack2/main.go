package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"

	"git.sr.ht/~alurm/notlang/stack2/code"
	"git.sr.ht/~alurm/notlang/stack2/parse"
)

type (
	Value interface {
		value()
	}
	String  string
	Command func(Stack) Value
)

type Stack struct {
	Up      *Stack
	Lookup  func(Value) Value
	Pointer int
	Command []Value
	Code    code.Code
	Args    []Value

	Depth int // Debug.
}

func (String) value()  {}
func (Command) value() {}

func Lookup(stack *Stack, v Value) Value {
	for {
		if stack.Lookup == nil {
			stack = stack.Up
		} else {
			lookup := stack.Lookup(v)
			if lookup != nil {
				return lookup
			}
		}
	}
}

func Evaluate(in Stack) Value {
	stack := &in
	for {
		switch c := stack.Code.(type) {
		case code.String:
			if stack.Up == nil {
				return String(c)
			}
			stack = stack.Up
			stack.Command = append(stack.Command, String(c))
			stack.Pointer++

		case code.Command:
			if stack.Pointer == 0 { // Head.
				lookup := Lookup(stack, String(c[0].(code.String))).(Command)
				stack.Command = append(stack.Command, lookup)
				stack.Pointer++
			}

			if stack.Pointer == len(c) { // Off by one.
				command := stack.Command[0].(Command)
				value := command(Stack{
					Up:   stack,
					Args: stack.Command[1:],
				})
				if stack.Up == nil {
					return value
				}
				stack = stack.Up
				stack.Command = append(stack.Command, value)
				stack.Pointer++
			} else { // Needs further processing.
				stack = &Stack{
					Up:   stack,
					Code: c[stack.Pointer],
				}
			}

		case code.Closure:
			// Shouldn't proceed evaluation, instead returning a thunk.

			copy := *stack

			closure := Command(func(forArgs Stack) Value {
				return Evaluate(copy)
			})

			if stack.Up == nil {
				return closure
			}
			stack = stack.Up
			stack.Command = append(stack.Command, closure)
			stack.Pointer++

		default:
			panic(nil)
		}
	}
}

func main() {
	var tokens chan parse.Token
	{
		c := make(chan byte)
		tokens = parse.Tokenize(c)
		go func() {
			in := bufio.NewReader(os.Stdin)
			for {
				b, err := in.ReadByte()
				if err != nil {
					close(c)
					return
				}
				c <- b
			}
		}()
	}

	switch true {
	case false:
		for token := range parse.Parse(parse.DollarThem(parse.GroupThem(tokens))) {
			fmt.Printf("%#v\n", token)
		}
	case true:
		var commands map[String]Command
		commands = map[String]Command{
			"hello": func(Stack) Value {
				return String("Hello, world")
			},
			"get": func(s Stack) Value {
				return Lookup(&s, s.Args[0])
			},
			"give": func(s Stack) Value {
				return s.Args[0]
			},
			"+": func(s Stack) Value {
				sum := 0
				for _, arg := range s.Args {
					str := arg.(String)
					i, err := strconv.Atoi(string(str))
					if err != nil {
						panic(err)
					}
					sum += i
				}
				return String(strconv.Itoa(sum))
			},
			"=": func(s Stack) Value {
				l := s.Args[0].(String)
				r := s.Args[1].(String)
				if l == r {
					return String("yes")
				} else {
					return String("")
				}
			},
			// Doesn't work :(
			// $[if true [print hi] [print bye]]
			"if": func(s Stack) Value {
				args := s.Args
				cond := args[0]
				then_ := args[1].(Command)
				else_ := args[2].(Command)
				if cond.(String) != "" {
					return then_(Stack{Up: &s})
				} else {
					return else_(Stack{Up: &s})
				}
			},
			"let": func(s Stack) Value { // Doesn't work :(
				key := s.Args[0].(String)
				fmt.Println("Key:", key)
				value := s.Args[1]
				up := s
				up.Lookup = func(v Value) Value {
					return commands["hello"]
					str, ok := v.(String)
					if !ok || str != key {
						return up.Lookup(v)
					}
					return value
				}
				return String("")
			},
			"print": func(s Stack) Value {
				fmt.Print(s.Args[0])
				return String("")
			},
			"stack": func(s Stack) Value {
				c := &s
				for c != nil {
					fmt.Printf(
						"Up: %#v\n\nLookup: %#v\n\nPointer: %#v\n\nCode: %#v\n\nArgs: %#v\n\n",
						c.Up,
						c.Lookup,
						c.Pointer,
						c.Code,
						c.Args,
					)
					c = c.Up
				}
				return String("")
			},
		}

		var frame Stack = Stack{
			Lookup: func(in Value) Value {
				s := in.(String)
				command, ok := commands[s]
				if !ok {
					panic(nil)
				}
				return command
			},
		}

		for code := range parse.Parse(parse.DollarThem(parse.GroupThem(tokens))) {
			frame.Code = code
			fmt.Printf("%s\n\n", Print(Evaluate(frame)))
		}
	}
}

func Print(v Value) string {
	switch v := v.(type) {
	case String:
		return string(v)
	case Command:
		return fmt.Sprint(v)
	default:
		panic(nil)
	}
}