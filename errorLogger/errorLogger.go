package errorLogger

import (
	"errors"
	"fmt"
	"lox/lox"
	"lox/tokens"
)

type ErrorLogger struct {
	runtimeErrorMessage string
	runtimeErrorLine    int
	Lox                 *lox.Lox
}

func NewErrorLogger(lox *lox.Lox) ErrorLogger {
	return ErrorLogger{Lox: lox}
}

func (el ErrorLogger) Error(line int, message string) {
	el.Report(line, "", message)
}

func (el ErrorLogger) Report(line int, where string, message string) {
	el.Lox.HadError = true
	fmt.Printf("[Line: %d] Error %s: %s", line, where, message)
}

func (el ErrorLogger) ErrorForToken(token tokens.Token, message string) error {
	if token.TokenType != tokens.EOF {
		el.Report(token.Line, fmt.Sprintf("at '%s '", token.Lexeme), message)
	}

	return errors.New("ParseError")
}

func (el ErrorLogger) RuntimeError(message string) {
	fmt.Println(message)
	el.Lox.HadRuntimeError = true
}
