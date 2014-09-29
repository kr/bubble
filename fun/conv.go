package fun

import (
	"log"
	"strconv"

	"go/token"

	"github.com/kr/bubble/ast"
	"github.com/kr/bubble/prim"
)

func Convert(p *ast.Program) Exp {
	r := env0
	r = bindlit(r, "false", Int(0))
	r = bindlit(r, "true", Int(1))
	r = bindlit(r, "println", Prim(prim.Println))
	return conv(p, r)
}

// env maps names to Var identities
// it keeps track of lexical scope
type env func(name *ast.Ident) Exp

func env0(name *ast.Ident) Exp {
	panic("undefined: " + name.Name)
}

// e must be a sanitary expression
// (e.g. literal constant int or operator)
type symbol struct {
	v *Var
	e Exp
}

func bindlit(r env, name string, e Exp) env {
	return bind(r, name, symbol{e: e})
}

// bindvar augments r with a newly introduced *Var bound to name.
func bindvar(r env, name *ast.Ident) (env, *Var) {
	v := new(Var)
	v.Ident = name.Name
	return bind(r, name.Name, symbol{v: v}), v

}

func bind(r env, name string, sym symbol) env {
	return func(get *ast.Ident) Exp {
		if get.Name == name {
			if sym.v != nil {
				return sym.v
			}
			return sym.e
		}
		return r(get)
	}
}

func conv(node ast.Node, r env) Exp {
	switch node := node.(type) {
	case *ast.Ident:
		return r(node)
	case *ast.BasicLit:
		return convlit(node.Kind, node.Value)
	case *ast.CallExpr:
		switch len(node.Args) {
		case 0:
			return App{conv(node.Fun, r), Int(0)}
		case 1:
			return App{conv(node.Fun, r), conv(node.Args[0], r)}
		default:
			return App{conv(node.Fun, r), Record(convl(node.Args, r))}
		}
	case *ast.BinaryExpr:
		el := []Exp{conv(node.X, r), conv(node.Y, r)}
		return App{convprim(node.Op), Record(el)}
	case *ast.FuncLit:
		return Fn{new(Var), conv(node.Body, r)}
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
	case *ast.Program:
		var fl []*ast.FuncDecl
		for _, file := range node.Files {
			fl = append(fl, file.Funcs...)
		}
		main := &ast.CallExpr{&ast.Ident{"main"}, nil}
		return convfix(fl, main, r)
	case *ast.ShortFuncLit:
		params := []*ast.Ident{{"x"}, {"y"}, {"z"}}
		return convfunc(params, node.Body, r)
	default:
		log.Printf("%T", node)
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
	return App{Fn{new(Var), convseq(sl[1:], r)}, conv(sl[0], r)}
}

func convl(xl []ast.Expr, r env) (el []Exp) {
	for _, x := range xl {
		el = append(el, conv(x, r))
	}
	return el
}

func convfix(fl []*ast.FuncDecl, exp ast.Node, r env) Exp {
	var fix Fix
	for _, f := range fl {
		var v *Var
		r, v = bindvar(r, f.Name)
		fix.Names = append(fix.Names, v)
	}
	fix.Body = conv(exp, r)
	for _, f := range fl {
		fix.Fns = append(fix.Fns, convfunc(f.Params, f.Body, r))
	}
	return fix
}

func convfunc(params []*ast.Ident, body ast.Node, r env) Fn {
	switch len(params) {
	case 0:
		return Fn{new(Var), conv(body, r)}
	case 1:
		r, v := bindvar(r, params[0])
		return Fn{v, conv(body, r)}
	default:
		v := new(Var)
		var pl []*Var
		for _, s := range params {
			var p *Var
			r, p = bindvar(r, s)
			pl = append(pl, p)
		}
		exp := conv(body, r)
		for i, p := range pl {
			exp = App{Fn{p, exp}, Select{i, v}}
		}
		return Fn{v, exp}
	}
}
