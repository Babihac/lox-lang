package parser

import (
	"errors"
	"fmt"
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
	// if r := recover(); r != nil {
	// 	p.errorLogger.Error(p.peek().Line, "ajajaj")
	// }
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

	var initializer stm.Expression = nil

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

	if p.match(tokens.BREAK) {
		return p.breakStatement()
	}

	if p.match(tokens.FUN) {
		return p.functionStatement("function")
	}

	if p.match(tokens.RETURN) {
		return p.returnStatement()
	}

	return p.expressionStatement()
}

func (p *Parser) forStatement() stm.Statement {
	p.consume(tokens.LEFT_PAREN, "Expect '(' after 'for'.")

	var initializer stm.Statement
	var condition stm.Expression = nil
	var increment stm.Expression = nil

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
		condition = stm.NewLiteral(true)
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

func (p *Parser) breakStatement() stm.BreakStmt {
	p.consume(tokens.SEMICOLON, "Expect ';' after break.\n")

	return *stm.NewBreak()
}

func (p *Parser) functionStatement(kind string) stm.Statement {

	name, err := p.consume(tokens.IDENTIFIER, fmt.Sprintf("Expect %s name\n", kind))

	if err != nil {
		return stm.NewError("error creating function statement")
	}

	functionComponents := p.parseFunctionComponents(kind)

	return *stm.NewFunction(*name, functionComponents.parameters, functionComponents.body)
}

func (p *Parser) returnStatement() stm.ReturnStmt {
	keyword := p.previous()
	var value stm.Expression = nil

	if !p.check(tokens.SEMICOLON) {
		value = p.expression()
	}

	p.consume(tokens.SEMICOLON, "Expect ';' after return value.")

	return *stm.NewReturn(keyword, value)
}

func (p *Parser) printStatement() stm.PrintStmt {
	value := p.expression()
	p.consume(tokens.SEMICOLON, "Expect ';' after value of print.\n")
	return *stm.NewPrint(value)
}

func (p *Parser) expressionStatement() stm.ExpressionStmt {
	expr := p.expression()
	p.consume(tokens.SEMICOLON, "Expect ';' after value.\n")

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

func (p *Parser) expression() stm.Expression {
	return p.assignemt()
}

func (p *Parser) assignemt() stm.Expression {
	expr := p.ternary()

	if p.match(tokens.EQUAL) {
		equals := p.previous()
		value := p.assignemt()

		if v, ok := expr.(*stm.Variable); ok {
			name := v.Name

			return stm.NewAssign(name, value)
		}
		p.errorLogger.ErrorForToken(equals, "Invalid assignment target.")
	}
	return expr

}

func (p *Parser) ternary() stm.Expression {
	expression := p.or()

	for {
		if !p.match(tokens.QUESTION_MARK) {
			break
		}
		operator := p.previous()
		consequent := p.ternary()
		_, err := p.consume(tokens.COLON, "expected alternative expression in ternary\n")

		if err != nil {
			return stm.NewErrorExpr("Invalid ternary expression\n")
		}

		alternative := p.ternary()

		expression = stm.NewTernary(operator, expression, consequent, alternative)
	}

	return expression
}

func (p *Parser) or() stm.Expression {
	expr := p.and()

	for {
		if !p.match(tokens.OR) {
			break
		}
		operator := p.previous()
		right := p.and()

		expr = stm.NewLogical(expr, operator, right)
	}
	return expr
}

func (p *Parser) and() stm.Expression {
	expr := p.equality()

	for {
		if !p.match(tokens.AND) {
			break
		}

		operator := p.previous()
		right := p.equality()

		expr = stm.NewLogical(expr, operator, right)
	}

	return expr
}

func (p *Parser) equality() stm.Expression {
	expression := p.comparison()

	for {
		if !p.match(tokens.BANG_EQUAL, tokens.EQUAL_EQUAL) {
			break
		}

		operator := p.previous()
		right := p.comparison()
		expression = stm.NewBinary(expression, operator, right)

	}

	return expression
}

func (p *Parser) comparison() stm.Expression {
	expr := p.term()

	for {
		if !p.match(tokens.GREATER, tokens.GREATER_EQUAL, tokens.LESS, tokens.LESS_EQUAL) {
			break
		}

		operator := p.previous()
		right := p.term()
		expr = stm.NewBinary(expr, operator, right)

	}
	return expr
}

func (p *Parser) term() stm.Expression {
	expr := p.factor()

	for {
		if !p.match(tokens.PLUS, tokens.MINUS) {
			break
		}

		operator := p.previous()
		right := p.factor()
		expr = stm.NewBinary(expr, operator, right)

	}
	return expr
}

func (p *Parser) factor() stm.Expression {
	expr := p.unary()

	for {
		if !p.match(tokens.STAR, tokens.SLASH) {
			break
		}

		operator := p.previous()
		right := p.unary()
		expr = stm.NewBinary(expr, operator, right)
	}
	return expr
}

func (p *Parser) unary() stm.Expression {
	expr := p.call()

	for {
		if !p.match(tokens.MINUS, tokens.BANG) {
			break
		}
		operator := p.previous()
		right := p.primary()
		expr = stm.NewUnary(operator, right)
	}

	return expr
}

func (p *Parser) call() stm.Expression {
	expr := p.anonymousFunction()

	for {
		if p.match(tokens.LEFT_PAREN) {
			expr = p.finishCall(expr)
		} else {
			break
		}
	}

	return expr
}

func (p *Parser) anonymousFunction() stm.Expression {
	expr := p.primary()

	if p.match(tokens.FUN) {

		functionComponents := p.parseFunctionComponents("anonymous function")

		return *stm.NewAnonymousFunction(functionComponents.parameters, functionComponents.body)
	}

	return expr
}

func (p *Parser) finishCall(expr stm.Expression) stm.Expression {
	arguments := make([]stm.Expression, 0)

	if !p.check(tokens.RIGHT_PAREN) {
		for {
			arguments = append(arguments, p.expression())

			if len(arguments) > 255 {
				p.errorLogger.ErrorForToken(p.peek(), "Can't have more than 255 arguments.")
			}

			if !p.match(tokens.COMMA) {
				break
			}
		}
	}
	paren, err := p.consume(tokens.RIGHT_PAREN, "Expect ')' after arguments.")

	if err != nil {
		return stm.NewErrorExpr("Error calling function")
	}

	return stm.NewCall(expr, *paren, arguments)

}

func (p *Parser) primary() stm.Expression {
	if p.match(tokens.TRUE) {
		return stm.NewLiteral(true)
	}

	if p.match(tokens.FALSE) {
		return stm.NewLiteral(false)
	}

	if p.match(tokens.NIL) {
		return stm.NewLiteral(nil)
	}

	if p.match(tokens.NUMBER, tokens.STRING) {
		return stm.NewLiteral(p.previous().Literal)
	}

	if p.match(tokens.IDENTIFIER) {
		return stm.NewVariable(p.previous())
	}

	if p.match(tokens.LEFT_PAREN) {
		expr := p.expression()
		_, err := p.consume(tokens.RIGHT_PAREN, "Expect ')' after expression.\n")

		if err == nil {
			return stm.NewGrouping(expr)
		}
	}

	errorMessage := fmt.Sprintf("Error during parsing: unexpected character: %s", p.peek().Lexeme)
	return stm.NewErrorExpr(errorMessage)

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
	err := p.errorLogger.ErrorForToken(p.previous(), errorMessage)
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

func (p *Parser) parseFunctionComponents(kind string) *FunctionComponents {
	parameters := make([]tokens.Token, 0)
	p.consume(tokens.LEFT_PAREN, fmt.Sprintf("Expect ( after %s name \n", kind))

	if !p.check(tokens.RIGHT_PAREN) {
		for {
			param, err := p.consume(tokens.IDENTIFIER, "Expect parameter name.\n")

			if err != nil {
				return nil
			}

			parameters = append(parameters, *param)

			if len(parameters) > 255 {
				p.errorLogger.ErrorForToken(p.peek(), "Can't have more than 255 arguments.\n")
			}

			if !p.match(tokens.COMMA) {
				break
			}
		}
	}
	p.consume(tokens.RIGHT_PAREN, "Expect ')' after arguments.")
	p.consume(tokens.LEFT_BRACE, fmt.Sprintf("Expect { before %s body\n", kind))
	body := p.block()

	return NewFunctionComponents(parameters, body)
}
