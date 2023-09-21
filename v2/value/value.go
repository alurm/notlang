package value

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"git.sr.ht/~alurm/notlang/v2/parse"
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
	// Not sure about this.
	Args *[]Value
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
			c.Args = &args
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
		var funcs map[String]Command
		funcs = map[String]Command{
			"return": func(c *Continuation, args []Value) Value {
				return args[0]
			},
			"args": func(c *Continuation, args []Value) Value {
				curr := c
				for curr != nil {
					if curr.Args != nil {
						return funcs["list"](
							c,
							*(curr.Args),
						)
					}
					curr = curr.Up
				}
				panic(nil)
			},
			// Crazy code.
			"upcall": func(c *Continuation, args []Value) Value {
				levelstr := string(args[0].(String))
				args = args[1:]
				head := args[0].(Command)
				//look := *Lookup(c, head.(String))
				list := args[1].(Command)
				var tail []Value
				sizeval := list(c, []Value{String("size")})
				sizestr := string(sizeval.(String))
				size := Must(strconv.Atoi(sizestr))
				for i := 0; i < size; i++ {
					istr := String(strconv.Itoa(i))
					v := list(c, []Value{istr})
					tail = append(tail, v)
				}
				level := Must(strconv.Atoi(levelstr))
				for i := 0; (i < level || level < 0) && c.Up != nil; i++ {
					c = c.Up
				}
				return head(c, tail)
			},
			"call": func(c *Continuation, args []Value) Value {
				head := args[0].(Command)
				//look := *Lookup(c, head.(String))
				list := args[1].(Command)
				var tail []Value
				sizeval := list(c, []Value{String("size")})
				sizestr := string(sizeval.(String))
				size := Must(strconv.Atoi(sizestr))
				for i := 0; i < size; i++ {
					istr := String(strconv.Itoa(i))
					v := list(c, []Value{istr})
					tail = append(tail, v)
				}
				return head(c, tail)
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
				fmt.Println(Print(args[0]))
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
				var names []Value
				var curr = c
				for curr != nil {
					for name := range curr.Names {
						names = append(names, name)
					}
					curr = curr.Up
				}
				return funcs["list"](c, names)
			},
			"unix": func(c *Continuation, args []Value) Value {
				var strs []string
				for _, v := range args {
					strs = append(strs, string(v.(String)))
				}
				cmd := exec.Command(strs[0], strs[1:]...)
				var out strings.Builder
				cmd.Stdout = &out
				err := cmd.Run()
				if err != nil {
					panic(err)
				}
				return String(out.String())
			},
			"lines": func(c *Continuation, args []Value) Value {
				str := string(args[0].(String))
				if str[len(str)-1] != '\n' {
					panic(nil)
				}
				var strs = strings.Split(str, "\n")
				strs = strs[:len(strs)-1]
				var values []Value
				for _, v := range strs {
					values = append(values, String(v))
				}
				return funcs["list"](c, values)
			},
			"list": func(c *Continuation, args []Value) Value {
				return Command(func(c *Continuation, args2 []Value) Value {
					if args2[0].(String) == "size" {
						return String(strconv.Itoa(len(args)))
					}
					i := Must(strconv.Atoi(string(args2[0].(String))))
					return args[i]
				})
			},
			"print-list": func(c *Continuation, args []Value) Value {
				list := args[0].(Command)
				size := Must(strconv.Atoi(string(list(c, []Value{String("size")}).(String))))
				for i := 0; i < size; i++ {
					e := list(
						c,
						[]Value{String(strconv.Itoa(i))},
					)
					funcs["print"](c, []Value{e})
				}
				return nil
			},
			"get": func(c *Continuation, args []Value) Value {
				key := args[0].(String)
				return *Lookup(c, key)
			},
			"env": func(c *Continuation, args []Value) Value {
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
