package optimizer

import (
	"reflect"

	"github.com/kr/bubble/cps"
)

// Optimize transforms exp in various ways
// in an attempt to improve code size
// or execution speed.
func Optimize(exp cps.Exp) cps.Exp {
	return fixedPoint(optimize1, exp)
}

var optimizers = []func(cps.Exp) cps.Exp{
	etaReduce1,
	betaCon1,
	selectFold1,
	deadVar1,
}

// optimize1 performs a single optimization pass:
// it applies each optimization function repeatedly
// until it produces no change, then moves on to
// the next function.
func optimize1(exp cps.Exp) cps.Exp {
	for _, f := range optimizers {
		exp = fixedPoint(f, exp)
	}
	return exp
}

// Finds the fixed point of f: iterates expᵢ₊₁ = f(expᵢ)
// until the result is unchanged.
func fixedPoint(f func(cps.Exp) cps.Exp, exp cps.Exp) cps.Exp {
	exp1 := f(exp)
	if reflect.DeepEqual(exp1, exp) {
		return exp
	}
	return fixedPoint(f, exp1)
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
