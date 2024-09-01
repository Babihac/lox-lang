package main

import (
	"lox/errorLogger"
	"lox/interpreter"
	"lox/lox"
	"lox/parser"
	"lox/resolver"
	"lox/scanner"
)

func main() {
	lox := lox.Lox{}

	errorLogger := errorLogger.ErrorLogger{
		Lox: &lox,
	}

	lox.ErrorLogger = errorLogger

	scanner := scanner.NewScanner(errorLogger)
	parses := parser.NewParser(errorLogger)
	interpreter := interpreter.NewInterpreter(errorLogger)
	resolver := resolver.NewResolver(interpreter, errorLogger)

	lox.SetComponents(scanner, parses, interpreter, resolver)

	lox.RunFile("testFiles/inheritance.txt")

	// expr := expressions.Binary{
	// 	Left:     expressions.Binary{Operator: tokens.NewToken(tokens.PLUS, "+", nil, 1), Left: expressions.Literal{Value: 1}, Right: expressions.Literal{Value: 2}},
	// 	Operator: tokens.NewToken(tokens.STAR, "*", nil, 1),
	// 	Right:    expressions.Binary{Operator: tokens.NewToken(tokens.MINUS, "-", nil, 1), Left: expressions.Literal{Value: 4}, Right: expressions.Literal{Value: 3}},
	// }

	// printer := ast.NewPrinter()

	// fmt.Println(printer.Print(expr))

}
