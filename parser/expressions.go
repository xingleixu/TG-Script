package parser

import (
	"github.com/xingleixu/TG-Script/ast"
	"github.com/xingleixu/TG-Script/lexer"
)

// ============================================================================
// EXPRESSION PARSING
// ============================================================================

// Pratt parser function types
type (
	prefixParseFn func() ast.Expression
	infixParseFn  func(ast.Expression) ast.Expression
)

// prefixParseFns maps token types to their prefix parse functions.
var prefixParseFns = map[lexer.Token]prefixParseFn{}

// infixParseFns maps token types to their infix parse functions.
var infixParseFns = map[lexer.Token]infixParseFn{}

// registerPrefix registers a prefix parse function for a token type.
func (p *Parser) registerPrefix(tokenType lexer.Token, fn prefixParseFn) {
	prefixParseFns[tokenType] = fn
}

// registerInfix registers an infix parse function for a token type.
func (p *Parser) registerInfix(tokenType lexer.Token, fn infixParseFn) {
	infixParseFns[tokenType] = fn
}

// parseExpression parses an expression using Pratt parsing.
func (p *Parser) parseExpression(precedence Precedence) ast.Expression {
	prefix := prefixParseFns[p.currentToken.Type]
	if prefix == nil {
		p.addErrorf("no prefix parse function for %s found", p.currentToken.Type)
		return nil
	}

	leftExp := prefix()

	for !p.peekTokenIs(lexer.SEMICOLON) && precedence < p.peekPrecedence() {
		infix := infixParseFns[p.peekToken.Type]
		if infix == nil {
			return leftExp
		}

		p.nextToken()
		leftExp = infix(leftExp)
	}

	return leftExp
}

// ============================================================================
// PREFIX EXPRESSIONS
// ============================================================================

// parseIdentifierExpression parses an identifier expression.
func (p *Parser) parseIdentifierExpression() ast.Expression {
	return &ast.Identifier{
		NamePos: p.currentToken.Position,
		Name:    p.currentToken.Literal,
	}
}

// parseIntegerLiteralExpression parses an integer literal expression.
func (p *Parser) parseIntegerLiteralExpression() ast.Expression {
	return p.parseIntegerLiteral()
}

// parseFloatLiteralExpression parses a float literal expression.
func (p *Parser) parseFloatLiteralExpression() ast.Expression {
	return p.parseFloatLiteral()
}

// parseStringLiteralExpression parses a string literal expression.
func (p *Parser) parseStringLiteralExpression() ast.Expression {
	return p.parseStringLiteral()
}

// parseBooleanLiteralExpression parses a boolean literal expression.
func (p *Parser) parseBooleanLiteralExpression() ast.Expression {
	return p.parseBooleanLiteral()
}

// parseNullLiteralExpression parses a null literal expression.
func (p *Parser) parseNullLiteralExpression() ast.Expression {
	return p.parseNullLiteral()
}

// parseUndefinedLiteralExpression parses an undefined literal expression.
func (p *Parser) parseUndefinedLiteralExpression() ast.Expression {
	return p.parseUndefinedLiteral()
}

// parseVoidLiteralExpression parses a void literal expression.
func (p *Parser) parseVoidLiteralExpression() ast.Expression {
	return p.parseVoidLiteral()
}

// parsePrefixExpression parses a prefix expression (unary operators).
func (p *Parser) parsePrefixExpression() ast.Expression {
	expression := &ast.UnaryExpression{
		OpPos:    p.currentToken.Position,
		Operator: lexer.Token(p.currentToken.Type),
	}

	p.nextToken()
	expression.Operand = p.parseExpression(UNARY)

	return expression
}

// parseGroupedExpression parses a grouped expression (parentheses).
func (p *Parser) parseGroupedExpression() ast.Expression {
	lparenPos := p.currentToken.Position
	p.nextToken()

	// Handle empty parentheses for arrow functions: () => expr
	if p.currentTokenIs(lexer.RPAREN) {
		rparenPos := p.currentToken.Position
		// Check if next token is '=>' to confirm this is arrow function params
		if p.peekTokenIs(lexer.ARROW) {
			return &ast.ArrowFunctionParams{
				LParen:     lparenPos,
				Parameters: []*ast.Parameter{}, // empty parameter list
				RParen:     rparenPos,
			}
		}
		// If not arrow function, this is an error - empty parentheses without arrow
		p.addErrorf("unexpected token %s", p.currentToken.Type)
		return nil
	}

	// Try to parse as arrow function parameters first
	if p.mightBeArrowFunctionParams() {
		// Save current position for potential backtracking
		savedCurrentToken := p.currentToken
		savedPeekToken := p.peekToken
		savedErrors := len(p.errors)

		// Try to parse as arrow function parameters
		params := p.parseArrowFunctionParameterList()
		if params != nil && p.expectPeek(lexer.RPAREN) {
			rparenPos := p.currentToken.Position
			// Check if next token is '=>' to confirm this is arrow function params
			if p.peekTokenIs(lexer.ARROW) {
				return &ast.ArrowFunctionParams{
					LParen:     lparenPos,
					Parameters: params,
					RParen:     rparenPos,
				}
			}
		}

		// If not arrow function params, restore state and parse as regular expression
		p.currentToken = savedCurrentToken
		p.peekToken = savedPeekToken
		// Remove any errors added during failed arrow function parsing
		p.errors = p.errors[:savedErrors]
	}

	// Parse as regular grouped expression
	exp := p.parseExpression(LOWEST)

	if !p.expectPeek(lexer.RPAREN) {
		return nil
	}

	return exp
}

// parseArrayLiteral parses an array literal.
func (p *Parser) parseArrayLiteral() ast.Expression {
	array := &ast.ArrayLiteral{
		LBracket: p.currentToken.Position,
	}

	array.Elements = p.parseExpressionList(lexer.RBRACKET)

	if p.currentTokenIs(lexer.RBRACKET) {
		array.RBracket = p.currentToken.Position
	}

	return array
}

// parseObjectLiteral parses an object literal.
func (p *Parser) parseObjectLiteral() ast.Expression {
	obj := &ast.ObjectLiteral{
		LBrace: p.currentToken.Position,
	}

	if p.peekTokenIs(lexer.RBRACE) {
		p.nextToken()
		obj.RBrace = p.currentToken.Position
		return obj
	}

	p.nextToken()

	for {
		prop := p.parseObjectProperty()
		if prop != nil {
			obj.Properties = append(obj.Properties, prop)
		}

		if !p.peekTokenIs(lexer.COMMA) {
			break
		}
		p.nextToken()
		p.nextToken()
	}

	if !p.expectPeek(lexer.RBRACE) {
		return nil
	}

	obj.RBrace = p.currentToken.Position
	return obj
}

// parseObjectProperty parses an object property.
func (p *Parser) parseObjectProperty() *ast.Property {
	prop := &ast.Property{}

	// Parse key
	switch p.currentToken.Type {
	case lexer.IDENT:
		prop.Key = p.parseIdentifierExpression()
	case lexer.STRING:
		prop.Key = p.parseStringLiteralExpression()
	case lexer.INT:
		prop.Key = p.parseIntegerLiteralExpression()
	case lexer.LBRACKET:
		// Computed property name
		prop.Computed = true
		p.nextToken()
		prop.Key = p.parseExpression(LOWEST)
		if !p.expectPeek(lexer.RBRACKET) {
			return nil
		}
	default:
		p.addErrorf("expected property key, got %s", p.currentToken.Type)
		return nil
	}

	if !p.expectPeek(lexer.COLON) {
		return nil
	}

	p.nextToken()
	prop.Value = p.parseExpression(LOWEST)

	return prop
}

// parseFunctionExpression parses a function expression.
func (p *Parser) parseFunctionExpression() ast.Expression {
	fn := &ast.FunctionExpression{
		FunctionPos: p.currentToken.Position,
	}

	// Check for async
	if p.currentTokenIs(lexer.ASYNC) {
		fn.Async = true
		if !p.expectPeek(lexer.FUNCTION) {
			return nil
		}
	}

	// Check for generator
	if p.peekTokenIs(lexer.MUL) {
		p.nextToken()
		fn.Generator = true
	}

	// Optional function name
	if p.peekTokenIs(lexer.IDENT) {
		p.nextToken()
		fn.Name = p.parseIdentifier()
	}

	if !p.expectPeek(lexer.LPAREN) {
		return nil
	}

	fn.LParen = p.currentToken.Position
	fn.Parameters = p.parseParameterList()

	if p.currentTokenIs(lexer.RPAREN) {
		fn.RParen = p.currentToken.Position
	}

	// Optional return type annotation
	if p.peekTokenIs(lexer.COLON) {
		p.nextToken()
		p.nextToken()
		fn.ReturnType = p.parseTypeAnnotation()
	}

	if !p.expectPeek(lexer.LBRACE) {
		return nil
	}

	fn.Body = p.parseBlockStatement()

	return fn
}

// ============================================================================
// INFIX EXPRESSIONS
// ============================================================================

// parseInfixExpression parses an infix expression (binary operators).
func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	expression := &ast.BinaryExpression{
		Left:     left,
		OpPos:    p.currentToken.Position,
		Operator: lexer.Token(p.currentToken.Type),
	}

	precedence := p.currentPrecedence()
	p.nextToken()
	expression.Right = p.parseExpression(precedence)

	return expression
}

// parseAssignmentExpression parses an assignment expression.
func (p *Parser) parseAssignmentExpression(left ast.Expression) ast.Expression {
	expression := &ast.AssignmentExpression{
		Left:     left,
		OpPos:    p.currentToken.Position,
		Operator: lexer.Token(p.currentToken.Type),
	}

	p.nextToken()
	expression.Right = p.parseExpression(LOWEST)

	return expression
}

// parseCallExpression parses a call expression.
func (p *Parser) parseCallExpression(fn ast.Expression) ast.Expression {
	exp := &ast.CallExpression{
		Callee: fn,
		LParen: p.currentToken.Position,
	}

	exp.Arguments = p.parseExpressionList(lexer.RPAREN)

	if p.currentTokenIs(lexer.RPAREN) {
		exp.RParen = p.currentToken.Position
	}

	return exp
}

// parseMemberExpression parses a member expression (dot notation).
func (p *Parser) parseMemberExpression(object ast.Expression) ast.Expression {
	exp := &ast.MemberExpression{
		Object: object,
		Dot:    p.currentToken.Position,
	}

	if !p.expectPeek(lexer.IDENT) {
		return nil
	}

	exp.Property = p.parseIdentifierExpression()
	return exp
}

// parseIndexExpression parses an index expression (bracket notation).
func (p *Parser) parseIndexExpression(left ast.Expression) ast.Expression {
	exp := &ast.MemberExpression{
		Object:    left,
		LBracket:  p.currentToken.Position,
		Computed:  true,
	}

	p.nextToken()
	exp.Property = p.parseExpression(LOWEST)

	if !p.expectPeek(lexer.RBRACKET) {
		return nil
	}

	exp.RBracket = p.currentToken.Position
	return exp
}

// parseTernaryExpression parses a ternary conditional expression.
func (p *Parser) parseTernaryExpression(condition ast.Expression) ast.Expression {
	exp := &ast.ConditionalExpression{
		Test:     condition,
		Question: p.currentToken.Position,
	}

	p.nextToken()
	exp.Consequent = p.parseExpression(LOWEST)

	if !p.expectPeek(lexer.COLON) {
		return nil
	}

	exp.Colon = p.currentToken.Position
	p.nextToken()
	exp.Alternate = p.parseExpression(TERNARY)

	return exp
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

// parseExpressionList parses a comma-separated list of expressions.
func (p *Parser) parseExpressionList(end lexer.Token) []ast.Expression {
	var args []ast.Expression

	if p.peekTokenIs(end) {
		p.nextToken()
		return args
	}

	p.nextToken()
	args = append(args, p.parseExpression(LOWEST))

	for p.peekTokenIs(lexer.COMMA) {
		p.nextToken()
		p.nextToken()
		args = append(args, p.parseExpression(LOWEST))
	}

	if !p.expectPeek(end) {
		return nil
	}

	return args
}

// parseParameterList parses a function parameter list.
func (p *Parser) parseParameterList() []*ast.Parameter {
	var params []*ast.Parameter

	if p.peekTokenIs(lexer.RPAREN) {
		p.nextToken()
		return params
	}

	p.nextToken()

	param := p.parseParameter()
	if param != nil {
		params = append(params, param)
	}

	for p.peekTokenIs(lexer.COMMA) {
		p.nextToken()
		p.nextToken()
		param := p.parseParameter()
		if param != nil {
			params = append(params, param)
		}
	}

	if !p.expectPeek(lexer.RPAREN) {
		return nil
	}

	return params
}

// parseParameter parses a function parameter.
func (p *Parser) parseParameter() *ast.Parameter {
	param := &ast.Parameter{}

	// Check for rest parameter
	if p.currentTokenIs(lexer.SPREAD) {
		param.Rest = true
		p.nextToken()
	}

	if !p.currentTokenIs(lexer.IDENT) {
		p.addErrorf("expected parameter name, got %s", p.currentToken.Type)
		return nil
	}

	param.Name = p.parseIdentifier()

	// Optional type annotation
	if p.peekTokenIs(lexer.COLON) {
		p.nextToken()
		p.nextToken()
		param.TypeAnnotation = p.parseTypeAnnotation()
	}

	// Optional default value
	if p.peekTokenIs(lexer.ASSIGN) {
		p.nextToken()
		p.nextToken()
		param.DefaultValue = p.parseExpression(LOWEST)
	}

	return param
}

// parseTypeofExpression parses typeof expressions
func (p *Parser) parseTypeofExpression() ast.Expression {
	expression := &ast.UnaryExpression{
		OpPos:    p.currentToken.Position,
		Operator: p.currentToken.Type,
	}

	p.nextToken()
	expression.Operand = p.parseExpression(UNARY)

	return expression
}

// parseDeleteExpression parses delete expressions
func (p *Parser) parseDeleteExpression() ast.Expression {
	expression := &ast.UnaryExpression{
		OpPos:    p.currentToken.Position,
		Operator: p.currentToken.Type,
	}

	p.nextToken()
	expression.Operand = p.parseExpression(UNARY)

	return expression
}

// parseNewExpression parses new expressions
func (p *Parser) parseNewExpression() ast.Expression {
	expression := &ast.UnaryExpression{
		OpPos:    p.currentToken.Position,
		Operator: p.currentToken.Type,
	}

	p.nextToken()
	expression.Operand = p.parseExpression(UNARY)

	return expression
}

// parseThisExpression parses this keyword
func (p *Parser) parseThisExpression() ast.Expression {
	return &ast.Identifier{
		NamePos: p.currentToken.Position,
		Name:    p.currentToken.Literal,
	}
}

// parseSuperExpression parses super keyword
func (p *Parser) parseSuperExpression() ast.Expression {
	return &ast.Identifier{
		NamePos: p.currentToken.Position,
		Name:    p.currentToken.Literal,
	}
}

// parseAwaitExpression parses await expressions
func (p *Parser) parseAwaitExpression() ast.Expression {
	expression := &ast.UnaryExpression{
		OpPos:    p.currentToken.Position,
		Operator: p.currentToken.Type,
	}

	p.nextToken()
	expression.Operand = p.parseExpression(UNARY)

	return expression
}

// parseYieldExpression parses yield expressions
func (p *Parser) parseYieldExpression() ast.Expression {
	expression := &ast.UnaryExpression{
		OpPos:    p.currentToken.Position,
		Operator: p.currentToken.Type,
	}

	// yield can be used without an expression
	if p.peekTokenIs(lexer.SEMICOLON) || p.peekTokenIs(lexer.RBRACE) || p.peekTokenIs(lexer.EOF) {
		return expression
	}

	p.nextToken()
	expression.Operand = p.parseExpression(UNARY)

	return expression
}

// parseIncrementExpression parses prefix ++ expressions
func (p *Parser) parseIncrementExpression() ast.Expression {
	expression := &ast.UnaryExpression{
		OpPos:    p.currentToken.Position,
		Operator: p.currentToken.Type,
	}

	p.nextToken()
	expression.Operand = p.parseExpression(UNARY)

	return expression
}

// parseDecrementExpression parses prefix -- expressions
func (p *Parser) parseDecrementExpression() ast.Expression {
	expression := &ast.UnaryExpression{
		OpPos:    p.currentToken.Position,
		Operator: p.currentToken.Type,
	}

	p.nextToken()
	expression.Operand = p.parseExpression(UNARY)

	return expression
}

// parseInstanceofExpression parses instanceof expressions
func (p *Parser) parseInstanceofExpression(left ast.Expression) ast.Expression {
	expression := &ast.BinaryExpression{
		Left:     left,
		OpPos:    p.currentToken.Position,
		Operator: p.currentToken.Type,
	}

	precedence := p.currentPrecedence()
	p.nextToken()
	expression.Right = p.parseExpression(precedence)

	return expression
}

// parseInExpression parses in expressions
func (p *Parser) parseInExpression(left ast.Expression) ast.Expression {
	expression := &ast.BinaryExpression{
		Left:     left,
		OpPos:    p.currentToken.Position,
		Operator: p.currentToken.Type,
	}

	precedence := p.currentPrecedence()
	p.nextToken()
	expression.Right = p.parseExpression(precedence)

	return expression
}

// parseNullishCoalescingExpression parses ?? expressions
func (p *Parser) parseNullishCoalescingExpression(left ast.Expression) ast.Expression {
	expression := &ast.BinaryExpression{
		Left:     left,
		OpPos:    p.currentToken.Position,
		Operator: p.currentToken.Type,
	}

	precedence := p.currentPrecedence()
	p.nextToken()
	expression.Right = p.parseExpression(precedence)

	return expression
}

// parseOptionalChainingExpression parses ?. expressions
func (p *Parser) parseOptionalChainingExpression(left ast.Expression) ast.Expression {
	expression := &ast.MemberExpression{
		Object:   left,
		Computed: false,
		Dot:      p.currentToken.Position,
	}

	p.nextToken()
	expression.Property = p.parseExpression(MEMBER)

	return expression
}

// parsePostfixIncrementExpression parses postfix ++ expressions
func (p *Parser) parsePostfixIncrementExpression(left ast.Expression) ast.Expression {
	return &ast.UnaryExpression{
		OpPos:    p.currentToken.Position,
		Operator: p.currentToken.Type,
		Operand:  left,
		Postfix:  true,
	}
}

// parsePostfixDecrementExpression parses postfix -- expressions
func (p *Parser) parsePostfixDecrementExpression(left ast.Expression) ast.Expression {
	return &ast.UnaryExpression{
		OpPos:    p.currentToken.Position,
		Operator: p.currentToken.Type,
		Operand:  left,
		Postfix:  true,
	}
}



// init registers all prefix and infix parse functions.

// parseArrowFunctionExpression parses arrow function expressions
// mightBeArrowFunctionParams checks if the current position might be arrow function parameters
func (p *Parser) mightBeArrowFunctionParams() bool {
	// Simple heuristic: if we see an identifier followed by ':' or ',' or ')', it might be parameters
	if p.currentTokenIs(lexer.IDENT) {
		return p.peekTokenIs(lexer.COLON) || p.peekTokenIs(lexer.COMMA) || p.peekTokenIs(lexer.RPAREN)
	}
	// Empty parameter list
	if p.currentTokenIs(lexer.RPAREN) {
		return true
	}
	return false
}

// parseArrowFunctionParameterList parses a list of arrow function parameters
func (p *Parser) parseArrowFunctionParameterList() []*ast.Parameter {
	var params []*ast.Parameter
	
	// Handle empty parameter list
	if p.currentTokenIs(lexer.RPAREN) {
		return params
	}
	
	// Parse first parameter
	param := p.parseParameter()
	if param == nil {
		return nil
	}
	params = append(params, param)
	
	// Parse remaining parameters
	for p.peekTokenIs(lexer.COMMA) {
		p.nextToken() // consume ','
		p.nextToken() // move to next parameter
		
		param := p.parseParameter()
		if param == nil {
			return nil
		}
		params = append(params, param)
	}
	
	return params
}

func (p *Parser) parseArrowFunctionExpression(left ast.Expression) ast.Expression {
	arrow := &ast.ArrowFunctionExpression{
		Arrow: p.currentToken.Position,
	}

	// Handle different parameter forms
	switch leftExpr := left.(type) {
	case *ast.Identifier:
		// Single parameter without parentheses: x => x * 2
		param := &ast.Parameter{
			Name: leftExpr,
		}
		arrow.Parameters = []*ast.Parameter{param}
	case *ast.ArrowFunctionParams:
		// Parameters in parentheses: (x: int, y: int) => x + y
		arrow.Parameters = leftExpr.Parameters
		arrow.LParen = leftExpr.LParen
		arrow.RParen = leftExpr.RParen
	default:
		// For now, we don't support other complex parameter forms
		p.addErrorf("arrow function currently only supports identifier or parenthesized parameters: %T", left)
		return nil
	}

	// Parse function body
	p.nextToken() // move past '=>'
	
	// Arrow function body can be either an expression or a block statement
	if p.currentTokenIs(lexer.LBRACE) {
		arrow.Body = p.parseBlockStatement()
	} else {
		// Expression body - wrap in a return statement
		expr := p.parseExpression(LOWEST)
		if expr != nil {
			returnStmt := &ast.ReturnStatement{
				ReturnPos: p.currentToken.Position,
				Argument:  expr,
			}
			arrow.Body = &ast.BlockStatement{
				LBrace: p.currentToken.Position,
				Body:   []ast.Statement{returnStmt},
				RBrace: p.currentToken.Position,
			}
		}
	}

	return arrow
}