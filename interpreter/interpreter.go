package interpreter

import (
	"fmt"
	env "lox/environment"
	"lox/expressions"
	"lox/interfaces"
	stm "lox/statement"
	"lox/tokens"
	"reflect"
	"strings"
)

type Interpreter struct {
	errorLogger interfaces.ErrorLogger
	environment *env.Environment
}

func NewInterpreter(errorLogger interfaces.ErrorLogger) *Interpreter {
	return &Interpreter{
		errorLogger: errorLogger,
		environment: env.NewEnvironment(),
	}
}

func (i *Interpreter) Interpret(statements []stm.Statement) {
	defer i.afterPanic()

	for _, stmt := range statements {
		i.execute(stmt)
	}
}

func (i *Interpreter) InterpretRepl(statements []stm.Statement) {

}

func (i *Interpreter) execute(stmt stm.Statement) {
	stmt.Accept(i)
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

func (i *Interpreter) evaluate(expr expressions.Expression) any {
	return expr.Accept(i)
}

func (i *Interpreter) VisitBlockStatement(stmt stm.BlockStmt) any {
	i.executeBlock(stmt.Statements, env.NewEnvironment(i.environment))
	return nil
}

// VisitExprStatement implements stm.Visitor.
func (i *Interpreter) VisitExprStatement(stmt stm.ExpressionStmt) any {
	value := i.evaluate(stmt.Expression)
	fmt.Println(value)
	return nil
}

// VisitPrintStatement implements stm.Visitor.
func (i *Interpreter) VisitPrintStatement(stmt stm.PrintStmt) any {
	value := i.evaluate(stmt.Expression)
	fmt.Println(i.stringify(value))
	return nil
}

func (i *Interpreter) VisitVarStatement(stmt stm.VarStmt) any {
	var value any = nil

	if stmt.Initializer != nil {
		value = i.evaluate(stmt.Initializer)
	}

	i.environment.Define(stmt.Name.Lexeme, value)

	return nil
}

func (i *Interpreter) VisitErrorStatement(stmt stm.ErrorStmt) any {
	fmt.Println(stmt.Message)
	return nil
}

// VisitBinaryExpr implements expressions.Visitor.
func (i *Interpreter) VisitBinaryExpr(expr expressions.Binary) any {
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

// VisitErrorExpr implements expressions.Visitor.
func (i *Interpreter) VisitErrorExpr(expr expressions.Error) any {
	return expr.Value
}

func (i *Interpreter) VisitVariableExpr(expr expressions.Variable) any {
	return i.environment.Get(expr.Name)
}

// VisitGroupingExpr implements expressions.Visitor.
func (i *Interpreter) VisitGroupingExpr(expr expressions.Grouping) any {
	return i.evaluate(expr.Expression)
}

// VisitLiteralExpr implements expressions.Visitor.
func (i *Interpreter) VisitLiteralExpr(expr expressions.Literal) any {
	return expr.Value
}

// VisitTernaryExpr implements expressions.Visitor.
func (i *Interpreter) VisitTernaryExpr(expr expressions.Ternary) any {
	condition := i.evaluate(expr.Condition)

	i.checkBoolOperands(expr.Operator, condition)

	if condition.(bool) {
		return i.evaluate(expr.Consequent)
	}
	return i.evaluate(expr.Alternative)
}

// VisitUnaryExpr implements expressions.Visitor.
func (i *Interpreter) VisitUnaryExpr(expr expressions.Unary) any {
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

func (i *Interpreter) VisitAssignExpr(expr expressions.Assign) any {
	value := i.evaluate(expr.Value)
	i.environment.Assign(expr.Name, value)
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
