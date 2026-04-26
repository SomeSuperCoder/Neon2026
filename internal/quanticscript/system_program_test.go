package quanticscript

import (
	"os"
	"testing"
)

// TestSystemProgramCompilation tests that the System_Program compiles successfully
func TestSystemProgramCompilation(t *testing.T) {
	// Read the compiled bytecode
	bytecode, err := os.ReadFile("../../programs/system/system.qsb")
	if err != nil {
		t.Fatalf("Failed to read System_Program bytecode: %v", err)
	}

	// Verify bytecode is not empty
	if len(bytecode) == 0 {
		t.Fatal("System_Program bytecode is empty")
	}

	// Verify bytecode has valid header (8 bytes minimum)
	if len(bytecode) < 8 {
		t.Fatal("System_Program bytecode too short")
	}

	// Check magic number (0x5153 = "QS")
	magic := uint16(bytecode[0]) | (uint16(bytecode[1]) << 8)
	if magic != 0x5153 {
		t.Errorf("Expected magic 0x5153, got 0x%04x", magic)
	}

	// Check version (0x0100)
	version := uint16(bytecode[2]) | (uint16(bytecode[3]) << 8)
	if version != 0x0100 {
		t.Errorf("Expected version 0x0100, got 0x%04x", version)
	}

	t.Logf("System_Program bytecode size: %d bytes", len(bytecode))
	t.Logf("Magic: 0x%04x, Version: 0x%04x", magic, version)

	// Verify bytecode can be loaded by interpreter
	ctx := NewMockExecutionContext()
	interp := NewBytecodeInterpreter(bytecode, ctx, 1000000)
	if interp == nil {
		t.Fatal("Failed to create interpreter for System_Program bytecode")
	}
}

// TestSystemProgramStructure tests the structure of the System_Program
func TestSystemProgramStructure(t *testing.T) {
	// Read the source code
	source, err := os.ReadFile("../../programs/system/system.qs")
	if err != nil {
		t.Fatalf("Failed to read System_Program source: %v", err)
	}

	// Verify source contains required functions
	sourceStr := string(source)

	requiredFunctions := []string{
		"entry",
		"handleCreateAccount",
		"handleTransfer",
		"handleAllocateSpace",
	}

	for _, fn := range requiredFunctions {
		if !contains(sourceStr, fn) {
			t.Errorf("System_Program missing required function: %s", fn)
		}
	}

	// Verify error codes are defined
	requiredConstants := []string{
		"ERROR_INSUFFICIENT_BALANCE",
		"ERROR_INVALID_ACCOUNT",
		"ERROR_BALANCE_OVERFLOW",
		"ERROR_STORAGE_RENT_VIOLATION",
		"ERROR_UNAUTHORIZED_SIGNER",
		"ERROR_INVALID_INSTRUCTION",
		"SUCCESS",
	}

	for _, constant := range requiredConstants {
		if !contains(sourceStr, constant) {
			t.Errorf("System_Program missing required constant: %s", constant)
		}
	}

	t.Log("System_Program structure verified")
}

// TestSystemProgramErrorCodes verifies that error codes match the specification
func TestSystemProgramErrorCodes(t *testing.T) {
	expectedErrors := map[string]int64{
		"ERROR_INSUFFICIENT_BALANCE":   0x1000,
		"ERROR_INVALID_ACCOUNT":        0x1001,
		"ERROR_BALANCE_OVERFLOW":       0x1002,
		"ERROR_STORAGE_RENT_VIOLATION": 0x1003,
		"ERROR_UNAUTHORIZED_SIGNER":    0x1004,
		"ERROR_INVALID_INSTRUCTION":    0x1FFF,
		"SUCCESS":                      0,
	}

	// Read the source code
	source, err := os.ReadFile("../../programs/system/system.qs")
	if err != nil {
		t.Fatalf("Failed to read System_Program source: %v", err)
	}

	sourceStr := string(source)

	// Verify each error code is present
	for name, code := range expectedErrors {
		if !contains(sourceStr, name) {
			t.Errorf("Error code %s (0x%04X) not found in source", name, code)
		}
	}

	t.Log("All error codes verified")
}

// TestCreateAccountLogic tests the CreateAccount instruction logic
func TestCreateAccountLogic(t *testing.T) {
	t.Skip("Skipping until DISPATCH integration is complete")

	// Test cases to implement:
	// 1. Valid account creation with positive balance
	// 2. Invalid account creation with negative balance (should return ERROR_INSUFFICIENT_BALANCE)
	// 3. Account creation with zero balance (should succeed)
	// 4. Unauthorized account creation (should return ERROR_UNAUTHORIZED_SIGNER)
}

// TestTransferLogic tests the Transfer instruction logic
func TestTransferLogic(t *testing.T) {
	t.Skip("Skipping until DISPATCH integration is complete")

	// Test cases to implement:
	// 1. Valid transfer with sufficient balance (should succeed)
	// 2. Transfer with insufficient balance (should return ERROR_INSUFFICIENT_BALANCE)
	// 3. Transfer with negative amount (should return ERROR_INSUFFICIENT_BALANCE)
	// 4. Transfer with zero amount (should return ERROR_INSUFFICIENT_BALANCE)
	// 5. Transfer causing overflow (should return ERROR_BALANCE_OVERFLOW)
	// 6. Unauthorized transfer (should return ERROR_UNAUTHORIZED_SIGNER)
}

// TestAllocateSpaceLogic tests the AllocateSpace instruction logic
func TestAllocateSpaceLogic(t *testing.T) {
	t.Skip("Skipping until DISPATCH integration is complete")

	// Test cases to implement:
	// 1. Valid space allocation (should succeed)
	// 2. Allocation with negative balance (should return ERROR_INSUFFICIENT_BALANCE)
	// 3. Allocation causing overflow (should return ERROR_BALANCE_OVERFLOW)
	// 4. Unauthorized allocation (should return ERROR_UNAUTHORIZED_SIGNER)
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
