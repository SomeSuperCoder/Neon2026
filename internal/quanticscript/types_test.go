package quanticscript

import (
	"testing"
)

func TestValueCreation(t *testing.T) {
	tests := []struct {
		name     string
		value    Value
		expected interface{}
	}{
		{"i64", NewI64(42), int64(42)},
		{"u64", NewU64(100), uint64(100)},
		{"bool_true", NewBool(true), true},
		{"bool_false", NewBool(false), false},
		{"bytes", NewBytes([]byte{1, 2, 3}), []byte{1, 2, 3}},
		{"string", NewString("hello"), "hello"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.value.Data == nil {
				t.Errorf("Value.Data is nil")
			}
		})
	}
}

func TestAsI64(t *testing.T) {
	v := NewI64(42)
	result, err := v.AsI64()
	if err != nil {
		t.Errorf("AsI64() error = %v", err)
	}
	if result != 42 {
		t.Errorf("AsI64() = %v, want 42", result)
	}

	// Test type mismatch
	wrongType := NewU64(100)
	_, err = wrongType.AsI64()
	if err == nil {
		t.Error("AsI64() should return error for wrong type")
	}
}

func TestAsU64(t *testing.T) {
	v := NewU64(100)
	result, err := v.AsU64()
	if err != nil {
		t.Errorf("AsU64() error = %v", err)
	}
	if result != 100 {
		t.Errorf("AsU64() = %v, want 100", result)
	}

	// Test type mismatch
	wrongType := NewI64(42)
	_, err = wrongType.AsU64()
	if err == nil {
		t.Error("AsU64() should return error for wrong type")
	}
}

func TestAsBool(t *testing.T) {
	v := NewBool(true)
	result, err := v.AsBool()
	if err != nil {
		t.Errorf("AsBool() error = %v", err)
	}
	if result != true {
		t.Errorf("AsBool() = %v, want true", result)
	}

	// Test type mismatch
	wrongType := NewI64(1)
	_, err = wrongType.AsBool()
	if err == nil {
		t.Error("AsBool() should return error for wrong type")
	}
}

func TestAsBytes(t *testing.T) {
	expected := []byte{1, 2, 3, 4}
	v := NewBytes(expected)
	result, err := v.AsBytes()
	if err != nil {
		t.Errorf("AsBytes() error = %v", err)
	}
	if len(result) != len(expected) {
		t.Errorf("AsBytes() length = %v, want %v", len(result), len(expected))
	}
	for i := range expected {
		if result[i] != expected[i] {
			t.Errorf("AsBytes()[%d] = %v, want %v", i, result[i], expected[i])
		}
	}

	// Test type mismatch
	wrongType := NewString("test")
	_, err = wrongType.AsBytes()
	if err == nil {
		t.Error("AsBytes() should return error for wrong type")
	}
}

func TestAsString(t *testing.T) {
	v := NewString("hello world")
	result, err := v.AsString()
	if err != nil {
		t.Errorf("AsString() error = %v", err)
	}
	if result != "hello world" {
		t.Errorf("AsString() = %v, want 'hello world'", result)
	}

	// Test type mismatch
	wrongType := NewBytes([]byte("test"))
	_, err = wrongType.AsString()
	if err == nil {
		t.Error("AsString() should return error for wrong type")
	}
}

func TestValueTypes(t *testing.T) {
	tests := []struct {
		name     string
		value    Value
		expected ValueType
	}{
		{"i64", NewI64(0), TypeI64},
		{"u64", NewU64(0), TypeU64},
		{"bool", NewBool(false), TypeBool},
		{"bytes", NewBytes(nil), TypeBytes},
		{"string", NewString(""), TypeString},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.value.Type != tt.expected {
				t.Errorf("Value.Type = %v, want %v", tt.value.Type, tt.expected)
			}
		})
	}
}
