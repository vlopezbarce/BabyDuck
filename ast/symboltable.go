package ast

// Attrib es la interfaz general para todo tipo en el árbol AST
type Attrib interface{}

// Información sobre una variable
type SymbolInfo struct {
	Type  string
	Scope string
	Value Attrib
}

// Tabla de símbolos organizada por ámbito (scope)
var symbolTables = map[string]map[string]SymbolInfo{
	"global": {},
}

// Tabla de funciones registradas
var functionTable = map[string]FuncNode{}

// Ámbito actual (inicia en global)
var currentScope = "global"

// Cambia al nuevo ámbito
func EnterScope(scope string) {
	currentScope = scope
	if _, exists := symbolTables[scope]; !exists {
		symbolTables[scope] = make(map[string]SymbolInfo)
	}
}

// Sale del ámbito actual y regresa a "global"
func ExitScope() {
	currentScope = "global"
}
