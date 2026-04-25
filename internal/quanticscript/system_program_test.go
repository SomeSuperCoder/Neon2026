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
}

// TestSystemProgramExecution tests basic execution of the System_Program
func TestSystemProgramExecution(t *testing.T) {
	t.Skip("Skipping execution test until ExecutionContext is fully implemented")

	// This test will be implemented when:
	// 1. ExecutionContext is properly set up
	// 2. Stdlib functions are available in QuanticScript
	// 3. Full instruction parsing is implemented
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

	t.Log("System_Program structure verified")
}

// TestCreateAccountLogic tests the CreateAccount instruction logic
// Note: This is a placeholder for when stdlib functions are available
func TestCreateAccountLogic(t *testing.T) {
	t.Skip("Skipping until stdlib functions are implemented")

	// Test cases:
	// 1. Valid account creation with positive balance
	// 2. Invalid account creation with negative balance
	// 3. Account creation with zero balance
}

// TestTransferLogic tests the Transfer instruction logic
// Note: This is a placeholder for when stdlib functions are available
func TestTransferLogic(t *testing.T) {
	t.Skip("Skipping until stdlib functions are implemented")

	// Test cases:
	// 1. Valid transfer with sufficient balance
	// 2. Transfer with insufficient balance (should fail)
	// 3. Transfer with negative amount (should fail)
	// 4. Transfer causing overflow (should fail)
	// 5. Unauthorized transfer (should fail)
}

// TestAllocateSpaceLogic tests the AllocateSpace instruction logic
// Note: This is a placeholder for when stdlib functions are available
func TestAllocateSpaceLogic(t *testing.T) {
	t.Skip("Skipping until stdlib functions are implemented")

	// Test cases:
	// 1. Valid space allocation
	// 2. Allocation with negative balance (should fail)
	// 3. Allocation causing overflow (should fail)
	// 4. Unauthorized allocation (should fail)
	// 5. Allocation violating storage rent (should fail)
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
