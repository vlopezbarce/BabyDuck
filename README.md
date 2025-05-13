# BabyDuck Compiler

**Instituto Tecnológico y de Estudios Superiores de Monterrey**

**Desarrollo de aplicaciones avanzadas de ciencias computacionales**

**Valeria López Barcelata A00833578**

**Profra. Elda Quiroga**

---

## Descripción

**BabyDuck** es un compilador en desarrollo como parte de un proyecto académico, implementado en el lenguaje Go. Utiliza [Gocc](https://github.com/goccmack/gocc), una herramienta generadora de analizadores léxicos y sintácticos para Go.
El compilador está diseñado para interpretar un lenguaje estructurado de propósito educativo, y actualmente cubre análisis léxico, sintáctico, semántico y generación de código intermedio (cuádruplos).

---

## 🔹 Entrega 1: Léxico y Sintaxis

- Se investigaron herramientas de generación de compiladores y se seleccionó **Gocc** por su integración con Go y buena documentación.
- Se definieron expresiones regulares y reglas gramaticales en formato `.bnf`.
- Se implementaron el **scanner** y el **parser** utilizando Gocc.
- Se diseñó un **plan de pruebas** para validar expresiones, declaraciones y estructuras básicas del lenguaje BabyDuck.

---

## 🔹 Entrega 2: Semántica de Variables
  
- Se diseñó un **Directorio de Funciones** (`map[string]FuncNode{}`) que almacena la información semántica de cada función declarada, incluyendo sus parámetros, cuerpo y variables locales.
- Cada función mantiene su propia **Tabla de Variables Locales**, implementada con un (`map[string]VarNode`).
- La primera función llamada (`program`) es tratada como la función principal y actúa como el contexto predeterminado del código.
- Se utilizó una variable de control (`currentScope`) que se actualiza durante el recorrido del AST para reflejar el contexto actual de ejecución.
- Se implementaron validaciones semánticas como:
  - Declaración de variables duplicadas.
  - Referencias a variables no definidas.
  - Verificación de tipos para asignaciones y expresiones.
- Se diseñó un **Cubo Semántico** para definir las reglas de compatibilidad entre tipos.

---

## 🔹 Entrega 3: Código de Expresiones y Estatutos Lineales

- Se implementó un sistema de generación de **cuádruplos** (`Quadruple`) para representar instrucciones intermedias.
- La generación de código (`Context`) utiliza:
  - **Pila semántica:** (`SemStack []VarNode`) para almacenar operandos durante el análisis.
  - **Fila de cuádruplos:** (`Quads []Quadruple`) que acumula el código intermedio generado.
  - **Contador de temporales:** (`TempCount int`) para generar identificadores únicos de variables temporales.
- Se imprimen los cuádruplos generados al final del análisis.
- Evaluación de cúadruplos:
  - Se soportan expresiones aritméticas y relacionales.
  - Se construye una memoria temporal (`temps map[string]VarNode`) para almacenar resultados parciales.
  - Las instrucciones **Assign** y **Print** evalúan expresiones y constantes.

---

## Estructura del Proyecto

📁 BabyDuck/

├── 📁 ast/

│    ├── 📜 ast.go                 # Métodos básicos para nodos del AST

│    ├── 📜 quads.go               # Generación de código intermedio (cuádruplos)

│    ├── 📜 semanticcube.go        # Implementación del cubo semántico

│    ├── 📜 types.go               # Definición de nodos del AST

├── 📜 parser.bnf                 # Definición léxica, gramatical y semántica del lenguaje

├── 📜 main_test.go               # Programa principal de prueba

└── 📜 README.md                  # Documentación general del proyecto

---

## 🛠 Requisitos

1. **Instalar Go:**  
   Descargar e instalar el lenguaje Go desde [https://golang.org](https://golang.org).

2. **Configurar la variable de entorno `GOPATH`:**  
   Configurar la variable `GOPATH` correctamente.  
   Consulta cómo hacerlo aquí: [Configuración de GOPATH](https://golang.org/doc/gopath_code.html)

## Instrucciones de Ejecución  

1️⃣ **Clonar este repositorio:**  
```
git clone https://github.com/vlopezbarce/BabyDuck.git
cd BabyDuck
```
2️⃣ **Instalar dependencias:**
```
go mod tidy
```
3️⃣ **Ejecutar las pruebas:**
```
go test -v
```
