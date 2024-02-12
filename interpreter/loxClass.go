package interpreter

import (
	"fmt"
	"lox/tokens"
)

type LoxClass struct {
	Name          string
	Methods       map[string]*LoxFunction
	StaticMethods map[string]*LoxFunction
	Fields        map[string]any
}

func NewLoxClass(name string, methods map[string]*LoxFunction, staticMethods map[string]*LoxFunction) *LoxClass {
	return &LoxClass{
		Name:          name,
		Methods:       methods,
		StaticMethods: staticMethods,
		Fields:        make(map[string]any),
	}
}

func (l *LoxClass) Set(name tokens.Token, value any) {
	l.Fields[name.Lexeme] = value
}

func (l *LoxClass) Get(name tokens.Token, interpreter *Interpreter) (any, error) {
	field, ok := l.Fields[name.Lexeme]

	if ok {
		return field, nil
	}

	method, ok := l.StaticMethods[name.Lexeme]

	if ok {
		return method, nil
	}

	return nil, fmt.Errorf(fmt.Sprintf("Undefined property \"%s\".", name.Lexeme))

}

func (l *LoxClass) FindMethod(name string) (*LoxFunction, bool) {
	method, ok := l.Methods[name]

	if !ok {
		return nil, false
	}

	return method, true
}

func (l *LoxClass) Call(interpreter *Interpreter, args []any) any {
	instance := NewLoxInstance(l)
	initializer, ok := l.FindMethod("init")

	if ok {
		initializer.Bind(instance, interpreter)
		initializer.Call(interpreter, args)
	}

	return instance
}

func (l *LoxClass) Arity() int {
	initializer, ok := l.FindMethod("init")

	if ok {
		return initializer.Arity()
	}
	return 0
}

func (l *LoxClass) String() string {
	return l.Name
}
