package fun

import (
	"log"
	"strconv"

	"go/token"

	"github.com/kr/bubble/ast"
	"github.com/kr/bubble/prim"
)

// exported symbol table for a package
type Tab struct {
	name string
	sym  map[string]Var
}

// Convert converts p to a functional expression.
// Function pkgtab must return the symbol table
// from a previous call to Convert
// for any package imported by p.
func Convert(p *ast.Package, pkgtab func(importPath string) Tab) (Exp, Tab) {
	tab := Tab{p.Name, make(map[string]Var)}
	fix := Fix{Body: Int(0)}
	var inits []Var
	r := globalEnv

	// first define the environment r containing all funcs
	for _, file := range p.Files {
		for _, f := range file.Funcs {
			var v Var
			if f.Name.Name == "init" {
				v = newVar("init") // do not bind init
				inits = append(inits, v)
			} else {
				r, v = bindvar(r, f.Name)
			}
			if p.Name == "main" && f.Name.Name == "main" {
				fix.Body = App{v, Int(0)}
			}
			if f.Name.IsExported() {
				tab.sym[f.Name.Name] = v
			}
			fix.Names = append(fix.Names, v)
		}
	}

	// then convert the funcs using r
	for _, file := range p.Files {
		r1 := bindimports(r, file.Imports, pkgtab)
		for _, f := range file.Funcs {
			fix.Fns = append(fix.Fns, convfunc(f.Params, f.Body, r1))
		}
	}

	for _, f := range inits {
		fix.Body = App{Fn{newVar(""), fix.Body}, App{f, Int(0)}}
	}
	return fix, tab
}

func conv(node ast.Node, r env) Exp {
	switch node := node.(type) {
	case *ast.Ident:
		v := r(node.Name)
		if _, ok := v.(pkg); ok {
			log.Fatal("use of package " + node.Name + " without selector")
		}
		return v
	case *ast.BasicLit:
		return convlit(node.Kind, node.Value)
	case *ast.CallExpr:
		return App{conv(node.Fun, r), Record(convl(node.Args, r))}
	case *ast.BinaryExpr:
		el := []Exp{conv(node.X, r), conv(node.Y, r)}
		return App{convprim(node.Op), Record(el)}
	case *ast.FuncLit:
		return convfunc(node.Params, node.Body, r)
	case *ast.BlockStmt:
		return convseq(node.List, r)
	case *ast.IfStmt:
		var alt Exp = Int(0)
		if node.Else != nil {
			alt = conv(node.Else, r)
		}
		return Switch{
			Value:   conv(node.Cond, r),
			Cases:   []Case{{IntCon(0), alt}},
			Default: conv(node.Body, r),
		}
	case *ast.ExprStmt:
		return conv(node.X, r)
	case *ast.ReturnStmt:
		return App{r("return"), Record{conv(node.V, r)}}
	case *ast.SelectorExpr:
		// If node.X is a package, don't call conv.
		// A package is not a valid expression.
		if id, ok := node.X.(*ast.Ident); ok {
			if p, ok := r(id.Name).(pkg); ok {
				return p.tab.sym[node.Sel.Name]
			}
		}
		log.Fatalf("cannot select from non-package %v", node.X)
	case *ast.ShortFuncLit:
		params := []*ast.Ident{{"x"}, {"y"}, {"z"}}
		return convfunc(params, node.Body, r)
	default:
		log.Fatalf("unhandled %T", node)
	}
	panic("unreached")
}

func convprim(kind token.Token) Exp {
	return Prim(primOps[kind])
}

func convlit(kind token.Token, s string) Exp {
	switch kind {
	case token.INT:
		v, err := strconv.ParseInt(s, 0, 0)
		if err != nil {
			panic("bad int literal: " + s)
		}
		return Int(v)
	case token.STRING:
		t, err := strconv.Unquote(s)
		if err != nil {
			panic("bad string literal: " + s)
		}
		return String(t)
	}
	panic("bad lit token")
}

func convseq(sl []ast.Stmt, r env) Exp {
	if len(sl) == 0 {
		return Int(0)
	}
	return App{Fn{newVar(""), convseq(sl[1:], r)}, conv(sl[0], r)}
}

func convl(xl []ast.Expr, r env) (el []Exp) {
	for _, x := range xl {
		el = append(el, conv(x, r))
	}
	return el
}

func convfunc(params []*ast.Ident, body ast.Node, r env) Fn {
	v := newVar("")
	var pl []Var
	for _, s := range params {
		var p Var
		r, p = bindvar(r, s)
		pl = append(pl, p)
	}
	exp := convfuncbody(body, r)
	for i, p := range pl {
		exp = App{Fn{p, exp}, Select{i, v}}
	}
	return Fn{v, exp}
}

// save continuation as "return", evaluate body
func convfuncbody(body ast.Node, r env) Exp {
	rec := newVar("")
	ret := newVar("")
	r = bind(r, "return", ret)
	return App{Prim(prim.Callcc), Record{Fn{rec,
		App{Fn{ret, conv(body, r)}, Select{0, rec}},
	}}}
}

var globalEnv env

func init() {
	r := env0
	r = bind(r, "false", Int(0))
	r = bind(r, "true", Int(1))
	r = bind(r, "println", Prim(prim.Println))
	r = bind(r, "callcc", Prim(prim.Callcc))
	globalEnv = r
}

// env maps names to Var identities
// it keeps track of lexical scope
type env func(name string) Value

func env0(name string) Value {
	panic("undefined: " + name)
}

func bind(r env, name string, v Value) env {
	return func(get string) Value {
		if get == name {
			return v
		}
		return r(get)
	}
}

// bindvar augments r with a newly introduced Var bound to name.
func bindvar(r env, name *ast.Ident) (env, Var) {
	v := newVar(name.Name)
	return bind(r, name.Name, v), v
}

// bindimports augments r with a binding
// for each package listed in a.
func bindimports(r env, a []*ast.ImportSpec, pkgtab func(string) Tab) env {
	for _, spec := range a {
		// TODO(kr): use local name from import spec
		dep := pkgtab(spec.ImportPath())
		r = bind(r, dep.name, pkg{dep})
	}
	return r
}
