package quanticscript

import (
	"testing"
	"time"

	"github.com/poh-blockchain/internal/filestore"
)

// compileSource is a helper function to compile QuanticScript source code
func compileSource(t *testing.T, source string) []byte {
	t.Helper()

	// Parse
	t.Log("compileSource: Starting lexer...")
	lexer := NewLexer(source, "test.qs")
	t.Log("compileSource: Starting parser...")
	parser := NewParser(lexer)
	t.Log("compileSource: Parsing program...")
	program := parser.ParseProgram()
	t.Logf("compileSource: Parser complete, errors: %d", len(parser.errors))
	if len(parser.errors) > 0 {
		t.Fatalf("Parser errors: %v", parser.errors)
	}

	// Type check
	t.Log("compileSource: Starting type checker...")
	tc := NewTypeChecker()
	t.Log("compileSource: Checking program...")
	tc.CheckProgram(program)
	t.Logf("compileSource: Type checker complete, errors: %d", len(tc.errors))
	if len(tc.errors) > 0 {
		t.Fatalf("Type checker errors: %v", tc.errors)
	}

	// Generate bytecode
	t.Log("compileSource: Starting code generator...")
	cg := NewCodeGenerator()
	t.Log("compileSource: Generating bytecode...")
	bytecode, err := cg.Generate(program)
	t.Logf("compileSource: Code generation complete, bytecode length: %d", len(bytecode))
	if err != nil {
		t.Fatalf("Code generator error: %v", err)
	}

	// Create bytecode file
	t.Log("compileSource: Creating bytecode file...")
	bytecodeFile := CreateBytecode(bytecode, 0)
	t.Log("compileSource: Getting bytecode body...")
	body, err := GetBytecodeBody(bytecodeFile)
	t.Logf("compileSource: Complete, body length: %d", len(body))
	if err != nil {
		t.Fatalf("Failed to get bytecode body: %v", err)
	}

	return body
}

// TestUpdateBalancePrivilege tests that only the system program can call updateBalance
func TestUpdateBalancePrivilege(t *testing.T) {
	t.Run("non-system program cannot call updateBalance", func(t *testing.T) {
		source := `
			export function entry(): i64 {
				let accountId: i64 = 100;
				let amount: i64 = 1000;
				updateBalance(accountId, amount);
				return 0;
			}
		`

		bytecode := compileSource(t, source)
		ctx := NewMockExecutionContext()
		ctx.programID[31] = 0x42

		interpreter := NewBytecodeInterpreter(bytecode, ctx, 100000)
		err := interpreter.Execute()

		if err == nil {
			t.Fatal("Expected security error, but execution succeeded")
		}

		if err.Error() != "UPDATEBALANCE can only be called by the system program" {
			t.Errorf("Expected security error message, got: %v", err)
		}
	})

	t.Run("system program can call updateBalance", func(t *testing.T) {
		source := `
			export function entry(): i64 {
				let accountId: i64 = 100;
				let amount: i64 = 1000;
				updateBalance(accountId, amount);
				return 0;
			}
		`

		bytecode := compileSource(t, source)
		ctx := NewMockExecutionContext()
		ctx.programID[31] = 0x01

		var accountID filestore.FileID
		accountID[31] = 0x64

		testAccount := &filestore.File{
			ID:      accountID,
			Balance: 5000,
			Data:    []byte{},
		}
		ctx.files[accountID] = testAccount

		interpreter := NewBytecodeInterpreter(bytecode, ctx, 100000)

		// Enable debug logging
		interpreter.SetDebugLogger(func(format string, args ...interface{}) {
			t.Logf(format, args...)
		})

		err := interpreter.Execute()

		if err != nil {
			t.Fatalf("Expected success, but got error: %v", err)
		}

		updatedAccount := ctx.files[accountID]
		expectedBalance := int64(6000)
		if updatedAccount.Balance != expectedBalance {
			t.Errorf("Expected balance %d, got: %d", expectedBalance, updatedAccount.Balance)
		}
	})

	t.Run("attempt to manipulate balance in loop", func(t *testing.T) {
		source := `
			export function entry(): i64 {
				let accountId: i64 = 100;
				let i: i64 = 0;
				while (i < 10) {
					updateBalance(accountId, 1000);
					i = i + 1;
				}
				return 0;
			}
		`

		bytecode := compileSource(t, source)
		ctx := NewMockExecutionContext()
		ctx.programID[31] = 0x42

		interpreter := NewBytecodeInterpreter(bytecode, ctx, 100000)
		err := interpreter.Execute()

		if err == nil {
			t.Fatal("Expected security error, but execution succeeded")
		}

		if err.Error() != "UPDATEBALANCE can only be called by the system program" {
			t.Errorf("Expected security error, got: %v", err)
		}
	})

	t.Run("attempt to call updateBalance in function", func(t *testing.T) {
		t.Log("Test starting...")
		source := `
			function transferTokens(fromAccount: i64, toAccount: i64, amount: i64): i64 {
				updateBalance(fromAccount, -amount);
				updateBalance(toAccount, amount);
				return 0;
			}
			
			export function entry(): i64 {
				return transferTokens(100, 200, 1000);
			}
		`

		t.Log("Compiling source...")
		bytecode := compileSource(t, source)
		t.Logf("Compilation complete, bytecode length: %d", len(bytecode))

		// Dump bytecode for debugging
		t.Logf("=== BYTECODE (%d bytes) ===", len(bytecode))
		for i := 0; i < len(bytecode) && i < 100; i++ {
			if name, ok := OpcodeNames[Opcode(bytecode[i])]; ok {
				t.Logf("[%03d] %s", i, name)
			} else {
				t.Logf("[%03d] 0x%02x", i, bytecode[i])
			}
		}

		t.Log("Creating context...")
		ctx := NewMockExecutionContext()
		ctx.programID[31] = 0x42

		t.Log("Creating interpreter...")
		interpreter := NewBytecodeInterpreter(bytecode, ctx, 100000)

		// Enable debug logging
		interpreter.SetDebugLogger(func(format string, args ...interface{}) {
			t.Logf(format, args...)
		})

		t.Log("Starting execution in goroutine...")
		// Use a timeout to detect real infinite loops
		done := make(chan error, 1)
		go func() {
			t.Log("Goroutine: calling Execute()")
			err := interpreter.Execute()
			t.Logf("Goroutine: Execute() returned with error: %v", err)
			done <- err
		}()

		t.Log("Waiting for result...")
		select {
		case err := <-done:
			t.Logf("Received result from goroutine: %v", err)
			// Normal completion or error
			if err == nil {
				t.Fatal("Expected security error, but execution succeeded")
			}
			if err.Error() != "UPDATEBALANCE can only be called by the system program" {
				t.Errorf("Expected security error, got: %v", err)
			}
		case <-time.After(1 * time.Second):
			t.Fatal("Test timed out - possible infinite loop")
		}
		t.Log("Test complete")
	})
}
