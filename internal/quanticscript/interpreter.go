package quanticscript

import (
	"fmt"

	"github.com/poh-blockchain/internal/filestore"
)

// BytecodeInterpreter executes QuanticScript bytecode with cost metering
// This implements Requirements 2.1, 2.2, 2.3, 2.4, 2.5
type BytecodeInterpreter struct {
	stack          []Value          // Execution stack for operands
	memory         []Value          // Local memory for variables
	programCounter int              // Current instruction pointer
	computeBudget  int64            // Remaining compute units
	ctx            ExecutionContext // Execution context for blockchain operations
	bytecode       []byte           // Program bytecode
	callStack      []int            // Call stack for function returns
	invokeDepth    int              // Current cross-program invocation depth
}

// NewBytecodeInterpreter creates a new interpreter instance
func NewBytecodeInterpreter(bytecode []byte, ctx ExecutionContext, computeBudget int64) *BytecodeInterpreter {
	return &BytecodeInterpreter{
		stack:          make([]Value, 0, 256),
		memory:         make([]Value, 256), // Pre-allocate local memory
		programCounter: 0,
		computeBudget:  computeBudget,
		ctx:            ctx,
		bytecode:       bytecode,
		callStack:      make([]int, 0, 64),
		invokeDepth:    0,
	}
}

// NewBytecodeInterpreterWithDepth creates a new interpreter instance with specified invocation depth
func NewBytecodeInterpreterWithDepth(bytecode []byte, ctx ExecutionContext, computeBudget int64, invokeDepth int) *BytecodeInterpreter {
	return &BytecodeInterpreter{
		stack:          make([]Value, 0, 256),
		memory:         make([]Value, 256), // Pre-allocate local memory
		programCounter: 0,
		computeBudget:  computeBudget,
		ctx:            ctx,
		bytecode:       bytecode,
		callStack:      make([]int, 0, 64),
		invokeDepth:    invokeDepth,
	}
}

// Execute runs the bytecode until completion or error
func (bi *BytecodeInterpreter) Execute() error {
	for bi.programCounter < len(bi.bytecode) {
		if err := bi.executeInstruction(); err != nil {
			return err
		}
	}
	return nil
}

// executeInstruction executes a single bytecode instruction
func (bi *BytecodeInterpreter) executeInstruction() error {
	if bi.programCounter >= len(bi.bytecode) {
		return fmt.Errorf("program counter out of bounds")
	}

	opcode := Opcode(bi.bytecode[bi.programCounter])
	bi.programCounter++

	// Deduct instruction cost from compute budget
	cost := GetInstructionCost(opcode)
	if err := bi.deductCost(cost); err != nil {
		return err
	}

	// Execute the instruction
	switch opcode {
	// Stack operations
	case OpPush:
		return bi.execPush()
	case OpPop:
		return bi.execPop()
	case OpDup:
		return bi.execDup()
	case OpSwap:
		return bi.execSwap()

	// Memory operations
	case OpLoad:
		return bi.execLoad()
	case OpStore:
		return bi.execStore()

	// Arithmetic operations
	case OpAdd:
		return bi.execAdd()
	case OpSub:
		return bi.execSub()
	case OpMul:
		return bi.execMul()
	case OpDiv:
		return bi.execDiv()
	case OpMod:
		return bi.execMod()

	// Comparison operations
	case OpEq:
		return bi.execEq()
	case OpLt:
		return bi.execLt()
	case OpGt:
		return bi.execGt()
	case OpLte:
		return bi.execLte()
	case OpGte:
		return bi.execGte()

	// Logical operations
	case OpAnd:
		return bi.execAnd()
	case OpOr:
		return bi.execOr()
	case OpNot:
		return bi.execNot()

	// Control flow
	case OpJmp:
		return bi.execJmp()
	case OpJmpIf:
		return bi.execJmpIf()
	case OpCall:
		return bi.execCall()
	case OpRet:
		return bi.execRet()

	// Blockchain operations
	case OpGetFile:
		return bi.execGetFile()
	case OpGetFileMut:
		return bi.execGetFileMut()
	case OpUpdateFile:
		return bi.execUpdateFile()
	case OpGetBalance:
		return bi.execGetBalance()
	case OpUpdateBalance:
		return bi.execUpdateBalance()
	case OpGetSigner:
		return bi.execGetSigner()
	case OpHasSigner:
		return bi.execHasSigner()
	case OpGetInstrData:
		return bi.execGetInstrData()
	case OpGetProgramID:
		return bi.execGetProgramID()

	// Cross-program invocation
	case OpInvoke:
		return bi.execInvoke()
	case OpInvokeRet:
		return bi.execInvokeRet()

	// Cryptographic operations
	case OpSha256:
		return bi.execSha256()
	case OpVerifySig:
		return bi.execVerifySig()
	case OpDerivePubKey:
		return bi.execDerivePubKey()

	// Query operations
	case OpQueryBlock:
		return bi.execQueryBlock()
	case OpQueryTx:
		return bi.execQueryTx()
	case OpQueryInstr:
		return bi.execQueryInstr()

	// Collection operations
	case OpArrayNew:
		return bi.execArrayNew()
	case OpArrayLen:
		return bi.execArrayLen()
	case OpArrayGet:
		return bi.execArrayGet()
	case OpArraySet:
		return bi.execArraySet()
	case OpArrayPush:
		return bi.execArrayPush()
	case OpArrayPop:
		return bi.execArrayPop()
	case OpMapNew:
		return bi.execMapNew()
	case OpMapGet:
		return bi.execMapGet()
	case OpMapSet:
		return bi.execMapSet()
	case OpMapHas:
		return bi.execMapHas()
	case OpMapDel:
		return bi.execMapDel()
	case OpSetNew:
		return bi.execSetNew()
	case OpSetAdd:
		return bi.execSetAdd()
	case OpSetHas:
		return bi.execSetHas()
	case OpSetDel:
		return bi.execSetDel()

	// String operations
	case OpStrConcat:
		return bi.execStrConcat()
	case OpStrSubstring:
		return bi.execStrSubstring()
	case OpStrLen:
		return bi.execStrLen()
	case OpStrToBytes:
		return bi.execStrToBytes()
	case OpStrFromBytes:
		return bi.execStrFromBytes()

	// Math operations
	case OpMathMin:
		return bi.execMathMin()
	case OpMathMax:
		return bi.execMathMax()
	case OpMathAbs:
		return bi.execMathAbs()
	case OpMathPow:
		return bi.execMathPow()

	default:
		return fmt.Errorf("unknown opcode: 0x%02x", opcode)
	}
}

// deductCost deducts cost from compute budget
func (bi *BytecodeInterpreter) deductCost(cost InstructionCost) error {
	bi.computeBudget -= int64(cost)
	if bi.computeBudget < 0 {
		return fmt.Errorf("out of compute budget")
	}
	return nil
}

// Stack operations

// push pushes a value onto the stack
func (bi *BytecodeInterpreter) push(value Value) error {
	bi.stack = append(bi.stack, value)
	return nil
}

// pop pops a value from the stack
func (bi *BytecodeInterpreter) pop() (Value, error) {
	if len(bi.stack) == 0 {
		return Value{}, fmt.Errorf("stack underflow")
	}
	value := bi.stack[len(bi.stack)-1]
	bi.stack = bi.stack[:len(bi.stack)-1]
	return value, nil
}

// peek returns the top stack value without removing it
func (bi *BytecodeInterpreter) peek() (Value, error) {
	if len(bi.stack) == 0 {
		return Value{}, fmt.Errorf("stack underflow")
	}
	return bi.stack[len(bi.stack)-1], nil
}

// execPush pushes a value onto the stack
func (bi *BytecodeInterpreter) execPush() error {
	// Read value type
	if bi.programCounter >= len(bi.bytecode) {
		return fmt.Errorf("unexpected end of bytecode in PUSH")
	}
	valueType := ValueType(bi.bytecode[bi.programCounter])
	bi.programCounter++

	// Read value based on type
	var value Value
	switch valueType {
	case TypeI64:
		if bi.programCounter+8 > len(bi.bytecode) {
			return fmt.Errorf("unexpected end of bytecode reading i64")
		}
		val := int64(bi.bytecode[bi.programCounter]) |
			int64(bi.bytecode[bi.programCounter+1])<<8 |
			int64(bi.bytecode[bi.programCounter+2])<<16 |
			int64(bi.bytecode[bi.programCounter+3])<<24 |
			int64(bi.bytecode[bi.programCounter+4])<<32 |
			int64(bi.bytecode[bi.programCounter+5])<<40 |
			int64(bi.bytecode[bi.programCounter+6])<<48 |
			int64(bi.bytecode[bi.programCounter+7])<<56
		bi.programCounter += 8
		value = NewI64(val)

	case TypeU64:
		if bi.programCounter+8 > len(bi.bytecode) {
			return fmt.Errorf("unexpected end of bytecode reading u64")
		}
		val := uint64(bi.bytecode[bi.programCounter]) |
			uint64(bi.bytecode[bi.programCounter+1])<<8 |
			uint64(bi.bytecode[bi.programCounter+2])<<16 |
			uint64(bi.bytecode[bi.programCounter+3])<<24 |
			uint64(bi.bytecode[bi.programCounter+4])<<32 |
			uint64(bi.bytecode[bi.programCounter+5])<<40 |
			uint64(bi.bytecode[bi.programCounter+6])<<48 |
			uint64(bi.bytecode[bi.programCounter+7])<<56
		bi.programCounter += 8
		value = NewU64(val)

	case TypeBool:
		if bi.programCounter >= len(bi.bytecode) {
			return fmt.Errorf("unexpected end of bytecode reading bool")
		}
		val := bi.bytecode[bi.programCounter] != 0
		bi.programCounter++
		value = NewBool(val)

	case TypeBytes, TypeString:
		// Read length (8 bytes)
		if bi.programCounter+8 > len(bi.bytecode) {
			return fmt.Errorf("unexpected end of bytecode reading bytes/string length")
		}
		length := uint64(bi.bytecode[bi.programCounter]) |
			uint64(bi.bytecode[bi.programCounter+1])<<8 |
			uint64(bi.bytecode[bi.programCounter+2])<<16 |
			uint64(bi.bytecode[bi.programCounter+3])<<24 |
			uint64(bi.bytecode[bi.programCounter+4])<<32 |
			uint64(bi.bytecode[bi.programCounter+5])<<40 |
			uint64(bi.bytecode[bi.programCounter+6])<<48 |
			uint64(bi.bytecode[bi.programCounter+7])<<56
		bi.programCounter += 8

		// Read data
		if bi.programCounter+int(length) > len(bi.bytecode) {
			return fmt.Errorf("unexpected end of bytecode reading bytes/string data")
		}
		data := make([]byte, length)
		copy(data, bi.bytecode[bi.programCounter:bi.programCounter+int(length)])
		bi.programCounter += int(length)

		if valueType == TypeString {
			value = NewString(string(data))
		} else {
			value = NewBytes(data)
		}

	default:
		return fmt.Errorf("unsupported value type for PUSH: %v", valueType)
	}

	return bi.push(value)
}

// execPop pops a value from the stack
func (bi *BytecodeInterpreter) execPop() error {
	_, err := bi.pop()
	return err
}

// execDup duplicates the top stack value
func (bi *BytecodeInterpreter) execDup() error {
	value, err := bi.peek()
	if err != nil {
		return err
	}
	return bi.push(value)
}

// execSwap swaps the top two stack values
func (bi *BytecodeInterpreter) execSwap() error {
	if len(bi.stack) < 2 {
		return fmt.Errorf("stack underflow: need 2 values for SWAP")
	}
	top := bi.stack[len(bi.stack)-1]
	second := bi.stack[len(bi.stack)-2]
	bi.stack[len(bi.stack)-1] = second
	bi.stack[len(bi.stack)-2] = top
	return nil
}

// Memory operations

// execLoad loads a value from local memory onto the stack
func (bi *BytecodeInterpreter) execLoad() error {
	// Read memory offset (2 bytes)
	if bi.programCounter+2 > len(bi.bytecode) {
		return fmt.Errorf("unexpected end of bytecode in LOAD")
	}
	offset := uint16(bi.bytecode[bi.programCounter]) |
		uint16(bi.bytecode[bi.programCounter+1])<<8
	bi.programCounter += 2

	if int(offset) >= len(bi.memory) {
		return fmt.Errorf("memory access out of bounds: offset %d", offset)
	}

	value := bi.memory[offset]
	return bi.push(value)
}

// execStore stores the top stack value to local memory
func (bi *BytecodeInterpreter) execStore() error {
	// Read memory offset (2 bytes)
	if bi.programCounter+2 > len(bi.bytecode) {
		return fmt.Errorf("unexpected end of bytecode in STORE")
	}
	offset := uint16(bi.bytecode[bi.programCounter]) |
		uint16(bi.bytecode[bi.programCounter+1])<<8
	bi.programCounter += 2

	if int(offset) >= len(bi.memory) {
		return fmt.Errorf("memory access out of bounds: offset %d", offset)
	}

	value, err := bi.pop()
	if err != nil {
		return err
	}

	bi.memory[offset] = value
	return nil
}

// GetComputeBudget returns the remaining compute budget
func (bi *BytecodeInterpreter) GetComputeBudget() int64 {
	return bi.computeBudget
}

// Arithmetic operations

// execAdd adds two integers
func (bi *BytecodeInterpreter) execAdd() error {
	b, err := bi.pop()
	if err != nil {
		return err
	}
	a, err := bi.pop()
	if err != nil {
		return err
	}

	// Support both signed and unsigned integers
	if a.Type == TypeI64 && b.Type == TypeI64 {
		aVal, _ := a.AsI64()
		bVal, _ := b.AsI64()
		return bi.push(NewI64(aVal + bVal))
	} else if a.Type == TypeU64 && b.Type == TypeU64 {
		aVal, _ := a.AsU64()
		bVal, _ := b.AsU64()
		return bi.push(NewU64(aVal + bVal))
	}

	return fmt.Errorf("type mismatch in ADD: %v and %v", a.Type, b.Type)
}

// execSub subtracts two integers
func (bi *BytecodeInterpreter) execSub() error {
	b, err := bi.pop()
	if err != nil {
		return err
	}
	a, err := bi.pop()
	if err != nil {
		return err
	}

	if a.Type == TypeI64 && b.Type == TypeI64 {
		aVal, _ := a.AsI64()
		bVal, _ := b.AsI64()
		return bi.push(NewI64(aVal - bVal))
	} else if a.Type == TypeU64 && b.Type == TypeU64 {
		aVal, _ := a.AsU64()
		bVal, _ := b.AsU64()
		return bi.push(NewU64(aVal - bVal))
	}

	return fmt.Errorf("type mismatch in SUB: %v and %v", a.Type, b.Type)
}

// execMul multiplies two integers
func (bi *BytecodeInterpreter) execMul() error {
	b, err := bi.pop()
	if err != nil {
		return err
	}
	a, err := bi.pop()
	if err != nil {
		return err
	}

	if a.Type == TypeI64 && b.Type == TypeI64 {
		aVal, _ := a.AsI64()
		bVal, _ := b.AsI64()
		return bi.push(NewI64(aVal * bVal))
	} else if a.Type == TypeU64 && b.Type == TypeU64 {
		aVal, _ := a.AsU64()
		bVal, _ := b.AsU64()
		return bi.push(NewU64(aVal * bVal))
	}

	return fmt.Errorf("type mismatch in MUL: %v and %v", a.Type, b.Type)
}

// execDiv divides two integers
func (bi *BytecodeInterpreter) execDiv() error {
	b, err := bi.pop()
	if err != nil {
		return err
	}
	a, err := bi.pop()
	if err != nil {
		return err
	}

	if a.Type == TypeI64 && b.Type == TypeI64 {
		aVal, _ := a.AsI64()
		bVal, _ := b.AsI64()
		if bVal == 0 {
			return fmt.Errorf("division by zero")
		}
		return bi.push(NewI64(aVal / bVal))
	} else if a.Type == TypeU64 && b.Type == TypeU64 {
		aVal, _ := a.AsU64()
		bVal, _ := b.AsU64()
		if bVal == 0 {
			return fmt.Errorf("division by zero")
		}
		return bi.push(NewU64(aVal / bVal))
	}

	return fmt.Errorf("type mismatch in DIV: %v and %v", a.Type, b.Type)
}

// execMod computes modulo of two integers
func (bi *BytecodeInterpreter) execMod() error {
	b, err := bi.pop()
	if err != nil {
		return err
	}
	a, err := bi.pop()
	if err != nil {
		return err
	}

	if a.Type == TypeI64 && b.Type == TypeI64 {
		aVal, _ := a.AsI64()
		bVal, _ := b.AsI64()
		if bVal == 0 {
			return fmt.Errorf("modulo by zero")
		}
		return bi.push(NewI64(aVal % bVal))
	} else if a.Type == TypeU64 && b.Type == TypeU64 {
		aVal, _ := a.AsU64()
		bVal, _ := b.AsU64()
		if bVal == 0 {
			return fmt.Errorf("modulo by zero")
		}
		return bi.push(NewU64(aVal % bVal))
	}

	return fmt.Errorf("type mismatch in MOD: %v and %v", a.Type, b.Type)
}

// Comparison operations

// execEq checks equality of two values
func (bi *BytecodeInterpreter) execEq() error {
	b, err := bi.pop()
	if err != nil {
		return err
	}
	a, err := bi.pop()
	if err != nil {
		return err
	}

	if a.Type != b.Type {
		return bi.push(NewBool(false))
	}

	var result bool
	switch a.Type {
	case TypeI64:
		aVal, _ := a.AsI64()
		bVal, _ := b.AsI64()
		result = aVal == bVal
	case TypeU64:
		aVal, _ := a.AsU64()
		bVal, _ := b.AsU64()
		result = aVal == bVal
	case TypeBool:
		aVal, _ := a.AsBool()
		bVal, _ := b.AsBool()
		result = aVal == bVal
	default:
		return fmt.Errorf("unsupported type for EQ: %v", a.Type)
	}

	return bi.push(NewBool(result))
}

// execLt checks if a < b
func (bi *BytecodeInterpreter) execLt() error {
	b, err := bi.pop()
	if err != nil {
		return err
	}
	a, err := bi.pop()
	if err != nil {
		return err
	}

	if a.Type != b.Type {
		return fmt.Errorf("type mismatch in LT: %v and %v", a.Type, b.Type)
	}

	var result bool
	switch a.Type {
	case TypeI64:
		aVal, _ := a.AsI64()
		bVal, _ := b.AsI64()
		result = aVal < bVal
	case TypeU64:
		aVal, _ := a.AsU64()
		bVal, _ := b.AsU64()
		result = aVal < bVal
	default:
		return fmt.Errorf("unsupported type for LT: %v", a.Type)
	}

	return bi.push(NewBool(result))
}

// execGt checks if a > b
func (bi *BytecodeInterpreter) execGt() error {
	b, err := bi.pop()
	if err != nil {
		return err
	}
	a, err := bi.pop()
	if err != nil {
		return err
	}

	if a.Type != b.Type {
		return fmt.Errorf("type mismatch in GT: %v and %v", a.Type, b.Type)
	}

	var result bool
	switch a.Type {
	case TypeI64:
		aVal, _ := a.AsI64()
		bVal, _ := b.AsI64()
		result = aVal > bVal
	case TypeU64:
		aVal, _ := a.AsU64()
		bVal, _ := b.AsU64()
		result = aVal > bVal
	default:
		return fmt.Errorf("unsupported type for GT: %v", a.Type)
	}

	return bi.push(NewBool(result))
}

// execLte checks if a <= b
func (bi *BytecodeInterpreter) execLte() error {
	b, err := bi.pop()
	if err != nil {
		return err
	}
	a, err := bi.pop()
	if err != nil {
		return err
	}

	if a.Type != b.Type {
		return fmt.Errorf("type mismatch in LTE: %v and %v", a.Type, b.Type)
	}

	var result bool
	switch a.Type {
	case TypeI64:
		aVal, _ := a.AsI64()
		bVal, _ := b.AsI64()
		result = aVal <= bVal
	case TypeU64:
		aVal, _ := a.AsU64()
		bVal, _ := b.AsU64()
		result = aVal <= bVal
	default:
		return fmt.Errorf("unsupported type for LTE: %v", a.Type)
	}

	return bi.push(NewBool(result))
}

// execGte checks if a >= b
func (bi *BytecodeInterpreter) execGte() error {
	b, err := bi.pop()
	if err != nil {
		return err
	}
	a, err := bi.pop()
	if err != nil {
		return err
	}

	if a.Type != b.Type {
		return fmt.Errorf("type mismatch in GTE: %v and %v", a.Type, b.Type)
	}

	var result bool
	switch a.Type {
	case TypeI64:
		aVal, _ := a.AsI64()
		bVal, _ := b.AsI64()
		result = aVal >= bVal
	case TypeU64:
		aVal, _ := a.AsU64()
		bVal, _ := b.AsU64()
		result = aVal >= bVal
	default:
		return fmt.Errorf("unsupported type for GTE: %v", a.Type)
	}

	return bi.push(NewBool(result))
}

// Logical operations

// execAnd performs logical AND
func (bi *BytecodeInterpreter) execAnd() error {
	b, err := bi.pop()
	if err != nil {
		return err
	}
	a, err := bi.pop()
	if err != nil {
		return err
	}

	if a.Type != TypeBool || b.Type != TypeBool {
		return fmt.Errorf("type mismatch in AND: expected bool, got %v and %v", a.Type, b.Type)
	}

	aVal, _ := a.AsBool()
	bVal, _ := b.AsBool()
	return bi.push(NewBool(aVal && bVal))
}

// execOr performs logical OR
func (bi *BytecodeInterpreter) execOr() error {
	b, err := bi.pop()
	if err != nil {
		return err
	}
	a, err := bi.pop()
	if err != nil {
		return err
	}

	if a.Type != TypeBool || b.Type != TypeBool {
		return fmt.Errorf("type mismatch in OR: expected bool, got %v and %v", a.Type, b.Type)
	}

	aVal, _ := a.AsBool()
	bVal, _ := b.AsBool()
	return bi.push(NewBool(aVal || bVal))
}

// execNot performs logical NOT
func (bi *BytecodeInterpreter) execNot() error {
	a, err := bi.pop()
	if err != nil {
		return err
	}

	if a.Type != TypeBool {
		return fmt.Errorf("type mismatch in NOT: expected bool, got %v", a.Type)
	}

	aVal, _ := a.AsBool()
	return bi.push(NewBool(!aVal))
}

// Control flow operations

// execJmp performs an unconditional jump
func (bi *BytecodeInterpreter) execJmp() error {
	// Read jump offset (4 bytes, signed)
	if bi.programCounter+4 > len(bi.bytecode) {
		return fmt.Errorf("unexpected end of bytecode in JMP")
	}
	offset := int32(bi.bytecode[bi.programCounter]) |
		int32(bi.bytecode[bi.programCounter+1])<<8 |
		int32(bi.bytecode[bi.programCounter+2])<<16 |
		int32(bi.bytecode[bi.programCounter+3])<<24
	bi.programCounter += 4

	// Calculate new program counter
	newPC := bi.programCounter + int(offset)
	if newPC < 0 || newPC > len(bi.bytecode) {
		return fmt.Errorf("jump target out of bounds: %d", newPC)
	}

	bi.programCounter = newPC
	return nil
}

// execJmpIf performs a conditional jump if top of stack is true
func (bi *BytecodeInterpreter) execJmpIf() error {
	// Read jump offset (4 bytes, signed)
	if bi.programCounter+4 > len(bi.bytecode) {
		return fmt.Errorf("unexpected end of bytecode in JMPIF")
	}
	offset := int32(bi.bytecode[bi.programCounter]) |
		int32(bi.bytecode[bi.programCounter+1])<<8 |
		int32(bi.bytecode[bi.programCounter+2])<<16 |
		int32(bi.bytecode[bi.programCounter+3])<<24
	bi.programCounter += 4

	// Pop condition from stack
	condition, err := bi.pop()
	if err != nil {
		return err
	}

	if condition.Type != TypeBool {
		return fmt.Errorf("JMPIF requires bool condition, got %v", condition.Type)
	}

	condVal, _ := condition.AsBool()
	if condVal {
		// Calculate new program counter
		newPC := bi.programCounter + int(offset)
		if newPC < 0 || newPC > len(bi.bytecode) {
			return fmt.Errorf("jump target out of bounds: %d", newPC)
		}
		bi.programCounter = newPC
	}

	return nil
}

// execCall calls a function at the specified offset
func (bi *BytecodeInterpreter) execCall() error {
	// Read function offset (4 bytes)
	if bi.programCounter+4 > len(bi.bytecode) {
		return fmt.Errorf("unexpected end of bytecode in CALL")
	}
	offset := int32(bi.bytecode[bi.programCounter]) |
		int32(bi.bytecode[bi.programCounter+1])<<8 |
		int32(bi.bytecode[bi.programCounter+2])<<16 |
		int32(bi.bytecode[bi.programCounter+3])<<24
	bi.programCounter += 4

	// Push return address onto call stack
	bi.callStack = append(bi.callStack, bi.programCounter)

	// Check call stack depth to prevent infinite recursion
	if len(bi.callStack) > 64 {
		return fmt.Errorf("call stack overflow: maximum depth exceeded")
	}

	// Calculate new program counter
	newPC := int(offset)
	if newPC < 0 || newPC >= len(bi.bytecode) {
		return fmt.Errorf("call target out of bounds: %d", newPC)
	}

	bi.programCounter = newPC
	return nil
}

// execRet returns from a function call
func (bi *BytecodeInterpreter) execRet() error {
	if len(bi.callStack) == 0 {
		// No return address, end execution
		bi.programCounter = len(bi.bytecode)
		return nil
	}

	// Pop return address from call stack
	returnAddr := bi.callStack[len(bi.callStack)-1]
	bi.callStack = bi.callStack[:len(bi.callStack)-1]

	bi.programCounter = returnAddr
	return nil
}

// Blockchain operations

// execGetFile gets a file from FileStore (read-only)
func (bi *BytecodeInterpreter) execGetFile() error {
	// Pop file ID from stack (32 bytes)
	fileIDValue, err := bi.pop()
	if err != nil {
		return err
	}

	if fileIDValue.Type != TypeFileID {
		return fmt.Errorf("GETFILE requires FileID, got %v", fileIDValue.Type)
	}

	fileIDBytes, _ := fileIDValue.AsBytes()
	fileID, err := filestore.FileIDFromBytes(fileIDBytes)
	if err != nil {
		return fmt.Errorf("invalid FileID: %w", err)
	}

	// Get file from context
	file, err := bi.ctx.GetFile(fileID)
	if err != nil {
		return fmt.Errorf("failed to get file: %w", err)
	}

	// Push file data onto stack as bytes
	return bi.push(NewBytes(file.Data))
}

// execGetFileMut gets a mutable file reference from FileStore
func (bi *BytecodeInterpreter) execGetFileMut() error {
	// Pop file ID from stack (32 bytes)
	fileIDValue, err := bi.pop()
	if err != nil {
		return err
	}

	if fileIDValue.Type != TypeFileID {
		return fmt.Errorf("GETFILEMUT requires FileID, got %v", fileIDValue.Type)
	}

	fileIDBytes, _ := fileIDValue.AsBytes()
	fileID, err := filestore.FileIDFromBytes(fileIDBytes)
	if err != nil {
		return fmt.Errorf("invalid FileID: %w", err)
	}

	// Get file from context (with write permission)
	file, err := bi.ctx.GetFileMut(fileID)
	if err != nil {
		return fmt.Errorf("failed to get mutable file: %w", err)
	}

	// Push file data onto stack as bytes
	return bi.push(NewBytes(file.Data))
}

// execUpdateFile updates a file in FileStore
func (bi *BytecodeInterpreter) execUpdateFile() error {
	// Pop new data from stack
	newDataValue, err := bi.pop()
	if err != nil {
		return err
	}

	if newDataValue.Type != TypeBytes {
		return fmt.Errorf("UPDATEFILE requires bytes for data, got %v", newDataValue.Type)
	}

	// Pop file ID from stack
	fileIDValue, err := bi.pop()
	if err != nil {
		return err
	}

	if fileIDValue.Type != TypeFileID {
		return fmt.Errorf("UPDATEFILE requires FileID, got %v", fileIDValue.Type)
	}

	fileIDBytes, _ := fileIDValue.AsBytes()
	fileID, err := filestore.FileIDFromBytes(fileIDBytes)
	if err != nil {
		return fmt.Errorf("invalid FileID: %w", err)
	}

	newData, _ := newDataValue.AsBytes()

	// Get the file first
	file, err := bi.ctx.GetFileMut(fileID)
	if err != nil {
		return fmt.Errorf("failed to get file for update: %w", err)
	}

	// Update file data
	file.Data = newData

	// Update file in context
	if err := bi.ctx.UpdateFile(file); err != nil {
		return fmt.Errorf("failed to update file: %w", err)
	}

	return nil
}

// execGetBalance gets a file's balance
func (bi *BytecodeInterpreter) execGetBalance() error {
	// Pop file ID from stack
	fileIDValue, err := bi.pop()
	if err != nil {
		return err
	}

	if fileIDValue.Type != TypeFileID {
		return fmt.Errorf("GETBALANCE requires FileID, got %v", fileIDValue.Type)
	}

	fileIDBytes, _ := fileIDValue.AsBytes()
	fileID, err := filestore.FileIDFromBytes(fileIDBytes)
	if err != nil {
		return fmt.Errorf("invalid FileID: %w", err)
	}

	// Get balance from context
	balance, err := bi.ctx.GetFileBalance(fileID)
	if err != nil {
		return fmt.Errorf("failed to get balance: %w", err)
	}

	// Push balance onto stack as i64
	return bi.push(NewI64(balance))
}

// execUpdateBalance updates a file's balance
func (bi *BytecodeInterpreter) execUpdateBalance() error {
	// Pop delta from stack
	deltaValue, err := bi.pop()
	if err != nil {
		return err
	}

	if deltaValue.Type != TypeI64 {
		return fmt.Errorf("UPDATEBALANCE requires i64 for delta, got %v", deltaValue.Type)
	}

	// Pop file ID from stack
	fileIDValue, err := bi.pop()
	if err != nil {
		return err
	}

	if fileIDValue.Type != TypeFileID {
		return fmt.Errorf("UPDATEBALANCE requires FileID, got %v", fileIDValue.Type)
	}

	fileIDBytes, _ := fileIDValue.AsBytes()
	fileID, err := filestore.FileIDFromBytes(fileIDBytes)
	if err != nil {
		return fmt.Errorf("invalid FileID: %w", err)
	}

	delta, _ := deltaValue.AsI64()

	// Update balance in context
	if err := bi.ctx.UpdateFileBalance(fileID, delta); err != nil {
		return fmt.Errorf("failed to update balance: %w", err)
	}

	return nil
}

// execGetSigner gets a transaction signer by index
func (bi *BytecodeInterpreter) execGetSigner() error {
	// Pop index from stack
	indexValue, err := bi.pop()
	if err != nil {
		return err
	}

	if indexValue.Type != TypeU64 {
		return fmt.Errorf("GETSIGNER requires u64 for index, got %v", indexValue.Type)
	}

	index, _ := indexValue.AsU64()

	// Get signers from context
	signers := bi.ctx.GetSigners()

	// Check bounds
	if int(index) >= len(signers) {
		return fmt.Errorf("signer index out of bounds: %d", index)
	}

	// Get signer public key
	signer := signers[index]

	// Push public key onto stack as bytes (TypePublicKey)
	value := Value{
		Type: TypePublicKey,
		Data: signer[:],
	}
	return bi.push(value)
}

// execHasSigner checks if a public key is a signer
func (bi *BytecodeInterpreter) execHasSigner() error {
	// Pop public key from stack
	pubkeyValue, err := bi.pop()
	if err != nil {
		return err
	}

	if pubkeyValue.Type != TypePublicKey {
		return fmt.Errorf("HASSIGNER requires PublicKey, got %v", pubkeyValue.Type)
	}

	pubkeyBytes, _ := pubkeyValue.AsBytes()

	// Convert to transaction.PublicKey
	var pubkey [32]byte
	copy(pubkey[:], pubkeyBytes)

	// Check if signer exists
	hasSigner := bi.ctx.HasSigner(pubkey)

	// Push result onto stack
	return bi.push(NewBool(hasSigner))
}

// execGetInstrData gets the instruction data
func (bi *BytecodeInterpreter) execGetInstrData() error {
	// Get instruction data from context
	data := bi.ctx.GetInstructionData()

	// Push data onto stack as bytes
	return bi.push(NewBytes(data))
}

// execGetProgramID gets the current program ID
func (bi *BytecodeInterpreter) execGetProgramID() error {
	// Get program ID from context
	programID := bi.ctx.GetProgramID()

	// Push program ID onto stack as FileID (bytes)
	value := Value{
		Type: TypeFileID,
		Data: programID[:],
	}
	return bi.push(value)
}

// Cryptographic operations

// execSha256 computes SHA-256 hash of input data
func (bi *BytecodeInterpreter) execSha256() error {
	// Pop data from stack
	dataValue, err := bi.pop()
	if err != nil {
		return err
	}

	if dataValue.Type != TypeBytes {
		return fmt.Errorf("SHA256 requires bytes, got %v", dataValue.Type)
	}

	data, _ := dataValue.AsBytes()

	// Compute SHA-256 hash
	hash := sha256Hash(data)

	// Push hash onto stack as bytes
	return bi.push(NewBytes(hash))
}

// execVerifySig verifies an Ed25519 signature
func (bi *BytecodeInterpreter) execVerifySig() error {
	// Pop signature from stack (64 bytes)
	sigValue, err := bi.pop()
	if err != nil {
		return err
	}

	if sigValue.Type != TypeBytes {
		return fmt.Errorf("VERIFYSIG requires bytes for signature, got %v", sigValue.Type)
	}

	// Pop message from stack
	msgValue, err := bi.pop()
	if err != nil {
		return err
	}

	if msgValue.Type != TypeBytes {
		return fmt.Errorf("VERIFYSIG requires bytes for message, got %v", msgValue.Type)
	}

	// Pop public key from stack (32 bytes)
	pubkeyValue, err := bi.pop()
	if err != nil {
		return err
	}

	if pubkeyValue.Type != TypePublicKey {
		return fmt.Errorf("VERIFYSIG requires PublicKey, got %v", pubkeyValue.Type)
	}

	pubkeyBytes, _ := pubkeyValue.AsBytes()
	msgBytes, _ := msgValue.AsBytes()
	sigBytes, _ := sigValue.AsBytes()

	// Verify signature
	valid := verifySignature(pubkeyBytes, msgBytes, sigBytes)

	// Push result onto stack
	return bi.push(NewBool(valid))
}

// execDerivePubKey derives a public key from a seed
func (bi *BytecodeInterpreter) execDerivePubKey() error {
	// Pop seed from stack (32 bytes)
	seedValue, err := bi.pop()
	if err != nil {
		return err
	}

	if seedValue.Type != TypeBytes {
		return fmt.Errorf("DERIVEPUBKEY requires bytes for seed, got %v", seedValue.Type)
	}

	seedBytes, _ := seedValue.AsBytes()

	// Derive public key
	pubkey, err := derivePublicKey(seedBytes)
	if err != nil {
		return fmt.Errorf("failed to derive public key: %w", err)
	}

	// Push public key onto stack
	return bi.push(NewPublicKey(pubkey))
}

// Query operations for finalized data

// execQueryBlock queries a finalized block by hash
func (bi *BytecodeInterpreter) execQueryBlock() error {
	// Pop block hash from stack (32 bytes)
	hashValue, err := bi.pop()
	if err != nil {
		return err
	}

	if hashValue.Type != TypeBytes {
		return fmt.Errorf("QUERYBLOCK requires bytes for block hash, got %v", hashValue.Type)
	}

	hashBytes, _ := hashValue.AsBytes()

	// Query block from context
	blockData, err := bi.ctx.QueryBlock(hashBytes)
	if err != nil {
		// Return null (empty bytes) if block not found or not finalized
		return bi.push(NewBytes(nil))
	}

	// Push block data onto stack as bytes
	return bi.push(NewBytes(blockData))
}

// execQueryTx queries a finalized transaction by ID
func (bi *BytecodeInterpreter) execQueryTx() error {
	// Pop transaction ID from stack (32 bytes)
	txIDValue, err := bi.pop()
	if err != nil {
		return err
	}

	if txIDValue.Type != TypeTxID {
		return fmt.Errorf("QUERYTX requires TxID, got %v", txIDValue.Type)
	}

	txIDBytes, _ := txIDValue.AsBytes()

	// Convert to TxID
	var txID [32]byte
	copy(txID[:], txIDBytes)

	// Query transaction from context
	txData, err := bi.ctx.QueryTransaction(txID)
	if err != nil {
		// Return null (empty bytes) if transaction not found or not finalized
		return bi.push(NewBytes(nil))
	}

	// Push transaction data onto stack as bytes
	return bi.push(NewBytes(txData))
}

// execQueryInstr queries a finalized instruction by reference
func (bi *BytecodeInterpreter) execQueryInstr() error {
	// Pop instruction index from stack
	indexValue, err := bi.pop()
	if err != nil {
		return err
	}

	if indexValue.Type != TypeU32 {
		return fmt.Errorf("QUERYINSTR requires u32 for instruction index, got %v", indexValue.Type)
	}

	// Pop transaction ID from stack (32 bytes)
	txIDValue, err := bi.pop()
	if err != nil {
		return err
	}

	if txIDValue.Type != TypeTxID {
		return fmt.Errorf("QUERYINSTR requires TxID, got %v", txIDValue.Type)
	}

	txIDBytes, _ := txIDValue.AsBytes()
	index, _ := indexValue.AsU32()

	// Convert to TxID
	var txID [32]byte
	copy(txID[:], txIDBytes)

	// Query instruction from context
	instrData, err := bi.ctx.QueryInstruction(txID, index)
	if err != nil {
		// Return null (empty bytes) if instruction not found or not finalized
		return bi.push(NewBytes(nil))
	}

	// Push instruction data onto stack as bytes
	return bi.push(NewBytes(instrData))
}

// Cross-program invocation operations

// MaxInvokeDepth is the maximum allowed cross-program invocation depth
const MaxInvokeDepth = 4

// execInvoke invokes another program with the given data
// Stack: [programID (FileID), invokeData (bytes), computeBudget (i64)] -> [result (bytes)]
// This implements Requirements 9.1, 9.2, 9.3, 9.4
func (bi *BytecodeInterpreter) execInvoke() error {
	// Check invocation depth limit (Requirement 9.4)
	if bi.invokeDepth >= MaxInvokeDepth {
		return fmt.Errorf("maximum cross-program invocation depth exceeded: %d", MaxInvokeDepth)
	}

	// Pop compute budget from stack
	budgetValue, err := bi.pop()
	if err != nil {
		return err
	}

	if budgetValue.Type != TypeI64 {
		return fmt.Errorf("INVOKE requires i64 for compute budget, got %v", budgetValue.Type)
	}

	budget, _ := budgetValue.AsI64()
	if budget <= 0 {
		return fmt.Errorf("INVOKE requires positive compute budget, got %d", budget)
	}

	// Pop invocation data from stack
	invokeDataValue, err := bi.pop()
	if err != nil {
		return err
	}

	if invokeDataValue.Type != TypeBytes {
		return fmt.Errorf("INVOKE requires bytes for invocation data, got %v", invokeDataValue.Type)
	}

	invokeData, _ := invokeDataValue.AsBytes()

	// Pop program ID from stack
	programIDValue, err := bi.pop()
	if err != nil {
		return err
	}

	if programIDValue.Type != TypeFileID {
		return fmt.Errorf("INVOKE requires FileID for program ID, got %v", programIDValue.Type)
	}

	programIDBytes, _ := programIDValue.AsBytes()
	programID, err := filestore.FileIDFromBytes(programIDBytes)
	if err != nil {
		return fmt.Errorf("invalid program ID: %w", err)
	}

	// Validate that the target program is in the declared program list (Requirement 9.2)
	declaredPrograms := bi.ctx.GetDeclaredPrograms()
	isProgramDeclared := false
	for _, declaredProgramID := range declaredPrograms {
		if declaredProgramID == programID {
			isProgramDeclared = true
			break
		}
	}

	if !isProgramDeclared {
		return fmt.Errorf("program %s not in declared program list", programID.String())
	}

	// Deduct the budget from the current interpreter's budget (Requirement 9.5)
	if bi.computeBudget < budget {
		return fmt.Errorf("insufficient compute budget for invocation: have %d, need %d", bi.computeBudget, budget)
	}
	bi.computeBudget -= budget

	// Invoke the program through the execution context (Requirement 9.1, 9.3)
	resultData, err := bi.ctx.InvokeProgram(programID, invokeData, budget, bi.invokeDepth+1)
	if err != nil {
		// Invocation failed, rollback is handled by the context (Requirement 9.5)
		return fmt.Errorf("cross-program invocation failed: %w", err)
	}

	// Push result data onto stack
	return bi.push(NewBytes(resultData))
}

// execInvokeRet returns from a cross-program invocation
// This is primarily used for explicit returns in invoked programs
// Stack: [result (bytes)] -> (returns from invocation)
func (bi *BytecodeInterpreter) execInvokeRet() error {
	// Pop result from stack
	resultValue, err := bi.pop()
	if err != nil {
		return err
	}

	if resultValue.Type != TypeBytes {
		return fmt.Errorf("INVOKERET requires bytes for result, got %v", resultValue.Type)
	}

	// Push result back onto stack for the caller
	if err := bi.push(resultValue); err != nil {
		return err
	}

	// End execution by setting program counter to end
	bi.programCounter = len(bi.bytecode)
	return nil
}

// GetInvokeDepth returns the current cross-program invocation depth
func (bi *BytecodeInterpreter) GetInvokeDepth() int {
	return bi.invokeDepth
}
