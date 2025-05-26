package ast

import "fmt"

// Almacena el contexto de ejecución actual
type Context struct {
	SemStack  []int
	Quads     []Quadruple
	TempCount int
}

// Representa una instrucción de código intermedio (cuádruplo)
type Quadruple struct {
	Operator int
	Left     int
	Right    int
	Result   int
}

// Agrega un operando a la pila semántica
func (ctx *Context) Push(addr int) {
	ctx.SemStack = append(ctx.SemStack, addr)
}

// Saca el operando superior de la pila semántica
func (ctx *Context) Pop() int {
	if len(ctx.SemStack) == 0 {
		panic("pop en pila vacía")
	}
	addr := ctx.SemStack[len(ctx.SemStack)-1]
	ctx.SemStack = ctx.SemStack[:len(ctx.SemStack)-1]
	return addr
}

// Genera un nombre de variable temporal nuevo
func (ctx *Context) NewTemp() string {
	ctx.TempCount++
	return fmt.Sprintf("t%d", ctx.TempCount)
}

// Agrega un nuevo cuádruplo a la lista
func (ctx *Context) AddQuad(operator, left, right, result int) {
	ctx.Quads = append(ctx.Quads, Quadruple{
		Operator: operator,
		Left:     left,
		Right:    right,
		Result:   result,
	})
}

// Resetea el contexto para una nueva función
func (ctx *Context) ClearLocalScope() {
	// Restablecer el contador de cuádruplos
	ctx.TempCount = 0

	// Limpiar la memoria local y temporal
	memory.Local.Clear()
	memory.Temp.Clear()

	// Reiniciar los contadores
	alloc.Local.Reset()
	alloc.Temp.Reset()
}

// Imprime todos los cuádruplos con sus índices
func (ctx *Context) PrintQuads() {
	fmt.Println()
	fmt.Println("Cuádruplos generados:")
	fmt.Println("===================================")

	var left string
	var right string
	var result string
	for i, q := range ctx.Quads {
		// DEBUG
		if q.Left == -1 {
			left = "_"
		} else {
			lNode, err := GetCompileTimeVar(q.Left)
			if err != nil {
				left = fmt.Sprintf("%d", q.Left)
			} else if lNode.Id == "" {
				left = lNode.Value
			} else {
				left = lNode.Id
			}
		}
		if q.Right == -1 {
			right = "_"
		} else {
			rNode, err := GetCompileTimeVar(q.Right)
			if err != nil {
				right = fmt.Sprintf("%d", q.Right)
			} else if rNode.Id == "" {
				right = rNode.Value
			} else {
				right = rNode.Id
			}
		}
		if q.Result == -1 {
			result = "_"
		} else {
			resNode, err := GetCompileTimeVar(q.Result)
			if err != nil {
				result = fmt.Sprintf("%d", q.Result)
			} else if resNode.Id == "" {
				result = resNode.Value
			} else {
				result = resNode.Id
			}
		}
		fmt.Printf("%d: (%s, %s, %s, %s)\n", i, opsList[q.Operator], left, right, result)
	}
}

// Obtiene una variable de tiempo de compilación por su dirección
func GetCompileTimeVar(a int) (*VarNode, error) {
	// Globales
	if a >= alloc.Global.Int.Start && a <= alloc.Global.Float.End {
		if node, found := memory.Global.FindByAddress(a); found {
			return node, nil
		}
	}
	// Locales
	if a >= alloc.Local.Int.Start && a <= alloc.Local.Float.End {
		if node, found := memory.Local.FindByAddress(a); found {
			return node, nil
		}
	}
	// Constantes
	if a >= alloc.Const.Int.Start && a <= alloc.Const.String.End {
		if node, found := memory.Const.FindByAddress(a); found {
			return node, nil
		}
	}
	// Temporales
	if a >= alloc.Temp.Int.Start && a <= alloc.Temp.Bool.End {
		if node, found := memory.Temp.FindByAddress(a); found {
			return node, nil
		}
	}

	return nil, fmt.Errorf("dirección '%d' no encontrada", a)
}
