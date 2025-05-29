package ast

import "fmt"

var semanticCube = map[int]map[string]map[string]string{
	PLUS: {
		"int": {
			"int":   "int",
			"float": "float",
		},
		"float": {
			"int":   "float",
			"float": "float",
		},
	},
	MINUS: {
		"int": {
			"int":   "int",
			"float": "float",
		},
		"float": {
			"int":   "float",
			"float": "float",
		},
	},
	TIMES: {
		"int": {
			"int":   "int",
			"float": "float",
		},
		"float": {
			"int":   "float",
			"float": "float",
		},
	},
	DIVIDE: {
		"int": {
			"int":   "int",
			"float": "float",
		},
		"float": {
			"int":   "float",
			"float": "float",
		},
	},
	GT: {
		"int": {
			"int":   "bool",
			"float": "bool",
		},
		"float": {
			"int":   "bool",
			"float": "bool",
		},
	},
	LT: {
		"int": {
			"int":   "bool",
			"float": "bool",
		},
		"float": {
			"int":   "bool",
			"float": "bool",
		},
	},
	NEQ: {
		"int": {
			"int":   "bool",
			"float": "bool",
		},
		"float": {
			"int":   "bool",
			"float": "bool",
		},
		"bool": {
			"bool": "bool",
		},
	},
	ASSIGN: {
		"int": {
			"int": "int",
		},
		"float": {
			"float": "float",
		},
	},
	RETURN: {
		"int": {
			"int": "int",
		},
		"float": {
			"float": "float",
		},
	},
}

func CheckSemantic(op int, left string, right string) (string, error) {
	if _, ok := semanticCube[op][left]; !ok {
		return "", fmt.Errorf("tipo izquierdo no soportado: %s", left)
	}
	if result, ok := semanticCube[op][left][right]; ok {
		return result, nil
	}
	return "", fmt.Errorf("operación inválida entre %s y %s", left, right)
}
