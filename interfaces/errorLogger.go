package interfaces

import "lox/tokens"

type ErrorLogger interface {
	Error(line int, message string)
	Report(line int, where string, message string)
	ErrorForToken(token tokens.Token, message string) error
	RuntimeError(message string)
}
