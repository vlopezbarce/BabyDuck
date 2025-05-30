# BabyDuck Compiler

**Instituto Tecnológico y de Estudios Superiores de Monterrey**

**Desarrollo de aplicaciones avanzadas de ciencias computacionales**

**Valeria López Barcelata A00833578**

**Profra. Elda Quiroga**

---

## Descripción

**BabyDuck** es un compilador educativo implementado en Go, desarrollado como parte de un proyecto académico. Usa [Gocc](https://github.com/goccmack/gocc) para realizar el análisis léxico y sintáctico. El compilador traduce programas escritos en un lenguaje estructurado tipo "Patito" y es capaz de evaluar el código intermedio generado para imprimir los resultados.

El proyecto incluye análisis léxico, sintáctico, semántico, generación de código intermedio (cuádruplos), y una máquina virtual para su ejecución. Implementa casos de prueba para validar el funcionamiento correcto del compilador.

---

## ➤ Entrega 1: Léxico y Sintaxis

- Implementación del analizador léxico y sintáctico usando Gocc.
- Definición de reglas gramaticales y expresiones regulares.
- Pruebas con programas simples que validan la estructura del lenguaje.

---

## ➤ Entrega 2: Semántica de Variables
  
- Directorio de funciones y tabla de variables por función.
- Validación de tipos, duplicados y referencias a variables.
- Implementación de un cubo semántico.

---

## ➤ Entrega 3: Expresiones y Estatutos Lineales

- Generación de cuádruplos para operaciones aritméticas y relacionales.
- Estructuras para pila semántica y cola de cuádruplos.
- Soporte para asignaciones y estatutos `print`.

---

### ➤ Entrega 4: Memoria y Control de Flujo
- Implementación del sistema de memoria con direcciones virtuales.
- Soporte para estatutos `if`, `else`, y `while`.
- Traductor a direcciones virtuales dinámicas para variables, constantes y temporales.

---

### ➤ Entrega 5: Funciones y Máquina Virtual
- Soporte para llamadas a funciones con parámetros.
- Implementación de una máquina virtual que ejecuta los cuádruplos.
- Gestión de contexto y memoria local por función.

---

### ➤ Entrega 6: Funciones que Retornan Valores
- Soporte completo para funciones que devuelven resultados.
- Evaluación de llamadas recursivas y expresiones anidadas.
- Ejecución de programas con lógica compleja y flujo de datos entre funciones.

---

## Estructura del Proyecto

<pre>
  📁 BabyDuck/
  ├── 📁 ast/ 
  │ ├── 📜 allocator.go      # Traducción a direcciones virtuales
  │ ├── 📜 ast.go            # Estructura del árbol sintáctico
  │ ├── 📜 memory.go         # Estructura de memoria
  │ ├── 📜 quads.go          # Generación de cuádruplos
  │ ├── 📜 runtime.go        # Ejecución del código intermedio
  │ ├── 📜 semanticcube.go   # Reglas de validación entre tipos
  │ ├── 📜 types.go          # Definición de nodos del AST
  ├── 📁 tests/              # Casos de prueba para el compilador
  ├── 📜 parser.bnf          # Definición léxica, gramatical y semántica del lenguaje
  └── 📜 compiler_test.go    # Programa principal de prueba
</pre>

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
