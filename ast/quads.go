package ast

import (
	"fmt"
	"strconv"
)

// Representa una instrucción de código intermedio (cuádruplo)
type Quadruple struct {
	Operator string  // Operación
	Left     VarNode // Operando 1
	Right    VarNode // Operando 2
	Result   string  // Resultado
}

// Almacena la pila semántica, cuádruplos y contador de temporales
type Context struct {
	SemStack  []VarNode   // Pila de operandos
	Quads     []Quadruple // Fila de cuádruplos generados
	TempCount int         // Contador de variables temporales
}

// Agrega un operando a la pila semántica
func (ctx *Context) Push(varNode VarNode) {
	ctx.SemStack = append(ctx.SemStack, varNode)
}

// Saca el operando superior de la pila semántica
func (ctx *Context) Pop() VarNode {
	if len(ctx.SemStack) == 0 {
		panic("pop en pila vacía")
	}
	val := ctx.SemStack[len(ctx.SemStack)-1]
	ctx.SemStack = ctx.SemStack[:len(ctx.SemStack)-1]
	return val
}

// Genera un nombre de variable temporal nuevo
func (ctx *Context) NewTemp() string {
	ctx.TempCount++
	return fmt.Sprintf("t%d", ctx.TempCount)
}

// Agrega un nuevo cuádruplo a la lista
func (ctx *Context) AddQuad(operator, result string, left, right VarNode) {
	ctx.Quads = append(ctx.Quads, Quadruple{
		Operator: operator,
		Left:     left,
		Right:    right,
		Result:   result,
	})
}

// Imprime todos los cuádruplos con sus índices
func PrintQuads(quads []Quadruple) {
	fmt.Println()
	fmt.Println("Cuádruplos generados:")
	fmt.Println("===================================")

	for i, q := range quads {
		fmt.Printf("  %2d: (%s, %s, %s, %s)\n", i, q.Operator, q.Left.Value, q.Right.Value, q.Result)
	}
}

// Genera el código intermedio para una asignación
func (n *AssignNode) Generate(ctx *Context) error {
	_, err := n.Exp.Generate(ctx)
	return err
}

// Genera el código intermedio para una expresión binaria
func (n *ExpressionNode) Generate(ctx *Context) (string, error) {
	if _, err := n.Left.Generate(ctx); err != nil {
		return "", err
	}

	if _, err := n.Right.Generate(ctx); err != nil {
		return "", err
	}

	// Extrae los operandos de la pila
	right := ctx.Pop()
	left := ctx.Pop()

	// Verifica la compatibilidad de tipos
	resultType, err := CheckSemantic(n.Op, left.Type, right.Type)
	if err != nil {
		return "", err
	}

	// Crea un temporal y agrega el cuádruplo
	temp := ctx.NewTemp()
	ctx.AddQuad(n.Op, temp, left, right)

	// Agrega el temporal a la pila
	ctx.Push(VarNode{
		Id:    temp,
		Type:  resultType,
		Value: temp,
	})

	return temp, nil
}

// Genera el código intermedio para una variable o constante
func (n *VarNode) Generate(ctx *Context) (string, error) {
	var val, typ string

	if n.Id != "" {
		// Buscar el valor en el árbol de símbolos si es una variable
		var info *VarNode
		var found bool

		// Buscar la variable en el ámbito global
		if scope != global {
			info, found = memory.Local.FindByName(n.Id)
		}

		// Si no se encuentra en el ámbito local o el ámbito actual es global, buscar en el global
		if !found {
			info, found = memory.Global.FindByName(n.Id)
		}

		if !found {
			return "", fmt.Errorf("variable '%s' no declarada", n.Id)
		}

		val = info.Value
		typ = info.Type
	} else {
		// Si es una constante, usar su valor y tipo directamente
		val = n.Value
		typ = n.Type
	}

	ctx.Push(VarNode{
		Id:    n.Id,
		Value: val,
		Type:  typ,
	})
	return val, nil
}

// Ejecuta los cuádruplos generados y devuelve el resultado
func (ctx *Context) Evaluate() VarNode {
	fmt.Println()
	fmt.Println("Ejecución de cuádruplos")
	fmt.Println("===================================")

	// Memoria para resultados intermedios
	temps := make(map[string]VarNode)

	for _, q := range ctx.Quads {
		// 1) Recuperar operando izquierdo desde memoria si es temporal
		left := q.Left
		if info, found := temps[left.Id]; found {
			left = info
		}

		// 2) Recuperar operando derecho desde memoria si es temporal
		right := q.Right
		if info, found := temps[right.Id]; found {
			right = info
		}

		// 3) Convertir el valor textual a un dato Go, según su tipo
		var leftVal, rightVal Attrib
		switch left.Type {
		case "int":
			intVal, _ := strconv.Atoi(left.Value)
			leftVal = intVal
		case "float":
			floatVal, _ := strconv.ParseFloat(left.Value, 64)
			leftVal = floatVal
		case "bool":
			leftVal = left.Value == "1"
		}

		switch right.Type {
		case "int":
			intVal, _ := strconv.Atoi(right.Value)
			rightVal = intVal
		case "float":
			floatVal, _ := strconv.ParseFloat(right.Value, 64)
			rightVal = floatVal
		case "bool":
			rightVal = right.Value == "1"
		}

		// 4) Verificar el tipo de resultado con el cubo semántico
		resultType, err := CheckSemantic(q.Operator, left.Type, right.Type)
		if err != nil {
			panic(err)
		}

		// 5) Ejecutar la operación
		var rawResult Attrib
		switch q.Operator {
		case "+":
			rawResult = toFloat64(leftVal) + toFloat64(rightVal)
		case "-":
			rawResult = toFloat64(leftVal) - toFloat64(rightVal)
		case "*":
			rawResult = toFloat64(leftVal) * toFloat64(rightVal)
		case "/":
			rawResult = toFloat64(leftVal) / toFloat64(rightVal)
		case ">":
			rawResult = toFloat64(leftVal) > toFloat64(rightVal)
		case "<":
			rawResult = toFloat64(leftVal) < toFloat64(rightVal)
		case "!=":
			rawResult = toFloat64(leftVal) != toFloat64(rightVal)
		default:
			panic(fmt.Sprintf("Operación no soportada: %s", q.Operator))
		}

		// 6) Normalizar a string según el tipo de resultado
		var outValue string
		switch resultType {
		case "int":
			outValue = fmt.Sprintf("%d", int(toFloat64(rawResult)))
		case "float":
			outValue = fmt.Sprintf("%f", toFloat64(rawResult))
		case "bool":
			if rawResult.(bool) {
				outValue = "1"
			} else {
				outValue = "0"
			}
		}

		// 7) Debug
		fmt.Printf("%s %s %s -> %s (%s)\n",
			q.Operator,
			left.Value,
			right.Value,
			outValue,
			resultType,
		)

		// 8) Almacenar resultado en memoria
		temps[q.Result] = VarNode{
			Id:    q.Result,
			Type:  resultType,
			Value: outValue,
		}
	}

	fmt.Println()
	fmt.Println("Memoria de temporales:")
	fmt.Println("===================================")
	fmt.Println(temps)

	// 9) Devolver el resultado final
	return temps[ctx.Quads[len(ctx.Quads)-1].Result]
}

// Convierte int, float64 o bool a float64
func toFloat64(v Attrib) float64 {
	switch x := v.(type) {
	case int:
		return float64(x)
	case float64:
		return x
	case bool:
		if x {
			return 1
		}
		return 0
	default:
		panic("tipo no soportado en toFloat64")
	}
}
