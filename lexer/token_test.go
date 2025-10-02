package lexer

import (
	"testing"
)

func TestTokenString(t *testing.T) {
	tests := []struct {
		token    Token
		expected string
	}{
		{IDENT, "IDENT"},
		{INT, "INT"},
		{FLOAT, "FLOAT"},
		{STRING, "STRING"},
		{ADD, "+"},
		{SUB, "-"},
		{MUL, "*"},
		{DIV, "/"},
		{ASSIGN, "="},
		{EQ, "=="},
		{STRICT_EQ, "==="},
		{IF, "if"},
		{FUNCTION, "function"},
		{CLASS, "class"},
		{INT_T, "int"},
		{FLOAT_T, "float"},
		{ARROW, "=>"},
		{NULLISH, "??"},
		{OPTIONAL, "?."},
	}

	for _, test := range tests {
		if got := test.token.String(); got != test.expected {
			t.Errorf("Token.String() for %v = %q, want %q", test.token, got, test.expected)
		}
	}
}

func TestLookup(t *testing.T) {
	tests := []struct {
		ident    string
		expected Token
	}{
		{"if", IF},
		{"else", ELSE},
		{"function", FUNCTION},
		{"class", CLASS},
		{"let", LET},
		{"const", CONST},
		{"var", VAR},
		{"true", BOOLEAN},
		{"false", BOOLEAN},
		{"null", NULL},
		{"undefined", UNDEFINED},
		{"int", INT_T},
		{"float", FLOAT_T},
		{"string", STRING_T},
		{"number", NUMBER_T},
		{"boolean", BOOLEAN_T},
		{"interface", INTERFACE},
		{"type", TYPE},
		{"public", PUBLIC},
		{"private", PRIVATE},
		{"protected", PROTECTED},
		{"async", ASYNC},
		{"await", AWAIT},
		{"typeof", TYPEOF},
		{"instanceof", INSTANCEOF},
		{"myVariable", IDENT}, // not a keyword
		{"customFunction", IDENT}, // not a keyword
	}

	for _, test := range tests {
		if got := Lookup(test.ident); got != test.expected {
			t.Errorf("Lookup(%q) = %v, want %v", test.ident, got, test.expected)
		}
	}
}

func TestTokenCategories(t *testing.T) {
	// Test IsLiteral
	literalTokens := []Token{IDENT, INT, FLOAT, STRING, BOOLEAN, NULL, UNDEFINED}
	for _, tok := range literalTokens {
		if !tok.IsLiteral() {
			t.Errorf("Token %v should be a literal", tok)
		}
	}

	// Test IsOperator
	operatorTokens := []Token{ADD, SUB, MUL, DIV, EQ, STRICT_EQ, LOGICAL_AND, ASSIGN}
	for _, tok := range operatorTokens {
		if !tok.IsOperator() {
			t.Errorf("Token %v should be an operator", tok)
		}
	}

	// Test IsKeyword
	keywordTokens := []Token{IF, ELSE, FUNCTION, CLASS, LET, CONST, INT_T, FLOAT_T}
	for _, tok := range keywordTokens {
		if !tok.IsKeyword() {
			t.Errorf("Token %v should be a keyword", tok)
		}
	}

	// Test IsDelimiter
	delimiterTokens := []Token{LPAREN, RPAREN, LBRACE, RBRACE, SEMICOLON, COMMA}
	for _, tok := range delimiterTokens {
		if !tok.IsDelimiter() {
			t.Errorf("Token %v should be a delimiter", tok)
		}
	}
}

func TestPrecedence(t *testing.T) {
	tests := []struct {
		token      Token
		precedence int
	}{
		{LOGICAL_OR, 1},
		{LOGICAL_AND, 2},
		{BIT_OR, 3},
		{BIT_XOR, 4},
		{BIT_AND, 5},
		{EQ, 6},
		{STRICT_EQ, 6},
		{LT, 7},
		{GT, 7},
		{ADD, 9},
		{SUB, 9},
		{MUL, 10},
		{DIV, 10},
		{POW, 11},
		{IDENT, 0}, // non-operator should have 0 precedence
	}

	for _, test := range tests {
		if got := test.token.Precedence(); got != test.precedence {
			t.Errorf("Token %v precedence = %d, want %d", test.token, got, test.precedence)
		}
	}
}

func TestIsAssignment(t *testing.T) {
	assignmentTokens := []Token{
		ASSIGN, ADD_ASSIGN, SUB_ASSIGN, MUL_ASSIGN, DIV_ASSIGN,
		MOD_ASSIGN, POW_ASSIGN, BIT_AND_ASSIGN, BIT_OR_ASSIGN,
	}

	for _, tok := range assignmentTokens {
		if !tok.IsAssignment() {
			t.Errorf("Token %v should be an assignment operator", tok)
		}
	}

	nonAssignmentTokens := []Token{ADD, SUB, EQ, STRICT_EQ, IF, IDENT}
	for _, tok := range nonAssignmentTokens {
		if tok.IsAssignment() {
			t.Errorf("Token %v should not be an assignment operator", tok)
		}
	}
}

func TestIsUnaryOperator(t *testing.T) {
	unaryTokens := []Token{ADD, SUB, LOGICAL_NOT, BIT_NOT, INCREMENT, DECREMENT, TYPEOF, DELETE}
	for _, tok := range unaryTokens {
		if !tok.IsUnaryOperator() {
			t.Errorf("Token %v should be a unary operator", tok)
		}
	}

	nonUnaryTokens := []Token{ASSIGN, EQ, IF, IDENT, LPAREN}
	for _, tok := range nonUnaryTokens {
		if tok.IsUnaryOperator() {
			t.Errorf("Token %v should not be a unary operator", tok)
		}
	}
}

func TestPosition(t *testing.T) {
	pos := Position{Line: 10, Column: 5, Offset: 100}
	
	if pos.Line != 10 {
		t.Errorf("Position.Line = %d, want 10", pos.Line)
	}
	if pos.Column != 5 {
		t.Errorf("Position.Column = %d, want 5", pos.Column)
	}
	if pos.Offset != 100 {
		t.Errorf("Position.Offset = %d, want 100", pos.Offset)
	}
}

func TestTokenInfo(t *testing.T) {
	pos := Position{Line: 1, Column: 1, Offset: 0}
	tokenInfo := TokenInfo{
		Type:     IF,
		Literal:  "if",
		Position: pos,
	}

	if tokenInfo.Type != IF {
		t.Errorf("TokenInfo.Type = %v, want %v", tokenInfo.Type, IF)
	}
	if tokenInfo.Literal != "if" {
		t.Errorf("TokenInfo.Literal = %q, want %q", tokenInfo.Literal, "if")
	}
	if tokenInfo.Position != pos {
		t.Errorf("TokenInfo.Position = %v, want %v", tokenInfo.Position, pos)
	}
}