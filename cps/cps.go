package cps

import "github.com/kr/bubble/prim"

type Value interface {
	value()
}

func (*Label) value()    {}
func (Int) value()       {}
func (String) value()    {}
func (Undefined) value() {}
func (Var) value()       {}

type Exp interface {
	cexp()
	mapexp(func(Exp) Exp) Exp
	mapval(func(Value) Value) Exp
}

func (App) cexp()    {}
func (Fix) cexp()    {}
func (Primop) cexp() {}
func (Record) cexp() {}
func (Select) cexp() {}
func (Switch) cexp() {}

type Label struct{ byte }
type Int int
type String string
type Undefined struct{}
type Var struct {
	ID   uint // nonzero for a valid Var
	Name string
}

var Undef = Undefined{}

type App struct {
	F  Value
	Vs []Value
}

type FixEnt struct {
	V Var
	A []Var
	B Exp
}

type Fix struct {
	Fs []FixEnt
	E  Exp
}

type Primop struct {
	Op prim.Op
	Vs []Value
	Ws []Var
	Es []Exp
}

type RecordEnt struct {
	V    Value
	Path Path
}

// Record exp binds W in the scope of E
// to the record containing Vs.
type Record struct {
	Vs []RecordEnt
	W  Var
	E  Exp
}

type Path interface {
	path()
}

type Offp int

func (o Offp) path() {}

// Select binds W in the scope of E
// to the Ith field of record V.
type Select struct {
	I int
	V Value
	W Var
	E Exp
}

// Switch proceeds with the Ith element in Es.
type Switch struct {
	I  Value
	Es []Exp
}
