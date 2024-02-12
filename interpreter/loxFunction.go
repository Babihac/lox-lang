package interpreter

import (
	env "lox/environment"
	stm "lox/statement"
)

type LoxFunction struct {
	Declaration   stm.FunctionStm
	Closure       *env.Environment
	isInitializer bool
}

func NewLoxFunction(declaration stm.FunctionStm, closure *env.Environment, isInitialzier bool) *LoxFunction {
	return &LoxFunction{
		Declaration:   declaration,
		Closure:       closure,
		isInitializer: isInitialzier,
	}
}

func (l *LoxFunction) Bind(instance *LoxInstance, interpreter *Interpreter) {
	thisIndex := l.Declaration.ThisIndex

	if thisIndex != -1 {
		interpreter.LocalVariables[thisIndex] = instance
	}
}

func (l *LoxFunction) Call(interpreter *Interpreter, args []any) (result any) {

	defer func() {
		value := recover()
		thisIndex := l.Declaration.ThisIndex

		if l.isInitializer {
			result = interpreter.LocalVariables[thisIndex]
		}

		if value != nil {

			returnValue, ok := value.(ReturnValue)

			if !ok {
				panic(value)
			}

			if !l.isInitializer {
				result = returnValue.Value
			}
		}

	}()

	environment := env.NewEnvironment(l.Closure)

	for i, param := range l.Declaration.Params {
		index := interpreter.locals[param]
		interpreter.LocalVariables[index] = args[i]
	}

	interpreter.executeBlock(l.Declaration.Body, environment)

	return
}

func (l *LoxFunction) Arity() int {
	return len(l.Declaration.Params)
}

func (l *LoxFunction) String() string {
	return "<fn " + l.Declaration.Name.Lexeme + ">"
}
