package main

import (
	"BabyDuck_A00833578/lexer"
	"BabyDuck_A00833578/parser"
	"testing"
)

// Estructura que define los casos de prueba
type TI struct {
	src    string // Código fuente del programa en BabyDuck
	expect bool   // Espera si el análisis debe ser exitoso (true) o fallar (false)
}

// testData contiene casos de prueba para el analizador de BabyDuck
// Cada caso tiene un código fuente y una expectativa de éxito o fracaso
var testData = []*TI{
	// Test 0: Palabras reservadas en una función
	{
		src: `program reservedWords;
			void mainA() [{
				print("Hello World");
			}];
			main {
				mainA();
			}
			end
		`,
		expect: true,
	},

	// Test 1: Programa básico sin errores
	{
		src: `program sumTest;
			var
				x, y : int;
			void sumNumbers(x: int, y: int) [
				var
					sum : int;
				{
					sum = x + y;
					print(sum);
				}
			];
			main {
				sumNumbers();
			}
			end
		`,
		expect: true,
	},

	// Test 2: Función sin variables
	{
		src: `program simpleFunction;
			void myFunction() [{
				print("Simple Function");
			}];
			main {
				myFunction();
			}
			end
		`,
		expect: true,
	},

	// Test 3: If-Else correcto
	{
		src: `program ifElseCondition;
			var
				a : int;
			main {
				if (a > 5) {
					a = 10;
				} else {
					a = 0;
				};
			}
			end
		`,
		expect: true,
	},

	// Test 4: Ciclo While
	{
		src: `program whileLoop;
			var
				b : int;
			main {
				while (b < 10) do {
					b = b + 1;
				};
			}
			end
		`,
		expect: true,
	},

	// Test 5: Llamada a función sin parámetros
	{
		src: `program functionCall;
			void printValue() [
				var
					c : int;
				{
					c = 15;
					print(c);
				}
			];
			main {
				printValue();
			}
			end
		`,
		expect: true,
	},

	// Test 6: Llamada a función con parámetros
	{
		src: `program functionCallWithParams;
			var
				d : float;
			void addValues(a : int, b : float) [{
				d = a + b;
			}];
			main {
				addValues(5, 2.5);
			}
			end
		`,
		expect: true,
	},

	// Test 7: Uso de print con expresiones
	{
		src: `program printTest;
			var
				e : int;
				f : float;
			main {
				e = 10;
				f = 3.14;
				print("El valor de e es:", e);
				print("El valor de f es:", f);
			}
			end
		`,
		expect: true,
	},

	// Test 8: Operadores lógicos
	{
		src: `program logicalOperators;
			var
				g : int;
			main {
				if (g > 5) {
					if (g < 10) {
						g = 1;
					};
				};
			}
			end
		`,
		expect: true,
	},

	// Test 9: Operaciones aritméticas
	{
		src: `program arithmeticOperations;
			var
				h : int;
			main {
				h = 5 * (3 + 2);
			}
			end
		`,
		expect: true,
	},

	// Test 10: Función compleja
	{
		src: `program complexFunction;
			var
				j : int;
			void calculate(a : int, b : int) [{
				j = a + (b * 2);
			}];
			main {
				calculate(3, 4);
			}
			end
		`,
		expect: true,
	},

	// Test 11: Programa con falta de punto y coma después de "program"
	{
		src: `program missingSemicolon
			var
				y : float;
			main {
				y = 5.5;
			}
			end
		`,
		expect: false, // Falta punto y coma después de "program missingSemicolon"
	},

	// Test 12: Tipo de variable inválido
	{
		src: `program invalidType;
			var
				z : string;
			main {
				print(z);
			}
			end
		`,
		expect: false, // 'string' no es tipo válido en el lenguaje
	},

	// Test 13: Uso incorrecto de void
	{
		src: `program badVoid;
			var
				k : int;
			void mainFunc() [{
				void k = 5;
			}];
			main {
				mainFunc();
			}
			end
		`,
		expect: false, // Palabra 'void' usada incorrectamente adentro del cuerpo de función
	},

	// Test 14: Declaración incorrecta de variables
	{
		src: `program badVariables;
			var
				k : int,
				a : float;
			main {
				print(k);
			}
			end
		`,
		expect: false, // Variables de diferente tipo deben ser declaradas con punto y coma al final
	},

	// Test 15: Declaración incorrecta de función
	{
		src: `program badFunction;
			var
				x : int;
			void myPrint(x: int) {
				print("El valor de x es: ", x);
			}
			main {
				myPrint(x);
			}
			end
		`,
		expect: false, // Funciones esperan brackets y corchetes y punto y coma al final
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
