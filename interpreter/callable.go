package interpreter

type NativeFunctionCallable struct {
	arityFn func() int
	callFn  func(interpreter *Interpreter, args []any) any
}

func NewNativeFnCallable(arityFn func() int, callFn func(interpreter *Interpreter, args []any) any) *NativeFunctionCallable {
	return &NativeFunctionCallable{
		arityFn: arityFn,
		callFn:  callFn,
	}
}

func (f NativeFunctionCallable) Call(interpreter *Interpreter, args []any) any {
	return f.callFn(interpreter, args)
}

func (f NativeFunctionCallable) Arity() int {
	return f.arityFn()
}

type Callable interface {
	Call(interpreter *Interpreter, args []any) any
	Arity() int
}
