package quanticscript

// Opcode represents a bytecode instruction opcode
type Opcode uint8

// Stack Operations
const (
	OpPush Opcode = iota // Push value onto stack
	OpPop                // Pop value from stack
	OpDup                // Duplicate top stack value
	OpSwap               // Swap top two stack values
)

// Arithmetic Operations
const (
	OpAdd Opcode = iota + 0x10 // Add two integers
	OpSub                      // Subtract two integers
	OpMul                      // Multiply two integers
	OpDiv                      // Divide two integers
	OpMod                      // Modulo operation
)

// Comparison Operations
const (
	OpEq  Opcode = iota + 0x20 // Equality comparison
	OpLt                       // Less than comparison
	OpGt                       // Greater than comparison
	OpLte                      // Less than or equal
	OpGte                      // Greater than or equal
)

// Logical Operations
const (
	OpAnd Opcode = iota + 0x30 // Logical AND
	OpOr                       // Logical OR
	OpNot                      // Logical NOT
)

// Bitwise Operations
const (
	OpBand Opcode = iota + 0x40 // Bitwise AND
	OpBor                       // Bitwise OR
	OpBxor                      // Bitwise XOR
	OpBnot                      // Bitwise NOT
	OpShl                       // Shift left
	OpShr                       // Shift right
)

// Control Flow
const (
	OpJmp   Opcode = iota + 0x50 // Unconditional jump
	OpJmpIf                      // Conditional jump if true
	OpCall                       // Call function
	OpRet                        // Return from function
)

// Memory Operations
const (
	OpLoad   Opcode = iota + 0x60 // Load from local memory
	OpStore                       // Store to local memory
	OpLoadG                       // Load from global state
	OpStoreG                      // Store to global state
)

// Blockchain Operations
const (
	OpGetFile       Opcode = iota + 0x70 // Get file from FileStore
	OpGetFileMut                         // Get mutable file reference
	OpUpdateFile                         // Update file in FileStore
	OpGetBalance                         // Get file balance
	OpUpdateBalance                      // Update file balance
	OpGetSigner                          // Get transaction signer
	OpHasSigner                          // Check if pubkey is signer
	OpGetInstrData                       // Get instruction data
	OpGetProgramID                       // Get current program ID
)

// Cross-Program Invocation
const (
	OpInvoke    Opcode = iota + 0x80 // Invoke another program
	OpInvokeRet                      // Return from cross-program call
)

// Cryptographic Operations
const (
	OpSha256       Opcode = iota + 0x90 // SHA-256 hash
	OpVerifySig                         // Verify Ed25519 signature
	OpDerivePubKey                      // Derive public key
)

// Query Operations (Finalized Data)
const (
	OpQueryBlock Opcode = iota + 0xA0 // Query finalized block
	OpQueryTx                         // Query finalized transaction
	OpQueryInstr                      // Query finalized instruction
)

// Collection Operations
const (
	OpArrayNew    Opcode = iota + 0xB0 // Create new array
	OpArrayLen                         // Get array length
	OpArrayGet                         // Get array element
	OpArraySet                         // Set array element
	OpArrayPush                        // Push element to array
	OpArrayPop                         // Pop element from array
	OpArrayMap                         // Map function over array
	OpArrayFilter                      // Filter array elements
	OpArrayReduce                      // Reduce array to single value
	OpArraySort                        // Sort array
	OpMapNew                           // Create new map
	OpMapGet                           // Get map value
	OpMapSet                           // Set map value
	OpMapHas                           // Check if map has key
	OpMapDel                           // Delete map key
	OpSetNew                           // Create new set
	OpSetAdd                           // Add element to set
	OpSetHas                           // Check if set has element
	OpSetDel                           // Delete element from set
)

// String Operations
const (
	OpStrConcat    Opcode = 0xC3 // Concatenate strings
	OpStrSubstring Opcode = 0xC4 // Get substring
	OpStrLen       Opcode = 0xC5 // Get string length
	OpStrToBytes   Opcode = 0xC6 // Convert string to bytes
	OpStrFromBytes Opcode = 0xC7 // Convert bytes to string
)

// Math Operations
const (
	OpMathMin Opcode = 0xD0 // Minimum of two values
	OpMathMax Opcode = 0xD1 // Maximum of two values
	OpMathAbs Opcode = 0xD2 // Absolute value
	OpMathPow Opcode = 0xD3 // Power (deterministic integer only)
)

// Conversion Operations
const (
	OpBytesToI64LE Opcode = 0xE0 // Decode TypeBytes (8 bytes) as little-endian i64
)

// Dispatch Operations
const (
	OpDispatch Opcode = 0xF0 // DISPATCH: pop raw instruction bytes, look up registry, push parsed args + handler name
)

// OpcodeNames maps opcodes to their human-readable names
var OpcodeNames = map[Opcode]string{
	// Stack Operations
	OpPush: "PUSH",
	OpPop:  "POP",
	OpDup:  "DUP",
	OpSwap: "SWAP",

	// Arithmetic Operations
	OpAdd: "ADD",
	OpSub: "SUB",
	OpMul: "MUL",
	OpDiv: "DIV",
	OpMod: "MOD",

	// Comparison Operations
	OpEq:  "EQ",
	OpLt:  "LT",
	OpGt:  "GT",
	OpLte: "LTE",
	OpGte: "GTE",

	// Logical Operations
	OpAnd: "AND",
	OpOr:  "OR",
	OpNot: "NOT",

	// Bitwise Operations
	OpBand: "BAND",
	OpBor:  "BOR",
	OpBxor: "BXOR",
	OpBnot: "BNOT",
	OpShl:  "SHL",
	OpShr:  "SHR",

	// Control Flow
	OpJmp:   "JMP",
	OpJmpIf: "JMPIF",
	OpCall:  "CALL",
	OpRet:   "RET",

	// Memory Operations
	OpLoad:   "LOAD",
	OpStore:  "STORE",
	OpLoadG:  "LOADG",
	OpStoreG: "STOREG",

	// Blockchain Operations
	OpGetFile:       "GETFILE",
	OpGetFileMut:    "GETFILEMUT",
	OpUpdateFile:    "UPDATEFILE",
	OpGetBalance:    "GETBALANCE",
	OpUpdateBalance: "UPDATEBALANCE",
	OpGetSigner:     "GETSIGNER",
	OpHasSigner:     "HASSIGNER",
	OpGetInstrData:  "GETINSTRDATA",
	OpGetProgramID:  "GETPROGRAMID",

	// Cross-Program Invocation
	OpInvoke:    "INVOKE",
	OpInvokeRet: "INVOKERET",

	// Cryptographic Operations
	OpSha256:       "SHA256",
	OpVerifySig:    "VERIFYSIG",
	OpDerivePubKey: "DERIVEPUBKEY",

	// Query Operations
	OpQueryBlock: "QUERYBLOCK",
	OpQueryTx:    "QUERYTX",
	OpQueryInstr: "QUERYINSTR",

	// Collection Operations
	OpArrayNew:    "ARRAYNEW",
	OpArrayLen:    "ARRAYLEN",
	OpArrayGet:    "ARRAYGET",
	OpArraySet:    "ARRAYSET",
	OpArrayPush:   "ARRAYPUSH",
	OpArrayPop:    "ARRAYPOP",
	OpArrayMap:    "ARRAYMAP",
	OpArrayFilter: "ARRAYFILTER",
	OpArrayReduce: "ARRAYREDUCE",
	OpArraySort:   "ARRAYSORT",
	OpMapNew:      "MAPNEW",
	OpMapGet:      "MAPGET",
	OpMapSet:      "MAPSET",
	OpMapHas:      "MAPHAS",
	OpMapDel:      "MAPDEL",
	OpSetNew:      "SETNEW",
	OpSetAdd:      "SETADD",
	OpSetHas:      "SETHAS",
	OpSetDel:      "SETDEL",

	// String Operations
	OpStrConcat:    "STRCONCAT",
	OpStrSubstring: "STRSUBSTRING",
	OpStrLen:       "STRLEN",
	OpStrToBytes:   "STRTOBYTES",
	OpStrFromBytes: "STRFROMBYTES",

	// Math Operations
	OpMathMin: "MATHMIN",
	OpMathMax: "MATHMAX",
	OpMathAbs: "MATHABS",
	OpMathPow: "MATHPOW",

	// Conversion Operations
	OpBytesToI64LE: "BYTESTOI64LE",

	// Dispatch Operations
	OpDispatch: "DISPATCH",
}
