package quanticscript

import (
	"fmt"
)

// ValueType represents the type of a value in QuanticScript
type ValueType uint8

const (
	TypeI8        ValueType = iota // Signed 8-bit integer
	TypeI16                        // Signed 16-bit integer
	TypeI32                        // Signed 32-bit integer
	TypeI64                        // Signed 64-bit integer
	TypeU8                         // Unsigned 8-bit integer
	TypeU16                        // Unsigned 16-bit integer
	TypeU32                        // Unsigned 32-bit integer
	TypeU64                        // Unsigned 64-bit integer
	TypeBool                       // Boolean
	TypeBytes                      // Byte array
	TypeString                     // UTF-8 string
	TypeFileID                     // 32-byte file identifier
	TypePublicKey                  // 32-byte Ed25519 public key
	TypeTxID                       // 32-byte transaction identifier
	TypeArray                      // Array of values
	TypeMap                        // Map/dictionary
	TypeSet                        // Set of unique values
)

// Value represents a runtime value in QuanticScript
type Value struct {
	Type ValueType
	Data interface{}
}

// NewI64 creates a new i64 value
func NewI64(v int64) Value {
	return Value{Type: TypeI64, Data: v}
}

// NewU64 creates a new u64 value
func NewU64(v uint64) Value {
	return Value{Type: TypeU64, Data: v}
}

// NewBool creates a new bool value
func NewBool(v bool) Value {
	return Value{Type: TypeBool, Data: v}
}

// NewBytes creates a new bytes value
func NewBytes(v []byte) Value {
	return Value{Type: TypeBytes, Data: v}
}

// NewString creates a new string value
func NewString(v string) Value {
	return Value{Type: TypeString, Data: v}
}

// AsI64 returns the value as an i64
func (v Value) AsI64() (int64, error) {
	if v.Type != TypeI64 {
		return 0, fmt.Errorf("type mismatch: expected i64, got %v", v.Type)
	}
	return v.Data.(int64), nil
}

// AsU64 returns the value as a u64
func (v Value) AsU64() (uint64, error) {
	if v.Type != TypeU64 {
		return 0, fmt.Errorf("type mismatch: expected u64, got %v", v.Type)
	}
	return v.Data.(uint64), nil
}

// AsBool returns the value as a bool
func (v Value) AsBool() (bool, error) {
	if v.Type != TypeBool {
		return false, fmt.Errorf("type mismatch: expected bool, got %v", v.Type)
	}
	return v.Data.(bool), nil
}

// AsBytes returns the value as bytes
func (v Value) AsBytes() ([]byte, error) {
	if v.Type != TypeBytes {
		return nil, fmt.Errorf("type mismatch: expected bytes, got %v", v.Type)
	}
	return v.Data.([]byte), nil
}

// AsString returns the value as a string
func (v Value) AsString() (string, error) {
	if v.Type != TypeString {
		return "", fmt.Errorf("type mismatch: expected string, got %v", v.Type)
	}
	return v.Data.(string), nil
}

// NewFileID creates a new FileID value
func NewFileID(v []byte) Value {
	return Value{Type: TypeFileID, Data: v}
}

// NewPublicKey creates a new PublicKey value
func NewPublicKey(v []byte) Value {
	return Value{Type: TypePublicKey, Data: v}
}

// NewU32 creates a new u32 value
func NewU32(v uint32) Value {
	return Value{Type: TypeU32, Data: v}
}

// AsU32 returns the value as a u32
func (v Value) AsU32() (uint32, error) {
	if v.Type != TypeU32 {
		return 0, fmt.Errorf("type mismatch: expected u32, got %v", v.Type)
	}
	return v.Data.(uint32), nil
}

// NewTxID creates a new TxID value
func NewTxID(v []byte) Value {
	return Value{Type: TypeTxID, Data: v}
}

// NewArray creates a new array value
func NewArray(v []Value) Value {
	return Value{Type: TypeArray, Data: v}
}

// AsArray returns the value as an array
func (v Value) AsArray() ([]Value, error) {
	if v.Type != TypeArray {
		return nil, fmt.Errorf("type mismatch: expected array, got %v", v.Type)
	}
	return v.Data.([]Value), nil
}

// NewMap creates a new map value
func NewMap(v map[string]Value) Value {
	return Value{Type: TypeMap, Data: v}
}

// AsMap returns the value as a map
func (v Value) AsMap() (map[string]Value, error) {
	if v.Type != TypeMap {
		return nil, fmt.Errorf("type mismatch: expected map, got %v", v.Type)
	}
	return v.Data.(map[string]Value), nil
}

// NewSet creates a new set value
func NewSet(v map[string]bool) Value {
	return Value{Type: TypeSet, Data: v}
}

// AsSet returns the value as a set
func (v Value) AsSet() (map[string]bool, error) {
	if v.Type != TypeSet {
		return nil, fmt.Errorf("type mismatch: expected set, got %v", v.Type)
	}
	return v.Data.(map[string]bool), nil
}
