package quanticscript

import (
	"testing"
)

func TestParserFunctionDeclaration(t *testing.T) {
	input := `export function entry(ctx: InstructionContext): void {
    return;
}`

	lexer := NewLexer(input, "test.qs")
	parser := NewParser(lexer)
	program := parser.ParseProgram()

	checkParserErrors(t, parser)

	if len(program.Declarations) != 1 {
		t.Fatalf("expected 1 declaration, got %d", len(program.Declarations))
	}

	fn, ok := program.Declarations[0].(*FunctionDecl)
	if !ok {
		t.Fatalf("expected FunctionDecl, got %T", program.Declarations[0])
	}

	if fn.Name != "entry" {
		t.Errorf("expected function name 'entry', got %s", fn.Name)
	}

	if !fn.IsExport {
		t.Errorf("expected function to be exported")
	}

	if len(fn.Parameters) != 1 {
		t.Fatalf("expected 1 parameter, got %d", len(fn.Parameters))
	}

	if fn.Parameters[0].Name != "ctx" {
		t.Errorf("expected parameter name 'ctx', got %s", fn.Parameters[0].Name)
	}

	if fn.ReturnType == nil || fn.ReturnType.Name != "void" {
		t.Errorf("expected return type 'void'")
	}
}

func TestParserLetStatement(t *testing.T) {
	input := `function test(): void {
    let x: i64 = 42;
    const y = 10;
}`

	lexer := NewLexer(input, "test.qs")
	parser := NewParser(lexer)
	program := parser.ParseProgram()

	checkParserErrors(t, parser)

	fn := program.Declarations[0].(*FunctionDecl)
	if len(fn.Body.Statements) != 2 {
		t.Fatalf("expected 2 statements, got %d", len(fn.Body.Statements))
	}

	// Check first let statement
	letStmt, ok := fn.Body.Statements[0].(*LetStmt)
	if !ok {
		t.Fatalf("expected LetStmt, got %T", fn.Body.Statements[0])
	}

	if letStmt.Name != "x" {
		t.Errorf("expected variable name 'x', got %s", letStmt.Name)
	}

	if letStmt.IsConst {
		t.Errorf("expected let, not const")
	}

	if letStmt.Type == nil || letStmt.Type.Name != "i64" {
		t.Errorf("expected type 'i64'")
	}

	// Check const statement
	constStmt, ok := fn.Body.Statements[1].(*LetStmt)
	if !ok {
		t.Fatalf("expected LetStmt, got %T", fn.Body.Statements[1])
	}

	if !constStmt.IsConst {
		t.Errorf("expected const")
	}
}

func TestParserIfStatement(t *testing.T) {
	input := `function test(): void {
    if (x > 10) {
        return;
    } else {
        return;
    }
}`

	lexer := NewLexer(input, "test.qs")
	parser := NewParser(lexer)
	program := parser.ParseProgram()

	checkParserErrors(t, parser)

	fn := program.Declarations[0].(*FunctionDecl)
	ifStmt, ok := fn.Body.Statements[0].(*IfStmt)
	if !ok {
		t.Fatalf("expected IfStmt, got %T", fn.Body.Statements[0])
	}

	if ifStmt.Condition == nil {
		t.Errorf("expected condition")
	}

	if ifStmt.ThenBlock == nil {
		t.Errorf("expected then block")
	}

	if ifStmt.ElseBlock == nil {
		t.Errorf("expected else block")
	}
}

func TestParserWhileStatement(t *testing.T) {
	input := `function test(): void {
    while (x < 100) {
        x = x + 1;
    }
}`

	lexer := NewLexer(input, "test.qs")
	parser := NewParser(lexer)
	program := parser.ParseProgram()

	checkParserErrors(t, parser)

	fn := program.Declarations[0].(*FunctionDecl)
	whileStmt, ok := fn.Body.Statements[0].(*WhileStmt)
	if !ok {
		t.Fatalf("expected WhileStmt, got %T", fn.Body.Statements[0])
	}

	if whileStmt.Condition == nil {
		t.Errorf("expected condition")
	}

	if whileStmt.Body == nil {
		t.Errorf("expected body")
	}
}

func TestParserForStatement(t *testing.T) {
	input := `function test(): void {
    for (let i = 0; i < 10; i = i + 1) {
        return;
    }
}`

	lexer := NewLexer(input, "test.qs")
	parser := NewParser(lexer)
	program := parser.ParseProgram()

	checkParserErrors(t, parser)

	fn := program.Declarations[0].(*FunctionDecl)
	forStmt, ok := fn.Body.Statements[0].(*ForStmt)
	if !ok {
		t.Fatalf("expected ForStmt, got %T", fn.Body.Statements[0])
	}

	if forStmt.Init == nil {
		t.Errorf("expected init")
	}

	if forStmt.Condition == nil {
		t.Errorf("expected condition")
	}

	if forStmt.Update == nil {
		t.Errorf("expected update")
	}

	if forStmt.Body == nil {
		t.Errorf("expected body")
	}
}

func TestParserAsmBlock(t *testing.T) {
	input := `function test(): void {
    __asm__ {
        PUSH 42
        PUSH 10
        ADD
    }
}`

	lexer := NewLexer(input, "test.qs")
	parser := NewParser(lexer)
	program := parser.ParseProgram()

	checkParserErrors(t, parser)

	fn := program.Declarations[0].(*FunctionDecl)
	asmBlock, ok := fn.Body.Statements[0].(*AsmBlock)
	if !ok {
		t.Fatalf("expected AsmBlock, got %T", fn.Body.Statements[0])
	}

	if len(asmBlock.Instructions) != 3 {
		t.Fatalf("expected 3 instructions, got %d", len(asmBlock.Instructions))
	}

	if asmBlock.Instructions[0].Mnemonic != "PUSH" {
		t.Errorf("expected mnemonic 'PUSH', got %s", asmBlock.Instructions[0].Mnemonic)
	}

	if len(asmBlock.Instructions[0].Operands) != 1 {
		t.Errorf("expected 1 operand, got %d", len(asmBlock.Instructions[0].Operands))
	}
}

func TestParserBinaryExpression(t *testing.T) {
	tests := []struct {
		input    string
		operator TokenType
	}{
		{"let x = 1 + 2;", TOKEN_PLUS},
		{"let x = 1 - 2;", TOKEN_MINUS},
		{"let x = 1 * 2;", TOKEN_STAR},
		{"let x = 1 / 2;", TOKEN_SLASH},
		{"let x = 1 % 2;", TOKEN_PERCENT},
		{"let x = 1 == 2;", TOKEN_EQ},
		{"let x = 1 != 2;", TOKEN_NEQ},
		{"let x = 1 < 2;", TOKEN_LT},
		{"let x = 1 > 2;", TOKEN_GT},
		{"let x = 1 <= 2;", TOKEN_LTE},
		{"let x = 1 >= 2;", TOKEN_GTE},
		{"let x = 1 && 2;", TOKEN_AND},
		{"let x = 1 || 2;", TOKEN_OR},
		{"let x = 1 & 2;", TOKEN_AMPERSAND},
		{"let x = 1 | 2;", TOKEN_PIPE},
		{"let x = 1 ^ 2;", TOKEN_CARET},
		{"let x = 1 << 2;", TOKEN_LSHIFT},
		{"let x = 1 >> 2;", TOKEN_RSHIFT},
	}

	for _, tt := range tests {
		input := "function test(): void { " + tt.input + " }"
		lexer := NewLexer(input, "test.qs")
		parser := NewParser(lexer)
		program := parser.ParseProgram()

		checkParserErrors(t, parser)

		fn := program.Declarations[0].(*FunctionDecl)
		letStmt := fn.Body.Statements[0].(*LetStmt)
		binExpr, ok := letStmt.Value.(*BinaryExpr)
		if !ok {
			t.Fatalf("expected BinaryExpr for %s, got %T", tt.input, letStmt.Value)
		}

		if binExpr.Operator != tt.operator {
			t.Errorf("expected operator %s, got %s", tt.operator, binExpr.Operator)
		}
	}
}

func TestParserUnaryExpression(t *testing.T) {
	tests := []struct {
		input    string
		operator TokenType
	}{
		{"let x = -5;", TOKEN_MINUS},
		{"let x = !true;", TOKEN_NOT},
		{"let x = ~5;", TOKEN_TILDE},
	}

	for _, tt := range tests {
		input := "function test(): void { " + tt.input + " }"
		lexer := NewLexer(input, "test.qs")
		parser := NewParser(lexer)
		program := parser.ParseProgram()

		checkParserErrors(t, parser)

		fn := program.Declarations[0].(*FunctionDecl)
		letStmt := fn.Body.Statements[0].(*LetStmt)
		unaryExpr, ok := letStmt.Value.(*UnaryExpr)
		if !ok {
			t.Fatalf("expected UnaryExpr for %s, got %T", tt.input, letStmt.Value)
		}

		if unaryExpr.Operator != tt.operator {
			t.Errorf("expected operator %s, got %s", tt.operator, unaryExpr.Operator)
		}
	}
}

func TestParserCallExpression(t *testing.T) {
	input := `function test(): void {
    let result = add(1, 2, 3);
}`

	lexer := NewLexer(input, "test.qs")
	parser := NewParser(lexer)
	program := parser.ParseProgram()

	checkParserErrors(t, parser)

	fn := program.Declarations[0].(*FunctionDecl)
	letStmt := fn.Body.Statements[0].(*LetStmt)
	callExpr, ok := letStmt.Value.(*CallExpr)
	if !ok {
		t.Fatalf("expected CallExpr, got %T", letStmt.Value)
	}

	if len(callExpr.Arguments) != 3 {
		t.Errorf("expected 3 arguments, got %d", len(callExpr.Arguments))
	}
}

func TestParserArrayExpression(t *testing.T) {
	input := `function test(): void {
    let arr = [1, 2, 3];
    let elem = arr[0];
}`

	lexer := NewLexer(input, "test.qs")
	parser := NewParser(lexer)
	program := parser.ParseProgram()

	checkParserErrors(t, parser)

	fn := program.Declarations[0].(*FunctionDecl)

	// Check array literal
	letStmt1 := fn.Body.Statements[0].(*LetStmt)
	arrayExpr, ok := letStmt1.Value.(*ArrayExpr)
	if !ok {
		t.Fatalf("expected ArrayExpr, got %T", letStmt1.Value)
	}

	if len(arrayExpr.Elements) != 3 {
		t.Errorf("expected 3 elements, got %d", len(arrayExpr.Elements))
	}

	// Check array indexing
	letStmt2 := fn.Body.Statements[1].(*LetStmt)
	indexExpr, ok := letStmt2.Value.(*IndexExpr)
	if !ok {
		t.Fatalf("expected IndexExpr, got %T", letStmt2.Value)
	}

	if indexExpr.Array == nil || indexExpr.Index == nil {
		t.Errorf("expected array and index")
	}
}

func TestParserMemberExpression(t *testing.T) {
	input := `function test(): void {
    let value = obj.field;
}`

	lexer := NewLexer(input, "test.qs")
	parser := NewParser(lexer)
	program := parser.ParseProgram()

	checkParserErrors(t, parser)

	fn := program.Declarations[0].(*FunctionDecl)
	letStmt := fn.Body.Statements[0].(*LetStmt)
	memberExpr, ok := letStmt.Value.(*MemberExpr)
	if !ok {
		t.Fatalf("expected MemberExpr, got %T", letStmt.Value)
	}

	if memberExpr.Member != "field" {
		t.Errorf("expected member 'field', got %s", memberExpr.Member)
	}
}

func TestParserImportDeclaration(t *testing.T) {
	input := `import { invoke } from "std/invoke";

export function entry(): void {
    return;
}`

	lexer := NewLexer(input, "test.qs")
	parser := NewParser(lexer)
	program := parser.ParseProgram()

	checkParserErrors(t, parser)

	if len(program.Imports) != 1 {
		t.Fatalf("expected 1 import, got %d", len(program.Imports))
	}

	imp := program.Imports[0]
	if len(imp.Names) != 1 || imp.Names[0] != "invoke" {
		t.Errorf("expected import name 'invoke'")
	}

	if imp.Path != "std/invoke" {
		t.Errorf("expected import path 'std/invoke', got %s", imp.Path)
	}
}

func TestParserAssignmentStatement(t *testing.T) {
	input := `function test(): void {
    x = 42;
    y += 10;
    z -= 5;
}`

	lexer := NewLexer(input, "test.qs")
	parser := NewParser(lexer)
	program := parser.ParseProgram()

	checkParserErrors(t, parser)

	fn := program.Declarations[0].(*FunctionDecl)

	// Check regular assignment
	assignStmt1, ok := fn.Body.Statements[0].(*AssignmentStmt)
	if !ok {
		t.Fatalf("expected AssignmentStmt, got %T", fn.Body.Statements[0])
	}

	if assignStmt1.Operator != TOKEN_ASSIGN {
		t.Errorf("expected = operator, got %s", assignStmt1.Operator)
	}

	// Check += assignment
	assignStmt2, ok := fn.Body.Statements[1].(*AssignmentStmt)
	if !ok {
		t.Fatalf("expected AssignmentStmt, got %T", fn.Body.Statements[1])
	}

	if assignStmt2.Operator != TOKEN_PLUS_ASSIGN {
		t.Errorf("expected += operator, got %s", assignStmt2.Operator)
	}
}

func checkParserErrors(t *testing.T, p *Parser) {
	errors := p.Errors()
	if len(errors) == 0 {
		return
	}

	t.Errorf("parser has %d errors", len(errors))
	for _, err := range errors {
		t.Errorf("parser error: %s", err)
	}
	t.FailNow()
}
