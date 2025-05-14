package ast

import (
	"fmt"
)

var globalScope string  // Nombre del programa
var currentScope string // Ámbito actual

// Inicializa el ámbito global y establece el ámbito actual
func SetGlobalScope(name string) {
	globalScope = name
	currentScope = name
}

// Declara una variable en el ámbito actual
func NewVariable(id, typ string) error {
	// Verificar si la variable ya existe en el ámbito actual
	if _, found := LookupVariable(id); found {
		return fmt.Errorf("variable '%s' ya declarada en función '%s'", id, currentScope)
	}

	// Agregar la variable a la tabla de símbolos del ámbito actual
	functionDirectory[currentScope].SymbolTable[id] = VarNode{
		Id:    id,
		Type:  typ,
		Value: "",
	}

	return nil
}

// Buscar una variable en el ámbito actual
func LookupVariable(id string) (VarNode, bool) {
	// Buscar en la tabla de símbolos del ámbito actual
	if info, exists := functionDirectory[currentScope].SymbolTable[id]; exists {
		return info, true
	}

	if currentScope != globalScope {
		// Buscar en el ámbito global
		if info, exists := functionDirectory[globalScope].SymbolTable[id]; exists {
			return info, true
		}
	}

	// No se encontró
	return VarNode{}, false
}

// Función constructora para FuncNode
func NewFunction(id string, vars []*VarNode, body []Attrib) (*FuncNode, error) {
	// Verificar si la función ya existe
	if _, exists := functionDirectory[id]; exists {
		return nil, fmt.Errorf("función '%s' ya declarada", id)
	}

	// Crear el nodo de función
	funcNode := &FuncNode{
		Id:          id,
		Body:        body,
		SymbolTable: make(map[string]VarNode),
	}

	// Agregar la función al directorio de funciones
	functionDirectory[id] = *funcNode

	// Establecer el ámbito actual a la nueva función
	currentScope = id

	// Registrar los parámetros como variables locales
	for _, param := range vars {
		if err := NewVariable(param.Id, param.Type); err != nil {
			return nil, err
		}
	}

	// Limpiar el contexto de función actual
	currentScope = globalScope

	return funcNode, nil
}

func ExecuteFunction(funcNode *FuncNode) error {
	// Limpiar variables locales anteriores
	for name, varNode := range funcNode.SymbolTable {
		varNode.Value = ""
		funcNode.SymbolTable[name] = varNode
	}

	// Establecer el ámbito actual a la función
	currentScope = funcNode.Id

	// Ejecutar las instrucciones del cuerpo
	for _, stmt := range funcNode.Body {
		if err := ExecuteStatement(stmt); err != nil {
			return fmt.Errorf("error al ejecutar en función '%s': %v", funcNode.Id, err)
		}
	}

	// Restablecer el ámbito global
	currentScope = globalScope

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
	info, exists := LookupVariable(assignNode.Id)
	if !exists {
		return fmt.Errorf("variable no declarada: %s", assignNode.Id)
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

	// Actualizar la tabla de símbolos con el valor calculado
	info.Value = result.Value
	functionDirectory[currentScope].SymbolTable[assignNode.Id] = info

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

	for name, funcNode := range functionDirectory {
		for varName, varNode := range funcNode.SymbolTable {
			fmt.Printf("Función: %s, Variable: %s, Tipo: %s, Valor: %v\n", name, varName, varNode.Type, varNode.Value)
		}
	}
}
