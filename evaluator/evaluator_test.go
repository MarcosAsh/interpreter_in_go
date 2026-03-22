package evaluator

import (
	"math"
	"pearl/lexer"
	"pearl/object"
	"pearl/parser"
	"strings"
	"testing"
)

// testEval runs the full lex -> parse -> eval pipeline and returns the result.
func testEval(input string) object.Object {
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	env := object.NewEnvironment()
	return Eval(program, env)
}

// --- assertion helpers ---

func testIntegerObject(t *testing.T, obj object.Object, expected int64) bool {
	t.Helper()
	result, ok := obj.(*object.Integer)
	if !ok {
		t.Errorf("object is not Integer. got=%T (%+v)", obj, obj)
		return false
	}
	if result.Value != expected {
		t.Errorf("object has wrong value. got=%d, want=%d", result.Value, expected)
		return false
	}
	return true
}

func testFloatObject(t *testing.T, obj object.Object, expected float64) bool {
	t.Helper()
	result, ok := obj.(*object.Float)
	if !ok {
		t.Errorf("object is not Float. got=%T (%+v)", obj, obj)
		return false
	}
	if math.Abs(result.Value-expected) > 1e-9 {
		t.Errorf("object has wrong value. got=%f, want=%f", result.Value, expected)
		return false
	}
	return true
}

func testStringObject(t *testing.T, obj object.Object, expected string) bool {
	t.Helper()
	result, ok := obj.(*object.String)
	if !ok {
		t.Errorf("object is not String. got=%T (%+v)", obj, obj)
		return false
	}
	if result.Value != expected {
		t.Errorf("object has wrong value. got=%q, want=%q", result.Value, expected)
		return false
	}
	return true
}

func testBooleanObject(t *testing.T, obj object.Object, expected bool) bool {
	t.Helper()
	result, ok := obj.(*object.Boolean)
	if !ok {
		t.Errorf("object is not Boolean. got=%T (%+v)", obj, obj)
		return false
	}
	if result.Value != expected {
		t.Errorf("object has wrong value. got=%t, want=%t", result.Value, expected)
		return false
	}
	return true
}

func testNullObject(t *testing.T, obj object.Object) bool {
	t.Helper()
	if obj == nil {
		t.Errorf("object is nil, expected Null object")
		return false
	}
	if obj.Type() != object.NULL_OBJ {
		t.Errorf("object is not Null. got=%T (%+v)", obj, obj)
		return false
	}
	return true
}

func testErrorObject(t *testing.T, obj object.Object, expectedSubstring string) bool {
	t.Helper()
	errObj, ok := obj.(*object.Error)
	if !ok {
		t.Errorf("object is not Error. got=%T (%+v)", obj, obj)
		return false
	}
	if !strings.Contains(errObj.Message, expectedSubstring) {
		t.Errorf("error message does not contain %q. got=%q", expectedSubstring, errObj.Message)
		return false
	}
	return true
}

// --- tests ---

func TestIntegerEvaluation(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int64
	}{
		{"single digit", "5", 5},
		{"double digit", "10", 10},
		{"zero", "0", 0},
		{"negative literal", "-5", -5},
		{"large number", "999999", 999999},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evaluated := testEval(tt.input)
			testIntegerObject(t, evaluated, tt.expected)
		})
	}
}

func TestFloatEvaluation(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected float64
	}{
		{"pi", "3.14", 3.14},
		{"small decimal", "0.5", 0.5},
		{"negative float", "-2.5", -2.5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evaluated := testEval(tt.input)
			testFloatObject(t, evaluated, tt.expected)
		})
	}
}

func TestStringEvaluation(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple string", `"hello"`, "hello"},
		{"empty string", `""`, ""},
		{"string with spaces", `"hello world"`, "hello world"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evaluated := testEval(tt.input)
			testStringObject(t, evaluated, tt.expected)
		})
	}
}

func TestBooleanEvaluation(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"true", "true", true},
		{"false", "false", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evaluated := testEval(tt.input)
			testBooleanObject(t, evaluated, tt.expected)
		})
	}
}

func TestNullEvaluation(t *testing.T) {
	evaluated := testEval("null")
	testNullObject(t, evaluated)
}

func TestIntegerArithmetic(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int64
	}{
		{"addition", "5 + 5", 10},
		{"subtraction", "5 - 3", 2},
		{"multiplication", "5 * 2", 10},
		{"division", "10 / 2", 5},
		{"modulo", "10 % 3", 1},
		{"compound expression", "2 + 3 * 4", 14},
		{"grouped expression", "(2 + 3) * 4", 20},
		{"negative result", "5 - 10", -5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evaluated := testEval(tt.input)
			testIntegerObject(t, evaluated, tt.expected)
		})
	}
}

func TestFloatArithmetic(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected float64
	}{
		{"float addition", "1.5 + 2.5", 4.0},
		{"float subtraction", "5.0 - 2.5", 2.5},
		{"float multiplication", "2.0 * 3.5", 7.0},
		{"float division", "10.0 / 4.0", 2.5},
		{"mixed int float add", "5 + 1.5", 6.5},
		{"mixed float int add", "1.5 + 5", 6.5},
		{"mixed int float mul", "3 * 2.5", 7.5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evaluated := testEval(tt.input)
			testFloatObject(t, evaluated, tt.expected)
		})
	}
}

func TestDivisionByZero(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"integer division by zero", "1 / 0"},
		{"integer modulo by zero", "1 % 0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evaluated := testEval(tt.input)
			testErrorObject(t, evaluated, "division by zero")
		})
	}
}

func TestStringConcatenation(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple concat", `"hello" ++ " world"`, "hello world"},
		{"empty concat", `"" ++ "test"`, "test"},
		{"multi concat", `"a" ++ "b" ++ "c"`, "abc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evaluated := testEval(tt.input)
			testStringObject(t, evaluated, tt.expected)
		})
	}
}

func TestComparisonOperators(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		// Integer comparisons
		{"int less than true", "1 < 2", true},
		{"int less than false", "2 < 1", false},
		{"int greater than true", "2 > 1", true},
		{"int greater than false", "1 > 2", false},
		{"int equal true", "1 == 1", true},
		{"int equal false", "1 == 2", false},
		{"int not equal true", "1 != 2", true},
		{"int not equal false", "1 != 1", false},
		{"int less or equal true eq", "1 <= 1", true},
		{"int less or equal true lt", "1 <= 2", true},
		{"int less or equal false", "2 <= 1", false},
		{"int greater or equal true eq", "1 >= 1", true},
		{"int greater or equal true gt", "2 >= 1", true},
		{"int greater or equal false", "1 >= 2", false},

		// Float comparisons
		{"float less than", "1.5 < 2.5", true},
		{"float greater than", "2.5 > 1.5", true},
		{"float equal", "1.5 == 1.5", true},
		{"float not equal", "1.5 != 2.5", true},
		{"float less or equal", "1.5 <= 1.5", true},
		{"float greater or equal", "2.5 >= 1.5", true},

		// String comparisons
		{"string equal", `"abc" == "abc"`, true},
		{"string not equal", `"abc" != "xyz"`, true},
		{"string less than", `"abc" < "def"`, true},
		{"string greater than", `"def" > "abc"`, true},

		// Boolean comparisons
		{"bool equal true", "true == true", true},
		{"bool equal false", "true == false", false},
		{"bool not equal", "true != false", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evaluated := testEval(tt.input)
			testBooleanObject(t, evaluated, tt.expected)
		})
	}
}

func TestPrefixOperators(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected interface{}
	}{
		{"negate integer", "-5", int64(-5)},
		{"negate float", "-3.14", float64(-3.14)},
		{"bang true", "!true", false},
		{"bang false", "!false", true},
		{"not true", "not true", false},
		{"not false", "not false", true},
		{"bang integer truthy", "!5", false},
		{"bang integer zero", "!0", true},
		{"bang null", "!null", true},
		{"double bang", "!!true", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evaluated := testEval(tt.input)
			switch expected := tt.expected.(type) {
			case int64:
				testIntegerObject(t, evaluated, expected)
			case float64:
				testFloatObject(t, evaluated, expected)
			case bool:
				testBooleanObject(t, evaluated, expected)
			}
		})
	}
}

func TestLogicalOperators(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected interface{}
	}{
		{"true and true", "true and true", true},
		{"true and false", "true and false", false},
		{"false and true", "false and true", false},
		{"false and false", "false and false", false},
		{"false or true", "false or true", true},
		{"true or false", "true or false", true},
		{"false or false", "false or false", false},

		// Short-circuit: "and" returns the first falsy value
		{"and short circuit returns falsy", "0 and 5", int64(0)},
		// "and" returns the last value when all truthy
		{"and returns last truthy", "1 and 2", int64(2)},
		// "or" returns first truthy value
		{"or short circuit returns truthy", "0 or 5", int64(5)},
		{"or returns first truthy", "3 or 5", int64(3)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evaluated := testEval(tt.input)
			switch expected := tt.expected.(type) {
			case bool:
				testBooleanObject(t, evaluated, expected)
			case int64:
				testIntegerObject(t, evaluated, expected)
			}
		})
	}
}

func TestIfElseExpressions(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected interface{}
	}{
		{"if true", "if true { 10 }", int64(10)},
		{"if false", "if false { 10 }", nil},
		{"if else true", "if true { 10 } else { 20 }", int64(10)},
		{"if else false", "if false { 10 } else { 20 }", int64(20)},
		{"if 1", "if 1 { 10 }", int64(10)},
		{"if 0 else", "if 0 { 10 } else { 20 }", int64(20)},
		{"if comparison", "if 1 < 2 { 10 }", int64(10)},
		{"if null else", "if null { 10 } else { 20 }", int64(20)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evaluated := testEval(tt.input)
			if tt.expected == nil {
				testNullObject(t, evaluated)
			} else {
				testIntegerObject(t, evaluated, tt.expected.(int64))
			}
		})
	}
}

func TestLetStatements(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int64
	}{
		{"simple let", "let x = 5; x", 5},
		{"let with expression", "let x = 5 * 5; x", 25},
		{"two lets", "let a = 5; let b = a; b", 5},
		{"let chain", "let a = 5; let b = a; let c = a + b; c", 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evaluated := testEval(tt.input)
			testIntegerObject(t, evaluated, tt.expected)
		})
	}
}

func TestVariableAssignment(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int64
	}{
		{"reassign variable", "let x = 5; x = 10; x", 10},
		{"reassign with expression", "let x = 5; x = x + 1; x", 6},
		{"multiple reassign", "let x = 1; x = 2; x = 3; x", 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evaluated := testEval(tt.input)
			testIntegerObject(t, evaluated, tt.expected)
		})
	}
}

func TestReturnStatements(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int64
	}{
		{"return value", "return 10", 10},
		{"return stops execution", "return 10; 9", 10},
		{"return expression", "return 2 * 5; 9", 10},
		{"nested return", "if true { return 10 }; 9", 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evaluated := testEval(tt.input)
			testIntegerObject(t, evaluated, tt.expected)
		})
	}
}

func TestFunctions(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int64
	}{
		{"simple function", "let add = fn(x, y) { x + y }; add(2, 3)", 5},
		{"identity function", "let id = fn(x) { x }; id(42)", 42},
		{"function with return", "let double = fn(x) { return x * 2 }; double(5)", 10},
		{"named function", "fn square(x) { x * x }; square(4)", 16},
		{"immediate call", "fn(x) { x + 1 }(5)", 6},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evaluated := testEval(tt.input)
			testIntegerObject(t, evaluated, tt.expected)
		})
	}
}

func TestClosures(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int64
	}{
		{
			"closure over argument",
			"let newAdder = fn(x) { fn(y) { x + y } }; let addTwo = newAdder(2); addTwo(3)",
			5,
		},
		{
			"closure over let",
			`let x = 10;
			 let add = fn(y) { x + y };
			 add(5)`,
			15,
		},
		{
			"nested closures",
			`let a = fn(x) { fn(y) { fn(z) { x + y + z } } };
			 a(1)(2)(3)`,
			6,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evaluated := testEval(tt.input)
			testIntegerObject(t, evaluated, tt.expected)
		})
	}
}

func TestDefaultParameters(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			"use default",
			`let greet = fn(name, greeting = "Hello") { greeting ++ " " ++ name }; greet("World")`,
			"Hello World",
		},
		{
			"override default",
			`let greet = fn(name, greeting = "Hello") { greeting ++ " " ++ name }; greet("World", "Hi")`,
			"Hi World",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evaluated := testEval(tt.input)
			testStringObject(t, evaluated, tt.expected)
		})
	}
}

func TestNamedArguments(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			"named arg overrides default",
			`let greet = fn(name, greeting = "Hello") { greeting ++ " " ++ name }; greet("World", greeting = "Hey")`,
			"Hey World",
		},
		{
			"named arg reorders params",
			`let build = fn(first, second) { first ++ " " ++ second }; build(second = "B", first = "A")`,
			"A B",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evaluated := testEval(tt.input)
			testStringObject(t, evaluated, tt.expected)
		})
	}
}

func TestRecursion(t *testing.T) {
	input := `
	fn factorial(n) {
		if n <= 1 { return 1 }
		return n * factorial(n - 1)
	}
	factorial(5)
	`
	evaluated := testEval(input)
	testIntegerObject(t, evaluated, 120)
}

func TestFunctionArity(t *testing.T) {
	input := `let add = fn(x, y) { x + y }; add(1, 2, 3)`
	evaluated := testEval(input)
	testErrorObject(t, evaluated, "takes 2 argument(s), got 3")
}

func TestArrays(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected interface{}
	}{
		{"array literal", "[1, 2, 3]", []int64{1, 2, 3}},
		{"array index 0", "[1, 2, 3][0]", int64(1)},
		{"array index 1", "[10, 20, 30][1]", int64(20)},
		{"array index 2", "[1, 2, 3][2]", int64(3)},
		{"negative index -1", "[1, 2, 3][-1]", int64(3)},
		{"negative index -2", "[1, 2, 3][-2]", int64(2)},
		{"out of bounds", "[1, 2, 3][5]", nil},
		{"empty array", "[]", []int64{}},
		{"expression index", "let a = [1, 2, 3]; a[1 + 1]", int64(3)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evaluated := testEval(tt.input)
			switch expected := tt.expected.(type) {
			case int64:
				testIntegerObject(t, evaluated, expected)
			case nil:
				testNullObject(t, evaluated)
			case []int64:
				arr, ok := evaluated.(*object.Array)
				if !ok {
					t.Fatalf("object is not Array. got=%T (%+v)", evaluated, evaluated)
				}
				if len(arr.Elements) != len(expected) {
					t.Fatalf("array has wrong num elements. got=%d, want=%d", len(arr.Elements), len(expected))
				}
				for i, exp := range expected {
					testIntegerObject(t, arr.Elements[i], exp)
				}
			}
		})
	}
}

func TestMaps(t *testing.T) {
	t.Run("map literal and indexing", func(t *testing.T) {
		evaluated := testEval(`{"a": 1}["a"]`)
		testIntegerObject(t, evaluated, 1)
	})

	t.Run("map missing key", func(t *testing.T) {
		evaluated := testEval(`{"a": 1}["b"]`)
		testNullObject(t, evaluated)
	})

	t.Run("map integer key", func(t *testing.T) {
		evaluated := testEval(`{1: "one"}[1]`)
		testStringObject(t, evaluated, "one")
	})

	t.Run("map boolean key", func(t *testing.T) {
		evaluated := testEval(`{true: "yes"}[true]`)
		testStringObject(t, evaluated, "yes")
	})
}

func TestStringIndexing(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected interface{}
	}{
		{"first char", `"hello"[0]`, "h"},
		{"last char positive", `"hello"[4]`, "o"},
		{"negative index", `"hello"[-1]`, "o"},
		{"negative index -2", `"hello"[-2]`, "l"},
		{"out of bounds", `"hello"[10]`, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evaluated := testEval(tt.input)
			switch expected := tt.expected.(type) {
			case string:
				testStringObject(t, evaluated, expected)
			case nil:
				testNullObject(t, evaluated)
			}
		})
	}
}

func TestStringInterpolation(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			"variable interpolation",
			`let x = 5; "x is {x}"`,
			"x is 5",
		},
		{
			"expression interpolation",
			`let a = 3; let b = 4; "sum is {a + b}"`,
			"sum is 7",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evaluated := testEval(tt.input)
			testStringObject(t, evaluated, tt.expected)
		})
	}
}

func TestForLoops(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int64
	}{
		{
			"for over range",
			"let sum = 0; for i in 1..4 { sum = sum + i }; sum",
			6, // 1+2+3
		},
		{
			"for over range from zero",
			"let sum = 0; for i in 0..5 { sum = sum + i }; sum",
			10, // 0+1+2+3+4
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evaluated := testEval(tt.input)
			testIntegerObject(t, evaluated, tt.expected)
		})
	}
}

func TestForOverArrays(t *testing.T) {
	input := `
	let arr = [10, 20, 30]
	let sum = 0
	for x in arr { sum = sum + x }
	sum
	`
	evaluated := testEval(input)
	testIntegerObject(t, evaluated, 60)
}

func TestWhileLoops(t *testing.T) {
	input := `
	let x = 0
	while x < 5 {
		x = x + 1
	}
	x
	`
	evaluated := testEval(input)
	testIntegerObject(t, evaluated, 5)
}

func TestBreakInForLoop(t *testing.T) {
	input := `
	let s = 0
	for i in 0..10 {
		if i == 3 { break }
		s = s + i
	}
	s
	`
	evaluated := testEval(input)
	testIntegerObject(t, evaluated, 3) // 0+1+2
}

func TestContinueInForLoop(t *testing.T) {
	input := `
	let s = 0
	for i in 0..6 {
		if i % 2 == 0 { continue }
		s = s + i
	}
	s
	`
	evaluated := testEval(input)
	testIntegerObject(t, evaluated, 9) // 1+3+5
}

func TestBreakInWhileLoop(t *testing.T) {
	input := `
	let x = 0
	while true {
		if x == 5 { break }
		x = x + 1
	}
	x
	`
	evaluated := testEval(input)
	testIntegerObject(t, evaluated, 5)
}

func TestTryCatch(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected interface{}
	}{
		{
			"catch division by zero",
			`try { 1 / 0 } catch e { e }`,
			"division by zero",
		},
		{
			"no error passes through",
			`try { 5 } catch e { 0 }`,
			int64(5),
		},
		{
			"catch undefined variable",
			`try { x } catch e { "caught" }`,
			"caught",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evaluated := testEval(tt.input)
			switch expected := tt.expected.(type) {
			case string:
				testStringObject(t, evaluated, expected)
			case int64:
				testIntegerObject(t, evaluated, expected)
			}
		})
	}
}

func TestPipeOperator(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected interface{}
	}{
		{
			"pipe to str",
			`5 |> str()`,
			"5",
		},
		{
			"pipe to len",
			`"hello" |> len()`,
			int64(5),
		},
		{
			"chained pipes",
			`"  hello  " |> trim() |> upper()`,
			"HELLO",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evaluated := testEval(tt.input)
			switch expected := tt.expected.(type) {
			case string:
				testStringObject(t, evaluated, expected)
			case int64:
				testIntegerObject(t, evaluated, expected)
			}
		})
	}
}

func TestRegexMatching(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"match true", `"hello" ~ /hel/`, true},
		{"match false", `"hello" ~ /xyz/`, false},
		{"not match true", `"hello" !~ /xyz/`, true},
		{"not match false", `"hello" !~ /hel/`, false},
		{"regex with digits", `"abc123" ~ /[0-9]+/`, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evaluated := testEval(tt.input)
			testBooleanObject(t, evaluated, tt.expected)
		})
	}
}

func TestErrorPropagation(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		expectedMessage string
	}{
		{"undefined variable", "foobar", "undefined variable: foobar"},
		{"type mismatch infix", `5 + "hello"`, "unknown operator"},
		{"unknown prefix", `-true`, "unknown operator"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evaluated := testEval(tt.input)
			testErrorObject(t, evaluated, tt.expectedMessage)
		})
	}
}

func TestUndefinedVariableWithLocation(t *testing.T) {
	evaluated := testEval("foobar")
	errObj, ok := evaluated.(*object.Error)
	if !ok {
		t.Fatalf("expected Error, got %T", evaluated)
	}
	if errObj.Line == 0 && errObj.Col == 0 {
		t.Errorf("expected error to have line/col info, got Line=%d Col=%d", errObj.Line, errObj.Col)
	}
	if !strings.Contains(errObj.Message, "undefined variable") {
		t.Errorf("expected 'undefined variable' in message, got %q", errObj.Message)
	}
}

func TestBuiltinLen(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected interface{}
	}{
		{"string len", `len("hello")`, int64(5)},
		{"empty string len", `len("")`, int64(0)},
		{"array len", `len([1, 2, 3])`, int64(3)},
		{"empty array len", `len([])`, int64(0)},
		{"wrong arg count", `len("a", "b")`, "takes 1 argument"},
		{"wrong type", `len(1)`, "not supported"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evaluated := testEval(tt.input)
			switch expected := tt.expected.(type) {
			case int64:
				testIntegerObject(t, evaluated, expected)
			case string:
				testErrorObject(t, evaluated, expected)
			}
		})
	}
}

func TestBuiltinUpperLower(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"upper", `upper("hello")`, "HELLO"},
		{"lower", `lower("HELLO")`, "hello"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evaluated := testEval(tt.input)
			testStringObject(t, evaluated, tt.expected)
		})
	}
}

func TestBuiltinTrim(t *testing.T) {
	evaluated := testEval(`trim("  hello  ")`)
	testStringObject(t, evaluated, "hello")
}

func TestBuiltinSplit(t *testing.T) {
	t.Run("default separator", func(t *testing.T) {
		evaluated := testEval(`split("a b c")`)
		arr, ok := evaluated.(*object.Array)
		if !ok {
			t.Fatalf("expected Array, got %T", evaluated)
		}
		if len(arr.Elements) != 3 {
			t.Fatalf("expected 3 elements, got %d", len(arr.Elements))
		}
		testStringObject(t, arr.Elements[0], "a")
		testStringObject(t, arr.Elements[1], "b")
		testStringObject(t, arr.Elements[2], "c")
	})

	t.Run("custom separator", func(t *testing.T) {
		evaluated := testEval(`split("a,b,c", ",")`)
		arr, ok := evaluated.(*object.Array)
		if !ok {
			t.Fatalf("expected Array, got %T", evaluated)
		}
		if len(arr.Elements) != 3 {
			t.Fatalf("expected 3 elements, got %d", len(arr.Elements))
		}
		testStringObject(t, arr.Elements[0], "a")
		testStringObject(t, arr.Elements[2], "c")
	})
}

func TestBuiltinJoin(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"join with separator", `join(["a", "b", "c"], ", ")`, "a, b, c"},
		{"join no separator", `join(["x", "y", "z"])`, "xyz"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evaluated := testEval(tt.input)
			testStringObject(t, evaluated, tt.expected)
		})
	}
}

func TestBuiltinPushPop(t *testing.T) {
	t.Run("push", func(t *testing.T) {
		evaluated := testEval(`let a = [1, 2]; push(a, 3); a[2]`)
		testIntegerObject(t, evaluated, 3)
	})

	t.Run("pop", func(t *testing.T) {
		evaluated := testEval(`let a = [1, 2, 3]; pop(a)`)
		testIntegerObject(t, evaluated, 3)
	})

	t.Run("pop modifies array", func(t *testing.T) {
		evaluated := testEval(`let a = [1, 2, 3]; pop(a); len(a)`)
		testIntegerObject(t, evaluated, 2)
	})
}

func TestBuiltinMapFilterReduce(t *testing.T) {
	t.Run("map", func(t *testing.T) {
		input := `map([1, 2, 3], fn(x) { x * 2 })`
		evaluated := testEval(input)
		arr, ok := evaluated.(*object.Array)
		if !ok {
			t.Fatalf("expected Array, got %T", evaluated)
		}
		if len(arr.Elements) != 3 {
			t.Fatalf("expected 3 elements, got %d", len(arr.Elements))
		}
		testIntegerObject(t, arr.Elements[0], 2)
		testIntegerObject(t, arr.Elements[1], 4)
		testIntegerObject(t, arr.Elements[2], 6)
	})

	t.Run("filter", func(t *testing.T) {
		input := `filter([1, 2, 3, 4, 5], fn(x) { x > 3 })`
		evaluated := testEval(input)
		arr, ok := evaluated.(*object.Array)
		if !ok {
			t.Fatalf("expected Array, got %T", evaluated)
		}
		if len(arr.Elements) != 2 {
			t.Fatalf("expected 2 elements, got %d", len(arr.Elements))
		}
		testIntegerObject(t, arr.Elements[0], 4)
		testIntegerObject(t, arr.Elements[1], 5)
	})

	t.Run("reduce", func(t *testing.T) {
		input := `reduce([1, 2, 3, 4], fn(acc, x) { acc + x }, 0)`
		evaluated := testEval(input)
		testIntegerObject(t, evaluated, 10)
	})
}

func TestBuiltinType(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"integer type", `type(5)`, "INTEGER"},
		{"float type", `type(3.14)`, "FLOAT"},
		{"string type", `type("hi")`, "STRING"},
		{"bool type", `type(true)`, "BOOLEAN"},
		{"null type", `type(null)`, "NULL"},
		{"array type", `type([1])`, "ARRAY"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evaluated := testEval(tt.input)
			testStringObject(t, evaluated, tt.expected)
		})
	}
}

func TestBuiltinStr(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"int to str", `str(42)`, "42"},
		{"float to str", `str(3.14)`, "3.14"},
		{"bool to str", `str(true)`, "true"},
		{"null to str", `str(null)`, "null"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evaluated := testEval(tt.input)
			testStringObject(t, evaluated, tt.expected)
		})
	}
}

func TestBuiltinInt(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int64
	}{
		{"float to int", `int(3.9)`, 3},
		{"string to int", `int("42")`, 42},
		{"bool true to int", `int(true)`, 1},
		{"bool false to int", `int(false)`, 0},
		{"int identity", `int(7)`, 7},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evaluated := testEval(tt.input)
			testIntegerObject(t, evaluated, tt.expected)
		})
	}
}

func TestBuiltinFloat(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected float64
	}{
		{"int to float", `float(5)`, 5.0},
		{"string to float", `float("3.14")`, 3.14},
		{"float identity", `float(2.5)`, 2.5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evaluated := testEval(tt.input)
			testFloatObject(t, evaluated, tt.expected)
		})
	}
}

func TestBuiltinReverse(t *testing.T) {
	t.Run("reverse string", func(t *testing.T) {
		evaluated := testEval(`reverse("hello")`)
		testStringObject(t, evaluated, "olleh")
	})

	t.Run("reverse array", func(t *testing.T) {
		evaluated := testEval(`reverse([1, 2, 3])`)
		arr, ok := evaluated.(*object.Array)
		if !ok {
			t.Fatalf("expected Array, got %T", evaluated)
		}
		testIntegerObject(t, arr.Elements[0], 3)
		testIntegerObject(t, arr.Elements[1], 2)
		testIntegerObject(t, arr.Elements[2], 1)
	})
}

func TestBuiltinContains(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"string contains true", `contains("hello world", "world")`, true},
		{"string contains false", `contains("hello", "xyz")`, false},
		{"array contains true", `contains([1, 2, 3], 2)`, true},
		{"array contains false", `contains([1, 2, 3], 5)`, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evaluated := testEval(tt.input)
			testBooleanObject(t, evaluated, tt.expected)
		})
	}
}

func TestBuiltinKeysValues(t *testing.T) {
	t.Run("keys", func(t *testing.T) {
		evaluated := testEval(`let m = {"a": 1}; keys(m)`)
		arr, ok := evaluated.(*object.Array)
		if !ok {
			t.Fatalf("expected Array, got %T", evaluated)
		}
		if len(arr.Elements) != 1 {
			t.Fatalf("expected 1 key, got %d", len(arr.Elements))
		}
		testStringObject(t, arr.Elements[0], "a")
	})

	t.Run("values", func(t *testing.T) {
		evaluated := testEval(`let m = {"a": 1}; values(m)`)
		arr, ok := evaluated.(*object.Array)
		if !ok {
			t.Fatalf("expected Array, got %T", evaluated)
		}
		if len(arr.Elements) != 1 {
			t.Fatalf("expected 1 value, got %d", len(arr.Elements))
		}
		testIntegerObject(t, arr.Elements[0], 1)
	})
}

func TestBuiltinSort(t *testing.T) {
	evaluated := testEval(`sort([3, 1, 2])`)
	arr, ok := evaluated.(*object.Array)
	if !ok {
		t.Fatalf("expected Array, got %T", evaluated)
	}
	testIntegerObject(t, arr.Elements[0], 1)
	testIntegerObject(t, arr.Elements[1], 2)
	testIntegerObject(t, arr.Elements[2], 3)
}

func TestBuiltinUnique(t *testing.T) {
	evaluated := testEval(`unique([1, 2, 2, 3, 3, 3])`)
	arr, ok := evaluated.(*object.Array)
	if !ok {
		t.Fatalf("expected Array, got %T", evaluated)
	}
	if len(arr.Elements) != 3 {
		t.Fatalf("expected 3 unique elements, got %d", len(arr.Elements))
	}
	testIntegerObject(t, arr.Elements[0], 1)
	testIntegerObject(t, arr.Elements[1], 2)
	testIntegerObject(t, arr.Elements[2], 3)
}

func TestBuiltinRange(t *testing.T) {
	t.Run("range with one arg", func(t *testing.T) {
		evaluated := testEval(`range(3)`)
		r, ok := evaluated.(*object.Range)
		if !ok {
			t.Fatalf("expected Range, got %T", evaluated)
		}
		if r.Start != 0 || r.End != 3 {
			t.Errorf("expected 0..3, got %d..%d", r.Start, r.End)
		}
	})

	t.Run("range with two args", func(t *testing.T) {
		evaluated := testEval(`range(2, 5)`)
		r, ok := evaluated.(*object.Range)
		if !ok {
			t.Fatalf("expected Range, got %T", evaluated)
		}
		if r.Start != 2 || r.End != 5 {
			t.Errorf("expected 2..5, got %d..%d", r.Start, r.End)
		}
	})
}

func TestBuiltinFind(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int64
	}{
		{"find in string", `find("hello world", "world")`, 6},
		{"find not found string", `find("hello", "xyz")`, -1},
		{"find in array", `find([10, 20, 30], 20)`, 1},
		{"find not found array", `find([1, 2, 3], 5)`, -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evaluated := testEval(tt.input)
			testIntegerObject(t, evaluated, tt.expected)
		})
	}
}

func TestBuiltinStartsWithEndsWith(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"starts_with true", `starts_with("hello", "hel")`, true},
		{"starts_with false", `starts_with("hello", "xyz")`, false},
		{"ends_with true", `ends_with("hello", "llo")`, true},
		{"ends_with false", `ends_with("hello", "xyz")`, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evaluated := testEval(tt.input)
			testBooleanObject(t, evaluated, tt.expected)
		})
	}
}

func TestBuiltinCharsLines(t *testing.T) {
	t.Run("chars", func(t *testing.T) {
		evaluated := testEval(`chars("abc")`)
		arr, ok := evaluated.(*object.Array)
		if !ok {
			t.Fatalf("expected Array, got %T", evaluated)
		}
		if len(arr.Elements) != 3 {
			t.Fatalf("expected 3 elements, got %d", len(arr.Elements))
		}
		testStringObject(t, arr.Elements[0], "a")
		testStringObject(t, arr.Elements[1], "b")
		testStringObject(t, arr.Elements[2], "c")
	})

	t.Run("lines", func(t *testing.T) {
		evaluated := testEval(`lines("a\nb\nc")`)
		arr, ok := evaluated.(*object.Array)
		if !ok {
			t.Fatalf("expected Array, got %T", evaluated)
		}
		if len(arr.Elements) != 3 {
			t.Fatalf("expected 3 elements, got %d", len(arr.Elements))
		}
		testStringObject(t, arr.Elements[0], "a")
		testStringObject(t, arr.Elements[1], "b")
		testStringObject(t, arr.Elements[2], "c")
	})
}

func TestBuiltinReplace(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"replace string", `replace("hello world", "world", "go")`, "hello go"},
		{"replace first only", `replace("aaa", "a", "b")`, "baa"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evaluated := testEval(tt.input)
			testStringObject(t, evaluated, tt.expected)
		})
	}
}

func TestBuiltinReplaceAll(t *testing.T) {
	evaluated := testEval(`replace_all("aaa", "a", "b")`)
	testStringObject(t, evaluated, "bbb")
}

func TestBuiltinFlatten(t *testing.T) {
	evaluated := testEval(`flatten([[1, 2], [3, [4, 5]]])`)
	arr, ok := evaluated.(*object.Array)
	if !ok {
		t.Fatalf("expected Array, got %T", evaluated)
	}
	if len(arr.Elements) != 5 {
		t.Fatalf("expected 5 elements, got %d", len(arr.Elements))
	}
	testIntegerObject(t, arr.Elements[0], 1)
	testIntegerObject(t, arr.Elements[4], 5)
}

func TestBuiltinSubstr(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"substr from start", `substr("hello", 0, 3)`, "hel"},
		{"substr from middle", `substr("hello", 2)`, "llo"},
		{"substr negative start", `substr("hello", -3)`, "llo"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evaluated := testEval(tt.input)
			testStringObject(t, evaluated, tt.expected)
		})
	}
}

func TestBuiltinRepeat(t *testing.T) {
	evaluated := testEval(`repeat("ab", 3)`)
	testStringObject(t, evaluated, "ababab")
}

func TestBuiltinSlice(t *testing.T) {
	t.Run("slice with end", func(t *testing.T) {
		evaluated := testEval(`slice([1, 2, 3, 4, 5], 1, 3)`)
		arr, ok := evaluated.(*object.Array)
		if !ok {
			t.Fatalf("expected Array, got %T", evaluated)
		}
		if len(arr.Elements) != 2 {
			t.Fatalf("expected 2 elements, got %d", len(arr.Elements))
		}
		testIntegerObject(t, arr.Elements[0], 2)
		testIntegerObject(t, arr.Elements[1], 3)
	})

	t.Run("slice without end", func(t *testing.T) {
		evaluated := testEval(`slice([1, 2, 3, 4, 5], 2)`)
		arr, ok := evaluated.(*object.Array)
		if !ok {
			t.Fatalf("expected Array, got %T", evaluated)
		}
		if len(arr.Elements) != 3 {
			t.Fatalf("expected 3 elements, got %d", len(arr.Elements))
		}
		testIntegerObject(t, arr.Elements[0], 3)
	})
}

func TestBuiltinShiftUnshift(t *testing.T) {
	t.Run("shift", func(t *testing.T) {
		evaluated := testEval(`let a = [1, 2, 3]; shift(a)`)
		testIntegerObject(t, evaluated, 1)
	})

	t.Run("shift modifies array", func(t *testing.T) {
		evaluated := testEval(`let a = [1, 2, 3]; shift(a); len(a)`)
		testIntegerObject(t, evaluated, 2)
	})

	t.Run("unshift", func(t *testing.T) {
		evaluated := testEval(`let a = [2, 3]; unshift(a, 1); a[0]`)
		testIntegerObject(t, evaluated, 1)
	})
}

func TestBuiltinMatch(t *testing.T) {
	t.Run("match found", func(t *testing.T) {
		evaluated := testEval(`match("hello123", /([a-z]+)([0-9]+)/)`)
		arr, ok := evaluated.(*object.Array)
		if !ok {
			t.Fatalf("expected Array, got %T", evaluated)
		}
		if len(arr.Elements) != 3 {
			t.Fatalf("expected 3 elements (full match + 2 groups), got %d", len(arr.Elements))
		}
		testStringObject(t, arr.Elements[0], "hello123")
		testStringObject(t, arr.Elements[1], "hello")
		testStringObject(t, arr.Elements[2], "123")
	})

	t.Run("match not found", func(t *testing.T) {
		evaluated := testEval(`match("hello", /[0-9]+/)`)
		testNullObject(t, evaluated)
	})
}

func TestBuiltinMatchAll(t *testing.T) {
	evaluated := testEval(`match_all("a1b2c3", /[a-z]([0-9])/)`)
	arr, ok := evaluated.(*object.Array)
	if !ok {
		t.Fatalf("expected Array, got %T", evaluated)
	}
	if len(arr.Elements) != 3 {
		t.Fatalf("expected 3 matches, got %d", len(arr.Elements))
	}
}

func TestBuiltinRegex(t *testing.T) {
	evaluated := testEval(`let r = regex("[0-9]+"); "abc123" ~ r`)
	testBooleanObject(t, evaluated, true)
}

func TestArrayAssignment(t *testing.T) {
	input := `let a = [1, 2, 3]; a[0] = 10; a[0]`
	evaluated := testEval(input)
	testIntegerObject(t, evaluated, 10)
}

func TestMapAssignment(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int64
	}{
		{
			"add new key",
			`let m = {"a": 1}; m["b"] = 2; m["b"]`,
			2,
		},
		{
			"overwrite key",
			`let m = {"a": 1}; m["a"] = 99; m["a"]`,
			99,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evaluated := testEval(tt.input)
			testIntegerObject(t, evaluated, tt.expected)
		})
	}
}

func TestTruthyFalsy(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"0 is falsy", "if 0 { true } else { false }", false},
		{"1 is truthy", "if 1 { true } else { false }", true},
		{"empty string is falsy", `if "" { true } else { false }`, false},
		{"non-empty string is truthy", `if "a" { true } else { false }`, true},
		{"null is falsy", "if null { true } else { false }", false},
		{"empty array is falsy", "if [] { true } else { false }", false},
		{"non-empty array is truthy", "if [1] { true } else { false }", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evaluated := testEval(tt.input)
			testBooleanObject(t, evaluated, tt.expected)
		})
	}
}
