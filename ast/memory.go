package ast

import "fmt"

var memory *Memory

func NewMemory() {
	memory = &Memory{
		Operators: &SymbolTree{Root: nil},
		Global:    &SymbolTree{Root: nil},
		Local:     &SymbolTree{Root: nil},
		Const:     &SymbolTree{Root: nil},
		Temp:      &SymbolTree{Root: nil},
	}
}

// Llenar el árbol de operadores con los operadores disponibles
func FillOperatorsTree() {
	operators := []string{"+", "-", "*", "/", ">", "<", "!=", "="}
	for addr, op := range operators {
		node := &VarNode{
			Address: addr,
			Id:      op,
			Type:    "operator",
			Value:   op,
		}
		memory.Operators.Insert(node)
	}
}

// Inserta un nuevo nodo en el árbol de símbolos
func (tree *SymbolTree) Insert(newNode *VarNode) {
	tree.Root = insertNode(tree.Root, newNode)
}

// Función auxiliar para insertar un nuevo nodo
func insertNode(currNode, newNode *VarNode) *VarNode {
	if currNode == nil {
		return newNode
	}
	if newNode.Address < currNode.Address {
		currNode.Left = insertNode(currNode.Left, newNode)
	} else {
		currNode.Right = insertNode(currNode.Right, newNode)
	}
	return currNode
}

// Actualiza un nodo existente en el árbol de símbolos
func (tree *SymbolTree) Update(newNode *VarNode) {
	updateNode(tree.Root, newNode)
}

// Función auxiliar para actualizar un nodo
func updateNode(currNode, newNode *VarNode) error {
	if currNode == nil {
		return nil
	}
	if currNode.Address == newNode.Address {
		currNode.Value = newNode.Value
		return nil
	}
	if newNode.Address < currNode.Address {
		return updateNode(currNode.Left, newNode)
	}
	return updateNode(currNode.Right, newNode)
}

// Busca en local primero y luego en global, retorna qué memoria usar
func lookupVar(name string) (*VarNode, *SymbolTree, error) {
	if scope != global {
		if info, found := memory.Local.FindByName(name); found {
			return info, memory.Local, nil
		}
	}
	if info, found := memory.Global.FindByName(name); found {
		return info, memory.Global, nil
	}
	return nil, nil, fmt.Errorf("variable '%s' no declarada", name)
}

// Busca una variable por su ID en el árbol de símbolos
func (tree *SymbolTree) FindByName(id string) (*VarNode, bool) {
	return findByName(tree.Root, id)
}

// Función auxiliar para buscar una variable por su ID
func findByName(node *VarNode, id string) (*VarNode, bool) {
	if node == nil {
		return nil, false
	}
	if node.Id == id {
		return node, true
	}

	// Revisar subárbol izquierdo
	if leftNode, found := findByName(node.Left, id); found {
		return leftNode, true
	}

	// Revisar subárbol derecho
	return findByName(node.Right, id)
}

// Busca una variable por su dirección en el árbol de símbolos
func (tree *SymbolTree) FindByAddress(address int) (*VarNode, bool) {
	return findByAddress(tree.Root, address)
}

// Función auxiliar para buscar una variable por su dirección
func findByAddress(node *VarNode, address int) (*VarNode, bool) {
	if node == nil {
		return nil, false
	}
	if node.Address == address {
		return node, true
	}
	if address < node.Address {
		return findByAddress(node.Left, address)
	} else {
		return findByAddress(node.Right, address)
	}
}

// Limpia los valores en un árbol de símbolos dentro de un rango
func (tree *SymbolTree) Clear() {
	tree.Root = nil
}

// Imprime el árbol de símbolos
func (tree *SymbolTree) Print() {
	printNode(tree.Root)
}

// Función auxiliar para imprimir el árbol de símbolos
func printNode(node *VarNode) {
	if node == nil {
		return
	}
	printNode(node.Left)
	fmt.Printf("Dirección: %d, Variable: %s, Tipo: %s, Valor: %s\n", node.Address, node.Id, node.Type, node.Value)
	printNode(node.Right)
}
