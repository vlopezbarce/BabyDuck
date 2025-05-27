package ast

import (
	"fmt"
	"strconv"
)

// Contexto de ejecución global
type Runtime struct {
	ExecutionStack []*StackFrame
	Quads          []Quadruple
	Output         []string
}

// Contexto de ejecución que almacena el estado actual
type StackFrame struct {
	Id       string
	Params   []*VarNode
	ReturnIP int
	Local    *MemorySegment
	Temp     *MemorySegment
}

func NewRuntime(ctx *Context) *Runtime {
	return &Runtime{
		ExecutionStack: []*StackFrame{},
		Quads:          ctx.Quads,
		Output:         []string{},
	}
}

// Agrega un nuevo contexto de llamada a la pila de ejecución
func (r *Runtime) Push(frame *StackFrame) {
	r.ExecutionStack = append(r.ExecutionStack, frame)
}

// Saca el contexto de llamada superior de la pila de ejecución
func (r *Runtime) Pop() *StackFrame {
	if len(r.ExecutionStack) == 0 {
		panic("pop en pila vacía")
	}
	frame := r.ExecutionStack[len(r.ExecutionStack)-1]
	r.ExecutionStack = r.ExecutionStack[:len(r.ExecutionStack)-1]
	return frame
}

// Obtiene el contexto de llamada superior de la pila de ejecución
func (r *Runtime) GetFrame() *StackFrame {
	if len(r.ExecutionStack) == 0 {
		return nil
	}
	frame := r.ExecutionStack[len(r.ExecutionStack)-1]
	return frame
}

// Obtiene el nombre de una función por su cuádruplo de inicio
func (runtime *Runtime) GetFunc(quadStart int) *FuncNode {
	for name, funcNode := range funcDir {
		if funcNode.QuadStart == quadStart {
			return funcDir[name]
		}
	}
	return &FuncNode{}
}

// Convierte el valor a tipo float64
func ValToFloat(val string, typ string) float64 {
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

// Ejecuta los cuádruplos generados
func (runtime *Runtime) RunProgram() error {
	fmt.Println()
	fmt.Println("Ejecución de cuádruplos")
	fmt.Println("===================================")

	// IP (Instruction Pointer) para iterar sobre los cuádruplos
	for IP := 0; IP < len(runtime.Quads); IP++ {
		q := runtime.Quads[IP]

		// Operaciones de estatutos condicionales, cíclicos, y de control de flujo
		switch q.Operator {
		case GOTO:
			// Saltar al cuádruplo indicado
			IP = q.Result - 1
			fmt.Printf("%s %d\n", opsList[q.Operator], q.Result)

			continue
		case GOTOF:
			// Obtener el resultado de la condición desde la memoria de temporales
			memSegment, allocSegment := alloc.GetSegment(q.Left, runtime)
			node, _ := memSegment.FindByAddress(allocSegment, q.Left)

			// Si la condición es falsa, saltar al cuádruplo indicado
			if node.Value == "0" {
				IP = q.Result - 1
				fmt.Printf("%s %d\n", opsList[q.Operator], q.Result)
			}

			continue
		case PRINTLN:
			// Imprimir un salto de línea
			runtime.Output = append(runtime.Output, "\n")

			continue
		case ERA:
			// Obtener la función desde el directorio
			funcNode := runtime.GetFunc(q.Left)

			// Crear un nuevo contexto de llamada
			frame := &StackFrame{
				Id:       funcNode.Id,
				Params:   make([]*VarNode, len(funcNode.Params)),
				ReturnIP: -1,
				Local: &MemorySegment{
					Int:   []*VarNode{},
					Float: []*VarNode{},
				},
				Temp: &MemorySegment{
					Int:   []*VarNode{},
					Float: []*VarNode{},
					Bool:  []*VarNode{},
				},
			}

			// Recrear los parámetros y variables locales de la función
			for _, v := range append(funcNode.Params, funcNode.Vars...) {
				frame.Local.Insert(&VarNode{
					Address: v.Address,
					Id:      v.Id,
					Type:    v.Type,
				})
			}

			// Recrear las variables temporales de la función
			for _, t := range funcNode.Temps {
				frame.Temp.Insert(&VarNode{
					Address: t.Address,
					Id:      t.Id,
					Type:    t.Type,
				})
			}

			// Agregar el contexto de llamada a la pila
			runtime.Push(frame)
			fmt.Printf("%s %s\n", opsList[q.Operator], funcNode.Id)

			continue
		case GOSUB:
			// Obtener el contexto de llamada actual
			frame := runtime.GetFrame()

			// Guardar la dirección de retorno
			frame.ReturnIP = IP + 1

			// Obtener la función desde el directorio
			funcNode := runtime.GetFunc(q.Left)

			// Actualizar los valores de los parámetros
			for i, p := range frame.Params {
				funcNode.Params[i].Value = p.Value
				frame.Local.Update(alloc.Local, funcNode.Params[i])
			}

			// Saltar al cuádruplo de inicio de la función
			IP = funcNode.QuadStart - 1
			fmt.Printf("%s %s\n", opsList[q.Operator], funcNode.Id)

			continue
		case ENDFUNC:
			// Obtener el contexto de llamada actual
			frame := runtime.Pop()
			IP = frame.ReturnIP - 1
			fmt.Printf("%s %s\n", opsList[q.Operator], frame.Id)

			continue
		}

		// Obtener el operando izquierdo desde memoria
		left, err := alloc.GetByAddress(q.Left, runtime)
		if err != nil {
			return err
		}

		switch q.Operator {
		case PARAM:
			// Obtener el contexto de llamada actual
			frame := runtime.GetFrame()

			// Asignar el valor del parámetro al frame actual
			frame.Params[q.Result-1] = left
			fmt.Printf("%s %s = %s\n", opsList[q.Operator], left.Id, left.Value)

			continue
		case PRINT:
			// Imprimir el valor de la variable
			switch left.Type {
			case "int", "float":
				runtime.Output = append(runtime.Output, left.Value)
			case "bool":
				if left.Value == "1" {
					runtime.Output = append(runtime.Output, "true")
				} else {
					runtime.Output = append(runtime.Output, "false")
				}
			case "string":
				// Imprimir el string sin comillas
				runtime.Output = append(runtime.Output, left.Value[1:len(left.Value)-1])
			}
			continue
		}

		// Obtener el nodo de resultado desde memoria
		memSegment, allocSegment := alloc.GetSegment(q.Result, runtime)
		result, err := memSegment.FindByAddress(allocSegment, q.Result)
		if err != nil {
			return err
		}

		// Operador de asignación
		switch q.Operator {
		case ASSIGN:
			// Verificar que el tipo de la variable izquierda y el resultado sean compatibles
			if result.Type != left.Type {
				return fmt.Errorf("tipo incompatible en asignación: se esperaba %s, se obtuvo %s", result.Type, left.Type)
			}
			result.Value = left.Value

			// Guardar el resultado en memoria
			memSegment.Update(allocSegment, result)

			fmt.Printf("%s %s %s (%s)\n", result.Id, opsList[q.Operator], result.Value, result.Type)

			continue
		}

		// Obtener el operando derecho desde memoria
		right, err := alloc.GetByAddress(q.Right, runtime)
		if err != nil {
			return err
		}

		// Ejecutar la operación
		var floatResult float64
		lVal := left.Value
		lTyp := left.Type
		rVal := right.Value
		rTyp := right.Type

		// Operadores aritméticos y relacionales
		switch q.Operator {
		case PLUS:
			floatResult = ValToFloat(lVal, lTyp) + ValToFloat(rVal, rTyp)
		case MINUS:
			floatResult = ValToFloat(lVal, lTyp) - ValToFloat(rVal, rTyp)
		case TIMES:
			floatResult = ValToFloat(lVal, lTyp) * ValToFloat(rVal, rTyp)
		case DIVIDE:
			floatResult = ValToFloat(lVal, lTyp) / ValToFloat(rVal, rTyp)
		case GT:
			boolResult := ValToFloat(lVal, lTyp) > ValToFloat(rVal, rTyp)
			floatResult = ValToFloat(fmt.Sprintf("%t", boolResult), "bool")
		case LT:
			boolResult := ValToFloat(lVal, lTyp) < ValToFloat(rVal, rTyp)
			floatResult = ValToFloat(fmt.Sprintf("%t", boolResult), "bool")
		case NEQ:
			boolResult := ValToFloat(lVal, lTyp) < ValToFloat(rVal, rTyp)
			floatResult = ValToFloat(fmt.Sprintf("%t", boolResult), "bool")
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
		memSegment.Update(allocSegment, result)

		// Debug
		fmt.Printf("%s %s %s -> %s (%s)\n", left.Value, opsList[q.Operator], right.Value, result.Value, result.Type)
	}
	return nil
}

func (runtime *Runtime) PrintOutput() {
	fmt.Println()
	fmt.Println("Salida del programa:")
	fmt.Println("===================================")
	for _, out := range runtime.Output {
		fmt.Print(out)
	}
}
