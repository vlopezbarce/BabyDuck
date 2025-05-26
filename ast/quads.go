package ast

import (
	"fmt"
	"strconv"
)

// Resetear el contexto para una nueva función
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
			lNode, err := GetVarByAddress(q.Left)
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
			rNode, err := GetVarByAddress(q.Right)
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
			resNode, err := GetVarByAddress(q.Result)
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

// Ejecuta los cuádruplos generados
func (ctx *Context) Evaluate() error {
	fmt.Println()
	fmt.Println("Ejecución de cuádruplos")
	fmt.Println("===================================")

	for i := 0; i < len(ctx.Quads); i++ {
		q := ctx.Quads[i]

		switch q.Operator {
		case GOTO:
			// Saltar al cuádruplo indicado
			i = q.Result - 1
			fmt.Printf("%s %d\n", opsList[q.Operator], q.Result)

			continue
		case GOTOF:
			// Obtener el resultado de la condición desde la memoria de temporales
			node, _ := memory.Temp.FindByAddress(q.Left)

			// Si la condición es falsa, saltar al cuádruplo indicado
			if node.Value == "0" {
				i = q.Result - 1
				fmt.Printf("%s %d\n", opsList[q.Operator], q.Result)
			}

			continue
		case PRINTLN:
			// Imprimir un salto de línea
			fmt.Println()

			continue
		}

		// Obtener el operando izquierdo desde memoria
		left, err := GetVarByAddress(q.Left)
		if err != nil {
			return err
		}

		switch q.Operator {
		case PRINT:
			// Imprimir el valor de la variable
			switch left.Type {
			case "int", "float":
				fmt.Print(left.Value, " ")
			case "bool":
				if left.Value == "1" {
					fmt.Print("true", " ")
				} else {
					fmt.Print("false", " ")
				}
			case "string":
				// Imprimir el string sin comillas
				fmt.Print(left.Value[1:len(left.Value)-1], " ")
			}
			continue
		}

		// Obtener el nodo de resultado desde memoria
		result, err := GetVarByAddress(q.Result)
		if err != nil {
			return err
		}

		switch q.Operator {
		case ASSIGN:
			// Verificar que el tipo de la variable izquierda y el resultado sean compatibles
			if result.Type != left.Type {
				return fmt.Errorf("tipo incompatible en asignación: se esperaba %s, se obtuvo %s", result.Type, left.Type)
			}
			result.Value = left.Value

			// Guardar el resultado en memoria
			if scope != global {
				memory.Local.Update(result)
			} else {
				memory.Global.Update(result)
			}
			fmt.Printf("%s %s %s (%s)\n", result.Id, opsList[q.Operator], result.Value, result.Type)

			continue
		}

		// Obtener el operando derecho desde memoria
		right, err := GetVarByAddress(q.Right)
		if err != nil {
			return err
		}

		// Ejecutar la operación
		var floatResult float64
		lVal := left.Value
		lTyp := left.Type
		rVal := right.Value
		rTyp := right.Type

		switch q.Operator {
		case PLUS:
			floatResult = valToFloat(lVal, lTyp) + valToFloat(rVal, rTyp)
		case MINUS:
			floatResult = valToFloat(lVal, lTyp) - valToFloat(rVal, rTyp)
		case TIMES:
			floatResult = valToFloat(lVal, lTyp) * valToFloat(rVal, rTyp)
		case DIVIDE:
			floatResult = valToFloat(lVal, lTyp) / valToFloat(rVal, rTyp)
		case GT:
			boolResult := valToFloat(lVal, lTyp) > valToFloat(rVal, rTyp)
			floatResult = valToFloat(fmt.Sprintf("%t", boolResult), "bool")
		case LT:
			boolResult := valToFloat(lVal, lTyp) < valToFloat(rVal, rTyp)
			floatResult = valToFloat(fmt.Sprintf("%t", boolResult), "bool")
		case NEQ:
			boolResult := valToFloat(lVal, lTyp) < valToFloat(rVal, rTyp)
			floatResult = valToFloat(fmt.Sprintf("%t", boolResult), "bool")
		}

		// Normalizar a string según el tipo de resultado
		var stringValue string
		switch result.Type {
		case "int", "bool":
			stringValue = fmt.Sprintf("%d", int(floatResult))
		case "float":
			stringValue = fmt.Sprintf("%f", floatResult)
		}

		// Actualizar el nodo de resultado
		result.Value = stringValue

		// Guardar el resultado en memoria
		memory.Temp.Update(result)

		// Debug
		fmt.Printf("%s %s %s -> %s (%s)\n", left.Value, opsList[q.Operator], right.Value, result.Value, result.Type)
	}
	return nil
}

// Convertir el valor a tipo float64
func valToFloat(val string, typ string) float64 {
	switch typ {
	case "int":
		intVal, _ := strconv.Atoi(val)
		return float64(intVal)
	case "float":
		floatVal, _ := strconv.ParseFloat(val, 64)
		return floatVal
	case "bool":
		if val == "true" || val == "1" {
			return 1
		}
		return 0
	}
	return 0
}
