package interpreter

import (
	"fmt"
	env "lox/environment"
	"lox/interfaces"
	stm "lox/statement"
	"lox/tokens"
	"reflect"
	"strings"
	"time"
)

type ReturnValue struct {
	Value any
}

type Interpreter struct {
	errorLogger          interfaces.ErrorLogger
	environment          *env.Environment
	nearestEnclosingLoop []*stm.WhileStmt
	breaking             bool
	globals              *env.Environment
	locals               map[tokens.Token]int
	LocalVariables       []any
}

func NewInterpreter(errorLogger interfaces.ErrorLogger) *Interpreter {
	globals := env.NewEnvironment()

	var clockCallable Callable = NewNativeFnCallable(
		func() int { return 0 },
		func(interpreter *Interpreter, args []any) any {
			return time.Now().UnixNano() / int64(time.Millisecond)
		})

	globals.Define("clock", clockCallable)

	return &Interpreter{
		errorLogger:    errorLogger,
		environment:    globals,
		globals:        globals,
		locals:         make(map[tokens.Token]int),
		LocalVariables: make([]any, 0),
	}
}

func (i *Interpreter) Interpret(statements []stm.Statement) {
	defer i.afterPanic()

	for _, stmt := range statements {
		i.execute(stmt)
	}
}

func (i *Interpreter) InterpretRepl(statements []stm.Statement) {
	defer i.afterPanic()

	for _, stmt := range statements {
		if exprStmt, ok := stmt.(*stm.ExpressionStmt); ok {
			value := i.evaluate(exprStmt.Expression)
			fmt.Println(value)
		} else {
			i.execute(stmt)
		}
	}

}

func (i *Interpreter) execute(stmt stm.Statement) {
	if i.breaking {
		return
	}
	stmt.Accept(i)
}

func (i *Interpreter) Resolve(token tokens.Token, depth int, index int) {
	i.locals[token] = index
}

func (i *Interpreter) executeBlock(statements []stm.Statement, environment *env.Environment) {
	previousEnv := i.environment

	defer func() {
		i.environment = previousEnv
	}()

	i.environment = environment

	for _, stmt := range statements {
		i.execute(stmt)
	}

}

func (i *Interpreter) evaluate(expr stm.Expression) any {
	return expr.Accept(i)
}

func (i *Interpreter) VisitFunctionStatement(stmt *stm.FunctionStm) any {
	index, ok := i.locals[stmt.Name]
	function := NewLoxFunction(*stmt, i.environment, false)

	if ok {
		i.LocalVariables[index] = function
	} else {
		i.globals.Define(stmt.Name.Lexeme, function)
	}

	return nil
}

func (i *Interpreter) VisitReturnStatement(stmt *stm.ReturnStmt) any {
	var value any = nil

	if stmt.Value != nil {
		value = i.evaluate(stmt.Value)
	}

	panic(ReturnValue{Value: value})
}

func (i *Interpreter) VisitBlockStatement(stmt *stm.BlockStmt) any {
	i.executeBlock(stmt.Statements, env.NewEnvironment(i.environment))
	return nil
}

func (i *Interpreter) VisitClassStatement(stmt *stm.ClassStmt) any {

	var superClass *LoxClass = nil

	if stmt.SuperClass != nil {
		value := i.evaluate(stmt.SuperClass)

		if _, ok := value.(*LoxClass); !ok {
			panic("Superclass must be a class.")
		}

		superClass = value.(*LoxClass)

		i.LocalVariables[stmt.SuperIndex] = superClass
	}

	index, ok := i.locals[stmt.Name]

	if !ok {
		i.globals.Define(stmt.Name.Lexeme, nil)
	}

	methods := make(map[string]*LoxFunction)
	staticMethods := make(map[string]*LoxFunction)

	for _, method := range stmt.Methods {
		function := NewLoxFunction(*method, i.globals, method.Name.Lexeme == "init")
		methods[method.Name.Lexeme] = function
	}

	for _, method := range stmt.StaticMethods {
		function := NewLoxFunction(*method, i.globals, method.Name.Lexeme == "init")
		staticMethods[method.Name.Lexeme] = function
	}

	class := NewLoxClass(stmt.Name.Lexeme, methods, staticMethods, superClass)

	if ok {
		i.LocalVariables[index] = class
	} else {
		i.globals.Assign(stmt.Name, class)
	}

	return nil

}

// VisitExprStatement implements stm.Visitor.
func (i *Interpreter) VisitExprStatement(stmt *stm.ExpressionStmt) any {
	i.evaluate(stmt.Expression)
	return nil
}

// VisitPrintStatement implements stm.Visitor.
func (i *Interpreter) VisitPrintStatement(stmt *stm.PrintStmt) any {
	value := i.evaluate(stmt.Expression)
	fmt.Println(i.stringify(value))
	return nil
}

func (i *Interpreter) VisitIfStatement(stmt *stm.IfStmt) any {
	if i.isTruthy(i.evaluate(stmt.Condition)) {
		i.execute(stmt.ThenBranch)
	} else if stmt.ElseBranch != nil {
		i.execute(stmt.ElseBranch)
	}
	return nil
}

func (i *Interpreter) VisitWhileStatement(stmt *stm.WhileStmt) any {
	i.nearestEnclosingLoop = append(i.nearestEnclosingLoop, stmt)

	for {
		if !i.isTruthy(i.evaluate(stmt.Condition)) {
			break
		}
		i.execute(stmt.Body)
	}
	i.nearestEnclosingLoop = i.nearestEnclosingLoop[:len(i.nearestEnclosingLoop)-1]
	i.breaking = false
	return nil
}

func (i *Interpreter) VisitBreakStatement(stmt *stm.BreakStmt) any {
	if len(i.nearestEnclosingLoop) == 0 {
		panic("Break not in loop")
	}
	enclosingLoop := i.nearestEnclosingLoop[len(i.nearestEnclosingLoop)-1]
	i.breaking = true
	enclosingLoop.Condition = stm.NewLiteral(false)
	return nil
}

func (i *Interpreter) VisitVarStatement(stmt *stm.VarStmt) any {
	var value any = nil

	if stmt.Initializer != nil {
		value = i.evaluate(stmt.Initializer)
	}

	if stmt.Local {
		index := i.locals[stmt.Name]
		i.LocalVariables[index] = value
	}

	i.environment.Define(stmt.Name.Lexeme, value)

	return nil
}

func (i *Interpreter) VisitErrorStatement(stmt *stm.ErrorStmt) any {
	return nil
}

func (i *Interpreter) VisitAnonymousFuncExpr(expr *stm.AnonymousFunction) any {
	function := NewAnonymousFunction(*expr, i.environment)

	return function
}

// VisitBinaryExpr implements stm.Visitor.
func (i *Interpreter) VisitBinaryExpr(expr *stm.Binary) any {
	left := i.evaluate(expr.Left)
	right := i.evaluate(expr.Right)

	switch expr.Operator.TokenType {
	case tokens.MINUS:
		i.checkNumberOperands(expr.Operator, left, right)
		return left.(float64) - right.(float64)

	case tokens.SLASH:
		i.checkNumberOperands(expr.Operator, left, right)
		return left.(float64) / right.(float64)

	case tokens.STAR:
		i.checkNumberOperands(expr.Operator, left, right)
		return left.(float64) * right.(float64)

	case tokens.PLUS:
		if i.tryTypeAssert(left, reflect.Float64) && i.tryTypeAssert(right, reflect.Float64) {
			return left.(float64) + right.(float64)
		}
		if i.tryTypeAssert(left, reflect.String) && i.tryTypeAssert(right, reflect.String) {
			return left.(string) + right.(string)
		}

		if str, ok := i.tryParseValuesToString(left, right); ok {
			return *str
		}

		panic("Inconsistent types for + operation\n")

	case tokens.GREATER:
		i.checkNumberOperands(expr.Operator, left, right)
		return left.(float64) > right.(float64)

	case tokens.GREATER_EQUAL:
		i.checkNumberOperands(expr.Operator, left, right)
		return left.(float64) >= right.(float64)

	case tokens.LESS:
		i.checkNumberOperands(expr.Operator, left, right)
		return left.(float64) < right.(float64)

	case tokens.LESS_EQUAL:
		i.checkNumberOperands(expr.Operator, left, right)

		return left.(float64) <= right.(float64)

	case tokens.BANG_EQUAL:
		return !i.isEqual(left, right)

	case tokens.EQUAL_EQUAL:
		return i.isEqual(left, right)
	}

	panic("Cannot execute expression\n")
}

// VisitErrorExpr implements stm.Visitor.
func (i *Interpreter) VisitErrorExpr(expr *stm.Error) any {
	return expr.Value
}

func (i *Interpreter) VisitVariableExpr(expr *stm.Variable) any {
	return i.lookupVariable(expr.Name, expr)
}

// VisitGroupingExpr implements stm.Visitor.
func (i *Interpreter) VisitGroupingExpr(expr *stm.Grouping) any {
	return i.evaluate(expr.Expression)
}

// VisitLiteralExpr implements stm.Visitor.
func (i *Interpreter) VisitLiteralExpr(expr *stm.Literal) any {
	return expr.Value
}

func (i *Interpreter) VisitLogicalExpr(expr *stm.Logical) any {
	left := i.evaluate(expr.Left)

	if expr.Operator.TokenType == tokens.OR {
		if i.isTruthy(left) {
			return left
		}
	} else {
		if !i.isTruthy(left) {
			return left
		}
	}

	return i.evaluate(expr.Right)
}

// VisitTernaryExpr implements stm.Visitor.
func (i *Interpreter) VisitTernaryExpr(expr *stm.Ternary) any {
	condition := i.evaluate(expr.Condition)

	i.checkBoolOperands(expr.Operator, condition)

	if condition.(bool) {
		return i.evaluate(expr.Consequent)
	}
	return i.evaluate(expr.Alternative)
}

// VisitUnaryExpr implements stm.Visitor.
func (i *Interpreter) VisitUnaryExpr(expr *stm.Unary) any {
	right := i.evaluate(expr.Right)

	switch expr.Operator.TokenType {
	case tokens.MINUS:
		i.checkNumberOperands(expr.Operator, right)
		return -right.(float64)
	case tokens.BANG:
		return !i.isTruthy(right)

	}
	return nil
}

func (i *Interpreter) VisitCallExpr(expr *stm.Call) any {
	callee := i.evaluate(expr.Callee)
	arguments := make([]any, 0)

	for _, arg := range expr.Arguments {
		arguments = append(arguments, i.evaluate(arg))
	}

	function, ok := callee.(Callable)

	if !ok {
		panic("Can only call functions and classes.")
	}

	if function.Arity() != len(arguments) {
		errorMsg := fmt.Sprintf("line[%d] Expected %d arguments but got %d", expr.Paren.Line, function.Arity(), len(arguments))
		panic(errorMsg)
	}

	return function.Call(i, arguments)

}

func (i *Interpreter) VisitGetExpr(expr *stm.Get) any {
	object := i.evaluate(expr.Object)

	instance, ok := object.(IloxInstance)

	if !ok {
		panic("Only instances have properties.")
	}

	property, err := instance.Get(expr.Name, i)

	if err != nil {
		panic(err.Error())
	}

	return property

}

func (i *Interpreter) VisitSetExpr(expr *stm.Set) any {
	object := i.evaluate(expr.Object)
	instance, ok := object.(*LoxInstance)

	if !ok {
		panic("Only instances have fields.")
	}

	value := i.evaluate(expr.Value)

	instance.Set(expr.Name, value)

	return value
}

func (i *Interpreter) VisitThisExpr(expr *stm.This) any {
	return i.lookupVariable(expr.Keyword, expr)
}

func (i *Interpreter) VisitSuperExpr(expr *stm.Super) any {
	class := i.lookupVariable(expr.Keyword, expr).(*LoxClass)

	method, ok := class.FindMethod(expr.Method.Lexeme)

	if ok {
		this := i.LocalVariables[expr.ThisIndex].(*LoxInstance)
		method.Bind(this, i)
		return method
	}

	panic("")
}

func (i *Interpreter) VisitAssignExpr(expr *stm.Assign) any {
	value := i.evaluate(expr.Value)
	index, ok := i.locals[expr.Name]

	if ok {
		i.LocalVariables[index] = value
	} else {
		i.globals.Assign(expr.Name, value)
	}

	return value
}

func (i *Interpreter) isTruthy(value any) bool {
	if value == nil {
		return false
	}

	boolValue, ok := value.(bool)

	if ok {
		return boolValue
	}
	return true
}

func (i *Interpreter) lookupVariable(name tokens.Token, expr stm.Expression) any {
	index, ok := i.locals[name]

	if ok {
		return i.LocalVariables[index]
	}
	return i.globals.Get(name)
}

func (i *Interpreter) tryTypeAssert(value any, targetType reflect.Kind) bool {
	return reflect.TypeOf(value).Kind() == targetType
}

func (i *Interpreter) tryParseValuesToString(valueA, valueB any) (*string, bool) {
	strA, okA := i.tryReflectToString(valueA)
	strB, okB := i.tryReflectToString(valueB)

	if okA && okB {
		str := *strA + *strB
		return &str, true
	}
	return nil, false
}

func (i *Interpreter) tryReflectToString(value any) (*string, bool) {

	switch v := value.(type) {
	case string:
		return &v, true
	case int, float64:
		str := fmt.Sprintf("%v", v)
		return &str, true
	}

	val, ok := value.(fmt.Stringer)

	if ok {
		str := val.String()
		return &str, true
	}

	return nil, false
}

func (i *Interpreter) isEqual(left any, right any) bool {
	return reflect.DeepEqual(left, right)
}

func (i *Interpreter) checkNumberOperands(operator tokens.Token, operands ...any) {
	for _, operand := range operands {
		if !i.tryTypeAssert(operand, reflect.Float64) {
			panic(fmt.Sprintf("%s Operand must be a number \n [line %d]", operator.Lexeme, operator.Line))
		}
	}
}

func (i *Interpreter) checkBoolOperands(operator tokens.Token, operands ...any) {
	for _, operand := range operands {
		if !i.tryTypeAssert(operand, reflect.Bool) {
			panic(fmt.Sprintf("Ternary operator needs boolean value before ?\n [line %d]", operator.Line))
		}
	}
}

func (i *Interpreter) stringify(value any) string {
	if value == nil {
		return "nil"
	}
	if i.tryTypeAssert(value, reflect.Float64) {
		textValue := fmt.Sprintf("%v", value)

		if strings.HasSuffix(textValue, ".0") {
			return textValue[:len(textValue)-2]
		}

		return textValue
	}

	return fmt.Sprintf("%v", value)
}

func (i *Interpreter) afterPanic() {
	if r := recover(); r != nil {
		i.errorLogger.RuntimeError(r.(string))
	}
}
