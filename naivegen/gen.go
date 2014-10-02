package naivegen

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/kr/bubble/cps"
)

const prelude = `
var F = [];
function drive(f0) {
	F.push(f0);
	while (F.length > 0) {
		var f = F.shift();
		var c = f[0].apply(null, f.slice(1));
		if (c.length > 0) {
			F.push(c);
		}
	}
}
`

func Gen(exp cps.Exp, r cps.Var) string {
	s := `(function() {`
	s += prelude
	s += "function " + jsvar(r) + "() { return []; };"
	s += `drive([function() {`
	s += gen(exp)
	return s + "}]);})();"
}

func gen(exp cps.Exp) string {
	switch exp := exp.(type) {
	case cps.Primop:
		var vl []string
		for _, v := range exp.Vs {
			vl = append(vl, genVal(v))
		}
		var wl []string
		for _, w := range exp.Ws {
			wl = append(wl, jsvar(w))
		}
		var cl []string
		for _, e := range exp.Es {
			cl = append(cl, gen(e))
		}
		return genPrim(exp.Op, vl, wl, cl)
	case cps.App:
		f := genVal(exp.F)
		var dl []string
		for _, v := range exp.Vs {
			dl = append(dl, genVal(v))
		}
		return "return [" + f + "," + strings.Join(dl, ",") + "];"
	case cps.Fix:
		s := ""
		for _, f := range exp.Fs {
			s += genFixent(f)
		}
		s += gen(exp.E)
		return s
	case cps.Record:
		s := "var " + jsvar(exp.W) + " = " + genRec(exp.Vs) + ";"
		return s + gen(exp.E)
	case cps.Select:
		s := genVal(exp.V) + "[" + strconv.Itoa(exp.I) + "]"
		return "var " + jsvar(exp.W) + " = " + s + ";" + gen(exp.E)
		panic("select")
	case cps.Switch:
		s := "switch (" + genVal(exp.I) + ") {"
		for i, e := range exp.Es {
			s += "case " + strconv.Itoa(i) + ": " + gen(e)
		}
		return s + "}"
	}
	fmt.Printf("exp %T\n", exp)
	panic("unreached")
}

func genFixent(f cps.FixEnt) string {
	body := gen(f.B)
	var al []string
	for _, a := range f.A {
		al = append(al, jsvar(a))
	}
	return `function ` + jsvar(f.V) + `(` + strings.Join(al, ",") + `) { ` + body + ` }`
}

func jsvar(v cps.Var) string {
	return fmt.Sprintf("v%d", v.ID)
}

func genVal(v cps.Value) string {
	switch v := v.(type) {
	case cps.Int:
		return strconv.Itoa(int(v))
	case cps.String:
		return strconv.QuoteToASCII(string(v))
	case cps.Undefined:
		return "undefined"
	case cps.Var:
		if v.Name != "" {
			return jsvar(v) + "/*" + v.Name + "*/"
		}
		return jsvar(v)
	}
	panic("unreached")
}

func genRec(vl []cps.RecordEnt) string {
	for _, v := range vl {
		if p, ok := v.Path.(cps.Offp); !ok || p != 0 {
			panic("unsupported path in record: " + fmt.Sprint(v.Path))
		}
	}
	var sl []string
	for _, v := range vl {
		sl = append(sl, genVal(v.V))
	}
	return "[" + strings.Join(sl, ",") + "]"
}
