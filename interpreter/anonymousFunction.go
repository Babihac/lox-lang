package interpreter

import (
	env "lox/environment"
	stm "lox/statement"
)

type AnonymousFunction struct {
	Declaration stm.AnonymousFunction
	Closure     *env.Environment
}

func NewAnonymousFunction(declaration stm.AnonymousFunction, closure *env.Environment) *AnonymousFunction {
	return &AnonymousFunction{
		Declaration: declaration,
		Closure:     closure,
	}
}

func (l AnonymousFunction) Call(interpreter *Interpreter, args []any) (result any) {

	defer func() {
		result = recover()
	}()

	environment := env.NewEnvironment(l.Closure)

	for i, param := range l.Declaration.Params {
		index := interpreter.locals[param]
		interpreter.LocalVariables[index] = args[i]
	}

	interpreter.executeBlock(l.Declaration.Body, environment)

	return
}

func (l AnonymousFunction) Arity() int {
	return len(l.Declaration.Params)
}

func (l AnonymousFunction) String() string {
	return "< anonymous function >"
}
