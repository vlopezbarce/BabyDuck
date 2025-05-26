package ast

import "fmt"

// Inicializa el asignador de direcciones
func NewAllocator() {
	alloc = &Allocator{
		Global: Segment{
			Int:   Range{Start: 1000, End: 1999, Counter: 1000},
			Float: Range{Start: 2000, End: 2999, Counter: 2000},
		},
		Local: Segment{
			Int:   Range{Start: 3000, End: 3999, Counter: 3000},
			Float: Range{Start: 4000, End: 4999, Counter: 4000},
		},
		Const: Segment{
			Int:    Range{Start: 5000, End: 5999, Counter: 5000},
			Float:  Range{Start: 6000, End: 6999, Counter: 6000},
			String: Range{Start: 7000, End: 7999, Counter: 7000},
		},
		Temp: Segment{
			Int:   Range{Start: 8000, End: 8499, Counter: 8000},
			Float: Range{Start: 8500, End: 8999, Counter: 8500},
			Bool:  Range{Start: 9000, End: 9499, Counter: 9000},
		},
	}
}

// Obtiene cantidad de variables creadas en un segmento
func (s *Segment) Count() int {
	return s.Int.Counter - s.Int.Start +
		s.Float.Counter - s.Float.Start +
		s.Bool.Counter - s.Bool.Start +
		s.String.Counter - s.String.Start
}

// Reinicia contadores para un segmento
func (s *Segment) Reset() {
	s.Int.Counter = s.Int.Start
	s.Float.Counter = s.Float.Start
	s.Bool.Counter = s.Bool.Start
	s.String.Counter = s.String.Start
}

// Global
func (a *Allocator) NextGlobal(typ string) (int, error) {
	var addr int
	var end int

	switch typ {
	case "int":
		addr = a.Global.Int.Counter
		end = a.Global.Int.End
		a.Global.Int.Counter++
	case "float":
		addr = a.Global.Float.Counter
		end = a.Global.Float.End
		a.Global.Float.Counter++
	}

	if addr > end {
		return -1, fmt.Errorf("espacio insuficiente para variables globales de tipo %s", typ)
	}

	return addr, nil
}

// Local
func (a *Allocator) NextLocal(typ string) (int, error) {
	var addr int
	var end int

	switch typ {
	case "int":
		addr = a.Local.Int.Counter
		end = a.Local.Int.End
		a.Local.Int.Counter++
	case "float":
		addr = a.Local.Float.Counter
		end = a.Local.Float.End
		a.Local.Float.Counter++
	}

	if addr > end {
		return -1, fmt.Errorf("espacio insuficiente para variables locales de tipo %s", typ)
	}

	return addr, nil
}

// Const
func (a *Allocator) NextConst(typ string) (int, error) {
	var addr int
	var end int

	switch typ {
	case "int":
		addr = a.Const.Int.Counter
		end = a.Const.Int.End
		a.Const.Int.Counter++
	case "float":
		addr = a.Const.Float.Counter
		end = a.Const.Float.End
		a.Const.Float.Counter++
	case "string":
		addr = a.Const.String.Counter
		end = a.Const.String.End
		a.Const.String.Counter++
	}

	if addr > end {
		return -1, fmt.Errorf("espacio insuficiente para variables constantes de tipo %s", typ)
	}

	return addr, nil
}

// Temp
func (a *Allocator) NextTemp(typ string) (int, error) {
	var addr int
	var end int

	switch typ {
	case "int":
		addr = a.Temp.Int.Counter
		end = a.Temp.Int.End
		a.Temp.Int.Counter++
	case "float":
		addr = a.Temp.Float.Counter
		end = a.Temp.Float.End
		a.Temp.Float.Counter++
	case "bool":
		addr = a.Temp.Bool.Counter
		end = a.Temp.Bool.End
		a.Temp.Bool.Counter++
	}

	if addr > end {
		return -1, fmt.Errorf("espacio insuficiente para variables temporales de tipo %s", typ)
	}

	return addr, nil
}
