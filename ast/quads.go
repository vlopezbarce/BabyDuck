package ast

import "fmt"

// Representa una instrucción de código intermedio (cuádruplo)
type Quadruple struct {
	Op   string // Operación
	Arg1 string // Operando 1
	Arg2 string // Operando 2
	Res  string // Resultado
}

// Almacena la pila semántica, cuádruplos y contador de temporales
type Context struct {
	SemStack  []string    // Pila de operandos
	Quads     []Quadruple // Fila de cuádruplos generados
	TempCount int         // Contador de variables temporales
}

// Agrega un operando a la pila semántica
func (ctx *Context) Push(operand string) {
	ctx.SemStack = append(ctx.SemStack, operand)
}

// Saca el operando superior de la pila semántica
func (ctx *Context) Pop() string {
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
func (ctx *Context) AddQuad(op, arg1, arg2, res string) {
	ctx.Quads = append(ctx.Quads, Quadruple{Op: op, Arg1: arg1, Arg2: arg2, Res: res})
}

// Imprime todos los cuádruplos con sus índices
func PrintQuads(quads []Quadruple) {
	for i, q := range quads {
		fmt.Printf("  %2d: (%s, %s, %s, %s)\n", i, q.Op, q.Arg1, q.Arg2, q.Res)
	}
}

// Genera el código intermedio para una asignación
func (n *AssignNode) Generate(ctx *Context) error {
	_, err := n.Exp.Generate(ctx)
	return err
}

// Genera el código intermedio para un valor literal
func (n *LiteralNode) Generate(ctx *Context) (string, error) {
	val := string(n.Tok.Lit)
	ctx.Push(val)
	return val, nil
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
	arg2 := ctx.Pop()
	arg1 := ctx.Pop()

	// Crea un temporal y agrega el cuádruplo
	temp := ctx.NewTemp()
	ctx.AddQuad(n.Op, arg1, arg2, temp)
	ctx.Push(temp) // Empuja el temporal a la pila

	return temp, nil
}
