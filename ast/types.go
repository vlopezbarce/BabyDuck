package ast

var memory *Memory                   // Memoria virtual para variables y constantes
var alloc *Allocator                 // Asignador de memoria para variables
var funcDir = map[string]*FuncNode{} // Tabla de funciones registradas

// Attrib es la interfaz general para todo tipo en el árbol AST
type Attrib interface {
	Generate(ct *Compilation) error
}

// Nodo de programa
type ProgramNode struct {
	Id    string
	Vars  []*VarNode
	Funcs []*FuncNode
	Body  []Attrib
}

// Nodo de función
type FuncNode struct {
	Id            string
	Params        []*VarNode
	Vars          []*VarNode
	Temps         []*VarNode
	Body          []Attrib
	QuadStart     int
	ReturnType    string
	ReturnAddress int
}

// Nodo de variable
type VarNode struct {
	Address int
	Id      string
	Type    string
	Value   string
}

// Nodo de asignación
type AssignNode struct {
	Id  string
	Exp Attrib
}

// Nodo de impresión
type PrintNode struct {
	Items []Attrib
}

// Nodo de expresión binaria
type ExpressionNode struct {
	Op    int
	Left  Attrib
	Right Attrib
}

// Nodo auxiliar para variables en expresiones
type ExpressionVar struct {
	Id string
}

// Nodo de condición
type IfNode struct {
	Condition Attrib
	ThenBlock []Attrib
	ElseBlock []Attrib
}

// Nodo de ciclo while
type WhileNode struct {
	Condition Attrib
	Body      []Attrib
}

// Nodo de llamada a función
type FCallNode struct {
	Id     string
	Params []Attrib
}

// Nodo de retorno de función
type ReturnNode struct {
	Exp Attrib
}
