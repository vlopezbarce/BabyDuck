package main

import (
	"BabyDuck/ast"
	"BabyDuck/lexer"
	"BabyDuck/parser"
	"os"
	"testing"
)

// Define los casos de prueba
type TestCase struct {
	Name   string
	Source string
	Expect bool
}

// Devuelve el contenido de un archivo como texto
func ReadTestCase(filename string) string {
	lines, err := os.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	return string(lines)
}

// Leer los casos de prueba desde los directorios tests/pass y tests/fail
func NewTestCases() []TestCase {
	var testCases []TestCase

	passCases, _ := os.ReadDir("tests/pass")
	for _, file := range passCases {
		source := ReadTestCase("tests/pass/" + file.Name())
		testCases = append(testCases, TestCase{Name: file.Name(), Source: source, Expect: true})
	}

	failCases, _ := os.ReadDir("tests/fail")
	for _, file := range failCases {
		source := ReadTestCase("tests/fail/" + file.Name())
		testCases = append(testCases, TestCase{Name: file.Name(), Source: source, Expect: false})
	}

	return testCases
}

func VerifyOutcome(t *testing.T, err error, expect bool) {
	// Si no se esperaba un error y se obtuvo uno, el caso falla
	if err != nil && expect {
		t.Error(err)
		t.FailNow()
	}
	// Si se esperaba un error y no hubo ninguno, el caso falla
	if err == nil && !expect {
		t.Errorf("Se esperaba un error, pero no se produjo")
		t.FailNow()
	}
}

func TestCompiler(t *testing.T) {
	testCases := NewTestCases()

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			// Analizar el léxico del código fuente
			s := lexer.NewLexer([]byte(tc.Source))
			p := parser.NewParser()

			// Analizar la sintaxis del código fuente
			program, err := p.Parse(s)

			// Verificar el resultado del análisis
			VerifyOutcome(t, err, tc.Expect)

			if program == nil {
				if tc.Expect {
					t.Error(err)
					t.FailNow()
				}
				return
			}

			// Si el análisis fue exitoso, generar el código intermedio
			ct := &ast.Compilation{}
			err = program.(*ast.ProgramNode).Generate(ct)

			// Verificar si hubo errores al generar el código intermedio
			VerifyOutcome(t, err, tc.Expect)

			// Ejecutar el programa con el código generado
			rt := ast.NewRuntime(ct)
			err = rt.RunProgram()

			// Verificar si hubo errores al ejecutar el programa
			VerifyOutcome(t, err, tc.Expect)

			// Imprimir la salida del programa
			rt.PrintOutput()
		})
	}
}
