package quanticscript

import (
	"testing"

	"github.com/poh-blockchain/internal/filestore"
)

// TestTransferOpcode tests the TRANSFER opcode with various scenarios
func TestTransferOpcode(t *testing.T) {
	t.Run("SuccessfulTransfer", func(t *testing.T) {
		// Create a mock execution context
		ctx := NewMockExecutionContext()

		// Create source and destination files
		sourceFileID := filestore.FileID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1}
		destFileID := filestore.FileID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 2}
		programID := filestore.FileID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 3}

		// Source file owned by the program with sufficient balance
		sourceFile := &filestore.File{
			ID:        sourceFileID,
			Balance:   10000,
			TxManager: programID,
			Data:      []byte{1, 2, 3}, // 3 bytes of data
		}

		// Destination file
		destFile := &filestore.File{
			ID:        destFileID,
			Balance:   5000,
			TxManager: programID,
			Data:      []byte{},
		}

		ctx.files[sourceFileID] = sourceFile
		ctx.files[destFileID] = destFile
		ctx.programID = programID

		// Create bytecode: PUSH sourceFileID, PUSH destFileID, PUSH amount, TRANSFER
		bytecode := []byte{
			byte(OpPush), byte(TypeFileID),
		}
		bytecode = append(bytecode, sourceFileID[:]...)
		bytecode = append(bytecode, byte(OpPush), byte(TypeFileID))
		bytecode = append(bytecode, destFileID[:]...)
		bytecode = append(bytecode, byte(OpPush), byte(TypeI64))
		bytecode = append(bytecode, 0xE8, 0x03, 0, 0, 0, 0, 0, 0) // 1000 in little-endian
		bytecode = append(bytecode, byte(OpTransfer))
		bytecode = append(bytecode, byte(OpRet))

		// Execute
		interp := NewBytecodeInterpreter(bytecode, ctx, 100000)
		err := interp.Execute()

		if err != nil {
			t.Fatalf("Transfer failed: %v", err)
		}

		// Verify balances
		if sourceFile.Balance != 9000 {
			t.Errorf("Expected source balance 9000, got %d", sourceFile.Balance)
		}
		if destFile.Balance != 6000 {
			t.Errorf("Expected dest balance 6000, got %d", destFile.Balance)
		}
	})

	t.Run("OwnershipViolation", func(t *testing.T) {
		ctx := NewMockExecutionContext()

		sourceFileID := filestore.FileID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1}
		destFileID := filestore.FileID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 2}
		programID := filestore.FileID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 3}
		otherProgramID := filestore.FileID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 4}

		// Source file owned by a DIFFERENT program
		sourceFile := &filestore.File{
			ID:        sourceFileID,
			Balance:   10000,
			TxManager: otherProgramID, // Different owner!
			Data:      []byte{},
		}

		destFile := &filestore.File{
			ID:        destFileID,
			Balance:   5000,
			TxManager: programID,
			Data:      []byte{},
		}

		ctx.files[sourceFileID] = sourceFile
		ctx.files[destFileID] = destFile
		ctx.programID = programID

		// Create bytecode
		bytecode := []byte{
			byte(OpPush), byte(TypeFileID),
		}
		bytecode = append(bytecode, sourceFileID[:]...)
		bytecode = append(bytecode, byte(OpPush), byte(TypeFileID))
		bytecode = append(bytecode, destFileID[:]...)
		bytecode = append(bytecode, byte(OpPush), byte(TypeI64))
		bytecode = append(bytecode, 0xE8, 0x03, 0, 0, 0, 0, 0, 0) // 1000
		bytecode = append(bytecode, byte(OpTransfer))
		bytecode = append(bytecode, byte(OpRet))

		// Execute
		interp := NewBytecodeInterpreter(bytecode, ctx, 100000)
		err := interp.Execute()

		// Should fail with ownership error
		if err == nil {
			t.Fatal("Expected ownership violation error, got nil")
		}

		// Verify balances unchanged
		if sourceFile.Balance != 10000 {
			t.Errorf("Source balance should be unchanged, got %d", sourceFile.Balance)
		}
		if destFile.Balance != 5000 {
			t.Errorf("Dest balance should be unchanged, got %d", destFile.Balance)
		}
	})

	t.Run("StorageCostViolation", func(t *testing.T) {
		ctx := NewMockExecutionContext()

		sourceFileID := filestore.FileID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1}
		destFileID := filestore.FileID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 2}
		programID := filestore.FileID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 3}

		// Source file with data that requires storage cost
		data := make([]byte, 1024) // 1KB of data
		storageCost := filestore.CalculateStorageCost(int64(len(data)))

		sourceFile := &filestore.File{
			ID:        sourceFileID,
			Balance:   storageCost + 100, // Just slightly above storage cost
			TxManager: programID,
			Data:      data,
		}

		destFile := &filestore.File{
			ID:        destFileID,
			Balance:   5000,
			TxManager: programID,
			Data:      []byte{},
		}

		ctx.files[sourceFileID] = sourceFile
		ctx.files[destFileID] = destFile
		ctx.programID = programID

		// Try to transfer more than available (would violate storage cost)
		bytecode := []byte{
			byte(OpPush), byte(TypeFileID),
		}
		bytecode = append(bytecode, sourceFileID[:]...)
		bytecode = append(bytecode, byte(OpPush), byte(TypeFileID))
		bytecode = append(bytecode, destFileID[:]...)
		bytecode = append(bytecode, byte(OpPush), byte(TypeI64))
		bytecode = append(bytecode, 0xC8, 0, 0, 0, 0, 0, 0, 0) // 200 (more than available)
		bytecode = append(bytecode, byte(OpTransfer))
		bytecode = append(bytecode, byte(OpRet))

		// Execute
		interp := NewBytecodeInterpreter(bytecode, ctx, 100000)
		err := interp.Execute()

		// Should fail with storage cost error
		if err == nil {
			t.Fatal("Expected storage cost violation error, got nil")
		}
	})

	t.Run("NegativeAmountRejection", func(t *testing.T) {
		ctx := NewMockExecutionContext()

		sourceFileID := filestore.FileID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1}
		destFileID := filestore.FileID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 2}
		programID := filestore.FileID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 3}

		sourceFile := &filestore.File{
			ID:        sourceFileID,
			Balance:   10000,
			TxManager: programID,
			Data:      []byte{},
		}

		destFile := &filestore.File{
			ID:        destFileID,
			Balance:   5000,
			TxManager: programID,
			Data:      []byte{},
		}

		ctx.files[sourceFileID] = sourceFile
		ctx.files[destFileID] = destFile
		ctx.programID = programID

		// Try to transfer negative amount
		bytecode := []byte{
			byte(OpPush), byte(TypeFileID),
		}
		bytecode = append(bytecode, sourceFileID[:]...)
		bytecode = append(bytecode, byte(OpPush), byte(TypeFileID))
		bytecode = append(bytecode, destFileID[:]...)
		bytecode = append(bytecode, byte(OpPush), byte(TypeI64))
		bytecode = append(bytecode, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF) // -1
		bytecode = append(bytecode, byte(OpTransfer))
		bytecode = append(bytecode, byte(OpRet))

		// Execute
		interp := NewBytecodeInterpreter(bytecode, ctx, 100000)
		err := interp.Execute()

		// Should fail
		if err == nil {
			t.Fatal("Expected negative amount error, got nil")
		}
	})
}
