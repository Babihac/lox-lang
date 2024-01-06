package stm

import (
	"lox/expressions"
	"lox/tokens"
)

type Visitor[T any] interface {
	VisitExprStatement(stmt ExpressionStmt) T
	VisitPrintStatement(stmt PrintStmt) T
	VisitVarStatement(stmt VarStmt) T
	VisitErrorStatement(stmt ErrorStmt) T
	VisitBlockStatement(stmt BlockStmt) T
}

type Statement interface {
	Accept(visitor Visitor[any]) any
}

type ExpressionStmt struct {
	Expression expressions.Expression
}

func NewExpression(expr expressions.Expression) *ExpressionStmt {
	return &ExpressionStmt{
		Expression: expr,
	}
}

func (e ExpressionStmt) Accept(visitor Visitor[any]) any {
	return visitor.VisitExprStatement(e)
}

type PrintStmt struct {
	Expression expressions.Expression
}

func NewPrint(expr expressions.Expression) *PrintStmt {
	return &PrintStmt{
		Expression: expr,
	}
}

func (p PrintStmt) Accept(visitor Visitor[any]) any {
	return visitor.VisitPrintStatement(p)
}

type VarStmt struct {
	Name        tokens.Token
	Initializer expressions.Expression
}

func NewVar(name tokens.Token, expr expressions.Expression) *VarStmt {
	return &VarStmt{
		Name:        name,
		Initializer: expr,
	}
}

func (v VarStmt) Accept(visitior Visitor[any]) any {
	return visitior.VisitVarStatement(v)
}

type BlockStmt struct {
	Statements []Statement
}

func NewBlock(statements []Statement) *BlockStmt {
	return &BlockStmt{
		Statements: statements,
	}
}

func (b BlockStmt) Accept(visitor Visitor[any]) any {
	return visitor.VisitBlockStatement(b)
}

type ErrorStmt struct {
	Message string
}

func NewError(msg string) *ErrorStmt {
	return &ErrorStmt{
		Message: msg,
	}
}

func (e ErrorStmt) Accept(visitor Visitor[any]) any {
	return visitor.VisitErrorStatement(e)
}
