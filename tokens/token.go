package tokens

import (
	"fmt"
)

type Token struct {
	TokenType TokenType
	Lexeme    string
	Literal   interface{}
	Line      int
}

func NewToken(tokenType TokenType, lexeme string, literal interface{}, line int) Token {
	return Token{
		Line:      line,
		TokenType: tokenType,
		Lexeme:    lexeme,
		Literal:   literal,
	}
}

func (t Token) String() string {
	return fmt.Sprintf("%v %s %v", t.TokenType, t.Lexeme, t.Literal)
}
