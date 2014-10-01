package optimizer

import "github.com/kr/bubble/cps"

// Optimize transforms exp in various ways
// in an attempt to improve code size
// or execution speed.
func Optimize(exp cps.Exp) cps.Exp {
	exp = betaCon(exp)
	return exp
}

// Function subVars computes B{v⃗ ↦ a⃗}.
// It returns B with each occurrence of vl[i]
// replaced by the corresponding al[i].
func subVars(B cps.Exp, vl []cps.Var, al []cps.Value) cps.Exp {
	return cps.MapValues(B, func(v cps.Value) cps.Value {
		if v, ok := v.(cps.Var); ok {
			for i := range vl {
				if v == vl[i] {
					return al[i]
				}
			}
		}
		return v
	})
}
