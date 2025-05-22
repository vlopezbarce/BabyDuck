package ast

import (
	"fmt"
	"strconv"
)

// Representa una instrucción de código intermedio (cuádruplo)
type Quadruple struct {
	Operator int // Operación
	Left     int // Operando 1
	Right    int // Operando 2
	Result   int // Resultado
}

// Almacena la pila semántica, cuádruplos y contador de temporales
type Context struct {
	SemStack  []int       // Pila de operandos
	Quads     []Quadruple // Fila de cuádruplos generados
	TempCount int         // Contador de variables temporales
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

// Imprime todos los cuádruplos con sus índices
func PrintQuads(quads []Quadruple) {
	fmt.Println()
	fmt.Println("Cuádruplos generados:")
	fmt.Println("===================================")

	for i, q := range quads {
		fmt.Printf("%d: %d %d %d %d\n", i, q.Operator, q.Left, q.Right, q.Result)
	}
}

// Genera el código intermedio para una asignación
func (n *AssignNode) Generate(ctx *Context, dest VarNode) error {
	// Obtiene el resultado de la expresión
	result, err := n.Exp.Generate(ctx)
	if err != nil {
		return err
	}

	// Obtiene el operador de la memoria
	opNode, found := memory.Operators.FindByName("=")
	if !found {
		return fmt.Errorf("operador '=' no encontrado")
	}

	// Agregar el cuádruplo
	ctx.AddQuad(opNode.Address, result, -1, dest.Address)

	return nil
}

// Genera el código intermedio para una impresión
func (n *PrintNode) Generate(ctx *Context) error {
	// Obtiene el resultado de la expresión
	result, err := n.Item.(Quad).Generate(ctx)
	if err != nil {
		return err
	}

	// Obtener el operador “print” de la memoria
	opNode, found := memory.Operators.FindByName("print")
	if !found {
		return fmt.Errorf("operador 'print' no encontrado")
	}

	// Agregar el cuádruplo
	ctx.AddQuad(opNode.Address, result, -1, -1)

	return nil
}

// Genera el código intermedio para una expresión binaria
func (n *ExpressionNode) Generate(ctx *Context) (int, error) {
	// Genera el código intermedio para los operandos izquierdo y derecho
	lAddr, err := n.Left.Generate(ctx)
	if err != nil {
		return -1, err
	}

	rAddr, err := n.Right.Generate(ctx)
	if err != nil {
		return -1, err
	}

	// Obtiene los nodos de memoria correspondientes
	leftNode, err := lookupVarByAddress(lAddr)
	if err != nil {
		return -1, err
	}
	rightNode, err := lookupVarByAddress(rAddr)
	if err != nil {
		return -1, err
	}

	// Verifica la compatibilidad de tipos
	resultType, err := CheckSemantic(n.Op, leftNode.Type, rightNode.Type)
	if err != nil {
		return -1, err
	}

	// Obtiene el operador de la memoria
	opNode, found := memory.Operators.FindByName(n.Op)
	if !found {
		return -1, fmt.Errorf("operador '%s' no encontrado", n.Op)
	}

	// Obtiene la dirección de memoria para el temporal
	var addr int
	switch resultType {
	case "int":
		addr, err = alloc.NextTempInt()
	case "float":
		addr, err = alloc.NextTempFloat()
	case "bool":
		addr, err = alloc.NextTempBool()
	}
	if err != nil {
		return -1, err
	}

	// Crea un nuevo nodo temporal
	tempId := ctx.NewTemp()

	tempNode := &VarNode{
		Address: addr,
		Id:      tempId,
		Type:    resultType,
		Value:   tempId,
	}

	// Inserta el temporal en la memoria
	memory.Temp.Insert(tempNode)

	// Agrega el cuádruplo
	ctx.AddQuad(opNode.Address, lAddr, rAddr, addr)

	// Agrega el temporal a la pila
	ctx.Push(addr)

	return addr, nil
}

// Genera el código intermedio para una variable o constante
func (n *VarNode) Generate(ctx *Context) (int, error) {
	if n.Id != "" {
		// Buscar en la memoria local o global
		varNode, _, err := lookupVar(n.Id)
		if err != nil {
			return -1, err
		}

		return varNode.Address, nil
	} else {
		// Buscar la constante en la memoria
		varNode, found := memory.Const.FindConst(n.Type, n.Value)

		if !found {
			// Obtener la dirección de memoria para la constante
			var addr int
			var err error

			switch n.Type {
			case "int":
				addr, err = alloc.NextConstInt()
			case "float":
				addr, err = alloc.NextConstFloat()
			}

			if err != nil {
				return -1, err
			}

			// Agregar la constante a la memoria
			constNode := &VarNode{
				Address: addr,
				Type:    n.Type,
				Value:   n.Value,
			}

			memory.Const.Insert(constNode)
			return addr, nil
		}

		return varNode.Address, nil
	}
}

// Ejecuta los cuádruplos generados y devuelve el resultado
func (ctx *Context) Evaluate() VarNode {
	fmt.Println()
	fmt.Println("Ejecución de cuádruplos")
	fmt.Println("===================================")
	var finalResult VarNode

	for _, q := range ctx.Quads {
		// Recuperar operando izquierdo desde memoria
		leftNode, err := lookupVarByAddress(q.Left)
		if err != nil {
			panic(err)
		}

		// Recuperar operando derecho desde memoria si no es parte de un cuádruplo unario
		var rightNode *VarNode

		if q.Right != -1 {
			node, err := lookupVarByAddress(q.Right)
			if err != nil {
				panic(err)
			}
			rightNode = node
		}

		// Convertir el string del valor a su tipo correspondiente
		var leftVal, rightVal Attrib
		switch leftNode.Type {
		case "int":
			leftVal, _ = strconv.Atoi(leftNode.Value)
		case "float":
			leftVal, _ = strconv.ParseFloat(leftNode.Value, 64)
		case "bool":
			leftVal = leftNode.Value == "1"
		}

		if q.Right != -1 {
			switch rightNode.Type {
			case "int":
				rightVal, _ = strconv.Atoi(rightNode.Value)
			case "float":
				rightVal, _ = strconv.ParseFloat(rightNode.Value, 64)
			case "bool":
				rightVal = rightNode.Value == "1"
			}
		}

		// Obtener el operador de la memoria
		opNode, found := memory.Operators.FindByAddress(q.Operator)
		if !found {
			panic(fmt.Sprintf("Operador '%d' no encontrado", q.Operator))
		}

		// Ejecutar la operación
		var rawResult Attrib
		switch opNode.Id {
		case "+":
			rawResult = toFloat64(leftVal) + toFloat64(rightVal)
		case "-":
			rawResult = toFloat64(leftVal) - toFloat64(rightVal)
		case "*":
			rawResult = toFloat64(leftVal) * toFloat64(rightVal)
		case "/":
			rawResult = toFloat64(leftVal) / toFloat64(rightVal)
		case ">":
			rawResult = toFloat64(leftVal) > toFloat64(rightVal)
		case "<":
			rawResult = toFloat64(leftVal) < toFloat64(rightVal)
		case "!=":
			rawResult = toFloat64(leftVal) != toFloat64(rightVal)
		case "=":
			rawResult = leftVal
		case "print":
			switch leftNode.Type {
			case "int", "float":
				fmt.Print(leftNode.Value)
			case "bool":
				if leftNode.Value == "1" {
					fmt.Print("true")
				} else {
					fmt.Print("false")
				}
			}
			return finalResult
		}

		// Obtener los datos de la variable de salida
		resultNode, err := lookupVarByAddress(q.Result)
		if err != nil {
			panic(err)
		}

		// Normalizar a string según el tipo de resultado
		var outValue string
		switch resultNode.Type {
		case "int":
			outValue = fmt.Sprintf("%d", int(toFloat64(rawResult)))
		case "float":
			outValue = fmt.Sprintf("%f", toFloat64(rawResult))
		case "bool":
			if rawResult.(bool) {
				outValue = "1"
			} else {
				outValue = "0"
			}
		}

		// Debug
		var debugRight string
		if q.Right != -1 {
			debugRight = rightNode.Value
		} else {
			debugRight = "_"
		}

		fmt.Printf("%s %s %s -> %s (%s)\n",
			opNode.Id,
			leftNode.Value,
			debugRight,
			outValue,
			resultNode.Type,
		)

		// Actualizar el nodo de resultado
		resultNode.Value = outValue

		// Guardar el resultado en memoria
		memory.Temp.Update(resultNode)
		finalResult = *resultNode
	}

	fmt.Println()
	fmt.Println("Memoria de temporales:")
	fmt.Println("===================================")
	memory.Temp.Print()
	return finalResult
}

// Convierte int, float64 o bool a float64
func toFloat64(v Attrib) float64 {
	switch x := v.(type) {
	case int:
		return float64(x)
	case float64:
		return x
	case bool:
		if x {
			return 1
		}
		return 0
	default:
		panic("tipo no soportado en toFloat64")
	}
}
