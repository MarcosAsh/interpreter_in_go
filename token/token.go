package token

type TokenType string

type Token struct {
	Type    TokenType
	Literal string
	Line    int
	Col     int
}

func (t Token) String() string {
	return t.Literal
}

const (
	ILLEGAL = "ILLEGAL"
	EOF     = "EOF"

	// literals
	IDENT  = "IDENT"
	INT    = "INT"
	FLOAT  = "FLOAT"
	STRING = "STRING"
	REGEX  = "REGEX"

	// operators
	ASSIGN   = "="
	PLUS     = "+"
	MINUS    = "-"
	BANG     = "!"
	ASTERISK = "*"
	SLASH    = "/"
	PERCENT  = "%"
	LT       = "<"
	GT       = ">"
	EQ       = "=="
	NOT_EQ   = "!="
	LTE      = "<="
	GTE      = ">="
	CONCAT   = "++"
	PIPE     = "|>"
	MATCH    = "~"
	NOTMATCH = "!~"
	RANGE    = ".."

	// delimiters
	COMMA     = ","
	COLON     = ":"
	SEMICOLON = ";"
	NEWLINE   = "NEWLINE"
	LPAREN    = "("
	RPAREN    = ")"
	LBRACE    = "{"
	RBRACE    = "}"
	LBRACKET  = "["
	RBRACKET  = "]"

	// keywords
	LET      = "LET"
	FN       = "FN"
	TRUE     = "TRUE"
	FALSE    = "FALSE"
	IF       = "IF"
	ELSE     = "ELSE"
	RETURN   = "RETURN"
	FOR      = "FOR"
	IN       = "IN"
	WHILE    = "WHILE"
	AND      = "AND"
	OR       = "OR"
	NOT      = "NOT"
	NULL     = "NULL"
	MATCH_KW = "MATCH_KW"
	TRY      = "TRY"
	CATCH    = "CATCH"
	ARROW    = "=>"
)

var keywords = map[string]TokenType{
	"fn":     FN,
	"let":    LET,
	"true":   TRUE,
	"false":  FALSE,
	"if":     IF,
	"else":   ELSE,
	"return": RETURN,
	"for":    FOR,
	"in":     IN,
	"while":  WHILE,
	"and":    AND,
	"or":     OR,
	"not":    NOT,
	"null":   NULL,
	"try":    TRY,
	"catch":  CATCH,
}

func LookupIdent(ident string) TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return IDENT
}
