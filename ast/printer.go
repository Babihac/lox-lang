package ast

import (
	"fmt"
	"lox/expressions"
	"strings"
)

type Printer struct{}

// VisitBinaryExpr implements expressions.Visitor.
func (p *Printer) VisitBinaryExpr(expr expressions.Binary) any {
	return p.parenthesize(expr.Operator.Lexeme, expr.Left, expr.Right)
}

func NewPrinter() *Printer {
	return &Printer{}
}

// VisitGroupingExpr implements expressions.Visitor.
func (p *Printer) VisitGroupingExpr(expr expressions.Grouping) any {
	return p.parenthesize("group", expr.Expression)
}

// VisitLiteralExpr implements expressions.Visitor.
func (p *Printer) VisitLiteralExpr(expr expressions.Literal) any {
	if expr.Value == nil {
		return "nil"
	}
	return fmt.Sprintf("%v", expr.Value)
}

func (p *Printer) VisitErrorExpr(expr expressions.Error) any {
	return expr.Value
}

// VisitUnaryExpr implements expressions.Visitor.
func (p *Printer) VisitUnaryExpr(expr expressions.Unary) any {
	return p.parenthesize(expr.Operator.Lexeme, expr.Right)
}

func (p *Printer) Print(expr expressions.Expression) string {
	return expr.Accept(p).(string)
}

func (p *Printer) VisitTernaryExpr(expr expressions.Ternary) any {
	return p.parenthesize("ternary", expr.Condition, expr.Consequent, expr.Alternative)
}

func (p *Printer) parenthesize(name string, exprs ...expressions.Expression) string {
	var builder strings.Builder

	builder.WriteRune('(')
	builder.WriteString(name)

	for _, expr := range exprs {
		builder.WriteRune(' ')
		builder.WriteString(expr.Accept(p).(string))
	}
	builder.WriteRune(')')

	return builder.String()

}
