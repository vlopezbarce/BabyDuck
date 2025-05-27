package main

import (
	"BabyDuck/ast"
	"BabyDuck/lexer"
	"BabyDuck/parser"
	"testing"
)

// Estructura que define los casos de prueba
type TI struct {
	src    string
	expect bool
}

// Casos de prueba con código fuente y una expectativa de éxito o fracaso
var testData = []*TI{
	{
		src: `program patito;
			
			void fib(n: int) [{
				if (n < 2) {
					print(n);
				}
				else {
					fib(n - 1);
					fib(n - 2);
				};
			}];

			main {
				fib(5);
			}
			end`,
		expect: true,
	},
}

func TestParser(t *testing.T) {
	// Iterar sobre cada caso de prueba
	for _, ts := range testData {
		t.Run(ts.src, func(t *testing.T) {
			// Crear el lexer (analizador léxico) y el parser (analizador sintáctico)
			s := lexer.NewLexer([]byte(ts.src))
			p := parser.NewParser()

			pass := true

			var errors []error

			// Parsear el código fuente (obtener el código intermedio generado)
			ctx, err := p.Parse(s)
			errors = append(errors, err)

			// Ejecutar el programa con el código generado
			runtime := ast.NewRuntime(ctx.(*ast.Context))
			err = runtime.RunProgram()
			errors = append(errors, err)

			for _, err := range errors {
				// Verificar si el análisis fue exitoso o falló según la expectativa
				if err != nil {
					// Si no se esperaba un error y se obtuvo uno, el caso falla
					if ts.expect {
						pass = false
						t.Errorf("Error inesperado: %s", err.Error())
						continue
					} else {
						// Si se esperaba un error y se obtuvo uno, el caso pasa
						t.Log(err.Error())
					}
				} else {
					if ts.expect {
						// Si se esperaba éxito y no hubo error, el caso pasa
						t.Log("Análisis exitoso")
					} else {
						// Si se esperaba un error y no hubo ninguno, el caso falla
						pass = false
						t.Errorf("Se esperaba un error, pero no se produjo")
						continue
					}
				}
			}

			// Imprimir la salida del programa
			runtime.PrintOutput()

			if !pass {
				t.Fail()
			}
		})
	}
}
