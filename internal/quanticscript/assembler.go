package quanticscript

import (
	"fmt"
	"strconv"
	"strings"
)

// Assembler converts assembly text to bytecode
type Assembler struct {
	lines       []string
	labels      map[string]int // label name -> bytecode offset
	unresolved  []unresolvedRef
	bytecode    []byte
	lineNumbers map[int]int // bytecode offset -> source line number
}

// unresolvedRef represents a label reference that needs to be resolved
type unresolvedRef struct {
	labelName      string
	bytecodeOffset int  // where to write the resolved offset
	isRelative     bool // true for JMP/JMPIF (relative), false for CALL (absolute)
	sourceLine     int
}

// NewAssembler creates a new assembler instance
func NewAssembler() *Assembler {
	return &Assembler{
		labels:      make(map[string]int),
		unresolved:  make([]unresolvedRef, 0),
		bytecode:    make([]byte, 0),
		lineNumbers: make(map[int]int),
	}
}

// Assemble converts assembly text to bytecode
func (a *Assembler) Assemble(source string) ([]byte, error) {
	// Split into lines and preprocess
	a.lines = strings.Split(source, "\n")

	// First pass: collect labels and generate bytecode
	if err := a.firstPass(); err != nil {
		return nil, err
	}

	// Second pass: resolve label references
	if err := a.secondPass(); err != nil {
		return nil, err
	}

	return a.bytecode, nil
}

// firstPass processes all lines, collecting labels and generating bytecode
func (a *Assembler) firstPass() error {
	for lineNum, line := range a.lines {
		// Remove comments and trim whitespace
		line = a.removeComment(line)
		line = strings.TrimSpace(line)

		// Skip empty lines
		if line == "" {
			continue
		}

		// Check for label definition
		if strings.HasSuffix(line, ":") {
			labelName := strings.TrimSuffix(line, ":")
			labelName = strings.TrimSpace(labelName)
			if labelName == "" {
				return fmt.Errorf("line %d: empty label name", lineNum+1)
			}
			// Store label position
			a.labels[labelName] = len(a.bytecode)
			continue
		}

		// Parse instruction
		if err := a.parseInstruction(line, lineNum+1); err != nil {
			return fmt.Errorf("line %d: %w", lineNum+1, err)
		}
	}

	return nil
}

// secondPass resolves all label references
func (a *Assembler) secondPass() error {
	for _, ref := range a.unresolved {
		targetOffset, ok := a.labels[ref.labelName]
		if !ok {
			return fmt.Errorf("line %d: undefined label '%s'", ref.sourceLine, ref.labelName)
		}

		var offsetValue int32
		if ref.isRelative {
			// For JMP/JMPIF: calculate relative offset from instruction end
			// The offset is relative to the position after reading the offset bytes
			offsetValue = int32(targetOffset - (ref.bytecodeOffset + 4))
		} else {
			// For CALL: use absolute offset
			offsetValue = int32(targetOffset)
		}

		// Write the resolved offset (4 bytes, little-endian)
		a.bytecode[ref.bytecodeOffset] = byte(offsetValue & 0xFF)
		a.bytecode[ref.bytecodeOffset+1] = byte((offsetValue >> 8) & 0xFF)
		a.bytecode[ref.bytecodeOffset+2] = byte((offsetValue >> 16) & 0xFF)
		a.bytecode[ref.bytecodeOffset+3] = byte((offsetValue >> 24) & 0xFF)
	}

	return nil
}

// removeComment removes comments from a line (comments start with ;)
func (a *Assembler) removeComment(line string) string {
	if idx := strings.Index(line, ";"); idx >= 0 {
		return line[:idx]
	}
	return line
}

// parseInstruction parses and encodes a single instruction
func (a *Assembler) parseInstruction(line string, lineNum int) error {
	// Split instruction and operands
	parts := strings.Fields(line)
	if len(parts) == 0 {
		return nil
	}

	mnemonic := strings.ToUpper(parts[0])
	operands := parts[1:]

	// Record line number for this bytecode offset
	a.lineNumbers[len(a.bytecode)] = lineNum

	// Look up opcode
	opcode, ok := mnemonicToOpcode(mnemonic)
	if !ok {
		return fmt.Errorf("unknown instruction '%s'", mnemonic)
	}

	// Emit opcode
	a.bytecode = append(a.bytecode, byte(opcode))

	// Parse and emit operands based on instruction type
	switch opcode {
	case OpPush:
		return a.parsePushOperands(operands, lineNum)
	case OpLoad, OpStore:
		return a.parseMemoryOperands(operands, lineNum)
	case OpJmp, OpJmpIf:
		return a.parseJumpOperands(operands, lineNum, true)
	case OpCall:
		return a.parseJumpOperands(operands, lineNum, false)
	case OpPop, OpDup, OpSwap:
		return a.parseNoOperands(operands, lineNum)
	case OpAdd, OpSub, OpMul, OpDiv, OpMod:
		return a.parseNoOperands(operands, lineNum)
	case OpEq, OpLt, OpGt, OpLte, OpGte:
		return a.parseNoOperands(operands, lineNum)
	case OpAnd, OpOr, OpNot:
		return a.parseNoOperands(operands, lineNum)
	case OpBand, OpBor, OpBxor, OpBnot, OpShl, OpShr:
		return a.parseNoOperands(operands, lineNum)
	case OpRet:
		return a.parseNoOperands(operands, lineNum)
	case OpGetFile, OpGetFileMut, OpUpdateFile:
		return a.parseNoOperands(operands, lineNum)
	case OpGetBalance, OpUpdateBalance:
		return a.parseNoOperands(operands, lineNum)
	case OpGetSigner, OpHasSigner:
		return a.parseNoOperands(operands, lineNum)
	case OpGetInstrData, OpGetProgramID:
		return a.parseNoOperands(operands, lineNum)
	case OpLoadG, OpStoreG:
		return a.parseNoOperands(operands, lineNum) // TODO: implement global state operands
	case OpInvoke, OpInvokeRet:
		return a.parseNoOperands(operands, lineNum)
	case OpSha256, OpVerifySig, OpDerivePubKey:
		return a.parseNoOperands(operands, lineNum)
	case OpQueryBlock, OpQueryTx, OpQueryInstr:
		return a.parseNoOperands(operands, lineNum)
	case OpBytesToI64LE:
		return a.parseNoOperands(operands, lineNum)
	case OpStrConcat, OpStrSubstring, OpStrLen, OpStrToBytes, OpStrFromBytes:
		return a.parseNoOperands(operands, lineNum)
	case OpDispatch:
		return a.parseNoOperands(operands, lineNum)
	// Collection Operations
	case OpArrayNew, OpArrayLen, OpArrayGet, OpArraySet, OpArrayPush, OpArrayPop:
		return a.parseNoOperands(operands, lineNum)
	case OpArrayMap, OpArrayFilter, OpArrayReduce, OpArraySort:
		return a.parseNoOperands(operands, lineNum)
	case OpMapNew, OpMapGet, OpMapSet, OpMapHas, OpMapDel:
		return a.parseNoOperands(operands, lineNum)
	case OpSetNew, OpSetAdd, OpSetHas, OpSetDel:
		return a.parseNoOperands(operands, lineNum)
	// Math Operations
	case OpMathMin, OpMathMax, OpMathAbs, OpMathPow:
		return a.parseNoOperands(operands, lineNum)
	// Conversion Operations
	case OpSlice, OpBytesToFileID:
		return a.parseNoOperands(operands, lineNum)
	// Transfer operation
	case OpTransfer:
		return a.parseNoOperands(operands, lineNum)
	default:
		return fmt.Errorf("unhandled opcode: %v", opcode)
	}
}

// parsePushOperands parses PUSH instruction operands
func (a *Assembler) parsePushOperands(operands []string, lineNum int) error {
	if len(operands) != 2 {
		return fmt.Errorf("PUSH requires 2 operands (type value), got %d", len(operands))
	}

	typeName := strings.ToLower(operands[0])
	valueStr := operands[1]

	switch typeName {
	case "i64":
		val, err := strconv.ParseInt(valueStr, 0, 64)
		if err != nil {
			return fmt.Errorf("invalid i64 value '%s': %w", valueStr, err)
		}
		// Emit type byte
		a.bytecode = append(a.bytecode, byte(TypeI64))
		// Emit value (8 bytes, little-endian)
		a.bytecode = append(a.bytecode,
			byte(val&0xFF),
			byte((val>>8)&0xFF),
			byte((val>>16)&0xFF),
			byte((val>>24)&0xFF),
			byte((val>>32)&0xFF),
			byte((val>>40)&0xFF),
			byte((val>>48)&0xFF),
			byte((val>>56)&0xFF),
		)

	case "u64":
		val, err := strconv.ParseUint(valueStr, 0, 64)
		if err != nil {
			return fmt.Errorf("invalid u64 value '%s': %w", valueStr, err)
		}
		// Emit type byte
		a.bytecode = append(a.bytecode, byte(TypeU64))
		// Emit value (8 bytes, little-endian)
		a.bytecode = append(a.bytecode,
			byte(val&0xFF),
			byte((val>>8)&0xFF),
			byte((val>>16)&0xFF),
			byte((val>>24)&0xFF),
			byte((val>>32)&0xFF),
			byte((val>>40)&0xFF),
			byte((val>>48)&0xFF),
			byte((val>>56)&0xFF),
		)

	case "bool":
		var val byte
		switch strings.ToLower(valueStr) {
		case "true", "1":
			val = 1
		case "false", "0":
			val = 0
		default:
			return fmt.Errorf("invalid bool value '%s': expected true/false", valueStr)
		}
		// Emit type byte
		a.bytecode = append(a.bytecode, byte(TypeBool))
		// Emit value (1 byte)
		a.bytecode = append(a.bytecode, val)

	default:
		return fmt.Errorf("unsupported type '%s' for PUSH", typeName)
	}

	return nil
}

// parseMemoryOperands parses LOAD/STORE instruction operands
func (a *Assembler) parseMemoryOperands(operands []string, lineNum int) error {
	if len(operands) != 1 {
		return fmt.Errorf("memory instruction requires 1 operand (offset), got %d", len(operands))
	}

	offset, err := strconv.ParseUint(operands[0], 0, 16)
	if err != nil {
		return fmt.Errorf("invalid memory offset '%s': %w", operands[0], err)
	}

	if offset > 0xFFFF {
		return fmt.Errorf("memory offset %d exceeds maximum (65535)", offset)
	}

	// Emit offset (2 bytes, little-endian)
	a.bytecode = append(a.bytecode,
		byte(offset&0xFF),
		byte((offset>>8)&0xFF),
	)

	return nil
}

// parseJumpOperands parses JMP/JMPIF/CALL instruction operands
func (a *Assembler) parseJumpOperands(operands []string, lineNum int, isRelative bool) error {
	if len(operands) != 1 {
		return fmt.Errorf("jump instruction requires 1 operand (label or offset), got %d", len(operands))
	}

	target := operands[0]

	// Check if it's a numeric offset or a label
	if offset, err := strconv.ParseInt(target, 0, 32); err == nil {
		// Numeric offset
		a.bytecode = append(a.bytecode,
			byte(offset&0xFF),
			byte((offset>>8)&0xFF),
			byte((offset>>16)&0xFF),
			byte((offset>>24)&0xFF),
		)
	} else {
		// Label reference - add placeholder and record for second pass
		offsetPos := len(a.bytecode)
		a.bytecode = append(a.bytecode, 0, 0, 0, 0) // Placeholder
		a.unresolved = append(a.unresolved, unresolvedRef{
			labelName:      target,
			bytecodeOffset: offsetPos,
			isRelative:     isRelative,
			sourceLine:     lineNum,
		})
	}

	return nil
}

// parseNoOperands validates that no operands are provided
func (a *Assembler) parseNoOperands(operands []string, lineNum int) error {
	if len(operands) > 0 {
		return fmt.Errorf("instruction does not take operands, got %d", len(operands))
	}
	return nil
}

// mnemonicToOpcode converts a mnemonic string to an opcode
func mnemonicToOpcode(mnemonic string) (Opcode, bool) {
	// Build reverse map from OpcodeNames
	for opcode, name := range OpcodeNames {
		if name == mnemonic {
			return opcode, true
		}
	}
	return 0, false
}

// AssembleToFile assembles assembly text and creates a complete bytecode file with header
func AssembleToFile(source string) ([]byte, error) {
	assembler := NewAssembler()
	body, err := assembler.Assemble(source)
	if err != nil {
		return nil, err
	}

	// Create bytecode file with header (entry point at offset 0)
	return CreateBytecode(body, 0), nil
}

// AssembleToBody assembles assembly text and returns just the bytecode body (no header)
func AssembleToBody(source string) ([]byte, error) {
	assembler := NewAssembler()
	return assembler.Assemble(source)
}
