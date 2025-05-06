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

// Crear un parámetro para una función
func NewParameter(id, typ Attrib) (*ParamNode, error) {
	return &ParamNode{
		Id:   id,
		Type: typ,
	}, nil
}

// Crear un nodo de asignación
func NewAssign(id, exp Attrib) (*AssignNode, error) {
	return &AssignNode{
		Id:  id,
		Exp: exp,
	}, nil
}

// Ejecutar una asignación
func ExecuteAssign(assignNode *AssignNode) error {
	idTok := assignNode.Id.(*token.Token)
	expNode := assignNode.Exp.(*ExpNode)

	varId := string(idTok.Lit)

	var currentFuncNode = functionDirectory[currentScope]

	// Verificar si la variable ya fue declarada en el ámbito actual
	info, found := LookupVariable(varId)
	if !found {
		return fmt.Errorf("variable '%s' no declarada", varId)
	}

	// Verificar compatibilidad de tipos
	if info.Type != expNode.Type {
		return fmt.Errorf("tipo incompatible en asignación a '%s'", string(idTok.Lit))
	}

	// Asignar el valor a la variable
	info.Value = expNode.Value
	currentFuncNode.SymbolTable[varId] = info

	return nil
}

// Crear un nodo de impresión
func NewPrint(printList []Attrib) (*PrintNode, error) {
	return &PrintNode{
		PrintList: printList,
	}, nil
}

// Función para procesar la instrucción Print
func ExecutePrint(printNode *PrintNode) error {
	fmt.Print("Print: ", printNode.PrintList)
	for i, exp := range printNode.PrintList {
		if exp == nil {
			return fmt.Errorf("printList[%d] es nil", i)
		}

		switch v := exp.(type) {
		case *ExpNode:
			if v == nil {
				return fmt.Errorf("printList[%d] es *ExpNode nil", i)
			}
			if v.Type == "id" {
				varId := v.Value.(string)
				info, found := LookupVariable(varId)
				if !found {
					return fmt.Errorf("variable no declarada: %s", varId)
				}
				if info.Value == nil {
					return fmt.Errorf("variable '%s' no inicializada", varId)
				}
				fmt.Print(info.Value)
			} else {
				// Si es otra expresión evaluada (int, float, bool, etc.)
				fmt.Print(v.Value)
			}
		case *token.Token:
			if v == nil {
				return fmt.Errorf("printList[%d] es *token.Token nil", i)
			}
			// Esto sería solo para cte_string directamente en el print
			str := string(v.Lit)
			if len(str) >= 2 && str[0] == '"' && str[len(str)-1] == '"' {
				// Quitar comillas si están presentes
				fmt.Print(str[1 : len(str)-1])
			} else {
				fmt.Print(str)
			}
		default:
			return fmt.Errorf("tipo no soportado en printList[%d]: %T", i, exp)
		}
		// Agregar un espacio entre las impresiones
		fmt.Print(" ")
	}
	// Salto de línea al final
	fmt.Println()
	return nil
}

// Comparar dos expresiones utilizando el operador relacional.
func CompareExpressions(op Attrib, left, right *ExpNode) (*ExpNode, error) {
	operatorTok := op.(*token.Token)
	operator := string(operatorTok.Lit)

	// Si el lado izquierdo es un id, obtener el valor y su tipo de la symbolTable
	if left.Type == "id" {
		symInfo, _ := LookupVariable(left.Value.(string))
		left.Value = symInfo.Value
		left.Type = symInfo.Type
	}

	// Si el lado derecho es un id, obtener el valor y su tipo de la symbolTable
	if right.Type == "id" {
		symInfo, _ := LookupVariable(right.Value.(string))
		right.Value = symInfo.Value
		right.Type = symInfo.Type
	}

	// Verificar la compatibilidad de tipos utilizando el semanticCube
	resultType, err := CheckSemantic(operator, left.Type, right.Type)
	if err != nil {
		return nil, err
	}

	var result bool

	// Verificar que ambos valores sean del tipo correcto para la comparación
	switch operator {
	case ">":
		if left.Type == "int" {
			result = left.Value.(int) > right.Value.(int)
		} else if left.Type == "float" {
			result = left.Value.(float64) > right.Value.(float64)
		}
	case "<":
		if left.Type == "int" {
			result = left.Value.(int) < right.Value.(int)
		} else if left.Type == "float" {
			result = left.Value.(float64) < right.Value.(float64)
		}
	case "!=":
		if left.Type == "int" {
			result = left.Value.(int) != right.Value.(int)
		} else if left.Type == "float" {
			result = left.Value.(float64) != right.Value.(float64)
		}
	default:
		return nil, fmt.Errorf("operador '%s' no soportado para comparación", operator)
	}

	// El tipo del resultado siempre será "bool" para las comparaciones
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
