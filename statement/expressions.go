package stm

import (
	"lox/tokens"
)

type ExprVisitor[T any] interface {
	VisitBinaryExpr(expr Binary) T
	VisitGroupingExpr(expr Grouping) T
	VisitLiteralExpr(expr Literal) T
	VisitUnaryExpr(expr Unary) T
	VisitErrorExpr(expr Error) T
	VisitTernaryExpr(expr Ternary) T
	VisitVariableExpr(expr Variable) T
	VisitAssignExpr(expr Assign) T
	VisitLogicalExpr(expr Logical) T
	VisitCallExpr(expr Call) T
	VisitAnonymousFuncExpr(expr AnonymousFunction) T
}

type Expression interface {
	Accept(visitor ExprVisitor[any]) any
}

type Grouping struct {
	Expression Expression
}

func NewGrouping(expr Expression) *Grouping {
	return &Grouping{
		Expression: expr,
	}
}

func (g Grouping) Accept(visitor ExprVisitor[any]) any {
	return visitor.VisitGroupingExpr(g)
}

type Assign struct {
	Name  tokens.Token
	Value Expression
}

func NewAssign(name tokens.Token, value Expression) *Assign {
	return &Assign{
		Name:  name,
		Value: value,
	}
}

func (a Assign) Accept(visitor ExprVisitor[any]) any {
	return visitor.VisitAssignExpr(a)
}

type Binary struct {
	Left     Expression
	Operator tokens.Token
	Right    Expression
}

func NewBinary(left Expression, operator tokens.Token, right Expression) *Binary {
	return &Binary{
		Left:     left,
		Operator: operator,
		Right:    right,
	}
}

func (b Binary) Accept(visitor ExprVisitor[any]) any {
	return visitor.VisitBinaryExpr(b)
}

type Literal struct {
	Value any
}

func NewLiteral(value any) *Literal {
	return &Literal{
		Value: value,
	}
}

func (l Literal) Accept(visitor ExprVisitor[any]) any {
	return visitor.VisitLiteralExpr(l)
}

type Logical struct {
	Left     Expression
	Operator tokens.Token
	Right    Expression
}

func NewLogical(left Expression, operator tokens.Token, right Expression) *Logical {
	return &Logical{
		Left:     left,
		Operator: operator,
		Right:    right,
	}
}

func (l Logical) Accept(visitor ExprVisitor[any]) any {
	return visitor.VisitLogicalExpr(l)
}

type Unary struct {
	Operator tokens.Token
	Right    Expression
}

func NewUnary(operator tokens.Token, right Expression) *Unary {
	return &Unary{
		Operator: operator,
		Right:    right,
	}
}

func (u Unary) Accept(visitor ExprVisitor[any]) any {
	return visitor.VisitUnaryExpr(u)
}

type Error struct {
	Value string
}

func NewErrorExpr(value string) *Error {
	return &Error{Value: value}
}

func (e Error) Accept(visitor ExprVisitor[any]) any {
	return visitor.VisitErrorExpr(e)
}

type Ternary struct {
	Operator    tokens.Token
	Condition   Expression
	Consequent  Expression
	Alternative Expression
}

func NewTernary(operator tokens.Token, condition Expression, consequent Expression, alternative Expression) *Ternary {
	return &Ternary{
		Operator:    operator,
		Condition:   condition,
		Consequent:  consequent,
		Alternative: alternative,
	}
}

func (t Ternary) Accept(visitor ExprVisitor[any]) any {
	return visitor.VisitTernaryExpr(t)
}

type Variable struct {
	Name tokens.Token
}

func NewVariable(name tokens.Token) *Variable {
	return &Variable{
		Name: name,
	}
}

func (v Variable) Accept(visitor ExprVisitor[any]) any {
	return visitor.VisitVariableExpr(v)
}

type Call struct {
	Callee    Expression
	Paren     tokens.Token
	Arguments []Expression
}

func NewCall(callee Expression, paren tokens.Token, arguments []Expression) *Call {
	return &Call{
		Callee:    callee,
		Paren:     paren,
		Arguments: arguments,
	}
}

func (c Call) Accept(visitor ExprVisitor[any]) any {
	return visitor.VisitCallExpr(c)
}

type AnonymousFunction struct {
	Params []tokens.Token
	Body   []Statement
}

func NewAnonymousFunction(params []tokens.Token, body []Statement) *AnonymousFunction {
	return &AnonymousFunction{
		Params: params,
		Body:   body,
	}
}

func (af AnonymousFunction) Accept(visitor ExprVisitor[any]) any {
	return visitor.VisitAnonymousFuncExpr(af)
}
