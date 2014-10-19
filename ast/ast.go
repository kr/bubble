package ast

import (
	"go/token"
	"strconv"
	"unicode"
	"unicode/utf8"
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
func (*Package) node()      {}
func (*ReturnStmt) node()   {}
func (*SelectorExpr) node() {}
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
func (*SelectorExpr) exp() {}
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

type Package struct {
	Name  string
	Files []*File
}

type File struct {
	Name    *Ident
	Imports []*ImportSpec
	Funcs   []*FuncDecl
}

type IfStmt struct {
	Cond Expr
	Body *BlockStmt
	Else Stmt // BlockStmt, IfStmt, or nil
}

type ImportSpec struct {
	Name *Ident    // maybe nil
	Path *BasicLit // import path (always a string)
}

func (is *ImportSpec) ImportPath() string {
	return is.Path.String()
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
	V Expr
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

// Returns the unquoted string value represented by b.
// Result is undefined if b is not a string.
func (b *BasicLit) String() string {
	s, _ := strconv.Unquote(b.Value)
	return s
}

type Ident struct {
	Name string
}

func (id *Ident) IsExported() bool { return IsExported(id.Name) }

// Symbols starting with anything other than
// a lower case letter and _ are exported.
// (This means upper case and case-less characters.)
func IsExported(name string) bool {
	ch, _ := utf8.DecodeRuneInString(name)
	return !unicode.IsLower(ch) && ch != '_'
}

type AssignStmt struct {
	Lhs Expr
	Tok token.Token
	Rhs Expr
}

type ExprStmt struct {
	X Expr
}

type SelectorExpr struct {
	X   Expr
	Sel *Ident
}
