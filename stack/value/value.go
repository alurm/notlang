package value

import (
	"os"
	"fmt"
	"strconv"

	"git.sr.ht/~alurm/notlang/stack/parse"
)

type (
	Value  interface{ value() }
	String string
	// Abstraction func(c *Continuation, args []Value, in chan Value) chan Value
	// Abstraction func(c *Continuation, in chan Value) chan Value
	/*
		grep $[cat] foo
		grep $[put foo; cat]
	*/
	// Maybe args and return type should be stored in the continuation.
	Command func(c *Continuation, args []Value) Value
	//Abstraction func(c *Continuation, args []Value) Value
)

func (String) value()  {}
func (Command) value() {}

type Continuation struct {
	Code   parse.Tree
	Names  map[String]*Value
	Up     *Continuation
	Lookup func(c *Continuation, s String) *Value
}

// Lookup of the lookup!
func Lookup(c *Continuation, s String) *Value {
	lookup := c
	for lookup.Lookup == nil {
		lookup = lookup.Up
	}
	if lookup == nil {
		panic(nil)
	}
	return lookup.Lookup(c, s)
}

func Evaluate(c *Continuation) Value {
	switch code := c.Code.(type) {
	case parse.String:
		return String(code)
	case parse.Command:
		if len(code) == 0 {
			panic(nil)
		}
		var command []Value
		for _, loop := range code {
			value := Evaluate(&Continuation{
				Code:  loop,
				Up:    c,
				Names: map[String]*Value{},
			})
			command = append(command, value)
		}
		head := command[0]
		tail := command[1:]
		switch head := head.(type) {
		case String:
			headValue := *Lookup(c, head)
			return headValue.(Command)(c, tail)
			panic(nil)
		case Command:
			// Keep the same continuation?
			return head(c, tail)
		default:
			panic(nil)
		}
	case parse.Application:
		// Confused about this.
		c := &Continuation{
			Up:    c,
			Names: map[String]*Value{},
		}
		var value Value
		for _, cmd := range code {
			c.Code = cmd
			value = Evaluate(c)
		}
		return value
	case parse.Abstraction:
		// Confused about this. Where to put args?
		return Command(func(_ *Continuation, args []Value) Value {
			/*c := &Continuation{
				Up:    c,
				Names: map[String]*Value{},
				Code: c.Code,
			}*/
			c.Code = parse.Application(code)
			return Evaluate(c)
		})
	default:
		panic(nil)
	}
}

func DefaultLookup(c *Continuation, s String) *Value {
	var out *Value
	for c != nil {
		out = c.Names[s]
		if out != nil {
			return out
		}
		c = c.Up
	}
	panic(nil)
}

func Ptr[T any](v T) *T { return &v }

func Must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}

func Shell(in chan parse.Tree) chan Value {
	out := make(chan Value)
	go func() {
		funcs := map[String]Command{
			"return": func(c *Continuation, args []Value) Value {
				return args[0]
			},
			"+": func(c *Continuation, args []Value) Value {
				return String(strconv.Itoa(
					Must(strconv.Atoi(string(args[0].(String)))) +
						Must(strconv.Atoi(string(args[1].(String)))),
				))
			},
			"paste": func(c *Continuation, args []Value) Value {
				var out String
				for _, s := range args {
					out += s.(String)
				}
				return out
			},
			"print": func(c *Continuation, args []Value) Value {
				fmt.Println(args[0])
				return nil
			},
			"foo": func(c *Continuation, args []Value) Value {
				return String("bar")
			},
			"set": func(c *Continuation, args []Value) Value {
				key := args[0].(String)
				value := args[1]
				*Lookup(c, key) = value
				return nil
			},
			"let": func(c *Continuation, args []Value) Value {
				key := args[0].(String)
				value := args[1]
				c.Names[key] = &value
				return nil
			},
			"names": func(c *Continuation, args []Value) Value {
				for name := range c.Names {
					fmt.Println(name)
				}
				return nil
			},
			"get": func(c *Continuation, args []Value) Value {
				key := args[0].(String)
				return *Lookup(c, key)
			},
			"env": func(c *Continuation, args[]Value) Value {
				switch args[0].(String) {
				case "get":
					return String(os.Getenv(string(args[1].(String))))
				default:
					panic(nil)
				}
			},
		}
		names := map[String]*Value{}
		for k, v := range funcs {
			v := v
			names[k] = Ptr(Value(v))
		}
		topContinuation := Continuation{
			Lookup: DefaultLookup,
			Names:  names,
		}

		for tree := range in {
			topContinuation.Code = tree
			out <- Evaluate(&topContinuation)
		}
		close(out)
	}()
	return out
}
