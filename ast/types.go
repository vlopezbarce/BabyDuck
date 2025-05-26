package ast

var memory *Memory                   // Memoria virtual para variables y constantes
var alloc *Allocator                 // Asignador de memoria para variables
var funcDir = map[string]*FuncNode{} // Tabla de funciones registradas

// Attrib es la interfaz general para todo tipo en el árbol AST
type Attrib interface {
	Generate(ctx *Context) error
}

// Nodo de función
type FuncNode struct {
	Id          string
	Params      []*VarNode
	Vars        []*VarNode
	Body        []Attrib
	ParamsCount int
	VarsCount   int
	TempCount   int
	QuadStart   int
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
	Global Segment
	Local  Segment
	Const  Segment
	Temp   Segment
}

// Segmento de memoria apartado
type Segment struct {
	Int    Range
	Float  Range
	Bool   Range
	String Range
}

// Rango de memoria para tipos de datos
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
