package optimizer

import "github.com/kr/bubble/cps"

func deadVar1(exp cps.Exp) cps.Exp {
	return cps.Map(exp, func(exp cps.Exp) cps.Exp {
		switch exp := exp.(type) {
		case cps.Fix:
			var fs []cps.FixEnt
			for _, ent := range exp.Fs {
				if isFree(exp, ent.V) {
					fs = append(fs, ent)
				}
			}
			if len(fs) == 0 {
				return exp.E
			}
			return cps.Fix{Fs: fs, E: exp.E}
		case cps.Primop:
			if !exp.Op.Pure() {
				return exp
			}
			free := false
			for _, w := range exp.Ws {
				isFree(exp, w)
			}
			if !free && len(exp.Es) == 1 {
				return exp.Es[0]
			}
		case cps.Record:
			if !isFree(exp, exp.W) {
				return exp.E
			}
		case cps.Select:
			if !isFree(exp, exp.W) {
				return exp.E
			}
		}
		return exp
	})
}

func isFree(exp cps.Exp, w cps.Var) bool {
	occur := false
	cps.WalkValues(exp, func(v cps.Value) {
		if r, ok := v.(cps.Var); ok && r == w {
			occur = true
		}
	})
	return occur
}
