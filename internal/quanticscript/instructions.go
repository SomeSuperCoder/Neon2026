package quanticscript

import (
	"encoding/binary"
	"fmt"
)

// ArgType defines the type of an instruction argument.
type ArgType int

const (
	ArgTypeI64   ArgType = iota // Signed 64-bit integer (8 bytes, little-endian)
	ArgTypeU64                  // Unsigned 64-bit integer (8 bytes, little-endian)
	ArgTypeBytes                // Fixed-length byte slice
	ArgTypeBool                 // Boolean (1 byte, 0=false, non-zero=true)
)

// ArgDef describes a single argument within an instruction's byte layout.
type ArgDef struct {
	Name   string  // Argument name (used as map key in parsed result)
	Type   ArgType // Argument type
	Offset int     // Byte offset within instruction data (after the 1-byte type code)
	Length int     // Byte length (required for ArgTypeBytes; 8 for i64/u64; 1 for bool)
}

// InstructionDef describes a complete instruction schema.
type InstructionDef struct {
	Code    int      // Instruction type code (first byte of instruction data)
	Name    string   // Human-readable instruction name
	Args    []ArgDef // Ordered argument definitions
	Handler string   // Assembly label to branch to after dispatch
}

// ParseArgs parses instruction data according to the given InstructionDef.
// It performs full bounds checking and returns a map of arg name → Value.
func ParseArgs(data []byte, def InstructionDef) (map[string]Value, error) {
	result := make(map[string]Value, len(def.Args))
	for _, arg := range def.Args {
		switch arg.Type {
		case ArgTypeI64:
			end := arg.Offset + 8
			if end > len(data) {
				return nil, fmt.Errorf("instruction %s: arg %q: data too short (need %d bytes, have %d)",
					def.Name, arg.Name, end, len(data))
			}
			v := int64(binary.LittleEndian.Uint64(data[arg.Offset:end]))
			result[arg.Name] = NewI64(v)

		case ArgTypeU64:
			end := arg.Offset + 8
			if end > len(data) {
				return nil, fmt.Errorf("instruction %s: arg %q: data too short (need %d bytes, have %d)",
					def.Name, arg.Name, end, len(data))
			}
			v := binary.LittleEndian.Uint64(data[arg.Offset:end])
			result[arg.Name] = NewU64(v)

		case ArgTypeBytes:
			end := arg.Offset + arg.Length
			if end > len(data) {
				return nil, fmt.Errorf("instruction %s: arg %q: data too short (need %d bytes, have %d)",
					def.Name, arg.Name, end, len(data))
			}
			b := make([]byte, arg.Length)
			copy(b, data[arg.Offset:end])
			result[arg.Name] = NewBytes(b)

		case ArgTypeBool:
			end := arg.Offset + 1
			if end > len(data) {
				return nil, fmt.Errorf("instruction %s: arg %q: data too short (need %d bytes, have %d)",
					def.Name, arg.Name, end, len(data))
			}
			result[arg.Name] = NewBool(data[arg.Offset] != 0)

		default:
			return nil, fmt.Errorf("instruction %s: arg %q: unknown ArgType %d", def.Name, arg.Name, arg.Type)
		}
	}
	return result, nil
}

// Dispatch reads the instruction type code from data[0], looks up the registry,
// and delegates to ParseArgs. Returns the matched InstructionDef and parsed args.
func Dispatch(data []byte, registry map[int]InstructionDef) (InstructionDef, map[string]Value, error) {
	if len(data) == 0 {
		return InstructionDef{}, nil, fmt.Errorf("dispatch: empty instruction data")
	}
	code := int(data[0])
	def, ok := registry[code]
	if !ok {
		return InstructionDef{}, nil, fmt.Errorf("dispatch: unknown instruction code %d (0x%04X)", code, code)
	}
	args, err := ParseArgs(data, def)
	if err != nil {
		return InstructionDef{}, nil, err
	}
	return def, args, nil
}

// ---------------------------------------------------------------------------
// System_Program registry
// ---------------------------------------------------------------------------

// SystemProgramRegistry maps instruction type codes to their schemas.
var SystemProgramRegistry = map[int]InstructionDef{
	// CREATE_ACCOUNT: [type(1)] [owner(32)] [balance(8)]
	0: {
		Code: 0, Name: "CREATE_ACCOUNT", Handler: "handle_create_account",
		Args: []ArgDef{
			{Name: "owner", Type: ArgTypeBytes, Offset: 1, Length: 32},
			{Name: "balance", Type: ArgTypeI64, Offset: 33, Length: 8},
		},
	},
	// TRANSFER: [type(1)] [from(32)] [to(32)] [amount(8)]
	1: {
		Code: 1, Name: "TRANSFER", Handler: "handle_transfer",
		Args: []ArgDef{
			{Name: "from", Type: ArgTypeBytes, Offset: 1, Length: 32},
			{Name: "to", Type: ArgTypeBytes, Offset: 33, Length: 32},
			{Name: "amount", Type: ArgTypeI64, Offset: 65, Length: 8},
		},
	},
	// ALLOCATE_SPACE: [type(1)] [account(32)] [extra_balance(8)]
	2: {
		Code: 2, Name: "ALLOCATE_SPACE", Handler: "handle_allocate_space",
		Args: []ArgDef{
			{Name: "account", Type: ArgTypeBytes, Offset: 1, Length: 32},
			{Name: "extra_balance", Type: ArgTypeI64, Offset: 33, Length: 8},
		},
	},
}

// ---------------------------------------------------------------------------
// Token_Program registry
// ---------------------------------------------------------------------------

// TokenProgramRegistry maps instruction type codes to their schemas.
// All 11 token instruction types are registered.
var TokenProgramRegistry = map[int]InstructionDef{
	// INITIALIZE_MINT: [type(1)] [decimals(1)] [has_mint_auth(1)] [mint_auth(32)] [has_freeze_auth(1)] [freeze_auth(32)]
	0: {
		Code: 0, Name: "INITIALIZE_MINT", Handler: "handle_initialize_mint",
		Args: []ArgDef{
			{Name: "decimals", Type: ArgTypeBool, Offset: 1, Length: 1}, // reuse Bool for u8 flag
			{Name: "has_mint_auth", Type: ArgTypeBool, Offset: 2, Length: 1},
			{Name: "mint_auth", Type: ArgTypeBytes, Offset: 3, Length: 32},
			{Name: "has_freeze_auth", Type: ArgTypeBool, Offset: 35, Length: 1},
			{Name: "freeze_auth", Type: ArgTypeBytes, Offset: 36, Length: 32},
		},
	},
	// INITIALIZE_ACCOUNT: [type(1)] [mint(32)] [owner(32)]
	1: {
		Code: 1, Name: "INITIALIZE_ACCOUNT", Handler: "handle_initialize_account",
		Args: []ArgDef{
			{Name: "mint", Type: ArgTypeBytes, Offset: 1, Length: 32},
			{Name: "owner", Type: ArgTypeBytes, Offset: 33, Length: 32},
		},
	},
	// TRANSFER: [type(1)] [from(32)] [to(32)] [amount(8)]
	2: {
		Code: 2, Name: "TRANSFER", Handler: "handle_transfer",
		Args: []ArgDef{
			{Name: "from", Type: ArgTypeBytes, Offset: 1, Length: 32},
			{Name: "to", Type: ArgTypeBytes, Offset: 33, Length: 32},
			{Name: "amount", Type: ArgTypeU64, Offset: 65, Length: 8},
		},
	},
	// MINT_TO: [type(1)] [mint(32)] [dest(32)] [amount(8)]
	3: {
		Code: 3, Name: "MINT_TO", Handler: "handle_mint_to",
		Args: []ArgDef{
			{Name: "mint", Type: ArgTypeBytes, Offset: 1, Length: 32},
			{Name: "dest", Type: ArgTypeBytes, Offset: 33, Length: 32},
			{Name: "amount", Type: ArgTypeU64, Offset: 65, Length: 8},
		},
	},
	// BURN: [type(1)] [account(32)] [amount(8)]
	4: {
		Code: 4, Name: "BURN", Handler: "handle_burn",
		Args: []ArgDef{
			{Name: "account", Type: ArgTypeBytes, Offset: 1, Length: 32},
			{Name: "amount", Type: ArgTypeU64, Offset: 33, Length: 8},
		},
	},
	// CLOSE_ACCOUNT: [type(1)] [account(32)] [dest(32)]
	5: {
		Code: 5, Name: "CLOSE_ACCOUNT", Handler: "handle_close_account",
		Args: []ArgDef{
			{Name: "account", Type: ArgTypeBytes, Offset: 1, Length: 32},
			{Name: "dest", Type: ArgTypeBytes, Offset: 33, Length: 32},
		},
	},
	// FREEZE_ACCOUNT: [type(1)] [account(32)]
	6: {
		Code: 6, Name: "FREEZE_ACCOUNT", Handler: "handle_freeze_account",
		Args: []ArgDef{
			{Name: "account", Type: ArgTypeBytes, Offset: 1, Length: 32},
		},
	},
	// THAW_ACCOUNT: [type(1)] [account(32)]
	7: {
		Code: 7, Name: "THAW_ACCOUNT", Handler: "handle_thaw_account",
		Args: []ArgDef{
			{Name: "account", Type: ArgTypeBytes, Offset: 1, Length: 32},
		},
	},
	// APPROVE: [type(1)] [account(32)] [delegate(32)] [amount(8)]
	8: {
		Code: 8, Name: "APPROVE", Handler: "handle_approve",
		Args: []ArgDef{
			{Name: "account", Type: ArgTypeBytes, Offset: 1, Length: 32},
			{Name: "delegate", Type: ArgTypeBytes, Offset: 33, Length: 32},
			{Name: "amount", Type: ArgTypeU64, Offset: 65, Length: 8},
		},
	},
	// REVOKE: [type(1)] [account(32)]
	9: {
		Code: 9, Name: "REVOKE", Handler: "handle_revoke",
		Args: []ArgDef{
			{Name: "account", Type: ArgTypeBytes, Offset: 1, Length: 32},
		},
	},
	// CREATE_ASSOCIATED_TOKEN_ACCOUNT: [type(1)] [owner(32)] [mint(32)]
	10: {
		Code: 10, Name: "CREATE_ASSOCIATED_TOKEN_ACCOUNT", Handler: "handle_create_associated_token_account",
		Args: []ArgDef{
			{Name: "owner", Type: ArgTypeBytes, Offset: 1, Length: 32},
			{Name: "mint", Type: ArgTypeBytes, Offset: 33, Length: 32},
		},
	},
}
