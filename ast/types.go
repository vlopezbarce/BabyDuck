package ast

type ExpNode struct {
	Type  string
	Value Attrib
}

type AssignNode struct {
	Id  string
	Exp *ExpNode
}

type FuncNode struct {
	Id         string
	Parameters []*ParamNode
	Body       Attrib
}

type ParamNode struct {
	Id   Attrib
	Type Attrib
}
