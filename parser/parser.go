package parser

import (
	"errors"
	"fmt"
	"lox/expressions"
	"lox/interfaces"
	stm "lox/statement"
	"lox/tokens"
)

type ErrorHandlerEvent struct {
	Err  error
	Done chan bool
}

type Parser struct {
	tokens       []tokens.Token
	errorLogger  interfaces.ErrorLogger
	current      int
	errorHandler chan ErrorHandlerEvent
}

func NewParser(errorLogger interfaces.ErrorLogger) *Parser {
	parser := Parser{
		current:      0,
		errorLogger:  errorLogger,
		errorHandler: make(chan ErrorHandlerEvent),
	}

	go parser.handleError()

	return &parser
}

func (p *Parser) LoadTokens(tokens []tokens.Token) {
	p.tokens = tokens
	p.current = 0
}

func (p *Parser) handleError() {
	for event := range p.errorHandler {
		p.synchronize()
		event.Done <- true
	}
}

func (p *Parser) handlePanic() {
	if r := recover(); r != nil {
		p.errorLogger.Error(p.peek().Line, r.(string))
	}
}

func (p *Parser) Parse() []stm.Statement {
	var statemets []stm.Statement

	for {
		if p.isAtEnd() {
			break
		}
		statemets = append(statemets, p.declaration())
	}
	return statemets
}

func (p *Parser) declaration() stm.Statement {

	defer p.handlePanic()

	if p.match(tokens.VAR) {
		return p.varDeclaration()
	}

	return p.statement()
}

func (p *Parser) varDeclaration() stm.Statement {
	name, err := p.consume(tokens.IDENTIFIER, "Expect variable name.")

	if err != nil {
		return stm.NewError("Invalid statement")
	}

	var initializer expressions.Expression = nil

	if p.match(tokens.EQUAL) {
		initializer = p.expression()
	}
	p.consume(tokens.SEMICOLON, "Expect ';' after variable declaration.")

	return stm.NewVar(*name, initializer)

}

func (p *Parser) statement() stm.Statement {
	if p.match(tokens.FOR) {
		return p.forStatement()
	}

	if p.match(tokens.IF) {
		return p.ifStatement()
	}

	if p.match(tokens.WHILE) {
		return p.whileStatement()
	}

	if p.match(tokens.PRINT) {
		return p.printStatement()
	}
	if p.match((tokens.LEFT_BRACE)) {
		return stm.NewBlock(p.block())
	}
	return p.expressionStatement()
}

func (p *Parser) forStatement() stm.Statement {
	p.consume(tokens.LEFT_PAREN, "Expect '(' after 'for'.")

	var initializer stm.Statement
	var condition expressions.Expression = nil
	var increment expressions.Expression = nil

	if p.match(tokens.SEMICOLON) {
		initializer = nil
	} else if p.match(tokens.VAR) {
		initializer = p.varDeclaration()
	} else {
		initializer = p.expressionStatement()
	}

	if !p.check(tokens.SEMICOLON) {
		condition = p.expression()
	}

	p.consume(tokens.SEMICOLON, "Expect ';' after loop condition.")

	if !p.check(tokens.RIGHT_PAREN) {
		increment = p.expression()
	}

	p.consume(tokens.RIGHT_PAREN, "Expect ')' after for clauses.")

	body := p.statement()

	if increment != nil {
		body = stm.NewBlock([]stm.Statement{body, stm.NewExpression(increment)})
	}

	if condition == nil {
		condition = expressions.NewLiteral(true)
	}

	body = stm.NewWhile(condition, body)

	if initializer != nil {
		body = stm.NewBlock([]stm.Statement{initializer, body})
	}

	return body

}

func (p *Parser) whileStatement() stm.WhileStmt {
	p.consume(tokens.LEFT_PAREN, "Expect '(' after 'while'.")
	condition := p.expression()
	p.consume(tokens.RIGHT_PAREN, "Expect ')' after condition.")
	body := p.statement()

	return *stm.NewWhile(condition, body)
}

func (p *Parser) ifStatement() stm.IfStmt {
	p.consume(tokens.LEFT_PAREN, "Expect '(' after 'if'.")
	condition := p.expression()
	p.consume(tokens.RIGHT_PAREN, "Expect ')' after if condition.")

	thenBranch := p.statement()
	var elseBranch stm.Statement = nil

	if p.match(tokens.ELSE) {
		elseBranch = p.statement()
	}

	return *stm.NewIf(condition, thenBranch, elseBranch)
}

func (p *Parser) printStatement() stm.PrintStmt {
	value := p.expression()
	p.consume(tokens.SEMICOLON, "Expect ';' after value.")
	return *stm.NewPrint(value)
}

func (p *Parser) expressionStatement() stm.ExpressionStmt {
	expr := p.expression()
	p.consume(tokens.SEMICOLON, "Expect ';' after value.")

	return *stm.NewExpression(expr)

}

func (p *Parser) block() []stm.Statement {
	statements := make([]stm.Statement, 0)

	for {
		if p.check(tokens.RIGHT_BRACE) || p.isAtEnd() {
			break
		}

		statements = append(statements, p.declaration())
	}

	p.consume(tokens.RIGHT_BRACE, "Expect '}' after block.")

	return statements
}

func (p *Parser) expression() expressions.Expression {
	return p.assignemt()
}

func (p *Parser) assignemt() expressions.Expression {
	expr := p.ternary()

	if p.match(tokens.EQUAL) {
		equals := p.previous()
		value := p.assignemt()

		if v, ok := expr.(*expressions.Variable); ok {
			name := v.Name

			return expressions.NewAssign(name, value)
		}
		p.errorLogger.ErrorForToken(equals, "Invalid assignment target.")
	}
	return expr

}

func (p *Parser) ternary() expressions.Expression {
	expression := p.or()

	for {
		if !p.match(tokens.QUESTION_MARK) {
			break
		}
		operator := p.previous()
		consequent := p.ternary()
		_, err := p.consume(tokens.COLON, "expected alternative expression in ternary\n")

		if err != nil {
			return expressions.NewError("Invalid ternary expression\n")
		}

		alternative := p.ternary()

		expression = expressions.NewTernary(operator, expression, consequent, alternative)
	}

	return expression
}

func (p *Parser) or() expressions.Expression {
	expr := p.and()

	for {
		if !p.match(tokens.OR) {
			break
		}
		operator := p.previous()
		right := p.and()

		expr = expressions.NewLogical(expr, operator, right)
	}
	return expr
}

func (p *Parser) and() expressions.Expression {
	expr := p.equality()

	for {
		if !p.match(tokens.AND) {
			break
		}

		operator := p.previous()
		right := p.equality()

		expr = expressions.NewLogical(expr, operator, right)
	}

	return expr
}

func (p *Parser) equality() expressions.Expression {
	expression := p.comparison()

	for {
		if !p.match(tokens.BANG_EQUAL, tokens.EQUAL_EQUAL) {
			break
		}

		operator := p.previous()
		right := p.comparison()
		expression = expressions.NewBinary(expression, operator, right)

	}

	return expression
}

func (p *Parser) comparison() expressions.Expression {
	expr := p.term()

	for {
		if !p.match(tokens.GREATER, tokens.GREATER_EQUAL, tokens.LESS, tokens.LESS_EQUAL) {
			break
		}

		operator := p.previous()
		right := p.term()
		expr = expressions.NewBinary(expr, operator, right)

	}
	return expr
}

func (p *Parser) term() expressions.Expression {
	expr := p.factor()

	for {
		if !p.match(tokens.PLUS, tokens.MINUS) {
			break
		}

		operator := p.previous()
		right := p.factor()
		expr = expressions.NewBinary(expr, operator, right)

	}
	return expr
}

func (p *Parser) factor() expressions.Expression {
	expr := p.unary()

	for {
		if !p.match(tokens.STAR, tokens.SLASH) {
			break
		}

		operator := p.previous()
		right := p.unary()
		expr = expressions.NewBinary(expr, operator, right)
	}
	return expr
}

func (p *Parser) unary() expressions.Expression {
	expr := p.primary()

	for {
		if !p.match(tokens.MINUS, tokens.BANG) {
			break
		}
		operator := p.previous()
		right := p.primary()
		expr = expressions.NewUnary(operator, right)
	}

	return expr
}

func (p *Parser) primary() expressions.Expression {
	if p.match(tokens.TRUE) {
		return expressions.NewLiteral(true)
	}

	if p.match(tokens.FALSE) {
		return expressions.NewLiteral(false)
	}

	if p.match(tokens.NIL) {
		return expressions.NewLiteral(nil)
	}

	if p.match(tokens.NUMBER, tokens.STRING) {
		return expressions.NewLiteral(p.previous().Literal)
	}

	if p.match(tokens.IDENTIFIER) {
		return expressions.NewVariable(p.previous())
	}

	if p.match(tokens.LEFT_PAREN) {
		expr := p.expression()
		_, err := p.consume(tokens.RIGHT_PAREN, "Expect ')' after expression.\n")

		if err == nil {
			return expressions.NewGrouping(expr)
		}
	}

	errorMessage := fmt.Sprintf("Error during parsing: unexpected character: %s", p.peek().Lexeme)
	return expressions.NewError(errorMessage)

}

func (p *Parser) match(tokenTypes ...tokens.TokenType) bool {
	for _, token := range tokenTypes {
		if p.check(token) {
			p.advance()
			return true
		}
	}
	return false
}

func (p *Parser) advance() tokens.Token {
	if !p.isAtEnd() {
		p.current++
	}
	return p.previous()
}

func (p *Parser) check(tokenType tokens.TokenType) bool {
	if p.isAtEnd() {
		return false
	}
	return p.peek().TokenType == tokenType
}

func (p *Parser) isAtEnd() bool {
	return p.peek().TokenType == tokens.EOF
}

func (p *Parser) peek() tokens.Token {
	return p.tokens[p.current]
}

func (p *Parser) previous() tokens.Token {
	return p.tokens[p.current-1]
}

func (p *Parser) consume(tokenType tokens.TokenType, errorMessage string) (*tokens.Token, error) {
	if p.check(tokenType) {
		token := p.advance()
		return &token, nil
	}
	err := p.errorLogger.ErrorForToken(p.peek(), errorMessage)
	done := make(chan bool)
	p.errorHandler <- ErrorHandlerEvent{Err: err, Done: done}
	<-done
	return nil, errors.New("parsing error")
}

func (p *Parser) synchronize() {
	p.advance()

	for {
		if p.isAtEnd() || p.previous().TokenType == tokens.SEMICOLON {
			return
		}
		switch p.peek().TokenType {
		case tokens.CLASS, tokens.FUN, tokens.VAR, tokens.FOR, tokens.IF, tokens.WHILE, tokens.PRINT, tokens.RETURN:
			return
		}
		p.advance()
	}
}
