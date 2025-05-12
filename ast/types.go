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

// Interfaz para nodos que pueden generar cuádruplos
type Quad interface {
	Generate(ctx *Context) (string, error)
}

// Nodo de asignación
type AssignNode struct {
	Id  Attrib
	Exp Quad
}

// Nodo de impresión
type PrintNode struct {
	Items []Attrib
}

// Nodo de expresión binaria
type ExpressionNode struct {
	Op    string
	Left  Quad
	Right Quad
}
