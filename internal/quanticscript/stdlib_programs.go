package quanticscript

import (
	"encoding/binary"
	"fmt"
)

// stdlib_programs.go provides serialization helpers and utilities for
// System_Program and Token_Program development in QuanticScript.
// This implements Requirements 3.1 and 3.2.

// FileID and PublicKey type conversion helpers

// NewFileIDFromBytes creates a FileID value from a byte slice
// Returns an error if the byte slice is not exactly 32 bytes
func NewFileIDFromBytes(data []byte) (Value, error) {
	if len(data) != 32 {
		return Value{}, fmt.Errorf("FileID must be exactly 32 bytes, got %d", len(data))
	}
	// Make a copy to avoid external modifications
	idCopy := make([]byte, 32)
	copy(idCopy, data)
	return NewFileID(idCopy), nil
}

// FileIDToBytes extracts the byte slice from a FileID value
func FileIDToBytes(v Value) ([]byte, error) {
	if v.Type != TypeFileID {
		return nil, fmt.Errorf("expected FileID, got %v", v.Type)
	}
	data, ok := v.Data.([]byte)
	if !ok {
		return nil, fmt.Errorf("FileID data is not a byte slice")
	}
	if len(data) != 32 {
		return nil, fmt.Errorf("FileID must be exactly 32 bytes, got %d", len(data))
	}
	return data, nil
}

// NewPublicKeyFromBytes creates a PublicKey value from a byte slice
// Returns an error if the byte slice is not exactly 32 bytes
func NewPublicKeyFromBytes(data []byte) (Value, error) {
	if len(data) != 32 {
		return Value{}, fmt.Errorf("PublicKey must be exactly 32 bytes, got %d", len(data))
	}
	// Make a copy to avoid external modifications
	keyCopy := make([]byte, 32)
	copy(keyCopy, data)
	return NewPublicKey(keyCopy), nil
}

// PublicKeyToBytes extracts the byte slice from a PublicKey value
func PublicKeyToBytes(v Value) ([]byte, error) {
	if v.Type != TypePublicKey {
		return nil, fmt.Errorf("expected PublicKey, got %v", v.Type)
	}
	data, ok := v.Data.([]byte)
	if !ok {
		return nil, fmt.Errorf("PublicKey data is not a byte slice")
	}
	if len(data) != 32 {
		return nil, fmt.Errorf("PublicKey must be exactly 32 bytes, got %d", len(data))
	}
	return data, nil
}

// Instruction data parsing utilities

// ParseInstructionU8 parses a u8 value from instruction data at the given offset
func ParseInstructionU8(data []byte, offset int) (uint8, int, error) {
	if offset < 0 || offset >= len(data) {
		return 0, offset, fmt.Errorf("offset %d out of bounds for data length %d", offset, len(data))
	}
	return data[offset], offset + 1, nil
}

// ParseInstructionU64 parses a u64 value (little-endian) from instruction data at the given offset
func ParseInstructionU64(data []byte, offset int) (uint64, int, error) {
	if offset < 0 || offset+8 > len(data) {
		return 0, offset, fmt.Errorf("insufficient data for u64 at offset %d (need 8 bytes, have %d)", offset, len(data)-offset)
	}
	value := binary.LittleEndian.Uint64(data[offset : offset+8])
	return value, offset + 8, nil
}

// ParseInstructionI64 parses an i64 value (little-endian) from instruction data at the given offset
func ParseInstructionI64(data []byte, offset int) (int64, int, error) {
	if offset < 0 || offset+8 > len(data) {
		return 0, offset, fmt.Errorf("insufficient data for i64 at offset %d (need 8 bytes, have %d)", offset, len(data)-offset)
	}
	value := int64(binary.LittleEndian.Uint64(data[offset : offset+8]))
	return value, offset + 8, nil
}

// ParseInstructionFileID parses a FileID (32 bytes) from instruction data at the given offset
func ParseInstructionFileID(data []byte, offset int) ([]byte, int, error) {
	if offset < 0 || offset+32 > len(data) {
		return nil, offset, fmt.Errorf("insufficient data for FileID at offset %d (need 32 bytes, have %d)", offset, len(data)-offset)
	}
	fileID := make([]byte, 32)
	copy(fileID, data[offset:offset+32])
	return fileID, offset + 32, nil
}

// ParseInstructionPublicKey parses a PublicKey (32 bytes) from instruction data at the given offset
func ParseInstructionPublicKey(data []byte, offset int) ([]byte, int, error) {
	if offset < 0 || offset+32 > len(data) {
		return nil, offset, fmt.Errorf("insufficient data for PublicKey at offset %d (need 32 bytes, have %d)", offset, len(data)-offset)
	}
	pubKey := make([]byte, 32)
	copy(pubKey, data[offset:offset+32])
	return pubKey, offset + 32, nil
}

// ParseInstructionBool parses a boolean value from instruction data at the given offset
func ParseInstructionBool(data []byte, offset int) (bool, int, error) {
	if offset < 0 || offset >= len(data) {
		return false, offset, fmt.Errorf("offset %d out of bounds for data length %d", offset, len(data))
	}
	return data[offset] != 0, offset + 1, nil
}

// ParseInstructionOptionalPublicKey parses an optional PublicKey from instruction data
// Format: 1 byte flag (0=null, 1=present) + 32 bytes if present
func ParseInstructionOptionalPublicKey(data []byte, offset int) ([]byte, bool, int, error) {
	if offset < 0 || offset >= len(data) {
		return nil, false, offset, fmt.Errorf("offset %d out of bounds for data length %d", offset, len(data))
	}

	flag := data[offset]
	offset++

	if flag == 0 {
		// Null value
		return nil, false, offset, nil
	}

	// Present value
	if offset+32 > len(data) {
		return nil, false, offset, fmt.Errorf("insufficient data for PublicKey at offset %d (need 32 bytes, have %d)", offset, len(data)-offset)
	}

	pubKey := make([]byte, 32)
	copy(pubKey, data[offset:offset+32])
	return pubKey, true, offset + 32, nil
}

// MintAccount serialization functions

// SerializeMintAccount serializes a MintAccount to bytes
// Format:
//   - Bytes 0-7: supply (little-endian u64)
//   - Byte 8: decimals (u8)
//   - Byte 9: mint_authority flag (0=null, 1=present)
//   - Bytes 10-41: mint_authority pubkey (if present)
//   - Byte 42 or 10: freeze_authority flag (0=null, 1=present)
//   - Bytes 43-74 or 11-42: freeze_authority pubkey (if present)
func SerializeMintAccount(supply uint64, decimals uint8, mintAuthority []byte, freezeAuthority []byte) ([]byte, error) {
	// Validate inputs
	if mintAuthority != nil && len(mintAuthority) != 32 {
		return nil, fmt.Errorf("mint_authority must be 32 bytes or nil, got %d", len(mintAuthority))
	}
	if freezeAuthority != nil && len(freezeAuthority) != 32 {
		return nil, fmt.Errorf("freeze_authority must be 32 bytes or nil, got %d", len(freezeAuthority))
	}

	// Calculate size
	size := 9 // supply (8) + decimals (1)
	size += 1 // mint_authority flag
	if mintAuthority != nil {
		size += 32
	}
	size += 1 // freeze_authority flag
	if freezeAuthority != nil {
		size += 32
	}

	data := make([]byte, size)
	offset := 0

	// Write supply
	binary.LittleEndian.PutUint64(data[offset:offset+8], supply)
	offset += 8

	// Write decimals
	data[offset] = decimals
	offset++

	// Write mint_authority
	if mintAuthority != nil {
		data[offset] = 1
		offset++
		copy(data[offset:offset+32], mintAuthority)
		offset += 32
	} else {
		data[offset] = 0
		offset++
	}

	// Write freeze_authority
	if freezeAuthority != nil {
		data[offset] = 1
		offset++
		copy(data[offset:offset+32], freezeAuthority)
		offset += 32
	} else {
		data[offset] = 0
		offset++
	}

	return data, nil
}

// DeserializeMintAccount deserializes a MintAccount from bytes
// Returns: supply, decimals, mintAuthority, hasMintAuthority, freezeAuthority, hasFreezeAuthority, error
func DeserializeMintAccount(data []byte) (uint64, uint8, []byte, bool, []byte, bool, error) {
	if len(data) < 10 {
		return 0, 0, nil, false, nil, false, fmt.Errorf("MintAccount data too short: need at least 10 bytes, got %d", len(data))
	}

	offset := 0

	// Read supply
	supply := binary.LittleEndian.Uint64(data[offset : offset+8])
	offset += 8

	// Read decimals
	decimals := data[offset]
	offset++

	// Read mint_authority
	var mintAuthority []byte
	hasMintAuthority := false
	if offset >= len(data) {
		return 0, 0, nil, false, nil, false, fmt.Errorf("MintAccount data truncated at mint_authority flag")
	}
	mintAuthorityFlag := data[offset]
	offset++

	if mintAuthorityFlag == 1 {
		if offset+32 > len(data) {
			return 0, 0, nil, false, nil, false, fmt.Errorf("MintAccount data truncated at mint_authority")
		}
		mintAuthority = make([]byte, 32)
		copy(mintAuthority, data[offset:offset+32])
		hasMintAuthority = true
		offset += 32
	}

	// Read freeze_authority
	var freezeAuthority []byte
	hasFreezeAuthority := false
	if offset >= len(data) {
		return 0, 0, nil, false, nil, false, fmt.Errorf("MintAccount data truncated at freeze_authority flag")
	}
	freezeAuthorityFlag := data[offset]
	offset++

	if freezeAuthorityFlag == 1 {
		if offset+32 > len(data) {
			return 0, 0, nil, false, nil, false, fmt.Errorf("MintAccount data truncated at freeze_authority")
		}
		freezeAuthority = make([]byte, 32)
		copy(freezeAuthority, data[offset:offset+32])
		hasFreezeAuthority = true
		offset += 32
	}

	return supply, decimals, mintAuthority, hasMintAuthority, freezeAuthority, hasFreezeAuthority, nil
}

// TokenAccount serialization functions

// SerializeTokenAccount serializes a TokenAccount to bytes
// Format:
//   - Bytes 0-31: mint FileID
//   - Bytes 32-63: owner FileID
//   - Bytes 64-71: token_balance (little-endian u64)
//   - Byte 72: delegate flag (0=null, 1=present)
//   - Bytes 73-104: delegate pubkey (if present)
//   - Bytes 105-112 or 73-80: delegated_amount (little-endian u64)
//   - Byte 113 or 81: frozen (0=false, 1=true)
func SerializeTokenAccount(mint []byte, owner []byte, tokenBalance uint64, delegate []byte, delegatedAmount uint64, frozen bool) ([]byte, error) {
	// Validate inputs
	if len(mint) != 32 {
		return nil, fmt.Errorf("mint must be 32 bytes, got %d", len(mint))
	}
	if len(owner) != 32 {
		return nil, fmt.Errorf("owner must be 32 bytes, got %d", len(owner))
	}
	if delegate != nil && len(delegate) != 32 {
		return nil, fmt.Errorf("delegate must be 32 bytes or nil, got %d", len(delegate))
	}

	// Calculate size
	size := 32 + 32 + 8 + 1 // mint + owner + token_balance + delegate flag
	if delegate != nil {
		size += 32
	}
	size += 8 + 1 // delegated_amount + frozen

	data := make([]byte, size)
	offset := 0

	// Write mint
	copy(data[offset:offset+32], mint)
	offset += 32

	// Write owner
	copy(data[offset:offset+32], owner)
	offset += 32

	// Write token_balance
	binary.LittleEndian.PutUint64(data[offset:offset+8], tokenBalance)
	offset += 8

	// Write delegate
	if delegate != nil {
		data[offset] = 1
		offset++
		copy(data[offset:offset+32], delegate)
		offset += 32
	} else {
		data[offset] = 0
		offset++
	}

	// Write delegated_amount
	binary.LittleEndian.PutUint64(data[offset:offset+8], delegatedAmount)
	offset += 8

	// Write frozen
	if frozen {
		data[offset] = 1
	} else {
		data[offset] = 0
	}

	return data, nil
}

// DeserializeTokenAccount deserializes a TokenAccount from bytes
// Returns: mint, owner, tokenBalance, delegate, hasDelegate, delegatedAmount, frozen, error
func DeserializeTokenAccount(data []byte) ([]byte, []byte, uint64, []byte, bool, uint64, bool, error) {
	if len(data) < 82 {
		return nil, nil, 0, nil, false, 0, false, fmt.Errorf("TokenAccount data too short: need at least 82 bytes, got %d", len(data))
	}

	offset := 0

	// Read mint
	mint := make([]byte, 32)
	copy(mint, data[offset:offset+32])
	offset += 32

	// Read owner
	owner := make([]byte, 32)
	copy(owner, data[offset:offset+32])
	offset += 32

	// Read token_balance
	tokenBalance := binary.LittleEndian.Uint64(data[offset : offset+8])
	offset += 8

	// Read delegate
	var delegate []byte
	hasDelegate := false
	if offset >= len(data) {
		return nil, nil, 0, nil, false, 0, false, fmt.Errorf("TokenAccount data truncated at delegate flag")
	}
	delegateFlag := data[offset]
	offset++

	if delegateFlag == 1 {
		if offset+32 > len(data) {
			return nil, nil, 0, nil, false, 0, false, fmt.Errorf("TokenAccount data truncated at delegate")
		}
		delegate = make([]byte, 32)
		copy(delegate, data[offset:offset+32])
		hasDelegate = true
		offset += 32
	}

	// Read delegated_amount
	if offset+8 > len(data) {
		return nil, nil, 0, nil, false, 0, false, fmt.Errorf("TokenAccount data truncated at delegated_amount")
	}
	delegatedAmount := binary.LittleEndian.Uint64(data[offset : offset+8])
	offset += 8

	// Read frozen
	if offset >= len(data) {
		return nil, nil, 0, nil, false, 0, false, fmt.Errorf("TokenAccount data truncated at frozen flag")
	}
	frozen := data[offset] != 0

	return mint, owner, tokenBalance, delegate, hasDelegate, delegatedAmount, frozen, nil
}

// Serialization helper for writing values to byte buffers

// SerializeU8 serializes a u8 value to bytes
func SerializeU8(value uint8) []byte {
	return []byte{value}
}

// SerializeU64 serializes a u64 value to bytes (little-endian)
func SerializeU64(value uint64) []byte {
	data := make([]byte, 8)
	binary.LittleEndian.PutUint64(data, value)
	return data
}

// SerializeI64 serializes an i64 value to bytes (little-endian)
func SerializeI64(value int64) []byte {
	data := make([]byte, 8)
	binary.LittleEndian.PutUint64(data, uint64(value))
	return data
}

// SerializeBool serializes a boolean value to bytes
func SerializeBool(value bool) []byte {
	if value {
		return []byte{1}
	}
	return []byte{0}
}

// SerializeOptionalPublicKey serializes an optional PublicKey to bytes
// Format: 1 byte flag (0=null, 1=present) + 32 bytes if present
func SerializeOptionalPublicKey(pubKey []byte) ([]byte, error) {
	if pubKey == nil {
		return []byte{0}, nil
	}

	if len(pubKey) != 32 {
		return nil, fmt.Errorf("PublicKey must be 32 bytes, got %d", len(pubKey))
	}

	data := make([]byte, 33)
	data[0] = 1
	copy(data[1:], pubKey)
	return data, nil
}
