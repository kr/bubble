// Interpreter for the CPS notation
// derived from the denotational semantics.
package sem

import (
	"fmt"
	"os"

	"github.com/kr/bubble/cps"
)

// A "denotable" value.
type dvalue interface {
	dvalue()
}

func (record) dvalue()  {}
func (dint) dvalue()    {}
func (dstring) dvalue() {}
func (fn) dvalue()      {}

type dint int

type dstring string

type record struct {
	L []dvalue
	I int
}

type fn func([]dvalue) thunk

type thunk func(store) answer

type answer interface{}

type store interface{}

type env func(v *cps.Var) dvalue

func bind(env env, v *cps.Var, d dvalue) env {
	return func(w *cps.Var) dvalue {
		if v == w {
			return d
		}
		return env(w)
	}
}

func bindn(env env, vl []*cps.Var, dl []dvalue) env {
	if len(vl) == 0 && len(dl) == 0 {
		return env
	}
	return bindn(bind(env, vl[0], dl[0]), vl[1:], dl[1:])
}

func env0(v *cps.Var) dvalue {
	panic(fmt.Errorf("undefined var %p", v))
}

func val(env env, v cps.Value) dvalue {
	switch v := v.(type) {
	case cps.Int:
		return dint(v)
	case cps.String:
		return dstring(v)
	case *cps.Var:
		return env(v)
	}
	fmt.Printf("val %T\n", v)
	panic("unreached")
}

func fetch(x dvalue, path cps.Path) dvalue {
	if p, ok := path.(cps.Offp); ok && p == 0 {
		return x
	}
	r := x.(record)
	switch p := path.(type) {
	case cps.Offp:
		return record{r.L, r.I + int(p)}
	}
	panic("unreached")
}

func eval(exp cps.Exp, r env) thunk {
	switch exp := exp.(type) {
	case cps.Record:
		var vs []dvalue
		for _, x := range exp.Vs {
			vs = append(vs, fetch(val(r, x.V), x.Path))
		}
		return eval(exp.E, bind(r, exp.W, record{vs, 0}))
	case cps.Primop:
		var dl []dvalue
		for _, v := range exp.Vs {
			dl = append(dl, val(r, v))
		}
		var cl []fn
		for _, e := range exp.Es {
			e := e
			cl = append(cl, func(al []dvalue) thunk {
				return eval(e, bindn(r, exp.Ws, al))
			})
		}
		return evalPrim(exp.Op, dl, cl)
	case cps.App:
		g := val(r, exp.F).(fn)
		var dl []dvalue
		for _, v := range exp.Vs {
			dl = append(dl, val(r, v))
		}
		return g(dl)
	case cps.Fix:
		var g func(env) env
		h := func(r1 env, f struct {
			V *cps.Var
			A []*cps.Var
			B cps.Exp
		}) dvalue {
			return fn(func(al []dvalue) thunk {
				return eval(f.B, bindn(g(r1), f.A, al))
			})
		}
		g = func(r env) env {
			var nl []*cps.Var
			for _, v := range exp.Fs {
				nl = append(nl, v.V)
			}
			var dl []dvalue
			for _, v := range exp.Fs {
				dl = append(dl, h(r, v))
			}
			return bindn(r, nl, dl)
		}
		return eval(exp.E, g(r))
	case cps.Select:
		rec := val(r, exp.V).(record)
		return eval(exp.E, bind(r, exp.W, rec.L[exp.I+rec.I]))
	}
	fmt.Printf("exp %T\n", exp)
	panic("unreached")
}

func Eval(exp cps.Exp, r *cps.Var) answer {
	f := fn(func(dl []dvalue) thunk {
		return func(store) answer {
			return dl[0]
		}
	})
	bootstrapOut = os.Stdout
	var st store
	return eval(exp, bind(env0, r, f))(st)
}
