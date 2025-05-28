package ast

import "fmt"

// Almacena el contexto de compilación actual
type Compilation struct {
	OperandStack []int
	Quads        []Quadruple
	TempCount    int
}

// Representa una instrucción de código intermedio (cuádruplo)
type Quadruple struct {
	Operator int
	Left     int
	Right    int
	Result   int
}

// Agrega un operando a la pila de operandos
func (ct *Compilation) Push(addr int) {
	ct.OperandStack = append(ct.OperandStack, addr)
}

// Saca el operando superior de la pila de operandos
func (ct *Compilation) Pop() int {
	if len(ct.OperandStack) == 0 {
		panic("pop en pila vacía")
	}
	addr := ct.OperandStack[len(ct.OperandStack)-1]
	ct.OperandStack = ct.OperandStack[:len(ct.OperandStack)-1]
	return addr
}

// Genera un nombre de variable temporal nuevo
func (ct *Compilation) NewTemp() string {
	ct.TempCount++
	return fmt.Sprintf("t%d", ct.TempCount)
}

// Agrega un nuevo cuádruplo a la lista
func (ct *Compilation) AddQuad(operator, left, right, result int) {
	ct.Quads = append(ct.Quads, Quadruple{
		Operator: operator,
		Left:     left,
		Right:    right,
		Result:   result,
	})
}

// Resetea el contexto para una nueva función
func (ct *Compilation) ClearLocalScope() {
	// Restablecer el ámbito global
	scope = global

	// Restablecer el contador de cuádruplos
	ct.TempCount = 0

	// Limpiar la memoria local y temporal
	memory.Local.Clear()
	memory.Temp.Clear()

	// Reiniciar los contadores
	alloc.Local.Reset()
	alloc.Temp.Reset()
}

// Imprime todos los cuádruplos con sus índices
func (ct *Compilation) PrintQuads() {
	fmt.Println()
	fmt.Println("Cuádruplos generados:")
	fmt.Println("===================================")

	var left string
	var right string
	var result string
	for i, q := range ct.Quads {
		// DEBUG
		if q.Left == -1 {
			left = "_"
		} else {
			left = fmt.Sprintf("%d", q.Left)
		}
		if q.Right == -1 {
			right = "_"
		} else {
			right = fmt.Sprintf("%d", q.Right)
		}
		if q.Result == -1 {
			result = "_"
		} else {
			result = fmt.Sprintf("%d", q.Result)
		}
		fmt.Printf("%d: (%s, %s, %s, %s)\n", i, opsList[q.Operator], left, right, result)
	}
}
