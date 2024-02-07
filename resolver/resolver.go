package resolver

import (
	"lox/interfaces"
	"lox/interpreter"
	stm "lox/statement"
	"lox/tokens"
)

type FunctionType int

const (
	NONE FunctionType = iota
	FUNCTION
	ANONYMOUS_FUNCTION
)

type LocalVariable struct {
	index   int
	defined bool
}

type Resolver struct {
	Interpreter     *interpreter.Interpreter
	Scopes          []map[string]*LocalVariable
	ErrorLogger     interfaces.ErrorLogger
	currentFunction FunctionType
	localIndex      int
}

func NewResolver(interpreter *interpreter.Interpreter, errorLogger interfaces.ErrorLogger) *Resolver {
	return &Resolver{
		Interpreter:     interpreter,
		ErrorLogger:     errorLogger,
		currentFunction: NONE,
		localIndex:      0,
	}
}

func (r *Resolver) ResolveBlock(statements []stm.Statement) {
	for _, stm := range statements {
		r.resolveStm(stm)
	}
}

func (r *Resolver) resolveLocal(expr stm.Expression, name tokens.Token) {
	for i := len(r.Scopes) - 1; i >= 0; i-- {
		variable, ok := r.Scopes[i][name.Lexeme]

		if ok {
			r.Interpreter.Resolve(name, len(r.Scopes)-1-i, variable.index)
			return
		}

	}
}

func (r *Resolver) resolveStm(statement stm.Statement) {
	statement.Accept(r)
}

func (r *Resolver) resolveExpr(expression stm.Expression) {
	expression.Accept(r)
}

func (r *Resolver) resolveFunction(function stm.FunctionStm, funcType FunctionType) {
	enclosingFunc := r.currentFunction
	r.currentFunction = funcType

	r.beginScope()

	for _, token := range function.Params {
		r.declare(token)
		r.define(token)
		r.Interpreter.Resolve(token, 0, r.localIndex-1)
		r.Interpreter.LocalVariables = append(r.Interpreter.LocalVariables, nil)
	}

	r.ResolveBlock(function.Body)
	r.endScope()

	r.currentFunction = enclosingFunc
}

func (r *Resolver) resolveAnonymousFunction(function stm.AnonymousFunction) {
	enclosingFunc := r.currentFunction
	r.currentFunction = ANONYMOUS_FUNCTION

	r.beginScope()

	for _, token := range function.Params {
		r.declare(token)
		r.define(token)
	}

	r.ResolveBlock(function.Body)

	r.endScope()

	r.currentFunction = enclosingFunc
}

func (r *Resolver) beginScope() {
	r.Scopes = append(r.Scopes, make(map[string]*LocalVariable))
}

func (r *Resolver) endScope() {
	r.Scopes = r.Scopes[:len((r.Scopes))-1]
}

func (r *Resolver) declare(name tokens.Token) {
	if len(r.Scopes) == 0 {
		return
	}
	scope := r.Scopes[len(r.Scopes)-1]

	_, ok := scope[name.Lexeme]

	if ok {
		r.ErrorLogger.ErrorForToken(name, "Already variable with this name in this scope.")
	}

	scope[name.Lexeme] = &LocalVariable{
		index:   r.localIndex,
		defined: false,
	}

	r.localIndex++

}
func (r *Resolver) define(name tokens.Token) {
	if len(r.Scopes) == 0 {
		return
	}
	scope := r.Scopes[len(r.Scopes)-1]
	scope[name.Lexeme].defined = true

	r.Interpreter.Resolve(name, 0, scope[name.Lexeme].index)
	r.Interpreter.LocalVariables = append(r.Interpreter.LocalVariables, nil)
}

// VisitAnonymousFuncExpr implements stm.ExprVisitor.
func (r *Resolver) VisitAnonymousFuncExpr(expr stm.AnonymousFunction) any {
	r.resolveAnonymousFunction(expr)

	return nil
}

// VisitAssignExpr implements stm.ExprVisitor.
func (r *Resolver) VisitAssignExpr(expr *stm.Assign) any {
	r.resolveExpr(expr.Value)
	r.resolveLocal(expr, expr.Name)

	return nil
}

// VisitBinaryExpr implements stm.ExprVisitor.
func (r *Resolver) VisitBinaryExpr(expr stm.Binary) any {
	r.resolveExpr(expr.Left)
	r.resolveExpr(expr.Right)

	return nil
}

// VisitCallExpr implements stm.ExprVisitor.
func (r *Resolver) VisitCallExpr(expr stm.Call) any {
	r.resolveExpr(expr.Callee)

	for _, arg := range expr.Arguments {
		r.resolveExpr(arg)
	}

	return nil
}

// VisitErrorExpr implements stm.ExprVisitor.
func (r *Resolver) VisitErrorExpr(expr stm.Error) any {
	return nil
}

// VisitGroupingExpr implements stm.ExprVisitor.
func (r *Resolver) VisitGroupingExpr(expr stm.Grouping) any {
	r.resolveExpr(expr.Expression)

	return nil
}

// VisitLiteralExpr implements stm.ExprVisitor.
func (r *Resolver) VisitLiteralExpr(expr stm.Literal) any {
	return nil
}

// VisitLogicalExpr implements stm.ExprVisitor.
func (r *Resolver) VisitLogicalExpr(expr stm.Logical) any {
	r.resolveExpr(expr.Left)
	r.resolveExpr(expr.Right)

	return nil
}

// VisitTernaryExpr implements stm.ExprVisitor.
func (r *Resolver) VisitTernaryExpr(expr stm.Ternary) any {
	r.resolveExpr(expr.Condition)
	r.resolveExpr(expr.Consequent)
	r.resolveExpr(expr.Alternative)

	return nil
}

// VisitUnaryExpr implements stm.ExprVisitor.
func (r *Resolver) VisitUnaryExpr(expr stm.Unary) any {
	r.resolveExpr(expr.Right)

	return nil
}

// VisitVariableExpr implements stm.ExprVisitor.
func (r *Resolver) VisitVariableExpr(expr *stm.Variable) any {
	if len(r.Scopes) != 0 {
		variable, ok := r.Scopes[len(r.Scopes)-1][expr.Name.Lexeme]
		if ok && !variable.defined {
			r.ErrorLogger.ErrorForToken(expr.Name, "Can't read local variable in its own initializer.")
		}
	}

	r.resolveLocal(expr, expr.Name)
	return nil
}

// VisitBlockStatement implements stm.StmVisitor.
func (r *Resolver) VisitBlockStatement(stmt stm.BlockStmt) any {
	r.beginScope()
	r.ResolveBlock(stmt.Statements)
	r.endScope()

	return nil
}

// VisitBreakStatement implements stm.StmVisitor.
func (r *Resolver) VisitBreakStatement(stmt stm.BreakStmt) any {
	return nil
}

// VisianyErrorSanyatement implements stm.StmVisitor.
func (r *Resolver) VisitErrorStatement(stmt stm.ErrorStmt) any {
	return nil
}

// VisitExprStatement implements stm.StmVisitor.
func (r *Resolver) VisitExprStatement(stmt stm.ExpressionStmt) any {
	r.resolveExpr(stmt.Expression)
	return nil
}

// VisitFunctionStatement implements stm.StmVisitor.
func (r *Resolver) VisitFunctionStatement(stmt stm.FunctionStm) any {
	r.declare(stmt.Name)
	r.define(stmt.Name)

	r.resolveFunction(stmt, FUNCTION)
	return nil
}

// VisitIfStatement implements stm.StmVisitor.
func (r *Resolver) VisitIfStatement(stmt stm.IfStmt) any {
	r.resolveExpr(stmt.Condition)
	r.resolveStm(stmt.ThenBranch)

	if stmt.ElseBranch != nil {
		r.resolveStm(stmt.ElseBranch)
	}

	return nil
}

// VisitPrintStatement implements stm.StmVisitor.
func (r *Resolver) VisitPrintStatement(stmt stm.PrintStmt) any {
	r.resolveExpr(stmt.Expression)

	return nil
}

// VisitReturnStatement implements stm.StmVisitor.
func (r *Resolver) VisitReturnStatement(stmt stm.ReturnStmt) any {

	if r.currentFunction == NONE {
		r.ErrorLogger.ErrorForToken(stmt.Keyword, "Can't return from top-level code.")
	}

	if stmt.Value != nil {
		r.resolveExpr(stmt.Value)
	}

	return nil
}

// VisitVarStatement implements stm.StmVisitor.
func (r *Resolver) VisitVarStatement(stmt *stm.VarStmt) any {
	r.declare(stmt.Name)

	if stmt.Initializer != nil {
		r.resolveExpr(stmt.Initializer)
	}

	r.define(stmt.Name)

	if len(r.Scopes) > 0 {
		stmt.Local = true
	}

	return nil
}

// VisitWhileStatement implements stm.StmVisitor.
func (r *Resolver) VisitWhileStatement(stmt stm.WhileStmt) any {
	r.resolveExpr(stmt.Condition)
	r.resolveStm(stmt.Body)

	return nil
}
