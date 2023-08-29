let-set make-counter {
	let-set count 0
	value {
		set count [+ 1 [get count]]
		get count
	}
}

let-set counter [make-counter]

counter
counter

let-set three [counter]

set counter [make-counter]

counter
counter
counter

let-set four [counter]

print [+ [get three] [get four]]

print hello, world

let-set pair {
	let-set left [get 0]
	let-set right [get 1]
	value {
		if [= [get 0] left] {
			get left
		} {
			get right
		}
	}
}

[pair 3 4] left

let-set closure {empty}

let-set list [pair 1 [pair 2 [pair 3 end]]]

let-set iter {
	let-set list [get 0]
	let-set func [get 1]
	if [= [get list] end] {
		empty
	} {
		func [list left]
		iter [list right] [get func]
	}

}

empty [
	let-set old-print [get print]
	let-set print {old-print [get 0]; old-print [get 0]}
	iter [get list] [get print]
	empty iter [get list] print # weird how that works.
]

let-set map {
	let-set list [get 0]
	let-set func [get 1]
	if [= [get list] end] {
		value end
	} {
		pair [func [list left]] [map [list right] [get func]]
	}
}

print

iter [map [get list] {+ [get 0] [get 0]}] [get print]

let-set e {}
let-set e2 []
