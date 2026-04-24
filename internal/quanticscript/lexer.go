package quanticscript

import (
	"fmt"
	"unicode"
)

// Lexer tokenizes QuanticScript source code
type Lexer struct {
	input    string
	filename string
	position int  // current position in input
	readPos  int  // current reading position (after current char)
	ch       byte // current char under examination
	line     int
	column   int
}

// NewLexer creates a new lexer for the given input
func NewLexer(input, filename string) *Lexer {
	l := &Lexer{
		input:    input,
		filename: filename,
		line:     1,
		column:   0,
	}
	l.readChar()
	return l
}

// readChar reads the next character and advances position
func (l *Lexer) readChar() {
	if l.readPos >= len(l.input) {
		l.ch = 0 // ASCII NUL
	} else {
		l.ch = l.input[l.readPos]
	}
	l.position = l.readPos
	l.readPos++
	l.column++
	if l.ch == '\n' {
		l.line++
		l.column = 0
	}
}

// peekChar returns the next character without advancing
func (l *Lexer) peekChar() byte {
	if l.readPos >= len(l.input) {
		return 0
	}
	return l.input[l.readPos]
}

// currentLocation returns the current source location
func (l *Lexer) currentLocation() SourceLocation {
	return SourceLocation{
		Filename: l.filename,
		Line:     l.line,
		Column:   l.column,
		Offset:   l.position,
	}
}

// NextToken returns the next token from the input
func (l *Lexer) NextToken() Token {
	l.skipWhitespace()

	loc := l.currentLocation()

	var tok Token
	tok.Location = loc

	switch l.ch {
	case 0:
		tok.Type = TOKEN_EOF
		tok.Literal = ""
	case '+':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok.Type = TOKEN_PLUS_ASSIGN
			tok.Literal = string(ch) + string(l.ch)
		} else {
			tok.Type = TOKEN_PLUS
			tok.Literal = string(l.ch)
		}
	case '-':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok.Type = TOKEN_MINUS_ASSIGN
			tok.Literal = string(ch) + string(l.ch)
		} else {
			tok.Type = TOKEN_MINUS
			tok.Literal = string(l.ch)
		}
	case '*':
		tok.Type = TOKEN_STAR
		tok.Literal = string(l.ch)
	case '/':
		if l.peekChar() == '/' {
			l.skipLineComment()
			return l.NextToken()
		} else if l.peekChar() == '*' {
			l.skipBlockComment()
			return l.NextToken()
		}
		tok.Type = TOKEN_SLASH
		tok.Literal = string(l.ch)
	case '%':
		tok.Type = TOKEN_PERCENT
		tok.Literal = string(l.ch)
	case '&':
		if l.peekChar() == '&' {
			ch := l.ch
			l.readChar()
			tok.Type = TOKEN_AND
			tok.Literal = string(ch) + string(l.ch)
		} else {
			tok.Type = TOKEN_AMPERSAND
			tok.Literal = string(l.ch)
		}
	case '|':
		if l.peekChar() == '|' {
			ch := l.ch
			l.readChar()
			tok.Type = TOKEN_OR
			tok.Literal = string(ch) + string(l.ch)
		} else {
			tok.Type = TOKEN_PIPE
			tok.Literal = string(l.ch)
		}
	case '^':
		tok.Type = TOKEN_CARET
		tok.Literal = string(l.ch)
	case '~':
		tok.Type = TOKEN_TILDE
		tok.Literal = string(l.ch)
	case '<':
		if l.peekChar() == '<' {
			ch := l.ch
			l.readChar()
			tok.Type = TOKEN_LSHIFT
			tok.Literal = string(ch) + string(l.ch)
		} else if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok.Type = TOKEN_LTE
			tok.Literal = string(ch) + string(l.ch)
		} else {
			tok.Type = TOKEN_LT
			tok.Literal = string(l.ch)
		}
	case '>':
		if l.peekChar() == '>' {
			ch := l.ch
			l.readChar()
			tok.Type = TOKEN_RSHIFT
			tok.Literal = string(ch) + string(l.ch)
		} else if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok.Type = TOKEN_GTE
			tok.Literal = string(ch) + string(l.ch)
		} else {
			tok.Type = TOKEN_GT
			tok.Literal = string(l.ch)
		}
	case '=':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok.Type = TOKEN_EQ
			tok.Literal = string(ch) + string(l.ch)
		} else if l.peekChar() == '>' {
			ch := l.ch
			l.readChar()
			tok.Type = TOKEN_ARROW
			tok.Literal = string(ch) + string(l.ch)
		} else {
			tok.Type = TOKEN_ASSIGN
			tok.Literal = string(l.ch)
		}
	case '!':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok.Type = TOKEN_NEQ
			tok.Literal = string(ch) + string(l.ch)
		} else {
			tok.Type = TOKEN_NOT
			tok.Literal = string(l.ch)
		}
	case '?':
		tok.Type = TOKEN_QUESTION
		tok.Literal = string(l.ch)
	case '(':
		tok.Type = TOKEN_LPAREN
		tok.Literal = string(l.ch)
	case ')':
		tok.Type = TOKEN_RPAREN
		tok.Literal = string(l.ch)
	case '{':
		tok.Type = TOKEN_LBRACE
		tok.Literal = string(l.ch)
	case '}':
		tok.Type = TOKEN_RBRACE
		tok.Literal = string(l.ch)
	case '[':
		tok.Type = TOKEN_LBRACKET
		tok.Literal = string(l.ch)
	case ']':
		tok.Type = TOKEN_RBRACKET
		tok.Literal = string(l.ch)
	case ',':
		tok.Type = TOKEN_COMMA
		tok.Literal = string(l.ch)
	case ';':
		tok.Type = TOKEN_SEMICOLON
		tok.Literal = string(l.ch)
	case ':':
		tok.Type = TOKEN_COLON
		tok.Literal = string(l.ch)
	case '.':
		tok.Type = TOKEN_DOT
		tok.Literal = string(l.ch)
	case '"':
		tok.Type = TOKEN_STRING
		tok.Literal = l.readString()
		return tok
	default:
		if isLetter(l.ch) {
			tok.Literal = l.readIdentifier()
			tok.Type = lookupIdent(tok.Literal)
			return tok
		} else if isDigit(l.ch) {
			tok.Type = TOKEN_INT
			tok.Literal = l.readNumber()
			return tok
		} else {
			tok.Type = TOKEN_ILLEGAL
			tok.Literal = string(l.ch)
		}
	}

	l.readChar()
	return tok
}

// skipWhitespace skips whitespace characters
func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		l.readChar()
	}
}

// skipLineComment skips a line comment (// ...)
func (l *Lexer) skipLineComment() {
	for l.ch != '\n' && l.ch != 0 {
		l.readChar()
	}
}

// skipBlockComment skips a block comment (/* ... */)
func (l *Lexer) skipBlockComment() {
	l.readChar() // skip '/'
	l.readChar() // skip '*'

	for {
		if l.ch == 0 {
			break
		}
		if l.ch == '*' && l.peekChar() == '/' {
			l.readChar() // skip '*'
			l.readChar() // skip '/'
			break
		}
		l.readChar()
	}
}

// readIdentifier reads an identifier or keyword
func (l *Lexer) readIdentifier() string {
	position := l.position
	for isLetter(l.ch) || isDigit(l.ch) {
		l.readChar()
	}
	return l.input[position:l.position]
}

// readNumber reads a number literal
func (l *Lexer) readNumber() string {
	position := l.position

	// Handle hex numbers (0x...)
	if l.ch == '0' && (l.peekChar() == 'x' || l.peekChar() == 'X') {
		l.readChar() // skip '0'
		l.readChar() // skip 'x'
		for isHexDigit(l.ch) {
			l.readChar()
		}
		return l.input[position:l.position]
	}

	// Handle decimal numbers
	for isDigit(l.ch) {
		l.readChar()
	}

	return l.input[position:l.position]
}

// readString reads a string literal
func (l *Lexer) readString() string {
	position := l.position + 1 // skip opening quote

	for {
		l.readChar()
		if l.ch == '"' {
			str := l.input[position:l.position]
			l.readChar() // skip closing quote
			return str
		}
		if l.ch == 0 {
			break
		}
		if l.ch == '\\' {
			l.readChar() // skip escape character
		}
	}

	return l.input[position:l.position]
}

// lookupIdent checks if an identifier is a keyword
func lookupIdent(ident string) TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return TOKEN_IDENT
}

// isLetter checks if a character is a letter or underscore
func isLetter(ch byte) bool {
	return unicode.IsLetter(rune(ch)) || ch == '_'
}

// isDigit checks if a character is a digit
func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}

// isHexDigit checks if a character is a hexadecimal digit
func isHexDigit(ch byte) bool {
	return ('0' <= ch && ch <= '9') ||
		('a' <= ch && ch <= 'f') ||
		('A' <= ch && ch <= 'F')
}

// LexerError represents a lexical error
type LexerError struct {
	Location SourceLocation
	Message  string
}

func (e *LexerError) Error() string {
	return fmt.Sprintf("%s: %s", e.Location, e.Message)
}
