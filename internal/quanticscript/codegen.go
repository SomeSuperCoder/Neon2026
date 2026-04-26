package quanticscript

import (
	"fmt"
	"strconv"
)

// CodeGenerator generates bytecode from AST
type CodeGenerator struct {
	bytecode       []byte
	errors         []error
	functions      map[string]int   // function name -> bytecode offset
	constants      map[string]int64 // constant name -> value (for i64 constants)
	localVars      map[string]int   // variable name -> local memory offset
	nextLocalSlot  int
	breakLabels    []int        // stack of break target offsets (for loops)
	continueLabels []int        // stack of continue target offsets (for loops)
	patchList      []patchEntry // list of offsets that need patching
}

// patchEntry represents a bytecode offset that needs to be patched later
type patchEntry struct {
	offset     int    // bytecode offset to patch
	targetName string // label/function name to resolve
	isRelative bool   // true for relative jumps (JMP/JMPIF), false for absolute (CALL)
}

// NewCodeGenerator creates a new code generator
func NewCodeGenerator() *CodeGenerator {
	return &CodeGenerator{
		bytecode:  make([]byte, 0),
		errors:    make([]error, 0),
		functions: make(map[string]int),
		constants: make(map[string]int64),
		localVars: make(map[string]int),
		patchList: make([]patchEntry, 0),
	}
}

// Generate generates bytecode from a program AST
func (cg *CodeGenerator) Generate(program *Program) ([]byte, error) {
	// First pass: process constant declarations
	for _, decl := range program.Declarations {
		if constDecl, ok := decl.(*ConstDecl); ok {
			cg.processConstDecl(constDecl)
		}
	}

	// Find the entry function
	var entryFunc *FunctionDecl
	for _, decl := range program.Declarations {
		if fn, ok := decl.(*FunctionDecl); ok {
			if fn.Name == "entry" && fn.IsExport {
				entryFunc = fn
				break
			}
		}
	}

	if entryFunc == nil {
		return nil, fmt.Errorf("no exported 'entry' function found")
	}

	// Generate entry wrapper at offset 0
	cg.generateEntryWrapper(entryFunc)

	// Second pass: collect function offsets
	for _, decl := range program.Declarations {
		if fn, ok := decl.(*FunctionDecl); ok {
			cg.functions[fn.Name] = len(cg.bytecode)
			cg.generateFunction(fn)
		}
	}

	// Third pass: patch function calls
	if err := cg.patchReferences(); err != nil {
		return nil, err
	}

	if len(cg.errors) > 0 {
		return nil, cg.errors[0]
	}

	return cg.bytecode, nil
}

// processConstDecl processes a constant declaration
func (cg *CodeGenerator) processConstDecl(constDecl *ConstDecl) {
	// For now, only support integer constants
	if intLit, ok := constDecl.Value.(*IntLiteral); ok {
		val, err := strconv.ParseInt(intLit.Value, 0, 64)
		if err != nil {
			cg.addError(constDecl.Location, "invalid integer constant: %v", err)
			return
		}
		cg.constants[constDecl.Name] = val
	} else if _, ok := constDecl.Value.(*ArrayExpr); ok {
		// For array constants (like SYSTEM_PROGRAM_ID), we'll handle them specially
		// Store a marker value to indicate it's an array constant
		cg.constants[constDecl.Name] = -1 // Special marker for array constants
	} else {
		cg.addError(constDecl.Location, "only integer and array constants are supported")
	}
}

// Errors returns the list of code generation errors
func (cg *CodeGenerator) Errors() []error {
	return cg.errors
}

// addError adds a code generation error
func (cg *CodeGenerator) addError(loc SourceLocation, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	err := &CodeGenError{
		Location: loc,
		Message:  msg,
	}
	cg.errors = append(cg.errors, err)
}

// emit emits a single byte
func (cg *CodeGenerator) emit(b byte) {
	cg.bytecode = append(cg.bytecode, b)
}

// emitOpcode emits an opcode
func (cg *CodeGenerator) emitOpcode(op Opcode) {
	cg.emit(byte(op))
}

// emitU16 emits a 16-bit unsigned integer (little-endian)
func (cg *CodeGenerator) emitU16(val uint16) {
	cg.emit(byte(val & 0xFF))
	cg.emit(byte((val >> 8) & 0xFF))
}

// emitI32 emits a 32-bit signed integer (little-endian)
func (cg *CodeGenerator) emitI32(val int32) {
	cg.emit(byte(val & 0xFF))
	cg.emit(byte((val >> 8) & 0xFF))
	cg.emit(byte((val >> 16) & 0xFF))
	cg.emit(byte((val >> 24) & 0xFF))
}

// emitI64 emits a 64-bit signed integer (little-endian)
func (cg *CodeGenerator) emitI64(val int64) {
	cg.emit(byte(val & 0xFF))
	cg.emit(byte((val >> 8) & 0xFF))
	cg.emit(byte((val >> 16) & 0xFF))
	cg.emit(byte((val >> 24) & 0xFF))
	cg.emit(byte((val >> 32) & 0xFF))
	cg.emit(byte((val >> 40) & 0xFF))
	cg.emit(byte((val >> 48) & 0xFF))
	cg.emit(byte((val >> 56) & 0xFF))
}

// emitU64 emits a 64-bit unsigned integer (little-endian)
func (cg *CodeGenerator) emitU64(val uint64) {
	cg.emit(byte(val & 0xFF))
	cg.emit(byte((val >> 8) & 0xFF))
	cg.emit(byte((val >> 16) & 0xFF))
	cg.emit(byte((val >> 24) & 0xFF))
	cg.emit(byte((val >> 32) & 0xFF))
	cg.emit(byte((val >> 40) & 0xFF))
	cg.emit(byte((val >> 48) & 0xFF))
	cg.emit(byte((val >> 56) & 0xFF))
}

// currentOffset returns the current bytecode offset
func (cg *CodeGenerator) currentOffset() int {
	return len(cg.bytecode)
}

// generateEntryWrapper generates the entry point wrapper
// This wrapper receives the ExecutionContext and marshals it to the entry function
func (cg *CodeGenerator) generateEntryWrapper(entryFunc *FunctionDecl) {
	// The entry wrapper is generated at offset 0
	// It expects the ExecutionContext to be available through special instructions

	// For now, we'll generate a simple wrapper that:
	// 1. Calls the entry function
	// 2. Returns the result

	// In a full implementation, this would:
	// - Extract instruction data using GETINSTRDATA
	// - Extract program ID using GETPROGRAMID
	// - Extract signers using GETSIGNER
	// - Marshal these into the entry function's expected parameters
	// - Handle the Result<void, Error> return type

	// For MVP, we'll just call the entry function directly
	cg.emitOpcode(OpCall)

	// Add to patch list (will be resolved when entry function is generated)
	offsetPos := cg.currentOffset()
	cg.emitI32(0) // Placeholder
	cg.patchList = append(cg.patchList, patchEntry{
		offset:     offsetPos,
		targetName: "entry",
		isRelative: false,
	})

	// Return from wrapper
	cg.emitOpcode(OpRet)
}

// generateFunction generates bytecode for a function
func (cg *CodeGenerator) generateFunction(fn *FunctionDecl) {
	// Reset local variable tracking for this function
	cg.localVars = make(map[string]int)
	cg.nextLocalSlot = 0

	// Allocate slots for parameters
	for _, param := range fn.Parameters {
		cg.localVars[param.Name] = cg.nextLocalSlot
		cg.nextLocalSlot++
	}

	// Pop arguments from stack into local memory slots (reverse order, stack is LIFO)
	for i := len(fn.Parameters) - 1; i >= 0; i-- {
		slot := cg.localVars[fn.Parameters[i].Name]
		cg.emitOpcode(OpStore)
		cg.emitU16(uint16(slot))
	}

	// Generate function body
	cg.generateBlock(fn.Body)

	// Ensure function ends with RET
	cg.emitOpcode(OpRet)
}

// generateBlock generates bytecode for a block statement
func (cg *CodeGenerator) generateBlock(block *BlockStmt) {
	for _, stmt := range block.Statements {
		cg.generateStatement(stmt)
	}
}

// generateStatement generates bytecode for a statement
func (cg *CodeGenerator) generateStatement(stmt Statement) {
	switch s := stmt.(type) {
	case *ReturnStmt:
		cg.generateReturnStmt(s)
	case *LetStmt:
		cg.generateLetStmt(s)
	case *ExprStmt:
		// Check if this is a void function call
		isVoidCall := false
		if callExpr, ok := s.Expression.(*CallExpr); ok {
			if ident, ok := callExpr.Function.(*IdentExpr); ok {
				if cg.isBuiltinFunction(ident.Name) && cg.isVoidBuiltinFunction(ident.Name) {
					isVoidCall = true
				}
			}
		}

		cg.generateExpression(s.Expression)

		// Only pop the result if it's not a void function call
		if !isVoidCall {
			cg.emitOpcode(OpPop)
		}
	case *IfStmt:
		cg.generateIfStmt(s)
	case *WhileStmt:
		cg.generateWhileStmt(s)
	case *ForStmt:
		cg.generateForStmt(s)
	case *AssignmentStmt:
		cg.generateAssignmentStmt(s)
	case *AsmBlock:
		cg.generateAsmBlock(s)
	case *BlockStmt:
		cg.generateBlock(s)
	}
}

// generateReturnStmt generates bytecode for a return statement
func (cg *CodeGenerator) generateReturnStmt(stmt *ReturnStmt) {
	if stmt.Value != nil {
		cg.generateExpression(stmt.Value)
	}
	cg.emitOpcode(OpRet)
}

// generateLetStmt generates bytecode for a variable declaration
func (cg *CodeGenerator) generateLetStmt(stmt *LetStmt) {
	// Allocate local slot for variable
	slot := cg.nextLocalSlot
	cg.localVars[stmt.Name] = slot
	cg.nextLocalSlot++

	// Generate initializer if present
	if stmt.Value != nil {
		cg.generateExpression(stmt.Value)
		// Store to local memory
		cg.emitOpcode(OpStore)
		cg.emitU16(uint16(slot))
	}
}

// generateIfStmt generates bytecode for an if statement
func (cg *CodeGenerator) generateIfStmt(stmt *IfStmt) {
	// Generate condition
	cg.generateExpression(stmt.Condition)

	// Emit conditional jump (will be patched)
	cg.emitOpcode(OpJmpIf)
	thenJumpOffset := cg.currentOffset()
	cg.emitI32(0) // Placeholder

	// Generate else block (if present)
	if stmt.ElseBlock != nil {
		cg.generateBlock(stmt.ElseBlock)
	}

	// Emit unconditional jump to skip then block
	cg.emitOpcode(OpJmp)
	elseJumpOffset := cg.currentOffset()
	cg.emitI32(0) // Placeholder

	// Patch then jump to here
	thenTarget := cg.currentOffset()
	cg.patchRelativeJump(thenJumpOffset, thenTarget)

	// Generate then block
	cg.generateBlock(stmt.ThenBlock)

	// Patch else jump to here
	elseTarget := cg.currentOffset()
	cg.patchRelativeJump(elseJumpOffset, elseTarget)
}

// generateWhileStmt generates bytecode for a while loop
func (cg *CodeGenerator) generateWhileStmt(stmt *WhileStmt) {
	// Loop start
	loopStart := cg.currentOffset()

	// Generate condition
	cg.generateExpression(stmt.Condition)

	// Emit conditional jump to exit (will be patched)
	cg.emitOpcode(OpJmpIf)
	exitJumpOffset := cg.currentOffset()
	cg.emitI32(0) // Placeholder

	// Emit unconditional jump to skip body
	cg.emitOpcode(OpJmp)
	skipBodyOffset := cg.currentOffset()
	cg.emitI32(0) // Placeholder

	// Patch condition jump to body
	bodyStart := cg.currentOffset()
	cg.patchRelativeJump(exitJumpOffset, bodyStart)

	// Generate body
	cg.generateBlock(stmt.Body)

	// Jump back to loop start
	cg.emitOpcode(OpJmp)
	cg.emitI32(int32(loopStart - (cg.currentOffset() + 4)))

	// Patch skip body jump to here
	loopExit := cg.currentOffset()
	cg.patchRelativeJump(skipBodyOffset, loopExit)
}

// generateForStmt generates bytecode for a for loop
func (cg *CodeGenerator) generateForStmt(stmt *ForStmt) {
	// Generate init
	if stmt.Init != nil {
		cg.generateStatement(stmt.Init)
	}

	// Loop start
	loopStart := cg.currentOffset()

	// Generate condition (if present)
	if stmt.Condition != nil {
		cg.generateExpression(stmt.Condition)

		// Emit conditional jump to exit
		cg.emitOpcode(OpJmpIf)
		exitJumpOffset := cg.currentOffset()
		cg.emitI32(0) // Placeholder

		// Emit unconditional jump to skip body
		cg.emitOpcode(OpJmp)
		skipBodyOffset := cg.currentOffset()
		cg.emitI32(0) // Placeholder

		// Patch condition jump to body
		bodyStart := cg.currentOffset()
		cg.patchRelativeJump(exitJumpOffset, bodyStart)

		// Generate body
		cg.generateBlock(stmt.Body)

		// Generate update
		if stmt.Update != nil {
			cg.generateStatement(stmt.Update)
		}

		// Jump back to loop start
		cg.emitOpcode(OpJmp)
		cg.emitI32(int32(loopStart - (cg.currentOffset() + 4)))

		// Patch skip body jump to here
		loopExit := cg.currentOffset()
		cg.patchRelativeJump(skipBodyOffset, loopExit)
	} else {
		// Infinite loop (no condition)
		cg.generateBlock(stmt.Body)

		// Generate update
		if stmt.Update != nil {
			cg.generateStatement(stmt.Update)
		}

		// Jump back to loop start
		cg.emitOpcode(OpJmp)
		cg.emitI32(int32(loopStart - (cg.currentOffset() + 4)))
	}
}

// generateAssignmentStmt generates bytecode for an assignment statement
func (cg *CodeGenerator) generateAssignmentStmt(stmt *AssignmentStmt) {
	// Generate value
	cg.generateExpression(stmt.Value)

	// Handle compound assignment operators
	if stmt.Operator != TOKEN_ASSIGN {
		// Load current value
		if ident, ok := stmt.Target.(*IdentExpr); ok {
			slot, ok := cg.localVars[ident.Name]
			if !ok {
				cg.addError(stmt.Location, "undefined variable '%s'", ident.Name)
				return
			}
			cg.emitOpcode(OpLoad)
			cg.emitU16(uint16(slot))

			// Swap so value is on top
			cg.emitOpcode(OpSwap)

			// Apply operator
			switch stmt.Operator {
			case TOKEN_PLUS_ASSIGN:
				cg.emitOpcode(OpAdd)
			case TOKEN_MINUS_ASSIGN:
				cg.emitOpcode(OpSub)
			}
		}
	}

	// Store to target
	if ident, ok := stmt.Target.(*IdentExpr); ok {
		slot, ok := cg.localVars[ident.Name]
		if !ok {
			cg.addError(stmt.Location, "undefined variable '%s'", ident.Name)
			return
		}
		cg.emitOpcode(OpStore)
		cg.emitU16(uint16(slot))
	} else {
		cg.addError(stmt.Location, "complex assignment targets not yet supported")
	}
}

// generateAsmBlock generates bytecode for an inline assembly block
func (cg *CodeGenerator) generateAsmBlock(block *AsmBlock) {
	for _, instr := range block.Instructions {
		cg.generateAsmInstruction(&instr)
	}
}

// generateAsmInstruction generates bytecode for a single assembly instruction
func (cg *CodeGenerator) generateAsmInstruction(instr *AsmInstruction) {
	// Look up opcode
	opcode, ok := mnemonicToOpcode(instr.Mnemonic)
	if !ok {
		cg.addError(instr.Location, "unknown assembly instruction '%s'", instr.Mnemonic)
		return
	}

	// Emit opcode
	cg.emitOpcode(opcode)

	// Handle operands based on instruction type
	switch opcode {
	case OpPush:
		cg.generateAsmPushOperands(instr)
	case OpLoad, OpStore:
		cg.generateAsmMemoryOperands(instr)
	case OpJmp, OpJmpIf, OpCall:
		cg.generateAsmJumpOperands(instr, opcode)
		// Most instructions have no operands
	}
}

// generateAsmPushOperands generates operands for PUSH instruction in assembly
func (cg *CodeGenerator) generateAsmPushOperands(instr *AsmInstruction) {
	if len(instr.Operands) < 2 {
		cg.addError(instr.Location, "PUSH requires 2 operands (type value)")
		return
	}

	typeName := instr.Operands[0]
	valueStr := instr.Operands[1]

	// Check if value is a variable reference
	if isIdentifier(valueStr) {
		// Load variable value instead
		slot, ok := cg.localVars[valueStr]
		if !ok {
			cg.addError(instr.Location, "undefined variable '%s'", valueStr)
			return
		}
		// Replace PUSH with LOAD
		cg.bytecode = cg.bytecode[:len(cg.bytecode)-1] // Remove PUSH opcode
		cg.emitOpcode(OpLoad)
		cg.emitU16(uint16(slot))
		return
	}

	// Parse literal value
	switch typeName {
	case "i64":
		val, err := strconv.ParseInt(valueStr, 0, 64)
		if err != nil {
			cg.addError(instr.Location, "invalid i64 value '%s': %v", valueStr, err)
			return
		}
		cg.emit(byte(TypeI64))
		cg.emitI64(val)

	case "u64":
		val, err := strconv.ParseUint(valueStr, 0, 64)
		if err != nil {
			cg.addError(instr.Location, "invalid u64 value '%s': %v", valueStr, err)
			return
		}
		cg.emit(byte(TypeU64))
		cg.emitU64(val)

	case "bool":
		var val byte
		if valueStr == "true" || valueStr == "1" {
			val = 1
		} else if valueStr == "false" || valueStr == "0" {
			val = 0
		} else {
			cg.addError(instr.Location, "invalid bool value '%s'", valueStr)
			return
		}
		cg.emit(byte(TypeBool))
		cg.emit(val)

	default:
		cg.addError(instr.Location, "unsupported type '%s' for PUSH", typeName)
	}
}

// generateAsmMemoryOperands generates operands for LOAD/STORE instructions in assembly
func (cg *CodeGenerator) generateAsmMemoryOperands(instr *AsmInstruction) {
	if len(instr.Operands) < 1 {
		cg.addError(instr.Location, "memory instruction requires 1 operand")
		return
	}

	operand := instr.Operands[0]

	// Check if operand is a variable name
	if slot, ok := cg.localVars[operand]; ok {
		cg.emitU16(uint16(slot))
		return
	}

	// Otherwise parse as numeric offset
	offset, err := strconv.ParseUint(operand, 0, 16)
	if err != nil {
		cg.addError(instr.Location, "invalid memory offset '%s': %v", operand, err)
		return
	}

	if offset > 0xFFFF {
		cg.addError(instr.Location, "memory offset %d exceeds maximum (65535)", offset)
		return
	}

	cg.emitU16(uint16(offset))
}

// generateAsmJumpOperands generates operands for jump instructions in assembly
func (cg *CodeGenerator) generateAsmJumpOperands(instr *AsmInstruction, opcode Opcode) {
	if len(instr.Operands) < 1 {
		cg.addError(instr.Location, "jump instruction requires 1 operand")
		return
	}

	target := instr.Operands[0]

	// Check if it's a numeric offset
	if offset, err := strconv.ParseInt(target, 0, 32); err == nil {
		cg.emitI32(int32(offset))
		return
	}

	// Otherwise it's a label/function reference - add to patch list
	offsetPos := cg.currentOffset()
	cg.emitI32(0) // Placeholder
	cg.patchList = append(cg.patchList, patchEntry{
		offset:     offsetPos,
		targetName: target,
		isRelative: opcode == OpJmp || opcode == OpJmpIf,
	})
}

// generateExpression generates bytecode for an expression
func (cg *CodeGenerator) generateExpression(expr Expression) {
	switch e := expr.(type) {
	case *BinaryExpr:
		cg.generateBinaryExpr(e)
	case *UnaryExpr:
		cg.generateUnaryExpr(e)
	case *CallExpr:
		cg.generateCallExpr(e)
	case *IdentExpr:
		cg.generateIdentExpr(e)
	case *IntLiteral:
		cg.generateIntLiteral(e)
	case *StringLiteral:
		cg.generateStringLiteral(e)
	case *BoolLiteral:
		cg.generateBoolLiteral(e)
	case *NullLiteral:
		cg.generateNullLiteral(e)
	case *ArrayExpr:
		cg.generateArrayExpr(e)
	case *IndexExpr:
		cg.generateIndexExpr(e)
	case *MemberExpr:
		cg.generateMemberExpr(e)
	}
}

// generateBinaryExpr generates bytecode for a binary expression
func (cg *CodeGenerator) generateBinaryExpr(expr *BinaryExpr) {
	// Generate left operand
	cg.generateExpression(expr.Left)

	// Generate right operand
	cg.generateExpression(expr.Right)

	// Generate operator
	switch expr.Operator {
	case TOKEN_PLUS:
		cg.emitOpcode(OpAdd)
	case TOKEN_MINUS:
		cg.emitOpcode(OpSub)
	case TOKEN_STAR:
		cg.emitOpcode(OpMul)
	case TOKEN_SLASH:
		cg.emitOpcode(OpDiv)
	case TOKEN_PERCENT:
		cg.emitOpcode(OpMod)
	case TOKEN_EQ:
		cg.emitOpcode(OpEq)
	case TOKEN_NEQ:
		// EQ followed by NOT
		cg.emitOpcode(OpEq)
		cg.emitOpcode(OpNot)
	case TOKEN_LT:
		cg.emitOpcode(OpLt)
	case TOKEN_GT:
		cg.emitOpcode(OpGt)
	case TOKEN_LTE:
		cg.emitOpcode(OpLte)
	case TOKEN_GTE:
		cg.emitOpcode(OpGte)
	case TOKEN_AND:
		cg.emitOpcode(OpAnd)
	case TOKEN_OR:
		cg.emitOpcode(OpOr)
	case TOKEN_AMPERSAND:
		cg.emitOpcode(OpBand)
	case TOKEN_PIPE:
		cg.emitOpcode(OpBor)
	case TOKEN_CARET:
		cg.emitOpcode(OpBxor)
	case TOKEN_LSHIFT:
		cg.emitOpcode(OpShl)
	case TOKEN_RSHIFT:
		cg.emitOpcode(OpShr)
	default:
		cg.addError(expr.Location, "unsupported binary operator")
	}
}

// generateUnaryExpr generates bytecode for a unary expression
func (cg *CodeGenerator) generateUnaryExpr(expr *UnaryExpr) {
	// Generate operand
	cg.generateExpression(expr.Operand)

	// Generate operator
	switch expr.Operator {
	case TOKEN_MINUS:
		// Negate: 0 - value
		cg.emitOpcode(OpPush)
		cg.emit(byte(TypeI64))
		cg.emitI64(0)
		cg.emitOpcode(OpSwap)
		cg.emitOpcode(OpSub)
	case TOKEN_NOT:
		cg.emitOpcode(OpNot)
	case TOKEN_TILDE:
		cg.emitOpcode(OpBnot)
	default:
		cg.addError(expr.Location, "unsupported unary operator")
	}
}

// generateCallExpr generates bytecode for a function call
func (cg *CodeGenerator) generateCallExpr(expr *CallExpr) {
	// Get function name
	var funcName string
	if ident, ok := expr.Function.(*IdentExpr); ok {
		funcName = ident.Name
	} else {
		cg.addError(expr.Location, "complex function expressions not yet supported")
		return
	}

	// Check if this is a builtin function
	if cg.isBuiltinFunction(funcName) {
		// Generate arguments (in order)
		for _, arg := range expr.Arguments {
			cg.generateExpression(arg)
		}

		// Emit the appropriate builtin opcode
		cg.emitBuiltinCall(funcName, expr.Location)
		return
	}

	// Generate arguments (in order) for regular function calls
	for _, arg := range expr.Arguments {
		cg.generateExpression(arg)
	}

	// Emit CALL instruction
	cg.emitOpcode(OpCall)

	// Add to patch list (will be resolved in second pass)
	offsetPos := cg.currentOffset()
	cg.emitI32(0) // Placeholder
	cg.patchList = append(cg.patchList, patchEntry{
		offset:     offsetPos,
		targetName: funcName,
		isRelative: false, // CALL uses absolute offset
	})
}

// generateIdentExpr generates bytecode for an identifier
func (cg *CodeGenerator) generateIdentExpr(expr *IdentExpr) {
	// Check if it's a constant first
	if val, ok := cg.constants[expr.Name]; ok {
		// Push constant value
		cg.emitOpcode(OpPush)
		cg.emit(byte(TypeI64))
		cg.emitI64(val)
		return
	}

	// Check if it's a local variable
	slot, ok := cg.localVars[expr.Name]
	if !ok {
		cg.addError(expr.Location, "undefined variable '%s'", expr.Name)
		return
	}

	// Load from local memory
	cg.emitOpcode(OpLoad)
	cg.emitU16(uint16(slot))
}

// generateIntLiteral generates bytecode for an integer literal
func (cg *CodeGenerator) generateIntLiteral(expr *IntLiteral) {
	val, err := strconv.ParseInt(expr.Value, 0, 64)
	if err != nil {
		cg.addError(expr.Location, "invalid integer literal '%s': %v", expr.Value, err)
		return
	}

	cg.emitOpcode(OpPush)
	cg.emit(byte(TypeI64))
	cg.emitI64(val)
}

// generateStringLiteral generates bytecode for a string literal
func (cg *CodeGenerator) generateStringLiteral(expr *StringLiteral) {
	cg.emitOpcode(OpPush)
	cg.emit(byte(TypeString))

	// Emit string length as 8-byte little-endian
	strLen := uint64(len(expr.Value))
	cg.emit(byte(strLen))
	cg.emit(byte(strLen >> 8))
	cg.emit(byte(strLen >> 16))
	cg.emit(byte(strLen >> 24))
	cg.emit(byte(strLen >> 32))
	cg.emit(byte(strLen >> 40))
	cg.emit(byte(strLen >> 48))
	cg.emit(byte(strLen >> 56))

	// Emit string bytes
	for _, ch := range expr.Value {
		cg.emit(byte(ch))
	}
}

// generateBoolLiteral generates bytecode for a boolean literal
func (cg *CodeGenerator) generateBoolLiteral(expr *BoolLiteral) {
	cg.emitOpcode(OpPush)
	cg.emit(byte(TypeBool))
	if expr.Value {
		cg.emit(1)
	} else {
		cg.emit(0)
	}
}

// generateNullLiteral generates bytecode for a null literal
func (cg *CodeGenerator) generateNullLiteral(expr *NullLiteral) {
	// Push null value (represented as 0)
	cg.emitOpcode(OpPush)
	cg.emit(byte(TypeI64))
	cg.emitI64(0)
}

// generateArrayExpr generates bytecode for an array literal
func (cg *CodeGenerator) generateArrayExpr(expr *ArrayExpr) {
	// Arrays are not yet fully supported
	cg.addError(expr.Location, "array literals not yet supported in bytecode generation")
}

// generateIndexExpr generates bytecode for array indexing
func (cg *CodeGenerator) generateIndexExpr(expr *IndexExpr) {
	// Array indexing is not yet fully supported
	cg.addError(expr.Location, "array indexing not yet supported in bytecode generation")
}

// generateMemberExpr generates bytecode for member access
func (cg *CodeGenerator) generateMemberExpr(expr *MemberExpr) {
	// Member access is not yet fully supported
	cg.addError(expr.Location, "member access not yet supported in bytecode generation")
}

// patchRelativeJump patches a relative jump instruction
func (cg *CodeGenerator) patchRelativeJump(jumpOffset int, targetOffset int) {
	// Calculate relative offset from the position after the jump instruction
	relativeOffset := int32(targetOffset - (jumpOffset + 4))

	// Patch the bytecode
	cg.bytecode[jumpOffset] = byte(relativeOffset & 0xFF)
	cg.bytecode[jumpOffset+1] = byte((relativeOffset >> 8) & 0xFF)
	cg.bytecode[jumpOffset+2] = byte((relativeOffset >> 16) & 0xFF)
	cg.bytecode[jumpOffset+3] = byte((relativeOffset >> 24) & 0xFF)
}

// patchReferences resolves all function/label references
func (cg *CodeGenerator) patchReferences() error {
	for _, patch := range cg.patchList {
		// Look up target offset
		targetOffset, ok := cg.functions[patch.targetName]
		if !ok {
			return fmt.Errorf("undefined function '%s'", patch.targetName)
		}

		var offsetValue int32
		if patch.isRelative {
			// Calculate relative offset
			offsetValue = int32(targetOffset - (patch.offset + 4))
		} else {
			// Use absolute offset
			offsetValue = int32(targetOffset)
		}

		// Patch the bytecode
		cg.bytecode[patch.offset] = byte(offsetValue & 0xFF)
		cg.bytecode[patch.offset+1] = byte((offsetValue >> 8) & 0xFF)
		cg.bytecode[patch.offset+2] = byte((offsetValue >> 16) & 0xFF)
		cg.bytecode[patch.offset+3] = byte((offsetValue >> 24) & 0xFF)
	}

	return nil
}

// isBuiltinFunction checks if a function name is a builtin
func (cg *CodeGenerator) isBuiltinFunction(name string) bool {
	builtins := map[string]bool{
		// Blockchain operations
		"getBalance":         true,
		"updateBalance":      true,
		"transfer":           true,
		"getInstructionData": true,
		"getProgramId":       true,
		"getProgramID":       true, // Alias
		"hasSigner":          true,
		"getSigner":          true,
		"getFile":            true,
		"getFileMut":         true,
		"updateFile":         true,
		"createFile":         true,
		"deleteFile":         true,
		// Cryptographic operations
		"sha256":          true,
		"hashBytes":       true,
		"verifySignature": true,
		"derivePublicKey": true,
		// Cross-program invocation
		"invoke":         true,
		"getInvokeDepth": true,
		// Query operations
		"queryBlock":       true,
		"queryTransaction": true,
		"queryInstruction": true,
		// Collection operations
		"arrayNew":    true,
		"arrayLength": true,
		"len":         true, // Alias for arrayLength
		"arrayGet":    true,
		"arraySet":    true,
		"arrayPush":   true,
		"arrayPop":    true,
		"mapNew":      true,
		"mapGet":      true,
		"mapSet":      true,
		"mapHas":      true,
		"mapDel":      true,
		"setNew":      true,
		"setAdd":      true,
		"setHas":      true,
		"setDel":      true,
		// String operations
		"stringConcat":    true,
		"stringSubstring": true,
		"stringLength":    true,
		"stringToBytes":   true,
		"bytesToString":   true,
		// Math operations
		"min":  true,
		"max":  true,
		"abs":  true,
		"pow":  true,
		"sqrt": true,
		// Debug operations
		"log": true,
		// Conversion operations
		"slice":         true,
		"bytesToI64LE":  true,
		"bytesToFileID": true,
	}
	return builtins[name]
}

// isVoidBuiltinFunction checks if a builtin function returns void
func (cg *CodeGenerator) isVoidBuiltinFunction(name string) bool {
	voidBuiltins := map[string]bool{
		"updateBalance": true,
		"updateFile":    true,
		"transfer":      true,
		"createFile":    true,
		"deleteFile":    true,
		"log":           true,
	}
	return voidBuiltins[name]
}

// emitBuiltinCall emits bytecode for a builtin function call
func (cg *CodeGenerator) emitBuiltinCall(name string, loc SourceLocation) {
	switch name {
	// Blockchain operations
	case "getBalance":
		cg.emitOpcode(OpGetBalance)
	case "updateBalance":
		cg.emitOpcode(OpUpdateBalance)
	case "transfer":
		cg.emitOpcode(OpTransfer)
	case "getInstructionData":
		cg.emitOpcode(OpGetInstrData)
	case "getProgramId", "getProgramID":
		cg.emitOpcode(OpGetProgramID)
	case "hasSigner":
		cg.emitOpcode(OpHasSigner)
	case "getSigner":
		cg.emitOpcode(OpGetSigner)
	case "getFile":
		cg.emitOpcode(OpGetFile)
	case "getFileMut":
		cg.emitOpcode(OpGetFileMut)
	case "updateFile":
		cg.emitOpcode(OpUpdateFile)
	case "createFile":
		cg.emitOpcode(OpCreateFile)
	case "deleteFile":
		cg.emitOpcode(OpDeleteFile)
	// Cryptographic operations
	case "sha256":
		cg.emitOpcode(OpSha256)
	case "hashBytes":
		cg.emitOpcode(OpHashBytes)
	case "verifySignature":
		cg.emitOpcode(OpVerifySig)
	case "derivePublicKey":
		cg.emitOpcode(OpDerivePubKey)
	// Cross-program invocation
	case "invoke":
		cg.emitOpcode(OpInvoke)
	case "getInvokeDepth":
		// Push current invoke depth (this would need runtime support)
		// For now, emit a placeholder
		cg.addError(loc, "getInvokeDepth not yet implemented")
	// Query operations
	case "queryBlock":
		cg.emitOpcode(OpQueryBlock)
	case "queryTransaction":
		cg.emitOpcode(OpQueryTx)
	case "queryInstruction":
		cg.emitOpcode(OpQueryInstr)
	// Collection operations
	case "arrayNew":
		cg.emitOpcode(OpArrayNew)
	case "arrayLength", "len":
		cg.emitOpcode(OpArrayLen)
	case "arrayGet":
		cg.emitOpcode(OpArrayGet)
	case "arraySet":
		cg.emitOpcode(OpArraySet)
	case "arrayPush":
		cg.emitOpcode(OpArrayPush)
	case "arrayPop":
		cg.emitOpcode(OpArrayPop)
	case "mapNew":
		cg.emitOpcode(OpMapNew)
	case "mapGet":
		cg.emitOpcode(OpMapGet)
	case "mapSet":
		cg.emitOpcode(OpMapSet)
	case "mapHas":
		cg.emitOpcode(OpMapHas)
	case "mapDel":
		cg.emitOpcode(OpMapDel)
	case "setNew":
		cg.emitOpcode(OpSetNew)
	case "setAdd":
		cg.emitOpcode(OpSetAdd)
	case "setHas":
		cg.emitOpcode(OpSetHas)
	case "setDel":
		cg.emitOpcode(OpSetDel)
	// String operations
	case "stringConcat":
		cg.emitOpcode(OpStrConcat)
	case "stringSubstring":
		cg.emitOpcode(OpStrSubstring)
	case "stringLength":
		cg.emitOpcode(OpStrLen)
	case "stringToBytes":
		cg.emitOpcode(OpStrToBytes)
	case "bytesToString":
		cg.emitOpcode(OpStrFromBytes)
	// Debug operations
	case "log":
		cg.emitOpcode(OpLog)
	// Math operations
	case "min":
		cg.emitOpcode(OpMathMin)
	case "max":
		cg.emitOpcode(OpMathMax)
	case "abs":
		cg.emitOpcode(OpMathAbs)
	case "pow":
		cg.emitOpcode(OpMathPow)
	case "sqrt":
		// sqrt not yet implemented
		cg.addError(loc, "sqrt not yet implemented")
	// Conversion operations
	case "slice":
		cg.emitOpcode(OpSlice)
	case "bytesToI64LE":
		cg.emitOpcode(OpBytesToI64LE)
	case "bytesToFileID":
		cg.emitOpcode(OpBytesToFileID)
	default:
		cg.addError(loc, "unknown builtin function '%s'", name)
	}
}

// CodeGenError represents a code generation error
type CodeGenError struct {
	Location SourceLocation
	Message  string
}

func (e *CodeGenError) Error() string {
	return fmt.Sprintf("%s: %s", e.Location, e.Message)
}
