package cps

import (
	"log"

	"github.com/kr/bubble/fun"
	"github.com/kr/bubble/prim"
)

// Convert converts into CPS
// from the minimal functional language
// defined in package fun.
// Returns the converted expression
// and an "exit" address
// (which must be linked separately).
func Convert(exps []fun.Exp) (Exp, Var) {
	r := newVar("exit")
	return convseq(exps, func(v Value) Exp {
		return App{r, []Value{v}}
	}), r
}

func convseq(exps []fun.Exp, c func(Value) Exp) Exp {
	if len(exps) == 1 {
		return conv(exps[0], c)
	}
	return conv(exps[0], func(v Value) Exp {
		return convseq(exps[1:], c)
	})
}

func conv(exp fun.Exp, c func(Value) Exp) Exp {
	switch exp := exp.(type) {
	case fun.Var:
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
			x := newVar("")
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
		return conv(exp.Rec, func(v Value) Exp {
			w := newVar("")
			return Select{exp.I, v, w, c(w)}
		})
	case fun.Switch:
		if isBool(exp) {
			return conv(exp.Value, func(v Value) Exp {
				k := newVar("")
				x := newVar("")
				return Fix{
					[]FixEnt{
						{k, []Var{x}, c(x)},
					},
					Primop{
						Op: prim.Ineq,
						Vs: []Value{v, Int(0)},
						Ws: nil,
						Es: []Exp{
							conv(exp.Default, func(z Value) Exp {
								return App{k, []Value{z}}
							}),
							conv(exp.Cases[0].Body, func(z Value) Exp {
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
			case op == prim.Callcc:
				k := newVar("")
				x := newVar("")
				kp := newVar("")
				xp := newVar("")
				wp := newVar("")
				return Fix{
					[]FixEnt{
						{k, []Var{x}, c(x)},
						{
							kp,
							[]Var{xp, newVar("")},
							Select{0, xp, wp, App{k, []Value{wp}}},
						},
					},
					conv(exp.V, func(v Value) Exp {
						w := newVar("")
						r := newVar("")
						return Select{0, v, w,
							Record{
								[]RecordEnt{{kp, Offp(0)}},
								r,
								App{w, []Value{r, k}},
							},
						}
					}),
				}
			case op.NArg() == 1 && op.NRes() == 0:
				return conv(exp.V, func(v Value) Exp {
					return Primop{
						op,
						[]Value{v},
						[]Var{},
						[]Exp{c(Int(0))},
					}
				})
			case op.NArg() == 1 && op.NRes() == 1:
				return conv(exp.V, func(v Value) Exp {
					w := newVar("")
					return Primop{
						op,
						[]Value{v},
						[]Var{w},
						[]Exp{c(w)},
					}
				})
			case op.NArg() > 1 && op.NRes() == 1:
				switch A := exp.V.(type) {
				case fun.Record:
					return fl(A, func(vs []Value) Exp {
						w := newVar("")
						return Primop{
							op,
							vs,
							[]Var{w},
							[]Exp{c(w)},
						}
					})
				default:
					panic("not implemented")
				}
			}
		default:
			r := newVar("")
			x := newVar("")
			return Fix{
				[]FixEnt{
					{r, []Var{x}, c(x)},
				},
				conv(exp.F, func(f Value) Exp {
					return conv(exp.V, func(e Value) Exp {
						return App{f, []Value{e, r}}
					})
				}),
			}
		}
	case fun.Fix:
		return Fix{
			fixfnl(exp.Names, exp.Fns),
			conv(exp.Body, c),
		}
	case fun.Fn:
		f := newVar("")
		k := newVar("")
		return Fix{
			[]FixEnt{
				{f, []Var{cpsvar(exp.V), k}, conv(exp.Body, func(z Value) Exp {
					return App{k, []Value{z}}
				})},
			},
			c(f),
		}
	}
	log.Fatalf("unhandled %T", exp)
	panic("unreached")
}

func fixfnl(h []fun.Var, b []fun.Fn) (vs []FixEnt) {
	if len(h) != len(b) {
		panic("mismatch")
	}
	for i := range h {
		f := b[i]
		w := newVar("")
		vs = append(vs, FixEnt{
			cpsvar(h[i]),
			[]Var{cpsvar(f.V), w},
			conv(f.Body, func(z Value) Exp {
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
		return conv(expl[0], func(v Value) Exp {
			return g(expl[1:], append(w, v))
		})
	}
	return g(expl, nil)
}

var nextVar uint

func newVar(name string) Var {
	nextVar++
	return Var{ID: nextVar, Name: name}
}

var cpsvars = map[uint]Var{}

func cpsvar(v fun.Var) Var {
	v1, ok := cpsvars[v.ID]
	if ok {
		return v1
	}
	v1 = newVar(v.Name)
	cpsvars[v.ID] = v1
	return v1
}

func isBool(s fun.Switch) bool {
	return len(s.Cases) == 1 &&
		s.Cases[0].Value == fun.IntCon(0) &&
		s.Default != nil
}
