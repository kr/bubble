package ast

import (
	"go/token"
)

type Node interface {
	node()
}

func (*AssignStmt) node()   {}
func (*BasicLit) node()     {}
func (*BinaryExpr) node()   {}
func (*BlockStmt) node()    {}
func (*CallExpr) node()     {}
func (*ExprStmt) node()     {}
func (*FuncDecl) node()     {}
func (*FuncLit) node()      {}
func (*Ident) node()        {}
func (*IfStmt) node()       {}
func (*Program) node()      {}
func (*ReturnStmt) node()   {}
func (*ShortFuncLit) node() {}

type Expr interface {
	Node
	exp()
}

func (*BasicLit) exp()     {}
func (*BinaryExpr) exp()   {}
func (*CallExpr) exp()     {}
func (*FuncDecl) exp()     {}
func (*FuncLit) exp()      {}
func (*Ident) exp()        {}
func (*ShortFuncLit) exp() {}

type Stmt interface {
	Node
	stmt()
}

func (*AssignStmt) stmt() {}
func (*BlockStmt) stmt()  {}
func (*ExprStmt) stmt()   {}
func (*IfStmt) stmt()     {}
func (*ReturnStmt) stmt() {}

type Program struct {
	Files []*File
}

type File struct {
	Funcs []*FuncDecl
}

type IfStmt struct {
	Cond Expr
	Body *BlockStmt
	Else Stmt // BlockStmt, IfStmt, or nil
}

type FuncDecl struct {
	Name   *Ident
	Params []*Ident
	Body   *BlockStmt
}

type ShortFuncLit struct {
	Body Expr
}

type FuncLit struct {
	Params []*Ident
	Body   *BlockStmt
}

type BlockStmt struct {
	List []Stmt
}

type Sequence struct {
	V []Expr
}

type ReturnStmt struct {
	Values []Expr
}

type CallExpr struct {
	Fun  Expr
	Args []Expr
}

type BinaryExpr struct {
	X  Expr
	Op token.Token
	Y  Expr
}

type BasicLit struct {
	Kind  token.Token
	Value string // literal string
}

type Ident struct {
	Name string
}

type AssignStmt struct {
	Lhs Expr
	Tok token.Token
	Rhs Expr
}

type ExprStmt struct {
	X Expr
}
