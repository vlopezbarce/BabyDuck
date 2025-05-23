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
		src: `program sumTest;
			var
                a, b, c, d, e, f, g, h, j, k, l : int;
                sum : int;
			void parserTest() [
				var
					x, y, z : int;
				{
					x = 1;
					y = 2;
					z = x + y;
					print("x + y = ", z);
				}
			];
			void memoryTest() [
				var
					x, y, z : int;
				{
					x = 1;
					y = 2;
					z = x + y;
					print("x + y = ", z);
				}
			];
            main {
				a = 5;
				b = 10;
                if (a < b) {
					c = a + b;
				} else {
					c = a - b;
				};

				a = 1;
				b = 2;
				c = 3;
				d = 4;
				e = 5;
				f = 6;
				g = 7;
				h = 8;
				j = 10;
				k = 11;
				l = 12;
				print(( ( a + b ) * c + d * e * f + k / h * j ) + g * l + h + j > ( a - c * d ) / f);
				print("hello world", a, 0, 1 < 2, 3.5);
            }
            end
		`,
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

			// Analizar los casos de prueba
			_, err := p.Parse(s)

			// Verificar si el análisis fue exitoso o falló según la expectativa
			if err != nil {
				// Si no se esperaba un error y se obtuvo uno, el caso falla
				if ts.expect {
					pass = false
					t.Errorf("Error inesperado: %s", err.Error())
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
				}
			}

			ast.PrintVariables()

			if !pass {
				t.Fail()
			}
		})
	}
}
