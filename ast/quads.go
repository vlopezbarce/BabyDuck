package ast

import (
	"fmt"
	"strconv"
)

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

// Genera un identificador único de etiqueta (label)
func (ctx *Context) NewLabel() int {
	ctx.LabelCount++
	return ctx.LabelCount
}

// Marca la posición actual del cuádruplo con la etiqueta dada
func (ctx *Context) SetLabel(label int) {
	// Insertar un cuádruplo especial de etiqueta
	opNode, _ := memory.Operators.FindByName("LABEL")
	ctx.AddQuad(opNode.Address, -1, -1, label)
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
func (ctx *Context) PrintQuads() {
	fmt.Println()
	fmt.Println("Cuádruplos generados:")
	fmt.Println("===================================")

	var left string
	var right string
	var result string
	for i, q := range ctx.Quads {
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
		fmt.Printf("%d: (%d, %s, %s, %s)\n", i, q.Operator, left, right, result)
	}
}

// Imprime los temporales generados
func (ctx *Context) PrintTemps() {
	if ctx.TempCount > 0 {
		fmt.Println()
		fmt.Println("Temporales:")
		fmt.Println("===================================")
		memory.Temp.Print()
	}
}

// Genera el código intermedio para una asignación
func (n *AssignNode) Generate(ctx *Context) (*VarNode, error) {
	// Buscar variable destino y memoria correcta
	var dest *VarNode
	var found bool

	if scope != global {
		if dest, found = memory.Local.FindByName(n.Id); !found {
			return nil, fmt.Errorf("variable '%s' no declarada en el ámbito actual", n.Id)
		}
	} else {
		if dest, found = memory.Global.FindByName(n.Id); !found {
			return nil, fmt.Errorf("variable '%s' no declarada en el ámbito actual", n.Id)
		}
	}

	// Obtener el resultado de la expresión
	if _, err := n.Exp.Generate(ctx); err != nil {
		return nil, err
	}
	result := ctx.Pop()

	// Obtener el operador de la memoria
	opNode, _ := memory.Operators.FindByName("=")

	// Agregar el cuádruplo
	ctx.AddQuad(opNode.Address, result, -1, dest.Address)

	return dest, nil
}

// Genera el código intermedio para una impresión
func (n *PrintNode) Generate(ctx *Context) error {
	// Obtener el resultado de la expresión
	if _, err := n.Item.(Quad).Generate(ctx); err != nil {
		return err
	}
	result := ctx.Pop()

	// Obtener el operador de la memoria
	opNode, _ := memory.Operators.FindByName("PRINT")

	// Agregar el cuádruplo
	ctx.AddQuad(opNode.Address, result, -1, -1)

	return nil
}

// Genera el código intermedio para una expresión binaria
func (n *ExpressionNode) Generate(ctx *Context) (int, error) {
	// Generar el código intermedio para los operandos izquierdo y derecho
	if _, err := n.Left.Generate(ctx); err != nil {
		return -1, err
	}
	if _, err := n.Right.Generate(ctx); err != nil {
		return -1, err
	}

	// Obtener los operandos izquierdo y derecho
	right := ctx.Pop()
	left := ctx.Pop()

	// Obtener los nodos de memoria correspondientes
	leftNode, err := lookupVarByAddress(left)
	if err != nil {
		return -1, err
	}
	rightNode, err := lookupVarByAddress(right)
	if err != nil {
		return -1, err
	}

	// Verificar la compatibilidad de tipos
	resultType, err := CheckSemantic(n.Op, leftNode.Type, rightNode.Type)
	if err != nil {
		return -1, err
	}

	// Obtener el operador de la memoria
	opNode, _ := memory.Operators.FindByName(n.Op)

	// Obtener la dirección de memoria para el temporal
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

	// Crear un nuevo nodo temporal
	tempId := ctx.NewTemp()

	tempNode := &VarNode{
		Address: addr,
		Id:      tempId,
		Type:    resultType,
		Value:   tempId,
	}

	// Insertar el temporal en la memoria
	memory.Temp.Insert(tempNode)

	// Agregar el cuádruplo
	ctx.AddQuad(opNode.Address, left, right, addr)

	// Agregar el temporal a la pila
	ctx.Push(addr)

	return addr, nil
}

// Genera el código intermedio para una variable o constante
func (n *VarNode) Generate(ctx *Context) (int, error) {
	if n.Id != "" {
		// Buscar en la memoria local o global
		var varNode *VarNode
		var found bool

		if scope != global {
			if varNode, found = memory.Local.FindByName(n.Id); found {
				if varNode.Value == "" {
					return -1, fmt.Errorf("variable '%s' no asignada", n.Id)
				}
			}
		}
		if !found {
			if varNode, found = memory.Global.FindByName(n.Id); found {
				if varNode.Value == "" {
					return -1, fmt.Errorf("variable '%s' no asignada", n.Id)
				}
			}
		}
		if !found {
			return -1, fmt.Errorf("variable '%s' no declarada en el ámbito actual", n.Id)
		}

		// Agregar la direción a la pila
		ctx.Push(varNode.Address)

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

			// Agregar la constante a la memoria
			memory.Const.Insert(constNode)

			// Agregar la dirección a la pila
			ctx.Push(addr)

			return addr, nil
		}

		// Agregar la dirección a la pila
		ctx.Push(varNode.Address)

		return varNode.Address, nil
	}
}

// Ejecuta los cuádruplos generados y devuelve el resultado
func (ctx *Context) Evaluate() VarNode {
	/*fmt.Println()
	fmt.Println("Ejecución de cuádruplos")
	fmt.Println("===================================")*/
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
		opNode, _ := memory.Operators.FindByAddress(q.Operator)

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
		case "PRINT":
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
		/*var debugRight string
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
		)*/

		// Actualizar el nodo de resultado
		resultNode.Value = outValue

		// Guardar el resultado en memoria
		memory.Temp.Update(resultNode)
		finalResult = *resultNode
	}

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
