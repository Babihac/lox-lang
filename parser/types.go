package parser

import (
	stm "lox/statement"
	"lox/tokens"
)

type FunctionComponents struct {
	parameters []tokens.Token
	body       []stm.Statement
}

func NewFunctionComponents(parameters []tokens.Token, body []stm.Statement) *FunctionComponents {
	return &FunctionComponents{
		parameters: parameters,
		body:       body,
	}
}
