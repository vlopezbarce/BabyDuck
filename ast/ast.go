package ast

import (
	"fmt"
	"math"
)

var global string // Nombre del programa
var scope string  // Ámbito actual

// Inicializa el programa principal
func NewProgram(name string) string {
	// Establece el ámbito global y actual
	global = name
	scope = name

	// Inicializa la memoria
	NewMemory()
	NewAddressAllocator()

	return name
}

// Declara una nueva variable
func NewVariable(id, typ string, start int) error {
	// Verificar si ya existe una variable con el mismo nombre
	if _, found := memory.Local.FindByName(id, start, math.MaxInt); found {
		return fmt.Errorf("variable '%s' ya declarada", id)
	}
	if scope != global {
		start = functionDirectory[global].Range.Int.Start
	}
	if _, found := memory.Global.FindByName(id, start, math.MaxInt); found {
		return fmt.Errorf("variable '%s' ya declarada", id)
	}

	// Asignar dirección de memoria
	var addr int
	var err error

	if scope == global {
		// Ámbito global
		switch typ {
		case "int":
			addr, err = allocator.NextGlobalInt()
		case "float":
			addr, err = allocator.NextGlobalFloat()
		default:
			return fmt.Errorf("tipo desconocido: %s", typ)
		}
	} else {
		// Ámbito local
		switch typ {
		case "int":
			addr, err = allocator.NextLocalInt()
		case "float":
			addr, err = allocator.NextLocalFloat()
		default:
			return fmt.Errorf("tipo desconocido: %s", typ)
		}
	}

	if err != nil {
		return fmt.Errorf("error al asignar dirección para variable '%s': %v", id, err)
	}

	// Insertar en el árbol correspondiente
	node := &VarNode{
		Address: addr,
		Id:      id,
		Type:    typ,
	}

	if scope == global {
		memory.Global.Insert(node)
	} else {
		memory.Local.Insert(node)
	}

	return nil
}

// Función constructora para FuncNode
func NewFunction(id string, vars []*VarNode, body []Attrib) (*FuncNode, error) {
	// Verificar si la función ya existe
	if _, exists := functionDirectory[id]; exists {
		return nil, fmt.Errorf("función '%s' ya declarada", id)
	}

	// Establecer el ámbito actual a la nueva función
	scope = id

	// Asignar rangos de memoria para la función
	var intStart, intEnd, floatStart, floatEnd int

	if scope == global {
		intStart = allocator.GlobalInt
		floatStart = allocator.GlobalFloat
	} else {
		intStart = allocator.LocalInt
		floatStart = allocator.LocalFloat
	}

	// Registrar los parámetros como variables locales
	for _, param := range vars {
		var start int
		if param.Type == "int" {
			start = intStart
		}
		if param.Type == "float" {
			start = floatStart
		}
		if err := NewVariable(param.Id, param.Type, start); err != nil {
			return nil, err
		}
	}

	// Asignar rangos de memoria para la función
	if scope == global {
		intEnd = allocator.GlobalInt - 1
		floatEnd = allocator.GlobalFloat - 1
	} else {
		intEnd = allocator.LocalInt - 1
		floatEnd = allocator.LocalFloat - 1
	}

	// Crear el nodo de función
	funcNode := &FuncNode{
		Id:   id,
		Body: body,
		Range: MemoryRanges{
			Int:   Range{Start: intStart, End: intEnd},
			Float: Range{Start: floatStart, End: floatEnd},
		},
	}

	// Agregar la función al directorio de funciones
	functionDirectory[id] = *funcNode

	// Limpiar el contexto de función actual
	scope = global

	return funcNode, nil
}

func ExecuteFunction(funcNode *FuncNode) error {
	// Limpiar variables locales anteriores
	memory.Local.Clear(funcNode.Range.Int.Start, funcNode.Range.Int.End)
	memory.Local.Clear(funcNode.Range.Float.Start, funcNode.Range.Float.End)

	// Establecer el ámbito actual a la función
	scope = funcNode.Id

	// Ejecutar las instrucciones del cuerpo
	for _, stmt := range funcNode.Body {
		if err := ExecuteStatement(stmt); err != nil {
			return fmt.Errorf("error al ejecutar en función '%s': %v", funcNode.Id, err)
		}
	}

	// Restablecer el ámbito global
	scope = global

	return nil
}

func ExecuteStatement(stmt Attrib) error {
	switch node := stmt.(type) {
	case *AssignNode:
		return ExecuteAssign(node)
	case *PrintNode:
		return ExecutePrint(node)
	//case *IfNode:
	//	return executeCondition(node)
	//case *WhileNode:
	//	return executeWhile(node)
	//case *FuncCallNode:
	//	return executeFunctionCall(node)
	default:
		return fmt.Errorf("tipo de instrucción no soportado: %T", node)
	}
}

// Función para ejecutar la asignación
func ExecuteAssign(assignNode *AssignNode) error {
	// Verificar si la variable está declarada
	var info *VarNode
	var found bool

	if scope == global {
		info, found = memory.Global.FindByName(assignNode.Id, functionDirectory[global].Range.Int.Start, functionDirectory[global].Range.Int.End)
	} else {
		info, found = memory.Local.FindByName(assignNode.Id, functionDirectory[scope].Range.Int.Start, functionDirectory[scope].Range.Int.End)
	}

	if !found {
		return fmt.Errorf("variable '%s' no declarada", assignNode.Id)
	}

	// Genera el código intermedio para la expresión
	ctx := &Context{}

	if err := assignNode.Generate(ctx); err != nil {
		return err
	}

	// Si hay cuádruplos generados, se evalúan
	var result VarNode

	if len(ctx.Quads) > 0 {
		PrintQuads(ctx.Quads)
		result = ctx.Evaluate()
	} else {
		// No hay cuádruplos: la pila semántica solo tiene la constante o id
		result = ctx.Pop()
	}

	// Verificar compatibilidad de tipos
	if info.Type != result.Type {
		return fmt.Errorf("tipo incompatible: se esperaba %s, se obtuvo %s", info.Type, result.Type)
	}

	// Actualizar el valor de la variable en la memoria
	info.Value = result.Value

	if scope == global {
		memory.Global.Update(info)
	} else {
		memory.Local.Update(info)
	}

	return nil
}

// Evalúa e imprime cada elemento de una lista
func ExecutePrint(printNode *PrintNode) error {
	for _, item := range printNode.Items {
		switch v := item.(type) {

		// Caso 1: es una expresión/constante numérica
		case Quad:
			// Genera el código intermedio para la expresión
			ctx := &Context{}

			if _, err := v.Generate(ctx); err != nil {
				return err
			}

			// Si hay cuádruplos generados, se evalúan
			var result VarNode

			if len(ctx.Quads) > 0 {
				PrintQuads(ctx.Quads)
				result = ctx.Evaluate()
			} else {
				// No hay cuádruplos: la pila semántica solo tiene la constante o id
				result = ctx.Pop()
			}
			fmt.Print(result.Value)

		// Caso 2: es un literal de cadena
		case string:
			// Imprimir la cadena sin comillas
			fmt.Print(v[1 : len(v)-1])

		default:
			return fmt.Errorf("elemento de print no soportado: %T", item)
		}

		// Agregar espacio entre elementos
		fmt.Print(" ")
	}

	// Salto de línea final
	fmt.Println()
	return nil
}

// Imprime todas las variables
func PrintVariables() {
	fmt.Println()
	fmt.Println("Variables registradas:")
	fmt.Println("===================================")

	fmt.Println("Global:")
	memory.Global.Print()
	fmt.Println("===================================")

	fmt.Println("Local:")
	memory.Local.Print()
	fmt.Println("===================================")

	fmt.Println("Constantes:")
	memory.Const.Print()
	fmt.Println("===================================")

	fmt.Println("Temporales:")
	memory.Temp.Print()
	fmt.Println("===================================")

	fmt.Println()
	fmt.Println("Funciones registradas:")
	fmt.Println("===================================")
	for id, funcNode := range functionDirectory {
		fmt.Printf("Función: %s\n", id)
		fmt.Printf("Rango de memoria: %d - %d\n", funcNode.Range.Int.Start, funcNode.Range.Int.End)
		fmt.Println("===================================")
	}
}
