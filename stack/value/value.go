package value

import (
	"fmt"

	"git.sr.ht/~alurm/notlang/stack/parse"
)

type (
	Value  interface{ value() }
	String string
	// Maybe args and return type should be stored in the continuation.
	Abstraction func(c *Continuation, args []Value) Value
)

func (String) value()      {}
func (Abstraction) value() {}

type Continuation struct {
	Code   parse.Tree
	Names  map[String]Value
	Up     *Continuation
	Lookup func(c *Continuation, s String) Value
}

// Lookup of the lookup!
func Lookup(c *Continuation, s String) Value {
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
				Code: loop,
				Up:   c,
			})
			command = append(command, value)
		}
		head := command[0]
		tail := command[1:]
		switch head := head.(type) {
		case String:
			headValue := Lookup(c, head)
			return headValue.(Abstraction)(c, tail)
			panic(nil)
		case Abstraction:
			// Keep the same continuation?
			return head(c, tail)
		default:
			panic(nil)
		}
	default:
		panic(nil)
	}
}

func DefaultLookup(c *Continuation, s String) (out Value) {
	for c != nil {
		out = c.Names[s]
		if out != nil {
			return out
		}
		c = c.Up
	}
	panic(nil)
}

func Shell(in chan parse.Tree) chan Value {
	out := make(chan Value)
	go func() {
		topContinuation := Continuation{
			Lookup: DefaultLookup,
			Names: map[String]Value{
				"foo": Abstraction(
					func(c *Continuation, args []Value) Value {
						return String("bar")
					},
				),
				"let": Abstraction(
					func(c *Continuation, args []Value) Value {
						key := args[0].(String)
						value := args[1]
						c.Names[key] = value
						return nil
					},
				),
				"names": Abstraction(
					func(c *Continuation, args []Value) Value {
						for name := range c.Names {
							fmt.Println(name)
						}
						return nil
					},
				),
				"get": Abstraction(
					func(c *Continuation, args []Value) Value {
						// Fix me: do a proper Lookup.
						key := args[0].(String)
						return c.Names[key]
					},
				),
			},
		}

		for tree := range in {
			topContinuation.Code = tree
			out <- Evaluate(&topContinuation)
		}
		close(out)
	}()
	return out
}
