package code

type (
	Code    interface{ code() }
	String  string
	Command []Code
	Closure []Code
)

func (String) code()  {}
func (Command) code() {}
func (Closure) code() {}
