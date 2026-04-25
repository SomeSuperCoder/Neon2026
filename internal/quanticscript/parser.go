package quanticscript

import (
	"fmt"
)

// Parser parses QuanticScript source code into an AST
type Parser struct {
	lexer   *Lexer
	current Token
	peek    Token
	errors  []error
}

// NewParser creates a new parser
func NewParser(lexer *Lexer) *Parser {
	p := &Parser{
		lexer:  lexer,
		errors: []error{},
	}
	// Read two tokens to initialize current and peek
	p.nextToken()
	p.nextToken()
	return p
}

// nextToken advances to the next token
func (p *Parser) nextToken() {
	p.current = p.peek
	p.peek = p.lexer.NextToken()
}

// Errors returns the list of parsing errors
func (p *Parser) Errors() []error {
	return p.errors
}

// addError adds a parsing error
func (p *Parser) addError(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	err := &ParserError{
		Location: p.current.Location,
		Message:  msg,
	}
	p.errors = append(p.errors, err)
}

// expect checks if the current token matches the expected type
func (p *Parser) expect(t TokenType) bool {
	if p.current.Type == t {
		p.nextToken()
		return true
	}
	p.addError("expected %s, got %s", t, p.current.Type)
	return false
}

// ParseProgram parses the entire program
func (p *Parser) ParseProgram() *Program {
	program := &Program{
		Location:     p.current.Location,
		Imports:      []*ImportDecl{},
		Declarations: []Node{},
	}

	// Parse imports
	for p.current.Type == TOKEN_IMPORT {
		if imp := p.parseImport(); imp != nil {
			program.Imports = append(program.Imports, imp)
		}
	}

	// Parse declarations
	for p.current.Type != TOKEN_EOF {
		if decl := p.parseDeclaration(); decl != nil {
			program.Declarations = append(program.Declarations, decl)
		} else {
			// Skip to next statement on error
			p.nextToken()
		}
	}

	return program
}

// parseImport parses an import declaration
func (p *Parser) parseImport() *ImportDecl {
	loc := p.current.Location
	p.nextToken() // consume 'import'

	imp := &ImportDecl{
		Location: loc,
		Names:    []string{},
	}

	// Parse import names: import { name1, name2 } from "path"
	if p.current.Type == TOKEN_LBRACE {
		p.nextToken()
		for p.current.Type == TOKEN_IDENT {
			imp.Names = append(imp.Names, p.current.Literal)
			p.nextToken()
			if p.current.Type == TOKEN_COMMA {
				p.nextToken()
			}
		}
		p.expect(TOKEN_RBRACE)
	}

	p.expect(TOKEN_FROM)

	if p.current.Type == TOKEN_STRING {
		imp.Path = p.current.Literal
		p.nextToken()
	} else {
		p.addError("expected string literal for import path")
	}

	p.expect(TOKEN_SEMICOLON)

	return imp
}

// parseDeclaration parses a top-level declaration
func (p *Parser) parseDeclaration() Node {
	isExport := false
	if p.current.Type == TOKEN_EXPORT {
		isExport = true
		p.nextToken()
	}

	switch p.current.Type {
	case TOKEN_FUNCTION:
		return p.parseFunctionDecl(isExport)
	default:
		p.addError("unexpected token %s at top level", p.current.Type)
		return nil
	}
}

// parseFunctionDecl parses a function declaration
func (p *Parser) parseFunctionDecl(isExport bool) *FunctionDecl {
	loc := p.current.Location
	p.nextToken() // consume 'function'

	if p.current.Type != TOKEN_IDENT {
		p.addError("expected function name")
		return nil
	}

	fn := &FunctionDecl{
		Location:   loc,
		Name:       p.current.Literal,
		Parameters: []*Parameter{},
		IsExport:   isExport,
	}
	p.nextToken()

	// Parse parameters
	p.expect(TOKEN_LPAREN)
	if p.current.Type != TOKEN_RPAREN {
		for {
			param := p.parseParameter()
			if param != nil {
				fn.Parameters = append(fn.Parameters, param)
			}
			if p.current.Type != TOKEN_COMMA {
				break
			}
			p.nextToken()
		}
	}
	p.expect(TOKEN_RPAREN)

	// Parse return type
	if p.current.Type == TOKEN_COLON {
		p.nextToken()
		fn.ReturnType = p.parseTypeAnnotation()
	}

	// Parse body
	fn.Body = p.parseBlockStmt()

	return fn
}

// parseParameter parses a function parameter
func (p *Parser) parseParameter() *Parameter {
	loc := p.current.Location

	if p.current.Type != TOKEN_IDENT {
		p.addError("expected parameter name")
		return nil
	}

	param := &Parameter{
		Location: loc,
		Name:     p.current.Literal,
	}
	p.nextToken()

	// Parse type annotation
	if p.current.Type == TOKEN_COLON {
		p.nextToken()
		param.Type = p.parseTypeAnnotation()
	}

	return param
}

// parseTypeAnnotation parses a type annotation
func (p *Parser) parseTypeAnnotation() *TypeAnnotation {
	loc := p.current.Location

	typeAnn := &TypeAnnotation{
		Location: loc,
	}

	// Parse base type
	if p.current.Type == TOKEN_IDENT || isTypeToken(p.current.Type) {
		typeAnn.Name = p.current.Literal
		p.nextToken()
	} else {
		p.addError("expected type name")
		return nil
	}

	// Check for array type
	if p.current.Type == TOKEN_LBRACKET {
		p.nextToken()
		p.expect(TOKEN_RBRACKET)
		typeAnn.IsArray = true
	}

	// Check for nullable type
	if p.current.Type == TOKEN_PIPE {
		p.nextToken()
		if p.current.Type == TOKEN_NULL {
			p.nextToken()
			typeAnn.Nullable = true
		}
	}

	return typeAnn
}

// parseBlockStmt parses a block statement
func (p *Parser) parseBlockStmt() *BlockStmt {
	loc := p.current.Location
	p.expect(TOKEN_LBRACE)

	block := &BlockStmt{
		Location:   loc,
		Statements: []Statement{},
	}

	for p.current.Type != TOKEN_RBRACE && p.current.Type != TOKEN_EOF {
		// Save current token to detect if we're stuck
		prevToken := p.current
		if stmt := p.parseStatement(); stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}
		// If we're still on the same token after parsing, advance to prevent infinite loop
		if p.current == prevToken {
			p.nextToken()
		}
	}

	p.expect(TOKEN_RBRACE)
	return block
}

// parseStatement parses a statement
func (p *Parser) parseStatement() Statement {
	switch p.current.Type {
	case TOKEN_RETURN:
		return p.parseReturnStmt()
	case TOKEN_LET, TOKEN_CONST:
		return p.parseLetStmt()
	case TOKEN_IF:
		return p.parseIfStmt()
	case TOKEN_WHILE:
		return p.parseWhileStmt()
	case TOKEN_FOR:
		return p.parseForStmt()
	case TOKEN_ASM:
		return p.parseAsmBlock()
	case TOKEN_LBRACE:
		return p.parseBlockStmt()
	default:
		// Try to parse as assignment or expression statement
		return p.parseExprOrAssignmentStmt()
	}
}

// parseReturnStmt parses a return statement
func (p *Parser) parseReturnStmt() *ReturnStmt {
	loc := p.current.Location
	p.nextToken() // consume 'return'

	stmt := &ReturnStmt{
		Location: loc,
	}

	if p.current.Type != TOKEN_SEMICOLON {
		stmt.Value = p.parseExpression()
	}

	p.expect(TOKEN_SEMICOLON)
	return stmt
}

// parseLetStmt parses a variable declaration
func (p *Parser) parseLetStmt() *LetStmt {
	loc := p.current.Location
	isConst := p.current.Type == TOKEN_CONST
	p.nextToken() // consume 'let' or 'const'

	if p.current.Type != TOKEN_IDENT {
		p.addError("expected variable name")
		return nil
	}

	stmt := &LetStmt{
		Location: loc,
		Name:     p.current.Literal,
		IsConst:  isConst,
	}
	p.nextToken()

	// Parse type annotation
	if p.current.Type == TOKEN_COLON {
		p.nextToken()
		stmt.Type = p.parseTypeAnnotation()
	}

	// Parse initializer
	if p.current.Type == TOKEN_ASSIGN {
		p.nextToken()
		stmt.Value = p.parseExpression()
	}

	p.expect(TOKEN_SEMICOLON)
	return stmt
}

// parseIfStmt parses an if statement
func (p *Parser) parseIfStmt() *IfStmt {
	loc := p.current.Location
	p.nextToken() // consume 'if'

	p.expect(TOKEN_LPAREN)
	condition := p.parseExpression()
	p.expect(TOKEN_RPAREN)

	stmt := &IfStmt{
		Location:  loc,
		Condition: condition,
		ThenBlock: p.parseBlockStmt(),
	}

	if p.current.Type == TOKEN_ELSE {
		p.nextToken()
		stmt.ElseBlock = p.parseBlockStmt()
	}

	return stmt
}

// parseWhileStmt parses a while statement
func (p *Parser) parseWhileStmt() *WhileStmt {
	loc := p.current.Location
	p.nextToken() // consume 'while'

	p.expect(TOKEN_LPAREN)
	condition := p.parseExpression()
	p.expect(TOKEN_RPAREN)

	stmt := &WhileStmt{
		Location:  loc,
		Condition: condition,
		Body:      p.parseBlockStmt(),
	}

	return stmt
}

// parseForStmt parses a for statement
func (p *Parser) parseForStmt() *ForStmt {
	loc := p.current.Location
	p.nextToken() // consume 'for'

	p.expect(TOKEN_LPAREN)

	stmt := &ForStmt{
		Location: loc,
	}

	// Parse init
	if p.current.Type != TOKEN_SEMICOLON {
		stmt.Init = p.parseStatement()
	} else {
		p.nextToken()
	}

	// Parse condition
	if p.current.Type != TOKEN_SEMICOLON {
		stmt.Condition = p.parseExpression()
	}
	p.expect(TOKEN_SEMICOLON)

	// Parse update (no semicolon expected)
	if p.current.Type != TOKEN_RPAREN {
		// Parse as assignment without semicolon
		expr := p.parseExpression()
		if isAssignmentOp(p.current.Type) {
			op := p.current.Type
			p.nextToken()
			value := p.parseExpression()
			stmt.Update = &AssignmentStmt{
				Location: expr.Loc(),
				Target:   expr,
				Operator: op,
				Value:    value,
			}
		} else {
			stmt.Update = &ExprStmt{
				Location:   expr.Loc(),
				Expression: expr,
			}
		}
	}
	p.expect(TOKEN_RPAREN)

	stmt.Body = p.parseBlockStmt()

	return stmt
}

// parseAsmBlock parses an inline assembly block
func (p *Parser) parseAsmBlock() *AsmBlock {
	loc := p.current.Location
	p.nextToken() // consume '__asm__'

	p.expect(TOKEN_LBRACE)

	block := &AsmBlock{
		Location:     loc,
		Instructions: []AsmInstruction{},
	}

	for p.current.Type != TOKEN_RBRACE && p.current.Type != TOKEN_EOF {
		if p.current.Type == TOKEN_IDENT {
			instr := AsmInstruction{
				Location: p.current.Location,
				Mnemonic: p.current.Literal,
				Operands: []string{},
			}
			p.nextToken()

			// Parse operands (identifiers, numbers, etc.)
			for p.current.Type != TOKEN_RBRACE && p.current.Type != TOKEN_EOF {
				// Skip commas
				if p.current.Type == TOKEN_COMMA {
					p.nextToken()
					continue
				}

				// If we hit an uppercase identifier, it's likely a new instruction
				if p.current.Type == TOKEN_IDENT {
					// Check if this looks like an instruction mnemonic (uppercase)
					if len(p.current.Literal) > 0 && p.current.Literal[0] >= 'A' && p.current.Literal[0] <= 'Z' {
						// This is a new instruction, stop parsing operands
						break
					}
					// Otherwise, it's an operand (lowercase identifier)
					instr.Operands = append(instr.Operands, p.current.Literal)
					p.nextToken()
				} else if p.current.Type == TOKEN_INT {
					// Numeric operand
					instr.Operands = append(instr.Operands, p.current.Literal)
					p.nextToken()
				} else {
					// Skip other tokens (whitespace, etc.)
					p.nextToken()
				}
			}

			block.Instructions = append(block.Instructions, instr)
		} else {
			p.nextToken()
		}
	}

	p.expect(TOKEN_RBRACE)
	return block
}

// parseExprOrAssignmentStmt parses an expression or assignment statement
func (p *Parser) parseExprOrAssignmentStmt() Statement {
	loc := p.current.Location
	expr := p.parseExpression()

	// Check if it's an assignment
	if isAssignmentOp(p.current.Type) {
		op := p.current.Type
		p.nextToken()
		value := p.parseExpression()
		p.expect(TOKEN_SEMICOLON)
		return &AssignmentStmt{
			Location: loc,
			Target:   expr,
			Operator: op,
			Value:    value,
		}
	}

	p.expect(TOKEN_SEMICOLON)
	return &ExprStmt{
		Location:   loc,
		Expression: expr,
	}
}

// parseExpression parses an expression
func (p *Parser) parseExpression() Expression {
	return p.parseLogicalOr()
}

// parseLogicalOr parses logical OR expressions
func (p *Parser) parseLogicalOr() Expression {
	left := p.parseLogicalAnd()

	for p.current.Type == TOKEN_OR {
		op := p.current.Type
		loc := p.current.Location
		p.nextToken()
		right := p.parseLogicalAnd()
		left = &BinaryExpr{
			Location: loc,
			Left:     left,
			Operator: op,
			Right:    right,
		}
	}

	return left
}

// parseLogicalAnd parses logical AND expressions
func (p *Parser) parseLogicalAnd() Expression {
	left := p.parseBitwiseOr()

	for p.current.Type == TOKEN_AND {
		op := p.current.Type
		loc := p.current.Location
		p.nextToken()
		right := p.parseBitwiseOr()
		left = &BinaryExpr{
			Location: loc,
			Left:     left,
			Operator: op,
			Right:    right,
		}
	}

	return left
}

// parseBitwiseOr parses bitwise OR expressions
func (p *Parser) parseBitwiseOr() Expression {
	left := p.parseBitwiseXor()

	for p.current.Type == TOKEN_PIPE {
		op := p.current.Type
		loc := p.current.Location
		p.nextToken()
		right := p.parseBitwiseXor()
		left = &BinaryExpr{
			Location: loc,
			Left:     left,
			Operator: op,
			Right:    right,
		}
	}

	return left
}

// parseBitwiseXor parses bitwise XOR expressions
func (p *Parser) parseBitwiseXor() Expression {
	left := p.parseBitwiseAnd()

	for p.current.Type == TOKEN_CARET {
		op := p.current.Type
		loc := p.current.Location
		p.nextToken()
		right := p.parseBitwiseAnd()
		left = &BinaryExpr{
			Location: loc,
			Left:     left,
			Operator: op,
			Right:    right,
		}
	}

	return left
}

// parseBitwiseAnd parses bitwise AND expressions
func (p *Parser) parseBitwiseAnd() Expression {
	left := p.parseEquality()

	for p.current.Type == TOKEN_AMPERSAND {
		op := p.current.Type
		loc := p.current.Location
		p.nextToken()
		right := p.parseEquality()
		left = &BinaryExpr{
			Location: loc,
			Left:     left,
			Operator: op,
			Right:    right,
		}
	}

	return left
}

// parseEquality parses equality expressions
func (p *Parser) parseEquality() Expression {
	left := p.parseComparison()

	for p.current.Type == TOKEN_EQ || p.current.Type == TOKEN_NEQ {
		op := p.current.Type
		loc := p.current.Location
		p.nextToken()
		right := p.parseComparison()
		left = &BinaryExpr{
			Location: loc,
			Left:     left,
			Operator: op,
			Right:    right,
		}
	}

	return left
}

// parseComparison parses comparison expressions
func (p *Parser) parseComparison() Expression {
	left := p.parseShift()

	for p.current.Type == TOKEN_LT || p.current.Type == TOKEN_GT ||
		p.current.Type == TOKEN_LTE || p.current.Type == TOKEN_GTE {
		op := p.current.Type
		loc := p.current.Location
		p.nextToken()
		right := p.parseShift()
		left = &BinaryExpr{
			Location: loc,
			Left:     left,
			Operator: op,
			Right:    right,
		}
	}

	return left
}

// parseShift parses shift expressions
func (p *Parser) parseShift() Expression {
	left := p.parseAdditive()

	for p.current.Type == TOKEN_LSHIFT || p.current.Type == TOKEN_RSHIFT {
		op := p.current.Type
		loc := p.current.Location
		p.nextToken()
		right := p.parseAdditive()
		left = &BinaryExpr{
			Location: loc,
			Left:     left,
			Operator: op,
			Right:    right,
		}
	}

	return left
}

// parseAdditive parses addition and subtraction
func (p *Parser) parseAdditive() Expression {
	left := p.parseMultiplicative()

	for p.current.Type == TOKEN_PLUS || p.current.Type == TOKEN_MINUS {
		op := p.current.Type
		loc := p.current.Location
		p.nextToken()
		right := p.parseMultiplicative()
		left = &BinaryExpr{
			Location: loc,
			Left:     left,
			Operator: op,
			Right:    right,
		}
	}

	return left
}

// parseMultiplicative parses multiplication, division, and modulo
func (p *Parser) parseMultiplicative() Expression {
	left := p.parseUnary()

	for p.current.Type == TOKEN_STAR || p.current.Type == TOKEN_SLASH ||
		p.current.Type == TOKEN_PERCENT {
		op := p.current.Type
		loc := p.current.Location
		p.nextToken()
		right := p.parseUnary()
		left = &BinaryExpr{
			Location: loc,
			Left:     left,
			Operator: op,
			Right:    right,
		}
	}

	return left
}

// parseUnary parses unary expressions
func (p *Parser) parseUnary() Expression {
	if p.current.Type == TOKEN_NOT || p.current.Type == TOKEN_MINUS ||
		p.current.Type == TOKEN_TILDE {
		loc := p.current.Location
		op := p.current.Type
		p.nextToken()
		operand := p.parseUnary()
		return &UnaryExpr{
			Location: loc,
			Operator: op,
			Operand:  operand,
		}
	}

	return p.parsePostfix()
}

// parsePostfix parses postfix expressions (calls, indexing, member access)
func (p *Parser) parsePostfix() Expression {
	expr := p.parsePrimary()

	for {
		switch p.current.Type {
		case TOKEN_LPAREN:
			expr = p.parseCallExpr(expr)
		case TOKEN_LBRACKET:
			expr = p.parseIndexExpr(expr)
		case TOKEN_DOT:
			expr = p.parseMemberExpr(expr)
		default:
			return expr
		}
	}
}

// parseCallExpr parses a function call
func (p *Parser) parseCallExpr(function Expression) *CallExpr {
	loc := p.current.Location
	p.nextToken() // consume '('

	call := &CallExpr{
		Location:  loc,
		Function:  function,
		Arguments: []Expression{},
	}

	if p.current.Type != TOKEN_RPAREN {
		for {
			arg := p.parseExpression()
			call.Arguments = append(call.Arguments, arg)
			if p.current.Type != TOKEN_COMMA {
				break
			}
			p.nextToken()
		}
	}

	p.expect(TOKEN_RPAREN)
	return call
}

// parseIndexExpr parses array indexing
func (p *Parser) parseIndexExpr(array Expression) *IndexExpr {
	loc := p.current.Location
	p.nextToken() // consume '['

	index := p.parseExpression()
	p.expect(TOKEN_RBRACKET)

	return &IndexExpr{
		Location: loc,
		Array:    array,
		Index:    index,
	}
}

// parseMemberExpr parses member access
func (p *Parser) parseMemberExpr(object Expression) *MemberExpr {
	loc := p.current.Location
	p.nextToken() // consume '.'

	if p.current.Type != TOKEN_IDENT {
		p.addError("expected member name")
		return nil
	}

	member := p.current.Literal
	p.nextToken()

	return &MemberExpr{
		Location: loc,
		Object:   object,
		Member:   member,
	}
}

// parsePrimary parses primary expressions
func (p *Parser) parsePrimary() Expression {
	switch p.current.Type {
	case TOKEN_IDENT:
		return p.parseIdentExpr()
	case TOKEN_INT:
		return p.parseIntLiteral()
	case TOKEN_STRING:
		return p.parseStringLiteral()
	case TOKEN_TRUE, TOKEN_FALSE:
		return p.parseBoolLiteral()
	case TOKEN_NULL:
		return p.parseNullLiteral()
	case TOKEN_LBRACKET:
		return p.parseArrayExpr()
	case TOKEN_LPAREN:
		return p.parseGroupedExpr()
	default:
		p.addError("unexpected token %s in expression", p.current.Type)
		return nil
	}
}

// parseIdentExpr parses an identifier
func (p *Parser) parseIdentExpr() *IdentExpr {
	expr := &IdentExpr{
		Location: p.current.Location,
		Name:     p.current.Literal,
	}
	p.nextToken()
	return expr
}

// parseIntLiteral parses an integer literal
func (p *Parser) parseIntLiteral() *IntLiteral {
	expr := &IntLiteral{
		Location: p.current.Location,
		Value:    p.current.Literal,
	}
	p.nextToken()
	return expr
}

// parseStringLiteral parses a string literal
func (p *Parser) parseStringLiteral() *StringLiteral {
	expr := &StringLiteral{
		Location: p.current.Location,
		Value:    p.current.Literal,
	}
	p.nextToken()
	return expr
}

// parseBoolLiteral parses a boolean literal
func (p *Parser) parseBoolLiteral() *BoolLiteral {
	expr := &BoolLiteral{
		Location: p.current.Location,
		Value:    p.current.Type == TOKEN_TRUE,
	}
	p.nextToken()
	return expr
}

// parseNullLiteral parses a null literal
func (p *Parser) parseNullLiteral() *NullLiteral {
	expr := &NullLiteral{
		Location: p.current.Location,
	}
	p.nextToken()
	return expr
}

// parseArrayExpr parses an array literal
func (p *Parser) parseArrayExpr() *ArrayExpr {
	loc := p.current.Location
	p.nextToken() // consume '['

	expr := &ArrayExpr{
		Location: loc,
		Elements: []Expression{},
	}

	if p.current.Type != TOKEN_RBRACKET {
		for {
			elem := p.parseExpression()
			expr.Elements = append(expr.Elements, elem)
			if p.current.Type != TOKEN_COMMA {
				break
			}
			p.nextToken()
		}
	}

	p.expect(TOKEN_RBRACKET)
	return expr
}

// parseGroupedExpr parses a grouped expression
func (p *Parser) parseGroupedExpr() Expression {
	p.nextToken() // consume '('
	expr := p.parseExpression()
	p.expect(TOKEN_RPAREN)
	return expr
}

// Helper functions

func isTypeToken(t TokenType) bool {
	return t == TOKEN_I8 || t == TOKEN_I16 || t == TOKEN_I32 || t == TOKEN_I64 ||
		t == TOKEN_U8 || t == TOKEN_U16 || t == TOKEN_U32 || t == TOKEN_U64 ||
		t == TOKEN_BOOL || t == TOKEN_STRING_TYPE || t == TOKEN_BYTES_TYPE ||
		t == TOKEN_VOID
}

func isAssignmentOp(t TokenType) bool {
	return t == TOKEN_ASSIGN || t == TOKEN_PLUS_ASSIGN || t == TOKEN_MINUS_ASSIGN
}

// ParserError represents a parsing error
type ParserError struct {
	Location SourceLocation
	Message  string
}

func (e *ParserError) Error() string {
	return fmt.Sprintf("%s: %s", e.Location, e.Message)
}
