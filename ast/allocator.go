package ast

import "fmt"

var alloc *Allocator

// Inicializa el asignador de direcciones
func NewAllocator() {
	alloc = &Allocator{
		Operators: Range{Start: 0, End: 999, Counter: 0},
		Global: MemoryRanges{
			Int:   Range{Start: 1000, End: 1999, Counter: 1000},
			Float: Range{Start: 2000, End: 2999, Counter: 2000},
		},
		Local: MemoryRanges{
			Int:   Range{Start: 3000, End: 3999, Counter: 3000},
			Float: Range{Start: 4000, End: 4999, Counter: 4000},
		},
		Const: MemoryRanges{
			Int:   Range{Start: 5000, End: 5999, Counter: 5000},
			Float: Range{Start: 6000, End: 6999, Counter: 6000},
		},
		Temp: MemoryRanges{
			Int:   Range{Start: 7000, End: 7499, Counter: 7000},
			Float: Range{Start: 7500, End: 7999, Counter: 7500},
			Bool:  Range{Start: 8000, End: 8499, Counter: 8000},
		},
	}
}

// Global
func (a *Allocator) NextGlobalInt() (int, error) {
	addr := a.Global.Int.Counter
	if addr > a.Global.Int.End {
		return -1, fmt.Errorf("espacio insuficiente para variables globales de tipo int")
	}
	a.Global.Int.Counter++
	return addr, nil
}

func (a *Allocator) NextGlobalFloat() (int, error) {
	addr := a.Global.Float.Counter
	if a.Global.Float.Counter > 2999 {
		return -1, fmt.Errorf("espacio insuficiente para variables globales de tipo float")
	}
	a.Global.Float.Counter++
	return addr, nil
}

// Local
func (a *Allocator) NextLocalInt() (int, error) {
	addr := a.Local.Int.Counter
	if addr > a.Local.Int.End {
		return -1, fmt.Errorf("espacio insuficiente para variables locales de tipo int")
	}
	a.Local.Int.Counter++
	return addr, nil
}

func (a *Allocator) NextLocalFloat() (int, error) {
	addr := a.Local.Float.Counter
	if addr > a.Local.Float.End {
		return -1, fmt.Errorf("espacio insuficiente para variables locales de tipo float")
	}
	a.Local.Float.Counter++
	return addr, nil
}

// Const
func (a *Allocator) NextConstInt() (int, error) {
	addr := a.Const.Int.Counter
	if addr > a.Const.Int.End {
		return -1, fmt.Errorf("espacio insuficiente para variables constantes de tipo int")
	}
	a.Const.Int.Counter++
	return addr, nil
}

func (a *Allocator) NextConstFloat() (int, error) {
	addr := a.Const.Float.Counter
	if addr > a.Const.Float.End {
		return -1, fmt.Errorf("espacio insuficiente para variables constantes de tipo float")
	}
	a.Const.Float.Counter++
	return addr, nil
}

// Temp
func (a *Allocator) NextTempInt() (int, error) {
	addr := a.Temp.Int.Counter
	if addr > a.Temp.Int.End {
		return -1, fmt.Errorf("espacio insuficiente para variables temporales de tipo int")
	}
	a.Temp.Int.Counter++
	return addr, nil
}

func (a *Allocator) NextTempFloat() (int, error) {
	addr := a.Temp.Float.Counter
	if addr > a.Temp.Float.End {
		return -1, fmt.Errorf("espacio insuficiente para variables temporales de tipo float")
	}
	a.Temp.Float.Counter++
	return addr, nil
}

func (a *Allocator) NextTempBool() (int, error) {
	addr := a.Temp.Bool.Counter
	if addr > a.Temp.Bool.End {
		return -1, fmt.Errorf("espacio insuficiente para variables temporales de tipo bool")
	}
	a.Temp.Bool.Counter++
	return addr, nil
}
