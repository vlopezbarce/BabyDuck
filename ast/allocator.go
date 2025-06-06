package ast

import "fmt"

// Asignador de direcciones de memoria
type Allocator struct {
	Global AllocSegment
	Local  AllocSegment
	Const  AllocSegment
	Temp   AllocSegment
}

// Segmentos de memoria para diferentes tipos de datos
type AllocSegment struct {
	Int    *Range
	Float  *Range
	Bool   *Range
	String *Range
}

// Rangos de direcciones para cada tipo de dato
type Range struct {
	Start int
	End   int
	Next  int
}

// Inicializa el asignador de direcciones
func NewAllocator() {
	alloc = &Allocator{
		Global: AllocSegment{
			Int:   &Range{Start: 1000, End: 1999, Next: 1000},
			Float: &Range{Start: 2000, End: 2999, Next: 2000},
		},
		Local: AllocSegment{
			Int:   &Range{Start: 3000, End: 3999, Next: 3000},
			Float: &Range{Start: 4000, End: 4999, Next: 4000},
		},
		Const: AllocSegment{
			Int:    &Range{Start: 5000, End: 5999, Next: 5000},
			Float:  &Range{Start: 6000, End: 6999, Next: 6000},
			String: &Range{Start: 7000, End: 7999, Next: 7000},
		},
		Temp: AllocSegment{
			Int:   &Range{Start: 8000, End: 8499, Next: 8000},
			Float: &Range{Start: 8500, End: 8999, Next: 8500},
			Bool:  &Range{Start: 9000, End: 9499, Next: 9000},
		},
	}
}

// Reinicia contadores para un segmento
func (s *AllocSegment) Reset() {
	s.Int.Next = s.Int.Start
	s.Float.Next = s.Float.Start
	if s.Bool != nil {
		s.Bool.Next = s.Bool.Start
	}
}

// Obtiene el segmento de memoria al que pertenece una dirección
func (a *Allocator) GetSegment(address int, frame *StackFrame) (*MemorySegment, AllocSegment) {
	var m *MemorySegment
	var s AllocSegment

	if address >= a.Global.Int.Start && address <= a.Global.Float.End {
		s = alloc.Global
		m = memory.Global
	}
	if address >= a.Const.Int.Start && address <= a.Const.String.End {
		s = alloc.Const
		m = memory.Const
	}
	if address >= a.Local.Int.Start && address <= a.Local.Float.End {
		s = alloc.Local
		// Verifica el contexto actual (nil si es durante la compilación)
		if frame != nil {
			m = frame.Local
		} else {
			m = memory.Local
		}
	}
	if address >= a.Temp.Int.Start && address <= a.Temp.Bool.End {
		s = alloc.Temp
		// Verifica el contexto actual (nil si es durante la compilación)
		if frame != nil {
			m = frame.Temp
		} else {
			m = memory.Temp
		}
	}

	return m, s
}

// Global
func (a *Allocator) NextGlobal(typ string) (int, error) {
	var r *Range
	switch typ {
	case "int":
		r = a.Global.Int
	case "float":
		r = a.Global.Float
	}
	if r.Next > r.End {
		return -1, fmt.Errorf("espacio insuficiente para variables globales de tipo %s", typ)
	}
	addr := r.Next
	r.Next++
	return addr, nil
}

// Local
func (a *Allocator) NextLocal(typ string) (int, error) {
	var r *Range
	switch typ {
	case "int":
		r = a.Local.Int
	case "float":
		r = a.Local.Float
	}
	if r.Next > r.End {
		return -1, fmt.Errorf("espacio insuficiente para variables locales de tipo %s", typ)
	}
	addr := r.Next
	r.Next++
	return addr, nil
}

// Const
func (a *Allocator) NextConst(typ string) (int, error) {
	var r *Range
	switch typ {
	case "int":
		r = a.Const.Int
	case "float":
		r = a.Const.Float
	case "string":
		r = a.Const.String
	}
	if r.Next > r.End {
		return -1, fmt.Errorf("espacio insuficiente para variables constantes de tipo %s", typ)
	}
	addr := r.Next
	r.Next++
	return addr, nil
}

// Temp
func (a *Allocator) NextTemp(typ string) (int, error) {
	var r *Range
	switch typ {
	case "int":
		r = a.Temp.Int
	case "float":
		r = a.Temp.Float
	case "bool":
		r = a.Temp.Bool
	}
	if r.Next > r.End {
		return -1, fmt.Errorf("espacio insuficiente para variables temporales de tipo %s", typ)
	}
	addr := r.Next
	r.Next++
	return addr, nil
}
