package scanner

import (
	"lox/interfaces"
	"lox/tokens"
	"strconv"
	"unicode"
)

type Scanner struct {
	source      string
	tokens      []tokens.Token
	start       int
	current     int
	line        int
	keywords    map[string]tokens.TokenType
	errorLogger interfaces.ErrorLogger
}

func (sc *Scanner) GetSource() string {
	return sc.source
}

func NewScanner(errorLogger interfaces.ErrorLogger) *Scanner {
	tokensInit := make([]tokens.Token, 0)
	return &Scanner{
		tokens:      tokensInit,
		start:       0,
		current:     0,
		line:        1,
		errorLogger: errorLogger,
		keywords: map[string]tokens.TokenType{
			"and":    tokens.AND,
			"class":  tokens.CLASS,
			"else":   tokens.ELSE,
			"false":  tokens.FALSE,
			"for":    tokens.FOR,
			"fun":    tokens.FUN,
			"if":     tokens.IF,
			"nil":    tokens.NIL,
			"or":     tokens.OR,
			"print":  tokens.PRINT,
			"return": tokens.RETURN,
			"super":  tokens.SUPER,
			"this":   tokens.THIS,
			"true":   tokens.TRUE,
			"var":    tokens.VAR,
			"while":  tokens.WHILE,
			"break":  tokens.BREAK,
		},
	}
}

func (sc *Scanner) LoadSource(source string) {
	tokensInit := make([]tokens.Token, 0)
	sc.start = 0
	sc.current = 0
	sc.line = 1
	sc.tokens = tokensInit
	sc.source = source
}

func (sc *Scanner) ScanTokens() []tokens.Token {

	for {
		if sc.isAtEnd() {
			break
		}
		sc.start = sc.current
		sc.scanToken()
	}
	eofToken := tokens.NewToken(tokens.EOF, "", nil, sc.line)
	sc.tokens = append(sc.tokens, eofToken)
	return sc.tokens
}

func (sc *Scanner) isAtEnd() bool {
	return sc.current >= len(sc.source)
}

func (sc *Scanner) scanToken() {
	c := sc.advance()
	switch c {
	case '(':
		sc.addToken(tokens.LEFT_PAREN)
	case ')':
		sc.addToken(tokens.RIGHT_PAREN)
	case '{':
		sc.addToken(tokens.LEFT_BRACE)
	case '}':
		sc.addToken(tokens.RIGHT_BRACE)
	case ',':
		sc.addToken(tokens.COMMA)
	case '.':
		sc.addToken(tokens.DOT)
	case '-':
		sc.addToken(tokens.MINUS)
	case '+':
		sc.addToken(tokens.PLUS)
	case ';':
		sc.addToken(tokens.SEMICOLON)
	case '*':
		sc.addToken(tokens.STAR)
	case '?':
		sc.addToken(tokens.QUESTION_MARK)
	case ':':
		sc.addToken(tokens.COLON)
	case '!':
		sc.addConditionalToken(sc.match('='), tokens.BANG_EQUAL, tokens.BANG)
	case '=':
		sc.addConditionalToken(sc.match('='), tokens.EQUAL_EQUAL, tokens.EQUAL)
	case '<':
		sc.addConditionalToken(sc.match('='), tokens.LESS_EQUAL, tokens.LESS)
	case '>':
		sc.addConditionalToken(sc.match('='), tokens.GREATER_EQUAL, tokens.GREATER)
	case ' ', '\r', '\t':
	case '\n':
		sc.line++
	case '/':
		if sc.match('/') {
			for {
				if sc.peek() == '\n' || sc.isAtEnd() {
					break
				}
				sc.advance()
			}
		} else if sc.match('*') {
			sc.multiLineComment()
		} else {
			sc.addToken(tokens.SLASH)
		}
	case '"':
		sc.string()
	default:
		if unicode.IsDigit(c) {
			sc.number()
		} else if sc.isAlpha(c) {
			sc.identifier()
		} else {
			sc.errorLogger.Error(sc.line, "Unexpected character.")
		}
	}
}

func (sc *Scanner) advance() rune {
	sc.current++
	return rune(sc.source[sc.current-1])
}

func (sc *Scanner) addToken(tokenType tokens.TokenType) {
	sc.addTokenWithLiteral(tokenType, nil)
}

func (sc *Scanner) addTokenWithLiteral(tokenType tokens.TokenType, literal interface{}) {
	text := sc.source[sc.start:sc.current]

	sc.tokens = append(sc.tokens, tokens.NewToken(tokenType, text, literal, sc.line))
}

func (sc *Scanner) match(expected rune) bool {
	if sc.isAtEnd() {
		return false
	}
	if rune(sc.source[sc.current]) != expected {
		return false
	}

	sc.current++

	return true
}

func (sc *Scanner) addConditionalToken(condition bool, tokenTypeIfTrue, tokenTypeIfFalse tokens.TokenType) {
	if condition {
		sc.addToken(tokenTypeIfTrue)
	} else {
		sc.addToken(tokenTypeIfFalse)
	}
}

func (sc *Scanner) peek() rune {
	if sc.isAtEnd() {
		return 0
	}
	return rune(sc.source[sc.current])
}

func (sc *Scanner) peekNext() rune {
	if sc.current+1 >= len(sc.source) {
		return 0
	}
	return rune(sc.source[sc.current+1])
}

func (sc *Scanner) string() {
	for {
		if sc.peek() == '"' || sc.isAtEnd() {
			break
		}
		if sc.peek() == '\n' {
			sc.line++
		}
		sc.advance()
	}

	if sc.isAtEnd() {
		sc.errorLogger.Error(sc.line, "Unterminated string")
		return
	}
	sc.advance()

	value := sc.source[sc.start+1 : sc.current-1]
	sc.addTokenWithLiteral(tokens.STRING, value)
}

func (sc *Scanner) number() {
	for {
		if !unicode.IsDigit(sc.peek()) {
			break
		}
		sc.advance()
	}

	if sc.peek() == '.' && unicode.IsDigit(sc.peekNext()) {
		sc.advance()
		for {
			if !unicode.IsDigit(sc.peek()) {
				break
			}
			sc.advance()
		}
	}

	number, err := strconv.ParseFloat(sc.source[sc.start:sc.current], 64)

	if err != nil {
		sc.errorLogger.Error(sc.line, "Invalid float number definition")
	}

	sc.addTokenWithLiteral(tokens.NUMBER, number)
}

func (sc *Scanner) multiLineComment() {
	for {
		if sc.isAtEnd() {
			sc.errorLogger.Error(sc.line, "Unclosed multiline comment")
			break
		}
		c := sc.advance()

		if c == '*' && sc.match('/') {
			break
		}
		if c == '\n' {
			sc.line++
		}
	}
}

func (sc *Scanner) identifier() {
	for {
		if !sc.isAlphaNumeric(sc.peek()) {
			break
		}
		sc.advance()
	}

	text := sc.source[sc.start:sc.current]

	tokenType, ok := sc.keywords[text]

	if !ok {
		tokenType = tokens.IDENTIFIER
	}
	sc.addToken(tokenType)
}

func (sc *Scanner) isAlpha(c rune) bool {
	return c == '_' || unicode.IsLetter(c)
}

func (sc *Scanner) isAlphaNumeric(c rune) bool {
	return sc.isAlpha(c) || unicode.IsDigit(c)
}
