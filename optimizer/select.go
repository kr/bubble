package optimizer

import "github.com/kr/bubble/cps"

// Replaces select expressions with the record fields
// being selected when they can be determined statically.
func selectFold1(exp cps.Exp) cps.Exp {
	return cps.Map(exp, func(exp cps.Exp) cps.Exp {
		if rec, ok := exp.(cps.Record); ok {
			rec.E = cps.Map(rec.E, func(exp cps.Exp) cps.Exp {
				sel, ok := exp.(cps.Select)
				if !ok {
					return exp
				}
				if v, ok := sel.V.(cps.Var); ok && v == rec.W {
					comb := combineRecSel(rec, sel)
					return comb
				}
				return exp
			})
			return rec
		}
		return exp
	})
}

func combineRecSel(r cps.Record, s cps.Select) cps.Exp {
	ent := cps.RecordEnt{cps.Undef, cps.Offp(0)}
	if s.I >= 0 && s.I < len(r.Vs) {
		ent = r.Vs[s.I]
	}
	if p, ok := ent.Path.(cps.Offp); !ok || p != cps.Offp(0) {
		return s
	}
	return subVars(s.E, []cps.Var{s.W}, []cps.Value{ent.V})
}
