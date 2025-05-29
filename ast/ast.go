package ast

import "fmt"

var scope string
var global string

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

func DeclareFunction(id string, params, vars []*VarNode, body []Attrib) (*FuncNode, error) {
	// Verificar si la función ya existe
	if _, exists := funcDir[id]; exists {
		return nil, fmt.Errorf("función '%s' ya declarada", id)
	}

	// Crear el nodo de función
	funcNode := &FuncNode{
		Id:     id,
		Params: params,
		Vars:   vars,
		Body:   body,
	}

	// Agregar la función al directorio
	funcDir[id] = funcNode

	// Verificar si hay variables duplicadas
	if err := ValidateVars(append(params, vars...)); err != nil {
		return nil, err
	}

	return funcNode, nil
}

func DeclareVariable(varNode *VarNode) error {
	// Obtener la dirección de memoria para la variable
	var addr int
	var err error

	if varNode.Id == "" {
		addr, err = alloc.NextConst(varNode.Type)
	} else if scope == global {
		addr, err = alloc.NextGlobal(varNode.Type)
	} else {
		addr, err = alloc.NextLocal(varNode.Type)
	}
	if err != nil {
		return err
	}

	// Actualizar el nodo de variable con la dirección
	varNode.Address = addr

	// Insertar la variable en la memoria correspondiente
	if varNode.Id == "" {
		memory.Const.Insert(varNode)
	} else if scope == global {
		memory.Global.Insert(varNode)
	} else {
		memory.Local.Insert(varNode)
	}

	return nil
}

func (n ProgramNode) Generate(ct *Compilation) error {
	// Inicializar la memoria y el asignador de direcciones
	NewMemory()
	NewAllocator()

	// Establecer el ámbito global
	global = n.Id
	scope = global

	// Registrar el programa en el directorio de funciones
	funcDir[n.Id] = &FuncNode{
		Id: n.Id,
	}

	// Verificar si hay variables duplicadas
	if err := ValidateVars(n.Vars); err != nil {
		return err
	}

	// Agregar el cuádruplo de inicio del programa
	ct.AddQuad(GOTO, -1, -1, -1)

	// Crear variables dentro del ámbito global
	for _, v := range n.Vars {
		if err := DeclareVariable(v); err != nil {
			return fmt.Errorf("error de compilación en '%s': %v", v.Id, err)
		}
	}

	// Generar cuádruplos para las funciones
	for _, funcNode := range n.Funcs {
		if err := funcNode.Generate(ct); err != nil {
			return fmt.Errorf("error de compilación en '%s': %v", funcNode.Id, err)
		}
	}

	// Marcar el inicio del programa
	ct.Quads[0].Result = len(ct.Quads)

	// Generar cuádruplos para el cuerpo del programa
	for _, stmt := range n.Body {
		if err := stmt.Generate(ct); err != nil {
			return fmt.Errorf("error de compilación en '%s': %v", n.Id, err)
		}
	}

	// Imprimir variables globales y constantes
	fmt.Println()
	fmt.Printf("Programa: %s\n", n.Id)
	fmt.Println("===================================")

	if memory.Global.Size() > 0 {
		fmt.Println()
		fmt.Println("Globales:")
		fmt.Println("===================================")
		memory.Global.Print()
	}

	if memory.Const.Size() > 0 {
		fmt.Println()
		fmt.Println("Constantes:")
		fmt.Println("===================================")
		memory.Const.Print()
	}

	if memory.Temp.Size() > 0 {
		fmt.Println()
		fmt.Println("Temporales:")
		fmt.Println("===================================")
		memory.Temp.Print()
	}

	// Imprimir cuádruplos generados
	ct.PrintQuads()

	return nil
}

func (n *FuncNode) Generate(ct *Compilation) error {
	// Marcar el inicio del cuádruplo de la función
	funcDir[n.Id].QuadStart = len(ct.Quads)

	// Establecer el ámbito actual a la función
	scope = n.Id

	// Crear parámetros dentro del ámbito de la función
	var paramNodes []*VarNode
	for _, p := range n.Params {
		if err := DeclareVariable(p); err != nil {
			return fmt.Errorf("error al declarar parámetro '%s' en función '%s': %v", p.Id, n.Id, err)
		}
		paramNodes = append(paramNodes, p)
	}

	// Crear variables dentro del ámbito de la función
	var varNodes []*VarNode
	for _, v := range n.Vars {
		if err := DeclareVariable(v); err != nil {
			return fmt.Errorf("error al declarar variable '%s' en función '%s': %v", v.Id, n.Id, err)
		}
		varNodes = append(varNodes, v)
	}

	// Generar cuádruplos para el cuerpo de la función
	for _, stmt := range n.Body {
		if err := stmt.Generate(ct); err != nil {
			return fmt.Errorf("error al generar cuádruplos para '%s': %v", n.Id, err)
		}
	}

	// Agregar el cuádruplo de retorno al final de la función
	ct.AddQuad(ENDFUNC, -1, -1, -1)

	// Guardar variables generadas en la función
	funcDir[n.Id].Params = paramNodes
	funcDir[n.Id].Vars = varNodes
	funcDir[n.Id].Temps = memory.Temp.GetAll()

	// Imprimir variables locales y temporales de la función
	fmt.Println()
	fmt.Printf("Función: %s\n", n.Id)
	fmt.Println("===================================")

	if memory.Local.Size() > 0 {
		fmt.Println()
		fmt.Println("Locales:")
		fmt.Println("===================================")
		memory.Local.Print()
	}

	if memory.Temp.Size() > 0 {
		fmt.Println()
		fmt.Println("Temporales:")
		fmt.Println("===================================")
		memory.Temp.Print()
	}

	// Limpiar el ámbito local
	ct.ClearLocalScope()

	return nil
}

func (n *VarNode) Generate(ct *Compilation) error {
	// Buscar la constante en la memoria
	varNode, found := memory.Const.FindConst(n.Type, n.Value)

	if !found {
		// Declarar la constante si no existe
		if err := DeclareVariable(n); err != nil {
			return err
		}
		// Agregar la dirección a la pila
		ct.Push(n.Address)
		return nil
	}
	// Agregar la dirección a la pila
	ct.Push(varNode.Address)

	return nil
}

func (n AssignNode) Generate(ct *Compilation) error {
	// Buscar variable destino y memoria correcta
	var destNode *VarNode
	var found bool

	if scope != global {
		destNode, found = memory.Local.FindByName(n.Id)
	}
	if !found {
		destNode, found = memory.Global.FindByName(n.Id)
	}
	if !found {
		return fmt.Errorf("variable '%s' no declarada", n.Id)
	}

	// Generar el código intermedio para la expresión
	if err := n.Exp.Generate(ct); err != nil {
		return err
	}
	result := ct.Pop()

	// Obtener el nodo de resultado desde memoria
	resultNode, err := GetByAddress(result, nil)
	if err != nil {
		return err
	}

	// Verificar que el tipo del resultado sea compatible con el tipo de la variable destino
	_, err = CheckSemantic(ASSIGN, resultNode.Type, destNode.Type)
	if err != nil {
		return err
	}

	// Agregar el cuádruplo de asignación
	ct.AddQuad(ASSIGN, result, -1, destNode.Address)

	return nil
}

func (n PrintNode) Generate(ct *Compilation) error {
	// Generar el código intermedio para los elementos a imprimir
	for _, item := range n.Items {
		if err := item.Generate(ct); err != nil {
			return err
		}
		result := ct.Pop()

		// Agregar el cuádruplo de impresión
		ct.AddQuad(PRINT, result, -1, -1)
	}

	// Agregar el cuádruplo de nueva línea
	ct.AddQuad(PRINTLN, -1, -1, -1)

	return nil
}

func (n ExpressionNode) Generate(ct *Compilation) error {
	// Generar el código intermedio para los operandos izquierdo y derecho
	if err := n.Left.Generate(ct); err != nil {
		return err
	}
	if err := n.Right.Generate(ct); err != nil {
		return err
	}

	// Obtener los operandos izquierdo y derecho
	right := ct.Pop()
	left := ct.Pop()

	// Obtener los nodos de memoria correspondientes
	leftNode, err := GetByAddress(left, nil)
	if err != nil {
		return err
	}
	rightNode, err := GetByAddress(right, nil)
	if err != nil {
		return err
	}

	if n.Op == DIVIDE && rightNode.Value == "0" {
		return fmt.Errorf("división por cero en la expresión")
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
	tempId := ct.NewTemp()

	tempNode := &VarNode{
		Address: addr,
		Id:      tempId,
		Type:    resultType,
		Value:   tempId,
	}

	// Insertar el temporal en la memoria
	memory.Temp.Insert(tempNode)

	// Agregar el cuádruplo de la operación
	ct.AddQuad(n.Op, left, right, addr)

	// Agregar el temporal a la pila
	ct.Push(addr)

	return nil
}

func (n ExpressionVar) Generate(ct *Compilation) error {
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
	ct.Push(varNode.Address)

	return nil
}

func (n IfNode) Generate(ct *Compilation) error {
	// Generar el código intermedio para la condición
	if err := n.Condition.Generate(ct); err != nil {
		return err
	}
	result := ct.Pop()

	// Buscar el tipo del resultado de la condición
	resultNode, _ := GetByAddress(result, nil)
	if resultNode.Type != "bool" {
		return fmt.Errorf("tipo incompatible en condición if: se esperaba bool, se obtuvo %s", resultNode.Type)
	}

	// Agregar el cuádruplo GOTOF
	indexGOTOF := len(ct.Quads)
	ct.AddQuad(GOTOF, result, -1, -1)

	// Generar los cuádruplos para el bloque Then
	for _, stmt := range n.ThenBlock {
		if err := stmt.Generate(ct); err != nil {
			return err
		}
	}

	// Agregar el cuádruplo GOTO
	indexGOTO := len(ct.Quads)
	ct.AddQuad(GOTO, -1, -1, -1)

	// Marcar la etiqueta para el cuádruplo GOTOF
	ct.Quads[indexGOTOF].Result = len(ct.Quads)

	// Generar los cuádruplos para el bloque Else
	for _, stmt := range n.ElseBlock {
		if err := stmt.Generate(ct); err != nil {
			return err
		}
	}

	// Marcar la etiqueta para el cuádruplo GOTO
	ct.Quads[indexGOTO].Result = len(ct.Quads)

	return nil
}

func (n WhileNode) Generate(ct *Compilation) error {
	// Marcar el inicio del ciclo
	start := len(ct.Quads)

	// Generar el código intermedio para el ciclo
	if err := n.Condition.Generate(ct); err != nil {
		return err
	}
	result := ct.Pop()

	// Buscar el tipo del resultado de la condición
	resultNode, _ := GetByAddress(result, nil)
	if resultNode.Type != "bool" {
		return fmt.Errorf("tipo incompatible en condición if: se esperaba bool, se obtuvo %s", resultNode.Type)
	}

	// Agregar el cuádruplo GOTOF
	indexGOTOF := len(ct.Quads)
	ct.AddQuad(GOTOF, result, -1, -1)

	// Generar los cuádruplos para el cuerpo del ciclo
	for _, stmt := range n.Body {
		if err := stmt.Generate(ct); err != nil {
			return err
		}
	}

	// Agregar el cuádruplo GOTO
	ct.AddQuad(GOTO, -1, -1, start)

	// Marcar la etiqueta para el cuádruplo GOTOF
	ct.Quads[indexGOTOF].Result = len(ct.Quads)

	return nil
}

func (n FCallNode) Generate(ct *Compilation) error {
	// Buscar la función en el directorio de funciones
	funcNode, found := funcDir[n.Id]
	if !found {
		return fmt.Errorf("función '%s' no declarada", n.Id)
	}
	if funcNode.Id == global {
		return fmt.Errorf("no se puede llamar a la función '%s'", n.Id)
	}

	// Verificar el número de parámetros
	if len(n.Params) != len(funcNode.Params) {
		return fmt.Errorf("número de parámetros incorrecto para la función '%s': se esperaban %d, se recibieron %d", n.Id, len(funcNode.Params), len(n.Params))
	}

	// Agregar el cuádruplo de ERA (Reservar Espacio de Registro)
	ct.AddQuad(ERA, funcNode.QuadStart, -1, -1)

	// Generar el código intermedio para los parámetros
	for i, param := range n.Params {
		if err := param.Generate(ct); err != nil {
			return fmt.Errorf("error al generar parámetro en llamada a función '%s': %v", n.Id, err)
		}
		result := ct.Pop()

		// Verificar el tipo del parámetro
		resultNode, _ := GetByAddress(result, nil)
		if resultNode.Type != funcNode.Params[i].Type {
			return fmt.Errorf("tipo de parámetro incorrecto en la función '%s': se esperaba %s, se recibió %s", n.Id, funcNode.Params[i].Type, resultNode.Type)
		}

		// Agregar el cuádruplo de asignación de parámetro
		ct.AddQuad(PARAM, result, -1, i+1)
	}

	// Agregar el cuádruplo de llamada a función
	ct.AddQuad(GOSUB, funcNode.QuadStart, -1, -1)

	return nil
}
