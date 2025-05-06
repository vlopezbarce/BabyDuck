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

// Interfaz para nodos que pueden ser evaluados
type Evaluable interface {
	Eval() (*ExpNode, error)
}

// Nodo de expresión
type ExpNode struct {
	Type  string
	Value Attrib
}

// Nodo de expresión binaria
type ExpressionNode struct {
	Operator Attrib
	Left     Evaluable
	Right    Evaluable
}

// Nodo de asignación
type AssignNode struct {
	Id  Attrib
	Exp Evaluable
}

// Nodo de impresión
type PrintNode struct {
	PrintList []Evaluable
}
