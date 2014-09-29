// Package fun is a small functional intermediate representation
// for Bubble programs. It is analogous to the "mini-ML" lambda
// language in the Standard ML of New Jersey compiler.
package fun

import (
	"go/token"

	"github.com/kr/bubble/prim"
)

type Exp interface {
	exp()
}

func (*Var) exp()   {}
func (App) exp()    {}
func (Fix) exp()    {}
func (Fn) exp()     {}
func (Int) exp()    {}
func (Prim) exp()   {}
func (Record) exp() {}
func (Select) exp() {}
func (String) exp() {}
func (Switch) exp() {}

type Var struct{ Ident string }

type Int int

type String string

type Fn struct {
	V    *Var
	Body Exp
}

type Fix struct {
	Names []*Var
	Fns   []Fn
	Body  Exp
}

type App struct {
	F Exp
	V Exp
}

type Case struct {
	Value Con
	Body  Exp
}

type Switch struct {
	Value   Exp
	Poss    []Conrep
	Cases   []Case
	Default Exp // can be nil
}

type Prim prim.Op

var primOps = [...]prim.Op{
	token.ADD: prim.Add,
	token.MUL: prim.Mul,
}

type Record []Exp

type Select struct {
	I   int
	Rec Exp
}

type Path interface {
	path()
}

func (o Offp) path() {}
func (s Selp) path() {}

type Offp int

type Selp struct {
	I int
	P Path
}

type Conrep interface {
	conrep()
}

func (c ConrepUnit) conrep() {}
func (t Tagged) conrep()     {}
func (c Constant) conrep()   {}

type ConrepUnit int

const (
	Undecided ConrepUnit = iota
)

type Tagged int

type Constant int

type Con interface {
	con()
}

func (IntCon) con() {}

type IntCon int
