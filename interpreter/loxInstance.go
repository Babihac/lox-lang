package interpreter

import (
	"fmt"
	"lox/tokens"
)

type IloxInstance interface {
	Get(name tokens.Token, interpreter *Interpreter) (any, error)
	Set(name tokens.Token, value any)
}

type LoxInstance struct {
	class  *LoxClass
	fields map[string]any
}

func NewLoxInstance(class *LoxClass) *LoxInstance {
	return &LoxInstance{
		class:  class,
		fields: make(map[string]any),
	}
}

func (l *LoxInstance) Get(name tokens.Token, interpreter *Interpreter) (any, error) {
	field, ok := l.fields[name.Lexeme]

	if ok {
		return field, nil
	}

	method, ok := l.class.FindMethod(name.Lexeme)

	if ok {
		method.Bind(l, interpreter)
		return method, nil
	}

	return nil, fmt.Errorf(fmt.Sprintf("Undefined property \"%s\".", name.Lexeme))

}

func (l *LoxInstance) Set(name tokens.Token, value any) {
	l.fields[name.Lexeme] = value
}

func (l *LoxInstance) String() string {
	return fmt.Sprintf("%s instance", l.class.Name)
}
