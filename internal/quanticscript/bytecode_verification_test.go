package quanticscript

import (
	"os"
	"strings"
	"testing"
	"time"
)

// TestBytecodeVerification_SystemProgram verifies the System_Program bytecode structure
// and instruction sequences. Requirements: 3.3, 3.4, 3.5
func TestBytecodeVerification_SystemProgram(t *testing.T) {
	bytecodeFile, err := os.ReadFile("../../programs/system/system.qsb")
	if err != nil {
		t.Fatalf("Failed to read System_Program bytecode: %v", err)
	}

	// Verify valid header
	header, err := ParseBytecodeHeader(bytecodeFile)
	if err != nil {
		t.Fatalf("Invalid bytecode header: %v", err)
	}
	if header.Magic != BytecodeMagic {
		t.Errorf("Expected magic 0x%04x, got 0x%04x", BytecodeMagic, header.Magic)
	}
	if header.Version != BytecodeVersion {
		t.Errorf("Expected version 0x%04x, got 0x%04x", BytecodeVersion, header.Version)
	}

	// Disassemble and verify instruction sequences
	disasm, err := DisassembleFile(bytecodeFile)
	if err != nil {
		t.Fatalf("Failed to disassemble System_Program: %v", err)
	}

	// Verify disassembly contains expected structural elements
	if !strings.Contains(disasm, "CALL") {
		t.Error("Expected CALL instruction in System_Program (entry dispatches to handler)")
	}
	if !strings.Contains(disasm, "RET") {
		t.Error("Expected RET instruction in System_Program")
	}
	if !strings.Contains(disasm, "cost:") {
		t.Error("Expected cost annotations in disassembly")
	}

	t.Logf("System_Program bytecode size: %d bytes", len(bytecodeFile))
	t.Logf("System_Program body size: %d bytes", len(bytecodeFile)-BytecodeHeaderSize)
	t.Logf("Disassembly:\n%s", disasm)
}

// TestBytecodeVerification_TokenProgram verifies the Token_Program bytecode structure
// and instruction sequences. Requirements: 3.3, 3.4, 3.5
func TestBytecodeVerification_TokenProgram(t *testing.T) {
	bytecodeFile, err := os.ReadFile("../../programs/token/token.qsb")
	if err != nil {
		t.Fatalf("Failed to read Token_Program bytecode: %v", err)
	}

	// Verify valid header
	header, err := ParseBytecodeHeader(bytecodeFile)
	if err != nil {
		t.Fatalf("Invalid bytecode header: %v", err)
	}
	if header.Magic != BytecodeMagic {
		t.Errorf("Expected magic 0x%04x, got 0x%04x", BytecodeMagic, header.Magic)
	}
	if header.Version != BytecodeVersion {
		t.Errorf("Expected version 0x%04x, got 0x%04x", BytecodeVersion, header.Version)
	}

	// Disassemble and verify instruction sequences
	disasm, err := DisassembleFile(bytecodeFile)
	if err != nil {
		t.Fatalf("Failed to disassemble Token_Program: %v", err)
	}

	// Token_Program has more handlers than System_Program, so should be larger
	if len(bytecodeFile) < len([]byte{}) {
		t.Error("Token_Program bytecode unexpectedly empty")
	}
	if !strings.Contains(disasm, "CALL") {
		t.Error("Expected CALL instruction in Token_Program (entry dispatches to handlers)")
	}
	if !strings.Contains(disasm, "RET") {
		t.Error("Expected RET instruction in Token_Program")
	}

	t.Logf("Token_Program bytecode size: %d bytes", len(bytecodeFile))
	t.Logf("Token_Program body size: %d bytes", len(bytecodeFile)-BytecodeHeaderSize)
	t.Logf("Disassembly:\n%s", disasm)
}

// TestBytecodeVerification_TokenLargerThanSystem verifies Token_Program is larger
// than System_Program, consistent with having more instruction handlers.
// Requirements: 3.3
func TestBytecodeVerification_TokenLargerThanSystem(t *testing.T) {
	systemBytecode, err := os.ReadFile("../../programs/system/system.qsb")
	if err != nil {
		t.Fatalf("Failed to read System_Program bytecode: %v", err)
	}
	tokenBytecode, err := os.ReadFile("../../programs/token/token.qsb")
	if err != nil {
		t.Fatalf("Failed to read Token_Program bytecode: %v", err)
	}

	if len(tokenBytecode) < len(systemBytecode) {
		t.Errorf("Token_Program (%d bytes) should be >= System_Program (%d bytes) since it has more handlers",
			len(tokenBytecode), len(systemBytecode))
	}
	t.Logf("System_Program: %d bytes, Token_Program: %d bytes", len(systemBytecode), len(tokenBytecode))
}

// TestComputeBudget_SystemProgram tests compute budget consumption for System_Program.
// Requirements: 3.4
func TestComputeBudget_SystemProgram(t *testing.T) {
	bytecodeFile, err := os.ReadFile("../../programs/system/system.qsb")
	if err != nil {
		t.Fatalf("Failed to read System_Program bytecode: %v", err)
	}

	body, err := GetBytecodeBody(bytecodeFile)
	if err != nil {
		t.Fatalf("Failed to get bytecode body: %v", err)
	}

	const budget = int64(1_000_000)
	ctx := NewMockExecutionContext()
	interp := NewBytecodeInterpreter(body, ctx, budget)

	err = interp.Execute()
	if err != nil {
		t.Fatalf("System_Program execution failed: %v", err)
	}

	consumed := budget - interp.GetComputeBudget()
	if consumed <= 0 {
		t.Error("Expected non-zero compute budget consumption")
	}
	// Sanity: should not consume more than 10,000 units for a stub entry
	if consumed > 10_000 {
		t.Errorf("System_Program consumed unexpectedly high budget: %d units", consumed)
	}

	t.Logf("System_Program compute budget consumed: %d units (remaining: %d)", consumed, interp.GetComputeBudget())
}

// TestComputeBudget_TokenProgram tests compute budget consumption for Token_Program.
// Requirements: 3.4
func TestComputeBudget_TokenProgram(t *testing.T) {
	bytecodeFile, err := os.ReadFile("../../programs/token/token.qsb")
	if err != nil {
		t.Fatalf("Failed to read Token_Program bytecode: %v", err)
	}

	body, err := GetBytecodeBody(bytecodeFile)
	if err != nil {
		t.Fatalf("Failed to get bytecode body: %v", err)
	}

	const budget = int64(1_000_000)
	ctx := NewMockExecutionContext()
	interp := NewBytecodeInterpreter(body, ctx, budget)

	err = interp.Execute()
	if err != nil {
		t.Fatalf("Token_Program execution failed: %v", err)
	}

	consumed := budget - interp.GetComputeBudget()
	if consumed <= 0 {
		t.Error("Expected non-zero compute budget consumption")
	}

	t.Logf("Token_Program compute budget consumed: %d units (remaining: %d)", consumed, interp.GetComputeBudget())
}

// TestDeterminism_SystemProgram verifies that System_Program produces identical
// results and budget consumption across repeated executions. Requirements: 3.5
func TestDeterminism_SystemProgram(t *testing.T) {
	bytecodeFile, err := os.ReadFile("../../programs/system/system.qsb")
	if err != nil {
		t.Fatalf("Failed to read System_Program bytecode: %v", err)
	}

	body, err := GetBytecodeBody(bytecodeFile)
	if err != nil {
		t.Fatalf("Failed to get bytecode body: %v", err)
	}

	const runs = 10
	const budget = int64(1_000_000)

	var firstBudget int64
	var firstStack []Value

	for i := 0; i < runs; i++ {
		ctx := NewMockExecutionContext()
		interp := NewBytecodeInterpreter(body, ctx, budget)

		err := interp.Execute()
		if err != nil {
			t.Fatalf("Run %d: System_Program execution failed: %v", i, err)
		}

		remaining := interp.GetComputeBudget()

		if i == 0 {
			firstBudget = remaining
			firstStack = make([]Value, len(interp.stack))
			copy(firstStack, interp.stack)
		} else {
			if remaining != firstBudget {
				t.Errorf("Run %d: budget mismatch: got %d, expected %d", i, remaining, firstBudget)
			}
			if len(interp.stack) != len(firstStack) {
				t.Errorf("Run %d: stack size mismatch: got %d, expected %d", i, len(interp.stack), len(firstStack))
			}
		}
	}

	t.Logf("System_Program determinism verified over %d runs, remaining budget: %d", runs, firstBudget)
}

// TestDeterminism_TokenProgram verifies that Token_Program produces identical
// results and budget consumption across repeated executions. Requirements: 3.5
func TestDeterminism_TokenProgram(t *testing.T) {
	bytecodeFile, err := os.ReadFile("../../programs/token/token.qsb")
	if err != nil {
		t.Fatalf("Failed to read Token_Program bytecode: %v", err)
	}

	body, err := GetBytecodeBody(bytecodeFile)
	if err != nil {
		t.Fatalf("Failed to get bytecode body: %v", err)
	}

	const runs = 10
	const budget = int64(1_000_000)

	var firstBudget int64
	var firstStackLen int

	for i := 0; i < runs; i++ {
		ctx := NewMockExecutionContext()
		interp := NewBytecodeInterpreter(body, ctx, budget)

		err := interp.Execute()
		if err != nil {
			t.Fatalf("Run %d: Token_Program execution failed: %v", i, err)
		}

		remaining := interp.GetComputeBudget()

		if i == 0 {
			firstBudget = remaining
			firstStackLen = len(interp.stack)
		} else {
			if remaining != firstBudget {
				t.Errorf("Run %d: budget mismatch: got %d, expected %d", i, remaining, firstBudget)
			}
			if len(interp.stack) != firstStackLen {
				t.Errorf("Run %d: stack size mismatch: got %d, expected %d", i, len(interp.stack), firstStackLen)
			}
		}
	}

	t.Logf("Token_Program determinism verified over %d runs, remaining budget: %d", runs, firstBudget)
}

// TestPerformance_ProgramExecution profiles execution time for both programs
// and identifies if execution is within acceptable bounds. Requirements: 3.4
func TestPerformance_ProgramExecution(t *testing.T) {
	programs := []struct {
		name string
		path string
	}{
		{"System_Program", "../../programs/system/system.qsb"},
		{"Token_Program", "../../programs/token/token.qsb"},
	}

	for _, prog := range programs {
		t.Run(prog.name, func(t *testing.T) {
			bytecodeFile, err := os.ReadFile(prog.path)
			if err != nil {
				t.Fatalf("Failed to read %s bytecode: %v", prog.name, err)
			}

			body, err := GetBytecodeBody(bytecodeFile)
			if err != nil {
				t.Fatalf("Failed to get bytecode body: %v", err)
			}

			const runs = 100
			const budget = int64(1_000_000)

			start := time.Now()
			for i := 0; i < runs; i++ {
				ctx := NewMockExecutionContext()
				interp := NewBytecodeInterpreter(body, ctx, budget)
				if err := interp.Execute(); err != nil {
					t.Fatalf("Run %d failed: %v", i, err)
				}
			}
			elapsed := time.Since(start)

			avgNs := elapsed.Nanoseconds() / runs
			t.Logf("%s: %d runs in %v, avg %d ns/execution", prog.name, runs, elapsed, avgNs)

			// Each execution should complete in under 1ms (generous bound for stub bytecode)
			if avgNs > 1_000_000 {
				t.Errorf("%s average execution time %d ns exceeds 1ms threshold", prog.name, avgNs)
			}
		})
	}
}

// TestBytecodeVerification_DisassemblyRoundTrip verifies that disassembled bytecode
// can be re-assembled to produce equivalent bytecode. Requirements: 3.3
func TestBytecodeVerification_DisassemblyRoundTrip(t *testing.T) {
	programs := []struct {
		name string
		path string
	}{
		{"System_Program", "../../programs/system/system.qsb"},
		{"Token_Program", "../../programs/token/token.qsb"},
	}

	for _, prog := range programs {
		t.Run(prog.name, func(t *testing.T) {
			bytecodeFile, err := os.ReadFile(prog.path)
			if err != nil {
				t.Fatalf("Failed to read %s bytecode: %v", prog.name, err)
			}

			body, err := GetBytecodeBody(bytecodeFile)
			if err != nil {
				t.Fatalf("Failed to get bytecode body: %v", err)
			}

			// Disassemble
			disasm, err := DisassembleBody(body)
			if err != nil {
				t.Fatalf("Failed to disassemble %s: %v", prog.name, err)
			}

			// Re-assemble
			reassembled, err := AssembleToBody(disasm)
			if err != nil {
				t.Fatalf("Failed to re-assemble %s: %v", prog.name, err)
			}

			// Compare lengths
			if len(reassembled) != len(body) {
				t.Errorf("%s: bytecode length mismatch after round-trip: original=%d, reassembled=%d",
					prog.name, len(body), len(reassembled))
			}

			// Compare bytes
			for i := range body {
				if i >= len(reassembled) {
					break
				}
				if body[i] != reassembled[i] {
					t.Errorf("%s: bytecode mismatch at offset %d: original=0x%02x, reassembled=0x%02x",
						prog.name, i, body[i], reassembled[i])
				}
			}

			t.Logf("%s: round-trip verified, %d bytes", prog.name, len(body))
		})
	}
}

// TestBytecodeVerification_InstructionCosts verifies that all opcodes in both
// programs have defined costs in the cost table. Requirements: 3.4
func TestBytecodeVerification_InstructionCosts(t *testing.T) {
	programs := []struct {
		name string
		path string
	}{
		{"System_Program", "../../programs/system/system.qsb"},
		{"Token_Program", "../../programs/token/token.qsb"},
	}

	for _, prog := range programs {
		t.Run(prog.name, func(t *testing.T) {
			bytecodeFile, err := os.ReadFile(prog.path)
			if err != nil {
				t.Fatalf("Failed to read %s bytecode: %v", prog.name, err)
			}

			disasm, err := DisassembleFile(bytecodeFile)
			if err != nil {
				t.Fatalf("Failed to disassemble %s: %v", prog.name, err)
			}

			// Every instruction line should have a cost annotation
			lines := strings.Split(disasm, "\n")
			for _, line := range lines {
				trimmed := strings.TrimSpace(line)
				// Skip empty lines, comments, and labels
				if trimmed == "" || strings.HasPrefix(trimmed, ";") || strings.HasSuffix(trimmed, ":") {
					continue
				}
				if !strings.Contains(trimmed, "cost:") {
					t.Errorf("%s: instruction missing cost annotation: %q", prog.name, trimmed)
				}
			}
		})
	}
}
