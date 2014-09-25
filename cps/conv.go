package cps

import (
	"fmt"

	"github.com/kr/bubble/fun"
	"github.com/kr/bubble/prim"
)

// Convert converts into CPS from the minimal functional
// language defined in package fun.
func Convert(exp fun.Exp, c func(Value) Exp) Exp {
	switch exp := exp.(type) {
	case *fun.Var:
		return c(cpsvar(exp))
	case fun.Int:
		return c(Int(exp))
	case fun.String:
		return c(String(exp))
	case fun.Prim:
		panic("not implemented")
	case fun.Record:
		if len(exp) == 0 {
			return c(Int(0))
		}
		return fl(exp, func(vs []Value) Exp {
			x := new(Var)
			r := Record{W: x, E: c(x)}
			for _, v := range vs {
				r.Vs = append(r.Vs, struct {
					V    Value
					Path Path
				}{v, Offp(0)})
			}
			return r
		})
	case fun.Select:
		return Convert(exp.Rec, func(v Value) Exp {
			w := new(Var)
			return Select{exp.I, v, w, c(w)}
		})
	case fun.Switch:
		if isBool(exp) {
			return Convert(exp.Value, func(v Value) Exp {
				k := new(Var)
				x := new(Var)
				return Fix{
					[]FixEnt{
						{k, []*Var{x}, c(x)},
					},
					Primop{
						Op: prim.Ineq,
						Vs: []Value{v, Int(0)},
						Ws: nil,
						Es: []Exp{
							Convert(exp.Default, func(z Value) Exp {
								return App{k, []Value{z}}
							}),
							Convert(exp.Cases[0].Body, func(z Value) Exp {
								return App{k, []Value{z}}
							}),
						},
					},
				}
			})
		} else {
			panic("not implemented")
		}
	case fun.App:
		switch f := exp.F.(type) {
		case fun.Prim:
			op := prim.Op(f)
			switch {
			case op.NArg() == 1 && op.NRes() == 0:
				return Convert(exp.V, func(v Value) Exp {
					return Primop{
						op,
						[]Value{v},
						[]*Var{},
						[]Exp{c(Int(0))},
					}
				})
			case op.NArg() == 1 && op.NRes() == 1:
				return Convert(exp.V, func(v Value) Exp {
					w := new(Var)
					return Primop{
						op,
						[]Value{v},
						[]*Var{w},
						[]Exp{c(w)},
					}
				})
			case op.NArg() > 1 && op.NRes() == 1:
				switch A := exp.V.(type) {
				case fun.Record:
					return fl(A, func(vs []Value) Exp {
						w := new(Var)
						return Primop{
							op,
							vs,
							[]*Var{w},
							[]Exp{c(w)},
						}
					})
				default:
					panic("not implemented")
				}
			}
		default:
			r := new(Var)
			x := new(Var)
			return Fix{
				[]FixEnt{
					{r, []*Var{x}, c(x)},
				},
				Convert(exp.F, func(f Value) Exp {
					return Convert(exp.V, func(e Value) Exp {
						return App{f, []Value{e, r}}
					})
				}),
			}
		}
	case fun.Fix:
		return Fix{
			fixfnl(exp.Names, exp.Fns),
			Convert(exp.Body, c),
		}
	case fun.Fn:
		f := new(Var)
		k := new(Var)
		return Fix{
			[]FixEnt{
				{f, []*Var{cpsvar(exp.V), k}, Convert(exp.Body, func(z Value) Exp {
					return App{k, []Value{z}}
				})},
			},
			c(f),
		}
	}
	panic("unreached " + fmt.Sprintf("%T", exp))
}

func fixfnl(h []*fun.Var, b []fun.Fn) (vs []FixEnt) {
	if len(h) != len(b) {
		panic("mismatch")
	}
	for i := range h {
		f := b[i]
		w := new(Var)
		vs = append(vs, FixEnt{
			cpsvar(h[i]),
			[]*Var{cpsvar(f.V), w},
			Convert(f.Body, func(z Value) Exp {
				return App{w, []Value{z}}
			}),
		})
	}
	return vs
}

func fl(expl []fun.Exp, c func([]Value) Exp) Exp {
	var g func(expl []fun.Exp, w []Value) Exp
	g = func(expl []fun.Exp, w []Value) Exp {
		if len(expl) == 0 {
			return c(w)
		}
		return Convert(expl[0], func(v Value) Exp {
			return g(expl[1:], append(w, v))
		})
	}
	return g(expl, nil)
}

var cpsvars = map[*fun.Var]*Var{}

func cpsvar(v *fun.Var) *Var {
	v1, ok := cpsvars[v]
	if ok {
		return v1
	}
	v1 = new(Var)
	v1.Ident = v.Ident
	cpsvars[v] = v1
	return v1
}

func isBool(s fun.Switch) bool {
	return len(s.Cases) == 1 &&
		s.Cases[0].Value == fun.IntCon(0) &&
		s.Default != nil
}
