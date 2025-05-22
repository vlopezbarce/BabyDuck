package ast

import "fmt"

var scope string
var global string
var globalVars map[string]bool

// Inicializa el ámbito global, la memoria y el asignador
func InitProgram(id string) {
	scope = id
	global = scope
	globalVars = make(map[string]bool)
	NewMemory()
	NewAllocator()
	FillOperatorsTree()
}

func ValidateVars(vars []*VarNode) error {
	localVars := make(map[string]bool)

	for _, v := range vars {
		// Verificar si la variable ya existe en el ámbito local
		if _, exists := localVars[v.Id]; exists {
			return fmt.Errorf("variable '%s' ya declarada en el ámbito local", v.Id)
		}

		// Verificar si la variable ya existe en el ámbito global
		if _, exists := globalVars[v.Id]; exists {
			return fmt.Errorf("variable '%s' ya declarada en el ámbito global", v.Id)
		}

		// Agregar la variable al mapa temporal para validación
		if scope == global {
			globalVars[v.Id] = true
		} else {
			localVars[v.Id] = true
		}
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

	// Si es el programa, ya se hizo la validación
	if id != global {
		// Establecer el ámbito actual a la función
		scope = id

		// Verificar si hay variables duplicadas
		if err := ValidateVars(vars); err != nil {
			return nil, err
		}

		// Reestablecer el ámbito global
		scope = global
	}

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
	//case *IfNode:
	//	return executeCondition(node)
	//case *WhileNode:
	//	return executeWhile(node)
	//case *FuncCallNode:
	//	return executeFunctionCall(node)
	default:
		return fmt.Errorf("tipo de instrucción no soportado: %T", node)
	}
}

// Función para ejecutar la asignación
func ExecuteAssign(assignNode *AssignNode) error {
	// Buscar variable destino y memoria correcta
	info, tree, err := lookupVar(assignNode.Id)
	if err != nil {
		return err
	}

	// Generar cuádruplos, pasando el VarNode destino
	ctx := &Context{}
	if err := assignNode.Generate(ctx, *info); err != nil {
		return err
	}

	// Ejecutar y evaluar
	PrintQuads(ctx.Quads)
	result := ctx.Evaluate()

	// Comprobar tipos
	if info.Type != result.Type {
		return fmt.Errorf(
			"tipo incompatible en '%s': se esperaba %s, se obtuvo %s",
			assignNode.Id, info.Type, result.Type,
		)
	}

	// Actualizar valor y escribir en la memoria correspondiente
	info.Value = result.Value
	tree.Update(info)

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

			// Ejecutar y evaluar
			PrintQuads(ctx.Quads)
			ctx.Evaluate()

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

// Imprime todas las variables
func PrintVariables() {
	fmt.Println()
	fmt.Println("Variables registradas:")
	fmt.Println("===================================")

	fmt.Println("Operadores:")
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

	fmt.Println()
	fmt.Println("Funciones registradas:")
	fmt.Println("===================================")
	for id := range funcDir {
		fmt.Printf("Función: %s\n", id)
	}
}
