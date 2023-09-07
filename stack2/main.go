package main

type (
	Code interface {
		code()
	}
	StringCode  string
	CommandCode struct {
		Head Code
		Tail []Code
	}
)

func (StringCode) code()  {}
func (CommandCode) code() {}

type (
	Frame interface {
		frame()
	}
	StringFrame  struct{}
	CommandFrame struct {
		Current int
		Head    Command
		Tail    []Value
	}
)

func (StringFrame) frame()  {}
func (CommandFrame) frame() {}

type (
	Value interface {
		value()
	}
	String string
	Stack  struct {
		Up     *Stack
		Lookup func(String) Value
		Frame
		Code Code
	}
	Command func([]Value) Value
)

func (String) value()  {}
func (Stack) value()   {}
func (Command) value() {}

func Evaluate(top Stack) {
	curr := top
	for {
		switch curr.Code.(type) {
		default:
			panic(nil)
		}
	}
}

func main() {
}
