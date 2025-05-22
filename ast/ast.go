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
	FillOperatorsTree()
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

	// Establecer el ámbito actual a la función
	scope = id

	// Verificar si hay variables duplicadas
	if err := ValidateVars(vars); err != nil {
		return nil, err
	}

	// Reestablecer el ámbito global
	scope = global

	return funcNode, nil
}

func DeclareVariable(id, typ string) error {
	// Obtiene la dirección de memoria para la variable
	getAddr := map[bool]map[string]func() (int, error){
		true: {
			"int":   alloc.NextGlobalInt,
			"float": alloc.NextGlobalFloat,
		},
		false: {
			"int":   alloc.NextLocalInt,
			"float": alloc.NextLocalFloat,
		},
	}

	addr, err := getAddr[scope == global][typ]()
	if err != nil {
		return fmt.Errorf("error al asignar dirección para variable '%s': %v", id, err)
	}

	// Crear el nodo de variable
	node := &VarNode{
		Address: addr,
		Id:      id,
		Type:    typ,
	}

	// Insertar en el árbol
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

	// Crear variables locales
	for _, v := range funcNode.Vars {
		if err := DeclareVariable(v.Id, v.Type); err != nil {
			return fmt.Errorf("error al declarar variable '%s' en función '%s': %v", v.Id, funcNode.Id, err)
		}
	}

	// Generar cuádruplos para el cuerpo de la función
	ctx := &Context{}
	for _, stmt := range funcNode.Body {
		if err := GenerateStatement(ctx, stmt); err != nil {
			return fmt.Errorf("error al generar cuádruplos para '%s': %v", funcNode.Id, err)
		}
	}

	// Imprimir cuádruplos y temporales
	ctx.PrintQuads()

	// Ejecutar el cuerpo de la función
	ctx.Evaluate()

	// Limpiar la memoria
	if scope != global {
		memory.Local.Clear()
	}
	//memory.Temp.Clear()

	// Restablecer el ámbito global
	scope = global

	return nil
}

func GenerateStatement(ctx *Context, stmt Attrib) error {
	var err error
	switch stmt := stmt.(type) {
	case *AssignNode:
		err = stmt.Generate(ctx)
	case *[]PrintNode:
		for _, printNode := range *stmt {
			err = printNode.Generate(ctx)
		}
	case *IfNode:
		err = stmt.Generate(ctx)
	}
	if err != nil {
		return err
	}
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
			var addr int
			var err error

			switch n.Type {
			case "int":
				addr, err = alloc.NextConstInt()
			case "float":
				addr, err = alloc.NextConstFloat()
			case "string":
				addr, err = alloc.NextConstString()
			}
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

	// Obtener el resultado de la expresión
	if err := n.Exp.Generate(ctx); err != nil {
		return err
	}
	result := ctx.Pop()

	// Obtener el operador de la memoria
	opNode, _ := memory.Operators.FindByName("=")

	// Agregar el cuádruplo
	ctx.AddQuad(opNode.Address, result, -1, dest.Address)

	return nil
}

func (n *PrintNode) Generate(ctx *Context) error {
	// Obtener el resultado de la expresión
	if err := n.Item.Generate(ctx); err != nil {
		return err
	}
	result := ctx.Pop()

	// Obtener el operador de la memoria
	opNode, _ := memory.Operators.FindByName("PRINT")

	// Agregar el cuádruplo
	ctx.AddQuad(opNode.Address, result, -1, -1)

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

	// Obtener el operador de la memoria
	opNode, _ := memory.Operators.FindByName(n.Op)

	// Obtener la dirección de memoria para el temporal
	var addr int
	switch resultType {
	case "int":
		addr, err = alloc.NextTempInt()
	case "float":
		addr, err = alloc.NextTempFloat()
	case "bool":
		addr, err = alloc.NextTempBool()
	}
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

	// Agregar el cuádruplo
	ctx.AddQuad(opNode.Address, left, right, addr)

	// Agregar el temporal a la pila
	ctx.Push(addr)

	return nil
}

func (n *IfNode) Generate(ctx *Context) error {
	// Generar código intermedio de la condición
	if err := n.Condition.Generate(ctx); err != nil {
		return err
	}
	result := ctx.Pop()

	// Buscar el tipo del resultado de la condición
	resultNode, err := lookupVarByAddress(result)
	if err != nil {
		return err
	}
	if resultNode.Type != "bool" {
		return fmt.Errorf("tipo incompatible en condición if: se esperaba bool, se obtuvo %s", resultNode.Type)
	}

	// Obtener el operador de la memoria
	opGOTOF, _ := memory.Operators.FindByName("GOTOF")
	idGOTOF := len(ctx.Quads)
	ctx.AddQuad(opGOTOF.Address, result, -1, -1)

	// Generar los cuádruplos para el bloque Then
	for _, stmt := range n.ThenBlock {
		if err := GenerateStatement(ctx, stmt); err != nil {
			return err
		}
	}

	// Obtener el operador de la memoria
	opGOTO, _ := memory.Operators.FindByName("GOTO")
	idGOTO := len(ctx.Quads)
	ctx.AddQuad(opGOTO.Address, -1, -1, -1)

	// Marca la etiqueta para el cuádruplo GOTOF
	ctx.Quads[idGOTOF].Result = len(ctx.Quads)

	// Generar los cuádruplos para el bloque Else
	for _, stmt := range n.ElseBlock {
		if err := GenerateStatement(ctx, stmt); err != nil {
			return err
		}
	}

	// Marca la etiqueta para el cuádruplo GOTO
	ctx.Quads[idGOTO].Result = len(ctx.Quads)

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

	fmt.Println("Operators:")
	memory.Operators.Print()
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
