package optimizer

import "github.com/kr/bubble/cps"

func etaReduce1(exp cps.Exp) cps.Exp {
	return cps.Map(exp, func(exp cps.Exp) cps.Exp {
		if fix, ok := exp.(cps.Fix); ok {
			var (
				fs []cps.FixEnt
				cs []cps.Var
				vs []cps.Value
			)
			for _, ent := range fix.Fs {
				if etaRedex(ent) {
					cs = append(cs, ent.V)
					vs = append(vs, ent.B.(cps.App).F)
				} else {
					fs = append(fs, ent)
				}
			}
			if len(fs) == 0 {
				return subVars(fix.E, cs, vs)
			}
			fix.Fs = fs
			return subVars(fix, cs, vs)
		}
		return exp
	})
}

// Returns whether ent is an η-redex.
func etaRedex(ent cps.FixEnt) bool {
	app, ok := ent.B.(cps.App)
	if !ok || len(ent.A) != len(app.Vs) {
		return false
	}
	for i := range ent.A {
		if v, ok := app.Vs[i].(cps.Var); !ok || v != ent.A[i] {
			return false
		}
	}
	return true
}
