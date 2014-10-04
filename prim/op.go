// Primitive operations
package prim

import "log"

type Op int

// New items in this list must also be added to
// the type switches in ../naivegen/prim.go
// and ../sem/prim.go.
const (
	invalid Op = iota
	Println
	Add
	Sub
	Mul
	Quo
	Lt
	Ineq
	Callcc
)

var opNames = [...]string{
	invalid: "invalid",
	Println: "Println",
	Add:     "Add",
	Sub:     "Sub",
	Mul:     "Mul",
	Quo:     "Quo",
	Lt:      "Lt",
	Ineq:    "Ineq",
	Callcc:  "Callcc",
}

var opNArg = [...]int{
	invalid: -1,
	Println: 1,
	Add:     2,
	Sub:     2,
	Mul:     2,
	Quo:     2,
	Lt:      2,
	Ineq:    2,
	Callcc:  -1, // unused; special case in ../cps/conf.go
}

var opNRes = [...]int{
	invalid: -1,
	Println: 0,
	Add:     1,
	Sub:     1,
	Mul:     1,
	Quo:     1,
	Lt:      0,
	Ineq:    0,
	Callcc:  -1, // unused; special case in ../cps/conf.go
}

var opPure = [...]bool{
	Lt:   true,
	Ineq: true,
}

func (o Op) String() string {
	return opNames[o]
}

func (o Op) GoString() string {
	return "prim." + opNames[o]
}

// NArg returns the number of arguments this operation takes.
func (o Op) NArg() int {
	if o == invalid {
		log.Fatal("invalid op")
	}
	return opNArg[o]
}

// NArg returns the number of results this operation yields.
func (o Op) NRes() int {
	if o == invalid {
		log.Fatal("invalid op")
	}
	return opNRes[o]
}

// Pure returns whether o has no side effects.
func (o Op) Pure() bool {
	if o == invalid {
		log.Fatal("invalid op")
	}
	return int(o) < len(opPure) && opPure[o]
}
