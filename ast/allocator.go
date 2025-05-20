package ast

import "fmt"

var allocator *AddressAllocator

// Gestiona la asignación de direcciones de memoria para variables
type AddressAllocator struct {
	OperatorCode map[string]int // 0000-0099
	GlobalInt    int            // 1000–1999
	GlobalFloat  int            // 2000–2999
	LocalInt     int            // 3000–3999
	LocalFloat   int            // 4000–4999
	ConstInt     int            // 5000–5999
	ConstFloat   int            // 6000–6999
	TempInt      int            // 7000–7499
	TempFloat    int            // 7500–7999
	TempBool     int            // 8000–8499
}

// Inicializa el asignador de direcciones
func NewAddressAllocator() {
	allocator = &AddressAllocator{
		OperatorCode: map[string]int{
			"+":  0,
			"-":  1,
			"*":  2,
			"/":  3,
			">":  4,
			"<":  5,
			"!=": 6,
			"=":  7,
		},
		GlobalInt:   1000,
		GlobalFloat: 2000,
		LocalInt:    3000,
		LocalFloat:  4000,
		ConstInt:    5000,
		ConstFloat:  6000,
		TempInt:     7000,
		TempFloat:   7500,
		TempBool:    8000,
	}
}

// Global
func (a *AddressAllocator) NextGlobalInt() (int, error) {
	addr := a.GlobalInt
	if addr > 1999 {
		return -1, fmt.Errorf("espacio insuficiente para variables globales de tipo int")
	}
	a.GlobalInt++
	return addr, nil
}

func (a *AddressAllocator) NextGlobalFloat() (int, error) {
	addr := a.GlobalFloat
	if addr > 2999 {
		return -1, fmt.Errorf("espacio insuficiente para variables globales de tipo float")
	}
	a.GlobalFloat++
	return addr, nil
}

// Local
func (a *AddressAllocator) NextLocalInt() (int, error) {
	addr := a.LocalInt
	if addr > 3999 {
		return -1, fmt.Errorf("espacio insuficiente para variables locales de tipo int")
	}
	a.LocalInt++
	return addr, nil
}

func (a *AddressAllocator) NextLocalFloat() (int, error) {
	addr := a.LocalFloat
	if addr > 4999 {
		return -1, fmt.Errorf("espacio insuficiente para variables locales de tipo float")
	}
	a.LocalFloat++
	return addr, nil
}

// Const
func (a *AddressAllocator) NextConstInt() (int, error) {
	addr := a.ConstInt
	if addr > 5999 {
		return -1, fmt.Errorf("espacio insuficiente para variables constantes de tipo int")
	}
	a.ConstInt++
	return addr, nil
}

func (a *AddressAllocator) NextConstFloat() (int, error) {
	addr := a.ConstFloat
	if addr > 6999 {
		return -1, fmt.Errorf("espacio insuficiente para variables constantes de tipo float")
	}
	a.ConstFloat++
	return addr, nil
}

// Temp
func (a *AddressAllocator) NextTempInt() (int, error) {
	addr := a.TempInt
	if addr > 7499 {
		return -1, fmt.Errorf("espacio insuficiente para variables temporales de tipo int")
	}
	a.TempInt++
	return addr, nil
}

func (a *AddressAllocator) NextTempFloat() (int, error) {
	addr := a.TempFloat
	if addr > 7999 {
		return -1, fmt.Errorf("espacio insuficiente para variables temporales de tipo float")
	}
	a.TempFloat++
	return addr, nil
}

func (a *AddressAllocator) NextTempBool() (int, error) {
	addr := a.TempBool
	if addr > 8499 {
		return -1, fmt.Errorf("espacio insuficiente para variables temporales de tipo bool")
	}
	a.TempBool++
	return addr, nil
}
