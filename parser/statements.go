package parser

import (
	"github.com/xingleixu/TG-Script/ast"
	"github.com/xingleixu/TG-Script/lexer"
)

// ============================================================================
// STATEMENT PARSING
// ============================================================================

// parseExpressionStatement parses an expression statement.
func (p *Parser) parseExpressionStatement() ast.Statement {
	stmt := &ast.ExpressionStatement{
		Expression: p.parseExpression(LOWEST),
	}

	// Use ASI logic for optional semicolon
	if p.peekTokenIs(lexer.SEMICOLON) {
		p.nextToken()
		stmt.Semicolon = p.currentToken.Position
	} else if !p.canInsertSemicolon() {
		// Only report error if ASI is not applicable
		p.addErrorf("expected ';' or line break after expression, got %s", p.peekToken.Type)
	}

	return stmt
}

// parseBlockStatement parses a block statement.
func (p *Parser) parseBlockStatement() *ast.BlockStatement {
	block := &ast.BlockStatement{
		LBrace: p.currentToken.Position,
	}

	p.nextToken()

	for !p.currentTokenIs(lexer.RBRACE) && !p.currentTokenIs(lexer.EOF) {
		stmt := p.parseStatement()
		if stmt != nil {
			block.Body = append(block.Body, stmt)
		}
		p.nextToken()
	}

	if p.currentTokenIs(lexer.RBRACE) {
		block.RBrace = p.currentToken.Position
	}

	return block
}

// parseEmptyStatement parses an empty statement (just a semicolon).
func (p *Parser) parseEmptyStatement() ast.Statement {
	return &ast.EmptyStatement{
		Semicolon: p.currentToken.Position,
	}
}

// parseVariableDeclaration parses a variable declaration (let, const, var).
func (p *Parser) parseVariableDeclaration() ast.Statement {
	stmt := &ast.VariableDeclaration{
		DeclPos: p.currentToken.Position,
		Kind:    p.currentToken.Type,
	}

	if !p.expectPeek(lexer.IDENT) {
		return nil
	}

	declarator := p.parseVariableDeclarator()
	if declarator != nil {
		stmt.Declarations = append(stmt.Declarations, declarator)
	}

	// Handle multiple declarations separated by commas
	for p.peekTokenIs(lexer.COMMA) {
		p.nextToken()
		if !p.expectPeek(lexer.IDENT) {
			return nil
		}
		declarator := p.parseVariableDeclarator()
		if declarator != nil {
			stmt.Declarations = append(stmt.Declarations, declarator)
		}
	}

	// Use ASI logic for optional semicolon
	if p.peekTokenIs(lexer.SEMICOLON) {
		p.nextToken()
		stmt.Semicolon = p.currentToken.Position
	} else if !p.canInsertSemicolon() {
		// Only report error if ASI is not applicable
		p.addErrorf("expected ';' or line break after variable declaration, got %s", p.peekToken.Type)
	}

	return stmt
}

// parseVariableDeclarator parses a single variable declarator.
func (p *Parser) parseVariableDeclarator() *ast.VariableDeclarator {
	declarator := &ast.VariableDeclarator{
		Id: p.parseIdentifier(),
	}

	// Optional type annotation
	if p.peekTokenIs(lexer.COLON) {
		p.nextToken() // consume ':'
		p.nextToken() // move to type
		declarator.TypeAnnotation = p.parseTypeAnnotation()
	}

	// Optional initializer
	if p.peekTokenIs(lexer.ASSIGN) {
		p.nextToken()
		p.nextToken()
		declarator.Init = p.parseExpression(LOWEST)
	}

	return declarator
}

// parseIfStatement parses an if statement.
func (p *Parser) parseIfStatement() ast.Statement {
	stmt := &ast.IfStatement{
		IfPos: p.currentToken.Position,
	}

	if !p.expectPeek(lexer.LPAREN) {
		return nil
	}

	stmt.LParen = p.currentToken.Position
	p.nextToken()
	stmt.Test = p.parseExpression(LOWEST)

	if !p.expectPeek(lexer.RPAREN) {
		return nil
	}

	stmt.RParen = p.currentToken.Position

	if !p.expectPeek(lexer.LBRACE) {
		return nil
	}

	stmt.Consequent = p.parseBlockStatement()

	// Optional else clause
	if p.peekTokenIs(lexer.ELSE) {
		p.nextToken()
		stmt.ElsePos = p.currentToken.Position

		if p.peekTokenIs(lexer.IF) {
			// else if
			p.nextToken()
			stmt.Alternate = p.parseIfStatement()
		} else if p.peekTokenIs(lexer.LBRACE) {
			// else block
			p.nextToken()
			stmt.Alternate = p.parseBlockStatement()
		}
	}

	return stmt
}

// parseWhileStatement parses a while statement.
func (p *Parser) parseWhileStatement() ast.Statement {
	stmt := &ast.WhileStatement{
		WhilePos: p.currentToken.Position,
	}

	if !p.expectPeek(lexer.LPAREN) {
		return nil
	}

	stmt.LParen = p.currentToken.Position
	p.nextToken()
	stmt.Test = p.parseExpression(LOWEST)

	if !p.expectPeek(lexer.RPAREN) {
		return nil
	}

	stmt.RParen = p.currentToken.Position

	if !p.expectPeek(lexer.LBRACE) {
		return nil
	}

	stmt.Body = p.parseBlockStatement()

	return stmt
}

// parseForStatement parses a for statement.
func (p *Parser) parseForStatement() ast.Statement {
	forPos := p.currentToken.Position

	if !p.expectPeek(lexer.LPAREN) {
		return nil
	}

	lParen := p.currentToken.Position
	p.nextToken()

	// Check if it's a for-in or for-of loop
	if p.currentTokenIs(lexer.LET) || p.currentTokenIs(lexer.CONST) || p.currentTokenIs(lexer.VAR) {
		// Could be for-in/for-of or regular for loop
		if !p.expectPeek(lexer.IDENT) {
			return nil
		}

		id := p.parseIdentifier()

		if p.peekTokenIs(lexer.IN) {
			// for-in loop
			p.nextToken()
			inPos := p.currentToken.Position
			p.nextToken()
			right := p.parseExpression(LOWEST)

			if !p.expectPeek(lexer.RPAREN) {
				return nil
			}

			rParen := p.currentToken.Position

			if !p.expectPeek(lexer.LBRACE) {
				return nil
			}

			body := p.parseBlockStatement()

			return &ast.ForInStatement{
				ForPos: forPos,
				LParen: lParen,
				Left:   id,
				InPos:  inPos,
				Right:  right,
				RParen: rParen,
				Body:   body,
			}
		} else if p.peekTokenIs(lexer.IDENT) && p.peekToken.Literal == "of" {
			// for-of loop (treat "of" as identifier for now)
			p.nextToken()
			ofPos := p.currentToken.Position
			p.nextToken()
			right := p.parseExpression(LOWEST)

			if !p.expectPeek(lexer.RPAREN) {
				return nil
			}

			rParen := p.currentToken.Position

			if !p.expectPeek(lexer.LBRACE) {
				return nil
			}

			body := p.parseBlockStatement()

			return &ast.ForOfStatement{
				ForPos: forPos,
				LParen: lParen,
				Left:   id,
				OfPos:  ofPos,
				Right:  right,
				RParen: rParen,
				Body:   body,
			}
		} else {
			// Regular for loop with declaration
			// Backtrack and parse as regular for loop
			// This is a simplified approach - in a full implementation,
			// we'd need better lookahead
		}
	}

	// Regular for loop
	var init ast.Statement
	if !p.currentTokenIs(lexer.SEMICOLON) {
		if p.currentTokenIs(lexer.LET) || p.currentTokenIs(lexer.CONST) || p.currentTokenIs(lexer.VAR) {
			init = p.parseVariableDeclaration()
		} else {
			init = p.parseExpressionStatement()
		}
	}

	if !p.expectPeek(lexer.SEMICOLON) {
		return nil
	}

	p.nextToken()

	var test ast.Expression
	if !p.currentTokenIs(lexer.SEMICOLON) {
		test = p.parseExpression(LOWEST)
	}

	if !p.expectPeek(lexer.SEMICOLON) {
		return nil
	}

	p.nextToken()

	var update ast.Expression
	if !p.currentTokenIs(lexer.RPAREN) {
		update = p.parseExpression(LOWEST)
	}

	if !p.expectPeek(lexer.RPAREN) {
		return nil
	}

	rParen := p.currentToken.Position

	if !p.expectPeek(lexer.LBRACE) {
		return nil
	}

	body := p.parseBlockStatement()

	return &ast.ForStatement{
		ForPos: forPos,
		LParen: lParen,
		Init:   init,
		Test:   test,
		Update: update,
		RParen: rParen,
		Body:   body,
	}
}

// parseReturnStatement parses a return statement.
func (p *Parser) parseReturnStatement() ast.Statement {
	stmt := &ast.ReturnStatement{
		ReturnPos: p.currentToken.Position,
	}

	// Check for ASI after return keyword (restricted production)
	if p.canInsertSemicolon() {
		// Return without argument due to ASI
		return stmt
	}

	if p.peekTokenIs(lexer.SEMICOLON) {
		p.nextToken()
		stmt.Semicolon = p.currentToken.Position
		return stmt
	}

	p.nextToken()
	stmt.Argument = p.parseExpression(LOWEST)

	// Use ASI logic for optional semicolon
	if p.peekTokenIs(lexer.SEMICOLON) {
		p.nextToken()
		stmt.Semicolon = p.currentToken.Position
	} else if !p.canInsertSemicolon() {
		// Only report error if ASI is not applicable
		p.addErrorf("expected ';' or line break after return statement, got %s", p.peekToken.Type)
	}

	return stmt
}

// parseBreakStatement parses a break statement.
func (p *Parser) parseBreakStatement() ast.Statement {
	stmt := &ast.BreakStatement{
		BreakPos: p.currentToken.Position,
	}

	// Check for ASI after break keyword (restricted production)
	if !p.canInsertSemicolon() && p.peekTokenIs(lexer.IDENT) {
		p.nextToken()
		stmt.Label = p.parseIdentifier()
	}

	// Use ASI logic for optional semicolon
	if p.peekTokenIs(lexer.SEMICOLON) {
		p.nextToken()
		stmt.Semicolon = p.currentToken.Position
	} else if !p.canInsertSemicolon() {
		// Only report error if ASI is not applicable
		p.addErrorf("expected ';' or line break after break statement, got %s", p.peekToken.Type)
	}

	return stmt
}

// parseContinueStatement parses a continue statement.
func (p *Parser) parseContinueStatement() ast.Statement {
	stmt := &ast.ContinueStatement{
		ContinuePos: p.currentToken.Position,
	}

	// Check for ASI after continue keyword (restricted production)
	if !p.canInsertSemicolon() && p.peekTokenIs(lexer.IDENT) {
		p.nextToken()
		stmt.Label = p.parseIdentifier()
	}

	// Use ASI logic for optional semicolon
	if p.peekTokenIs(lexer.SEMICOLON) {
		p.nextToken()
		stmt.Semicolon = p.currentToken.Position
	} else if !p.canInsertSemicolon() {
		// Only report error if ASI is not applicable
		p.addErrorf("expected ';' or line break after continue statement, got %s", p.peekToken.Type)
	}

	return stmt
}

// parseFunctionDeclaration parses a function declaration.
func (p *Parser) parseFunctionDeclaration() ast.Statement {
	fn := &ast.FunctionDeclaration{
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

	if !p.expectPeek(lexer.IDENT) {
		return nil
	}

	fn.Name = p.parseIdentifier()

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

// parseClassDeclaration parses a class declaration.
func (p *Parser) parseClassDeclaration() ast.Statement {
	class := &ast.ClassDeclaration{
		ClassPos: p.currentToken.Position,
	}

	if !p.expectPeek(lexer.IDENT) {
		return nil
	}

	class.Name = p.parseIdentifier()

	// Optional extends clause
	if p.peekTokenIs(lexer.EXTENDS) {
		p.nextToken()
		p.nextToken()
		class.SuperClass = p.parseExpression(LOWEST)
	}

	if !p.expectPeek(lexer.LBRACE) {
		return nil
	}

	class.LBrace = p.currentToken.Position
	p.nextToken()

	// Parse class body
	for !p.currentTokenIs(lexer.RBRACE) && !p.currentTokenIs(lexer.EOF) {
		member := p.parseClassMember()
		if member != nil {
			class.Body = append(class.Body, member)
		}
		p.nextToken()
	}

	if p.currentTokenIs(lexer.RBRACE) {
		class.RBrace = p.currentToken.Position
	}

	return class
}

// parseClassMember parses a class member (method or property).
func (p *Parser) parseClassMember() ast.Node {
	// This is a simplified implementation
	// In a full implementation, this would handle:
	// - static methods/properties
	// - private/protected/public modifiers
	// - getters/setters
	// - computed property names
	// - etc.

	if p.currentTokenIs(lexer.IDENT) && p.peekTokenIs(lexer.LPAREN) {
		// Method
		method := &ast.MethodDefinition{
			Key:  p.parseIdentifierExpression(),
			Kind: "method",
		}

		// Create a function expression for the method
		fn := &ast.FunctionExpression{
			FunctionPos: p.currentToken.Position,
		}

		if !p.expectPeek(lexer.LPAREN) {
			return nil
		}

		fn.LParen = p.currentToken.Position
		fn.Parameters = p.parseParameterList()

		if p.currentTokenIs(lexer.RPAREN) {
			fn.RParen = p.currentToken.Position
		}

		if !p.expectPeek(lexer.LBRACE) {
			return nil
		}

		fn.Body = p.parseBlockStatement()
		method.Value = fn

		return method
	}

	// Property (simplified)
	if p.currentTokenIs(lexer.IDENT) {
		prop := &ast.PropertyDefinition{
			Key: p.parseIdentifierExpression(),
		}

		if p.peekTokenIs(lexer.ASSIGN) {
			p.nextToken()
			p.nextToken()
			prop.Value = p.parseExpression(LOWEST)
		}

		return prop
	}

	return nil
}

// parseInterfaceDeclaration parses an interface declaration.
func (p *Parser) parseInterfaceDeclaration() ast.Statement {
	iface := &ast.InterfaceDeclaration{
		InterfacePos: p.currentToken.Position,
	}

	if !p.expectPeek(lexer.IDENT) {
		return nil
	}

	iface.Name = p.parseIdentifier()

	// Optional generic type parameters
	if p.peekTokenIs(lexer.LT) {
		p.nextToken()
		iface.TypeParameters = p.parseTypeParameterList()
	}

	// Optional extends clause
	if p.peekTokenIs(lexer.EXTENDS) {
		p.nextToken()
		p.nextToken()
		iface.Extends = append(iface.Extends, p.parseTypeAnnotation())

		for p.peekTokenIs(lexer.COMMA) {
			p.nextToken()
			p.nextToken()
			iface.Extends = append(iface.Extends, p.parseTypeAnnotation())
		}
	}

	if !p.expectPeek(lexer.LBRACE) {
		return nil
	}

	iface.LBrace = p.currentToken.Position
	p.nextToken()

	// Parse interface body
	for !p.currentTokenIs(lexer.RBRACE) && !p.currentTokenIs(lexer.EOF) {
		member := p.parseTypeMember()
		if member != nil {
			iface.Body = append(iface.Body, member)
		}
		p.nextToken()
	}

	if p.currentTokenIs(lexer.RBRACE) {
		iface.RBrace = p.currentToken.Position
	}

	return iface
}

// parseTypeAliasDeclaration parses a type alias declaration.
func (p *Parser) parseTypeAliasDeclaration() ast.Statement {
	alias := &ast.TypeAliasDeclaration{
		TypePos: p.currentToken.Position,
	}

	if !p.expectPeek(lexer.IDENT) {
		return nil
	}

	alias.Name = p.parseIdentifier()

	// Optional generic type parameters
	if p.peekTokenIs(lexer.LT) {
		p.nextToken()
		alias.TypeParameters = p.parseTypeParameterList()
	}

	if !p.expectPeek(lexer.ASSIGN) {
		return nil
	}

	alias.Assign = p.currentToken.Position
	p.nextToken()
	alias.Type = p.parseTypeAnnotation()

	if p.peekTokenIs(lexer.SEMICOLON) {
		p.nextToken()
	}

	return alias
}

// parseEnumDeclaration parses an enum declaration.
func (p *Parser) parseEnumDeclaration() ast.Statement {
	enum := &ast.EnumDeclaration{
		EnumPos: p.currentToken.Position,
	}

	if !p.expectPeek(lexer.IDENT) {
		return nil
	}

	enum.Name = p.parseIdentifier()

	if !p.expectPeek(lexer.LBRACE) {
		return nil
	}

	enum.LBrace = p.currentToken.Position
	p.nextToken()

	// Parse enum members
	for !p.currentTokenIs(lexer.RBRACE) && !p.currentTokenIs(lexer.EOF) {
		member := p.parseEnumMember()
		if member != nil {
			enum.Members = append(enum.Members, member)
		}

		if p.peekTokenIs(lexer.COMMA) {
			p.nextToken()
		}
		p.nextToken()
	}

	if p.currentTokenIs(lexer.RBRACE) {
		enum.RBrace = p.currentToken.Position
	}

	return enum
}

// parseEnumMember parses an enum member.
func (p *Parser) parseEnumMember() *ast.EnumMember {
	if !p.currentTokenIs(lexer.IDENT) {
		return nil
	}

	member := &ast.EnumMember{
		Name: p.parseIdentifier(),
	}

	if p.peekTokenIs(lexer.ASSIGN) {
		p.nextToken()
		p.nextToken()
		member.Value = p.parseExpression(LOWEST)
	}

	return member
}

// parseTypeMember parses a type member (for interfaces and object types).
func (p *Parser) parseTypeMember() *ast.TypeMember {
	if !p.currentTokenIs(lexer.IDENT) {
		return nil
	}

	member := &ast.TypeMember{
		Key: p.parseIdentifierExpression(),
	}

	if p.peekTokenIs(lexer.QUESTION) {
		p.nextToken()
		member.Optional = true
	}

	if !p.expectPeek(lexer.COLON) {
		return nil
	}

	p.nextToken()
	member.Type = p.parseTypeAnnotation()

	if p.peekTokenIs(lexer.SEMICOLON) {
		p.nextToken()
	}

	return member
}

// parseTypeParameterList parses a list of generic type parameters.
func (p *Parser) parseTypeParameterList() []*ast.TypeParameter {
	var params []*ast.TypeParameter

	if p.peekTokenIs(lexer.GT) {
		p.nextToken()
		return params
	}

	p.nextToken()

	param := p.parseTypeParameter()
	if param != nil {
		params = append(params, param)
	}

	for p.peekTokenIs(lexer.COMMA) {
		p.nextToken()
		p.nextToken()
		param := p.parseTypeParameter()
		if param != nil {
			params = append(params, param)
		}
	}

	if !p.expectPeek(lexer.GT) {
		return nil
	}

	return params
}

// parseTypeParameter parses a generic type parameter.
func (p *Parser) parseTypeParameter() *ast.TypeParameter {
	if !p.currentTokenIs(lexer.IDENT) {
		return nil
	}

	param := &ast.TypeParameter{
		Name: p.parseIdentifier(),
	}

	if p.peekTokenIs(lexer.EXTENDS) {
		p.nextToken()
		p.nextToken()
		param.Constraint = p.parseTypeAnnotation()
	}

	if p.peekTokenIs(lexer.ASSIGN) {
		p.nextToken()
		p.nextToken()
		param.Default = p.parseTypeAnnotation()
	}

	return param
}
