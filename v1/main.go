package main

import (
	"fmt"
	"os"
	"strconv"
)

type (
	Token     interface{ Token() }
	String    string
	Semicolon struct{}
	Bracket   struct {
		Open   bool
		Square bool
	}
)

// for (i in String Semicolon Bracket) echo 'func ('$i') Token() {}'
func (String) Token()    {}
func (Semicolon) Token() {}
func (Bracket) Token()   {}

type (
	SyntaxNode  interface{ SyntaxNode() }
	SyntaxBlock struct {
		Call       bool
		Statements []SyntaxStatement
	}
	SyntaxStatement []SyntaxNode
)

func (String) SyntaxNode()      {}
func (SyntaxBlock) SyntaxNode() {}

func Tokenize(input string) (out []Token) {
	var i int
	for i < len(input) {
		switch b := input[i]; b {
		case ' ', '\t':
			i++
		case ';', '\n': // Hack.
			out = append(out, Semicolon{})
			i++
		case '[', ']', '{', '}':
			out = append(out, Bracket{
				Open:   b == '[' || b == '{',
				Square: b == '[' || b == ']',
			})
			i++
		default:
			var s string
			special := func(input byte) bool {
				switch input {
				case ' ', '\t', ';', '[', ']', '{', '}', '\n':
					return true
				default:
					return false
				}
			}
			for i < len(input) && !special(input[i]) {
				s += string(input[i])
				i++
			}
			out = append(out, String(s))
		}
	}
	return out
}

func Parse(tokens []Token, call bool) (out SyntaxBlock, i int) {
	var statement []SyntaxNode
	out.Call = call
	end := func() bool {
		if i >= len(tokens) {
			return true
		}
		token, ok := tokens[i].(Bracket)
		if ok && !token.Open {
			return true
		}
		return false
	}
	for !end() {
		switch token := tokens[i].(type) {
		case String:
			statement = append(statement, token)
			//out.Statements = append(out.Statements, token)
			i++
		case Bracket:
			i++
			block, shift := Parse(
				tokens[i:],
				token.Square,
			)
			i += shift
			i++
			statement = append(statement, block)
			//out.Statements = append(out.Statements, block)
		case Semicolon:
			out.Statements = append(out.Statements, statement)
			statement = nil
			i++
		}
	}
	out.Statements = append(out.Statements, statement)
	return
}

func ParseTop(tokens []Token) SyntaxBlock {
	b, _ := Parse(tokens, true)
	return b
}

type (
	Value     interface{ Value() }
	Closure   func([]Value) Value
	Primitive func([]Value, *Frame) Value
)

func (String) Value()    {}
func (Closure) Value()   {}
func (Primitive) Value() {}

func BuiltinPrint(values []Value, _ *Frame) Value {
	var result String
	for i, v := range values {
		if i != 0 {
			result += " "
		}
		result += v.(String)
	}
	Print(result)
	return result
}

func BuiltinGet(values []Value, f *Frame) Value {
	return *Lookup(values[0].(String), f)
}

func BuiltinSet(values []Value, f *Frame) Value {
	name := values[0].(String)
	value := values[1]
	*Lookup(name, f) = value
	return String("")
}

func Lookup(name String, f *Frame) *Value {
	if f == nil {
		return nil
	}
	if f.Names[name] != nil {
		return f.Names[name]
	}
	return Lookup(name, f.Up)
}

func BuiltinLet(value []Value, f *Frame) Value {
	name := value[0].(String)
	var nothing Value
	f.Names[name] = &nothing
	return String("")
}

func EvaluateSubstatements(s SyntaxStatement, f *Frame) (values []Value) {
	//fmt.Printf("Substatement: %#v\n", s)
	for _, n := range s {
		values = append(values, Evaluate(n, f))
	}
	return
}

func assert[T any](x T, y error) T {
	if y != nil {
		panic(y)
	}
	return x
}

func CallStatementValue(values []Value, f *Frame) Value {
	tail := values[1:]
	switch head := values[0].(type) {
	case String:
		lookupResult := Lookup(head, f)
		if lookupResult != nil {
			values[0] = *lookupResult
			return CallStatementValue(values, f)
		}
		switch head {
		case "let-set":
			BuiltinLet(tail, f)
			return BuiltinSet(tail, f)
		case "-":
			l := string(tail[0].(String))
			r := string(tail[1].(String))
			li := assert(strconv.Atoi(l))
			ri := assert(strconv.Atoi(r))
			return String(strconv.Itoa(li - ri))
		case "+":
			l := string(tail[0].(String))
			r := string(tail[1].(String))
			li := assert(strconv.Atoi(l))
			ri := assert(strconv.Atoi(r))
			return String(strconv.Itoa(li + ri))
		case "*":
			l := string(tail[0].(String))
			r := string(tail[1].(String))
			li := assert(strconv.Atoi(l))
			ri := assert(strconv.Atoi(r))
			return String(strconv.Itoa(li * ri))
		case "+1":
			l := string(tail[0].(String))
			li := assert(strconv.Atoi(l))
			return String(strconv.Itoa(li + 1))
		case "empty":
			return String("")
		case "panic":
			panic(nil)
		case "if":
			//f := &Frame{Up: f, Names: map[String]*Value{}}
			cond := tail[0].(String)
			then := tail[1].(Closure)
			otherwise := tail[2].(Closure)
			if string(cond) != "" {
				return then([]Value{})
			} else {
				return otherwise([]Value{})
			}
		case "=":
			switch l := tail[0].(type) {
			case String:
				r, ok := tail[1].(String)
				if !ok {
					return String("")
				}
				if l == r {
					return String("yes")
				}
				return String("")
			case Closure:
				return String("")
			default:
				panic(nil)
			}
		default:
			panic("Not implemented: " + string(head))
		}
	case Closure:
		//fmt.Println("Evaluating")
		//fmt.Printf("Head: %#v\n", head)
		result := head(tail)
		//Print(result)
		return result
	case Primitive:
		return head(tail, f)
	default:
		panic("Not implemented")
	}
}

func EvaluateStatement(s SyntaxStatement, f *Frame) Value {
	values := EvaluateSubstatements(s, f)
	return CallStatementValue(values, f)
}

type Frame struct {
	Up    *Frame
	Names map[String]*Value
}

func Evaluate(n SyntaxNode, f *Frame) Value {
	switch n := n.(type) {
	case String:
		return n
	case SyntaxBlock:
		f := &Frame{
			Up:    f,
			Names: map[String]*Value{},
		}
		if n.Call {
			var v Value
			for _, s := range n.Statements {
				if len(s) != 0 {
					v = EvaluateStatement(s, f)
				}
			}
			return v
		} else {
			return Closure(func(args []Value) Value {
				// Create new frame with arguments.
				f := &Frame{
					Up:    f,
					Names: map[String]*Value{},
				}
				for i, v := range args {
					v := v
					f.Names[String(strconv.Itoa(i))] = &v
				}
				asCall := n
				asCall.Call = true
				return Evaluate(asCall, f)
			})
		}
	default:
		panic(nil)
	}
}

/*
let make-counter {
	let counter
	set counter 0
	value {
		set counter [increment [get counter]]
		counter
	}
}
let c1
set c1 [make-counter]
let c2
set c2 [make-counter]

c1's and c2's counter variables are different

Closure below saves reference to upper frame?
*/

func Print(v Value) {
	s := v.(String)
	fmt.Println(s)
}

func Ptr[T any](value T) *T { return &value }

func MakeClosure(source string, f *Frame) Value {
	out, _ := Parse(Tokenize(source), false)
	return Evaluate(out, f)
}

func Builtins() map[String]*Value {
	builtins := map[String]*Value{
		"print": Ptr(Value(Primitive(BuiltinPrint))),
		"let":   Ptr(Value(Primitive(BuiltinLet))),
		"get":   Ptr(Value(Primitive(BuiltinGet))),
		"set":   Ptr(Value(Primitive(BuiltinSet))),
		"value": Ptr(Value(Primitive(func(v []Value, _ *Frame) Value {
			return v[0]
		}))),
	}
	//builtins[String("let-set")] = Ptr(MakeClosure(""))
	return builtins
}

func main() {
	tokens := Tokenize(
		//"print hi world,  good",
		os.Args[1],
	)
	//tokens := Tokenize("echo hello world; echo [echo good]")
	//tokens := Tokenize("[foo] hello world [good]")
	for _, token := range tokens {
		//fmt.Printf("%#v\n", token)
		_ = token
	}
	//fmt.Println()
	syntax := ParseTop(tokens)
	//fmt.Printf("%+v\n", syntax)
	//fmt.Println()
	//fmt.Printf("%#v\n", Evaluate(syntax, nil))
	value := Evaluate(
		syntax,
		&Frame{Up: nil, Names: Builtins()},
	)
	Print(value)
}
