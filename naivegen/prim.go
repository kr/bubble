package naivegen

import "github.com/kr/bubble/prim"

// return value placed in its own statement block
func genPrim(op prim.Op, dl, wl, cl []string) string {
	switch op {
	case prim.Println:
		return `console.log.apply(null, ` + dl[0] + `);` + cl[0]
	case prim.Add:
		return `var ` + wl[0] + ` = ` + dl[0] + ` + ` + dl[1] + `;` + cl[0]
	case prim.Mul:
		return `var ` + wl[0] + ` = ` + dl[0] + ` * ` + dl[1] + `;` + cl[0]
	case prim.Lt:
		return `if (` + dl[0] + ` < ` + dl[1] + `) { ` + cl[0] + ` } else { ` + cl[1] + ` }`
	case prim.Ineq:
		return `if (` + dl[0] + ` !== ` + dl[1] + `) { ` + cl[0] + ` } else { ` + cl[1] + ` }`
	}
	panic("unreached")
}
