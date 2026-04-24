package quanticscript

import "fmt"

// TokenType represents the type of a token
type TokenType int

const (
	// Special tokens
	TOKEN_EOF TokenType = iota
	TOKEN_ILLEGAL

	// Identifiers and literals
	TOKEN_IDENT
	TOKEN_INT
	TOKEN_STRING
	TOKEN_BYTES

	// Keywords
	TOKEN_FUNCTION
	TOKEN_EXPORT
	TOKEN_IMPORT
	TOKEN_FROM
	TOKEN_LET
	TOKEN_CONST
	TOKEN_IF
	TOKEN_ELSE
	TOKEN_WHILE
	TOKEN_FOR
	TOKEN_RETURN
	TOKEN_STRUCT
	TOKEN_ENUM
	TOKEN_INTERFACE
	TOKEN_TYPE
	TOKEN_TRUE
	TOKEN_FALSE
	TOKEN_NULL
	TOKEN_ASM // __asm__

	// Types
	TOKEN_I8
	TOKEN_I16
	TOKEN_I32
	TOKEN_I64
	TOKEN_U8
	TOKEN_U16
	TOKEN_U32
	TOKEN_U64
	TOKEN_BOOL
	TOKEN_STRING_TYPE
	TOKEN_BYTES_TYPE
	TOKEN_VOID

	// Operators
	TOKEN_PLUS
	TOKEN_MINUS
	TOKEN_STAR
	TOKEN_SLASH
	TOKEN_PERCENT
	TOKEN_AMPERSAND
	TOKEN_PIPE
	TOKEN_CARET
	TOKEN_TILDE
	TOKEN_LSHIFT
	TOKEN_RSHIFT
	TOKEN_EQ
	TOKEN_NEQ
	TOKEN_LT
	TOKEN_GT
	TOKEN_LTE
	TOKEN_GTE
	TOKEN_AND
	TOKEN_OR
	TOKEN_NOT
	TOKEN_ASSIGN
	TOKEN_PLUS_ASSIGN
	TOKEN_MINUS_ASSIGN
	TOKEN_QUESTION

	// Delimiters
	TOKEN_LPAREN
	TOKEN_RPAREN
	TOKEN_LBRACE
	TOKEN_RBRACE
	TOKEN_LBRACKET
	TOKEN_RBRACKET
	TOKEN_COMMA
	TOKEN_SEMICOLON
	TOKEN_COLON
	TOKEN_DOT
	TOKEN_ARROW
)

var tokenNames = map[TokenType]string{
	TOKEN_EOF:     "EOF",
	TOKEN_ILLEGAL: "ILLEGAL",
	TOKEN_IDENT:   "IDENT",
	TOKEN_INT:     "INT",
	TOKEN_STRING:  "STRING",
	TOKEN_BYTES:   "BYTES",

	TOKEN_FUNCTION:  "function",
	TOKEN_EXPORT:    "export",
	TOKEN_IMPORT:    "import",
	TOKEN_FROM:      "from",
	TOKEN_LET:       "let",
	TOKEN_CONST:     "const",
	TOKEN_IF:        "if",
	TOKEN_ELSE:      "else",
	TOKEN_WHILE:     "while",
	TOKEN_FOR:       "for",
	TOKEN_RETURN:    "return",
	TOKEN_STRUCT:    "struct",
	TOKEN_ENUM:      "enum",
	TOKEN_INTERFACE: "interface",
	TOKEN_TYPE:      "type",
	TOKEN_TRUE:      "true",
	TOKEN_FALSE:     "false",
	TOKEN_NULL:      "null",
	TOKEN_ASM:       "__asm__",

	TOKEN_I8:          "i8",
	TOKEN_I16:         "i16",
	TOKEN_I32:         "i32",
	TOKEN_I64:         "i64",
	TOKEN_U8:          "u8",
	TOKEN_U16:         "u16",
	TOKEN_U32:         "u32",
	TOKEN_U64:         "u64",
	TOKEN_BOOL:        "bool",
	TOKEN_STRING_TYPE: "string",
	TOKEN_BYTES_TYPE:  "bytes",
	TOKEN_VOID:        "void",

	TOKEN_PLUS:         "+",
	TOKEN_MINUS:        "-",
	TOKEN_STAR:         "*",
	TOKEN_SLASH:        "/",
	TOKEN_PERCENT:      "%",
	TOKEN_AMPERSAND:    "&",
	TOKEN_PIPE:         "|",
	TOKEN_CARET:        "^",
	TOKEN_TILDE:        "~",
	TOKEN_LSHIFT:       "<<",
	TOKEN_RSHIFT:       ">>",
	TOKEN_EQ:           "==",
	TOKEN_NEQ:          "!=",
	TOKEN_LT:           "<",
	TOKEN_GT:           ">",
	TOKEN_LTE:          "<=",
	TOKEN_GTE:          ">=",
	TOKEN_AND:          "&&",
	TOKEN_OR:           "||",
	TOKEN_NOT:          "!",
	TOKEN_ASSIGN:       "=",
	TOKEN_PLUS_ASSIGN:  "+=",
	TOKEN_MINUS_ASSIGN: "-=",
	TOKEN_QUESTION:     "?",

	TOKEN_LPAREN:    "(",
	TOKEN_RPAREN:    ")",
	TOKEN_LBRACE:    "{",
	TOKEN_RBRACE:    "}",
	TOKEN_LBRACKET:  "[",
	TOKEN_RBRACKET:  "]",
	TOKEN_COMMA:     ",",
	TOKEN_SEMICOLON: ";",
	TOKEN_COLON:     ":",
	TOKEN_DOT:       ".",
	TOKEN_ARROW:     "=>",
}

func (t TokenType) String() string {
	if name, ok := tokenNames[t]; ok {
		return name
	}
	return fmt.Sprintf("TokenType(%d)", t)
}

var keywords = map[string]TokenType{
	"function":  TOKEN_FUNCTION,
	"export":    TOKEN_EXPORT,
	"import":    TOKEN_IMPORT,
	"from":      TOKEN_FROM,
	"let":       TOKEN_LET,
	"const":     TOKEN_CONST,
	"if":        TOKEN_IF,
	"else":      TOKEN_ELSE,
	"while":     TOKEN_WHILE,
	"for":       TOKEN_FOR,
	"return":    TOKEN_RETURN,
	"struct":    TOKEN_STRUCT,
	"enum":      TOKEN_ENUM,
	"interface": TOKEN_INTERFACE,
	"type":      TOKEN_TYPE,
	"true":      TOKEN_TRUE,
	"false":     TOKEN_FALSE,
	"null":      TOKEN_NULL,
	"__asm__":   TOKEN_ASM,
	"i8":        TOKEN_I8,
	"i16":       TOKEN_I16,
	"i32":       TOKEN_I32,
	"i64":       TOKEN_I64,
	"u8":        TOKEN_U8,
	"u16":       TOKEN_U16,
	"u32":       TOKEN_U32,
	"u64":       TOKEN_U64,
	"bool":      TOKEN_BOOL,
	"string":    TOKEN_STRING_TYPE,
	"bytes":     TOKEN_BYTES_TYPE,
	"void":      TOKEN_VOID,
}

// SourceLocation represents a position in the source code
type SourceLocation struct {
	Filename string
	Line     int
	Column   int
	Offset   int
}

func (loc SourceLocation) String() string {
	return fmt.Sprintf("%s:%d:%d", loc.Filename, loc.Line, loc.Column)
}

// Token represents a lexical token
type Token struct {
	Type     TokenType
	Literal  string
	Location SourceLocation
}

func (t Token) String() string {
	return fmt.Sprintf("%s(%q) at %s", t.Type, t.Literal, t.Location)
}
