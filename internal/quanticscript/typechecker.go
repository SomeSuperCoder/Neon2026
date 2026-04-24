package quanticscript

import (
	"fmt"
)

// TypeChecker performs type checking and inference on the AST
type TypeChecker struct {
	errors      []error
	scopes      []*Scope
	currentFunc *FunctionDecl
}

// Scope represents a lexical scope for variables and functions
type Scope struct {
	parent    *Scope
	variables map[string]*TypeInfo
	functions map[string]*FunctionType
}

// TypeInfo represents type information for a variable or expression
type TypeInfo struct {
	Type     *TypeAnnotation
	IsConst  bool
	Location SourceLocation
}

// FunctionType represents a function signature
type FunctionType struct {
	Parameters []*TypeAnnotation
	ReturnType *TypeAnnotation
	Location   SourceLocation
}

// NewTypeChecker creates a new type checker
func NewTypeChecker() *TypeChecker {
	tc := &TypeChecker{
		errors: []error{},
		scopes: []*Scope{},
	}
	tc.pushScope() // Global scope
	tc.registerBuiltinFunctions()
	return tc
}

// registerBuiltinFunctions registers built-in functions and marks non-deterministic ones
func (tc *TypeChecker) registerBuiltinFunctions() {
	// These are placeholder registrations for standard library functions
	// In a full implementation, these would be loaded from the standard library

	// Deterministic functions are allowed
	tc.registerFunction("sha256", []*TypeAnnotation{{Name: "bytes"}}, &TypeAnnotation{Name: "bytes"})
	tc.registerFunction("verifySignature", []*TypeAnnotation{{Name: "PublicKey"}, {Name: "bytes"}, {Name: "bytes"}}, &TypeAnnotation{Name: "bool"})

	// Non-deterministic functions are registered but will be rejected
	tc.registerNonDeterministicFunction("random")
	tc.registerNonDeterministicFunction("rand")
	tc.registerNonDeterministicFunction("getRandom")
	tc.registerNonDeterministicFunction("randomBytes")
	tc.registerNonDeterministicFunction("now")
	tc.registerNonDeterministicFunction("currentTime")
	tc.registerNonDeterministicFunction("getTime")
	tc.registerNonDeterministicFunction("timestamp")
	tc.registerNonDeterministicFunction("readFile")
	tc.registerNonDeterministicFunction("writeFile")
	tc.registerNonDeterministicFunction("openFile")
	tc.registerNonDeterministicFunction("fetch")
	tc.registerNonDeterministicFunction("httpGet")
	tc.registerNonDeterministicFunction("httpPost")
	tc.registerNonDeterministicFunction("connect")
}

// registerFunction registers a deterministic function
func (tc *TypeChecker) registerFunction(name string, params []*TypeAnnotation, returnType *TypeAnnotation) {
	tc.declareFunction(name, &FunctionType{
		Parameters: params,
		ReturnType: returnType,
		Location:   SourceLocation{},
	})
}

// registerNonDeterministicFunction registers a non-deterministic function (for error detection)
func (tc *TypeChecker) registerNonDeterministicFunction(name string) {
	// Mark as special non-deterministic function
	tc.declareFunction(name, &FunctionType{
		Parameters: []*TypeAnnotation{}, // Placeholder
		ReturnType: nil,
		Location:   SourceLocation{},
	})
}

// Errors returns the list of type checking errors
func (tc *TypeChecker) Errors() []error {
	return tc.errors
}

// addError adds a type checking error
func (tc *TypeChecker) addError(loc SourceLocation, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	err := &TypeCheckerError{
		Location: loc,
		Message:  msg,
	}
	tc.errors = append(tc.errors, err)
}

// pushScope creates a new scope
func (tc *TypeChecker) pushScope() {
	scope := &Scope{
		variables: make(map[string]*TypeInfo),
		functions: make(map[string]*FunctionType),
	}
	if len(tc.scopes) > 0 {
		scope.parent = tc.scopes[len(tc.scopes)-1]
	}
	tc.scopes = append(tc.scopes, scope)
}

// popScope removes the current scope
func (tc *TypeChecker) popScope() {
	if len(tc.scopes) > 0 {
		tc.scopes = tc.scopes[:len(tc.scopes)-1]
	}
}

// currentScope returns the current scope
func (tc *TypeChecker) currentScope() *Scope {
	if len(tc.scopes) == 0 {
		return nil
	}
	return tc.scopes[len(tc.scopes)-1]
}

// declareVariable declares a variable in the current scope
func (tc *TypeChecker) declareVariable(name string, typeInfo *TypeInfo) {
	scope := tc.currentScope()
	if scope != nil {
		if _, exists := scope.variables[name]; exists {
			tc.addError(typeInfo.Location, "variable '%s' already declared in this scope", name)
			return
		}
		scope.variables[name] = typeInfo
	}
}

// lookupVariable looks up a variable in the scope chain
func (tc *TypeChecker) lookupVariable(name string) *TypeInfo {
	for i := len(tc.scopes) - 1; i >= 0; i-- {
		if typeInfo, exists := tc.scopes[i].variables[name]; exists {
			return typeInfo
		}
	}
	return nil
}

// declareFunction declares a function in the current scope
func (tc *TypeChecker) declareFunction(name string, funcType *FunctionType) {
	scope := tc.currentScope()
	if scope != nil {
		if _, exists := scope.functions[name]; exists {
			tc.addError(funcType.Location, "function '%s' already declared in this scope", name)
			return
		}
		scope.functions[name] = funcType
	}
}

// lookupFunction looks up a function in the scope chain
func (tc *TypeChecker) lookupFunction(name string) *FunctionType {
	for i := len(tc.scopes) - 1; i >= 0; i-- {
		if funcType, exists := tc.scopes[i].functions[name]; exists {
			return funcType
		}
	}
	return nil
}

// CheckProgram type checks the entire program
func (tc *TypeChecker) CheckProgram(program *Program) {
	// First pass: declare all functions
	for _, decl := range program.Declarations {
		if fn, ok := decl.(*FunctionDecl); ok {
			tc.declareFunctionSignature(fn)
		}
	}

	// Second pass: check function bodies
	for _, decl := range program.Declarations {
		if fn, ok := decl.(*FunctionDecl); ok {
			tc.checkFunctionDecl(fn)
		}
	}
}

// declareFunctionSignature declares a function signature
func (tc *TypeChecker) declareFunctionSignature(fn *FunctionDecl) {
	paramTypes := make([]*TypeAnnotation, len(fn.Parameters))
	for i, param := range fn.Parameters {
		if param.Type == nil {
			tc.addError(param.Location, "parameter '%s' must have a type annotation", param.Name)
			return
		}
		paramTypes[i] = param.Type
	}

	funcType := &FunctionType{
		Parameters: paramTypes,
		ReturnType: fn.ReturnType,
		Location:   fn.Location,
	}

	tc.declareFunction(fn.Name, funcType)
}

// checkFunctionDecl checks a function declaration
func (tc *TypeChecker) checkFunctionDecl(fn *FunctionDecl) {
	tc.currentFunc = fn
	tc.pushScope()

	// Declare parameters in function scope
	for _, param := range fn.Parameters {
		if param.Type == nil {
			tc.addError(param.Location, "parameter '%s' must have a type annotation", param.Name)
			continue
		}
		tc.declareVariable(param.Name, &TypeInfo{
			Type:     param.Type,
			IsConst:  false,
			Location: param.Location,
		})
	}

	// Check function body
	tc.checkBlockStmt(fn.Body)

	tc.popScope()
	tc.currentFunc = nil
}

// checkBlockStmt checks a block statement
func (tc *TypeChecker) checkBlockStmt(block *BlockStmt) {
	tc.pushScope()
	for _, stmt := range block.Statements {
		tc.checkStatement(stmt)
	}
	tc.popScope()
}

// checkStatement checks a statement
func (tc *TypeChecker) checkStatement(stmt Statement) {
	switch s := stmt.(type) {
	case *ReturnStmt:
		tc.checkReturnStmt(s)
	case *LetStmt:
		tc.checkLetStmt(s)
	case *ExprStmt:
		tc.checkExpression(s.Expression)
	case *IfStmt:
		tc.checkIfStmt(s)
	case *WhileStmt:
		tc.checkWhileStmt(s)
	case *ForStmt:
		tc.checkForStmt(s)
	case *AssignmentStmt:
		tc.checkAssignmentStmt(s)
	case *AsmBlock:
		tc.checkAsmBlock(s)
	case *BlockStmt:
		tc.checkBlockStmt(s)
	}
}

// checkReturnStmt checks a return statement
func (tc *TypeChecker) checkReturnStmt(stmt *ReturnStmt) {
	if tc.currentFunc == nil {
		tc.addError(stmt.Location, "return statement outside of function")
		return
	}

	var returnType *TypeAnnotation
	if stmt.Value != nil {
		returnType = tc.checkExpression(stmt.Value)
	}

	expectedType := tc.currentFunc.ReturnType
	if expectedType == nil && returnType != nil {
		tc.addError(stmt.Location, "function does not return a value")
	} else if expectedType != nil && returnType == nil {
		tc.addError(stmt.Location, "function must return a value of type %s", expectedType.Name)
	} else if expectedType != nil && returnType != nil {
		if !tc.typesCompatible(returnType, expectedType) {
			tc.addError(stmt.Location, "return type mismatch: expected %s, got %s",
				tc.typeToString(expectedType), tc.typeToString(returnType))
		}
	}
}

// checkLetStmt checks a variable declaration
func (tc *TypeChecker) checkLetStmt(stmt *LetStmt) {
	var inferredType *TypeAnnotation

	if stmt.Value != nil {
		inferredType = tc.checkExpression(stmt.Value)
	}

	var finalType *TypeAnnotation
	if stmt.Type != nil {
		finalType = stmt.Type

		// Reject floating-point types
		if tc.isFloatingPointType(finalType) {
			tc.addError(stmt.Location, "floating-point type '%s' is non-deterministic and not allowed", finalType.Name)
			return
		}

		// Check if initializer type matches declared type
		if inferredType != nil && !tc.typesCompatible(inferredType, finalType) {
			tc.addError(stmt.Location, "type mismatch: cannot assign %s to %s",
				tc.typeToString(inferredType), tc.typeToString(finalType))
		}
	} else if inferredType != nil {
		finalType = inferredType

		// Reject floating-point types
		if tc.isFloatingPointType(finalType) {
			tc.addError(stmt.Location, "floating-point type '%s' is non-deterministic and not allowed", finalType.Name)
			return
		}
	} else {
		tc.addError(stmt.Location, "variable '%s' must have a type annotation or initializer", stmt.Name)
		return
	}

	tc.declareVariable(stmt.Name, &TypeInfo{
		Type:     finalType,
		IsConst:  stmt.IsConst,
		Location: stmt.Location,
	})
}

// checkIfStmt checks an if statement
func (tc *TypeChecker) checkIfStmt(stmt *IfStmt) {
	condType := tc.checkExpression(stmt.Condition)
	if condType != nil && condType.Name != "bool" {
		tc.addError(stmt.Condition.Loc(), "if condition must be a boolean, got %s", tc.typeToString(condType))
	}

	tc.checkBlockStmt(stmt.ThenBlock)
	if stmt.ElseBlock != nil {
		tc.checkBlockStmt(stmt.ElseBlock)
	}
}

// checkWhileStmt checks a while statement
func (tc *TypeChecker) checkWhileStmt(stmt *WhileStmt) {
	condType := tc.checkExpression(stmt.Condition)
	if condType != nil && condType.Name != "bool" {
		tc.addError(stmt.Condition.Loc(), "while condition must be a boolean, got %s", tc.typeToString(condType))
	}

	tc.checkBlockStmt(stmt.Body)
}

// checkForStmt checks a for statement
func (tc *TypeChecker) checkForStmt(stmt *ForStmt) {
	tc.pushScope()

	if stmt.Init != nil {
		tc.checkStatement(stmt.Init)
	}

	if stmt.Condition != nil {
		condType := tc.checkExpression(stmt.Condition)
		if condType != nil && condType.Name != "bool" {
			tc.addError(stmt.Condition.Loc(), "for condition must be a boolean, got %s", tc.typeToString(condType))
		}
	}

	if stmt.Update != nil {
		tc.checkStatement(stmt.Update)
	}

	tc.checkBlockStmt(stmt.Body)

	tc.popScope()
}

// checkAssignmentStmt checks an assignment statement
func (tc *TypeChecker) checkAssignmentStmt(stmt *AssignmentStmt) {
	targetType := tc.checkExpression(stmt.Target)
	valueType := tc.checkExpression(stmt.Value)

	if targetType == nil || valueType == nil {
		return
	}

	// Check if target is assignable
	if ident, ok := stmt.Target.(*IdentExpr); ok {
		varInfo := tc.lookupVariable(ident.Name)
		if varInfo != nil && varInfo.IsConst {
			tc.addError(stmt.Location, "cannot assign to const variable '%s'", ident.Name)
			return
		}
	}

	// Check type compatibility
	if !tc.typesCompatible(valueType, targetType) {
		tc.addError(stmt.Location, "type mismatch: cannot assign %s to %s",
			tc.typeToString(valueType), tc.typeToString(targetType))
	}
}

// checkAsmBlock checks an inline assembly block
func (tc *TypeChecker) checkAsmBlock(block *AsmBlock) {
	// Validate that all variable references in assembly are properly typed
	for _, instr := range block.Instructions {
		tc.checkAsmInstruction(&instr)
	}
}

// checkAsmInstruction checks a single assembly instruction
func (tc *TypeChecker) checkAsmInstruction(instr *AsmInstruction) {
	// Validate operands based on instruction type
	for _, operand := range instr.Operands {
		// Check if operand is a variable reference
		if isIdentifier(operand) {
			varInfo := tc.lookupVariable(operand)
			if varInfo == nil {
				tc.addError(instr.Location, "undefined variable '%s' in assembly block", operand)
				continue
			}

			// Validate type safety at assembly boundaries
			tc.validateAsmOperandType(instr, operand, varInfo)
		}
	}

	// Check for unsafe operations in assembly
	tc.checkAsmInstructionSafety(instr)
}

// validateAsmOperandType validates that a variable used in assembly has a compatible type
func (tc *TypeChecker) validateAsmOperandType(instr *AsmInstruction, operand string, varInfo *TypeInfo) {
	// Instructions that load/store variables must use compatible types
	switch instr.Mnemonic {
	case "LOAD", "STORE":
		// These instructions work with any type, but we ensure the variable is properly typed
		if varInfo.Type == nil {
			tc.addError(instr.Location, "variable '%s' used in assembly must have a type", operand)
			return
		}

		// Prevent using arrays or complex types directly in assembly without explicit handling
		if varInfo.Type.IsArray {
			tc.addError(instr.Location, "cannot use array variable '%s' directly in assembly; use element access", operand)
		}

	case "GETFILE", "GETFILEMUT", "UPDATEFILE":
		// File operations require FileID type
		if varInfo.Type != nil && varInfo.Type.Name != "FileID" {
			tc.addError(instr.Location, "file operation requires FileID type, got %s for variable '%s'",
				tc.typeToString(varInfo.Type), operand)
		}

	case "GETBALANCE", "UPDATEBALANCE":
		// Balance operations work with FileID and numeric types
		if varInfo.Type != nil && varInfo.Type.Name != "FileID" && !tc.isNumericType(varInfo.Type) {
			tc.addError(instr.Location, "balance operation requires FileID or numeric type, got %s for variable '%s'",
				tc.typeToString(varInfo.Type), operand)
		}

	case "GETSIGNER", "HASSIGNER":
		// Signer operations require PublicKey type
		if varInfo.Type != nil && varInfo.Type.Name != "PublicKey" && !tc.isIntegerType(varInfo.Type) {
			tc.addError(instr.Location, "signer operation requires PublicKey or integer type, got %s for variable '%s'",
				tc.typeToString(varInfo.Type), operand)
		}

	case "ADD", "SUB", "MUL", "DIV", "MOD":
		// Arithmetic operations require numeric types
		if varInfo.Type != nil && !tc.isNumericType(varInfo.Type) {
			tc.addError(instr.Location, "arithmetic operation '%s' requires numeric type, got %s for variable '%s'",
				instr.Mnemonic, tc.typeToString(varInfo.Type), operand)
		}

	case "AND", "OR", "NOT":
		// Logical operations require boolean types
		if varInfo.Type != nil && varInfo.Type.Name != "bool" {
			tc.addError(instr.Location, "logical operation requires boolean type, got %s for variable '%s'",
				tc.typeToString(varInfo.Type), operand)
		}

	case "BAND", "BOR", "BXOR", "BNOT", "SHL", "SHR":
		// Bitwise operations require integer types
		if varInfo.Type != nil && !tc.isIntegerType(varInfo.Type) {
			tc.addError(instr.Location, "bitwise operation requires integer type, got %s for variable '%s'",
				tc.typeToString(varInfo.Type), operand)
		}
	}
}

// checkAsmInstructionSafety checks for unsafe operations in assembly
func (tc *TypeChecker) checkAsmInstructionSafety(instr *AsmInstruction) {
	// Prevent unsafe type coercions
	switch instr.Mnemonic {
	case "CAST", "REINTERPRET", "TRANSMUTE":
		// These are unsafe operations that could break type safety
		tc.addError(instr.Location, "unsafe type coercion '%s' not allowed in assembly", instr.Mnemonic)

	case "MEMCPY", "MEMMOVE", "MEMSET":
		// Direct memory operations bypass type safety
		tc.addError(instr.Location, "direct memory operation '%s' not allowed in assembly", instr.Mnemonic)

	case "SYSCALL", "CALL_EXTERNAL":
		// External calls could break sandbox
		tc.addError(instr.Location, "external call '%s' not allowed in assembly", instr.Mnemonic)

	// Detect non-deterministic operations in assembly
	case "RANDOM", "RAND", "GETRANDOM":
		tc.addError(instr.Location, "random number generation '%s' is non-deterministic and not allowed", instr.Mnemonic)

	case "TIME", "CLOCK", "TIMESTAMP", "NOW":
		tc.addError(instr.Location, "system time access '%s' is non-deterministic and not allowed", instr.Mnemonic)

	case "FILEREAD", "FILEWRITE", "FILEOPEN", "FILECLOSE":
		tc.addError(instr.Location, "file I/O operation '%s' is non-deterministic and not allowed", instr.Mnemonic)

	case "NETCONNECT", "NETSEND", "NETRECV", "HTTP", "FETCH":
		tc.addError(instr.Location, "network operation '%s' is non-deterministic and not allowed", instr.Mnemonic)

	case "FADD", "FSUB", "FMUL", "FDIV", "FMOD", "FSQRT", "FSIN", "FCOS", "FTAN":
		tc.addError(instr.Location, "floating-point operation '%s' is non-deterministic and not allowed", instr.Mnemonic)

	case "F32", "F64", "FLOAT", "DOUBLE":
		tc.addError(instr.Location, "floating-point type '%s' is non-deterministic and not allowed", instr.Mnemonic)
	}
}

// checkExpression checks an expression and returns its type
func (tc *TypeChecker) checkExpression(expr Expression) *TypeAnnotation {
	if expr == nil {
		return nil
	}

	switch e := expr.(type) {
	case *BinaryExpr:
		return tc.checkBinaryExpr(e)
	case *UnaryExpr:
		return tc.checkUnaryExpr(e)
	case *CallExpr:
		return tc.checkCallExpr(e)
	case *IdentExpr:
		return tc.checkIdentExpr(e)
	case *IntLiteral:
		return tc.checkIntLiteral(e)
	case *StringLiteral:
		return tc.checkStringLiteral(e)
	case *BoolLiteral:
		return tc.checkBoolLiteral(e)
	case *NullLiteral:
		return tc.checkNullLiteral(e)
	case *ArrayExpr:
		return tc.checkArrayExpr(e)
	case *IndexExpr:
		return tc.checkIndexExpr(e)
	case *MemberExpr:
		return tc.checkMemberExpr(e)
	default:
		tc.addError(expr.Loc(), "unknown expression type")
		return nil
	}
}

// checkBinaryExpr checks a binary expression
func (tc *TypeChecker) checkBinaryExpr(expr *BinaryExpr) *TypeAnnotation {
	leftType := tc.checkExpression(expr.Left)
	rightType := tc.checkExpression(expr.Right)

	if leftType == nil || rightType == nil {
		return nil
	}

	// Check operator compatibility
	switch expr.Operator {
	case TOKEN_PLUS, TOKEN_MINUS, TOKEN_STAR, TOKEN_SLASH, TOKEN_PERCENT:
		// Arithmetic operators require numeric types
		if !tc.isNumericType(leftType) || !tc.isNumericType(rightType) {
			tc.addError(expr.Location, "arithmetic operator requires numeric types, got %s and %s",
				tc.typeToString(leftType), tc.typeToString(rightType))
			return nil
		}
		if !tc.typesCompatible(leftType, rightType) {
			tc.addError(expr.Location, "type mismatch in arithmetic: %s and %s",
				tc.typeToString(leftType), tc.typeToString(rightType))
			return nil
		}
		return leftType

	case TOKEN_EQ, TOKEN_NEQ:
		// Equality operators work on any compatible types
		if !tc.typesCompatible(leftType, rightType) {
			tc.addError(expr.Location, "type mismatch in comparison: %s and %s",
				tc.typeToString(leftType), tc.typeToString(rightType))
		}
		return &TypeAnnotation{Name: "bool", Location: expr.Location}

	case TOKEN_LT, TOKEN_GT, TOKEN_LTE, TOKEN_GTE:
		// Comparison operators require numeric types
		if !tc.isNumericType(leftType) || !tc.isNumericType(rightType) {
			tc.addError(expr.Location, "comparison operator requires numeric types, got %s and %s",
				tc.typeToString(leftType), tc.typeToString(rightType))
			return nil
		}
		if !tc.typesCompatible(leftType, rightType) {
			tc.addError(expr.Location, "type mismatch in comparison: %s and %s",
				tc.typeToString(leftType), tc.typeToString(rightType))
		}
		return &TypeAnnotation{Name: "bool", Location: expr.Location}

	case TOKEN_AND, TOKEN_OR:
		// Logical operators require boolean types
		if leftType.Name != "bool" || rightType.Name != "bool" {
			tc.addError(expr.Location, "logical operator requires boolean types, got %s and %s",
				tc.typeToString(leftType), tc.typeToString(rightType))
			return nil
		}
		return &TypeAnnotation{Name: "bool", Location: expr.Location}

	case TOKEN_AMPERSAND, TOKEN_PIPE, TOKEN_CARET, TOKEN_LSHIFT, TOKEN_RSHIFT:
		// Bitwise operators require integer types
		if !tc.isIntegerType(leftType) || !tc.isIntegerType(rightType) {
			tc.addError(expr.Location, "bitwise operator requires integer types, got %s and %s",
				tc.typeToString(leftType), tc.typeToString(rightType))
			return nil
		}
		if !tc.typesCompatible(leftType, rightType) {
			tc.addError(expr.Location, "type mismatch in bitwise operation: %s and %s",
				tc.typeToString(leftType), tc.typeToString(rightType))
			return nil
		}
		return leftType

	default:
		tc.addError(expr.Location, "unknown binary operator")
		return nil
	}
}

// checkUnaryExpr checks a unary expression
func (tc *TypeChecker) checkUnaryExpr(expr *UnaryExpr) *TypeAnnotation {
	operandType := tc.checkExpression(expr.Operand)
	if operandType == nil {
		return nil
	}

	switch expr.Operator {
	case TOKEN_MINUS:
		if !tc.isNumericType(operandType) {
			tc.addError(expr.Location, "unary minus requires numeric type, got %s", tc.typeToString(operandType))
			return nil
		}
		return operandType

	case TOKEN_NOT:
		if operandType.Name != "bool" {
			tc.addError(expr.Location, "logical not requires boolean type, got %s", tc.typeToString(operandType))
			return nil
		}
		return operandType

	case TOKEN_TILDE:
		if !tc.isIntegerType(operandType) {
			tc.addError(expr.Location, "bitwise not requires integer type, got %s", tc.typeToString(operandType))
			return nil
		}
		return operandType

	default:
		tc.addError(expr.Location, "unknown unary operator")
		return nil
	}
}

// checkCallExpr checks a function call expression
func (tc *TypeChecker) checkCallExpr(expr *CallExpr) *TypeAnnotation {
	// Get function name
	var funcName string
	if ident, ok := expr.Function.(*IdentExpr); ok {
		funcName = ident.Name
	} else {
		tc.addError(expr.Location, "complex function expressions not yet supported")
		return nil
	}

	// Check for non-deterministic function calls
	if tc.isNonDeterministicFunction(funcName) {
		tc.addError(expr.Location, "non-deterministic function '%s' is not allowed in QuanticScript", funcName)
		return nil
	}

	// Look up function
	funcType := tc.lookupFunction(funcName)
	if funcType == nil {
		tc.addError(expr.Location, "undefined function '%s'", funcName)
		return nil
	}

	// Check argument count
	if len(expr.Arguments) != len(funcType.Parameters) {
		tc.addError(expr.Location, "function '%s' expects %d arguments, got %d",
			funcName, len(funcType.Parameters), len(expr.Arguments))
		return nil
	}

	// Check argument types
	for i, arg := range expr.Arguments {
		argType := tc.checkExpression(arg)
		expectedType := funcType.Parameters[i]
		if argType != nil && !tc.typesCompatible(argType, expectedType) {
			tc.addError(arg.Loc(), "argument %d: type mismatch: expected %s, got %s",
				i+1, tc.typeToString(expectedType), tc.typeToString(argType))
		}
	}

	return funcType.ReturnType
}

// isNonDeterministicFunction checks if a function is non-deterministic
func (tc *TypeChecker) isNonDeterministicFunction(name string) bool {
	nonDeterministicFuncs := []string{
		"random", "rand", "getRandom", "randomBytes",
		"now", "currentTime", "getTime", "timestamp",
		"readFile", "writeFile", "openFile",
		"fetch", "httpGet", "httpPost", "connect",
	}

	for _, fn := range nonDeterministicFuncs {
		if name == fn {
			return true
		}
	}
	return false
}

// checkIdentExpr checks an identifier expression
func (tc *TypeChecker) checkIdentExpr(expr *IdentExpr) *TypeAnnotation {
	varInfo := tc.lookupVariable(expr.Name)
	if varInfo == nil {
		tc.addError(expr.Location, "undefined variable '%s'", expr.Name)
		return nil
	}
	return varInfo.Type
}

// checkIntLiteral checks an integer literal
func (tc *TypeChecker) checkIntLiteral(expr *IntLiteral) *TypeAnnotation {
	// Default to i64 for integer literals
	return &TypeAnnotation{Name: "i64", Location: expr.Location}
}

// checkStringLiteral checks a string literal
func (tc *TypeChecker) checkStringLiteral(expr *StringLiteral) *TypeAnnotation {
	return &TypeAnnotation{Name: "string", Location: expr.Location}
}

// checkBoolLiteral checks a boolean literal
func (tc *TypeChecker) checkBoolLiteral(expr *BoolLiteral) *TypeAnnotation {
	return &TypeAnnotation{Name: "bool", Location: expr.Location}
}

// checkNullLiteral checks a null literal
func (tc *TypeChecker) checkNullLiteral(expr *NullLiteral) *TypeAnnotation {
	return &TypeAnnotation{Name: "null", Location: expr.Location, Nullable: true}
}

// checkArrayExpr checks an array expression
func (tc *TypeChecker) checkArrayExpr(expr *ArrayExpr) *TypeAnnotation {
	if len(expr.Elements) == 0 {
		tc.addError(expr.Location, "cannot infer type of empty array")
		return nil
	}

	// Infer element type from first element
	elemType := tc.checkExpression(expr.Elements[0])
	if elemType == nil {
		return nil
	}

	// Check all elements have the same type
	for i, elem := range expr.Elements[1:] {
		t := tc.checkExpression(elem)
		if t != nil && !tc.typesCompatible(t, elemType) {
			tc.addError(elem.Loc(), "array element %d: type mismatch: expected %s, got %s",
				i+1, tc.typeToString(elemType), tc.typeToString(t))
		}
	}

	return &TypeAnnotation{
		Name:     elemType.Name,
		IsArray:  true,
		Location: expr.Location,
	}
}

// checkIndexExpr checks an index expression
func (tc *TypeChecker) checkIndexExpr(expr *IndexExpr) *TypeAnnotation {
	arrayType := tc.checkExpression(expr.Array)
	indexType := tc.checkExpression(expr.Index)

	if arrayType == nil || indexType == nil {
		return nil
	}

	if !arrayType.IsArray {
		tc.addError(expr.Location, "cannot index non-array type %s", tc.typeToString(arrayType))
		return nil
	}

	if !tc.isIntegerType(indexType) {
		tc.addError(expr.Index.Loc(), "array index must be an integer, got %s", tc.typeToString(indexType))
		return nil
	}

	// Return element type
	return &TypeAnnotation{
		Name:     arrayType.Name,
		Location: expr.Location,
	}
}

// checkMemberExpr checks a member expression
func (tc *TypeChecker) checkMemberExpr(expr *MemberExpr) *TypeAnnotation {
	objectType := tc.checkExpression(expr.Object)
	if objectType == nil {
		return nil
	}

	// For now, we don't support struct types, so member access is not allowed
	tc.addError(expr.Location, "member access not yet supported for type %s", tc.typeToString(objectType))
	return nil
}

// Helper functions

// typesCompatible checks if two types are compatible
func (tc *TypeChecker) typesCompatible(t1, t2 *TypeAnnotation) bool {
	if t1 == nil || t2 == nil {
		return false
	}

	// Check base type
	if t1.Name != t2.Name {
		return false
	}

	// Check array dimension
	if t1.IsArray != t2.IsArray {
		return false
	}

	// Nullable types can accept non-nullable values
	// but not vice versa
	if t2.Nullable && !t1.Nullable {
		return false
	}

	return true
}

// isNumericType checks if a type is numeric
func (tc *TypeChecker) isNumericType(t *TypeAnnotation) bool {
	return tc.isIntegerType(t)
}

// isIntegerType checks if a type is an integer
func (tc *TypeChecker) isIntegerType(t *TypeAnnotation) bool {
	if t == nil {
		return false
	}
	switch t.Name {
	case "i8", "i16", "i32", "i64", "u8", "u16", "u32", "u64":
		return true
	}
	return false
}

// isFloatingPointType checks if a type is a floating-point type
func (tc *TypeChecker) isFloatingPointType(t *TypeAnnotation) bool {
	if t == nil {
		return false
	}
	switch t.Name {
	case "f32", "f64", "float", "double":
		return true
	}
	return false
}

// typeToString converts a type annotation to a string
func (tc *TypeChecker) typeToString(t *TypeAnnotation) string {
	if t == nil {
		return "unknown"
	}
	s := t.Name
	if t.IsArray {
		s += "[]"
	}
	if t.Nullable {
		s += " | null"
	}
	return s
}

// isIdentifier checks if a string is a valid identifier
func isIdentifier(s string) bool {
	if len(s) == 0 {
		return false
	}
	// Simple check: starts with letter or underscore
	c := s[0]
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || c == '_'
}

// TypeCheckerError represents a type checking error
type TypeCheckerError struct {
	Location SourceLocation
	Message  string
}

func (e *TypeCheckerError) Error() string {
	return fmt.Sprintf("%s: %s", e.Location, e.Message)
}
