package ast

import (
	"fmt"
	"strconv"
)

// Contexto de ejecución global
type Runtime struct {
	ExecutionStack []*StackFrame
	ReservedFrame  *StackFrame
	Quads          []Quadruple
	Output         []string
}

// Contexto de ejecución que almacena el estado actual
type StackFrame struct {
	Id       string
	Params   []string
	Local    *MemorySegment
	Temp     *MemorySegment
	ReturnIP int
}

func NewRuntime(ct *Compilation) *Runtime {
	return &Runtime{
		ExecutionStack: []*StackFrame{},
		ReservedFrame:  nil,
		Quads:          ct.Quads,
		Output:         []string{},
	}
}

// Agrega un nuevo contexto de llamada a la pila de ejecución
func (rt *Runtime) PushFrame() {
	rt.ExecutionStack = append(rt.ExecutionStack, rt.ReservedFrame)
	rt.ReservedFrame = nil
}

// Saca el contexto de llamada superior de la pila de ejecución
func (rt *Runtime) PopFrame() *StackFrame {
	if len(rt.ExecutionStack) == 0 {
		panic("pop en pila vacía")
	}
	frame := rt.ExecutionStack[len(rt.ExecutionStack)-1]
	rt.ExecutionStack = rt.ExecutionStack[:len(rt.ExecutionStack)-1]
	return frame
}

// Obtiene el contexto de llamada superior de la pila de ejecución
func (rt *Runtime) CurrentFrame() *StackFrame {
	if len(rt.ExecutionStack) == 0 {
		return nil
	}
	frame := rt.ExecutionStack[len(rt.ExecutionStack)-1]
	return frame
}

// Obtiene el nombre de una función por su cuádruplo de inicio
func (rt *Runtime) GetFunc(quadStart int) *FuncNode {
	for name, funcNode := range funcDir {
		if funcNode.QuadStart == quadStart {
			return funcDir[name]
		}
	}
	return &FuncNode{}
}

// Maneja operaciones de control de flujo
func (rt *Runtime) handleControlFlow(q Quadruple, ip int) (int, bool, error) {
	switch q.Operator {
	case GOTO:
		// Saltar al cuádruplo indicado
		ip = q.Result - 1
		return ip, true, nil

	case GOTOF:
		// Obtener el resultado de la condición desde memoria
		left, err := GetByAddress(q.Left, rt.CurrentFrame())
		if err != nil {
			return ip, true, err
		}

		// Si la condición es falsa, saltar al cuádruplo indicado
		if left.Value == "0" {
			ip = q.Result - 1
		}
		return ip, true, nil
	}

	return ip, false, nil
}

// Maneja operaciones de entrada/salida
func (rt *Runtime) handleIO(q Quadruple) (bool, error) {
	switch q.Operator {
	case PRINTLN:
		// Imprimir un salto de línea
		rt.Output = append(rt.Output, "\n")
		return true, nil

	case PRINT:
		// Obtener el operando izquierdo desde memoria
		left, err := GetByAddress(q.Left, rt.CurrentFrame())
		if err != nil {
			return true, err
		} else if left.Value == "" {
			return true, fmt.Errorf("variable %s no inicializada", left.Id)
		}

		switch left.Type {
		case "int", "float":
			// Imprimir el valor de la variable
			rt.Output = append(rt.Output, left.Value)
		case "bool":
			if left.Value == "1" {
				rt.Output = append(rt.Output, "true")
			} else {
				rt.Output = append(rt.Output, "false")
			}
		case "string":
			// Imprimir el string sin comillas
			rt.Output = append(rt.Output, left.Value[1:len(left.Value)-1])
		}

		rt.Output = append(rt.Output, " ")
		return true, nil
	}

	return false, nil
}

// Maneja llamadas a funciones
func (rt *Runtime) handleFunctionCalls(q Quadruple, ip int) (int, bool, error) {
	switch q.Operator {
	case ERA:
		// Obtener la función desde el directorio
		funcNode := rt.GetFunc(q.Left)

		// Crear un nuevo contexto de llamada
		newFrame := &StackFrame{
			Id:     funcNode.Id,
			Params: make([]string, len(funcNode.Params)),
			Local: &MemorySegment{
				Int:   []*VarNode{},
				Float: []*VarNode{},
			},
			Temp: &MemorySegment{
				Int:   []*VarNode{},
				Float: []*VarNode{},
				Bool:  []*VarNode{},
			},
			ReturnIP: -1,
		}

		// Recrear los parámetros y variables locales de la función
		for _, v := range append(funcNode.Params, funcNode.Vars...) {
			newFrame.Local.Insert(&VarNode{
				Address: v.Address,
				Id:      v.Id,
				Type:    v.Type,
			})
		}

		// Recrear las variables temporales de la función
		for _, t := range funcNode.Temps {
			newFrame.Temp.Insert(&VarNode{
				Address: t.Address,
				Id:      t.Id,
				Type:    t.Type,
			})
		}

		// Reservar el espacio de memoria para el nuevo contexto
		rt.ReservedFrame = newFrame
		return ip, true, nil

	case PARAM:
		// Obtener el operando izquierdo desde la memoria actual
		left, err := GetByAddress(q.Left, rt.CurrentFrame())
		if err != nil {
			return ip, true, err
		} else if left.Value == "" {
			return ip, true, fmt.Errorf("variable %s no inicializada", left.Id)
		}

		// Obtener el espacio reservado para el nuevo contexto
		frame := rt.ReservedFrame

		// Pasar el parámetro al contexto de llamada
		frame.Params[q.Result-1] = left.Value
		return ip, true, nil

	case GOSUB:
		// Obtener el espacio reservado para el nuevo contexto
		frame := rt.ReservedFrame

		// Guardar la dirección de retorno
		frame.ReturnIP = ip + 1

		// Obtener la función desde el directorio
		funcNode := funcDir[frame.Id]

		// Actualizar los valores locales con los parámetros pasados
		for i, val := range frame.Params {
			localNode, err := GetByAddress(funcNode.Params[i].Address, frame)
			if err != nil {
				return ip, true, err
			}
			localNode.Value = val
		}

		// Agregar el nuevo contexto de llamada a la pila
		rt.PushFrame()

		// Saltar al cuádruplo de inicio de la función
		ip = funcNode.QuadStart - 1
		return ip, true, nil

	case ENDFUNC:
		// Sacar el contexto de llamada actual
		frame := rt.PopFrame()

		// Si hay un contexto de llamada anterior, volver a él
		ip = frame.ReturnIP - 1
		return ip, true, nil
	case RETURN:
		// Obtener el contexto de llamada actual
		frame := rt.CurrentFrame()

		// Obtener la función desde el directorio
		funcNode := funcDir[frame.Id]

		// Obtener el resultado de la operación desde memoria
		leftNode, err := GetByAddress(q.Left, frame)
		if err != nil {
			return ip, true, err
		}

		// Obtener el nodo de retorno desde la memoria global
		returnNode, err := GetByAddress(funcNode.ReturnAddr, frame)
		if err != nil {
			return ip, true, err
		}

		// Actualiza el valor de retorno
		returnNode.Value = leftNode.Value

		return ip, true, nil
	}
	return ip, false, nil
}

// Maneja asignaciones
func (rt *Runtime) handleAssign(q Quadruple) (bool, error) {
	switch q.Operator {
	case ASSIGN:
		// Obtener el contexto de llamada actual
		frame := rt.CurrentFrame()

		// Obtener el operando izquierdo desde memoria
		left, err := GetByAddress(q.Left, frame)
		if err != nil {
			return true, err
		} else if left.Value == "" {
			return true, fmt.Errorf("variable %s no inicializada", left.Id)
		}

		// Obtener el nodo de resultado desde memoria
		result, err := GetByAddress(q.Result, frame)
		if err != nil {
			return true, err
		}

		// Guardar el resultado en memoria
		result.Value = left.Value
		return true, nil
	}
	return false, nil
}

// Maneja operaciones aritméticas y relacionales
func (rt *Runtime) handleArithmetic(q Quadruple) error {
	// Obtener el contexto de llamada actual
	frame := rt.CurrentFrame()

	// Obtener el operando izquierdo desde memoria
	left, err := GetByAddress(q.Left, frame)
	if err != nil {
		return err
	} else if left.Value == "" {
		return fmt.Errorf("variable %s no inicializada", left.Id)
	}

	// Obtener el operando derecho desde memoria
	right, err := GetByAddress(q.Right, frame)
	if err != nil {
		return err
	} else if right.Value == "" {
		return fmt.Errorf("variable %s no inicializada", right.Id)
	}

	// Obtener el nodo de resultado desde memoria
	result, err := GetByAddress(q.Result, frame)
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
		boolResult := valToFloat(lVal, lTyp) != valToFloat(rVal, rTyp)
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

	// Guardar el resultado en memoria
	result.Value = stringValue
	return nil
}

// Ejecuta los cuádruplos generados
func (rt *Runtime) RunProgram() error {
	fmt.Println()
	fmt.Println("Ejecución de cuádruplos")
	fmt.Println("===================================")

	// IP (Instruction Pointer) para iterar sobre los cuádruplos
	for ip := 0; ip < len(rt.Quads); ip++ {
		q := rt.Quads[ip]

		// Manejar operaciones de control de flujo
		if newIP, handled, err := rt.handleControlFlow(q, ip); handled {
			if err != nil {
				return err
			}
			ip = newIP
			continue
		}
		// Manejar operaciones de entrada/salida
		if handled, err := rt.handleIO(q); handled {
			if err != nil {
				return err
			}
			continue
		}
		// Manejar llamadas a funciones
		if newIP, handled, err := rt.handleFunctionCalls(q, ip); handled {
			if err != nil {
				return err
			}
			ip = newIP
			continue
		}
		// Manejar asignaciones
		if handled, err := rt.handleAssign(q); handled {
			if err != nil {
				return err
			}
			continue
		}
		// Manejar operaciones aritméticas y relacionales
		if err := rt.handleArithmetic(q); err != nil {
			return err
		}
	}

	return nil
}

// Limpia el contexto de ejecución y la memoria
func (rt *Runtime) Clear() {
	NewMemory()
	alloc = &Allocator{}
	rt.ExecutionStack = nil
	rt.ReservedFrame = nil
	for k := range funcDir {
		delete(funcDir, k)
	}
}

func (Runtime *Runtime) PrintOutput() {
	fmt.Println()
	fmt.Println("Salida del programa:")
	fmt.Println("===================================")
	for _, out := range Runtime.Output {
		fmt.Print(out)
	}
	fmt.Println()
}

// Convierte el valor a tipo float64
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
