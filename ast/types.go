package ast

// Attrib es la interfaz general para todo tipo en el árbol AST
type Attrib interface{}

// Tabla de funciones registradas
var funcDir = map[string]*FuncNode{}

// Nodo de función
type FuncNode struct {
	Id   string
	Vars []*VarNode
	Body []Attrib
}

// Memoria de direcciones virtuales
type Memory struct {
	Global *SymbolTree
	Const  *SymbolTree
	Temp   *SymbolTree
	Local  *SymbolTree
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

// Gestiona la asignación de direcciones de memoria
type Allocator struct {
	Operators Range
	Global    MemoryRanges
	Local     MemoryRanges
	Const     MemoryRanges
	Temp      MemoryRanges
}

// Rangos para tipos de datos
type MemoryRanges struct {
	Int    Range
	Float  Range
	Bool   Range
	String Range
}

// Rango de memoria para funciones
type Range struct {
	Start   int
	End     int
	Counter int
}

// Representa una instrucción de código intermedio (cuádruplo)
type Quadruple struct {
	Operator int
	Left     int
	Right    int
	Result   int
}

// Almacena la pila semántica, cuádruplos y contador de temporales
type Context struct {
	SemStack  []int
	Quads     []Quadruple
	TempCount int
}

// Interfaz para nodos que pueden generar cuádruplos
type Quad interface {
	Generate(ctx *Context) error
}

// Nodo de asignación
type AssignNode struct {
	Id  string
	Exp Quad
}

// Nodo de impresión
type PrintNode struct {
	Item Quad
}

// Nodo de expresión binaria
type ExpressionNode struct {
	Op    int
	Left  Quad
	Right Quad
}

// Nodo de condición
type IfNode struct {
	Condition Quad
	ThenBlock []Attrib
	ElseBlock []Attrib
}

// Nodo de ciclo while
type WhileNode struct {
	Condition Quad
	Body      []Attrib
}
