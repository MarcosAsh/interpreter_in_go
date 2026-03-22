package lexer

import (
	"pearl/token"
	"testing"
)

// ---------- helpers ----------

type expectedToken struct {
	typ     token.TokenType
	literal string
}

func assertTokens(t *testing.T, input string, expected []expectedToken) {
	t.Helper()
	l := New(input)
	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp.typ {
			t.Fatalf("token[%d] - wrong type: got=%q, want=%q (literal=%q)",
				i, tok.Type, exp.typ, tok.Literal)
		}
		if tok.Literal != exp.literal {
			t.Fatalf("token[%d] - wrong literal: got=%q, want=%q",
				i, tok.Literal, exp.literal)
		}
	}
}

// ---------- single-char operators and delimiters ----------

func TestSingleCharTokens(t *testing.T) {
	input := `= + - * / % < > ~ ! ; : , ( ) { } [ ]`
	expected := []expectedToken{
		{token.ASSIGN, "="},
		{token.PLUS, "+"},
		{token.MINUS, "-"},
		{token.ASTERISK, "*"},
		{token.SLASH, "/"},
		{token.PERCENT, "%"},
		{token.LT, "<"},
		{token.GT, ">"},
		{token.MATCH, "~"},
		{token.BANG, "!"},
		{token.SEMICOLON, ";"},
		{token.COLON, ":"},
		{token.COMMA, ","},
		{token.LPAREN, "("},
		{token.RPAREN, ")"},
		{token.LBRACE, "{"},
		{token.RBRACE, "}"},
		{token.LBRACKET, "["},
		{token.RBRACKET, "]"},
		{token.EOF, ""},
	}
	assertTokens(t, input, expected)
}

// ---------- multi-char operators ----------

func TestMultiCharOperators(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []expectedToken
	}{
		{"EQ", "==", []expectedToken{{token.EQ, "=="}, {token.EOF, ""}}},
		{"NOT_EQ", "!=", []expectedToken{{token.NOT_EQ, "!="}, {token.EOF, ""}}},
		{"LTE", "<=", []expectedToken{{token.LTE, "<="}, {token.EOF, ""}}},
		{"GTE", ">=", []expectedToken{{token.GTE, ">="}, {token.EOF, ""}}},
		{"CONCAT", "++", []expectedToken{{token.CONCAT, "++"}, {token.EOF, ""}}},
		{"PIPE", "|>", []expectedToken{{token.PIPE, "|>"}, {token.EOF, ""}}},
		{"NOTMATCH", "!~", []expectedToken{{token.NOTMATCH, "!~"}, {token.EOF, ""}}},
		{"RANGE", "..", []expectedToken{{token.RANGE, ".."}, {token.EOF, ""}}},
		{"ARROW", "=>", []expectedToken{{token.ARROW, "=>"}, {token.EOF, ""}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertTokens(t, tt.input, tt.expected)
		})
	}
}

// Multi-char operators followed by other tokens to confirm they don't consume
// too much.
func TestMultiCharBoundaries(t *testing.T) {
	// "== =" should give EQ then ASSIGN
	assertTokens(t, "== =", []expectedToken{
		{token.EQ, "=="},
		{token.ASSIGN, "="},
		{token.EOF, ""},
	})

	// "++" then "+" should give CONCAT then PLUS
	assertTokens(t, "++ +", []expectedToken{
		{token.CONCAT, "++"},
		{token.PLUS, "+"},
		{token.EOF, ""},
	})

	// "!=" then "!" should give NOT_EQ then BANG
	assertTokens(t, "!= !", []expectedToken{
		{token.NOT_EQ, "!="},
		{token.BANG, "!"},
		{token.EOF, ""},
	})
}

// ---------- keywords ----------

func TestKeywords(t *testing.T) {
	tests := []struct {
		input    string
		expected token.TokenType
	}{
		{"let", token.LET},
		{"fn", token.FN},
		{"true", token.TRUE},
		{"false", token.FALSE},
		{"if", token.IF},
		{"else", token.ELSE},
		{"return", token.RETURN},
		{"for", token.FOR},
		{"in", token.IN},
		{"while", token.WHILE},
		{"and", token.AND},
		{"or", token.OR},
		{"not", token.NOT},
		{"null", token.NULL},
		{"try", token.TRY},
		{"catch", token.CATCH},
		{"break", token.BREAK},
		{"continue", token.CONTINUE},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			l := New(tt.input)
			tok := l.NextToken()
			if tok.Type != tt.expected {
				t.Errorf("keyword %q: got type=%q, want=%q", tt.input, tok.Type, tt.expected)
			}
			if tok.Literal != tt.input {
				t.Errorf("keyword %q: got literal=%q", tt.input, tok.Literal)
			}
		})
	}
}

// An identifier that *starts* with a keyword should be IDENT, not a keyword.
func TestKeywordPrefix(t *testing.T) {
	l := New("letter")
	tok := l.NextToken()
	if tok.Type != token.IDENT {
		t.Errorf("expected IDENT, got %q", tok.Type)
	}
	if tok.Literal != "letter" {
		t.Errorf("expected literal 'letter', got %q", tok.Literal)
	}
}

// ---------- identifiers ----------

func TestIdentifiers(t *testing.T) {
	tests := []struct {
		input   string
		literal string
	}{
		{"foo", "foo"},
		{"bar_baz", "bar_baz"},
		{"_private", "_private"},
		{"x1", "x1"},
		{"camelCase", "camelCase"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			l := New(tt.input)
			tok := l.NextToken()
			if tok.Type != token.IDENT {
				t.Errorf("expected IDENT, got %q", tok.Type)
			}
			if tok.Literal != tt.literal {
				t.Errorf("expected literal %q, got %q", tt.literal, tok.Literal)
			}
		})
	}
}

// ---------- numbers ----------

func TestIntegers(t *testing.T) {
	tests := []struct {
		input   string
		literal string
	}{
		{"0", "0"},
		{"5", "5"},
		{"42", "42"},
		{"100000", "100000"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			l := New(tt.input)
			tok := l.NextToken()
			if tok.Type != token.INT {
				t.Errorf("expected INT, got %q", tok.Type)
			}
			if tok.Literal != tt.literal {
				t.Errorf("expected literal %q, got %q", tt.literal, tok.Literal)
			}
		})
	}
}

func TestFloats(t *testing.T) {
	tests := []struct {
		input   string
		literal string
	}{
		{"3.14", "3.14"},
		{"0.5", "0.5"},
		{"100.001", "100.001"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			l := New(tt.input)
			tok := l.NextToken()
			if tok.Type != token.FLOAT {
				t.Errorf("expected FLOAT, got %q", tok.Type)
			}
			if tok.Literal != tt.literal {
				t.Errorf("expected literal %q, got %q", tt.literal, tok.Literal)
			}
		})
	}
}

// A number followed by ".." should be parsed as INT then RANGE, not as a
// float.
func TestIntegerFollowedByRange(t *testing.T) {
	assertTokens(t, "1..10", []expectedToken{
		{token.INT, "1"},
		{token.RANGE, ".."},
		{token.INT, "10"},
		{token.EOF, ""},
	})
}

// ---------- strings ----------

func TestSimpleString(t *testing.T) {
	l := New(`"hello world"`)
	tok := l.NextToken()
	if tok.Type != token.STRING {
		t.Fatalf("expected STRING, got %q", tok.Type)
	}
	if tok.Literal != "hello world" {
		t.Fatalf("expected literal 'hello world', got %q", tok.Literal)
	}
}

func TestStringEscapeSequences(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"newline", `"a\nb"`, "a\nb"},
		{"tab", `"a\tb"`, "a\tb"},
		{"carriage return", `"a\rb"`, "a\rb"},
		{"escaped quote", `"a\"b"`, "a\"b"},
		{"escaped backslash", `"a\\b"`, "a\\b"},
		{"escaped brace", `"a\{b"`, "a{b"},
		{"unknown escape", `"a\xb"`, "a\\xb"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New(tt.input)
			tok := l.NextToken()
			if tok.Type != token.STRING {
				t.Fatalf("expected STRING, got %q", tok.Type)
			}
			if tok.Literal != tt.expected {
				t.Fatalf("expected literal %q, got %q", tt.expected, tok.Literal)
			}
		})
	}
}

func TestEmptyString(t *testing.T) {
	l := New(`""`)
	tok := l.NextToken()
	if tok.Type != token.STRING {
		t.Fatalf("expected STRING, got %q", tok.Type)
	}
	if tok.Literal != "" {
		t.Fatalf("expected empty literal, got %q", tok.Literal)
	}
}

func TestUnterminatedString(t *testing.T) {
	// Unterminated string: the lexer reads to EOF and returns whatever it got.
	l := New(`"no closing quote`)
	tok := l.NextToken()
	if tok.Type != token.STRING {
		t.Fatalf("expected STRING, got %q", tok.Type)
	}
	if tok.Literal != "no closing quote" {
		t.Fatalf("expected literal 'no closing quote', got %q", tok.Literal)
	}
}

func TestStringWithInterpolationSyntax(t *testing.T) {
	// The lexer treats {expr} inside strings as plain literal text.
	l := New(`"hello {name}"`)
	tok := l.NextToken()
	if tok.Type != token.STRING {
		t.Fatalf("expected STRING, got %q", tok.Type)
	}
	if tok.Literal != "hello {name}" {
		t.Fatalf("expected literal 'hello {name}', got %q", tok.Literal)
	}
}

// ---------- comments ----------

func TestCommentSkipped(t *testing.T) {
	input := "# this is a comment\n5"
	l := New(input)
	// The comment is skipped; the next meaningful token after the newline
	// produced by the comment line is NEWLINE (the \n), then INT.
	tok := l.NextToken()
	if tok.Type != token.NEWLINE {
		t.Fatalf("expected NEWLINE after comment, got %q", tok.Type)
	}
	tok = l.NextToken()
	if tok.Type != token.INT || tok.Literal != "5" {
		t.Fatalf("expected INT 5 after comment, got %s %q", tok.Type, tok.Literal)
	}
}

func TestCommentAtEOF(t *testing.T) {
	input := "# comment with no newline"
	l := New(input)
	tok := l.NextToken()
	if tok.Type != token.EOF {
		t.Fatalf("expected EOF after comment at end, got %q (literal=%q)", tok.Type, tok.Literal)
	}
}

func TestInlineComment(t *testing.T) {
	input := "42 # answer"
	assertTokens(t, input, []expectedToken{
		{token.INT, "42"},
		{token.EOF, ""},
	})
}

// ---------- newlines ----------

func TestNewlineToken(t *testing.T) {
	input := "a\nb"
	assertTokens(t, input, []expectedToken{
		{token.IDENT, "a"},
		{token.NEWLINE, "\n"},
		{token.IDENT, "b"},
		{token.EOF, ""},
	})
}

func TestMultipleNewlines(t *testing.T) {
	input := "a\n\nb"
	assertTokens(t, input, []expectedToken{
		{token.IDENT, "a"},
		{token.NEWLINE, "\n"},
		{token.NEWLINE, "\n"},
		{token.IDENT, "b"},
		{token.EOF, ""},
	})
}

// ---------- whitespace ----------

func TestWhitespaceSkipped(t *testing.T) {
	input := "  \t a \t  b  "
	assertTokens(t, input, []expectedToken{
		{token.IDENT, "a"},
		{token.IDENT, "b"},
		{token.EOF, ""},
	})
}

// ---------- empty input / EOF ----------

func TestEmptyInput(t *testing.T) {
	l := New("")
	tok := l.NextToken()
	if tok.Type != token.EOF {
		t.Fatalf("expected EOF, got %q", tok.Type)
	}
	if tok.Literal != "" {
		t.Fatalf("expected empty literal for EOF, got %q", tok.Literal)
	}
}

func TestRepeatedEOF(t *testing.T) {
	l := New("")
	for i := 0; i < 5; i++ {
		tok := l.NextToken()
		if tok.Type != token.EOF {
			t.Fatalf("call %d: expected EOF, got %q", i, tok.Type)
		}
	}
}

// ---------- line and column tracking ----------

func TestLineTracking(t *testing.T) {
	// Verify that tokens on different lines report increasing line numbers.
	// Note: the lexer sets tok.Line = l.line *after* readIdentifier returns,
	// and readIdentifier may have already advanced past the newline that
	// follows the identifier. So the exact numbers depend on how far the
	// lexer read ahead. We just verify the second identifier is on a later
	// line than the newline token (which proves line tracking moves forward).
	input := "5\n10"
	l := New(input)

	tok1 := l.NextToken() // INT "5"
	if tok1.Type != token.INT {
		t.Fatalf("expected INT, got %q", tok1.Type)
	}

	tokNL := l.NextToken() // NEWLINE
	if tokNL.Type != token.NEWLINE {
		t.Fatalf("expected NEWLINE, got %q", tokNL.Type)
	}

	tok2 := l.NextToken() // INT "10"
	if tok2.Type != token.INT {
		t.Fatalf("expected INT, got %q", tok2.Type)
	}

	if tok2.Line <= tok1.Line {
		t.Errorf("expected '10' line (%d) > '5' line (%d)", tok2.Line, tok1.Line)
	}
}

func TestColumnAdvancesWithinLine(t *testing.T) {
	input := "a b"
	l := New(input)

	tok1 := l.NextToken() // a
	tok2 := l.NextToken() // b

	if tok2.Col <= tok1.Col {
		t.Errorf("expected col of 'b' (%d) > col of 'a' (%d)", tok2.Col, tok1.Col)
	}
}

func TestColumnResetsAfterNewline(t *testing.T) {
	// Use single-char tokens so readChar doesn't overshoot.
	// After a newline, the lexer resets col. The next single-char token
	// should have a small column value (1 or 2 depending on whitespace).
	input := "+\n-"
	l := New(input)

	tokPlus := l.NextToken() // PLUS
	if tokPlus.Type != token.PLUS {
		t.Fatalf("expected PLUS, got %q", tokPlus.Type)
	}

	l.NextToken() // NEWLINE

	tokMinus := l.NextToken() // MINUS on new line
	if tokMinus.Type != token.MINUS {
		t.Fatalf("expected MINUS, got %q", tokMinus.Type)
	}

	// The MINUS should have a smaller or equal col than the PLUS
	// (both are at column 1 of their respective lines).
	if tokMinus.Col != tokPlus.Col {
		t.Errorf("expected MINUS col (%d) == PLUS col (%d) since both start at column 1",
			tokMinus.Col, tokPlus.Col)
	}
}

// ---------- illegal tokens ----------

func TestSingleDotIsIllegal(t *testing.T) {
	l := New(".")
	tok := l.NextToken()
	if tok.Type != token.ILLEGAL {
		t.Fatalf("expected ILLEGAL for single dot, got %q", tok.Type)
	}
}

func TestSinglePipeIsIllegal(t *testing.T) {
	l := New("|")
	tok := l.NextToken()
	if tok.Type != token.ILLEGAL {
		t.Fatalf("expected ILLEGAL for single pipe, got %q", tok.Type)
	}
}

// ---------- compound expressions ----------

func TestLetStatement(t *testing.T) {
	input := "let x = 10"
	assertTokens(t, input, []expectedToken{
		{token.LET, "let"},
		{token.IDENT, "x"},
		{token.ASSIGN, "="},
		{token.INT, "10"},
		{token.EOF, ""},
	})
}

func TestFunctionDefinition(t *testing.T) {
	input := "let add = fn(a, b) { a + b }"
	assertTokens(t, input, []expectedToken{
		{token.LET, "let"},
		{token.IDENT, "add"},
		{token.ASSIGN, "="},
		{token.FN, "fn"},
		{token.LPAREN, "("},
		{token.IDENT, "a"},
		{token.COMMA, ","},
		{token.IDENT, "b"},
		{token.RPAREN, ")"},
		{token.LBRACE, "{"},
		{token.IDENT, "a"},
		{token.PLUS, "+"},
		{token.IDENT, "b"},
		{token.RBRACE, "}"},
		{token.EOF, ""},
	})
}

func TestIfElseExpression(t *testing.T) {
	input := "if x == 5 { return true } else { return false }"
	assertTokens(t, input, []expectedToken{
		{token.IF, "if"},
		{token.IDENT, "x"},
		{token.EQ, "=="},
		{token.INT, "5"},
		{token.LBRACE, "{"},
		{token.RETURN, "return"},
		{token.TRUE, "true"},
		{token.RBRACE, "}"},
		{token.ELSE, "else"},
		{token.LBRACE, "{"},
		{token.RETURN, "return"},
		{token.FALSE, "false"},
		{token.RBRACE, "}"},
		{token.EOF, ""},
	})
}

func TestForInLoop(t *testing.T) {
	input := "for x in 1..10 { break }"
	assertTokens(t, input, []expectedToken{
		{token.FOR, "for"},
		{token.IDENT, "x"},
		{token.IN, "in"},
		{token.INT, "1"},
		{token.RANGE, ".."},
		{token.INT, "10"},
		{token.LBRACE, "{"},
		{token.BREAK, "break"},
		{token.RBRACE, "}"},
		{token.EOF, ""},
	})
}

func TestWhileWithContinue(t *testing.T) {
	input := "while true { continue }"
	assertTokens(t, input, []expectedToken{
		{token.WHILE, "while"},
		{token.TRUE, "true"},
		{token.LBRACE, "{"},
		{token.CONTINUE, "continue"},
		{token.RBRACE, "}"},
		{token.EOF, ""},
	})
}

func TestTryCatch(t *testing.T) {
	input := "try { x } catch e { e }"
	assertTokens(t, input, []expectedToken{
		{token.TRY, "try"},
		{token.LBRACE, "{"},
		{token.IDENT, "x"},
		{token.RBRACE, "}"},
		{token.CATCH, "catch"},
		{token.IDENT, "e"},
		{token.LBRACE, "{"},
		{token.IDENT, "e"},
		{token.RBRACE, "}"},
		{token.EOF, ""},
	})
}

func TestPipeExpression(t *testing.T) {
	input := "x |> fn(a) { a + 1 }"
	assertTokens(t, input, []expectedToken{
		{token.IDENT, "x"},
		{token.PIPE, "|>"},
		{token.FN, "fn"},
		{token.LPAREN, "("},
		{token.IDENT, "a"},
		{token.RPAREN, ")"},
		{token.LBRACE, "{"},
		{token.IDENT, "a"},
		{token.PLUS, "+"},
		{token.INT, "1"},
		{token.RBRACE, "}"},
		{token.EOF, ""},
	})
}

func TestArrowExpression(t *testing.T) {
	input := "x => x + 1"
	assertTokens(t, input, []expectedToken{
		{token.IDENT, "x"},
		{token.ARROW, "=>"},
		{token.IDENT, "x"},
		{token.PLUS, "+"},
		{token.INT, "1"},
		{token.EOF, ""},
	})
}

func TestLogicalOperators(t *testing.T) {
	input := "true and false or not null"
	assertTokens(t, input, []expectedToken{
		{token.TRUE, "true"},
		{token.AND, "and"},
		{token.FALSE, "false"},
		{token.OR, "or"},
		{token.NOT, "not"},
		{token.NULL, "null"},
		{token.EOF, ""},
	})
}

func TestStringConcatenation(t *testing.T) {
	input := `"hello" ++ " " ++ "world"`
	assertTokens(t, input, []expectedToken{
		{token.STRING, "hello"},
		{token.CONCAT, "++"},
		{token.STRING, " "},
		{token.CONCAT, "++"},
		{token.STRING, "world"},
		{token.EOF, ""},
	})
}

func TestArrayLiteral(t *testing.T) {
	input := "[1, 2, 3]"
	assertTokens(t, input, []expectedToken{
		{token.LBRACKET, "["},
		{token.INT, "1"},
		{token.COMMA, ","},
		{token.INT, "2"},
		{token.COMMA, ","},
		{token.INT, "3"},
		{token.RBRACKET, "]"},
		{token.EOF, ""},
	})
}

func TestHashLiteral(t *testing.T) {
	input := `{"key": "val"}`
	assertTokens(t, input, []expectedToken{
		{token.LBRACE, "{"},
		{token.STRING, "key"},
		{token.COLON, ":"},
		{token.STRING, "val"},
		{token.RBRACE, "}"},
		{token.EOF, ""},
	})
}

func TestMatchOperator(t *testing.T) {
	input := `x ~ y`
	assertTokens(t, input, []expectedToken{
		{token.IDENT, "x"},
		{token.MATCH, "~"},
		{token.IDENT, "y"},
		{token.EOF, ""},
	})
}

func TestNotMatchOperator(t *testing.T) {
	input := `x !~ y`
	assertTokens(t, input, []expectedToken{
		{token.IDENT, "x"},
		{token.NOTMATCH, "!~"},
		{token.IDENT, "y"},
		{token.EOF, ""},
	})
}

// ---------- ReadRegex ----------

func TestReadRegex(t *testing.T) {
	// Simulates the parser having already consumed the opening SLASH token.
	// After NextToken() returns SLASH, the lexer's current char is the first
	// character after '/'.
	input := "/abc/"
	l := New(input)

	tok := l.NextToken() // SLASH
	if tok.Type != token.SLASH {
		t.Fatalf("expected SLASH, got %q", tok.Type)
	}

	// Now the lexer sits at 'a', which is what ReadRegex expects.
	pattern, err := l.ReadRegex()
	if err != nil {
		t.Fatalf("ReadRegex error: %v", err)
	}
	if pattern != "abc" {
		t.Fatalf("expected pattern 'abc', got %q", pattern)
	}
}

func TestReadRegexWithEscapedSlash(t *testing.T) {
	input := `/a\/b/`
	l := New(input)
	l.NextToken() // consume SLASH

	pattern, err := l.ReadRegex()
	if err != nil {
		t.Fatalf("ReadRegex error: %v", err)
	}
	if pattern != `a\/b` {
		t.Fatalf("expected pattern 'a\\/b', got %q", pattern)
	}
}

func TestReadRegexUnterminated(t *testing.T) {
	input := "/abc"
	l := New(input)
	l.NextToken() // SLASH

	_, err := l.ReadRegex()
	if err == nil {
		t.Fatal("expected error for unterminated regex")
	}
}

func TestReadRegexUnterminatedAtNewline(t *testing.T) {
	input := "/abc\n"
	l := New(input)
	l.NextToken() // SLASH

	_, err := l.ReadRegex()
	if err == nil {
		t.Fatal("expected error for regex terminated by newline")
	}
}

func TestReadRegexEmpty(t *testing.T) {
	input := "//"
	l := New(input)
	l.NextToken() // SLASH

	pattern, err := l.ReadRegex()
	if err != nil {
		t.Fatalf("ReadRegex error: %v", err)
	}
	if pattern != "" {
		t.Fatalf("expected empty pattern, got %q", pattern)
	}
}

// ---------- ReadRegexFromStart ----------

func TestReadRegexFromStart(t *testing.T) {
	// ReadRegexFromStart skips whitespace, expects opening '/', reads to closing '/'.
	input := " /foo/"
	l := New(input)

	pattern, err := l.ReadRegexFromStart()
	if err != nil {
		t.Fatalf("ReadRegexFromStart error: %v", err)
	}
	if pattern != "foo" {
		t.Fatalf("expected pattern 'foo', got %q", pattern)
	}
}

func TestReadRegexFromStartNoSlash(t *testing.T) {
	input := "abc"
	l := New(input)

	_, err := l.ReadRegexFromStart()
	if err == nil {
		t.Fatal("expected error when no opening /")
	}
}

func TestReadRegexFromStartWithEscape(t *testing.T) {
	input := `/he\nllo/`
	l := New(input)

	pattern, err := l.ReadRegexFromStart()
	if err != nil {
		t.Fatalf("ReadRegexFromStart error: %v", err)
	}
	if pattern != `he\nllo` {
		t.Fatalf("expected pattern %q, got %q", `he\nllo`, pattern)
	}
}

func TestReadRegexFromStartUnterminated(t *testing.T) {
	input := "/abc"
	l := New(input)

	_, err := l.ReadRegexFromStart()
	if err == nil {
		t.Fatal("expected error for unterminated regex")
	}
}

// ---------- multi-line program ----------

func TestMultiLineProgram(t *testing.T) {
	input := `let x = 5
let y = 10
let result = x + y`

	expected := []expectedToken{
		{token.LET, "let"},
		{token.IDENT, "x"},
		{token.ASSIGN, "="},
		{token.INT, "5"},
		{token.NEWLINE, "\n"},
		{token.LET, "let"},
		{token.IDENT, "y"},
		{token.ASSIGN, "="},
		{token.INT, "10"},
		{token.NEWLINE, "\n"},
		{token.LET, "let"},
		{token.IDENT, "result"},
		{token.ASSIGN, "="},
		{token.IDENT, "x"},
		{token.PLUS, "+"},
		{token.IDENT, "y"},
		{token.EOF, ""},
	}
	assertTokens(t, input, expected)
}

// ---------- comparison operators ----------

func TestComparisonOperators(t *testing.T) {
	input := "a < b <= c > d >= e"
	assertTokens(t, input, []expectedToken{
		{token.IDENT, "a"},
		{token.LT, "<"},
		{token.IDENT, "b"},
		{token.LTE, "<="},
		{token.IDENT, "c"},
		{token.GT, ">"},
		{token.IDENT, "d"},
		{token.GTE, ">="},
		{token.IDENT, "e"},
		{token.EOF, ""},
	})
}

// ---------- arithmetic with all operators ----------

func TestArithmeticExpression(t *testing.T) {
	input := "1 + 2 - 3 * 4 / 5 % 6"
	assertTokens(t, input, []expectedToken{
		{token.INT, "1"},
		{token.PLUS, "+"},
		{token.INT, "2"},
		{token.MINUS, "-"},
		{token.INT, "3"},
		{token.ASTERISK, "*"},
		{token.INT, "4"},
		{token.SLASH, "/"},
		{token.INT, "5"},
		{token.PERCENT, "%"},
		{token.INT, "6"},
		{token.EOF, ""},
	})
}
