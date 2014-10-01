package optimizer

import "github.com/kr/bubble/cps"

func tailCall1(exp cps.Exp) cps.Exp {
	return cps.Map(exp, func(exp cps.Exp) cps.Exp {
		if fix, ok := exp.(cps.Fix); ok {
			var (
				fs []cps.FixEnt
				cs []cps.Var
				vs []cps.Value
			)
			for _, ent := range fix.Fs {
				if v, ok := tailCont(ent); ok {
					cs = append(cs, ent.V)
					vs = append(vs, v)
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

// Returns the further continuation
// of the tail call continuation in ent, if any,
// and whether it is a tail call continuation.
func tailCont(ent cps.FixEnt) (cps.Value, bool) {
	app, ok := ent.B.(cps.App)
	if !ok || len(ent.A) != 1 || len(app.Vs) != 1 {
		return cps.Var{}, false
	}
	if v, ok := app.Vs[0].(cps.Var); ok && v == ent.A[0] {
		return app.F, true
	}
	return cps.Var{}, false
}
