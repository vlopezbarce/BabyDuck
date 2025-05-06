package ast

// Attrib es la interfaz general para todo tipo en el árbol AST
type Attrib interface{}

// Tabla de funciones registradas
var functionDirectory = map[string]FuncNode{}

// Nodo de variable
type VarNode struct {
	Type  string
	Value Attrib
}

// Nodo de función
type FuncNode struct {
	Id          string
	Parameters  []*ParamNode
	Body        []Attrib
	SymbolTable map[string]VarNode // Tabla de símbolos para almacenar información sobre variables
}

// Nodo de parámetro
type ParamNode struct {
	Id   Attrib
	Type Attrib
}

// Nodo de asignación
type AssignNode struct {
	Id  Attrib
	Exp Attrib
}

// Nodo de expresión
type ExpNode struct {
	Type  string
	Value Attrib
}

// Nodo de impresión
type PrintNode struct {
	PrintList []Attrib
}
