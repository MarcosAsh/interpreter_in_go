package lexer

import (
	"fmt"
	"pearl/token"
)

type Lexer struct {
	input   string
	pos     int  // current position
	readPos int  // next position
	ch      byte // current char
	line    int
	col     int
}

func New(input string) *Lexer {
	l := &Lexer{input: input, line: 1, col: 0}
	l.readChar()
	return l
}

func (l *Lexer) readChar() {
	if l.readPos >= len(l.input) {
		l.ch = 0
	} else {
		l.ch = l.input[l.readPos]
	}
	l.pos = l.readPos
	l.readPos++
	l.col++

	if l.ch == '\n' {
		l.line++
		l.col = 0
	}
}

func (l *Lexer) peekChar() byte {
	if l.readPos >= len(l.input) {
		return 0
	}
	return l.input[l.readPos]
}

func (l *Lexer) NextToken() token.Token {
	var tok token.Token

	l.skipWhitespace()

	tok.Line = l.line
	tok.Col = l.col

	switch l.ch {
	case '=':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = token.Token{Type: token.EQ, Literal: string(ch) + string(l.ch), Line: l.line, Col: l.col}
		} else if l.peekChar() == '>' {
			l.readChar()
			tok = token.Token{Type: token.ARROW, Literal: "=>", Line: l.line, Col: l.col}
		} else {
			tok = l.newToken(token.ASSIGN, l.ch)
		}
	case '+':
		if l.peekChar() == '+' {
			l.readChar()
			tok = token.Token{Type: token.CONCAT, Literal: "++", Line: l.line, Col: l.col}
		} else {
			tok = l.newToken(token.PLUS, l.ch)
		}
	case '-':
		tok = l.newToken(token.MINUS, l.ch)
	case '!':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = token.Token{Type: token.NOT_EQ, Literal: string(ch) + string(l.ch), Line: l.line, Col: l.col}
		} else if l.peekChar() == '~' {
			l.readChar()
			tok = token.Token{Type: token.NOTMATCH, Literal: "!~", Line: l.line, Col: l.col}
		} else {
			tok = l.newToken(token.BANG, l.ch)
		}
	case '*':
		tok = l.newToken(token.ASTERISK, l.ch)
	case '/':
		// could be division or regex
		// for now treat as division, parser will handle context
		tok = l.newToken(token.SLASH, l.ch)
	case '%':
		tok = l.newToken(token.PERCENT, l.ch)
	case '<':
		if l.peekChar() == '=' {
			l.readChar()
			tok = token.Token{Type: token.LTE, Literal: "<=", Line: l.line, Col: l.col}
		} else {
			tok = l.newToken(token.LT, l.ch)
		}
	case '>':
		if l.peekChar() == '=' {
			l.readChar()
			tok = token.Token{Type: token.GTE, Literal: ">=", Line: l.line, Col: l.col}
		} else {
			tok = l.newToken(token.GT, l.ch)
		}
	case '~':
		tok = l.newToken(token.MATCH, l.ch)
	case '.':
		if l.peekChar() == '.' {
			l.readChar()
			tok = token.Token{Type: token.RANGE, Literal: "..", Line: l.line, Col: l.col}
		} else {
			tok = l.newToken(token.ILLEGAL, l.ch)
		}
	case '|':
		if l.peekChar() == '>' {
			l.readChar()
			tok = token.Token{Type: token.PIPE, Literal: "|>", Line: l.line, Col: l.col}
		} else {
			tok = l.newToken(token.ILLEGAL, l.ch)
		}
	case ';':
		tok = l.newToken(token.SEMICOLON, l.ch)
	case ':':
		tok = l.newToken(token.COLON, l.ch)
	case ',':
		tok = l.newToken(token.COMMA, l.ch)
	case '(':
		tok = l.newToken(token.LPAREN, l.ch)
	case ')':
		tok = l.newToken(token.RPAREN, l.ch)
	case '{':
		tok = l.newToken(token.LBRACE, l.ch)
	case '}':
		tok = l.newToken(token.RBRACE, l.ch)
	case '[':
		tok = l.newToken(token.LBRACKET, l.ch)
	case ']':
		tok = l.newToken(token.RBRACKET, l.ch)
	case '"':
		tok.Type = token.STRING
		tok.Literal = l.readString()
		tok.Line = l.line
		tok.Col = l.col
		return tok
	case '#':
		l.skipComment()
		return l.NextToken()
	case '\n':
		tok = l.newToken(token.NEWLINE, l.ch)
	case 0:
		tok.Literal = ""
		tok.Type = token.EOF
	default:
		if isLetter(l.ch) {
			tok.Literal = l.readIdentifier()
			tok.Type = token.LookupIdent(tok.Literal)
			tok.Line = l.line
			tok.Col = l.col
			return tok
		} else if isDigit(l.ch) {
			tok.Line = l.line
			tok.Col = l.col
			lit, isFloat := l.readNumber()
			tok.Literal = lit
			if isFloat {
				tok.Type = token.FLOAT
			} else {
				tok.Type = token.INT
			}
			return tok
		} else {
			tok = l.newToken(token.ILLEGAL, l.ch)
		}
	}

	l.readChar()
	return tok
}

func (l *Lexer) readIdentifier() string {
	pos := l.pos
	for isLetter(l.ch) || isDigit(l.ch) || l.ch == '_' {
		l.readChar()
	}
	return l.input[pos:l.pos]
}

func (l *Lexer) readNumber() (string, bool) {
	pos := l.pos
	isFloat := false

	for isDigit(l.ch) {
		l.readChar()
	}

	if l.ch == '.' && isDigit(l.peekChar()) {
		isFloat = true
		l.readChar() // consume the dot
		for isDigit(l.ch) {
			l.readChar()
		}
	}

	return l.input[pos:l.pos], isFloat
}

func (l *Lexer) readString() string {
	var result string
	l.readChar() // skip opening quote

	for l.ch != '"' && l.ch != 0 {
		if l.ch == '\\' {
			l.readChar()
			switch l.ch {
			case 'n':
				result += "\n"
			case 't':
				result += "\t"
			case 'r':
				result += "\r"
			case '"':
				result += "\""
			case '\\':
				result += "\\"
			case '{':
				result += "{"
			default:
				result += "\\" + string(l.ch)
			}
		} else {
			result += string(l.ch)
		}
		l.readChar()
	}

	// consume closing quote
	if l.ch == '"' {
		l.readChar()
	}

	return result
}

// ReadRegexFromStart reads a regex when we haven't yet tokenized the opening /
// Used when parser knows a regex is coming (after ~ or !~)
func (l *Lexer) ReadRegexFromStart() (string, error) {
	// skip whitespace first
	for l.ch == ' ' || l.ch == '\t' {
		l.readChar()
	}

	if l.ch != '/' {
		return "", fmt.Errorf("expected '/' to start regex, got '%c'", l.ch)
	}
	l.readChar() // skip opening /

	var result string
	for l.ch != '/' && l.ch != 0 && l.ch != '\n' {
		if l.ch == '\\' {
			result += string(l.ch)
			l.readChar()
			if l.ch != 0 {
				result += string(l.ch)
			}
		} else {
			result += string(l.ch)
		}
		l.readChar()
	}

	if l.ch != '/' {
		return "", fmt.Errorf("unterminated regex")
	}
	l.readChar() // skip closing /

	return result, nil
}

// ReadRegex reads a regex pattern. Called when curToken is SLASH.
// At this point the lexer has already consumed the opening / and advanced.
// So we just read until the closing /
func (l *Lexer) ReadRegex() (string, error) {
	var result string

	for l.ch != '/' && l.ch != 0 && l.ch != '\n' {
		if l.ch == '\\' {
			result += string(l.ch)
			l.readChar()
			if l.ch != 0 {
				result += string(l.ch)
			}
		} else {
			result += string(l.ch)
		}
		l.readChar()
	}

	if l.ch != '/' {
		return "", fmt.Errorf("unterminated regex")
	}
	l.readChar() // skip closing /

	return result, nil
}

func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\r' {
		l.readChar()
	}
}

func (l *Lexer) skipComment() {
	for l.ch != '\n' && l.ch != 0 {
		l.readChar()
	}
}

func (l *Lexer) newToken(tokenType token.TokenType, ch byte) token.Token {
	return token.Token{Type: tokenType, Literal: string(ch), Line: l.line, Col: l.col}
}

func isLetter(ch byte) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_'
}

func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}
func (l *Lexer) GetCh() byte { return l.ch }
