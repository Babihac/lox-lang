package stm

import (
	"lox/tokens"
)

type StmVisitor[T any] interface {
	VisitExprStatement(stmt *ExpressionStmt) T
	VisitPrintStatement(stmt *PrintStmt) T
	VisitVarStatement(stmt *VarStmt) T
	VisitErrorStatement(stmt *ErrorStmt) T
	VisitBlockStatement(stmt *BlockStmt) T
	VisitIfStatement(stmt *IfStmt) T
	VisitWhileStatement(stmt *WhileStmt) T
	VisitBreakStatement(stmt *BreakStmt) T
	VisitFunctionStatement(stmt *FunctionStm) T
	VisitReturnStatement(stmt *ReturnStmt) T
	VisitClassStatement(stmt *ClassStmt) T
}

type Statement interface {
	Accept(visitor StmVisitor[any]) any
}

type ExpressionStmt struct {
	Expression Expression
}

func NewExpression(expr Expression) *ExpressionStmt {
	return &ExpressionStmt{
		Expression: expr,
	}
}

func (e *ExpressionStmt) Accept(visitor StmVisitor[any]) any {
	return visitor.VisitExprStatement(e)
}

type IfStmt struct {
	Condition  Expression
	ThenBranch Statement
	ElseBranch Statement
}

func NewIf(condition Expression, thenStm, elseStm Statement) *IfStmt {
	return &IfStmt{
		Condition:  condition,
		ThenBranch: thenStm,
		ElseBranch: elseStm,
	}
}

func (i *IfStmt) Accept(visitor StmVisitor[any]) any {
	return visitor.VisitIfStatement(i)
}

type PrintStmt struct {
	Expression Expression
}

func NewPrint(expr Expression) *PrintStmt {
	return &PrintStmt{
		Expression: expr,
	}
}

func (p *PrintStmt) Accept(visitor StmVisitor[any]) any {
	return visitor.VisitPrintStatement(p)
}

type WhileStmt struct {
	Condition Expression
	Body      Statement
}

func NewWhile(condition Expression, body Statement) *WhileStmt {
	return &WhileStmt{
		Condition: condition,
		Body:      body,
	}
}

func (w *WhileStmt) Accept(visitor StmVisitor[any]) any {
	return visitor.VisitWhileStatement(w)
}

type VarStmt struct {
	Name        tokens.Token
	Initializer Expression
	Local       bool
}

func NewVar(name tokens.Token, expr Expression) *VarStmt {
	return &VarStmt{
		Name:        name,
		Initializer: expr,
		Local:       false,
	}
}

func (v *VarStmt) Accept(visitior StmVisitor[any]) any {
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

func (b *BlockStmt) Accept(visitor StmVisitor[any]) any {
	return visitor.VisitBlockStatement(b)
}

type BreakStmt struct {
}

func NewBreak() *BreakStmt {
	return &BreakStmt{}
}

func (b *BreakStmt) Accept(visitor StmVisitor[any]) any {
	return visitor.VisitBreakStatement(b)
}

type FunctionStm struct {
	Name      tokens.Token
	Params    []tokens.Token
	Body      []Statement
	ThisIndex int
}

func NewFunction(name tokens.Token, params []tokens.Token, body []Statement) *FunctionStm {
	return &FunctionStm{
		Name:      name,
		Params:    params,
		Body:      body,
		ThisIndex: -1,
	}
}

func (f *FunctionStm) Accept(visitor StmVisitor[any]) any {
	return visitor.VisitFunctionStatement(f)
}

type ErrorStmt struct {
	Message string
}

func NewError(msg string) *ErrorStmt {
	return &ErrorStmt{
		Message: msg,
	}
}

func (e *ErrorStmt) Accept(visitor StmVisitor[any]) any {
	return visitor.VisitErrorStatement(e)
}

type ReturnStmt struct {
	Keyword tokens.Token
	Value   Expression
}

func NewReturn(keyword tokens.Token, value Expression) *ReturnStmt {
	return &ReturnStmt{
		Keyword: keyword,
		Value:   value,
	}
}

func (r *ReturnStmt) Accept(visitor StmVisitor[any]) any {
	return visitor.VisitReturnStatement(r)
}

type ClassStmt struct {
	Name          tokens.Token
	Methods       []*FunctionStm
	StaticMethods []*FunctionStm
	SuperClass    *Variable
	SuperIndex    int
}

func NewClass(name tokens.Token, methods []*FunctionStm, staticMethods []*FunctionStm, superClass *Variable) *ClassStmt {
	return &ClassStmt{
		Name:          name,
		Methods:       methods,
		StaticMethods: staticMethods,
		SuperClass:    superClass,
		SuperIndex:    -1,
	}
}

func (c *ClassStmt) Accept(visitor StmVisitor[any]) any {
	return visitor.VisitClassStatement(c)
}
