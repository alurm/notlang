package value

import "git.sr.ht/~alurm/notlang/stack/parse"

type (
	Value  interface{ value() }
	String string
)

func (String) value() {}

/*func Evaluate(parse.Tree, Continuation) {
}*/

func Shell(in chan parse.Tree) chan Value {
	out := make(chan Value)
	go func() {
		for _ = range in {
			out <- String("work in progress") /*Evaluate(t, nil)*/
		}
		close(out)
	}()
	return out
}
