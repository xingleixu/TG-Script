package parser

import (
	"testing"

	"github.com/xingleixu/TG-Script/ast"
	"github.com/xingleixu/TG-Script/lexer"
)

// Helper function to create a parser from source code
func createParser(input string) *Parser {
	l := lexer.New(input)
	return New(l)
}

// Helper function to check parser errors
func checkParserErrors(t *testing.T, p *Parser) {
	errors := p.Errors()
	if len(errors) == 0 {
		return
	}

	t.Errorf("parser has %d errors", len(errors))
	for _, msg := range errors {
		t.Errorf("parser error: %q", msg)
	}
	t.FailNow()
}

func TestLetStatements(t *testing.T) {
	tests := []struct {
		input              string
		expectedIdentifier string
		expectedValue      interface{}
	}{
		{"let x = 5;", "x", 5},
		{"let y = true;", "y", true},
		{"let foobar = y;", "foobar", "y"},
	}

	for _, tt := range tests {
		p := createParser(tt.input)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		if len(program.Body) != 1 {
			t.Fatalf("program.Body does not contain 1 statement. got=%d",
				len(program.Body))
		}

		stmt := program.Body[0]
		if !testLetStatement(t, stmt, tt.expectedIdentifier) {
			return
		}

		val := stmt.(*ast.VariableDeclaration).Declarations[0].Init
		if !testLiteralExpression(t, val, tt.expectedValue) {
			return
		}
	}
}

func testLetStatement(t *testing.T, s ast.Statement, name string) bool {
	if s.String() == "" {
		t.Errorf("s.String() returned empty string")
		return false
	}

	letStmt, ok := s.(*ast.VariableDeclaration)
	if !ok {
		t.Errorf("s not *ast.VariableDeclaration. got=%T", s)
		return false
	}

	if letStmt.Kind != lexer.LET {
		t.Errorf("letStmt.Kind not 'let'. got=%q", letStmt.Kind)
		return false
	}

	if len(letStmt.Declarations) != 1 {
		t.Errorf("letStmt.Declarations length not 1. got=%d", len(letStmt.Declarations))
		return false
	}

	decl := letStmt.Declarations[0]
	if decl.Id.(*ast.Identifier).Name != name {
		t.Errorf("decl.Id.Name not '%s'. got=%s", name, decl.Id.(*ast.Identifier).Name)
		return false
	}

	return true
}

func TestReturnStatements(t *testing.T) {
	tests := []struct {
		input         string
		expectedValue interface{}
	}{
		{"return 5;", 5},
		{"return true;", true},
		{"return foobar;", "foobar"},
	}

	for _, tt := range tests {
		p := createParser(tt.input)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		if len(program.Body) != 1 {
			t.Fatalf("program.Body does not contain 1 statement. got=%d",
				len(program.Body))
		}

		stmt := program.Body[0]
		returnStmt, ok := stmt.(*ast.ReturnStatement)
		if !ok {
			t.Fatalf("stmt not *ast.ReturnStatement. got=%T", stmt)
		}

		if !testLiteralExpression(t, returnStmt.Argument, tt.expectedValue) {
			return
		}
	}
}

func TestIdentifierExpression(t *testing.T) {
	input := "foobar;"

	p := createParser(input)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Body) != 1 {
		t.Fatalf("program has not enough statements. got=%d",
			len(program.Body))
	}

	stmt, ok := program.Body[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Body[0] is not ast.ExpressionStatement. got=%T",
			program.Body[0])
	}

	ident, ok := stmt.Expression.(*ast.Identifier)
	if !ok {
		t.Fatalf("exp not *ast.Identifier. got=%T", stmt.Expression)
	}

	if ident.Name != "foobar" {
		t.Errorf("ident.Name not %s. got=%s", "foobar", ident.Name)
	}
}

func TestIntegerLiteralExpression(t *testing.T) {
	input := "5;"

	p := createParser(input)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Body) != 1 {
		t.Fatalf("program has not enough statements. got=%d",
			len(program.Body))
	}

	stmt, ok := program.Body[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Body[0] is not ast.ExpressionStatement. got=%T",
			program.Body[0])
	}

	literal, ok := stmt.Expression.(*ast.IntegerLiteral)
	if !ok {
		t.Fatalf("exp not *ast.IntegerLiteral. got=%T", stmt.Expression)
	}

	if literal.Value != 5 {
		t.Errorf("literal.Value not %d. got=%d", 5, literal.Value)
	}
}

func TestParsingPrefixExpressions(t *testing.T) {
	prefixTests := []struct {
		input    string
		operator string
		value    interface{}
	}{
		{"!5;", "!", 5},
		{"-15;", "-", 15},
		{"!foobar;", "!", "foobar"},
		{"-foobar;", "-", "foobar"},
		{"!true;", "!", true},
		{"!false;", "!", false},
	}

	for _, tt := range prefixTests {
		p := createParser(tt.input)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		if len(program.Body) != 1 {
			t.Fatalf("program.Body does not contain %d statements. got=%d\n",
				1, len(program.Body))
		}

		stmt, ok := program.Body[0].(*ast.ExpressionStatement)
		if !ok {
			t.Fatalf("program.Body[0] is not ast.ExpressionStatement. got=%T",
				program.Body[0])
		}

		exp, ok := stmt.Expression.(*ast.UnaryExpression)
		if !ok {
			t.Fatalf("stmt is not ast.UnaryExpression. got=%T", stmt.Expression)
		}

		if exp.Operator.String() != tt.operator {
			t.Fatalf("exp.Operator is not '%s'. got=%s",
				tt.operator, exp.Operator)
		}

		if !testLiteralExpression(t, exp.Operand, tt.value) {
			return
		}
	}
}

func TestParsingInfixExpressions(t *testing.T) {
	infixTests := []struct {
		input      string
		leftValue  interface{}
		operator   string
		rightValue interface{}
	}{
		{"5 + 5;", 5, "+", 5},
		{"5 - 5;", 5, "-", 5},
		{"5 * 5;", 5, "*", 5},
		{"5 / 5;", 5, "/", 5},
		{"5 > 5;", 5, ">", 5},
		{"5 < 5;", 5, "<", 5},
		{"5 == 5;", 5, "==", 5},
		{"5 != 5;", 5, "!=", 5},
		{"foobar + barfoo;", "foobar", "+", "barfoo"},
		{"foobar - barfoo;", "foobar", "-", "barfoo"},
		{"foobar * barfoo;", "foobar", "*", "barfoo"},
		{"foobar / barfoo;", "foobar", "/", "barfoo"},
		{"foobar > barfoo;", "foobar", ">", "barfoo"},
		{"foobar < barfoo;", "foobar", "<", "barfoo"},
		{"foobar == barfoo;", "foobar", "==", "barfoo"},
		{"foobar != barfoo;", "foobar", "!=", "barfoo"},
		{"true == true", true, "==", true},
		{"true != false", true, "!=", false},
		{"false == false", false, "==", false},
	}

	for _, tt := range infixTests {
		p := createParser(tt.input)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		if len(program.Body) != 1 {
			t.Fatalf("program.Body does not contain %d statements. got=%d\n",
				1, len(program.Body))
		}

		stmt, ok := program.Body[0].(*ast.ExpressionStatement)
		if !ok {
			t.Fatalf("program.Body[0] is not ast.ExpressionStatement. got=%T",
				program.Body[0])
		}

		if !testInfixExpression(t, stmt.Expression, tt.leftValue,
			tt.operator, tt.rightValue) {
			return
		}
	}
}

func TestOperatorPrecedenceParsing(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			"-a * b",
			"((-a) * b)",
		},
		{
			"!-a",
			"(!(-a))",
		},
		{
			"a + b + c",
			"((a + b) + c)",
		},
		{
			"a + b - c",
			"((a + b) - c)",
		},
		{
			"a * b * c",
			"((a * b) * c)",
		},
		{
			"a * b / c",
			"((a * b) / c)",
		},
		{
			"a + b / c",
			"(a + (b / c))",
		},
		{
			"a + b * c + d / e - f",
			"(((a + (b * c)) + (d / e)) - f)",
		},
		{
			"3 + 4; -5 * 5",
			"(3 + 4)((-5) * 5)",
		},
		{
			"5 > 4 == 3 < 4",
			"((5 > 4) == (3 < 4))",
		},
		{
			"5 < 4 != 3 > 4",
			"((5 < 4) != (3 > 4))",
		},
		{
			"3 + 4 * 5 == 3 * 1 + 4 * 5",
			"((3 + (4 * 5)) == ((3 * 1) + (4 * 5)))",
		},
		{
			"true",
			"true",
		},
		{
			"false",
			"false",
		},
		{
			"3 > 5 == false",
			"((3 > 5) == false)",
		},
		{
			"3 < 5 == true",
			"((3 < 5) == true)",
		},
		{
			"1 + (2 + 3) + 4",
			"((1 + (2 + 3)) + 4)",
		},
		{
			"(5 + 5) * 2",
			"((5 + 5) * 2)",
		},
		{
			"2 / (5 + 5)",
			"(2 / (5 + 5))",
		},
		{
			"(5 + 5) * 2 * (5 + 5)",
			"(((5 + 5) * 2) * (5 + 5))",
		},
		{
			"-(5 + 5)",
			"(-(5 + 5))",
		},
		{
			"!(true == true)",
			"(!(true == true))",
		},
		{
			"a + add(b * c) + d",
			"((a + add((b * c))) + d)",
		},
		{
			"add(a, b, 1, 2 * 3, 4 + 5, add(6, 7 * 8))",
			"add(a, b, 1, (2 * 3), (4 + 5), add(6, (7 * 8)))",
		},
		{
			"add(a + b + c * d / f + g)",
			"add((((a + b) + ((c * d) / f)) + g))",
		},
		{
			"a * [1, 2, 3, 4][b * c] * d",
			"((a * ([1, 2, 3, 4][(b * c)])) * d)",
		},
		{
			"add(a * b[2], b[1], 2 * [1, 2][1])",
			"add((a * (b[2])), (b[1]), (2 * ([1, 2][1])))",
		},
	}

	for _, tt := range tests {
		p := createParser(tt.input)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		actual := program.String()
		if actual != tt.expected {
			t.Errorf("expected=%q, got=%q", tt.expected, actual)
		}
	}
}

func testLiteralExpression(
	t *testing.T,
	exp ast.Expression,
	expected interface{},
) bool {
	switch v := expected.(type) {
	case int:
		return testIntegerLiteral(t, exp, int64(v))
	case int64:
		return testIntegerLiteral(t, exp, v)
	case string:
		return testIdentifier(t, exp, v)
	case bool:
		return testBooleanLiteral(t, exp, v)
	}
	t.Errorf("type of exp not handled. got=%T", exp)
	return false
}

func testInfixExpression(t *testing.T, exp ast.Expression, left interface{},
	operator string, right interface{}) bool {

	opExp, ok := exp.(*ast.BinaryExpression)
	if !ok {
		t.Errorf("exp is not ast.BinaryExpression. got=%T(%s)", exp, exp)
		return false
	}

	if !testLiteralExpression(t, opExp.Left, left) {
		return false
	}

	if opExp.Operator.String() != operator {
		t.Errorf("exp.Operator is not '%s'. got=%q", operator, opExp.Operator)
		return false
	}

	if !testLiteralExpression(t, opExp.Right, right) {
		return false
	}

	return true
}

func testIntegerLiteral(t *testing.T, il ast.Expression, value int64) bool {
	integ, ok := il.(*ast.IntegerLiteral)
	if !ok {
		t.Errorf("il not *ast.IntegerLiteral. got=%T", il)
		return false
	}

	if integ.Value != value {
		t.Errorf("integ.Value not %d. got=%d", value, integ.Value)
		return false
	}

	return true
}

func testIdentifier(t *testing.T, exp ast.Expression, value string) bool {
	ident, ok := exp.(*ast.Identifier)
	if !ok {
		t.Errorf("exp not *ast.Identifier. got=%T", exp)
		return false
	}

	if ident.Name != value {
		t.Errorf("ident.Name not %s. got=%s", value, ident.Name)
		return false
	}

	return true
}

func testBooleanLiteral(t *testing.T, exp ast.Expression, value bool) bool {
	bo, ok := exp.(*ast.BooleanLiteral)
	if !ok {
		t.Errorf("exp not *ast.BooleanLiteral. got=%T", exp)
		return false
	}

	if bo.Value != value {
		t.Errorf("bo.Value not %t. got=%t", value, bo.Value)
		return false
	}

	return true
}

func TestIfExpression(t *testing.T) {
	input := `if (x < y) { x }`

	p := createParser(input)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Body) != 1 {
		t.Fatalf("program.Body does not contain %d statements. got=%d\n",
			1, len(program.Body))
	}

	stmt, ok := program.Body[0].(*ast.IfStatement)
	if !ok {
		t.Fatalf("program.Body[0] is not ast.IfStatement. got=%T",
			program.Body[0])
	}

	if !testInfixExpression(t, stmt.Test, "x", "<", "y") {
		return
	}

	consequence, ok := stmt.Consequent.(*ast.BlockStatement)
	if !ok {
		t.Fatalf("stmt.Consequent is not ast.BlockStatement. got=%T",
			stmt.Consequent)
	}

	if len(consequence.Body) != 1 {
		t.Errorf("consequence is not 1 statement. got=%d\n",
			len(consequence.Body))
	}

	consequenceStmt, ok := consequence.Body[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("Body[0] is not ast.ExpressionStatement. got=%T",
			consequence.Body[0])
	}

	if !testIdentifier(t, consequenceStmt.Expression, "x") {
		return
	}

	if stmt.Alternate != nil {
		t.Errorf("exp.Alternate was not nil. got=%+v", stmt.Alternate)
	}
}

func TestFunctionLiteralParsing(t *testing.T) {
	input := `function(x, y) { x + y; }`

	p := createParser(input)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Body) != 1 {
		t.Fatalf("program.Body does not contain %d statements. got=%d\n",
			1, len(program.Body))
	}

	stmt, ok := program.Body[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Body[0] is not ast.ExpressionStatement. got=%T",
			program.Body[0])
	}

	function, ok := stmt.Expression.(*ast.FunctionExpression)
	if !ok {
		t.Fatalf("stmt.Expression is not ast.FunctionExpression. got=%T",
			stmt.Expression)
	}

	if len(function.Parameters) != 2 {
		t.Fatalf("function literal parameters wrong. want 2, got=%d\n",
			len(function.Parameters))
	}

	testLiteralExpression(t, function.Parameters[0].Name, "x")
	testLiteralExpression(t, function.Parameters[1].Name, "y")

	if len(function.Body.Body) != 1 {
		t.Fatalf("function.Body.Body has not 1 statement. got=%d\n",
			len(function.Body.Body))
	}

	bodyStmt, ok := function.Body.Body[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("function body stmt is not ast.ExpressionStatement. got=%T",
			function.Body.Body[0])
	}

	testInfixExpression(t, bodyStmt.Expression, "x", "+", "y")
}

func TestCallExpressionParsing(t *testing.T) {
	input := "add(1, 2 * 3, 4 + 5);"

	p := createParser(input)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Body) != 1 {
		t.Fatalf("program.Body does not contain %d statements. got=%d\n",
			1, len(program.Body))
	}

	stmt, ok := program.Body[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("stmt is not ast.ExpressionStatement. got=%T",
			program.Body[0])
	}

	exp, ok := stmt.Expression.(*ast.CallExpression)
	if !ok {
		t.Fatalf("stmt.Expression is not ast.CallExpression. got=%T",
			stmt.Expression)
	}

	if !testIdentifier(t, exp.Callee, "add") {
		return
	}

	if len(exp.Arguments) != 3 {
		t.Fatalf("wrong length of arguments. got=%d", len(exp.Arguments))
	}

	testLiteralExpression(t, exp.Arguments[0], 1)
	testInfixExpression(t, exp.Arguments[1], 2, "*", 3)
	testInfixExpression(t, exp.Arguments[2], 4, "+", 5)
}

func TestStringLiteralExpression(t *testing.T) {
	input := `"hello world";`

	p := createParser(input)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	stmt := program.Body[0].(*ast.ExpressionStatement)
	literal, ok := stmt.Expression.(*ast.StringLiteral)
	if !ok {
		t.Fatalf("exp not *ast.StringLiteral. got=%T", stmt.Expression)
	}

	if literal.Value != "hello world" {
		t.Errorf("literal.Value not %q. got=%q", "hello world", literal.Value)
	}
}

func TestArrayLiterals(t *testing.T) {
	input := "[1, 2 * 2, 3 + 3]"

	p := createParser(input)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	stmt, ok := program.Body[0].(*ast.ExpressionStatement)
	array, ok := stmt.Expression.(*ast.ArrayLiteral)
	if !ok {
		t.Fatalf("exp not ast.ArrayLiteral. got=%T", stmt.Expression)
	}

	if len(array.Elements) != 3 {
		t.Fatalf("len(array.Elements) not 3. got=%d", len(array.Elements))
	}

	testIntegerLiteral(t, array.Elements[0], 1)
	testInfixExpression(t, array.Elements[1], 2, "*", 2)
	testInfixExpression(t, array.Elements[2], 3, "+", 3)
}

func TestParsingIndexExpressions(t *testing.T) {
	input := "myArray[1 + 1]"

	p := createParser(input)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	stmt, ok := program.Body[0].(*ast.ExpressionStatement)
	indexExp, ok := stmt.Expression.(*ast.MemberExpression)
	if !ok {
		t.Fatalf("exp not *ast.MemberExpression. got=%T", stmt.Expression)
	}

	if !testIdentifier(t, indexExp.Object, "myArray") {
		return
	}

	if !testInfixExpression(t, indexExp.Property, 1, "+", 1) {
		return
	}
}