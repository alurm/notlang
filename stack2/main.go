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
	// Set to 0 in closures to indicate where to stop.
	Depth int
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

func Evaluate(stack *Stack) Value {
	for {
		switch c := stack.Code.(type) {
		case code.String:
			if stack.Depth == 0 {
				return String(c)
			}
			stack = stack.Up
			stack.Command = append(stack.Command, String(c))
			stack.Pointer++

		case code.Command:
			/*if stack.Pointer == 0 { // Head.
				lookup := Lookup(stack, String(c[0].(code.String))).(Command)
				stack.Command = append(stack.Command, lookup)
				stack.Pointer++
			}*/

			if stack.Pointer == len(c) { // Off by one.
				var command Command
				switch head := stack.Command[0].(type) {
				case String:
					command = Lookup(stack, head).(Command)
				case Command:
					command = head
				default:
					panic(nil)
				}
				value := command(Stack{
					Up:    stack,
					Args:  stack.Command[1:],
					Depth: stack.Depth + 1,
				})
				if stack.Depth == 0 {
					return value
				}
				stack = stack.Up
				stack.Command = append(stack.Command, value)
				stack.Pointer++
			} else { // Needs further processing.
				stack = &Stack{
					Up:    stack,
					Code:  c[stack.Pointer],
					Depth: stack.Depth + 1,
				}
			}

		case code.Closure:
			// Shouldn't proceed evaluation, instead returning a thunk.

			// Is copy needed?
			copy := *stack
			copy.Code = code.Command(c)
			copy.Depth = 0

			closure := Command(func(forArgs Stack) Value {
				return Evaluate(&copy)
			})

			if stack.Depth == 0 {
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

func PrintFrame(c Stack) {
	fmt.Printf(
		"# Frame:\n\nUp: %#v\n\nLookup: %#v\n\nCommand: %#v\n\nPointer: %#v\n\nCode: %#v\n\nArgs: %#v\n\n",
		c.Up,
		c.Lookup,
		c.Command,
		c.Pointer,
		c.Code,
		c.Args,
	)
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
				fmt.Print("Hello, world")
				return String("")
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
			"if": func(s Stack) Value {
				args := s.Args
				cond := args[0]
				then_ := args[1].(Command)
				else_ := args[2].(Command)
				if cond.(String) != "" {
					return then_(Stack{Up: &s, Depth: s.Depth + 1})
				} else {
					return else_(Stack{Up: &s, Depth: s.Depth + 1})
				}
			},
			"empty": func(s Stack) Value {
				return String("")
			},
			/* Doesn't work properly.
			$[let double [+ $[arg 0] $[arg 0]]]

			$[double 20]
			40

			$[double 10]
			40 # Mistake.
			*/
			"arg": func(s Stack) Value {
				// Confusing.
				// Wanted to use s.Up.Args but need to use s.Up.Command[1:]?
				// s.Up.Up because $[arg x] creates a frame as well.

				// Should find nearest frame with Depth = 0 probably.

				n, err := strconv.Atoi(string(s.Args[0].(String)))
				if err != nil {
					panic(err)
				}
				/*for c := &s; c != nil; c = c.Up {
					PrintFrame(*c)
				}*/
				/*for _, v := range s.Up.Args {
					fmt.Println("arg:", Print(v))
				}*/
				/*cmd := s.Up.Command
				for _, v := range cmd[1:] {
					fmt.Println("cmd:", Print(v))
				}*/
				var curr = &s
				for ; curr.Depth != 0; curr = curr.Up { // Call to arg.
				}
				curr = curr.Up
				for ; curr.Depth != 0; curr = curr.Up { // Real call.
				}
				return curr.Command[n+1]
				return s.Up.Up.Command[n+1]
			},
			"let": func(s Stack) Value { // Doesn't work :(
				key := s.Args[0].(String)
				value := s.Args[1]
				up := s.Up
				prev := up.Lookup
				up.Lookup = func(v Value) Value {
					str, ok := v.(String)
					if !ok || str != key {
						return prev(v)
					}
					return value
				}
				return nil
			},
			"print": func(s Stack) Value {
				fmt.Print(s.Args[0])
				return String("")
			},
			"stack": func(s Stack) Value {
				c := &s
				for c != nil {
					PrintFrame(*c)
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
			Depth: 0,
		}

		for code := range parse.Parse(parse.DollarThem(parse.GroupThem(tokens))) {
			func() {
				defer func() {
					/*err := recover()
					if err != nil {
						fmt.Println(err)
					}*/
				}()
				// Ugly.
				frame.Pointer = 0
				frame.Command = nil
				frame.Code = code
				frame.Args = nil
				frame.Depth = 0
				// Pass by pointer so let works.
				v := Evaluate(&frame)
				if v == nil {
					fmt.Println()
				} else {
					fmt.Printf("%s\n\n", Print(v))
				}
			}()
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