// Primitive operations
package prim

type Op int

// New items in this list must also be added to
// the type switches in ../naivegen/prim.go
// and ../sem/prim.go.
const (
	Println Op = iota
	Add
	Mul
	Lt
	Ineq
	Callcc
)

var opNames = [...]string{
	Println: "Println",
	Add:     "Add",
	Mul:     "Mul",
	Lt:      "Lt",
	Ineq:    "Ineq",
	Callcc:  "Callcc",
}

var opNArg = [...]int{
	Println: 1,
	Add:     2,
	Mul:     2,
	Lt:      2,
	Ineq:    2,
	Callcc:  -1, // unused; special case in ../cps/conf.go
}

var opNRes = [...]int{
	Println: 0,
	Add:     1,
	Mul:     1,
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
	return opNArg[o]
}

// NArg returns the number of results this operation yields.
func (o Op) NRes() int {
	return opNRes[o]
}

// Pure returns whether o has no side effects.
func (o Op) Pure() bool {
	return int(o) < len(opPure) && opPure[o]
}
