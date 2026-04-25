package quanticscript

import (
	"bytes"
	"testing"
)

// TestFileIDConversion tests FileID type conversion helpers
func TestFileIDConversion(t *testing.T) {
	t.Run("NewFileIDFromBytes_Valid", func(t *testing.T) {
		data := make([]byte, 32)
		for i := range data {
			data[i] = byte(i)
		}

		value, err := NewFileIDFromBytes(data)
		if err != nil {
			t.Fatalf("Failed to create FileID: %v", err)
		}

		if value.Type != TypeFileID {
			t.Errorf("Expected TypeFileID, got %v", value.Type)
		}

		extracted, err := FileIDToBytes(value)
		if err != nil {
			t.Fatalf("Failed to extract FileID bytes: %v", err)
		}

		if !bytes.Equal(data, extracted) {
			t.Errorf("FileID bytes mismatch")
		}
	})

	t.Run("NewFileIDFromBytes_InvalidLength", func(t *testing.T) {
		data := make([]byte, 16) // Wrong length

		_, err := NewFileIDFromBytes(data)
		if err == nil {
			t.Error("Expected error for invalid FileID length")
		}
	})

	t.Run("FileIDToBytes_WrongType", func(t *testing.T) {
		value := NewU64(42)

		_, err := FileIDToBytes(value)
		if err == nil {
			t.Error("Expected error for wrong type")
		}
	})
}

// TestPublicKeyConversion tests PublicKey type conversion helpers
func TestPublicKeyConversion(t *testing.T) {
	t.Run("NewPublicKeyFromBytes_Valid", func(t *testing.T) {
		data := make([]byte, 32)
		for i := range data {
			data[i] = byte(i + 100)
		}

		value, err := NewPublicKeyFromBytes(data)
		if err != nil {
			t.Fatalf("Failed to create PublicKey: %v", err)
		}

		if value.Type != TypePublicKey {
			t.Errorf("Expected TypePublicKey, got %v", value.Type)
		}

		extracted, err := PublicKeyToBytes(value)
		if err != nil {
			t.Fatalf("Failed to extract PublicKey bytes: %v", err)
		}

		if !bytes.Equal(data, extracted) {
			t.Errorf("PublicKey bytes mismatch")
		}
	})

	t.Run("NewPublicKeyFromBytes_InvalidLength", func(t *testing.T) {
		data := make([]byte, 20) // Wrong length

		_, err := NewPublicKeyFromBytes(data)
		if err == nil {
			t.Error("Expected error for invalid PublicKey length")
		}
	})
}

// TestInstructionParsing tests instruction data parsing utilities
func TestInstructionParsing(t *testing.T) {
	t.Run("ParseInstructionU8", func(t *testing.T) {
		data := []byte{42, 100, 200}

		value, offset, err := ParseInstructionU8(data, 0)
		if err != nil {
			t.Fatalf("Failed to parse u8: %v", err)
		}
		if value != 42 {
			t.Errorf("Expected 42, got %d", value)
		}
		if offset != 1 {
			t.Errorf("Expected offset 1, got %d", offset)
		}

		value, offset, err = ParseInstructionU8(data, 2)
		if err != nil {
			t.Fatalf("Failed to parse u8: %v", err)
		}
		if value != 200 {
			t.Errorf("Expected 200, got %d", value)
		}
		if offset != 3 {
			t.Errorf("Expected offset 3, got %d", offset)
		}
	})

	t.Run("ParseInstructionU64", func(t *testing.T) {
		data := []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}

		value, offset, err := ParseInstructionU64(data, 0)
		if err != nil {
			t.Fatalf("Failed to parse u64: %v", err)
		}
		if value != 0x0807060504030201 {
			t.Errorf("Expected 0x0807060504030201, got 0x%x", value)
		}
		if offset != 8 {
			t.Errorf("Expected offset 8, got %d", offset)
		}
	})

	t.Run("ParseInstructionI64", func(t *testing.T) {
		data := []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF} // -1

		value, offset, err := ParseInstructionI64(data, 0)
		if err != nil {
			t.Fatalf("Failed to parse i64: %v", err)
		}
		if value != -1 {
			t.Errorf("Expected -1, got %d", value)
		}
		if offset != 8 {
			t.Errorf("Expected offset 8, got %d", offset)
		}
	})

	t.Run("ParseInstructionFileID", func(t *testing.T) {
		data := make([]byte, 32)
		for i := range data {
			data[i] = byte(i)
		}

		fileID, offset, err := ParseInstructionFileID(data, 0)
		if err != nil {
			t.Fatalf("Failed to parse FileID: %v", err)
		}
		if len(fileID) != 32 {
			t.Errorf("Expected FileID length 32, got %d", len(fileID))
		}
		if !bytes.Equal(data, fileID) {
			t.Errorf("FileID bytes mismatch")
		}
		if offset != 32 {
			t.Errorf("Expected offset 32, got %d", offset)
		}
	})

	t.Run("ParseInstructionPublicKey", func(t *testing.T) {
		data := make([]byte, 32)
		for i := range data {
			data[i] = byte(i + 50)
		}

		pubKey, offset, err := ParseInstructionPublicKey(data, 0)
		if err != nil {
			t.Fatalf("Failed to parse PublicKey: %v", err)
		}
		if len(pubKey) != 32 {
			t.Errorf("Expected PublicKey length 32, got %d", len(pubKey))
		}
		if !bytes.Equal(data, pubKey) {
			t.Errorf("PublicKey bytes mismatch")
		}
		if offset != 32 {
			t.Errorf("Expected offset 32, got %d", offset)
		}
	})

	t.Run("ParseInstructionBool", func(t *testing.T) {
		data := []byte{1, 0, 255}

		value, offset, err := ParseInstructionBool(data, 0)
		if err != nil {
			t.Fatalf("Failed to parse bool: %v", err)
		}
		if !value {
			t.Errorf("Expected true, got false")
		}
		if offset != 1 {
			t.Errorf("Expected offset 1, got %d", offset)
		}

		value, offset, err = ParseInstructionBool(data, 1)
		if err != nil {
			t.Fatalf("Failed to parse bool: %v", err)
		}
		if value {
			t.Errorf("Expected false, got true")
		}
		if offset != 2 {
			t.Errorf("Expected offset 2, got %d", offset)
		}
	})

	t.Run("ParseInstructionOptionalPublicKey_Null", func(t *testing.T) {
		data := []byte{0} // Null flag

		pubKey, hasValue, offset, err := ParseInstructionOptionalPublicKey(data, 0)
		if err != nil {
			t.Fatalf("Failed to parse optional PublicKey: %v", err)
		}
		if hasValue {
			t.Errorf("Expected null value")
		}
		if pubKey != nil {
			t.Errorf("Expected nil pubKey")
		}
		if offset != 1 {
			t.Errorf("Expected offset 1, got %d", offset)
		}
	})

	t.Run("ParseInstructionOptionalPublicKey_Present", func(t *testing.T) {
		data := make([]byte, 33)
		data[0] = 1 // Present flag
		for i := 1; i < 33; i++ {
			data[i] = byte(i)
		}

		pubKey, hasValue, offset, err := ParseInstructionOptionalPublicKey(data, 0)
		if err != nil {
			t.Fatalf("Failed to parse optional PublicKey: %v", err)
		}
		if !hasValue {
			t.Errorf("Expected present value")
		}
		if len(pubKey) != 32 {
			t.Errorf("Expected PublicKey length 32, got %d", len(pubKey))
		}
		if !bytes.Equal(data[1:], pubKey) {
			t.Errorf("PublicKey bytes mismatch")
		}
		if offset != 33 {
			t.Errorf("Expected offset 33, got %d", offset)
		}
	})
}

// TestMintAccountSerialization tests MintAccount serialization
func TestMintAccountSerialization(t *testing.T) {
	t.Run("SerializeDeserialize_WithAuthorities", func(t *testing.T) {
		supply := uint64(1000000)
		decimals := uint8(6)
		mintAuthority := make([]byte, 32)
		freezeAuthority := make([]byte, 32)
		for i := range mintAuthority {
			mintAuthority[i] = byte(i)
			freezeAuthority[i] = byte(i + 100)
		}

		data, err := SerializeMintAccount(supply, decimals, mintAuthority, freezeAuthority)
		if err != nil {
			t.Fatalf("Failed to serialize MintAccount: %v", err)
		}

		// Expected size: 9 + 1 + 32 + 1 + 32 = 75 bytes
		if len(data) != 75 {
			t.Errorf("Expected data length 75, got %d", len(data))
		}

		s, d, ma, hasMA, fa, hasFA, err := DeserializeMintAccount(data)
		if err != nil {
			t.Fatalf("Failed to deserialize MintAccount: %v", err)
		}

		if s != supply {
			t.Errorf("Supply mismatch: expected %d, got %d", supply, s)
		}
		if d != decimals {
			t.Errorf("Decimals mismatch: expected %d, got %d", decimals, d)
		}
		if !hasMA {
			t.Error("Expected mint authority to be present")
		}
		if !bytes.Equal(ma, mintAuthority) {
			t.Error("Mint authority mismatch")
		}
		if !hasFA {
			t.Error("Expected freeze authority to be present")
		}
		if !bytes.Equal(fa, freezeAuthority) {
			t.Error("Freeze authority mismatch")
		}
	})

	t.Run("SerializeDeserialize_WithoutAuthorities", func(t *testing.T) {
		supply := uint64(5000000)
		decimals := uint8(9)

		data, err := SerializeMintAccount(supply, decimals, nil, nil)
		if err != nil {
			t.Fatalf("Failed to serialize MintAccount: %v", err)
		}

		// Expected size: 9 + 1 + 1 = 11 bytes
		if len(data) != 11 {
			t.Errorf("Expected data length 11, got %d", len(data))
		}

		s, d, ma, hasMA, fa, hasFA, err := DeserializeMintAccount(data)
		if err != nil {
			t.Fatalf("Failed to deserialize MintAccount: %v", err)
		}

		if s != supply {
			t.Errorf("Supply mismatch: expected %d, got %d", supply, s)
		}
		if d != decimals {
			t.Errorf("Decimals mismatch: expected %d, got %d", decimals, d)
		}
		if hasMA {
			t.Error("Expected mint authority to be null")
		}
		if ma != nil {
			t.Error("Expected nil mint authority")
		}
		if hasFA {
			t.Error("Expected freeze authority to be null")
		}
		if fa != nil {
			t.Error("Expected nil freeze authority")
		}
	})

	t.Run("SerializeDeserialize_MixedAuthorities", func(t *testing.T) {
		supply := uint64(2000000)
		decimals := uint8(2)
		mintAuthority := make([]byte, 32)
		for i := range mintAuthority {
			mintAuthority[i] = byte(i * 2)
		}

		data, err := SerializeMintAccount(supply, decimals, mintAuthority, nil)
		if err != nil {
			t.Fatalf("Failed to serialize MintAccount: %v", err)
		}

		// Expected size: 9 + 1 + 32 + 1 = 43 bytes
		if len(data) != 43 {
			t.Errorf("Expected data length 43, got %d", len(data))
		}

		s, d, ma, hasMA, fa, hasFA, err := DeserializeMintAccount(data)
		if err != nil {
			t.Fatalf("Failed to deserialize MintAccount: %v", err)
		}

		if s != supply {
			t.Errorf("Supply mismatch: expected %d, got %d", supply, s)
		}
		if d != decimals {
			t.Errorf("Decimals mismatch: expected %d, got %d", decimals, d)
		}
		if !hasMA {
			t.Error("Expected mint authority to be present")
		}
		if !bytes.Equal(ma, mintAuthority) {
			t.Error("Mint authority mismatch")
		}
		if hasFA {
			t.Error("Expected freeze authority to be null")
		}
		if fa != nil {
			t.Error("Expected nil freeze authority")
		}
	})
}

// TestTokenAccountSerialization tests TokenAccount serialization
func TestTokenAccountSerialization(t *testing.T) {
	t.Run("SerializeDeserialize_WithDelegate", func(t *testing.T) {
		mint := make([]byte, 32)
		owner := make([]byte, 32)
		delegate := make([]byte, 32)
		for i := range mint {
			mint[i] = byte(i)
			owner[i] = byte(i + 50)
			delegate[i] = byte(i + 100)
		}
		tokenBalance := uint64(500000)
		delegatedAmount := uint64(100000)
		frozen := true

		data, err := SerializeTokenAccount(mint, owner, tokenBalance, delegate, delegatedAmount, frozen)
		if err != nil {
			t.Fatalf("Failed to serialize TokenAccount: %v", err)
		}

		// Expected size: 32 + 32 + 8 + 1 + 32 + 8 + 1 = 114 bytes
		if len(data) != 114 {
			t.Errorf("Expected data length 114, got %d", len(data))
		}

		m, o, tb, d, hasD, da, f, err := DeserializeTokenAccount(data)
		if err != nil {
			t.Fatalf("Failed to deserialize TokenAccount: %v", err)
		}

		if !bytes.Equal(m, mint) {
			t.Error("Mint mismatch")
		}
		if !bytes.Equal(o, owner) {
			t.Error("Owner mismatch")
		}
		if tb != tokenBalance {
			t.Errorf("Token balance mismatch: expected %d, got %d", tokenBalance, tb)
		}
		if !hasD {
			t.Error("Expected delegate to be present")
		}
		if !bytes.Equal(d, delegate) {
			t.Error("Delegate mismatch")
		}
		if da != delegatedAmount {
			t.Errorf("Delegated amount mismatch: expected %d, got %d", delegatedAmount, da)
		}
		if f != frozen {
			t.Errorf("Frozen mismatch: expected %v, got %v", frozen, f)
		}
	})

	t.Run("SerializeDeserialize_WithoutDelegate", func(t *testing.T) {
		mint := make([]byte, 32)
		owner := make([]byte, 32)
		for i := range mint {
			mint[i] = byte(i * 3)
			owner[i] = byte(i * 5)
		}
		tokenBalance := uint64(750000)
		delegatedAmount := uint64(0)
		frozen := false

		data, err := SerializeTokenAccount(mint, owner, tokenBalance, nil, delegatedAmount, frozen)
		if err != nil {
			t.Fatalf("Failed to serialize TokenAccount: %v", err)
		}

		// Expected size: 32 + 32 + 8 + 1 + 8 + 1 = 82 bytes
		if len(data) != 82 {
			t.Errorf("Expected data length 82, got %d", len(data))
		}

		m, o, tb, d, hasD, da, f, err := DeserializeTokenAccount(data)
		if err != nil {
			t.Fatalf("Failed to deserialize TokenAccount: %v", err)
		}

		if !bytes.Equal(m, mint) {
			t.Error("Mint mismatch")
		}
		if !bytes.Equal(o, owner) {
			t.Error("Owner mismatch")
		}
		if tb != tokenBalance {
			t.Errorf("Token balance mismatch: expected %d, got %d", tokenBalance, tb)
		}
		if hasD {
			t.Error("Expected delegate to be null")
		}
		if d != nil {
			t.Error("Expected nil delegate")
		}
		if da != delegatedAmount {
			t.Errorf("Delegated amount mismatch: expected %d, got %d", delegatedAmount, da)
		}
		if f != frozen {
			t.Errorf("Frozen mismatch: expected %v, got %v", frozen, f)
		}
	})
}

// TestSerializationHelpers tests basic serialization helper functions
func TestSerializationHelpers(t *testing.T) {
	t.Run("SerializeU8", func(t *testing.T) {
		data := SerializeU8(42)
		if len(data) != 1 {
			t.Errorf("Expected length 1, got %d", len(data))
		}
		if data[0] != 42 {
			t.Errorf("Expected 42, got %d", data[0])
		}
	})

	t.Run("SerializeU64", func(t *testing.T) {
		data := SerializeU64(0x0102030405060708)
		if len(data) != 8 {
			t.Errorf("Expected length 8, got %d", len(data))
		}
		// Little-endian
		if data[0] != 0x08 || data[7] != 0x01 {
			t.Errorf("Incorrect little-endian encoding")
		}
	})

	t.Run("SerializeI64", func(t *testing.T) {
		data := SerializeI64(-1)
		if len(data) != 8 {
			t.Errorf("Expected length 8, got %d", len(data))
		}
		// -1 in two's complement
		for i := 0; i < 8; i++ {
			if data[i] != 0xFF {
				t.Errorf("Expected 0xFF at position %d, got 0x%02x", i, data[i])
			}
		}
	})

	t.Run("SerializeBool", func(t *testing.T) {
		dataTrue := SerializeBool(true)
		if len(dataTrue) != 1 || dataTrue[0] != 1 {
			t.Error("Expected [1] for true")
		}

		dataFalse := SerializeBool(false)
		if len(dataFalse) != 1 || dataFalse[0] != 0 {
			t.Error("Expected [0] for false")
		}
	})

	t.Run("SerializeOptionalPublicKey_Null", func(t *testing.T) {
		data, err := SerializeOptionalPublicKey(nil)
		if err != nil {
			t.Fatalf("Failed to serialize null PublicKey: %v", err)
		}
		if len(data) != 1 || data[0] != 0 {
			t.Error("Expected [0] for null PublicKey")
		}
	})

	t.Run("SerializeOptionalPublicKey_Present", func(t *testing.T) {
		pubKey := make([]byte, 32)
		for i := range pubKey {
			pubKey[i] = byte(i)
		}

		data, err := SerializeOptionalPublicKey(pubKey)
		if err != nil {
			t.Fatalf("Failed to serialize PublicKey: %v", err)
		}
		if len(data) != 33 {
			t.Errorf("Expected length 33, got %d", len(data))
		}
		if data[0] != 1 {
			t.Error("Expected flag byte to be 1")
		}
		if !bytes.Equal(data[1:], pubKey) {
			t.Error("PublicKey bytes mismatch")
		}
	})
}
