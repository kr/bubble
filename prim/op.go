// Primitive operations
package prim

type Op int

// New items in this list must also be added to
// the type switches in ../naivegen/prim.go
// and ../sem/prim.go.
const (
	Println Op = iota
	Add
	Lt
	Ineq
)

var opNames = [...]string{
	Println: "Println",
	Add:     "Add",
	Lt:      "Lt",
	Ineq:    "Ineq",
}

var opNArg = [...]int{
	Println: 1,
	Add:     2,
	Lt:      2,
	Ineq:    2,
}

var opNRes = [...]int{
	Println: 0,
	Add:     1,
	Lt:      0,
	Ineq:    0,
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
