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

// Genera un identificador único de etiqueta
func (ctx *Context) NewLabel() int {
	ctx.LabelCount++
	return ctx.LabelCount
}

// Marca el cuádruplo especial de etiqueta
func (ctx *Context) SetLabel(label int) {
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

// Ejecuta los cuádruplos generados
func (ctx *Context) Evaluate() error {
	/*fmt.Println()
	fmt.Println("Ejecución de cuádruplos")
	fmt.Println("===================================")*/

	for _, q := range ctx.Quads {
		// Ignorar cuádruplos de etiquetas
		if q.Left == -1 {
			continue
		}

		// Recuperar operando izquierdo desde memoria
		leftNode, err := lookupVarByAddress(q.Left)
		if err != nil {
			return err
		}

		// Recuperar operando derecho desde memoria si no es parte de un cuádruplo unario
		var rightNode *VarNode
		if q.Right != -1 {
			node, err := lookupVarByAddress(q.Right)
			if err != nil {
				return err
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
		case "string":
			leftVal = leftNode.Value
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
				fmt.Print(leftVal)
			case "bool":
				if leftVal == "1" {
					fmt.Print("true")
				} else {
					fmt.Print("false")
				}
			case "string":
				fmt.Print(leftVal)
			}
			fmt.Print(" ")
			return nil
		}

		// Obtener los datos de la variable de salida
		resultNode, err := lookupVarByAddress(q.Result)
		if err != nil {
			return err
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

		if opNode.Id == "=" {
			// Actualizar el nodo de destino
			if scope != global {
				memory.Local.Update(resultNode)
			} else {
				memory.Global.Update(resultNode)
			}
		}
	}
	return nil
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
