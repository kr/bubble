package sem

import (
	"fmt"
	"io"

	"github.com/kr/bubble/prim"
)

var bootstrapOut io.Writer

func evalPrim(op prim.Op, dl []dvalue, cl []fn) thunk {
	switch op {
	case prim.Println:
		var vl []interface{}
		for _, d := range dl {
			vl = append(vl, d)
		}
		fmt.Fprintln(bootstrapOut, vl...)
		return cl[0](nil)
	case prim.Add:
		i := dl[0].(dint)
		j := dl[1].(dint)
		return cl[0]([]dvalue{i + j})
	case prim.Mul:
		i := dl[0].(dint)
		j := dl[1].(dint)
		return cl[0]([]dvalue{i * j})
	case prim.Lt:
	case prim.Ineq:
		if eq(dl[0], dl[1]) {
			return cl[1](nil)
		} else {
			return cl[0](nil)
		}
	}
	panic("unreached")
}

func eq(a, b dvalue) bool {
	switch a := a.(type) {
	case dint:
		return int(a) == int(b.(dint))
	}
	panic("unreached")
}
