package ast

import "fmt"

var scope string
var global string

// Inicializa el ámbito global, la memoria y el asignador
func InitProgram(id string) string {
	scope = id
	global = scope
	NewMemory()
	NewAllocator()
	return id
}

func ValidateVars(vars []*VarNode) error {
	tempVars := make(map[string]bool)

	for _, v := range vars {
		// Verificar si la variable ya existe en el ámbito actual
		if _, exists := tempVars[v.Id]; exists {
			return fmt.Errorf("variable '%s' ya declarada en el ámbito actual", v.Id)
		}
		// Agregar la variable al mapa temporal para validación
		tempVars[v.Id] = true
	}

	return nil
}

func DeclareFunction(id string, vars []*VarNode, body []Attrib) (*FuncNode, error) {
	// Verificar si la función ya existe
	if _, exists := funcDir[id]; exists {
		return nil, fmt.Errorf("función '%s' ya declarada", id)
	}

	// Crear el nodo de función
	funcNode := &FuncNode{
		Id:   id,
		Vars: vars,
		Body: body,
	}

	// Agregar la función al directorio de funciones
	funcDir[id] = funcNode

	// Verificar si hay variables duplicadas
	if err := ValidateVars(vars); err != nil {
		return nil, err
	}

	return funcNode, nil
}

func DeclareVariable(id, typ string) error {
	// Obtener la dirección de memoria para la variable
	var addr int
	var err error

	if scope == global {
		addr, err = alloc.NextGlobal(typ)
	} else {
		addr, err = alloc.NextLocal(typ)
	}
	if err != nil {
		return err
	}

	// Crear el nodo de variable
	node := &VarNode{
		Address: addr,
		Id:      id,
		Type:    typ,
	}

	// Insertar en el árbol correspondiente
	if scope == global {
		memory.Global.Insert(node)
	} else {
		memory.Local.Insert(node)
	}

	return nil
}

func ExecuteFunction(funcNode *FuncNode) error {
	// Establecer el ámbito actual a la función
	scope = funcNode.Id

	// Crear variables dentro del ámbito actual
	for _, v := range funcNode.Vars {
		if err := DeclareVariable(v.Id, v.Type); err != nil {
			return fmt.Errorf("error al declarar variable '%s' en función '%s': %v", v.Id, funcNode.Id, err)
		}
	}

	// Generar cuádruplos para el cuerpo de la función
	ctx := &Context{}
	for _, stmt := range funcNode.Body {
		if err := stmt.Generate(ctx); err != nil {
			return fmt.Errorf("error al generar cuádruplos para '%s': %v", funcNode.Id, err)
		}
	}

	// Imprimir cuádruplos y temporales
	ctx.PrintQuads()

	// Ejecutar el cuerpo de la función
	if err := ctx.Evaluate(); err != nil {
		return fmt.Errorf("error al ejecutar cuádruplos de '%s': %v", funcNode.Id, err)
	}

	// Limpiar la memoria
	if scope != global {
		memory.Local.Clear()
	}
	//memory.Temp.Clear()

	// Restablecer el ámbito global
	scope = global

	return nil
}

func (n *VarNode) Generate(ctx *Context) error {
	if n.Id != "" {
		// Buscar en la memoria local o global
		var varNode *VarNode
		var found bool

		if scope != global {
			varNode, found = memory.Local.FindByName(n.Id)
		}
		if !found {
			varNode, found = memory.Global.FindByName(n.Id)
		}
		if !found {
			return fmt.Errorf("variable '%s' no declarada", n.Id)
		}

		// Agregar la direción a la pila
		ctx.Push(varNode.Address)

		return nil
	} else {
		// Buscar la constante en la memoria
		varNode, found := memory.Const.FindConst(n.Type, n.Value)

		if !found {
			// Obtener la dirección de memoria para la constante
			addr, err := alloc.NextConst(n.Type)
			if err != nil {
				return err
			}

			// Agregar la constante a la memoria
			constNode := &VarNode{
				Address: addr,
				Type:    n.Type,
				Value:   n.Value,
			}

			// Agregar la constante a la memoria
			memory.Const.Insert(constNode)

			// Agregar la dirección a la pila
			ctx.Push(addr)

			return nil
		}

		// Agregar la dirección a la pila
		ctx.Push(varNode.Address)

		return nil
	}
}

func (n *AssignNode) Generate(ctx *Context) error {
	// Buscar variable destino y memoria correcta
	var dest *VarNode
	var found bool

	if scope != global {
		if dest, found = memory.Local.FindByName(n.Id); !found {
			return fmt.Errorf("variable '%s' no declarada en el ámbito actual", n.Id)
		}
	} else {
		if dest, found = memory.Global.FindByName(n.Id); !found {
			return fmt.Errorf("variable '%s' no declarada en el ámbito actual", n.Id)
		}
	}

	// Generar el código intermedio para la expresión
	if err := n.Exp.Generate(ctx); err != nil {
		return err
	}
	result := ctx.Pop()

	// Agregar el cuádruplo de asignación
	ctx.AddQuad(ASSIGN, result, -1, dest.Address)

	return nil
}

func (n *PrintNode) Generate(ctx *Context) error {
	// Generar el código intermedio para los elementos a imprimir
	for _, item := range n.Items {
		if err := item.Generate(ctx); err != nil {
			return err
		}
		result := ctx.Pop()

		// Agregar el cuádruplo de impresión
		ctx.AddQuad(PRINT, result, -1, -1)
	}

	// Agregar el cuádruplo de nueva línea
	ctx.AddQuad(PRINTLN, -1, -1, -1)

	return nil
}

func (n *ExpressionNode) Generate(ctx *Context) error {
	// Generar el código intermedio para los operandos izquierdo y derecho
	if err := n.Left.Generate(ctx); err != nil {
		return err
	}
	if err := n.Right.Generate(ctx); err != nil {
		return err
	}

	// Obtener los operandos izquierdo y derecho
	right := ctx.Pop()
	left := ctx.Pop()

	// Obtener los nodos de memoria correspondientes
	leftNode, err := lookupVarByAddress(left)
	if err != nil {
		return err
	}
	rightNode, err := lookupVarByAddress(right)
	if err != nil {
		return err
	}

	// Verificar la compatibilidad de tipos
	resultType, err := CheckSemantic(n.Op, leftNode.Type, rightNode.Type)
	if err != nil {
		return err
	}

	// Obtener la dirección de memoria para el temporal
	addr, err := alloc.NextTemp(resultType)
	if err != nil {
		return err
	}

	// Crear un nuevo nodo temporal
	tempId := ctx.NewTemp()

	tempNode := &VarNode{
		Address: addr,
		Id:      tempId,
		Type:    resultType,
		Value:   tempId,
	}

	// Insertar el temporal en la memoria
	memory.Temp.Insert(tempNode)

	// Agregar el cuádruplo de la operación
	ctx.AddQuad(n.Op, left, right, addr)

	// Agregar el temporal a la pila
	ctx.Push(addr)

	return nil
}

func (n *IfNode) Generate(ctx *Context) error {
	// Generar el código intermedio para la condición
	if err := n.Condition.Generate(ctx); err != nil {
		return err
	}
	result := ctx.Pop()

	// Buscar el tipo del resultado de la condición
	resultNode, _ := memory.Temp.FindByAddress(result)
	if resultNode.Type != "bool" {
		return fmt.Errorf("tipo incompatible en condición if: se esperaba bool, se obtuvo %s", resultNode.Type)
	}

	// Agregar el cuádruplo GOTOF
	indexGOTOF := len(ctx.Quads)
	ctx.AddQuad(GOTOF, result, -1, -1)

	// Generar los cuádruplos para el bloque Then
	for _, stmt := range n.ThenBlock {
		if err := stmt.Generate(ctx); err != nil {
			return err
		}
	}

	// Agregar el cuádruplo GOTO
	indexGOTO := len(ctx.Quads)
	ctx.AddQuad(GOTO, -1, -1, -1)

	// Marcar la etiqueta para el cuádruplo GOTOF
	ctx.Quads[indexGOTOF].Result = len(ctx.Quads)

	// Generar los cuádruplos para el bloque Else
	for _, stmt := range n.ElseBlock {
		if err := stmt.Generate(ctx); err != nil {
			return err
		}
	}

	// Marcar la etiqueta para el cuádruplo GOTO
	ctx.Quads[indexGOTO].Result = len(ctx.Quads)

	return nil
}

func (n *WhileNode) Generate(ctx *Context) error {
	// Marcar el inicio del ciclo
	start := len(ctx.Quads)

	// Generar el código intermedio para el ciclo
	if err := n.Condition.Generate(ctx); err != nil {
		return err
	}
	result := ctx.Pop()

	// Buscar el tipo del resultado de la condición
	resultNode, _ := memory.Temp.FindByAddress(result)
	if resultNode.Type != "bool" {
		return fmt.Errorf("tipo incompatible en condición if: se esperaba bool, se obtuvo %s", resultNode.Type)
	}

	// Agregar el cuádruplo GOTOF
	indexGOTOF := len(ctx.Quads)
	ctx.AddQuad(GOTOF, result, -1, -1)

	// Generar los cuádruplos para el cuerpo del ciclo
	for _, stmt := range n.Body {
		if err := stmt.Generate(ctx); err != nil {
			return err
		}
	}

	// Agregar el cuádruplo GOTO
	ctx.AddQuad(GOTO, -1, -1, start)

	// Marcar la etiqueta para el cuádruplo GOTOF
	ctx.Quads[indexGOTOF].Result = len(ctx.Quads)

	return nil
}

func PrintVariables() {
	fmt.Println()
	fmt.Println("Funciones registradas:")
	fmt.Println("===================================")
	for id := range funcDir {
		fmt.Println(id)
	}

	fmt.Println()
	fmt.Println("Variables registradas:")
	fmt.Println("===================================")

	fmt.Println("Global:")
	memory.Global.Print()
	fmt.Println("===================================")

	fmt.Println("Local:")
	memory.Local.Print()
	fmt.Println("===================================")

	fmt.Println("Constantes:")
	memory.Const.Print()
	fmt.Println("===================================")

	fmt.Println("Temporales:")
	memory.Temp.Print()
	fmt.Println("===================================")
}
