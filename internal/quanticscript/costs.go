package quanticscript

// InstructionCost represents the computational cost of a bytecode instruction
type InstructionCost int64

// CostTable maps opcodes to their computational costs
var CostTable = map[Opcode]InstructionCost{
	// Stack Operations (cost: 1)
	OpPush: 1,
	OpPop:  1,
	OpDup:  1,
	OpSwap: 1,

	// Arithmetic Operations (cost: 2-4)
	OpAdd: 2,
	OpSub: 2,
	OpMul: 3,
	OpDiv: 4,
	OpMod: 4,

	// Comparison Operations (cost: 2)
	OpEq:  2,
	OpLt:  2,
	OpGt:  2,
	OpLte: 2,
	OpGte: 2,

	// Logical Operations (cost: 1-2)
	OpAnd: 2,
	OpOr:  2,
	OpNot: 1,

	// Bitwise Operations (cost: 1-2)
	OpBand: 2,
	OpBor:  2,
	OpBxor: 2,
	OpBnot: 1,
	OpShl:  2,
	OpShr:  2,

	// Control Flow (cost: 2-5)
	OpJmp:   2,
	OpJmpIf: 2,
	OpCall:  5,
	OpRet:   2,

	// Memory Operations (cost: 3-15)
	OpLoad:   3,
	OpStore:  3,
	OpLoadG:  10,
	OpStoreG: 15,

	// Blockchain Operations (cost: 10-100)
	OpGetFile:       50,
	OpGetFileMut:    50,
	OpUpdateFile:    100,
	OpGetBalance:    30,
	OpUpdateBalance: 80,
	OpGetSigner:     10,
	OpHasSigner:     15,
	OpGetInstrData:  5,
	OpGetProgramID:  5,

	// Cross-Program Invocation (cost: 10-200)
	OpInvoke:    200,
	OpInvokeRet: 10,

	// Cryptographic Operations (cost: 50-100)
	OpSha256:       50,
	OpVerifySig:    100,
	OpDerivePubKey: 80,

	// Query Operations (cost: 80-100)
	OpQueryBlock: 100,
	OpQueryTx:    80,
	OpQueryInstr: 80,

	// Collection Operations (cost: 2-20)
	OpArrayNew:    5,
	OpArrayLen:    2,
	OpArrayGet:    3,
	OpArraySet:    3,
	OpArrayPush:   3,
	OpArrayPop:    3,
	OpArrayMap:    10, // Base cost, actual cost depends on array size
	OpArrayFilter: 10, // Base cost, actual cost depends on array size
	OpArrayReduce: 10, // Base cost, actual cost depends on array size
	OpArraySort:   20, // Base cost, actual cost depends on array size
	OpMapNew:      5,
	OpMapGet:      3,
	OpMapSet:      4,
	OpMapHas:      3,
	OpMapDel:      4,
	OpSetNew:      5,
	OpSetAdd:      4,
	OpSetHas:      3,
	OpSetDel:      4,

	// String Operations (cost: 2-5)
	OpStrConcat:    3,
	OpStrSubstring: 3,
	OpStrLen:       2,
	OpStrToBytes:   2,
	OpStrFromBytes: 2,

	// Math Operations (cost: 2-10)
	OpMathMin: 2,
	OpMathMax: 2,
	OpMathAbs: 2,
	OpMathPow: 10, // Higher cost due to potential computation

	// Dispatch Operations (base 10 + 2 per arg, charged at runtime)
	OpDispatch: 10,
}

// GetInstructionCost returns the computational cost for a given opcode
func GetInstructionCost(op Opcode) InstructionCost {
	if cost, ok := CostTable[op]; ok {
		return cost
	}
	// Default cost for unknown opcodes
	return 1
}
