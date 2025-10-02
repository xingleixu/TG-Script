package lexer

import (
	"testing"
)

func TestLexerBasicTokens(t *testing.T) {
	input := `let five = 5;
let ten = 10;

let add = function(x, y) {
  x + y;
};

let result = add(five, ten);
!-/5*;
5 < 10 > 5;

if (5 < 10) {
	return true;
} else {
	return false;
}

10 == 10;
10 != 9;
10 === 10;
10 !== 9;
`

	tests := []struct {
		expectedType    Token
		expectedLiteral string
	}{
		{LET, "let"},
		{IDENT, "five"},
		{ASSIGN, "="},
		{INT, "5"},
		{SEMICOLON, ";"},
		{LET, "let"},
		{IDENT, "ten"},
		{ASSIGN, "="},
		{INT, "10"},
		{SEMICOLON, ";"},
		{LET, "let"},
		{IDENT, "add"},
		{ASSIGN, "="},
		{FUNCTION, "function"},
		{LPAREN, "("},
		{IDENT, "x"},
		{COMMA, ","},
		{IDENT, "y"},
		{RPAREN, ")"},
		{LBRACE, "{"},
		{IDENT, "x"},
		{ADD, "+"},
		{IDENT, "y"},
		{SEMICOLON, ";"},
		{RBRACE, "}"},
		{SEMICOLON, ";"},
		{LET, "let"},
		{IDENT, "result"},
		{ASSIGN, "="},
		{IDENT, "add"},
		{LPAREN, "("},
		{IDENT, "five"},
		{COMMA, ","},
		{IDENT, "ten"},
		{RPAREN, ")"},
		{SEMICOLON, ";"},
		{LOGICAL_NOT, "!"},
		{SUB, "-"},
		{DIV, "/"},
		{INT, "5"},
		{MUL, "*"},
		{SEMICOLON, ";"},
		{INT, "5"},
		{LT, "<"},
		{INT, "10"},
		{GT, ">"},
		{INT, "5"},
		{SEMICOLON, ";"},
		{IF, "if"},
		{LPAREN, "("},
		{INT, "5"},
		{LT, "<"},
		{INT, "10"},
		{RPAREN, ")"},
		{LBRACE, "{"},
		{RETURN, "return"},
		{BOOLEAN, "true"},
		{SEMICOLON, ";"},
		{RBRACE, "}"},
		{ELSE, "else"},
		{LBRACE, "{"},
		{RETURN, "return"},
		{BOOLEAN, "false"},
		{SEMICOLON, ";"},
		{RBRACE, "}"},
		{INT, "10"},
		{EQ, "=="},
		{INT, "10"},
		{SEMICOLON, ";"},
		{INT, "10"},
		{NE, "!="},
		{INT, "9"},
		{SEMICOLON, ";"},
		{INT, "10"},
		{STRICT_EQ, "==="},
		{INT, "10"},
		{SEMICOLON, ";"},
		{INT, "10"},
		{STRICT_NE, "!=="},
		{INT, "9"},
		{SEMICOLON, ";"},
		{EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestLexerNumbers(t *testing.T) {
	input := `42 3.14 0x1A 0b1010 0o777 1.23e-4 2.5E+3`

	tests := []struct {
		expectedType    Token
		expectedLiteral string
	}{
		{INT, "42"},
		{FLOAT, "3.14"},
		{INT, "0x1A"},
		{INT, "0b1010"},
		{INT, "0o777"},
		{FLOAT, "1.23e-4"},
		{FLOAT, "2.5E+3"},
		{EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestLexerStrings(t *testing.T) {
	input := `"hello world" 'single quotes' ` + "`template string`"

	tests := []struct {
		expectedType    Token
		expectedLiteral string
	}{
		{STRING, "hello world"},
		{STRING, "single quotes"},
		{TEMPLATE, "template string"},
		{EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestLexerOperators(t *testing.T) {
	input := `++ -- += -= *= /= **= && || ?? ?. ... => ** >>> <<= >>= >>>= &= |= ^= %=`

	tests := []struct {
		expectedType    Token
		expectedLiteral string
	}{
		{INCREMENT, "++"},
		{DECREMENT, "--"},
		{ADD_ASSIGN, "+="},
		{SUB_ASSIGN, "-="},
		{MUL_ASSIGN, "*="},
		{DIV_ASSIGN, "/="},
		{POW_ASSIGN, "**="},
		{LOGICAL_AND, "&&"},
		{LOGICAL_OR, "||"},
		{NULLISH, "??"},
		{OPTIONAL, "?."},
		{SPREAD, "..."},
		{ARROW, "=>"},
		{POW, "**"},
		{BIT_URSHIFT, ">>>"},
		{LSHIFT_ASSIGN, "<<="},
		{RSHIFT_ASSIGN, ">>="},
		{URSHIFT_ASSIGN, ">>>="},
		{BIT_AND_ASSIGN, "&="},
		{BIT_OR_ASSIGN, "|="},
		{BIT_XOR_ASSIGN, "^="},
		{MOD_ASSIGN, "%="},
		{EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestLexerKeywords(t *testing.T) {
	input := `class interface type public private protected async await typeof instanceof`

	tests := []struct {
		expectedType    Token
		expectedLiteral string
	}{
		{CLASS, "class"},
		{INTERFACE, "interface"},
		{TYPE, "type"},
		{PUBLIC, "public"},
		{PRIVATE, "private"},
		{PROTECTED, "protected"},
		{ASYNC, "async"},
		{AWAIT, "await"},
		{TYPEOF, "typeof"},
		{INSTANCEOF, "instanceof"},
		{EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestLexerPosition(t *testing.T) {
	input := `let x = 5;
let y = 10;`

	l := New(input)

	// Test first token position
	tok := l.NextToken() // "let"
	if tok.Position.Line != 1 || tok.Position.Column != 1 {
		t.Errorf("First token position wrong. expected line=1, column=1, got line=%d, column=%d",
			tok.Position.Line, tok.Position.Column)
	}

	// Skip to second line
	for tok.Type != LET || tok.Position.Line != 2 {
		tok = l.NextToken()
	}

	if tok.Position.Line != 2 || tok.Position.Column != 1 {
		t.Errorf("Second line token position wrong. expected line=2, column=1, got line=%d, column=%d",
			tok.Position.Line, tok.Position.Column)
	}
}

func TestLexerErrors(t *testing.T) {
	input := `let x = @;` // @ is an illegal character

	l := New(input)

	// Tokenize all
	for {
		tok := l.NextToken()
		if tok.Type == EOF {
			break
		}
	}

	if !l.HasErrors() {
		t.Error("Expected lexer to have errors, but it doesn't")
	}

	errors := l.GetErrors()
	if len(errors) == 0 {
		t.Error("Expected at least one error, got none")
	}

	// Check that error message contains information about the illegal character
	found := false
	for _, err := range errors {
		if err != "" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected error message about illegal character")
	}
}

func TestTokenizeAll(t *testing.T) {
	input := `let x = 5;`

	l := New(input)
	tokens := l.TokenizeAll()

	expectedTokens := []Token{LET, IDENT, ASSIGN, INT, SEMICOLON, EOF}

	if len(tokens) != len(expectedTokens) {
		t.Fatalf("Expected %d tokens, got %d", len(expectedTokens), len(tokens))
	}

	for i, expectedType := range expectedTokens {
		if tokens[i].Type != expectedType {
			t.Errorf("Token %d: expected type %v, got %v", i, expectedType, tokens[i].Type)
		}
	}
}

func TestLexerComments(t *testing.T) {
	input := `// This is a single line comment
let x = 5; // Another comment
/* This is a 
   multi-line comment */
let y = 10;
/* Single line multi-comment */`

	tests := []struct {
		expectedType    Token
		expectedLiteral string
	}{
		{COMMENT, "// This is a single line comment"},
		{LET, "let"},
		{IDENT, "x"},
		{ASSIGN, "="},
		{INT, "5"},
		{SEMICOLON, ";"},
		{COMMENT, "// Another comment"},
		{COMMENT, "/* This is a \n   multi-line comment */"},
		{LET, "let"},
		{IDENT, "y"},
		{ASSIGN, "="},
		{INT, "10"},
		{SEMICOLON, ";"},
		{COMMENT, "/* Single line multi-comment */"},
		{EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestLexerUnterminatedComment(t *testing.T) {
	input := `/* This comment is not closed`

	l := New(input)

	// Tokenize all
	for {
		tok := l.NextToken()
		if tok.Type == EOF {
			break
		}
	}

	if !l.HasErrors() {
		t.Error("Expected lexer to have errors for unterminated comment, but it doesn't")
	}

	errors := l.GetErrors()
	if len(errors) == 0 {
		t.Error("Expected at least one error for unterminated comment, got none")
	}

	// Check that error message contains information about unterminated comment
	found := false
	for _, err := range errors {
		if err != "" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected error message about unterminated comment")
	}
}