package quanticscript

import (
	"testing"
)

func TestLexerBasicTokens(t *testing.T) {
	input := `let x = 42;`

	tests := []struct {
		expectedType    TokenType
		expectedLiteral string
	}{
		{TOKEN_LET, "let"},
		{TOKEN_IDENT, "x"},
		{TOKEN_ASSIGN, "="},
		{TOKEN_INT, "42"},
		{TOKEN_SEMICOLON, ";"},
		{TOKEN_EOF, ""},
	}

	l := NewLexer(input, "test.qs")

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
	input := `function export import from let const if else while for return struct enum interface type true false null __asm__`

	expectedTypes := []TokenType{
		TOKEN_FUNCTION, TOKEN_EXPORT, TOKEN_IMPORT, TOKEN_FROM,
		TOKEN_LET, TOKEN_CONST, TOKEN_IF, TOKEN_ELSE, TOKEN_WHILE,
		TOKEN_FOR, TOKEN_RETURN, TOKEN_STRUCT, TOKEN_ENUM,
		TOKEN_INTERFACE, TOKEN_TYPE, TOKEN_TRUE, TOKEN_FALSE,
		TOKEN_NULL, TOKEN_ASM, TOKEN_EOF,
	}

	l := NewLexer(input, "test.qs")

	for i, expectedType := range expectedTypes {
		tok := l.NextToken()
		if tok.Type != expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, expectedType, tok.Type)
		}
	}
}

func TestLexerTypes(t *testing.T) {
	input := `i8 i16 i32 i64 u8 u16 u32 u64 bool string bytes void`

	expectedTypes := []TokenType{
		TOKEN_I8, TOKEN_I16, TOKEN_I32, TOKEN_I64,
		TOKEN_U8, TOKEN_U16, TOKEN_U32, TOKEN_U64,
		TOKEN_BOOL, TOKEN_STRING_TYPE, TOKEN_BYTES_TYPE, TOKEN_VOID,
		TOKEN_EOF,
	}

	l := NewLexer(input, "test.qs")

	for i, expectedType := range expectedTypes {
		tok := l.NextToken()
		if tok.Type != expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, expectedType, tok.Type)
		}
	}
}

func TestLexerOperators(t *testing.T) {
	input := `+ - * / % & | ^ ~ << >> == != < > <= >= && || ! = += -= ?`

	expectedTypes := []TokenType{
		TOKEN_PLUS, TOKEN_MINUS, TOKEN_STAR, TOKEN_SLASH, TOKEN_PERCENT,
		TOKEN_AMPERSAND, TOKEN_PIPE, TOKEN_CARET, TOKEN_TILDE,
		TOKEN_LSHIFT, TOKEN_RSHIFT, TOKEN_EQ, TOKEN_NEQ,
		TOKEN_LT, TOKEN_GT, TOKEN_LTE, TOKEN_GTE,
		TOKEN_AND, TOKEN_OR, TOKEN_NOT, TOKEN_ASSIGN,
		TOKEN_PLUS_ASSIGN, TOKEN_MINUS_ASSIGN, TOKEN_QUESTION,
		TOKEN_EOF,
	}

	l := NewLexer(input, "test.qs")

	for i, expectedType := range expectedTypes {
		tok := l.NextToken()
		if tok.Type != expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, expectedType, tok.Type)
		}
	}
}

func TestLexerDelimiters(t *testing.T) {
	input := `( ) { } [ ] , ; : . =>`

	expectedTypes := []TokenType{
		TOKEN_LPAREN, TOKEN_RPAREN, TOKEN_LBRACE, TOKEN_RBRACE,
		TOKEN_LBRACKET, TOKEN_RBRACKET, TOKEN_COMMA, TOKEN_SEMICOLON,
		TOKEN_COLON, TOKEN_DOT, TOKEN_ARROW, TOKEN_EOF,
	}

	l := NewLexer(input, "test.qs")

	for i, expectedType := range expectedTypes {
		tok := l.NextToken()
		if tok.Type != expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, expectedType, tok.Type)
		}
	}
}

func TestLexerStrings(t *testing.T) {
	input := `"hello world" "test"`

	tests := []struct {
		expectedType    TokenType
		expectedLiteral string
	}{
		{TOKEN_STRING, "hello world"},
		{TOKEN_STRING, "test"},
		{TOKEN_EOF, ""},
	}

	l := NewLexer(input, "test.qs")

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
	input := `42 0 123456 0xFF 0x1a2b`

	tests := []struct {
		expectedType    TokenType
		expectedLiteral string
	}{
		{TOKEN_INT, "42"},
		{TOKEN_INT, "0"},
		{TOKEN_INT, "123456"},
		{TOKEN_INT, "0xFF"},
		{TOKEN_INT, "0x1a2b"},
		{TOKEN_EOF, ""},
	}

	l := NewLexer(input, "test.qs")

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

func TestLexerComments(t *testing.T) {
	input := `let x = 42; // line comment
let y = 10; /* block comment */ let z = 5;`

	tests := []struct {
		expectedType    TokenType
		expectedLiteral string
	}{
		{TOKEN_LET, "let"},
		{TOKEN_IDENT, "x"},
		{TOKEN_ASSIGN, "="},
		{TOKEN_INT, "42"},
		{TOKEN_SEMICOLON, ";"},
		{TOKEN_LET, "let"},
		{TOKEN_IDENT, "y"},
		{TOKEN_ASSIGN, "="},
		{TOKEN_INT, "10"},
		{TOKEN_SEMICOLON, ";"},
		{TOKEN_LET, "let"},
		{TOKEN_IDENT, "z"},
		{TOKEN_ASSIGN, "="},
		{TOKEN_INT, "5"},
		{TOKEN_SEMICOLON, ";"},
		{TOKEN_EOF, ""},
	}

	l := NewLexer(input, "test.qs")

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

func TestLexerSourceLocation(t *testing.T) {
	input := `let x = 42;
let y = 10;`

	l := NewLexer(input, "test.qs")

	tok := l.NextToken() // let
	if tok.Location.Line != 1 || tok.Location.Column != 1 {
		t.Fatalf("wrong location for 'let'. expected line=1, col=1, got line=%d, col=%d",
			tok.Location.Line, tok.Location.Column)
	}

	l.NextToken() // x
	l.NextToken() // =
	l.NextToken() // 42
	l.NextToken() // ;

	tok = l.NextToken() // let on line 2
	if tok.Location.Line != 2 {
		t.Fatalf("wrong location for second 'let'. expected line=2, got line=%d",
			tok.Location.Line)
	}
}

func TestLexerAsmKeyword(t *testing.T) {
	input := `__asm__ { PUSH 42 }`

	tests := []struct {
		expectedType    TokenType
		expectedLiteral string
	}{
		{TOKEN_ASM, "__asm__"},
		{TOKEN_LBRACE, "{"},
		{TOKEN_IDENT, "PUSH"},
		{TOKEN_INT, "42"},
		{TOKEN_RBRACE, "}"},
		{TOKEN_EOF, ""},
	}

	l := NewLexer(input, "test.qs")

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

func TestLexerFunctionDeclaration(t *testing.T) {
	input := `export function entry(ctx: InstructionContext): void {
    return;
}`

	expectedTypes := []TokenType{
		TOKEN_EXPORT, TOKEN_FUNCTION, TOKEN_IDENT, TOKEN_LPAREN,
		TOKEN_IDENT, TOKEN_COLON, TOKEN_IDENT, TOKEN_RPAREN,
		TOKEN_COLON, TOKEN_VOID, TOKEN_LBRACE, TOKEN_RETURN,
		TOKEN_SEMICOLON, TOKEN_RBRACE, TOKEN_EOF,
	}

	l := NewLexer(input, "test.qs")

	for i, expectedType := range expectedTypes {
		tok := l.NextToken()
		if tok.Type != expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q (literal=%q)",
				i, expectedType, tok.Type, tok.Literal)
		}
	}
}
