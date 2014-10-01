package optimizer

import (
	"reflect"

	"github.com/kr/bubble/cps"
)

// Repeatedly performs 𝛽-contraction on exp
// until no more can be performed.
func betaCon(exp cps.Exp) cps.Exp {
	for {
		E1 := betaCon1(exp)
		if reflect.DeepEqual(E1, exp) {
			return exp
		}
		exp = E1
	}
}

// Performs a single 𝛽-contraction on exp if possible,
// or returns exp unchanged.
func betaCon1(exp cps.Exp) cps.Exp {
	for f, s := range countfns(exp) {
		if s.napp == 1 && s.noccur == 1 && s.V.ID != 0 {
			return delFixent(betaReduce(exp, f, s.A, s.B), f)
		}
	}
	return exp
}

type fncount struct {
	cps.FixEnt
	noccur int
	napp   int
}

// For each Var in exp, countfns records:
// the number of times it occurs as a Value,
// the number of times it appears in function position,
// and the Fix entry where it is bound, if any.
func countfns(exp cps.Exp) map[cps.Var]fncount {
	fntab := make(map[cps.Var]fncount)
	cps.WalkValues(exp, func(v cps.Value) {
		if v, ok := v.(cps.Var); ok {
			fn := fntab[v]
			fn.noccur++
			fntab[v] = fn
		}
	})
	cps.Walk(exp, func(exp cps.Exp) {
		switch exp := exp.(type) {
		case cps.App:
			if v, ok := exp.F.(cps.Var); ok {
				fn := fntab[v]
				fn.napp++
				fntab[v] = fn
			}
		case cps.Fix:
			for _, f := range exp.Fs {
				fn := fntab[f.V]
				fn.FixEnt = f
				fntab[f.V] = fn
			}
		}
	})
	return fntab
}

// Replaces application of f in E with B,
// substituting the actual arguments
// for occurrences of the formal parameters vs.
func betaReduce(E cps.Exp, f cps.Var, vs []cps.Var, B cps.Exp) cps.Exp {
	return cps.Map(E, func(exp cps.Exp) cps.Exp {
		if app, ok := exp.(cps.App); ok && app.F == f {
			return subVars(B, vs, app.Vs)
		}
		return exp
	})
}

// Deletes the definition of f where it occurs in a Fix.
func delFixent(exp cps.Exp, f cps.Var) cps.Exp {
	return cps.Map(exp, func(exp cps.Exp) cps.Exp {
		if fix, ok := exp.(cps.Fix); ok {
			var fs []cps.FixEnt
			for _, fent := range fix.Fs {
				if fent.V != f {
					fs = append(fs, fent)
				}
			}
			if len(fs) == 0 {
				return fix.E
			}
			return cps.Fix{Fs: fs, E: fix.E}
		}
		return exp
	})
}
