package ast

import (
	"BabyDuck_A00833578/token"
	"fmt"
)

var currentScope = "" // Ámbito actual

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
	previousScope := currentScope
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
	currentScope = previousScope

	return funcNode, nil
}

func ExecuteFunction(funcNode *FuncNode) error {
	previousScope := currentScope
	currentScope = funcNode.Id

	// Ejecutar las instrucciones del cuerpo
	for _, stmt := range funcNode.Body {
		err := ExecuteStatement(stmt)
		if err != nil {
			return fmt.Errorf("error al ejecutar en función '%s': %v", funcNode.Id, err)
		}
	}

	currentScope = previousScope
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

// Ejecutar una asignación
func ExecuteAssign(assignNode *AssignNode) error {
	idTok := assignNode.Id.(*token.Token)
	expNode, _ := assignNode.Exp.Eval()

	varId := string(idTok.Lit)

	var currentFuncNode = functionDirectory[currentScope]

	// Verificar si la variable ya fue declarada en el ámbito actual
	info, found := LookupVariable(varId)
	if !found {
		return fmt.Errorf("variable '%s' no declarada", varId)
	}

	// Verificar compatibilidad de tipos
	if info.Type != expNode.Type {
		return fmt.Errorf("tipo incompatible en asignación a '%s'", varId)
	}

	// Asignar el valor a la variable
	info.Value = expNode.Value
	currentFuncNode.SymbolTable[varId] = info

	return nil
}

// Función para procesar la instrucción Print
func ExecutePrint(printNode *PrintNode) error {
	for _, exp := range printNode.PrintList {
		// Imprime el valor de la expresión
		evaluated, err := exp.Eval()
		if err != nil {
			return err
		}

		fmt.Print(evaluated.Value, " ")
	}
	// Salto de línea al final
	fmt.Println()
	return nil
}

// Función para evaluar expresiones
func (e *ExpNode) Eval() (*ExpNode, error) {
	if e.Type == "id" {
		info, found := LookupVariable(e.Value.(string))
		if !found {
			return nil, fmt.Errorf("variable '%s' no declarada", e.Value.(string))
		}
		if info.Value == nil {
			return nil, fmt.Errorf("variable '%s' no inicializada", e.Value.(string))
		}
		return &ExpNode{
			Type:  info.Type,
			Value: info.Value,
		}, nil
	}
	// Si ya tiene valor (int, float, bool, string), devolver directamente
	return e, nil
}

// Función para evaluar expresiones binarias
func (e *ExpressionNode) Eval() (*ExpNode, error) {
	leftEval, err := e.Left.Eval()
	if err != nil {
		return nil, err
	}
	rightEval, err := e.Right.Eval()
	if err != nil {
		return nil, err
	}
	return CompareExpressions(e.Operator, leftEval, rightEval)
}

// Comparar dos expresiones utilizando el operador relacional
func CompareExpressions(op Attrib, left, right *ExpNode) (*ExpNode, error) {
	operatorTok := op.(*token.Token)
	operator := string(operatorTok.Lit)

	// Verificar la compatibilidad de tipos utilizando el semanticCube
	resultType, err := CheckSemantic(operator, left.Type, right.Type)
	if err != nil {
		return nil, err
	}

	// Verificar que ambos valores sean del tipo correcto para la comparación
	var result bool
	switch operator {
	case ">":
		switch left.Type {
		case "int":
			result = left.Value.(int) > right.Value.(int)
		case "float":
			result = left.Value.(float64) > right.Value.(float64)
		}
	case "<":
		switch left.Type {
		case "int":
			result = left.Value.(int) < right.Value.(int)
		case "float":
			result = left.Value.(float64) < right.Value.(float64)
		}
	case "!=":
		switch left.Type {
		case "int":
			result = left.Value.(int) != right.Value.(int)
		case "float":
			result = left.Value.(float64) != right.Value.(float64)
		}
	default:
		return nil, fmt.Errorf("operador '%s' no soportado", operator)
	}

	return &ExpNode{
		Type:  resultType,
		Value: result,
	}, nil
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
