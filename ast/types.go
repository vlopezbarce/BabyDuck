package ast

// Attrib es la interfaz general para todo tipo en el árbol AST
type Attrib interface{}

// Tabla de funciones registradas
var funcDir = map[string]*FuncNode{}

// Memoria de direcciones virtuales
type Memory struct {
	Operators *SymbolTree
	Global    *SymbolTree
	Const     *SymbolTree
	Temp      *SymbolTree
	Local     *SymbolTree
}

// Estructura del árbol de símbolos
type SymbolTree struct {
	Root *VarNode
}

// Nodo de variable
type VarNode struct {
	Address int
	Id      string
	Type    string
	Value   string
	Left    *VarNode
	Right   *VarNode
}

// Rango de memoria para funciones
type Range struct {
	Start   int
	End     int
	Counter int
}

// Rangos para tipos de datos
type MemoryRanges struct {
	Int   Range
	Float Range
	Bool  Range
}

// Nodo de función
type FuncNode struct {
	Id   string
	Vars []*VarNode
	Body []Attrib
}

// Interfaz para nodos que pueden generar cuádruplos
type Quad interface {
	Generate(ctx *Context) (string, error)
}

// Nodo de asignación
type AssignNode struct {
	Id  string
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
