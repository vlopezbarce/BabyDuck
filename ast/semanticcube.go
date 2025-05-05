package ast

import "fmt"

var semanticCube = map[string]map[string]map[string]string{
	"+": {
		"int": {
			"int":   "int",
			"float": "float",
		},
		"float": {
			"int":   "float",
			"float": "float",
		},
	},
	"-": {
		"int": {
			"int":   "int",
			"float": "float",
		},
		"float": {
			"int":   "float",
			"float": "float",
		},
	},
	">": {
		"int": {
			"int":   "bool",
			"float": "bool",
		},
		"float": {
			"int":   "bool",
			"float": "bool",
		},
	},
	"<": {
		"int": {
			"int":   "bool",
			"float": "bool",
		},
		"float": {
			"int":   "bool",
			"float": "bool",
		},
	},
	"!=": {
		"int": {
			"int":   "bool",
			"float": "bool",
		},
		"float": {
			"int":   "bool",
			"float": "bool",
		},
	},
}

func CheckSemantic(op string, left string, right string) (string, error) {
	if _, ok := semanticCube[op]; !ok {
		return "", fmt.Errorf("operador no soportado: %s", op)
	}
	if _, ok := semanticCube[op][left]; !ok {
		return "", fmt.Errorf("tipo izquierdo no soportado: %s", left)
	}
	if result, ok := semanticCube[op][left][right]; ok {
		return result, nil
	}
	return "", fmt.Errorf("operación inválida: %s %s %s", left, op, right)
}
