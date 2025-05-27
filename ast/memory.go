package ast

import "fmt"

// Direcciones fijas para operadores
const (
	PLUS    = 0
	MINUS   = 1
	TIMES   = 2
	DIVIDE  = 3
	GT      = 4
	LT      = 5
	NEQ     = 6
	ASSIGN  = 7
	PRINT   = 8
	PRINTLN = 9
	GOTO    = 10
	GOTOF   = 11
	ERA     = 12
	PARAM   = 13
	GOSUB   = 14
	ENDFUNC = 15
)

// DEBUG: Lista de operadores para imprimir operación
var opsList = []string{
	"+",
	"-",
	"*",
	"/",
	">",
	"<",
	"!=",
	"=",
	"PRINT",
	"PRINTLN",
	"GOTO",
	"GOTOF",
	"ERA",
	"PARAM",
	"GOSUB",
	"ENDFUNC",
}

// Memoria de direcciones virtuales
type Memory struct {
	Global *MemorySegment
	Local  *MemorySegment
	Const  *MemorySegment
	Temp   *MemorySegment
}

// Segmentos de memoria apartados
type MemorySegment struct {
	Int    []*VarNode
	Float  []*VarNode
	Bool   []*VarNode
	String []*VarNode
}

func NewMemory() {
	memory = &Memory{
		Global: &MemorySegment{
			Int:   []*VarNode{},
			Float: []*VarNode{},
		},
		Local: &MemorySegment{
			Int:   []*VarNode{},
			Float: []*VarNode{},
		},
		Const: &MemorySegment{
			Int:    []*VarNode{},
			Float:  []*VarNode{},
			String: []*VarNode{},
		},
		Temp: &MemorySegment{
			Int:   []*VarNode{},
			Float: []*VarNode{},
			Bool:  []*VarNode{},
		},
	}
}

// Obtiene un nodo de memoria por dirección
func GetByAddress(address int, frame *StackFrame) (*VarNode, error) {
	// Obtener el segmento de memoria al que pertenece la dirección
	m, s := alloc.GetSegment(address, frame)

	// Buscar el nodo en el segmento de memoria
	if address >= s.Int.Start && address <= s.Int.End {
		index := address - s.Int.Start
		return m.Int[index], nil
	}
	if address >= s.Float.Start && address <= s.Float.End {
		index := address - s.Float.Start
		return m.Float[index], nil
	}
	if s.Bool != nil && address >= s.Bool.Start && address <= s.Bool.End {
		index := address - s.Bool.Start
		return m.Bool[index], nil
	}
	if s.String != nil && address >= s.String.Start && address <= s.String.End {
		index := address - s.String.Start
		return m.String[index], nil
	}
	return nil, fmt.Errorf("variable con dirección %d no encontrada", address)
}

// Inserta un nuevo nodo en el segmento de memoria
func (m *MemorySegment) Insert(node *VarNode) {
	switch node.Type {
	case "int":
		m.Int = append(m.Int, node)
	case "float":
		m.Float = append(m.Float, node)
	case "bool":
		m.Bool = append(m.Bool, node)
	case "string":
		m.String = append(m.String, node)
	}
}

// Busca una variable por su ID en el segmento de memoria
func (m *MemorySegment) FindByName(id string) (*VarNode, bool) {
	for _, node := range m.Int {
		if node.Id == id {
			return node, true
		}
	}
	for _, node := range m.Float {
		if node.Id == id {
			return node, true
		}
	}
	for _, node := range m.Bool {
		if node.Id == id {
			return node, true
		}
	}
	for _, node := range m.String {
		if node.Id == id {
			return node, true
		}
	}
	return nil, false
}

// Busca una constante por tipo y valor en el segmento de memoria
func (m *MemorySegment) FindConst(typ string, val string) (*VarNode, bool) {
	switch typ {
	case "int":
		for _, node := range m.Int {
			if node.Value == val {
				return node, true
			}
		}
	case "float":
		for _, node := range m.Float {
			if node.Value == val {
				return node, true
			}
		}
	case "bool":
		for _, node := range m.Bool {
			if node.Value == val {
				return node, true
			}
		}
	case "string":
		for _, node := range m.String {
			if node.Value == val {
				return node, true
			}
		}
	}
	return nil, false
}

// Obtiene todos los nodos de un segmento de memoria
func (m *MemorySegment) GetAll() []*VarNode {
	var result []*VarNode
	result = append(result, m.Int...)
	result = append(result, m.Float...)
	result = append(result, m.Bool...)
	result = append(result, m.String...)
	return result
}

// Limpia un segmento de memoria
func (m *MemorySegment) Clear() {
	m.Int = []*VarNode{}
	m.Float = []*VarNode{}
	m.Bool = []*VarNode{}
	m.String = []*VarNode{}
}

// Imprime el segmento de memoria
func (m *MemorySegment) Print() {
	for _, node := range append(m.Int, append(m.Float, append(m.Bool, m.String...)...)...) {
		var nodeId string
		if node.Id != "" {
			nodeId = fmt.Sprintf("  ID: %s", node.Id)
		} else {
			nodeId = ""
		}

		var nodeType string
		if node.Type != "" {
			nodeType = fmt.Sprintf("  TYPE: %s", node.Type)
		} else {
			nodeType = ""
		}

		var nodeValue string
		if node.Value != "" {
			nodeValue = fmt.Sprintf("  VALUE: %s", node.Value)
		} else {
			nodeValue = ""
		}

		fmt.Printf("ADDR: %d%s%s%s\n", node.Address, nodeId, nodeType, nodeValue)
	}
}
