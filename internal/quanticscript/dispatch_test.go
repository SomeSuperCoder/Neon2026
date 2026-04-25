package quanticscript

import (
	"encoding/binary"
	"testing"
)

// buildInstrData builds a raw instruction byte slice with the given type code
// followed by the provided fields concatenated.
func buildInstrData(typeCode byte, fields ...[]byte) []byte {
	data := []byte{typeCode}
	for _, f := range fields {
		data = append(data, f...)
	}
	return data
}

func i64LE(v int64) []byte {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(v))
	return b
}

func u64LE(v uint64) []byte {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, v)
	return b
}

func bytes32(fill byte) []byte {
	b := make([]byte, 32)
	for i := range b {
		b[i] = fill
	}
	return b
}

// ---------------------------------------------------------------------------
// ParseArgs tests
// ---------------------------------------------------------------------------

func TestParseArgs_I64(t *testing.T) {
	def := InstructionDef{
		Code: 0, Name: "TEST", Handler: "h",
		Args: []ArgDef{{Name: "amount", Type: ArgTypeI64, Offset: 1, Length: 8}},
	}
	data := buildInstrData(0, i64LE(-42))
	args, err := ParseArgs(data, def)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	v, ok := args["amount"]
	if !ok {
		t.Fatal("missing 'amount' key")
	}
	got, err := v.AsI64()
	if err != nil {
		t.Fatalf("AsI64: %v", err)
	}
	if got != -42 {
		t.Errorf("expected -42, got %d", got)
	}
}

func TestParseArgs_U64(t *testing.T) {
	def := InstructionDef{
		Code: 0, Name: "TEST", Handler: "h",
		Args: []ArgDef{{Name: "amount", Type: ArgTypeU64, Offset: 1, Length: 8}},
	}
	data := buildInstrData(0, u64LE(1000))
	args, err := ParseArgs(data, def)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got, err := args["amount"].AsU64()
	if err != nil {
		t.Fatalf("AsU64: %v", err)
	}
	if got != 1000 {
		t.Errorf("expected 1000, got %d", got)
	}
}

func TestParseArgs_Bytes(t *testing.T) {
	def := InstructionDef{
		Code: 0, Name: "TEST", Handler: "h",
		Args: []ArgDef{{Name: "owner", Type: ArgTypeBytes, Offset: 1, Length: 32}},
	}
	owner := bytes32(0xAB)
	data := buildInstrData(0, owner)
	args, err := ParseArgs(data, def)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got, err := args["owner"].AsBytes()
	if err != nil {
		t.Fatalf("AsBytes: %v", err)
	}
	if len(got) != 32 || got[0] != 0xAB {
		t.Errorf("unexpected bytes value")
	}
}

func TestParseArgs_Bool(t *testing.T) {
	def := InstructionDef{
		Code: 0, Name: "TEST", Handler: "h",
		Args: []ArgDef{{Name: "flag", Type: ArgTypeBool, Offset: 1, Length: 1}},
	}
	data := buildInstrData(0, []byte{1})
	args, err := ParseArgs(data, def)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got, err := args["flag"].AsBool()
	if err != nil {
		t.Fatalf("AsBool: %v", err)
	}
	if !got {
		t.Error("expected true")
	}
}

func TestParseArgs_BoundsCheck(t *testing.T) {
	def := InstructionDef{
		Code: 0, Name: "TEST", Handler: "h",
		Args: []ArgDef{{Name: "amount", Type: ArgTypeI64, Offset: 1, Length: 8}},
	}
	// Only 4 bytes after type code — too short for i64
	data := []byte{0, 1, 2, 3, 4}
	_, err := ParseArgs(data, def)
	if err == nil {
		t.Fatal("expected bounds-check error, got nil")
	}
}

// ---------------------------------------------------------------------------
// Dispatch tests
// ---------------------------------------------------------------------------

func TestDispatch_KnownCode(t *testing.T) {
	registry := map[int]InstructionDef{
		1: {Code: 1, Name: "TRANSFER", Handler: "handle_transfer",
			Args: []ArgDef{
				{Name: "from", Type: ArgTypeBytes, Offset: 1, Length: 32},
				{Name: "to", Type: ArgTypeBytes, Offset: 33, Length: 32},
				{Name: "amount", Type: ArgTypeI64, Offset: 65, Length: 8},
			}},
	}
	from := bytes32(0x01)
	to := bytes32(0x02)
	data := buildInstrData(1, from, to, i64LE(500))

	def, args, err := Dispatch(data, registry)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if def.Name != "TRANSFER" {
		t.Errorf("expected TRANSFER, got %s", def.Name)
	}
	amt, _ := args["amount"].AsI64()
	if amt != 500 {
		t.Errorf("expected 500, got %d", amt)
	}
}

func TestDispatch_UnknownCode(t *testing.T) {
	registry := map[int]InstructionDef{}
	data := []byte{99}
	_, _, err := Dispatch(data, registry)
	if err == nil {
		t.Fatal("expected error for unknown instruction code")
	}
}

func TestDispatch_TruncatedData(t *testing.T) {
	registry := map[int]InstructionDef{
		0: {Code: 0, Name: "CREATE_ACCOUNT", Handler: "handle_create_account",
			Args: []ArgDef{
				{Name: "owner", Type: ArgTypeBytes, Offset: 1, Length: 32},
				{Name: "balance", Type: ArgTypeI64, Offset: 33, Length: 8},
			}},
	}
	// Only type byte — no args
	data := []byte{0}
	_, _, err := Dispatch(data, registry)
	if err == nil {
		t.Fatal("expected error for truncated data")
	}
}

// ---------------------------------------------------------------------------
// System_Program registry tests
// ---------------------------------------------------------------------------

func TestSystemProgramRegistry_CreateAccount(t *testing.T) {
	owner := bytes32(0xAA)
	data := buildInstrData(0, owner, i64LE(1000))

	def, args, err := Dispatch(data, SystemProgramRegistry)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if def.Name != "CREATE_ACCOUNT" {
		t.Errorf("expected CREATE_ACCOUNT, got %s", def.Name)
	}
	bal, _ := args["balance"].AsI64()
	if bal != 1000 {
		t.Errorf("expected balance 1000, got %d", bal)
	}
}

func TestSystemProgramRegistry_Transfer(t *testing.T) {
	from := bytes32(0x01)
	to := bytes32(0x02)
	data := buildInstrData(1, from, to, i64LE(250))

	def, args, err := Dispatch(data, SystemProgramRegistry)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if def.Name != "TRANSFER" {
		t.Errorf("expected TRANSFER, got %s", def.Name)
	}
	amt, _ := args["amount"].AsI64()
	if amt != 250 {
		t.Errorf("expected 250, got %d", amt)
	}
}

func TestSystemProgramRegistry_AllocateSpace(t *testing.T) {
	account := bytes32(0x03)
	data := buildInstrData(2, account, i64LE(512))

	def, args, err := Dispatch(data, SystemProgramRegistry)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if def.Name != "ALLOCATE_SPACE" {
		t.Errorf("expected ALLOCATE_SPACE, got %s", def.Name)
	}
	extra, _ := args["extra_balance"].AsI64()
	if extra != 512 {
		t.Errorf("expected 512, got %d", extra)
	}
}

// ---------------------------------------------------------------------------
// Token_Program registry tests
// ---------------------------------------------------------------------------

func TestTokenProgramRegistry_InitializeMint(t *testing.T) {
	mintAuth := bytes32(0xAA)
	freezeAuth := bytes32(0xBB)
	// code=0, decimals(1), has_mint_auth(1), mint_auth(32), has_freeze_auth(1), freeze_auth(32)
	data := buildInstrData(0, []byte{6}, []byte{1}, mintAuth, []byte{1}, freezeAuth)

	def, args, err := Dispatch(data, TokenProgramRegistry)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if def.Name != "INITIALIZE_MINT" {
		t.Errorf("expected INITIALIZE_MINT, got %s", def.Name)
	}
	hasMintAuth, _ := args["has_mint_auth"].AsBool()
	if !hasMintAuth {
		t.Error("expected has_mint_auth=true")
	}
	_ = args["decimals"]
}

func TestTokenProgramRegistry_Transfer(t *testing.T) {
	from := bytes32(0x01)
	to := bytes32(0x02)
	data := buildInstrData(2, from, to, u64LE(100))

	def, args, err := Dispatch(data, TokenProgramRegistry)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if def.Name != "TRANSFER" {
		t.Errorf("expected TRANSFER, got %s", def.Name)
	}
	amt, _ := args["amount"].AsU64()
	if amt != 100 {
		t.Errorf("expected 100, got %d", amt)
	}
}

func TestTokenProgramRegistry_AllCodesRegistered(t *testing.T) {
	// All 11 token instruction codes (0-10) must be registered
	for code := 0; code <= 10; code++ {
		if _, ok := TokenProgramRegistry[code]; !ok {
			t.Errorf("token instruction code %d not registered", code)
		}
	}
}

func TestTokenProgramRegistry_UnknownCode(t *testing.T) {
	data := []byte{99}
	_, _, err := Dispatch(data, TokenProgramRegistry)
	if err == nil {
		t.Fatal("expected error for unknown token instruction code")
	}
}
