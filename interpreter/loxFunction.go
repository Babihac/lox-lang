package interpreter

import (
	env "lox/environment"
	stm "lox/statement"
)

type LoxFunction struct {
	Declaration stm.FunctionStm
	Closure     *env.Environment
}

func NewLoxFunction(declaration stm.FunctionStm, closure *env.Environment) *LoxFunction {
	return &LoxFunction{
		Declaration: declaration,
		Closure:     closure,
	}
}

func (l LoxFunction) Call(interpreter *Interpreter, args []any) (result any) {

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

func (l LoxFunction) Arity() int {
	return len(l.Declaration.Params)
}

func (l LoxFunction) String() string {
	return "<fn " + l.Declaration.Name.Lexeme + ">"
}
