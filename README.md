# BabyDuck Compiler

**Tecnológico de Monterrey**
**TC3002B: Desarrollo de aplicaciones avanzadas de ciencias computacionales**
**Módulo 3: Compiladores**
**Profra. Elda Quiroga**
**Valeria López Barcelata A00833578**

---

## Descripción general

**BabyDuck** es un compilador en desarrollo como parte de un proyecto académico, implementado en el lenguaje Go. El compilador está diseñado para interpretar un lenguaje estructurado de propósito educativo, y actualmente cubre análisis léxico, sintáctico, semántico y generación de código intermedio (cuádruplos).

---

## Entrega 1: Léxico y Sintaxis

- Se investigaron herramientas de generación de compiladores y se seleccionó **Gocc** por su integración con Go y buena documentación.
- Se definieron expresiones regulares y reglas gramaticales en formato `.bnf`.
- Se implementaron el **scanner** y el **parser** utilizando Gocc.
- Se diseñó un **plan de pruebas** para validar expresiones, declaraciones y estructuras básicas del lenguaje BabyDuck.
- La documentación incluye un resumen de herramientas exploradas, y cómo se definieron las reglas gramaticales.

---

## Entrega 2: Semántica de Variables

- Se diseñó un **Directorio de Funciones** y una **Tabla de Variables** para manejar información semántica.
- Se implementaron validaciones como **variables duplicadas** y manejo de **tipos**.
- Se utilizó un mapa (`map[string]VarNode`) para la tabla de símbolos por función.
- Se diseñó un **cubo semántico** para definir reglas de compatibilidad entre tipos.
- Todo el análisis semántico ocurre durante el recorrido del AST tras el parseo.

---

## Entrega 3: Código de Expresiones y Estatutos Lineales

- Se implementó un sistema de generación de **cuádruplos** para representar instrucciones intermedias.
- Se utilizaron estructuras como:
  - **Pila semántica**: para operandos.
  - **Contador de temporales**: para generar nombres de variables intermedias.
  - **Fila de cuádruplos**: que representa el código generado.
- Se soportan expresiones aritméticas y relacionales.

---

## Estructura del Proyecto

BabyDuck_A00833578/
├── ast/
│   ├── ast.go                 # Métodos del AST
│   ├── semanticcube.go        # Implementación del cubo semántico
│   ├── symboltable.go         # Funciones para la tabla de variables
│   ├── types.go               # Definición de nodos del AST
├── parser.bnf                 # Definición léxica, gramatical y semántica del lenguaje
├── main_test.go               # Programa principal de prueba
└── README.md                  # Documentación general del proyecto

---

## Herramientas utilizadas

- **Go (Golang)** como lenguaje principal de implementación.
- **Gocc** como generador de parser y lexer.
- Terminal e IDE: Visual Studio Code.