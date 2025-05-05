package ast

import (
	"BabyDuck_A00833578/token"
	"fmt"
)

// Declara una variable en el ámbito actual
func NewVariable(id, typ Attrib) error {
	idTok := id.(*token.Token)
	typTok := typ.(*token.Token)

	varId := string(idTok.Lit)
	varType := string(typTok.Lit)

	// Verificar si ya existe en el ámbito actual
	scopeTable := symbolTables[currentScope]
	if _, exists := scopeTable[varId]; exists {
		return fmt.Errorf("variable '%s' ya declarada en ámbito '%s'", varId, currentScope)
	}

	// Registrar la variable en la tabla de símbolos
	scopeTable[varId] = SymbolInfo{
		Type:  varType,
		Scope: currentScope,
		Value: nil,
	}

	return nil
}

// Busca una variable, primero en el ámbito actual y luego en global
func LookupVariable(id string) (SymbolInfo, bool) {
	// Buscar en el ámbito actual
	if info, ok := symbolTables[currentScope][id]; ok {
		return info, true
	}
	// Buscar en ámbito global
	if info, ok := symbolTables["global"][id]; ok {
		return info, true
	}
	return SymbolInfo{}, false
}

// Crear una asignación
func NewAssign(id, exp Attrib) (*AssignNode, error) {
	idTok := id.(*token.Token)
	expTok := exp.(*ExpNode)

	// Buscar la variable en la tabla de símbolos
	info, ok := LookupVariable(string(idTok.Lit))
	if !ok {
		return nil, fmt.Errorf("variable '%s' no declarada", string(idTok.Lit))
	}
	// Verificar compatibilidad de tipos
	if info.Type != expTok.Type {
		return nil, fmt.Errorf("tipo incompatible en asignación a '%s'", string(idTok.Lit))
	}

	// Actualizar el valor en la tabla de símbolos
	scopeTable := symbolTables[currentScope]
	if _, exists := scopeTable[string(idTok.Lit)]; exists {
		scopeTable[string(idTok.Lit)] = SymbolInfo{
			Type:  info.Type,
			Scope: info.Scope,
			Value: expTok.Value,
		}
	}

	return &AssignNode{
		Id:  string(idTok.Lit),
		Exp: expTok,
	}, nil
}

// Función constructora para FuncNode
func NewFunction(id Attrib, params []*ParamNode, body Attrib) (*FuncNode, error) {
	idTok := id.(*token.Token)

	// Guardar la función en la tabla de funciones
	funcId := string(idTok.Lit)
	functionTable[funcId] = FuncNode{
		Id:         funcId,
		Parameters: params,
		Body:       body,
	}

	// Entrar al nuevo ámbito
	EnterScope(string(idTok.Lit))

	// Guardar parámetros en la tabla de símbolos
	for _, param := range params {
		if err := NewVariable(param.Id, param.Type); err != nil {
			return nil, err
		}
	}

	// Salir del ámbito
	ExitScope()

	return &FuncNode{
		Id:         funcId,
		Parameters: params,
		Body:       body,
	}, nil
}

// Crear un parámetro para una función
func NewParameter(id, typ Attrib) (*ParamNode, error) {
	return &ParamNode{
		Id:   id,
		Type: typ,
	}, nil
}

// Imprime todas las variables en el ámbito actual
func PrintVariables() {
	fmt.Println("Variables en el ámbito actual:")
	for name, info := range symbolTables[currentScope] {
		fmt.Printf("Variable: %s, Tipo: %s, Ámbito: %s, Valor: %v\n", name, info.Type, info.Scope, info.Value)
	}
}

// Imprime todas las funciones registradas
func PrintFunctions() {
	fmt.Println("Funciones registradas:")
	for name, funcInfo := range functionTable {
		fmt.Printf("Función: %s, Parámetros: %v\n", name, funcInfo.Parameters)
	}
}
