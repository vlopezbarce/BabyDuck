package ast

import (
	"fmt"
	"strconv"
)

// Representa una instrucción de código intermedio (cuádruplo)
type Quadruple struct {
	Operator VarNode // Operación
	Left     VarNode // Operando 1
	Right    VarNode // Operando 2
	Result   VarNode // Resultado
}

// Almacena la pila semántica, cuádruplos y contador de temporales
type Context struct {
	SemStack  []VarNode   // Pila de operandos
	Quads     []Quadruple // Fila de cuádruplos generados
	TempCount int         // Contador de variables temporales
}

// Agrega un operando a la pila semántica
func (ctx *Context) Push(varNode VarNode) {
	ctx.SemStack = append(ctx.SemStack, varNode)
}

// Saca el operando superior de la pila semántica
func (ctx *Context) Pop() VarNode {
	if len(ctx.SemStack) == 0 {
		panic("pop en pila vacía")
	}
	val := ctx.SemStack[len(ctx.SemStack)-1]
	ctx.SemStack = ctx.SemStack[:len(ctx.SemStack)-1]
	return val
}

// Genera un nombre de variable temporal nuevo
func (ctx *Context) NewTemp() string {
	ctx.TempCount++
	return fmt.Sprintf("t%d", ctx.TempCount)
}

// Agrega un nuevo cuádruplo a la lista
func (ctx *Context) AddQuad(operator, left, right, result VarNode) {
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
		fmt.Printf("  %2d: (%s, %s, %s, %s)\n", i, q.Operator.Address, q.Left.Address, q.Right.Address, q.Result.Address)
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
	ctx.AddQuad(opNode, result, VarNode{}, dest)

	return nil
}

// Genera el código intermedio para una expresión binaria
func (n *ExpressionNode) Generate(ctx *Context) (VarNode, error) {
	if _, err := n.Left.Generate(ctx); err != nil {
		return VarNode{}, err
	}

	if _, err := n.Right.Generate(ctx); err != nil {
		return VarNode{}, err
	}

	// Extrae los operandos de la pila
	right := ctx.Pop()
	left := ctx.Pop()

	// Verifica la compatibilidad de tipos
	resultType, err := CheckSemantic(n.Op, left.Type, right.Type)
	if err != nil {
		return VarNode{}, err
	}

	// Obtiene el operador de la memoria
	opNode, found := memory.Operators.FindByName(n.Op)
	if !found {
		return VarNode{}, fmt.Errorf("operador '%s' no encontrado", n.Op)
	}

	// Obtiene la dirección de memoria para el temporal
	var addr int
	var err error

	switch resultType {
	case "int":
		addr, err = memory.Temp.NextTempInt()
	case "float":
		addr, err = memory.Temp.NextTempFloat()
	case "bool":
		addr, err = memory.Temp.NextTempBool()
	}
	
	if err != nil {
		return VarNode{}, err
	}
	
	// Genera un nuevo temporal
	id := ctx.NewTemp()

	tempNode := VarNode{
		Address: addr,
		Id:      id,
		Type:    resultType,
		Value:   id,
	}

	// Agrega el cuádruplo
	ctx.AddQuad(opNode, left, right, tempNode)

	// Agrega el temporal a la pila
	ctx.Push(tempNode)

	return tempNode, nil
}

// Genera el código intermedio para una variable o constante
func (n *VarNode) Generate(ctx *Context) (VarNode, error) {
	// Si es una variable o un temporal
	if n.Id != "" {
		if n.Type == "operator" {
			// Buscar el operador en la memoria
			varNode, found := memory.Operators.FindByAddress(n.Address)
			if !found {
				return "", fmt.Errorf("operador '%s' no encontrado", n.Id)
			}

			return varNode, nil
		} else if n.Id[0] == 't' {
			// Buscar en la memoria de temporales
			varNode, found := memory.Temp.FindByAddress(n.Address)
			if !found {
				return "", fmt.Errorf("variable temporal '%s' no encontrada", n.Id)
			}
			return varNode, nil
		} else {
			// Buscar en la memoria local o global
			var varNode *VarNode
			var found bool

			if scope != global {
				varNode, found := memory.Local.FindByAddress(n.Address)
			}
			if !found {
				varNode, found := memory.Global.FindByAddress(n.Address)
			}
			if !found {
				return "", fmt.Errorf("variable '%s' no encontrada", n.Id)
			}
			
			return varNode, nil
		}
	} else {
		// Buscar la constante en la memoria
		varNode, found := memory.Const.FindByAddress(n.Address)

		if !found {
			// Obtener la dirección de memoria para la constante
			var addr int
			var err error
			
			switch n.Type {
			case "int":
				addr, err := memory.Const.NextConstInt()
			case "float":
				addr, err := memory.Const.NextConstFloat()
			}
			
			if err != nil {
				return "", err
			}

			// Agregar la constante a la memoria
			constNode = &VarNode{
				Address: addr,
				Type:    n.Type,
				Value:   n.Value,
			}

			memory.Const.Insert(constNode)
			return constNode, nil
		}

		return varNode, nil
	}
}

// Ejecuta los cuádruplos generados y devuelve el resultado
func (ctx *Context) Evaluate() VarNode {
	fmt.Println()
	fmt.Println("Ejecución de cuádruplos")
	fmt.Println("===================================")

	// Memoria para resultados intermedios
	memory.Temp.Clear()

	for _, q := range ctx.Quads {
		// 1) Recuperar operando izquierdo desde memoria si es temporal
		left := q.Left
		if info, found := ; found {
			left = info
		}

		// 2) Recuperar operando derecho desde memoria si es temporal
		right := q.Right
		if info, found := ; found {
			right = info
		}

		// 3) Convertir el valor textual a un dato Go, según su tipo
		var leftVal, rightVal Attrib
		switch left.Type {
		case "int":
			intVal, _ := strconv.Atoi(left.Value)
			leftVal = intVal
		case "float":
			floatVal, _ := strconv.ParseFloat(left.Value, 64)
			leftVal = floatVal
		case "bool":
			leftVal = left.Value == "1"
		}

		switch right.Type {
		case "int":
			intVal, _ := strconv.Atoi(right.Value)
			rightVal = intVal
		case "float":
			floatVal, _ := strconv.ParseFloat(right.Value, 64)
			rightVal = floatVal
		case "bool":
			rightVal = right.Value == "1"
		}

		// 4) Verificar el tipo de resultado con el cubo semántico
		var resultType string
		switch q.Operator {
		case "=":
			resultType = left.Type
		default:
			var err error
			resultType, err = CheckSemantic(q.Operator, left.Type, right.Type)
			if err != nil {
				panic(err)
			}
		}

		// 5) Ejecutar la operación
		var rawResult Attrib
		switch q.Operator {
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
		default:
			panic(fmt.Sprintf("Operación no soportada: %s", q.Operator))
		}

		// 6) Normalizar a string según el tipo de resultado
		var outValue string
		switch resultType {
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

		// 7) Debug
		fmt.Printf("%s %s %s -> %s (%s)\n",
			q.Operator,
			left.Value,
			right.Value,
			outValue,
			resultType,
		)
		
		// 8) Obtener la dirección de memoria para el resultado
	
		outAddr, _ = memory.Temp.NextGlobalInt()
		memory.Temp.Insert(&VarNode{
			Address: outAddr,
			Id:      q.Result,
			Type:    resultType,
			Value:   outValue,
			})

		// 9) Guardar el resultado en memoria
	}

	fmt.Println()
	fmt.Println("Memoria de temporales:")
	fmt.Println("===================================")
	fmt.Println(temps)

	// 9) Devolver el resultado final
	return temps[ctx.Quads[len(ctx.Quads)-1].Result]
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
