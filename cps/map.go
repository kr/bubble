package cps

// Walk traverses exp depth first
// and applies f to every Exp that occurs,
// starting with exp itself.
func Walk(exp Exp, f func(Exp)) {
	Map(exp, func(exp Exp) Exp {
		f(exp)
		return exp
	})
}

// WalkValues applies f to every Value that occurs in exp.
func WalkValues(exp Exp, f func(Value)) {
	MapValues(exp, func(v Value) Value {
		f(v)
		return v
	})
}

// Map traverses exp depth first
// and applies f to every Exp that occurs,
// starting with exp itself.
// It returns the resulting Exp.
func Map(exp Exp, f func(Exp) Exp) Exp {
	return f(exp).mapexp(f)
}

func (a App) mapexp(f func(Exp) Exp) Exp {
	return a
}

func (fx Fix) mapexp(f func(Exp) Exp) Exp {
	var fl []FixEnt
	for _, fent := range fx.Fs {
		fent.B = Map(fent.B, f)
		fl = append(fl, fent)
	}
	return Fix{Fs: fl, E: Map(fx.E, f)}
}

func (p Primop) mapexp(f func(Exp) Exp) Exp {
	p1 := p
	p1.Es = nil
	for _, e := range p.Es {
		p1.Es = append(p1.Es, Map(e, f))
	}
	return p1
}

func (r Record) mapexp(f func(Exp) Exp) Exp {
	r.E = Map(r.E, f)
	return r
}

func (s Select) mapexp(f func(Exp) Exp) Exp {
	s.E = Map(s.E, f)
	return s
}

func (s Switch) mapexp(f func(Exp) Exp) Exp {
	s1 := s
	s1.Es = nil
	for _, e := range s.Es {
		s1.Es = append(s1.Es, Map(e, f))
	}
	return s1
}

// MapValues applies f to every Value that occurs in exp,
// and returns the resulting Exp.
func MapValues(exp Exp, f func(Value) Value) Exp {
	return exp.mapval(f)
}

func (a App) mapval(f func(Value) Value) Exp {
	var vs []Value
	for _, v := range a.Vs {
		vs = append(vs, f(v))
	}
	return App{F: f(a.F), Vs: vs}
}

func (fx Fix) mapval(f func(Value) Value) Exp {
	var fs []FixEnt
	for _, ent := range fx.Fs {
		ent.B = MapValues(ent.B, f)
		fs = append(fs, ent)
	}
	return Fix{Fs: fs, E: MapValues(fx.E, f)}
}

func (p Primop) mapval(f func(Value) Value) Exp {
	p1 := p
	p1.Vs = nil
	for _, v := range p.Vs {
		p1.Vs = append(p1.Vs, f(v))
	}
	p1.Es = nil
	for _, e := range p.Es {
		p1.Es = append(p1.Es, MapValues(e, f))
	}
	return p1
}

func (r Record) mapval(f func(Value) Value) Exp {
	r1 := Record{
		W: r.W,
		E: MapValues(r.E, f),
	}
	for _, v := range r.Vs {
		v.V = f(v.V)
		r1.Vs = append(r1.Vs, v)
	}
	return r1
}

func (s Select) mapval(f func(Value) Value) Exp {
	s.V = f(s.V)
	s.E = MapValues(s.E, f)
	return s
}

func (s Switch) mapval(f func(Value) Value) Exp {
	s1 := Switch{I: f(s.I)}
	for _, e := range s.Es {
		s1.Es = append(s1.Es, MapValues(e, f))
	}
	return s1
}
