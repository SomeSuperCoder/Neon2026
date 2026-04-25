package quanticscript

import (
	"fmt"
	"testing"

	"github.com/poh-blockchain/internal/filestore"
)

// InvocationRecord tracks a single cross-program invocation call
type InvocationRecord struct {
	ProgramID     filestore.FileID
	InvokeData    []byte
	ComputeBudget int64
	Depth         int
}

// InvokableMockContext extends MockExecutionContext with cross-program invocation support
type InvokableMockContext struct {
	MockExecutionContext
	// registeredPrograms maps program IDs to handler functions
	registeredPrograms map[filestore.FileID]func(data []byte, budget int64, depth int) ([]byte, error)
	// declaredPrograms is the list of programs this context is allowed to invoke
	declaredPrograms []filestore.FileID
	// invocationHistory records all invocations for verification
	invocationHistory []InvocationRecord
}

// NewInvokableMockContext creates a new InvokableMockContext
func NewInvokableMockContext() *InvokableMockContext {
	return &InvokableMockContext{
		MockExecutionContext: MockExecutionContext{
			files:   make(map[filestore.FileID]*filestore.File),
			signers: nil,
		},
		registeredPrograms: make(map[filestore.FileID]func(data []byte, budget int64, depth int) ([]byte, error)),
		declaredPrograms:   []filestore.FileID{},
		invocationHistory:  []InvocationRecord{},
	}
}

// RegisterProgram registers a mock program handler for the given program ID
// and adds it to the declared programs list
func (m *InvokableMockContext) RegisterProgram(programID filestore.FileID, handler func(data []byte, budget int64, depth int) ([]byte, error)) {
	m.registeredPrograms[programID] = handler
	m.declaredPrograms = append(m.declaredPrograms, programID)
}

// InvokeProgram implements ExecutionContext.InvokeProgram
func (m *InvokableMockContext) InvokeProgram(programID filestore.FileID, invokeData []byte, computeBudget int64, depth int) ([]byte, error) {
	// Record the invocation
	m.invocationHistory = append(m.invocationHistory, InvocationRecord{
		ProgramID:     programID,
		InvokeData:    invokeData,
		ComputeBudget: computeBudget,
		Depth:         depth,
	})

	handler, ok := m.registeredPrograms[programID]
	if !ok {
		return nil, fmt.Errorf("program not found: %s", programID.String())
	}

	return handler(invokeData, computeBudget, depth)
}

// GetDeclaredPrograms implements ExecutionContext.GetDeclaredPrograms
func (m *InvokableMockContext) GetDeclaredPrograms() []filestore.FileID {
	return m.declaredPrograms
}

// GetInvocationHistory returns the recorded invocation history
func (m *InvokableMockContext) GetInvocationHistory() []InvocationRecord {
	return m.invocationHistory
}

// buildPushFileIDBytes creates a PUSH instruction for a FileID from raw bytes
func buildPushFileIDBytes(id filestore.FileID) []byte {
	return buildPushFileID(id)
}

// buildInvoke builds the bytecode sequence to invoke a program:
// PUSH programID (as FileID), PUSH invokeData, PUSH budget, INVOKE
func buildInvokeSequence(programID filestore.FileID, invokeData []byte, budget int64) []byte {
	var bc []byte
	// The interpreter's execInvoke expects TypeFileID on the stack
	bc = append(bc, buildPushFileIDBytes(programID)...)
	bc = append(bc, buildPushBytes(invokeData)...)
	bc = append(bc, buildPushI64(budget)...)
	bc = append(bc, byte(OpInvoke))
	return bc
}

// TestCrossProgramBasicInvocation tests basic cross-program invocation
func TestCrossProgramBasicInvocation(t *testing.T) {
	t.Run("invoke registered program returns result", func(t *testing.T) {
		ctx := NewInvokableMockContext()

		// Register a mock program that returns a fixed result
		var targetProgramID filestore.FileID
		targetProgramID[31] = 0x10

		expectedResult := []byte("invocation result")
		ctx.RegisterProgram(targetProgramID, func(data []byte, budget int64, depth int) ([]byte, error) {
			return expectedResult, nil
		})

		// Build bytecode: invoke target program, return result
		var bytecode []byte
		bytecode = append(bytecode, buildInvokeSequence(targetProgramID, []byte("input data"), 1000)...)
		bytecode = append(bytecode, byte(OpRet))

		interp := NewBytecodeInterpreter(bytecode, ctx, 100000)
		err := interp.Execute()
		if err != nil {
			t.Fatalf("Execution failed: %v", err)
		}

		if len(interp.stack) != 1 {
			t.Fatalf("Expected 1 value on stack, got %d", len(interp.stack))
		}

		result, err := interp.stack[0].AsBytes()
		if err != nil {
			t.Fatalf("Failed to get result: %v", err)
		}

		if string(result) != string(expectedResult) {
			t.Errorf("Expected result %q, got %q", expectedResult, result)
		}
	})

	t.Run("invoke data is passed correctly to target program", func(t *testing.T) {
		ctx := NewInvokableMockContext()

		var targetProgramID filestore.FileID
		targetProgramID[31] = 0x11

		sentData := []byte("hello from caller")
		var receivedData []byte

		ctx.RegisterProgram(targetProgramID, func(data []byte, budget int64, depth int) ([]byte, error) {
			receivedData = data
			return []byte("ok"), nil
		})

		var bytecode []byte
		bytecode = append(bytecode, buildInvokeSequence(targetProgramID, sentData, 1000)...)
		bytecode = append(bytecode, byte(OpRet))

		interp := NewBytecodeInterpreter(bytecode, ctx, 100000)
		err := interp.Execute()
		if err != nil {
			t.Fatalf("Execution failed: %v", err)
		}

		if string(receivedData) != string(sentData) {
			t.Errorf("Expected invoke data %q, got %q", sentData, receivedData)
		}
	})

	t.Run("invocation is recorded in history", func(t *testing.T) {
		ctx := NewInvokableMockContext()

		var targetProgramID filestore.FileID
		targetProgramID[31] = 0x12

		ctx.RegisterProgram(targetProgramID, func(data []byte, budget int64, depth int) ([]byte, error) {
			return []byte("done"), nil
		})

		var bytecode []byte
		bytecode = append(bytecode, buildInvokeSequence(targetProgramID, []byte("data"), 500)...)
		bytecode = append(bytecode, byte(OpRet))

		interp := NewBytecodeInterpreter(bytecode, ctx, 100000)
		err := interp.Execute()
		if err != nil {
			t.Fatalf("Execution failed: %v", err)
		}

		history := ctx.GetInvocationHistory()
		if len(history) != 1 {
			t.Fatalf("Expected 1 invocation in history, got %d", len(history))
		}

		if history[0].ProgramID != targetProgramID {
			t.Errorf("Expected program ID %v, got %v", targetProgramID, history[0].ProgramID)
		}

		if history[0].ComputeBudget != 500 {
			t.Errorf("Expected compute budget 500, got %d", history[0].ComputeBudget)
		}
	})
}

// TestCrossProgramDepthTracking tests invocation depth tracking
func TestCrossProgramDepthTracking(t *testing.T) {
	t.Run("invocation depth is passed to target program", func(t *testing.T) {
		ctx := NewInvokableMockContext()

		var targetProgramID filestore.FileID
		targetProgramID[31] = 0x20

		var receivedDepth int
		ctx.RegisterProgram(targetProgramID, func(data []byte, budget int64, depth int) ([]byte, error) {
			receivedDepth = depth
			return []byte("ok"), nil
		})

		var bytecode []byte
		bytecode = append(bytecode, buildInvokeSequence(targetProgramID, []byte{}, 1000)...)
		bytecode = append(bytecode, byte(OpRet))

		// Start at depth 0, so invocation should be at depth 1
		interp := NewBytecodeInterpreter(bytecode, ctx, 100000)
		err := interp.Execute()
		if err != nil {
			t.Fatalf("Execution failed: %v", err)
		}

		if receivedDepth != 1 {
			t.Errorf("Expected invocation depth 1, got %d", receivedDepth)
		}
	})

	t.Run("invocation depth increments for nested calls", func(t *testing.T) {
		ctx := NewInvokableMockContext()

		var programA filestore.FileID
		programA[31] = 0x21
		var programB filestore.FileID
		programB[31] = 0x22

		var depthAtB int

		// Program B records its depth
		ctx.RegisterProgram(programB, func(data []byte, budget int64, depth int) ([]byte, error) {
			depthAtB = depth
			return []byte("b result"), nil
		})

		// Program A invokes Program B
		ctx.RegisterProgram(programA, func(data []byte, budget int64, depth int) ([]byte, error) {
			// Simulate program A invoking program B at depth+1
			// We record that A was called at 'depth', B will be called at depth+1
			result, err := ctx.InvokeProgram(programB, []byte{}, budget/2, depth+1)
			if err != nil {
				return nil, err
			}
			return result, nil
		})

		var bytecode []byte
		bytecode = append(bytecode, buildInvokeSequence(programA, []byte{}, 10000)...)
		bytecode = append(bytecode, byte(OpRet))

		interp := NewBytecodeInterpreter(bytecode, ctx, 100000)
		err := interp.Execute()
		if err != nil {
			t.Fatalf("Execution failed: %v", err)
		}

		// A was called at depth 1, B was called at depth 2
		if depthAtB != 2 {
			t.Errorf("Expected program B to be invoked at depth 2, got %d", depthAtB)
		}
	})

	t.Run("maximum depth limit is enforced", func(t *testing.T) {
		ctx := NewInvokableMockContext()

		var targetProgramID filestore.FileID
		targetProgramID[31] = 0x23

		ctx.RegisterProgram(targetProgramID, func(data []byte, budget int64, depth int) ([]byte, error) {
			return []byte("ok"), nil
		})

		// Create interpreter already at max depth
		var bytecode []byte
		bytecode = append(bytecode, buildInvokeSequence(targetProgramID, []byte{}, 1000)...)
		bytecode = append(bytecode, byte(OpRet))

		// Start at MaxInvokeDepth so the next invoke should fail
		interp := NewBytecodeInterpreterWithDepth(bytecode, ctx, 100000, MaxInvokeDepth)
		err := interp.Execute()
		if err == nil {
			t.Fatal("Expected error for exceeding max invocation depth, got nil")
		}
	})

	t.Run("depth tracking accuracy across multiple invocations", func(t *testing.T) {
		ctx := NewInvokableMockContext()

		var prog1 filestore.FileID
		prog1[31] = 0x24
		var prog2 filestore.FileID
		prog2[31] = 0x25

		ctx.RegisterProgram(prog1, func(data []byte, budget int64, depth int) ([]byte, error) {
			return []byte("p1"), nil
		})
		ctx.RegisterProgram(prog2, func(data []byte, budget int64, depth int) ([]byte, error) {
			return []byte("p2"), nil
		})

		// Invoke two programs sequentially
		var bytecode []byte
		bytecode = append(bytecode, buildInvokeSequence(prog1, []byte{}, 1000)...)
		bytecode = append(bytecode, byte(OpPop)) // discard first result
		bytecode = append(bytecode, buildInvokeSequence(prog2, []byte{}, 1000)...)
		bytecode = append(bytecode, byte(OpRet))

		interp := NewBytecodeInterpreter(bytecode, ctx, 100000)
		err := interp.Execute()
		if err != nil {
			t.Fatalf("Execution failed: %v", err)
		}

		history := ctx.GetInvocationHistory()
		if len(history) != 2 {
			t.Fatalf("Expected 2 invocations in history, got %d", len(history))
		}

		// Both should be at depth 1 (called from depth 0)
		if history[0].Depth != 1 {
			t.Errorf("Expected first invocation at depth 1, got %d", history[0].Depth)
		}
		if history[1].Depth != 1 {
			t.Errorf("Expected second invocation at depth 1, got %d", history[1].Depth)
		}
	})
}

// TestCrossProgramPermissions tests invocation permission and budget enforcement
func TestCrossProgramPermissions(t *testing.T) {
	t.Run("invocation fails for undeclared program", func(t *testing.T) {
		ctx := NewInvokableMockContext()

		// Create a program ID that is NOT registered/declared
		var undeclaredProgramID filestore.FileID
		undeclaredProgramID[31] = 0x30

		var bytecode []byte
		bytecode = append(bytecode, buildInvokeSequence(undeclaredProgramID, []byte{}, 1000)...)
		bytecode = append(bytecode, byte(OpRet))

		interp := NewBytecodeInterpreter(bytecode, ctx, 100000)
		err := interp.Execute()
		if err == nil {
			t.Fatal("Expected error for invoking undeclared program, got nil")
		}
	})

	t.Run("compute budget is deducted on invocation", func(t *testing.T) {
		ctx := NewInvokableMockContext()

		var targetProgramID filestore.FileID
		targetProgramID[31] = 0x31

		ctx.RegisterProgram(targetProgramID, func(data []byte, budget int64, depth int) ([]byte, error) {
			return []byte("ok"), nil
		})

		invokeBudget := int64(5000)
		initialBudget := int64(100000)

		var bytecode []byte
		bytecode = append(bytecode, buildInvokeSequence(targetProgramID, []byte{}, invokeBudget)...)
		bytecode = append(bytecode, byte(OpRet))

		interp := NewBytecodeInterpreter(bytecode, ctx, initialBudget)
		budgetBeforeInvoke := interp.GetComputeBudget()

		err := interp.Execute()
		if err != nil {
			t.Fatalf("Execution failed: %v", err)
		}

		budgetAfterInvoke := interp.GetComputeBudget()

		// Budget should have decreased by at least the invokeBudget plus instruction costs
		budgetDeducted := budgetBeforeInvoke - budgetAfterInvoke
		if budgetDeducted < invokeBudget {
			t.Errorf("Expected at least %d budget deducted, got %d", invokeBudget, budgetDeducted)
		}
	})

	t.Run("insufficient budget returns error", func(t *testing.T) {
		ctx := NewInvokableMockContext()

		var targetProgramID filestore.FileID
		targetProgramID[31] = 0x32

		ctx.RegisterProgram(targetProgramID, func(data []byte, budget int64, depth int) ([]byte, error) {
			return []byte("ok"), nil
		})

		// Request more budget than available
		var bytecode []byte
		bytecode = append(bytecode, buildInvokeSequence(targetProgramID, []byte{}, 999999)...)
		bytecode = append(bytecode, byte(OpRet))

		// Give interpreter only 100 units
		interp := NewBytecodeInterpreter(bytecode, ctx, 100)
		err := interp.Execute()
		if err == nil {
			t.Fatal("Expected error for insufficient compute budget, got nil")
		}
	})

	t.Run("invocation failure propagates error", func(t *testing.T) {
		ctx := NewInvokableMockContext()

		var targetProgramID filestore.FileID
		targetProgramID[31] = 0x33

		ctx.RegisterProgram(targetProgramID, func(data []byte, budget int64, depth int) ([]byte, error) {
			return nil, fmt.Errorf("program execution failed")
		})

		var bytecode []byte
		bytecode = append(bytecode, buildInvokeSequence(targetProgramID, []byte{}, 1000)...)
		bytecode = append(bytecode, byte(OpRet))

		interp := NewBytecodeInterpreter(bytecode, ctx, 100000)
		err := interp.Execute()
		if err == nil {
			t.Fatal("Expected error from failed invocation, got nil")
		}
	})
}
