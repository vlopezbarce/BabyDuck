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

	// Limpiar la memoria local
	if scope != global {
		memory.Local.Clear()
	}

	// Crear variables locales
	for _, v := range funcNode.Vars {
		if err := DeclareVariable(v.Id, v.Type); err != nil {
			return fmt.Errorf("error al declarar variable '%s' en función '%s': %v", v.Id, funcNode.Id, err)
		}
	}

	// Ejecutar las instrucciones del cuerpo
	for _, stmt := range funcNode.Body {
		if err := ExecuteStatement(stmt); err != nil {
			return fmt.Errorf("error al ejecutar en función '%s': %v", funcNode.Id, err)
		}
	}

	// Restablecer el ámbito global
	scope = global

	return nil
}

func ExecuteStatement(stmt Attrib) error {
	switch node := stmt.(type) {
	case *AssignNode:
		return ExecuteAssign(node)
	case []*PrintNode:
		return ExecutePrint(node)
	case *IfNode:
		return ExecuteCondition(node)
	//case *WhileNode:
	//	return executeWhile(node)
	//case *FuncCallNode:
	//	return executeFunctionCall(node)
	default:
		return fmt.Errorf("tipo de instrucción no soportado: %T", node)
	}
}

// Ejecuta la asignación de una variable
func ExecuteAssign(assignNode *AssignNode) error {
	ctx := &Context{}
	dest, err := assignNode.Generate(ctx)
	if err != nil {
		return err
	}

	// Imprimir cuádruplos y temporales
	ctx.PrintQuads()
	ctx.PrintTemps()

	// Ejecutar y evaluar
	result := ctx.Evaluate()
	ctx.PrintTemps()

	// Comprobar tipos
	if dest.Type != result.Type {
		return fmt.Errorf(
			"tipo incompatible en '%s': se esperaba %s, se obtuvo %s",
			assignNode.Id, dest.Type, result.Type,
		)
	}

	// Actualizar valor y escribir en la memoria correspondiente
	dest.Value = result.Value
	if scope != global {
		memory.Local.Update(dest)
	} else {
		memory.Global.Update(dest)
	}

	// Limpiar la memoria temporal
	memory.Temp.Clear()

	return nil
}

// Evalúa e imprime cada elemento de una lista
func ExecutePrint(printNodes []*PrintNode) error {
	// Generar todos los cuádruplos de print
	for _, n := range printNodes {
		switch n.Item.(type) {
		case Quad:
			// Generar cuádruplos para el nodo
			ctx := &Context{}
			if err := n.Generate(ctx); err != nil {
				return err
			}

			// Imprimir cuádruplos y temporales
			ctx.PrintQuads()
			ctx.PrintTemps()

			// Ejecutar y evaluar
			ctx.Evaluate()
			ctx.PrintTemps()

			// Limpiar la memoria temporal
			memory.Temp.Clear()

		case string:
			// Imprimir el string directamente
			fmt.Print(n.Item.(string))
		}

		// Espacio entre elementos
		fmt.Print(" ")
	}

	// Salto de línea final
	fmt.Println()

	// Limpiar temporales
	memory.Temp.Clear()

	return nil
}

// Ejecuta una condición if
func ExecuteCondition(ifNode *IfNode) error {
	// Generar código intermedio de la condición
	ctx := &Context{}
	condAddr, err := ifNode.Condition.Generate(ctx)
	if err != nil {
		return err
	}

	// Buscar el tipo del resultado de la condición
	result, err := lookupVarByAddress(condAddr)
	if err != nil {
		return err
	}
	if result.Type != "bool" {
		return fmt.Errorf("tipo incompatible en condición if: se esperaba bool, se obtuvo %s", result.Type)
	}

	// Crear etiquetas para el salto
	falseLabel := ctx.NewLabel()
	endLabel := ctx.NewLabel()

	// Obtener el operador de la memoria
	opGOTOF, _ := memory.Operators.FindByName("GOTOF")
	ctx.AddQuad(opGOTOF.Address, condAddr, -1, falseLabel)

	// Generar los cuádruplos para el bloque Then
	for _, stmt := range ifNode.ThenBlock {
		var err error
		switch stmt := stmt.(type) {
		case *AssignNode:
			_, err = stmt.Generate(ctx)
		case *PrintNode:
			err = stmt.Generate(ctx)
		}
		if err != nil {
			return fmt.Errorf("error al generar bloque Then: %v", err)
		}
	}

	// Obtener el operador de la memoria
	opGOTO, _ := memory.Operators.FindByName("GOTO")
	ctx.AddQuad(opGOTO.Address, -1, -1, endLabel)

	// Generar los cuádruplos para el bloque Else
	ctx.SetLabel(falseLabel)
	for _, stmt := range ifNode.ElseBlock {
		var err error
		switch stmt := stmt.(type) {
		case *AssignNode:
			_, err = stmt.Generate(ctx)
		case *PrintNode:
			err = stmt.Generate(ctx)
		}
		if err != nil {
			return fmt.Errorf("error al generar bloque Then: %v", err)
		}
	}

	ctx.SetLabel(endLabel)

	// Imprimir cuádruplos y temporales
	ctx.PrintQuads()
	ctx.PrintTemps()

	// Limpiar la memoria temporal
	memory.Temp.Clear()

	return nil
}

// Imprime todas las variables
func PrintVariables() {
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

	fmt.Println()
	fmt.Println("Funciones registradas:")
	fmt.Println("===================================")
	for id := range funcDir {
		fmt.Printf("Función: %s\n", id)
	}
}
