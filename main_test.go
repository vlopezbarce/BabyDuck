package main

import (
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
			var
                i, j: int;
                f: float;
			void uno(a: int, b: int) [
				{
					if (a > 0) {
						a = a + b * j + i;
						print(a + b);
					}
					else {
						print(i + j);
					};
				}
			];
            main {
				i = 2;
				j = 1;
				f = 3.14;

				while (i > 0) do {
					print(i, j * 2, f * 2 + 1.5);
					i = i - j;
				};
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

			if !pass {
				t.Fail()
			}
		})
	}
}
