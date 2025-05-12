package ast

import (
	"BabyDuck_A00833578/token"
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
func NewVariable(id, typ Attrib) error {
	idTok := id.(*token.Token)
	typTok := typ.(*token.Token)

	varId := string(idTok.Lit)
	varType := string(typTok.Lit)

	var currentFuncNode = functionDirectory[currentScope]

	// Verificar si la variable ya existe en el ámbito actual
	_, found := LookupVariable(varId)
	if found {
		return fmt.Errorf("variable '%s' ya declarada en función '%s'", varId, currentScope)
	}

	// Agregar la variable a la tabla de símbolos del ámbito actual
	currentFuncNode.SymbolTable[varId] = VarNode{
		Type:  varType,
		Value: nil,
	}

	return nil
}

// Buscar una variable en el ámbito actual
func LookupVariable(varId string) (VarNode, bool) {
	var currentFuncNode = functionDirectory[currentScope]

	// Buscar en la tabla de símbolos del ámbito actual
	if info, exists := currentFuncNode.SymbolTable[varId]; exists {
		return info, true
	}

	if currentScope != globalScope {
		// Buscar en el ámbito global
		if info, exists := functionDirectory[globalScope].SymbolTable[varId]; exists {
			return info, true
		}
	}

	// No se encontró
	return VarNode{}, false
}

// Función constructora para FuncNode
func NewFunction(id Attrib, params []*ParamNode, body []Attrib) (*FuncNode, error) {
	idTok := id.(*token.Token)
	funcId := string(idTok.Lit)

	// Verificar si la función ya existe
	if _, exists := functionDirectory[funcId]; exists {
		return nil, fmt.Errorf("función '%s' ya declarada", funcId)
	}

	// Crear nodo de función con su tabla local vacía
	funcNode := &FuncNode{
		Id:          funcId,
		Parameters:  params,
		Body:        body,
		SymbolTable: make(map[string]VarNode),
	}

	// Establecer función actual para contexto de variables
	currentScope = funcId

	// Registrar la función en el directorio
	functionDirectory[funcId] = *funcNode

	// Registrar los parámetros como variables locales
	for _, param := range params {
		err := NewVariable(param.Id, param.Type)
		if err != nil {
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
		varNode.Value = nil
		funcNode.SymbolTable[name] = varNode
	}

	// Establecer el ámbito actual a la función
	currentScope = funcNode.Id

	// Ejecutar las instrucciones del cuerpo
	for _, stmt := range funcNode.Body {
		err := ExecuteStatement(stmt)
		if err != nil {
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
	//case *PrintNode:
	//	return ExecutePrint(node)
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
	ctx := &Context{}

	// Genera el código intermedio para la expresión
	if err := assignNode.Generate(ctx); err != nil {
		return err
	}

	idTok := assignNode.Id.(*token.Token)
	varId := string(idTok.Lit)

	// Obtener el valor final de la expresión
	PrintQuads(ctx.Quads)
	resultNode := ctx.Evaluate()

	// Actualizar la tabla de símbolos con el valor calculado
	info, exists := LookupVariable(varId)
	if !exists {
		return fmt.Errorf("variable no declarada: %s", varId)
	}

	// Verificar compatibilidad de tipos

	// Asignar el valor a la variable en la tabla de símbolos
	info.Value = resultNode.Value
	functionDirectory[currentScope].SymbolTable[varId] = info

	return nil
}

// Imprime todas las variables
func PrintVariables() {
	fmt.Println("Variables registradas:")
	for name, funcNode := range functionDirectory {
		for varName, varNode := range funcNode.SymbolTable {
			fmt.Printf("Función: %s, Variable: %s, Tipo: %s, Valor: %v\n", name, varName, varNode.Type, varNode.Value)
		}
	}
}
