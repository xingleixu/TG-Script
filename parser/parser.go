package parser

import (
	"fmt"
	"strconv"

	"github.com/xingleixu/TG-Script/ast"
	"github.com/xingleixu/TG-Script/lexer"
)

// Parser represents the parser state.
type Parser struct {
	lexer *lexer.Lexer

	currentToken lexer.TokenInfo
	peekToken    lexer.TokenInfo

	errors []string
}

// New creates a new parser instance.
func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		lexer:  l,
		errors: []string{},
	}

	// Read two tokens, so currentToken and peekToken are both set
	p.nextToken()
	p.nextToken()

	// Register prefix parse functions
	p.registerPrefix(lexer.IDENT, p.parseIdentifierExpression)
	p.registerPrefix(lexer.INT, p.parseIntegerLiteralExpression)
	p.registerPrefix(lexer.FLOAT, p.parseFloatLiteralExpression)
	p.registerPrefix(lexer.STRING, p.parseStringLiteralExpression)
	p.registerPrefix(lexer.BOOLEAN, p.parseBooleanLiteralExpression)
	p.registerPrefix(lexer.NULL, p.parseNullLiteralExpression)
	p.registerPrefix(lexer.UNDEFINED, p.parseUndefinedLiteralExpression)
	p.registerPrefix(lexer.VOID, p.parseVoidLiteralExpression)
	p.registerPrefix(lexer.LOGICAL_NOT, p.parsePrefixExpression)
	p.registerPrefix(lexer.SUB, p.parsePrefixExpression)
	p.registerPrefix(lexer.ADD, p.parsePrefixExpression)
	p.registerPrefix(lexer.BIT_NOT, p.parsePrefixExpression)
	p.registerPrefix(lexer.LPAREN, p.parseGroupedExpression)
	p.registerPrefix(lexer.LBRACKET, p.parseArrayLiteral)
	p.registerPrefix(lexer.LBRACE, p.parseObjectLiteral)
	p.registerPrefix(lexer.FUNCTION, p.parseFunctionExpression)
	p.registerPrefix(lexer.TYPEOF, p.parseTypeofExpression)
	p.registerPrefix(lexer.DELETE, p.parseDeleteExpression)
	p.registerPrefix(lexer.NEW, p.parseNewExpression)
	p.registerPrefix(lexer.THIS, p.parseThisExpression)
	p.registerPrefix(lexer.SUPER, p.parseSuperExpression)
	p.registerPrefix(lexer.AWAIT, p.parseAwaitExpression)
	p.registerPrefix(lexer.YIELD, p.parseYieldExpression)
	p.registerPrefix(lexer.INCREMENT, p.parseIncrementExpression)
	p.registerPrefix(lexer.DECREMENT, p.parseDecrementExpression)

	// Register infix parse functions
	p.registerInfix(lexer.ADD, p.parseInfixExpression)
	p.registerInfix(lexer.SUB, p.parseInfixExpression)
	p.registerInfix(lexer.MUL, p.parseInfixExpression)
	p.registerInfix(lexer.DIV, p.parseInfixExpression)
	p.registerInfix(lexer.MOD, p.parseInfixExpression)
	p.registerInfix(lexer.POW, p.parseInfixExpression)
	p.registerInfix(lexer.EQ, p.parseInfixExpression)
	p.registerInfix(lexer.NE, p.parseInfixExpression)
	p.registerInfix(lexer.STRICT_EQ, p.parseInfixExpression)
	p.registerInfix(lexer.STRICT_NE, p.parseInfixExpression)
	p.registerInfix(lexer.LT, p.parseInfixExpression)
	p.registerInfix(lexer.GT, p.parseInfixExpression)
	p.registerInfix(lexer.LE, p.parseInfixExpression)
	p.registerInfix(lexer.GE, p.parseInfixExpression)
	p.registerInfix(lexer.LOGICAL_AND, p.parseInfixExpression)
	p.registerInfix(lexer.LOGICAL_OR, p.parseInfixExpression)
	p.registerInfix(lexer.BIT_AND, p.parseInfixExpression)
	p.registerInfix(lexer.BIT_OR, p.parseInfixExpression)
	p.registerInfix(lexer.BIT_XOR, p.parseInfixExpression)
	p.registerInfix(lexer.BIT_LSHIFT, p.parseInfixExpression)
	p.registerInfix(lexer.BIT_RSHIFT, p.parseInfixExpression)
	p.registerInfix(lexer.BIT_URSHIFT, p.parseInfixExpression)
	p.registerInfix(lexer.ASSIGN, p.parseAssignmentExpression)
	p.registerInfix(lexer.ADD_ASSIGN, p.parseAssignmentExpression)
	p.registerInfix(lexer.SUB_ASSIGN, p.parseAssignmentExpression)
	p.registerInfix(lexer.MUL_ASSIGN, p.parseAssignmentExpression)
	p.registerInfix(lexer.DIV_ASSIGN, p.parseAssignmentExpression)
	p.registerInfix(lexer.MOD_ASSIGN, p.parseAssignmentExpression)
	p.registerInfix(lexer.LPAREN, p.parseCallExpression)
	p.registerInfix(lexer.LBRACKET, p.parseIndexExpression)
	p.registerInfix(lexer.DOT, p.parseMemberExpression)
	p.registerInfix(lexer.QUESTION, p.parseTernaryExpression)
	p.registerInfix(lexer.INSTANCEOF, p.parseInstanceofExpression)
	p.registerInfix(lexer.IN, p.parseInExpression)
	p.registerInfix(lexer.NULLISH, p.parseNullishCoalescingExpression)
	p.registerInfix(lexer.OPTIONAL, p.parseOptionalChainingExpression)
	p.registerInfix(lexer.INCREMENT, p.parsePostfixIncrementExpression)
	p.registerInfix(lexer.DECREMENT, p.parsePostfixDecrementExpression)
	p.registerInfix(lexer.ARROW, p.parseArrowFunctionExpression)

	return p
}

// nextToken advances both currentToken and peekToken.
func (p *Parser) nextToken() {
	p.currentToken = p.peekToken
	p.peekToken = p.lexer.NextToken()
	
	// Skip comments
	for p.peekToken.Type == lexer.COMMENT {
		p.peekToken = p.lexer.NextToken()
	}
}

// Errors returns the list of parsing errors.
func (p *Parser) Errors() []string {
	return p.errors
}

// addError adds an error message to the parser's error list.
func (p *Parser) addError(msg string) {
	p.errors = append(p.errors, msg)
}

// addErrorf adds a formatted error message to the parser's error list.
func (p *Parser) addErrorf(format string, args ...interface{}) {
	p.addError(fmt.Sprintf(format, args...))
}

// expectToken checks if the current token is of the expected type and advances.
func (p *Parser) expectToken(tokenType lexer.Token) bool {
	if p.currentToken.Type == tokenType {
		p.nextToken()
		return true
	}
	p.addErrorf("expected %s, got %s", tokenType, p.currentToken.Type)
	return false
}

// expectPeek checks if the peek token is of the expected type and advances if so.
func (p *Parser) expectPeek(tokenType lexer.Token) bool {
	if p.peekToken.Type == tokenType {
		p.nextToken()
		return true
	}
	p.addErrorf("expected next token to be %s, got %s", tokenType, p.peekToken.Type)
	return false
}

// currentTokenIs checks if the current token is of the given type.
func (p *Parser) currentTokenIs(tokenType lexer.Token) bool {
	return p.currentToken.Type == tokenType
}

// peekTokenIs checks if the peek token is of the given type.
func (p *Parser) peekTokenIs(tokenType lexer.Token) bool {
	return p.peekToken.Type == tokenType
}

// skipSemicolon skips optional semicolons.
func (p *Parser) skipSemicolon() {
	if p.currentTokenIs(lexer.SEMICOLON) {
		p.nextToken()
	}
}

// canInsertSemicolon checks if a semicolon can be automatically inserted
// according to TypeScript/JavaScript ASI rules.
func (p *Parser) canInsertSemicolon() bool {
	// ASI rules:
	// 1. At the end of input (EOF)
	// 2. Before a closing brace '}'
	// 3. After a line terminator (newline)
	// 4. Before certain restricted tokens (return, break, continue, etc.)
	
	if p.peekToken.Type == lexer.EOF {
		return true
	}
	
	if p.peekToken.Type == lexer.RBRACE {
		return true
	}
	
	// Check if there's a line break between current and peek token
	if p.currentToken.Position.Line < p.peekToken.Position.Line {
		return true
	}
	
	return false
}

// expectSemicolonOrASI expects a semicolon or allows automatic semicolon insertion
func (p *Parser) expectSemicolonOrASI() bool {
	if p.peekTokenIs(lexer.SEMICOLON) {
		p.nextToken()
		return true
	}
	
	if p.canInsertSemicolon() {
		return true
	}
	
	p.addErrorf("expected ';' or line break, got %s", p.peekToken.Type)
	return false
}

// ============================================================================
// PRECEDENCE HANDLING
// ============================================================================

// Precedence represents operator precedence levels.
type Precedence int

const (
	_ Precedence = iota
	LOWEST
	ARROW       // =>
	ASSIGN      // =, +=, -=, etc.
	TERNARY     // ? :
	NULLISH     // ??
	LOGICAL_OR  // ||
	LOGICAL_AND // &&
	BITWISE_OR  // |
	BITWISE_XOR // ^
	BITWISE_AND // &
	EQUALITY    // ==, !=, ===, !==
	RELATIONAL  // <, >, <=, >=, instanceof, in
	SHIFT       // <<, >>, >>>
	SUM         // +, -
	PRODUCT     // *, /, %
	EXPONENT    // **
	UNARY       // !, -, +, ~, typeof, void, delete
	POSTFIX     // ++, --
	CALL        // (), [], .
	MEMBER      // .
	OPTIONAL    // ?.
	PRIMARY     // literals, identifiers
)

// precedences maps token types to their precedence levels.
var precedences = map[lexer.Token]Precedence{
	lexer.ARROW:         ARROW,

	lexer.ASSIGN:        ASSIGN,
	lexer.ADD_ASSIGN:    ASSIGN,
	lexer.SUB_ASSIGN:    ASSIGN,
	lexer.MUL_ASSIGN:    ASSIGN,
	lexer.DIV_ASSIGN:    ASSIGN,
	lexer.MOD_ASSIGN:    ASSIGN,

	lexer.QUESTION:      TERNARY,

	lexer.NULLISH: NULLISH,

	lexer.LOGICAL_OR:    LOGICAL_OR,
	lexer.LOGICAL_AND:   LOGICAL_AND,

	lexer.BIT_OR:        BITWISE_OR,
	lexer.BIT_XOR:       BITWISE_XOR,
	lexer.BIT_AND:       BITWISE_AND,

	lexer.EQ:            EQUALITY,
	lexer.NE:            EQUALITY,
	lexer.STRICT_EQ:     EQUALITY,
	lexer.STRICT_NE:     EQUALITY,

	lexer.LT:            RELATIONAL,
	lexer.GT:            RELATIONAL,
	lexer.LE:            RELATIONAL,
	lexer.GE:            RELATIONAL,
	lexer.INSTANCEOF:    RELATIONAL,
	lexer.IN:            RELATIONAL,

	lexer.BIT_LSHIFT:    SHIFT,
	lexer.BIT_RSHIFT:    SHIFT,
	lexer.BIT_URSHIFT:   SHIFT,

	lexer.ADD:           SUM,
	lexer.SUB:           SUM,

	lexer.MUL:           PRODUCT,
	lexer.DIV:           PRODUCT,
	lexer.MOD:           PRODUCT,

	lexer.POW:           EXPONENT,

	lexer.LPAREN:        CALL,
	lexer.LBRACKET:      CALL,
	lexer.DOT:           MEMBER,
	lexer.OPTIONAL:      OPTIONAL,
	lexer.INCREMENT:     POSTFIX,
	lexer.DECREMENT:     POSTFIX,
}

// peekPrecedence returns the precedence of the peek token.
func (p *Parser) peekPrecedence() Precedence {
	if p, ok := precedences[p.peekToken.Type]; ok {
		return p
	}
	return LOWEST
}

// currentPrecedence returns the precedence of the current token.
func (p *Parser) currentPrecedence() Precedence {
	if p, ok := precedences[p.currentToken.Type]; ok {
		return p
	}
	return LOWEST
}

// ============================================================================
// MAIN PARSING FUNCTIONS
// ============================================================================

// ParseProgram parses the entire program and returns the AST.
func (p *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{
		Body: []ast.Statement{},
	}

	for !p.currentTokenIs(lexer.EOF) {
		stmt := p.parseStatement()
		if stmt != nil {
			program.Body = append(program.Body, stmt)
		}
		p.nextToken()
	}

	return program
}

// parseStatement parses a statement.
func (p *Parser) parseStatement() ast.Statement {
	switch p.currentToken.Type {
	case lexer.LET, lexer.CONST, lexer.VAR:
		return p.parseVariableDeclaration()
	case lexer.FUNCTION:
		// Check if this is a function declaration or function expression
		// If the next token is not an identifier, it's likely a function expression
		if p.peekTokenIs(lexer.LPAREN) {
			return p.parseExpressionStatement()
		}
		return p.parseFunctionDeclaration()
	case lexer.CLASS:
		return p.parseClassDeclaration()
	case lexer.INTERFACE:
		return p.parseInterfaceDeclaration()
	case lexer.TYPE:
		return p.parseTypeAliasDeclaration()
	case lexer.ENUM:
		return p.parseEnumDeclaration()
	case lexer.IF:
		return p.parseIfStatement()
	case lexer.WHILE:
		return p.parseWhileStatement()
	case lexer.FOR:
		return p.parseForStatement()
	case lexer.RETURN:
		return p.parseReturnStatement()
	case lexer.BREAK:
		return p.parseBreakStatement()
	case lexer.CONTINUE:
		return p.parseContinueStatement()
	case lexer.LBRACE:
		return p.parseBlockStatement()
	case lexer.SEMICOLON:
		return p.parseEmptyStatement()
	default:
		return p.parseExpressionStatement()
	}
}

// ============================================================================
// UTILITY FUNCTIONS
// ============================================================================

// parseIdentifier parses an identifier.
func (p *Parser) parseIdentifier() *ast.Identifier {
	if !p.currentTokenIs(lexer.IDENT) {
		p.addErrorf("expected identifier, got %s", p.currentToken.Type)
		return nil
	}

	ident := &ast.Identifier{
		NamePos: p.currentToken.Position,
		Name:    p.currentToken.Literal,
	}

	return ident
}

// parseIntegerLiteral parses an integer literal.
func (p *Parser) parseIntegerLiteral() *ast.IntegerLiteral {
	lit := &ast.IntegerLiteral{
		ValuePos: p.currentToken.Position,
		Raw:      p.currentToken.Literal,
	}

	value, err := strconv.ParseInt(p.currentToken.Literal, 0, 64)
	if err != nil {
		p.addErrorf("could not parse %q as integer", p.currentToken.Literal)
		return nil
	}

	lit.Value = value
	return lit
}

// parseFloatLiteral parses a float literal.
func (p *Parser) parseFloatLiteral() *ast.FloatLiteral {
	lit := &ast.FloatLiteral{
		ValuePos: p.currentToken.Position,
		Raw:      p.currentToken.Literal,
	}

	value, err := strconv.ParseFloat(p.currentToken.Literal, 64)
	if err != nil {
		p.addErrorf("could not parse %q as float", p.currentToken.Literal)
		return nil
	}

	lit.Value = value
	return lit
}

// parseStringLiteral parses a string literal.
func (p *Parser) parseStringLiteral() *ast.StringLiteral {
	return &ast.StringLiteral{
		ValuePos: p.currentToken.Position,
		Value:    p.currentToken.Literal,
		Raw:      p.currentToken.Literal,
	}
}

// parseBooleanLiteral parses a boolean literal.
func (p *Parser) parseBooleanLiteral() *ast.BooleanLiteral {
	return &ast.BooleanLiteral{
		ValuePos: p.currentToken.Position,
		Value:    p.currentToken.Literal == "true",
		Raw:      p.currentToken.Literal,
	}
}

// parseNullLiteral parses a null literal.
func (p *Parser) parseNullLiteral() *ast.NullLiteral {
	return &ast.NullLiteral{
		ValuePos: p.currentToken.Position,
	}
}

// parseUndefinedLiteral parses an undefined literal.
func (p *Parser) parseUndefinedLiteral() *ast.UndefinedLiteral {
	return &ast.UndefinedLiteral{
		ValuePos: p.currentToken.Position,
	}
}

// parseVoidLiteral parses a void literal.
func (p *Parser) parseVoidLiteral() *ast.VoidLiteral {
	return &ast.VoidLiteral{
		ValuePos: p.currentToken.Position,
	}
}