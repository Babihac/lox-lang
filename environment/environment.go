package env

import (
	"fmt"
	"lox/tokens"
)

type Environment struct {
	Values    map[string]any
	Enclosing *Environment
}

func NewEnvironment(enclosing ...*Environment) *Environment {
	values := make(map[string]any)
	var env *Environment = nil

	if len(enclosing) > 0 {
		env = enclosing[0]
	}

	return &Environment{
		Values:    values,
		Enclosing: env,
	}
}

func (e *Environment) Define(name string, value any) {
	e.Values[name] = value
}

func (e *Environment) Get(name tokens.Token) any {
	value, ok := e.Values[name.Lexeme]
	if ok {
		return value
	}

	if e.Enclosing != nil {
		return e.Enclosing.Get(name)
	}

	msg := fmt.Sprintf("Undefined variable: %s.\n", name.Lexeme)
	panic(msg)
}

func (e *Environment) GetAt(depth int, name string) any {
	value, ok := e.ancestor(depth).Values[name]

	if ok {
		return value
	}

	msg := fmt.Sprintf("Undefined variable: %s.\n", name)
	panic(msg)
}

func (e *Environment) ancestor(depth int) *Environment {
	bindedEnv := e

	for i := 0; i < depth; i++ {
		bindedEnv = bindedEnv.Enclosing
	}
	return bindedEnv
}

func (e *Environment) Assign(name tokens.Token, value any) {
	if _, ok := e.Values[name.Lexeme]; ok {
		e.Values[name.Lexeme] = value
		return
	}

	if e.Enclosing != nil {
		e.Enclosing.Assign(name, value)
		return
	}

	panic(fmt.Sprintf("Undefined variable %s.", name.Lexeme))
}

func (e *Environment) AssignAt(depth int, name tokens.Token, value any) {
	e.ancestor(depth).Values[name.Lexeme] = value
}
