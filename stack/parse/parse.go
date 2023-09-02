package parse

type (
	Tree     interface{ tree() }
	String   string    // echo
	Command  []Tree    // echo hi
	Block    []Command // $[echo hi; echo lol]
	Function []Command // [echo hi; echo lol]
)

func (String) tree()   {}
func (Command) tree()  {}
func (Block) tree()    {}
func (Function) tree() {}
