package lox

import (
	"bufio"
	"fmt"
	"log"
	"lox/interfaces"
	"lox/interpreter"
	"lox/parser"
	"lox/resolver"
	"lox/scanner"
	"os"
)

type Token struct{}

type Lox struct {
	HadError        bool
	HadRuntimeError bool
	ErrorLogger     interfaces.ErrorLogger
	scanner         *scanner.Scanner
	parser          *parser.Parser
	interpreter     *interpreter.Interpreter
	resolver        *resolver.Resolver
}

func (l *Lox) SetComponents(scanner *scanner.Scanner, parser *parser.Parser, interpreter *interpreter.Interpreter, resolver *resolver.Resolver) {
	l.scanner = scanner
	l.parser = parser
	l.interpreter = interpreter
	l.resolver = resolver
}

func (l *Lox) RunFile(path string) {
	file, err := os.ReadFile(path)

	if err != nil {
		log.Fatal(err)
	}

	l.run(string(file))

	if l.HadError {
		os.Exit(65)
	}

	if l.HadRuntimeError {
		os.Exit(70)
	}
}

func (l *Lox) RunPrompt() {

	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print(">")

		if !scanner.Scan() {
			fmt.Println("Exit")
			break
		}

		source := scanner.Text()
		l.runRepl(source)
		l.HadError = false
		l.HadRuntimeError = false

	}

	if scanner.Err() != nil {
		log.Fatal(scanner.Err())
	}

}

func (l *Lox) run(source string) {

	l.scanner.LoadSource(source)

	tokens := l.scanner.ScanTokens()

	l.parser.LoadTokens(tokens)

	stmts := l.parser.Parse()

	if l.HadError {
		return
	}

	l.resolver.ResolveBlock(stmts)

	if l.HadError {
		return
	}

	l.interpreter.Interpret(stmts)

}

func (l *Lox) runRepl(source string) {
	l.scanner.LoadSource(source)

	tokens := l.scanner.ScanTokens()

	l.parser.LoadTokens(tokens)

	stmts := l.parser.Parse()

	if l.HadError {
		return
	}

	l.interpreter.InterpretRepl(stmts)
}
