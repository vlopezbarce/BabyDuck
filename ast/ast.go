package ast

import "fmt"

var scope string
var global string
var ctx *Context

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

func ValidateFunction(id string, params, vars []*VarNode, body []Attrib) (*FuncNode, error) {
	// Verificar si la función ya existe
	if _, exists := funcDir[id]; exists {
		return nil, fmt.Errorf("función '%s' ya declarada", id)
	}

	// Crear el nodo de función
	funcNode := &FuncNode{
		Id:          id,
		Params:      params,
		Vars:        vars,
		Body:        body,
		ParamsCount: len(params),
		VarsCount:   len(vars),
	}

	// Agregar la función al directorio de funciones
	funcDir[id] = funcNode

	// Verificar si hay variables duplicadas
	if err := ValidateVars(append(params, vars...)); err != nil {
		return nil, err
	}

	return funcNode, nil
}

func GenerateProgram(id string, vars []*VarNode, funcs []*FuncNode, body []Attrib) error {
	// Inicializar el contexto global, la memoria y el asignador de direcciones
	ctx = &Context{}
	NewMemory()
	NewAllocator()

	// Establecer el ámbito global
	global = id
	scope = global

	// Crear el nodo de función principal
	programNode := &FuncNode{
		Id:        id,
		Vars:      vars,
		Body:      body,
		VarsCount: len(vars),
	}

	// Registrar la función en el directorio de funciones
	funcDir[id] = programNode

	// Verificar si hay variables duplicadas
	if err := ValidateVars(vars); err != nil {
		return err
	}

	// Agregar el cuádruplo de inicio del programa
	ctx.AddQuad(GOTO, -1, -1, -1)

	// Crear variables dentro del ámbito global
	for _, v := range vars {
		if err := DeclareVariable(v.Id, v.Type); err != nil {
			return fmt.Errorf("error al declarar variable '%s': %v", v.Id, err)
		}
	}

	// Generar cuádruplos para las funciones
	for _, funcNode := range funcs {
		if err := funcNode.Generate(ctx); err != nil {
			return fmt.Errorf("error al generar cuádruplos para '%s': %v", funcNode.Id, err)
		}
	}

	// Marcar el inicio del programa
	ctx.Quads[0].Result = len(ctx.Quads)

	// Generar cuádruplos para el cuerpo del programa
	for _, stmt := range body {
		if err := stmt.Generate(ctx); err != nil {
			return fmt.Errorf("error al generar cuádruplos para '%s': %v", id, err)
		}
	}

	// Imprimir variables globales y constantes
	fmt.Println()
	fmt.Printf("Programa: %s\n", id)
	fmt.Println("===================================")

	fmt.Println("Globales:")
	memory.Global.Print()
	fmt.Println("===================================")

	fmt.Println("Temporales:")
	memory.Temp.Print()
	fmt.Println("===================================")

	fmt.Println("Constantes:")
	memory.Const.Print()
	fmt.Println("===================================")

	fmt.Println("Operadores:")
	for op, name := range opsList {
		fmt.Printf("ADDR: %d, ID: %s\n", op, name)
	}
	fmt.Println("===================================")

	// Imprimir cuádruplos generados
	ctx.PrintQuads()

	return nil
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

func (n *FuncNode) Generate(ctx *Context) error {
	// Marcar el inicio del cuádruplo de la función
	funcDir[n.Id].QuadStart = len(ctx.Quads)

	// Establecer el ámbito actual a la función
	scope = n.Id

	// Crear parámetros dentro del ámbito de la función
	for _, p := range n.Params {
		if err := DeclareVariable(p.Id, p.Type); err != nil {
			return fmt.Errorf("error al declarar parámetro '%s' en función '%s': %v", p.Id, n.Id, err)
		}
	}

	// Crear variables dentro del ámbito de la función
	for _, v := range n.Vars {
		if err := DeclareVariable(v.Id, v.Type); err != nil {
			return fmt.Errorf("error al declarar variable '%s' en función '%s': %v", v.Id, n.Id, err)
		}
	}

	// Generar cuádruplos para el cuerpo de la función
	for _, stmt := range n.Body {
		if err := stmt.Generate(ctx); err != nil {
			return fmt.Errorf("error al generar cuádruplos para '%s': %v", n.Id, err)
		}
	}

	// Agregar el cuádruplo de retorno al final de la función
	ctx.AddQuad(ENDFUNC, -1, -1, -1)

	// Actualizar contador de temporales de la función
	funcDir[n.Id].TempsCount = alloc.Temp.Count()

	// Restablecer el ámbito global
	scope = global

	// Imprimir variables locales y temporales de la función
	fmt.Println()
	fmt.Printf("Función: %s\n", n.Id)
	fmt.Println("===================================")

	fmt.Println("Locales:")
	memory.Local.Print()
	fmt.Println("===================================")

	fmt.Println("Temporales:")
	memory.Temp.Print()
	fmt.Println("===================================")

	// Limpiar el ámbito local
	ctx.ClearLocalScope()

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
	leftNode, err := GetVarByAddress(left)
	if err != nil {
		return err
	}
	rightNode, err := GetVarByAddress(right)
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

func (n *FCallNode) Generate(ctx *Context) error {
	// Buscar la función en el directorio de funciones
	funcNode, found := funcDir[n.Id]
	if !found {
		return fmt.Errorf("función '%s' no declarada", n.Id)
	}

	// Verificar el número de parámetros
	if len(n.Params) != funcNode.ParamsCount {
		return fmt.Errorf("número de parámetros incorrecto para la función '%s': se esperaban %d, se recibieron %d", n.Id, funcNode.ParamsCount, len(n.Params))
	}

	// Agregar el cuádruplo de ERA (Reservar Espacio de Registro)
	ctx.AddQuad(ERA, funcNode.QuadStart, -1, -1)

	// Generar el código intermedio para los parámetros
	for i, param := range n.Params {
		if err := param.Generate(ctx); err != nil {
			return fmt.Errorf("error al generar parámetro en llamada a función '%s': %v", n.Id, err)
		}
		result := ctx.Pop()

		// Agregar el cuádruplo de asignación de parámetro
		ctx.AddQuad(PARAM, result, -1, i+1)
	}

	// Agregar el cuádruplo de llamada a función
	ctx.AddQuad(GOSUB, funcNode.QuadStart, -1, -1)

	return nil
}
