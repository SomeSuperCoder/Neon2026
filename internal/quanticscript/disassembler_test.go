package quanticscript

import (
	"strings"
	"testing"
)

func TestDisassemblerBasicInstructions(t *testing.T) {
	// Assemble some code
	assembly := `
		PUSH i64 42
		PUSH i64 10
		ADD
		RET
	`

	bytecode, err := AssembleToBody(assembly)
	if err != nil {
		t.Fatalf("Assemble() error = %v", err)
	}

	// Disassemble it
	disassembled, err := DisassembleBody(bytecode)
	if err != nil {
		t.Fatalf("Disassemble() error = %v", err)
	}

	// Check that key instructions are present
	if !strings.Contains(disassembled, "PUSH") {
		t.Error("Expected PUSH instruction in disassembly")
	}
	if !strings.Contains(disassembled, "ADD") {
		t.Error("Expected ADD instruction in disassembly")
	}
	if !strings.Contains(disassembled, "RET") {
		t.Error("Expected RET instruction in disassembly")
	}
}

func TestDisassemblerCostAnnotations(t *testing.T) {
	assembly := `
		PUSH i64 1
		PUSH i64 2
		MUL
		RET
	`

	bytecode, err := AssembleToBody(assembly)
	if err != nil {
		t.Fatalf("Assemble() error = %v", err)
	}

	disassembled, err := DisassembleBody(bytecode)
	if err != nil {
		t.Fatalf("Disassemble() error = %v", err)
	}

	// Check for cost annotations
	if !strings.Contains(disassembled, "cost:") {
		t.Error("Expected cost annotations in disassembly")
	}

	// MUL should have cost 3
	if !strings.Contains(disassembled, "MUL") || !strings.Contains(disassembled, "cost: 3") {
		t.Error("Expected MUL instruction with cost 3")
	}
}

func TestDisassemblerLabels(t *testing.T) {
	assembly := `
		PUSH i64 0
		JMPIF target
		PUSH i64 1
	target:
		RET
	`

	bytecode, err := AssembleToBody(assembly)
	if err != nil {
		t.Fatalf("Assemble() error = %v", err)
	}

	disassembled, err := DisassembleBody(bytecode)
	if err != nil {
		t.Fatalf("Disassemble() error = %v", err)
	}

	// Check that labels are generated
	if !strings.Contains(disassembled, "label_") {
		t.Error("Expected label in disassembly")
	}
	if !strings.Contains(disassembled, "JMPIF") {
		t.Error("Expected JMPIF instruction in disassembly")
	}
}

func TestDisassemblerFunctionCalls(t *testing.T) {
	assembly := `
		CALL func_0
		RET
	func_0:
		PUSH i64 42
		RET
	`

	bytecode, err := AssembleToBody(assembly)
	if err != nil {
		t.Fatalf("Assemble() error = %v", err)
	}

	disassembled, err := DisassembleBody(bytecode)
	if err != nil {
		t.Fatalf("Disassemble() error = %v", err)
	}

	// Check for CALL instruction and function label
	if !strings.Contains(disassembled, "CALL") {
		t.Error("Expected CALL instruction in disassembly")
	}
	if !strings.Contains(disassembled, "func_") {
		t.Error("Expected function label in disassembly")
	}
}

func TestDisassembleFile(t *testing.T) {
	assembly := `
		PUSH i64 100
		RET
	`

	bytecode, err := AssembleToFile(assembly)
	if err != nil {
		t.Fatalf("AssembleToFile() error = %v", err)
	}

	disassembled, err := DisassembleFile(bytecode)
	if err != nil {
		t.Fatalf("DisassembleFile() error = %v", err)
	}

	// Check for header comments
	if !strings.Contains(disassembled, "QuanticScript Bytecode") {
		t.Error("Expected header comment in disassembly")
	}
	if !strings.Contains(disassembled, "Version:") {
		t.Error("Expected version in disassembly")
	}
	if !strings.Contains(disassembled, "Entry Offset:") {
		t.Error("Expected entry offset in disassembly")
	}
}

func TestRoundTrip(t *testing.T) {
	// Test that assemble -> disassemble -> assemble produces equivalent bytecode
	original := `
		PUSH i64 10
		PUSH i64 20
		ADD
		STORE 0
		LOAD 0
		RET
	`

	// First assembly
	bytecode1, err := AssembleToBody(original)
	if err != nil {
		t.Fatalf("First Assemble() error = %v", err)
	}

	// Disassemble
	disassembled, err := DisassembleBody(bytecode1)
	if err != nil {
		t.Fatalf("Disassemble() error = %v", err)
	}

	// Second assembly
	bytecode2, err := AssembleToBody(disassembled)
	if err != nil {
		t.Fatalf("Second Assemble() error = %v", err)
	}

	// Compare bytecode
	if len(bytecode1) != len(bytecode2) {
		t.Errorf("Bytecode length mismatch: %d vs %d", len(bytecode1), len(bytecode2))
	}

	for i := range bytecode1 {
		if i >= len(bytecode2) {
			break
		}
		if bytecode1[i] != bytecode2[i] {
			t.Errorf("Bytecode mismatch at offset %d: 0x%02x vs 0x%02x", i, bytecode1[i], bytecode2[i])
		}
	}
}
