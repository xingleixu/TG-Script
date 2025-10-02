package parser

import (
	"github.com/xingleixu/TG-Script/ast"
	"github.com/xingleixu/TG-Script/lexer"
)

// ============================================================================
// TYPE PARSING
// ============================================================================

// parseTypeAnnotation parses a type annotation.
func (p *Parser) parseTypeAnnotation() ast.TypeNode {
	return p.parseUnionType()
}

// parseUnionType parses a union type (type1 | type2 | ...).
func (p *Parser) parseUnionType() ast.TypeNode {
	left := p.parseIntersectionType()

	if !p.peekTokenIs(lexer.BIT_OR) {
		return left
	}

	union := &ast.UnionType{
		Types: []ast.TypeNode{left},
	}

	for p.peekTokenIs(lexer.BIT_OR) {
		p.nextToken() // consume '|'
		p.nextToken()
		right := p.parseIntersectionType()
		union.Types = append(union.Types, right)
	}

	return union
}

// parseIntersectionType parses an intersection type (type1 & type2 & ...).
func (p *Parser) parseIntersectionType() ast.TypeNode {
	left := p.parsePrimaryType()

	if !p.peekTokenIs(lexer.BIT_AND) {
		return left
	}

	intersection := &ast.IntersectionType{
		Types: []ast.TypeNode{left},
	}

	for p.peekTokenIs(lexer.BIT_AND) {
		p.nextToken() // consume '&'
		p.nextToken()
		right := p.parsePrimaryType()
		intersection.Types = append(intersection.Types, right)
	}

	return intersection
}

// parsePrimaryType parses a primary type.
func (p *Parser) parsePrimaryType() ast.TypeNode {
	var baseType ast.TypeNode
	
	switch p.currentToken.Type {
	case lexer.IDENT:
		baseType = p.parseTypeReference()
	case lexer.LPAREN:
		baseType = p.parseGroupedType()
	case lexer.LBRACE:
		baseType = p.parseObjectType()
	case lexer.LBRACKET:
		baseType = p.parseArrayOrTupleType()
	case lexer.FUNCTION:
		baseType = p.parseFunctionType()
	case lexer.STRING_T, lexer.NUMBER_T, lexer.BOOLEAN_T, lexer.INT_T, lexer.FLOAT_T, lexer.VOID, lexer.NULL, lexer.UNDEFINED,
		 lexer.INT8_T, lexer.INT16_T, lexer.INT32_T, lexer.INT64_T, lexer.FLOAT32_T, lexer.FLOAT64_T:
		// Handle primitive type tokens
		baseType = p.parseTypeReference()
	default:
		// Handle primitive types by literal for backward compatibility
		switch p.currentToken.Literal {
		case "string", "number", "boolean", "void", "any", "unknown", "never", "undefined", "null":
			baseType = p.parseTypeReference()
		default:
			p.addErrorf("unexpected token in type: %s", p.currentToken.Literal)
			return nil
		}
	}
	
	// Check for array suffix: elementType[]
	for p.peekTokenIs(lexer.LBRACKET) {
		p.nextToken() // consume '['
		if !p.expectPeek(lexer.RBRACKET) {
			p.addError("expected ']' after '['")
			return nil
		}
		// Create array type with the base type as element type
		baseType = &ast.ArrayType{
			LBracket:    p.currentToken.Position,
			ElementType: baseType,
			RBracket:    p.currentToken.Position,
		}
	}
	
	return baseType
}

// parseTypeReference parses a type reference (identifier or qualified name).
func (p *Parser) parseTypeReference() ast.TypeNode {
	// Handle basic type keywords
	switch p.currentToken.Type {
	case lexer.STRING_T, lexer.NUMBER_T, lexer.BOOLEAN_T, lexer.INT_T, lexer.FLOAT_T, lexer.VOID, lexer.NULL, lexer.UNDEFINED, lexer.ANY, lexer.UNKNOWN, lexer.NEVER,
		 lexer.INT8_T, lexer.INT16_T, lexer.INT32_T, lexer.INT64_T, lexer.FLOAT32_T, lexer.FLOAT64_T:
		// Create BasicType for built-in types
		return &ast.BasicType{
			TypePos: p.currentToken.Position,
			Kind:    p.currentToken.Type,
		}
	case lexer.IDENT:
		// Create TypeReference for user-defined types
		name := p.parseIdentifier()
		ref := &ast.TypeReference{
			Name: name,
		}

		// Optional generic type arguments
		if p.peekTokenIs(lexer.LT) {
			p.nextToken()
			ref.TypeArgs = p.parseTypeArgumentList()
		}

		return ref
	default:
		p.addErrorf("expected type name, got %s", p.currentToken.Literal)
		return nil
	}
}

// parseGroupedType parses a grouped type (parenthesized type).
func (p *Parser) parseGroupedType() ast.TypeNode {
	p.nextToken() // consume '('
	typ := p.parseTypeAnnotation()
	if !p.expectPeek(lexer.RPAREN) {
		return nil
	}
	return typ
}

// parseObjectType parses an object type.
func (p *Parser) parseObjectType() *ast.ObjectType {
	obj := &ast.ObjectType{
		LBrace: p.currentToken.Position,
	}

	p.nextToken()

	for !p.currentTokenIs(lexer.RBRACE) && !p.currentTokenIs(lexer.EOF) {
		member := p.parseTypeMember()
		if member != nil {
			obj.Members = append(obj.Members, member)
		}

		if p.peekTokenIs(lexer.SEMICOLON) || p.peekTokenIs(lexer.COMMA) {
			p.nextToken()
		}
		p.nextToken()
	}

	if p.currentTokenIs(lexer.RBRACE) {
		obj.RBrace = p.currentToken.Position
	}

	return obj
}

// parseArrayOrTupleType parses an array type or tuple type.
func (p *Parser) parseArrayOrTupleType() ast.TypeNode {
	lBracket := p.currentToken.Position
	p.nextToken()

	// Empty array type []
	if p.currentTokenIs(lexer.RBRACKET) {
		// This is actually invalid syntax, but we'll handle it gracefully
		p.addError("empty array type syntax is invalid")
		return nil
	}

	// Check if it's a tuple type (multiple types separated by commas)
	firstType := p.parseTypeAnnotation()

	if p.peekTokenIs(lexer.COMMA) {
		// Tuple type
		tuple := &ast.TupleType{
			LBracket: lBracket,
			Elements: []ast.TypeNode{firstType},
		}

		for p.peekTokenIs(lexer.COMMA) {
			p.nextToken() // consume ','
			p.nextToken()
			typ := p.parseTypeAnnotation()
			tuple.Elements = append(tuple.Elements, typ)
		}

		if !p.expectPeek(lexer.RBRACKET) {
			return nil
		}

		tuple.RBracket = p.currentToken.Position
		return tuple
	}

	// Array type
	if !p.expectPeek(lexer.RBRACKET) {
		return nil
	}

	return &ast.ArrayType{
		LBracket:    lBracket,
		ElementType: firstType,
		RBracket:    p.currentToken.Position,
	}
}

// parseFunctionType parses a function type.
func (p *Parser) parseFunctionType() *ast.FunctionType {
	fn := &ast.FunctionType{}

	if !p.expectPeek(lexer.LPAREN) {
		return nil
	}

	fn.LParen = p.currentToken.Position
	fn.Parameters = p.parseParameterList()

	if p.currentTokenIs(lexer.RPAREN) {
		fn.RParen = p.currentToken.Position
	}

	if !p.expectPeek(lexer.COLON) {
		return nil
	}

	p.nextToken()
	fn.ReturnType = p.parseTypeAnnotation()

	return fn
}

// parseTypeArgumentList parses a list of type arguments.
func (p *Parser) parseTypeArgumentList() []ast.TypeNode {
	var args []ast.TypeNode

	if p.peekTokenIs(lexer.GT) {
		p.nextToken()
		return args
	}

	p.nextToken()
	arg := p.parseTypeAnnotation()
	args = append(args, arg)

	for p.peekTokenIs(lexer.COMMA) {
		p.nextToken()
		p.nextToken()
		arg := p.parseTypeAnnotation()
		args = append(args, arg)
	}

	if !p.expectPeek(lexer.GT) {
		return nil
	}

	return args
}

// parseModifiers parses modifiers (public, private, protected, static, etc.).
func (p *Parser) parseModifiers() []ast.Modifier {
	var modifiers []ast.Modifier

	for {
		switch p.currentToken.Literal {
		case "public", "private", "protected", "static", "readonly", "abstract", "async":
			modifiers = append(modifiers, ast.Modifier{
				Kind: p.currentToken.Literal,
				Pos:  p.currentToken.Position,
			})
			p.nextToken()
		default:
			return modifiers
		}
	}
}

// parseTypeAssertion parses a type assertion (value as Type).
func (p *Parser) parseTypeAssertion(expression ast.Expression) *ast.TypeAssertion {
	assertion := &ast.TypeAssertion{
		Expression: expression,
		AsPos:      p.currentToken.Position,
	}

	p.nextToken()
	assertion.Type = p.parseTypeAnnotation()

	return assertion
}

// parseNonNullAssertion parses a non-null assertion (value!).
func (p *Parser) parseNonNullAssertion(expression ast.Expression) *ast.NonNullAssertion {
	return &ast.NonNullAssertion{
		Expression: expression,
		Bang:       p.currentToken.Position,
	}
}

// Helper function to check if current token is a type keyword
func (p *Parser) isTypeKeyword() bool {
	switch p.currentToken.Literal {
	case "string", "number", "boolean", "void", "any", "unknown", "never", "undefined", "null":
		return true
	default:
		return false
	}
}

// Helper function to check if we're at the start of a type annotation
func (p *Parser) isAtTypeAnnotation() bool {
	return p.currentTokenIs(lexer.IDENT) || p.currentTokenIs(lexer.LBRACE) ||
		p.currentTokenIs(lexer.LBRACKET) || p.currentTokenIs(lexer.LPAREN) ||
		p.isTypeKeyword()
}
