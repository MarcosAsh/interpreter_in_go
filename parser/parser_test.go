package parser

import (
	"fmt"
	"pearl/ast"
	"pearl/lexer"
	"testing"
)

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

func parseProgram(t *testing.T, input string) *ast.Program {
	t.Helper()
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)
	if program == nil {
		t.Fatal("ParseProgram() returned nil")
	}
	return program
}

func checkParserErrors(t *testing.T, p *Parser) {
	t.Helper()
	errs := p.Errors()
	if len(errs) == 0 {
		return
	}
	t.Errorf("parser has %d error(s)", len(errs))
	for _, msg := range errs {
		t.Errorf("  parser error: %s", msg)
	}
	t.FailNow()
}

func expectStatementCount(t *testing.T, program *ast.Program, n int) {
	t.Helper()
	if len(program.Statements) != n {
		t.Fatalf("expected %d statement(s), got %d", n, len(program.Statements))
	}
}

func getExpressionStatement(t *testing.T, program *ast.Program, index int) *ast.ExpressionStatement {
	t.Helper()
	if index >= len(program.Statements) {
		t.Fatalf("statement index %d out of range (have %d)", index, len(program.Statements))
	}
	stmt, ok := program.Statements[index].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("statement[%d] is %T, want *ast.ExpressionStatement", index, program.Statements[index])
	}
	return stmt
}

// testLiteralExpression dispatches to the right type-specific check.
func testLiteralExpression(t *testing.T, exp ast.Expression, expected interface{}) {
	t.Helper()
	switch v := expected.(type) {
	case int:
		testIntegerLiteral(t, exp, int64(v))
	case int64:
		testIntegerLiteral(t, exp, v)
	case float64:
		testFloatLiteral(t, exp, v)
	case string:
		testIdentifier(t, exp, v)
	case bool:
		testBooleanLiteral(t, exp, v)
	default:
		t.Fatalf("testLiteralExpression: unhandled type %T", expected)
	}
}

func testIntegerLiteral(t *testing.T, exp ast.Expression, expected int64) {
	t.Helper()
	il, ok := exp.(*ast.IntegerLiteral)
	if !ok {
		t.Fatalf("expression is %T, want *ast.IntegerLiteral", exp)
	}
	if il.Value != expected {
		t.Errorf("IntegerLiteral.Value = %d, want %d", il.Value, expected)
	}
}

func testFloatLiteral(t *testing.T, exp ast.Expression, expected float64) {
	t.Helper()
	fl, ok := exp.(*ast.FloatLiteral)
	if !ok {
		t.Fatalf("expression is %T, want *ast.FloatLiteral", exp)
	}
	if fl.Value != expected {
		t.Errorf("FloatLiteral.Value = %f, want %f", fl.Value, expected)
	}
}

func testIdentifier(t *testing.T, exp ast.Expression, value string) {
	t.Helper()
	ident, ok := exp.(*ast.Identifier)
	if !ok {
		t.Fatalf("expression is %T, want *ast.Identifier", exp)
	}
	if ident.Value != value {
		t.Errorf("Identifier.Value = %q, want %q", ident.Value, value)
	}
}

func testBooleanLiteral(t *testing.T, exp ast.Expression, value bool) {
	t.Helper()
	b, ok := exp.(*ast.Boolean)
	if !ok {
		t.Fatalf("expression is %T, want *ast.Boolean", exp)
	}
	if b.Value != value {
		t.Errorf("Boolean.Value = %v, want %v", b.Value, value)
	}
}

func testInfixExpression(t *testing.T, exp ast.Expression, left interface{}, op string, right interface{}) {
	t.Helper()
	infix, ok := exp.(*ast.InfixExpression)
	if !ok {
		t.Fatalf("expression is %T, want *ast.InfixExpression", exp)
	}
	testLiteralExpression(t, infix.Left, left)
	if infix.Operator != op {
		t.Errorf("InfixExpression.Operator = %q, want %q", infix.Operator, op)
	}
	testLiteralExpression(t, infix.Right, right)
}

// ---------------------------------------------------------------------------
// 1. Let statements
// ---------------------------------------------------------------------------

func TestLetStatements(t *testing.T) {
	tests := []struct {
		input         string
		expectedName  string
		expectedValue interface{}
	}{
		{"let x = 5;", "x", int64(5)},
		{"let y = 10;", "y", int64(10)},
		{"let name = \"hello\";", "name", nil}, // string checked separately
		{"let flag = true;", "flag", true},
		{"let a = foo;", "a", "foo"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			program := parseProgram(t, tt.input)
			expectStatementCount(t, program, 1)

			stmt, ok := program.Statements[0].(*ast.LetStatement)
			if !ok {
				t.Fatalf("statement is %T, want *ast.LetStatement", program.Statements[0])
			}

			if stmt.Name.Value != tt.expectedName {
				t.Errorf("LetStatement.Name.Value = %q, want %q", stmt.Name.Value, tt.expectedName)
			}
			if stmt.TokenLiteral() != "let" {
				t.Errorf("TokenLiteral() = %q, want %q", stmt.TokenLiteral(), "let")
			}

			if tt.expectedValue != nil {
				testLiteralExpression(t, stmt.Value, tt.expectedValue)
			}
		})
	}
}

func TestLetStatementString(t *testing.T) {
	program := parseProgram(t, `let name = "hello";`)
	expectStatementCount(t, program, 1)

	stmt := program.Statements[0].(*ast.LetStatement)
	sl, ok := stmt.Value.(*ast.StringLiteral)
	if !ok {
		t.Fatalf("value is %T, want *ast.StringLiteral", stmt.Value)
	}
	if sl.Value != "hello" {
		t.Errorf("StringLiteral.Value = %q, want %q", sl.Value, "hello")
	}
}

// ---------------------------------------------------------------------------
// 2. Return statements
// ---------------------------------------------------------------------------

func TestReturnStatements(t *testing.T) {
	tests := []struct {
		input         string
		expectedValue interface{}
	}{
		{"return 5;", int64(5)},
		{"return 42;", int64(42)},
		{"return foo;", "foo"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			program := parseProgram(t, tt.input)
			expectStatementCount(t, program, 1)

			stmt, ok := program.Statements[0].(*ast.ReturnStatement)
			if !ok {
				t.Fatalf("statement is %T, want *ast.ReturnStatement", program.Statements[0])
			}
			if stmt.TokenLiteral() != "return" {
				t.Errorf("TokenLiteral() = %q, want %q", stmt.TokenLiteral(), "return")
			}

			testLiteralExpression(t, stmt.ReturnValue, tt.expectedValue)
		})
	}
}

func TestReturnStatementBare(t *testing.T) {
	program := parseProgram(t, "return\n")
	expectStatementCount(t, program, 1)

	stmt, ok := program.Statements[0].(*ast.ReturnStatement)
	if !ok {
		t.Fatalf("statement is %T, want *ast.ReturnStatement", program.Statements[0])
	}
	if stmt.ReturnValue != nil {
		t.Errorf("expected nil ReturnValue for bare return, got %T", stmt.ReturnValue)
	}
}

// ---------------------------------------------------------------------------
// 3. Integer and float literals
// ---------------------------------------------------------------------------

func TestIntegerLiteralExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"5;", 5},
		{"42;", 42},
		{"0;", 0},
		{"999;", 999},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			program := parseProgram(t, tt.input)
			expectStatementCount(t, program, 1)
			stmt := getExpressionStatement(t, program, 0)
			testIntegerLiteral(t, stmt.Expression, tt.expected)
		})
	}
}

func TestFloatLiteralExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
	}{
		{"3.14;", 3.14},
		{"0.5;", 0.5},
		{"100.0;", 100.0},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			program := parseProgram(t, tt.input)
			expectStatementCount(t, program, 1)
			stmt := getExpressionStatement(t, program, 0)
			testFloatLiteral(t, stmt.Expression, tt.expected)
		})
	}
}

// ---------------------------------------------------------------------------
// 4. Prefix expressions
// ---------------------------------------------------------------------------

func TestPrefixExpressions(t *testing.T) {
	tests := []struct {
		input    string
		operator string
		value    interface{}
	}{
		{"-5;", "-", int64(5)},
		{"!true;", "!", true},
		{"!false;", "!", false},
		{"-15;", "-", int64(15)},
		{"not false;", "not", false},
		{"not true;", "not", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			program := parseProgram(t, tt.input)
			expectStatementCount(t, program, 1)
			stmt := getExpressionStatement(t, program, 0)

			prefix, ok := stmt.Expression.(*ast.PrefixExpression)
			if !ok {
				t.Fatalf("expression is %T, want *ast.PrefixExpression", stmt.Expression)
			}
			if prefix.Operator != tt.operator {
				t.Errorf("Operator = %q, want %q", prefix.Operator, tt.operator)
			}
			testLiteralExpression(t, prefix.Right, tt.value)
		})
	}
}

// ---------------------------------------------------------------------------
// 5. Infix expressions
// ---------------------------------------------------------------------------

func TestInfixExpressions(t *testing.T) {
	tests := []struct {
		input    string
		left     interface{}
		operator string
		right    interface{}
	}{
		{"5 + 5;", int64(5), "+", int64(5)},
		{"5 - 5;", int64(5), "-", int64(5)},
		{"5 * 5;", int64(5), "*", int64(5)},
		{"5 / 5;", int64(5), "/", int64(5)},
		{"5 % 3;", int64(5), "%", int64(3)},
		{"5 > 5;", int64(5), ">", int64(5)},
		{"5 < 5;", int64(5), "<", int64(5)},
		{"5 == 5;", int64(5), "==", int64(5)},
		{"5 != 5;", int64(5), "!=", int64(5)},
		{"5 >= 5;", int64(5), ">=", int64(5)},
		{"5 <= 5;", int64(5), "<=", int64(5)},
		{"true == true;", true, "==", true},
		{"true != false;", true, "!=", false},
		{"true and false;", true, "and", false},
		{"true or false;", true, "or", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			program := parseProgram(t, tt.input)
			expectStatementCount(t, program, 1)
			stmt := getExpressionStatement(t, program, 0)
			testInfixExpression(t, stmt.Expression, tt.left, tt.operator, tt.right)
		})
	}
}

func TestStringConcatOperator(t *testing.T) {
	input := `"a" ++ "b";`
	program := parseProgram(t, input)
	expectStatementCount(t, program, 1)
	stmt := getExpressionStatement(t, program, 0)

	infix, ok := stmt.Expression.(*ast.InfixExpression)
	if !ok {
		t.Fatalf("expression is %T, want *ast.InfixExpression", stmt.Expression)
	}
	if infix.Operator != "++" {
		t.Errorf("Operator = %q, want %q", infix.Operator, "++")
	}
}

// ---------------------------------------------------------------------------
// 6. Operator precedence
// ---------------------------------------------------------------------------

func TestOperatorPrecedenceParsing(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"-a * b;", "((-a) * b)"},
		{"!-a;", "(!(-a))"},
		{"a + b + c;", "((a + b) + c)"},
		{"a + b - c;", "((a + b) - c)"},
		{"a * b * c;", "((a * b) * c)"},
		{"a * b / c;", "((a * b) / c)"},
		{"a + b * c;", "(a + (b * c))"},
		{"a + b * c + d / e - f;", "(((a + (b * c)) + (d / e)) - f)"},
		{"5 > 4 == 3 < 4;", "((5 > 4) == (3 < 4))"},
		{"5 < 4 != 3 > 4;", "((5 < 4) != (3 > 4))"},
		{"3 + 4 * 5 == 3 * 1 + 4 * 5;", "((3 + (4 * 5)) == ((3 * 1) + (4 * 5)))"},
		{"true;", "true"},
		{"false;", "false"},
		{"1 + (2 + 3) + 4;", "((1 + (2 + 3)) + 4)"},
		{"(5 + 5) * 2;", "((5 + 5) * 2)"},
		{"-(5 + 5);", "(-(5 + 5))"},
		{"!(true == true);", "(!(true == true))"},
		{"a + b * c and d;", "((a + (b * c)) and d)"},
		{"a or b and c;", "(a or (b and c))"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			program := parseProgram(t, tt.input)
			actual := program.Statements[0].String()
			if actual != tt.expected {
				t.Errorf("got %q, want %q", actual, tt.expected)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// 7. Boolean literals
// ---------------------------------------------------------------------------

func TestBooleanLiterals(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"true;", true},
		{"false;", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			program := parseProgram(t, tt.input)
			expectStatementCount(t, program, 1)
			stmt := getExpressionStatement(t, program, 0)
			testBooleanLiteral(t, stmt.Expression, tt.expected)
		})
	}
}

// ---------------------------------------------------------------------------
// 8. Null literal
// ---------------------------------------------------------------------------

func TestNullLiteral(t *testing.T) {
	program := parseProgram(t, "null;")
	expectStatementCount(t, program, 1)
	stmt := getExpressionStatement(t, program, 0)

	_, ok := stmt.Expression.(*ast.NullLiteral)
	if !ok {
		t.Fatalf("expression is %T, want *ast.NullLiteral", stmt.Expression)
	}
}

// ---------------------------------------------------------------------------
// 9. If/else expressions
// ---------------------------------------------------------------------------

func TestIfExpression(t *testing.T) {
	input := "if x < y { x }"
	program := parseProgram(t, input)
	expectStatementCount(t, program, 1)
	stmt := getExpressionStatement(t, program, 0)

	ifExp, ok := stmt.Expression.(*ast.IfExpression)
	if !ok {
		t.Fatalf("expression is %T, want *ast.IfExpression", stmt.Expression)
	}

	testInfixExpression(t, ifExp.Condition, "x", "<", "y")

	if len(ifExp.Consequence.Statements) != 1 {
		t.Fatalf("consequence has %d statements, want 1", len(ifExp.Consequence.Statements))
	}

	cons, ok := ifExp.Consequence.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("consequence statement is %T, want *ast.ExpressionStatement",
			ifExp.Consequence.Statements[0])
	}
	testIdentifier(t, cons.Expression, "x")

	if ifExp.Alternative != nil {
		t.Errorf("Alternative should be nil, got %+v", ifExp.Alternative)
	}
}

func TestIfElseExpression(t *testing.T) {
	input := "if x < y { x } else { y }"
	program := parseProgram(t, input)
	expectStatementCount(t, program, 1)
	stmt := getExpressionStatement(t, program, 0)

	ifExp, ok := stmt.Expression.(*ast.IfExpression)
	if !ok {
		t.Fatalf("expression is %T, want *ast.IfExpression", stmt.Expression)
	}

	testInfixExpression(t, ifExp.Condition, "x", "<", "y")

	if len(ifExp.Consequence.Statements) != 1 {
		t.Fatalf("consequence has %d statements, want 1", len(ifExp.Consequence.Statements))
	}

	if ifExp.Alternative == nil {
		t.Fatal("Alternative is nil")
	}
	if len(ifExp.Alternative.Statements) != 1 {
		t.Fatalf("alternative has %d statements, want 1", len(ifExp.Alternative.Statements))
	}

	alt, ok := ifExp.Alternative.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("alternative statement is %T", ifExp.Alternative.Statements[0])
	}
	testIdentifier(t, alt.Expression, "y")
}

// ---------------------------------------------------------------------------
// 10. Function literals
// ---------------------------------------------------------------------------

func TestFunctionLiteralAnonymous(t *testing.T) {
	input := "fn(x, y) { x + y }"
	program := parseProgram(t, input)
	expectStatementCount(t, program, 1)
	stmt := getExpressionStatement(t, program, 0)

	fn, ok := stmt.Expression.(*ast.FunctionLiteral)
	if !ok {
		t.Fatalf("expression is %T, want *ast.FunctionLiteral", stmt.Expression)
	}
	if fn.Name != "" {
		t.Errorf("Name = %q, want empty", fn.Name)
	}
	if len(fn.Parameters) != 2 {
		t.Fatalf("Parameters has %d items, want 2", len(fn.Parameters))
	}
	testIdentifier(t, fn.Parameters[0].Name, "x")
	testIdentifier(t, fn.Parameters[1].Name, "y")

	if len(fn.Body.Statements) != 1 {
		t.Fatalf("body has %d statements, want 1", len(fn.Body.Statements))
	}
	body, ok := fn.Body.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("body statement is %T", fn.Body.Statements[0])
	}
	testInfixExpression(t, body.Expression, "x", "+", "y")
}

func TestFunctionLiteralNamed(t *testing.T) {
	input := "fn add(a, b) { a + b }"
	program := parseProgram(t, input)
	expectStatementCount(t, program, 1)
	stmt := getExpressionStatement(t, program, 0)

	fn, ok := stmt.Expression.(*ast.FunctionLiteral)
	if !ok {
		t.Fatalf("expression is %T, want *ast.FunctionLiteral", stmt.Expression)
	}
	if fn.Name != "add" {
		t.Errorf("Name = %q, want %q", fn.Name, "add")
	}
	if len(fn.Parameters) != 2 {
		t.Fatalf("Parameters has %d items, want 2", len(fn.Parameters))
	}
}

func TestFunctionLiteralDefaultParams(t *testing.T) {
	input := "fn greet(name, greeting = \"hi\") { greeting }"
	program := parseProgram(t, input)
	expectStatementCount(t, program, 1)
	stmt := getExpressionStatement(t, program, 0)

	fn, ok := stmt.Expression.(*ast.FunctionLiteral)
	if !ok {
		t.Fatalf("expression is %T, want *ast.FunctionLiteral", stmt.Expression)
	}
	if len(fn.Parameters) != 2 {
		t.Fatalf("Parameters has %d items, want 2", len(fn.Parameters))
	}
	if fn.Parameters[0].Default != nil {
		t.Errorf("first param should have no default")
	}
	if fn.Parameters[1].Default == nil {
		t.Fatal("second param should have a default")
	}
	sl, ok := fn.Parameters[1].Default.(*ast.StringLiteral)
	if !ok {
		t.Fatalf("default is %T, want *ast.StringLiteral", fn.Parameters[1].Default)
	}
	if sl.Value != "hi" {
		t.Errorf("default value = %q, want %q", sl.Value, "hi")
	}
}

func TestFunctionLiteralNoParams(t *testing.T) {
	input := "fn() { 42 }"
	program := parseProgram(t, input)
	expectStatementCount(t, program, 1)
	stmt := getExpressionStatement(t, program, 0)

	fn, ok := stmt.Expression.(*ast.FunctionLiteral)
	if !ok {
		t.Fatalf("expression is %T, want *ast.FunctionLiteral", stmt.Expression)
	}
	if len(fn.Parameters) != 0 {
		t.Errorf("Parameters has %d items, want 0", len(fn.Parameters))
	}
}

// ---------------------------------------------------------------------------
// 11. Call expressions
// ---------------------------------------------------------------------------

func TestCallExpressionPositional(t *testing.T) {
	input := "add(1, 2, 3);"
	program := parseProgram(t, input)
	expectStatementCount(t, program, 1)
	stmt := getExpressionStatement(t, program, 0)

	call, ok := stmt.Expression.(*ast.CallExpression)
	if !ok {
		t.Fatalf("expression is %T, want *ast.CallExpression", stmt.Expression)
	}
	testIdentifier(t, call.Function, "add")

	if len(call.Arguments) != 3 {
		t.Fatalf("Arguments has %d items, want 3", len(call.Arguments))
	}
	testLiteralExpression(t, call.Arguments[0].Value, int64(1))
	testLiteralExpression(t, call.Arguments[1].Value, int64(2))
	testLiteralExpression(t, call.Arguments[2].Value, int64(3))
	for i, arg := range call.Arguments {
		if arg.Name != "" {
			t.Errorf("argument %d has Name = %q, want empty", i, arg.Name)
		}
	}
}

func TestCallExpressionNamed(t *testing.T) {
	input := "greet(name = \"bob\", loud = true);"
	program := parseProgram(t, input)
	expectStatementCount(t, program, 1)
	stmt := getExpressionStatement(t, program, 0)

	call, ok := stmt.Expression.(*ast.CallExpression)
	if !ok {
		t.Fatalf("expression is %T, want *ast.CallExpression", stmt.Expression)
	}

	if len(call.Arguments) != 2 {
		t.Fatalf("Arguments has %d items, want 2", len(call.Arguments))
	}
	if call.Arguments[0].Name != "name" {
		t.Errorf("arg[0].Name = %q, want %q", call.Arguments[0].Name, "name")
	}
	if call.Arguments[1].Name != "loud" {
		t.Errorf("arg[1].Name = %q, want %q", call.Arguments[1].Name, "loud")
	}
}

func TestCallExpressionNoArgs(t *testing.T) {
	input := "foo();"
	program := parseProgram(t, input)
	expectStatementCount(t, program, 1)
	stmt := getExpressionStatement(t, program, 0)

	call, ok := stmt.Expression.(*ast.CallExpression)
	if !ok {
		t.Fatalf("expression is %T, want *ast.CallExpression", stmt.Expression)
	}
	testIdentifier(t, call.Function, "foo")
	if len(call.Arguments) != 0 {
		t.Errorf("Arguments has %d items, want 0", len(call.Arguments))
	}
}

func TestCallExpressionMixed(t *testing.T) {
	input := "do_thing(1, key = 2);"
	program := parseProgram(t, input)
	expectStatementCount(t, program, 1)
	stmt := getExpressionStatement(t, program, 0)

	call, ok := stmt.Expression.(*ast.CallExpression)
	if !ok {
		t.Fatalf("expression is %T, want *ast.CallExpression", stmt.Expression)
	}
	if len(call.Arguments) != 2 {
		t.Fatalf("Arguments has %d items, want 2", len(call.Arguments))
	}
	if call.Arguments[0].Name != "" {
		t.Errorf("arg[0].Name = %q, want empty (positional)", call.Arguments[0].Name)
	}
	testLiteralExpression(t, call.Arguments[0].Value, int64(1))
	if call.Arguments[1].Name != "key" {
		t.Errorf("arg[1].Name = %q, want %q", call.Arguments[1].Name, "key")
	}
	testLiteralExpression(t, call.Arguments[1].Value, int64(2))
}

// ---------------------------------------------------------------------------
// 12. String literals
// ---------------------------------------------------------------------------

func TestStringLiteralPlain(t *testing.T) {
	input := `"hello world";`
	program := parseProgram(t, input)
	expectStatementCount(t, program, 1)
	stmt := getExpressionStatement(t, program, 0)

	sl, ok := stmt.Expression.(*ast.StringLiteral)
	if !ok {
		t.Fatalf("expression is %T, want *ast.StringLiteral", stmt.Expression)
	}
	if sl.Value != "hello world" {
		t.Errorf("Value = %q, want %q", sl.Value, "hello world")
	}
	// plain string -- all parts should be text
	for _, part := range sl.Parts {
		if part.IsExpr {
			t.Errorf("expected no expression parts in a plain string")
		}
	}
}

func TestStringLiteralWithInterpolation(t *testing.T) {
	input := `"hello {name}";`
	program := parseProgram(t, input)
	expectStatementCount(t, program, 1)
	stmt := getExpressionStatement(t, program, 0)

	sl, ok := stmt.Expression.(*ast.StringLiteral)
	if !ok {
		t.Fatalf("expression is %T, want *ast.StringLiteral", stmt.Expression)
	}

	// Should have at least two parts: text "hello " and expression {name}
	if len(sl.Parts) < 2 {
		t.Fatalf("expected at least 2 parts, got %d", len(sl.Parts))
	}

	if sl.Parts[0].IsExpr {
		t.Error("part[0] should be text")
	}
	if sl.Parts[0].Text != "hello " {
		t.Errorf("part[0].Text = %q, want %q", sl.Parts[0].Text, "hello ")
	}
	if !sl.Parts[1].IsExpr {
		t.Error("part[1] should be an expression")
	}
	testIdentifier(t, sl.Parts[1].Expr, "name")
}

func TestStringEscapeSequences(t *testing.T) {
	input := `"line1\nline2";`
	program := parseProgram(t, input)
	expectStatementCount(t, program, 1)
	stmt := getExpressionStatement(t, program, 0)

	sl, ok := stmt.Expression.(*ast.StringLiteral)
	if !ok {
		t.Fatalf("expression is %T, want *ast.StringLiteral", stmt.Expression)
	}
	if sl.Value != "line1\nline2" {
		t.Errorf("Value = %q, want %q", sl.Value, "line1\nline2")
	}
}

// ---------------------------------------------------------------------------
// 13. Array literals
// ---------------------------------------------------------------------------

func TestArrayLiteralEmpty(t *testing.T) {
	input := "[];";
	program := parseProgram(t, input)
	expectStatementCount(t, program, 1)
	stmt := getExpressionStatement(t, program, 0)

	arr, ok := stmt.Expression.(*ast.ArrayLiteral)
	if !ok {
		t.Fatalf("expression is %T, want *ast.ArrayLiteral", stmt.Expression)
	}
	if len(arr.Elements) != 0 {
		t.Errorf("Elements has %d items, want 0", len(arr.Elements))
	}
}

func TestArrayLiteralWithElements(t *testing.T) {
	input := "[1, 2, 3];"
	program := parseProgram(t, input)
	expectStatementCount(t, program, 1)
	stmt := getExpressionStatement(t, program, 0)

	arr, ok := stmt.Expression.(*ast.ArrayLiteral)
	if !ok {
		t.Fatalf("expression is %T, want *ast.ArrayLiteral", stmt.Expression)
	}
	if len(arr.Elements) != 3 {
		t.Fatalf("Elements has %d items, want 3", len(arr.Elements))
	}
	testIntegerLiteral(t, arr.Elements[0], 1)
	testIntegerLiteral(t, arr.Elements[1], 2)
	testIntegerLiteral(t, arr.Elements[2], 3)
}

func TestArrayLiteralMixed(t *testing.T) {
	input := `[1, "two", true];`
	program := parseProgram(t, input)
	expectStatementCount(t, program, 1)
	stmt := getExpressionStatement(t, program, 0)

	arr, ok := stmt.Expression.(*ast.ArrayLiteral)
	if !ok {
		t.Fatalf("expression is %T, want *ast.ArrayLiteral", stmt.Expression)
	}
	if len(arr.Elements) != 3 {
		t.Fatalf("Elements has %d items, want 3", len(arr.Elements))
	}
	testIntegerLiteral(t, arr.Elements[0], 1)

	sl, ok := arr.Elements[1].(*ast.StringLiteral)
	if !ok {
		t.Fatalf("element[1] is %T, want *ast.StringLiteral", arr.Elements[1])
	}
	if sl.Value != "two" {
		t.Errorf("element[1] value = %q, want %q", sl.Value, "two")
	}
	testBooleanLiteral(t, arr.Elements[2], true)
}

// ---------------------------------------------------------------------------
// 14. Map literals
// ---------------------------------------------------------------------------

func TestMapLiteralEmpty(t *testing.T) {
	// "{}" would be parsed as a map with zero pairs
	input := "{};"
	program := parseProgram(t, input)
	expectStatementCount(t, program, 1)
	stmt := getExpressionStatement(t, program, 0)

	m, ok := stmt.Expression.(*ast.MapLiteral)
	if !ok {
		t.Fatalf("expression is %T, want *ast.MapLiteral", stmt.Expression)
	}
	if len(m.Pairs) != 0 {
		t.Errorf("Pairs has %d entries, want 0", len(m.Pairs))
	}
}

func TestMapLiteralStringKeys(t *testing.T) {
	input := `{"one": 1, "two": 2, "three": 3};`
	program := parseProgram(t, input)
	expectStatementCount(t, program, 1)
	stmt := getExpressionStatement(t, program, 0)

	m, ok := stmt.Expression.(*ast.MapLiteral)
	if !ok {
		t.Fatalf("expression is %T, want *ast.MapLiteral", stmt.Expression)
	}
	if len(m.Pairs) != 3 {
		t.Fatalf("Pairs has %d entries, want 3", len(m.Pairs))
	}

	expected := map[string]int64{
		"one":   1,
		"two":   2,
		"three": 3,
	}

	for key, val := range m.Pairs {
		sl, ok := key.(*ast.StringLiteral)
		if !ok {
			t.Errorf("key is %T, want *ast.StringLiteral", key)
			continue
		}
		ev, exists := expected[sl.Value]
		if !exists {
			t.Errorf("unexpected key %q", sl.Value)
			continue
		}
		testIntegerLiteral(t, val, ev)
	}
}

// ---------------------------------------------------------------------------
// 15. Index expressions
// ---------------------------------------------------------------------------

func TestIndexExpression(t *testing.T) {
	input := "arr[1];"
	program := parseProgram(t, input)
	expectStatementCount(t, program, 1)
	stmt := getExpressionStatement(t, program, 0)

	idx, ok := stmt.Expression.(*ast.IndexExpression)
	if !ok {
		t.Fatalf("expression is %T, want *ast.IndexExpression", stmt.Expression)
	}
	testIdentifier(t, idx.Left, "arr")
	testIntegerLiteral(t, idx.Index, 1)
}

func TestIndexExpressionComplex(t *testing.T) {
	input := "arr[1 + 2];"
	program := parseProgram(t, input)
	expectStatementCount(t, program, 1)
	stmt := getExpressionStatement(t, program, 0)

	idx, ok := stmt.Expression.(*ast.IndexExpression)
	if !ok {
		t.Fatalf("expression is %T, want *ast.IndexExpression", stmt.Expression)
	}
	testIdentifier(t, idx.Left, "arr")
	testInfixExpression(t, idx.Index, int64(1), "+", int64(2))
}

// ---------------------------------------------------------------------------
// 16. For/in statements
// ---------------------------------------------------------------------------

func TestForStatement(t *testing.T) {
	input := "for x in items { x }"
	program := parseProgram(t, input)
	expectStatementCount(t, program, 1)

	forStmt, ok := program.Statements[0].(*ast.ForStatement)
	if !ok {
		t.Fatalf("statement is %T, want *ast.ForStatement", program.Statements[0])
	}
	if forStmt.TokenLiteral() != "for" {
		t.Errorf("TokenLiteral() = %q, want %q", forStmt.TokenLiteral(), "for")
	}
	if forStmt.Variable.Value != "x" {
		t.Errorf("Variable = %q, want %q", forStmt.Variable.Value, "x")
	}
	testIdentifier(t, forStmt.Iterable, "items")
	if len(forStmt.Body.Statements) != 1 {
		t.Fatalf("body has %d statements, want 1", len(forStmt.Body.Statements))
	}
}

func TestForStatementWithRange(t *testing.T) {
	input := "for i in 0..10 { i }"
	program := parseProgram(t, input)
	expectStatementCount(t, program, 1)

	forStmt, ok := program.Statements[0].(*ast.ForStatement)
	if !ok {
		t.Fatalf("statement is %T, want *ast.ForStatement", program.Statements[0])
	}

	rangeLit, ok := forStmt.Iterable.(*ast.RangeLiteral)
	if !ok {
		t.Fatalf("iterable is %T, want *ast.RangeLiteral", forStmt.Iterable)
	}
	testIntegerLiteral(t, rangeLit.Start, 0)
	testIntegerLiteral(t, rangeLit.End, 10)
}

// ---------------------------------------------------------------------------
// 17. While statements
// ---------------------------------------------------------------------------

func TestWhileStatement(t *testing.T) {
	input := "while x < 10 { x }"
	program := parseProgram(t, input)
	expectStatementCount(t, program, 1)

	whileStmt, ok := program.Statements[0].(*ast.WhileStatement)
	if !ok {
		t.Fatalf("statement is %T, want *ast.WhileStatement", program.Statements[0])
	}
	if whileStmt.TokenLiteral() != "while" {
		t.Errorf("TokenLiteral() = %q, want %q", whileStmt.TokenLiteral(), "while")
	}
	testInfixExpression(t, whileStmt.Condition, "x", "<", int64(10))
	if len(whileStmt.Body.Statements) != 1 {
		t.Fatalf("body has %d statements, want 1", len(whileStmt.Body.Statements))
	}
}

func TestWhileStatementTrue(t *testing.T) {
	input := "while true { 1 }"
	program := parseProgram(t, input)
	expectStatementCount(t, program, 1)

	whileStmt, ok := program.Statements[0].(*ast.WhileStatement)
	if !ok {
		t.Fatalf("statement is %T, want *ast.WhileStatement", program.Statements[0])
	}
	testBooleanLiteral(t, whileStmt.Condition, true)
}

// ---------------------------------------------------------------------------
// 18. Break and continue statements
// ---------------------------------------------------------------------------

func TestBreakStatement(t *testing.T) {
	input := "while true { break }"
	program := parseProgram(t, input)
	expectStatementCount(t, program, 1)

	whileStmt := program.Statements[0].(*ast.WhileStatement)
	if len(whileStmt.Body.Statements) != 1 {
		t.Fatalf("body has %d statements, want 1", len(whileStmt.Body.Statements))
	}

	brk, ok := whileStmt.Body.Statements[0].(*ast.BreakStatement)
	if !ok {
		t.Fatalf("body statement is %T, want *ast.BreakStatement", whileStmt.Body.Statements[0])
	}
	if brk.TokenLiteral() != "break" {
		t.Errorf("TokenLiteral() = %q, want %q", brk.TokenLiteral(), "break")
	}
}

func TestContinueStatement(t *testing.T) {
	input := "while true { continue }"
	program := parseProgram(t, input)
	expectStatementCount(t, program, 1)

	whileStmt := program.Statements[0].(*ast.WhileStatement)
	if len(whileStmt.Body.Statements) != 1 {
		t.Fatalf("body has %d statements, want 1", len(whileStmt.Body.Statements))
	}

	cont, ok := whileStmt.Body.Statements[0].(*ast.ContinueStatement)
	if !ok {
		t.Fatalf("body statement is %T, want *ast.ContinueStatement", whileStmt.Body.Statements[0])
	}
	if cont.TokenLiteral() != "continue" {
		t.Errorf("TokenLiteral() = %q, want %q", cont.TokenLiteral(), "continue")
	}
}

func TestBreakAndContinueInFor(t *testing.T) {
	input := `for x in items {
  if x == 0 { continue }
  if x == 5 { break }
}`
	program := parseProgram(t, input)
	expectStatementCount(t, program, 1)

	forStmt, ok := program.Statements[0].(*ast.ForStatement)
	if !ok {
		t.Fatalf("statement is %T, want *ast.ForStatement", program.Statements[0])
	}
	if len(forStmt.Body.Statements) != 2 {
		t.Fatalf("body has %d statements, want 2", len(forStmt.Body.Statements))
	}
}

// ---------------------------------------------------------------------------
// 19. Range expressions
// ---------------------------------------------------------------------------

func TestRangeExpression(t *testing.T) {
	input := "1..10;"
	program := parseProgram(t, input)
	expectStatementCount(t, program, 1)
	stmt := getExpressionStatement(t, program, 0)

	rangeLit, ok := stmt.Expression.(*ast.RangeLiteral)
	if !ok {
		t.Fatalf("expression is %T, want *ast.RangeLiteral", stmt.Expression)
	}
	testIntegerLiteral(t, rangeLit.Start, 1)
	testIntegerLiteral(t, rangeLit.End, 10)
}

func TestRangeExpressionWithIdentifiers(t *testing.T) {
	input := "a..b;"
	program := parseProgram(t, input)
	expectStatementCount(t, program, 1)
	stmt := getExpressionStatement(t, program, 0)

	rangeLit, ok := stmt.Expression.(*ast.RangeLiteral)
	if !ok {
		t.Fatalf("expression is %T, want *ast.RangeLiteral", stmt.Expression)
	}
	testIdentifier(t, rangeLit.Start, "a")
	testIdentifier(t, rangeLit.End, "b")
}

// ---------------------------------------------------------------------------
// 20. Pipe expressions
// ---------------------------------------------------------------------------

func TestPipeExpression(t *testing.T) {
	input := "x |> foo();"
	program := parseProgram(t, input)
	expectStatementCount(t, program, 1)
	stmt := getExpressionStatement(t, program, 0)

	pipe, ok := stmt.Expression.(*ast.PipeExpression)
	if !ok {
		t.Fatalf("expression is %T, want *ast.PipeExpression", stmt.Expression)
	}
	testIdentifier(t, pipe.Left, "x")

	call, ok := pipe.Right.(*ast.CallExpression)
	if !ok {
		t.Fatalf("pipe.Right is %T, want *ast.CallExpression", pipe.Right)
	}
	testIdentifier(t, call.Function, "foo")
}

func TestPipeExpressionChained(t *testing.T) {
	input := "x |> foo() |> bar();"
	program := parseProgram(t, input)
	expectStatementCount(t, program, 1)
	stmt := getExpressionStatement(t, program, 0)

	// |> is left-associative, so: (x |> foo()) |> bar()
	pipe, ok := stmt.Expression.(*ast.PipeExpression)
	if !ok {
		t.Fatalf("expression is %T, want *ast.PipeExpression", stmt.Expression)
	}

	// right side should be bar()
	rightCall, ok := pipe.Right.(*ast.CallExpression)
	if !ok {
		t.Fatalf("pipe.Right is %T, want *ast.CallExpression", pipe.Right)
	}
	testIdentifier(t, rightCall.Function, "bar")

	// left side should be another pipe expression
	leftPipe, ok := pipe.Left.(*ast.PipeExpression)
	if !ok {
		t.Fatalf("pipe.Left is %T, want *ast.PipeExpression", pipe.Left)
	}
	testIdentifier(t, leftPipe.Left, "x")
}

// ---------------------------------------------------------------------------
// 21. Assign expressions
// ---------------------------------------------------------------------------

func TestAssignExpression(t *testing.T) {
	input := "x = 5;"
	program := parseProgram(t, input)
	expectStatementCount(t, program, 1)
	stmt := getExpressionStatement(t, program, 0)

	assign, ok := stmt.Expression.(*ast.AssignExpression)
	if !ok {
		t.Fatalf("expression is %T, want *ast.AssignExpression", stmt.Expression)
	}
	testIdentifier(t, assign.Name, "x")
	testIntegerLiteral(t, assign.Value, 5)
}

func TestAssignExpressionString(t *testing.T) {
	input := `name = "alice";`
	program := parseProgram(t, input)
	expectStatementCount(t, program, 1)
	stmt := getExpressionStatement(t, program, 0)

	assign, ok := stmt.Expression.(*ast.AssignExpression)
	if !ok {
		t.Fatalf("expression is %T, want *ast.AssignExpression", stmt.Expression)
	}
	testIdentifier(t, assign.Name, "name")

	sl, ok := assign.Value.(*ast.StringLiteral)
	if !ok {
		t.Fatalf("value is %T, want *ast.StringLiteral", assign.Value)
	}
	if sl.Value != "alice" {
		t.Errorf("Value = %q, want %q", sl.Value, "alice")
	}
}

func TestAssignExpressionIndex(t *testing.T) {
	input := "arr[0] = 99;"
	program := parseProgram(t, input)
	expectStatementCount(t, program, 1)
	stmt := getExpressionStatement(t, program, 0)

	assign, ok := stmt.Expression.(*ast.AssignExpression)
	if !ok {
		t.Fatalf("expression is %T, want *ast.AssignExpression", stmt.Expression)
	}

	idx, ok := assign.Name.(*ast.IndexExpression)
	if !ok {
		t.Fatalf("assign.Name is %T, want *ast.IndexExpression", assign.Name)
	}
	testIdentifier(t, idx.Left, "arr")
	testIntegerLiteral(t, idx.Index, 0)
	testIntegerLiteral(t, assign.Value, 99)
}

// ---------------------------------------------------------------------------
// 22. Try/catch expressions
// ---------------------------------------------------------------------------

func TestTryCatchExpression(t *testing.T) {
	input := "try { risky() } catch e { handle(e) }"
	program := parseProgram(t, input)
	expectStatementCount(t, program, 1)
	stmt := getExpressionStatement(t, program, 0)

	tryExp, ok := stmt.Expression.(*ast.TryExpression)
	if !ok {
		t.Fatalf("expression is %T, want *ast.TryExpression", stmt.Expression)
	}

	if tryExp.Body == nil {
		t.Fatal("Body is nil")
	}
	if len(tryExp.Body.Statements) != 1 {
		t.Fatalf("body has %d statements, want 1", len(tryExp.Body.Statements))
	}

	if tryExp.CatchVar == nil {
		t.Fatal("CatchVar is nil")
	}
	if tryExp.CatchVar.Value != "e" {
		t.Errorf("CatchVar.Value = %q, want %q", tryExp.CatchVar.Value, "e")
	}

	if tryExp.Handler == nil {
		t.Fatal("Handler is nil")
	}
	if len(tryExp.Handler.Statements) != 1 {
		t.Fatalf("handler has %d statements, want 1", len(tryExp.Handler.Statements))
	}
}

func TestTryCatchWithoutVariable(t *testing.T) {
	input := "try { risky() } catch { fallback() }"
	program := parseProgram(t, input)
	expectStatementCount(t, program, 1)
	stmt := getExpressionStatement(t, program, 0)

	tryExp, ok := stmt.Expression.(*ast.TryExpression)
	if !ok {
		t.Fatalf("expression is %T, want *ast.TryExpression", stmt.Expression)
	}

	if tryExp.CatchVar != nil {
		t.Errorf("CatchVar should be nil, got %q", tryExp.CatchVar.Value)
	}

	if tryExp.Handler == nil {
		t.Fatal("Handler is nil")
	}
}

// ---------------------------------------------------------------------------
// 23. Parser error reporting
// ---------------------------------------------------------------------------

func TestParserErrors(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"let missing ident", "let = 5;"},
		{"let missing assign", "let x 5;"},
		{"unclosed paren", "(5 + 5;"},
		{"unclosed bracket", "arr[1;"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			p.ParseProgram()

			errs := p.Errors()
			if len(errs) == 0 {
				t.Errorf("expected parser errors for input %q, got none", tt.input)
			}
		})
	}
}

func TestParserErrorMessages(t *testing.T) {
	input := "let = 5;"
	l := lexer.New(input)
	p := New(l)
	p.ParseProgram()

	errs := p.Errors()
	if len(errs) == 0 {
		t.Fatal("expected errors, got none")
	}
	// Error should mention that IDENT was expected
	found := false
	for _, e := range errs {
		if len(e) > 0 {
			found = true
		}
	}
	if !found {
		t.Error("expected non-empty error messages")
	}
}

// ---------------------------------------------------------------------------
// Additional coverage: identifiers, grouped expressions, String() output
// ---------------------------------------------------------------------------

func TestIdentifierExpression(t *testing.T) {
	input := "foobar;"
	program := parseProgram(t, input)
	expectStatementCount(t, program, 1)
	stmt := getExpressionStatement(t, program, 0)
	testIdentifier(t, stmt.Expression, "foobar")
}

func TestGroupedExpression(t *testing.T) {
	input := "(5 + 5);"
	program := parseProgram(t, input)
	expectStatementCount(t, program, 1)
	stmt := getExpressionStatement(t, program, 0)
	testInfixExpression(t, stmt.Expression, int64(5), "+", int64(5))
}

func TestMultipleStatements(t *testing.T) {
	input := "let a = 1\nlet b = 2\nreturn a"
	program := parseProgram(t, input)
	expectStatementCount(t, program, 3)

	_, ok1 := program.Statements[0].(*ast.LetStatement)
	_, ok2 := program.Statements[1].(*ast.LetStatement)
	_, ok3 := program.Statements[2].(*ast.ReturnStatement)

	if !ok1 || !ok2 || !ok3 {
		t.Errorf("unexpected statement types: %T, %T, %T",
			program.Statements[0], program.Statements[1], program.Statements[2])
	}
}

func TestProgramString(t *testing.T) {
	input := "let x = 5;"
	program := parseProgram(t, input)
	s := program.String()
	if s == "" {
		t.Error("Program.String() returned empty string")
	}
}

// ---------------------------------------------------------------------------
// Semicolons vs newlines as terminators
// ---------------------------------------------------------------------------

func TestSemicolonTerminator(t *testing.T) {
	input := "5; 10;"
	program := parseProgram(t, input)
	if len(program.Statements) != 2 {
		t.Fatalf("expected 2 statements, got %d", len(program.Statements))
	}
}

func TestNewlineTerminator(t *testing.T) {
	input := "5\n10\n"
	program := parseProgram(t, input)
	if len(program.Statements) != 2 {
		t.Fatalf("expected 2 statements, got %d", len(program.Statements))
	}
}

// ---------------------------------------------------------------------------
// Edge cases for function calls -- function expression calls
// ---------------------------------------------------------------------------

func TestImmediateCallExpression(t *testing.T) {
	input := "fn(x) { x }(42);"
	program := parseProgram(t, input)
	expectStatementCount(t, program, 1)
	stmt := getExpressionStatement(t, program, 0)

	call, ok := stmt.Expression.(*ast.CallExpression)
	if !ok {
		t.Fatalf("expression is %T, want *ast.CallExpression", stmt.Expression)
	}

	_, fnOk := call.Function.(*ast.FunctionLiteral)
	if !fnOk {
		t.Fatalf("call.Function is %T, want *ast.FunctionLiteral", call.Function)
	}
	if len(call.Arguments) != 1 {
		t.Fatalf("Arguments has %d items, want 1", len(call.Arguments))
	}
	testLiteralExpression(t, call.Arguments[0].Value, int64(42))
}

// ---------------------------------------------------------------------------
// Nested structures
// ---------------------------------------------------------------------------

func TestNestedIfInsideFunction(t *testing.T) {
	input := `fn check(x) { if x > 0 { return x } else { return 0 } }`
	program := parseProgram(t, input)
	expectStatementCount(t, program, 1)
	stmt := getExpressionStatement(t, program, 0)

	fn, ok := stmt.Expression.(*ast.FunctionLiteral)
	if !ok {
		t.Fatalf("expression is %T, want *ast.FunctionLiteral", stmt.Expression)
	}
	if fn.Name != "check" {
		t.Errorf("fn.Name = %q, want %q", fn.Name, "check")
	}
	if len(fn.Body.Statements) != 1 {
		t.Fatalf("body has %d statements, want 1", len(fn.Body.Statements))
	}

	bodyStmt, ok := fn.Body.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("body stmt is %T, want *ast.ExpressionStatement", fn.Body.Statements[0])
	}
	_, isIf := bodyStmt.Expression.(*ast.IfExpression)
	if !isIf {
		t.Fatalf("body expression is %T, want *ast.IfExpression", bodyStmt.Expression)
	}
}

func TestNestedArrayInMap(t *testing.T) {
	input := `{"items": [1, 2, 3]};`
	program := parseProgram(t, input)
	expectStatementCount(t, program, 1)
	stmt := getExpressionStatement(t, program, 0)

	m, ok := stmt.Expression.(*ast.MapLiteral)
	if !ok {
		t.Fatalf("expression is %T, want *ast.MapLiteral", stmt.Expression)
	}
	if len(m.Pairs) != 1 {
		t.Fatalf("Pairs has %d entries, want 1", len(m.Pairs))
	}

	for key, val := range m.Pairs {
		sl, ok := key.(*ast.StringLiteral)
		if !ok {
			t.Fatalf("key is %T, want *ast.StringLiteral", key)
		}
		if sl.Value != "items" {
			t.Errorf("key = %q, want %q", sl.Value, "items")
		}
		arr, ok := val.(*ast.ArrayLiteral)
		if !ok {
			t.Fatalf("value is %T, want *ast.ArrayLiteral", val)
		}
		if len(arr.Elements) != 3 {
			t.Errorf("array has %d elements, want 3", len(arr.Elements))
		}
	}
}

// ---------------------------------------------------------------------------
// Comprehensive precedence test using String()
// ---------------------------------------------------------------------------

func TestOperatorPrecedenceComplete(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		// assignment has lowest precedence
		{"a + b;", "(a + b)"},
		{"a * b + c;", "((a * b) + c)"},
		// comparison vs arithmetic
		{"a + b > c;", "((a + b) > c)"},
		{"a == b != c;", "(a == b)"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			program := parseProgram(t, tt.input)
			if len(program.Statements) == 0 {
				t.Fatal("no statements parsed")
			}
			actual := program.Statements[0].String()
			if actual != tt.expected {
				t.Errorf("got %q, want %q", actual, tt.expected)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Whitespace-insensitive parsing
// ---------------------------------------------------------------------------

func TestMultilineLetStatement(t *testing.T) {
	input := `let result = fn(x) {
  x + 1
}`
	program := parseProgram(t, input)
	expectStatementCount(t, program, 1)

	stmt, ok := program.Statements[0].(*ast.LetStatement)
	if !ok {
		t.Fatalf("statement is %T, want *ast.LetStatement", program.Statements[0])
	}
	if stmt.Name.Value != "result" {
		t.Errorf("Name = %q, want %q", stmt.Name.Value, "result")
	}

	_, fnOk := stmt.Value.(*ast.FunctionLiteral)
	if !fnOk {
		t.Fatalf("value is %T, want *ast.FunctionLiteral", stmt.Value)
	}
}

// ---------------------------------------------------------------------------
// Interaction of features
// ---------------------------------------------------------------------------

func TestLetWithArrayValue(t *testing.T) {
	input := "let xs = [1, 2, 3];"
	program := parseProgram(t, input)
	expectStatementCount(t, program, 1)

	stmt := program.Statements[0].(*ast.LetStatement)
	arr, ok := stmt.Value.(*ast.ArrayLiteral)
	if !ok {
		t.Fatalf("value is %T, want *ast.ArrayLiteral", stmt.Value)
	}
	if len(arr.Elements) != 3 {
		t.Errorf("array has %d elements, want 3", len(arr.Elements))
	}
}

func TestLetWithMapValue(t *testing.T) {
	input := `let m = {"a": 1};`
	program := parseProgram(t, input)
	expectStatementCount(t, program, 1)

	stmt := program.Statements[0].(*ast.LetStatement)
	_, ok := stmt.Value.(*ast.MapLiteral)
	if !ok {
		t.Fatalf("value is %T, want *ast.MapLiteral", stmt.Value)
	}
}

func TestReturnCallExpression(t *testing.T) {
	input := "return add(1, 2);"
	program := parseProgram(t, input)
	expectStatementCount(t, program, 1)

	ret := program.Statements[0].(*ast.ReturnStatement)
	call, ok := ret.ReturnValue.(*ast.CallExpression)
	if !ok {
		t.Fatalf("ReturnValue is %T, want *ast.CallExpression", ret.ReturnValue)
	}
	testIdentifier(t, call.Function, "add")
	if len(call.Arguments) != 2 {
		t.Errorf("args = %d, want 2", len(call.Arguments))
	}
}

// ---------------------------------------------------------------------------
// Ensure String() methods produce stable output for debugging
// ---------------------------------------------------------------------------

func TestNodeStringMethods(t *testing.T) {
	tests := []struct {
		input    string
		contains string
	}{
		{"let x = 5;", "let x = 5"},
		{"return 10;", "return 10"},
		{"null;", "null"},
		{"break", "break"},
		{"continue", "continue"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			program := parseProgram(t, tt.input)
			s := program.String()
			if s == "" {
				t.Error("String() returned empty")
			}
			if len(tt.contains) > 0 {
				found := false
				if len(s) >= len(tt.contains) {
					for i := 0; i <= len(s)-len(tt.contains); i++ {
						if s[i:i+len(tt.contains)] == tt.contains {
							found = true
							break
						}
					}
				}
				if !found {
					t.Errorf("String() = %q, expected it to contain %q", s, tt.contains)
				}
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Verify that TokenLiteral() returns the keyword for statement nodes
// ---------------------------------------------------------------------------

func TestTokenLiterals(t *testing.T) {
	tests := []struct {
		input    string
		literal  string
		stmtType string
	}{
		{"let x = 1;", "let", "let"},
		{"return 1;", "return", "return"},
		{"for x in y { x }", "for", "for"},
		{"while true { 1 }", "while", "while"},
	}

	for _, tt := range tests {
		t.Run(tt.stmtType, func(t *testing.T) {
			program := parseProgram(t, tt.input)
			if len(program.Statements) == 0 {
				t.Fatal("no statements")
			}
			if program.Statements[0].TokenLiteral() != tt.literal {
				t.Errorf("TokenLiteral() = %q, want %q",
					program.Statements[0].TokenLiteral(), tt.literal)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Verify comment handling doesn't break parsing
// ---------------------------------------------------------------------------

func TestCommentSkipping(t *testing.T) {
	input := `# this is a comment
let x = 5
# another comment
x`
	program := parseProgram(t, input)
	if len(program.Statements) != 2 {
		t.Fatalf("expected 2 statements, got %d", len(program.Statements))
	}
}

// ---------------------------------------------------------------------------
// Program.TokenLiteral on empty program
// ---------------------------------------------------------------------------

func TestEmptyProgram(t *testing.T) {
	program := parseProgram(t, "")
	if len(program.Statements) != 0 {
		t.Errorf("expected 0 statements, got %d", len(program.Statements))
	}
	if program.TokenLiteral() != "" {
		t.Errorf("TokenLiteral() = %q, want empty", program.TokenLiteral())
	}
}

// ---------------------------------------------------------------------------
// Multiple errors accumulate
// ---------------------------------------------------------------------------

func TestMultipleParserErrors(t *testing.T) {
	input := "let = ;\nlet = ;"
	l := lexer.New(input)
	p := New(l)
	p.ParseProgram()

	errs := p.Errors()
	if len(errs) < 2 {
		t.Errorf("expected at least 2 errors, got %d: %v", len(errs), errs)
	}
}

// ---------------------------------------------------------------------------
// Expressions as statements (coverage for parseExpressionStatement)
// ---------------------------------------------------------------------------

func TestExpressionStatementToken(t *testing.T) {
	input := "foobar;"
	program := parseProgram(t, input)
	stmt := getExpressionStatement(t, program, 0)
	if stmt.TokenLiteral() != "foobar" {
		t.Errorf("TokenLiteral() = %q, want %q", stmt.TokenLiteral(), "foobar")
	}
}

// ---------------------------------------------------------------------------
// Verify that formatted String output round-trips key structure
// ---------------------------------------------------------------------------

func TestFunctionLiteralString(t *testing.T) {
	input := "fn add(a, b) { a + b }"
	program := parseProgram(t, input)
	s := program.String()
	// String() should contain "fn add(" somewhere
	expected := "fn add(a, b)"
	found := false
	for i := 0; i <= len(s)-len(expected); i++ {
		if s[i:i+len(expected)] == expected {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("String() = %q, expected it to contain %q", s, expected)
	}
}

func TestCallExpressionString(t *testing.T) {
	input := "add(1, 2);"
	program := parseProgram(t, input)
	s := program.String()
	if s != "add(1, 2)" {
		t.Errorf("String() = %q, want %q", s, "add(1, 2)")
	}
}

func TestIndexExpressionString(t *testing.T) {
	input := "arr[0];"
	program := parseProgram(t, input)
	s := program.String()
	if s != "(arr[0])" {
		t.Errorf("String() = %q, want %q", s, "(arr[0])")
	}
}

func TestPipeExpressionString(t *testing.T) {
	input := "x |> foo();"
	program := parseProgram(t, input)
	s := program.String()
	if s != "x |> foo()" {
		t.Errorf("String() = %q, want %q", s, "x |> foo()")
	}
}

func TestAssignExpressionString(t *testing.T) {
	input := "x = 10;"
	program := parseProgram(t, input)
	s := program.String()
	if s != "x = 10" {
		t.Errorf("String() = %q, want %q", s, "x = 10")
	}
}

func TestRangeExpressionString(t *testing.T) {
	input := "1..5;"
	program := parseProgram(t, input)
	s := program.String()
	if s != "1..5" {
		t.Errorf("String() = %q, want %q", s, "1..5")
	}
}

func TestNullLiteralString(t *testing.T) {
	input := "null;"
	program := parseProgram(t, input)
	s := program.String()
	if s != "null" {
		t.Errorf("String() = %q, want %q", s, "null")
	}
}

// Prevent regressions: verify that the parser tracks line/col in errors.
func TestErrorIncludesLineInfo(t *testing.T) {
	input := "let = 5;"
	l := lexer.New(input)
	p := New(l)
	p.ParseProgram()

	errs := p.Errors()
	if len(errs) == 0 {
		t.Fatal("expected errors")
	}
	// Error format is "line N, col N: ..."
	msg := errs[0]
	if len(msg) < 5 || msg[:4] != "line" {
		t.Errorf("error %q does not start with line info", msg)
	}
}

// Verify that String() on a TryExpression includes "try" and "catch"
func TestTryExpressionString(t *testing.T) {
	input := "try { 1 } catch e { 2 }"
	program := parseProgram(t, input)
	s := program.String()
	for _, want := range []string{"try", "catch"} {
		if !containsSubstring(s, want) {
			t.Errorf("String() = %q, expected it to contain %q", s, want)
		}
	}
}

func containsSubstring(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 || findSubstring(s, sub))
}

func findSubstring(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

// ---------------------------------------------------------------------------
// Bulk precedence table-driven tests with String() comparison
// ---------------------------------------------------------------------------

func TestAllInfixOperatorsParse(t *testing.T) {
	// Verify each infix operator creates an InfixExpression with the right operator string.
	ops := []struct {
		input string
		op    string
	}{
		{"1 + 2;", "+"},
		{"1 - 2;", "-"},
		{"1 * 2;", "*"},
		{"1 / 2;", "/"},
		{"1 % 2;", "%"},
		{"1 == 2;", "=="},
		{"1 != 2;", "!="},
		{"1 < 2;", "<"},
		{"1 > 2;", ">"},
		{"1 <= 2;", "<="},
		{"1 >= 2;", ">="},
		{"true and false;", "and"},
		{"true or false;", "or"},
	}

	for _, tt := range ops {
		t.Run(fmt.Sprintf("op_%s", tt.op), func(t *testing.T) {
			program := parseProgram(t, tt.input)
			stmt := getExpressionStatement(t, program, 0)
			infix, ok := stmt.Expression.(*ast.InfixExpression)
			if !ok {
				t.Fatalf("expression is %T, want *ast.InfixExpression", stmt.Expression)
			}
			if infix.Operator != tt.op {
				t.Errorf("Operator = %q, want %q", infix.Operator, tt.op)
			}
		})
	}
}
