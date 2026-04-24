package access

import (
	"testing"

	"github.com/poh-blockchain/internal/filestore"
	"github.com/poh-blockchain/internal/transaction"
)

func TestNewAccessController(t *testing.T) {
	ac := NewAccessController()

	if ac == nil {
		t.Fatal("NewAccessController returned nil")
	}
	if ac.accessLog == nil {
		t.Error("accessLog not initialized")
	}
	if ac.declaredAccess == nil {
		t.Error("declaredAccess not initialized")
	}
}

func TestSetDeclaredAccess(t *testing.T) {
	ac := NewAccessController()

	fileID1 := filestore.GenerateFileID([]byte("file1"))
	fileID2 := filestore.GenerateFileID([]byte("file2"))

	inputs := map[string]transaction.FileAccess{
		"from": {FileID: fileID1, Permission: transaction.Write},
		"to":   {FileID: fileID2, Permission: transaction.Read},
	}

	ac.SetDeclaredAccess(inputs)

	declared := ac.GetDeclaredAccess()
	if len(declared) != 2 {
		t.Errorf("Expected 2 declared accesses, got %d", len(declared))
	}

	if declared[fileID1] != transaction.Write {
		t.Errorf("Expected Write permission for file1, got %v", declared[fileID1])
	}
	if declared[fileID2] != transaction.Read {
		t.Errorf("Expected Read permission for file2, got %v", declared[fileID2])
	}
}

func TestSetDeclaredAccessMultipleDeclarations(t *testing.T) {
	ac := NewAccessController()

	fileID := filestore.GenerateFileID([]byte("file"))

	// Declare same file multiple times with different permissions
	inputs := map[string]transaction.FileAccess{
		"input1": {FileID: fileID, Permission: transaction.Read},
		"input2": {FileID: fileID, Permission: transaction.Write},
	}

	ac.SetDeclaredAccess(inputs)

	declared := ac.GetDeclaredAccess()
	// Should use the highest permission level (Write)
	if declared[fileID] != transaction.Write {
		t.Errorf("Expected Write permission (highest), got %v", declared[fileID])
	}
}

func TestSetDeclaredAccessClearsPreviousState(t *testing.T) {
	ac := NewAccessController()

	fileID1 := filestore.GenerateFileID([]byte("file1"))
	fileID2 := filestore.GenerateFileID([]byte("file2"))

	// First set
	inputs1 := map[string]transaction.FileAccess{
		"file1": {FileID: fileID1, Permission: transaction.Write},
	}
	ac.SetDeclaredAccess(inputs1)

	// Second set should clear the first
	inputs2 := map[string]transaction.FileAccess{
		"file2": {FileID: fileID2, Permission: transaction.Read},
	}
	ac.SetDeclaredAccess(inputs2)

	declared := ac.GetDeclaredAccess()
	if len(declared) != 1 {
		t.Errorf("Expected 1 declared access after reset, got %d", len(declared))
	}
	if _, exists := declared[fileID1]; exists {
		t.Error("Previous declared access should be cleared")
	}
}

func TestValidateAccessSuccess(t *testing.T) {
	ac := NewAccessController()

	fileID := filestore.GenerateFileID([]byte("file"))

	inputs := map[string]transaction.FileAccess{
		"file": {FileID: fileID, Permission: transaction.Write},
	}
	ac.SetDeclaredAccess(inputs)

	// Write access should be allowed
	err := ac.ValidateAccess(fileID, transaction.Write)
	if err != nil {
		t.Errorf("ValidateAccess failed: %v", err)
	}

	// Read access should also be allowed when Write is declared
	err = ac.ValidateAccess(fileID, transaction.Read)
	if err != nil {
		t.Errorf("ValidateAccess for Read failed when Write declared: %v", err)
	}
}

func TestValidateAccessUndeclaredFile(t *testing.T) {
	ac := NewAccessController()

	fileID1 := filestore.GenerateFileID([]byte("file1"))
	fileID2 := filestore.GenerateFileID([]byte("file2"))

	inputs := map[string]transaction.FileAccess{
		"file1": {FileID: fileID1, Permission: transaction.Write},
	}
	ac.SetDeclaredAccess(inputs)

	// Accessing undeclared file should fail
	err := ac.ValidateAccess(fileID2, transaction.Read)
	if err == nil {
		t.Error("Expected error for undeclared file access")
	}
}

func TestValidateAccessPermissionViolation(t *testing.T) {
	ac := NewAccessController()

	fileID := filestore.GenerateFileID([]byte("file"))

	inputs := map[string]transaction.FileAccess{
		"file": {FileID: fileID, Permission: transaction.Read},
	}
	ac.SetDeclaredAccess(inputs)

	// Write access should fail when only Read is declared
	err := ac.ValidateAccess(fileID, transaction.Write)
	if err == nil {
		t.Error("Expected error for write access when only read declared")
	}
}

func TestRecordAccess(t *testing.T) {
	ac := NewAccessController()

	fileID := filestore.GenerateFileID([]byte("file"))

	ac.RecordAccess(fileID, transaction.Read)

	accessLog := ac.GetAccessLog()
	if len(accessLog) != 1 {
		t.Errorf("Expected 1 access log entry, got %d", len(accessLog))
	}
	if accessLog[fileID] != transaction.Read {
		t.Errorf("Expected Read permission in log, got %v", accessLog[fileID])
	}
}

func TestRecordAccessUpgrade(t *testing.T) {
	ac := NewAccessController()

	fileID := filestore.GenerateFileID([]byte("file"))

	// Record Read first
	ac.RecordAccess(fileID, transaction.Read)

	// Then record Write (should upgrade)
	ac.RecordAccess(fileID, transaction.Write)

	accessLog := ac.GetAccessLog()
	if accessLog[fileID] != transaction.Write {
		t.Errorf("Expected Write permission after upgrade, got %v", accessLog[fileID])
	}
}

func TestRecordAccessNoDowngrade(t *testing.T) {
	ac := NewAccessController()

	fileID := filestore.GenerateFileID([]byte("file"))

	// Record Write first
	ac.RecordAccess(fileID, transaction.Write)

	// Then record Read (should not downgrade)
	ac.RecordAccess(fileID, transaction.Read)

	accessLog := ac.GetAccessLog()
	if accessLog[fileID] != transaction.Write {
		t.Errorf("Expected Write permission (no downgrade), got %v", accessLog[fileID])
	}
}

func TestGetAccessLog(t *testing.T) {
	ac := NewAccessController()

	fileID1 := filestore.GenerateFileID([]byte("file1"))
	fileID2 := filestore.GenerateFileID([]byte("file2"))

	ac.RecordAccess(fileID1, transaction.Read)
	ac.RecordAccess(fileID2, transaction.Write)

	accessLog := ac.GetAccessLog()

	if len(accessLog) != 2 {
		t.Errorf("Expected 2 access log entries, got %d", len(accessLog))
	}
	if accessLog[fileID1] != transaction.Read {
		t.Error("file1 access log incorrect")
	}
	if accessLog[fileID2] != transaction.Write {
		t.Error("file2 access log incorrect")
	}
}

func TestGetAccessLogReturnsCopy(t *testing.T) {
	ac := NewAccessController()

	fileID := filestore.GenerateFileID([]byte("file"))
	ac.RecordAccess(fileID, transaction.Read)

	// Get log and modify it
	accessLog := ac.GetAccessLog()
	accessLog[fileID] = transaction.Write

	// Original should be unchanged
	originalLog := ac.GetAccessLog()
	if originalLog[fileID] != transaction.Read {
		t.Error("GetAccessLog should return a copy, not the original")
	}
}

func TestValidateAndRecordSuccess(t *testing.T) {
	ac := NewAccessController()

	fileID := filestore.GenerateFileID([]byte("file"))

	inputs := map[string]transaction.FileAccess{
		"file": {FileID: fileID, Permission: transaction.Write},
	}
	ac.SetDeclaredAccess(inputs)

	// Should validate and record
	err := ac.ValidateAndRecord(fileID, transaction.Write)
	if err != nil {
		t.Errorf("ValidateAndRecord failed: %v", err)
	}

	// Check that access was recorded
	accessLog := ac.GetAccessLog()
	if accessLog[fileID] != transaction.Write {
		t.Error("Access was not recorded")
	}
}

func TestValidateAndRecordFailure(t *testing.T) {
	ac := NewAccessController()

	fileID := filestore.GenerateFileID([]byte("file"))

	inputs := map[string]transaction.FileAccess{
		"file": {FileID: fileID, Permission: transaction.Read},
	}
	ac.SetDeclaredAccess(inputs)

	// Should fail validation and not record
	err := ac.ValidateAndRecord(fileID, transaction.Write)
	if err == nil {
		t.Error("Expected error for permission violation")
	}

	// Check that access was not recorded
	accessLog := ac.GetAccessLog()
	if _, exists := accessLog[fileID]; exists {
		t.Error("Access should not be recorded on validation failure")
	}
}

func TestReset(t *testing.T) {
	ac := NewAccessController()

	fileID := filestore.GenerateFileID([]byte("file"))

	inputs := map[string]transaction.FileAccess{
		"file": {FileID: fileID, Permission: transaction.Write},
	}
	ac.SetDeclaredAccess(inputs)
	ac.RecordAccess(fileID, transaction.Write)

	// Reset should clear everything
	ac.Reset()

	declared := ac.GetDeclaredAccess()
	accessLog := ac.GetAccessLog()

	if len(declared) != 0 {
		t.Errorf("Expected empty declared access after reset, got %d entries", len(declared))
	}
	if len(accessLog) != 0 {
		t.Errorf("Expected empty access log after reset, got %d entries", len(accessLog))
	}
}

func TestGetDeclaredAccess(t *testing.T) {
	ac := NewAccessController()

	fileID1 := filestore.GenerateFileID([]byte("file1"))
	fileID2 := filestore.GenerateFileID([]byte("file2"))

	inputs := map[string]transaction.FileAccess{
		"file1": {FileID: fileID1, Permission: transaction.Read},
		"file2": {FileID: fileID2, Permission: transaction.Write},
	}
	ac.SetDeclaredAccess(inputs)

	declared := ac.GetDeclaredAccess()

	if len(declared) != 2 {
		t.Errorf("Expected 2 declared accesses, got %d", len(declared))
	}
	if declared[fileID1] != transaction.Read {
		t.Error("file1 declared access incorrect")
	}
	if declared[fileID2] != transaction.Write {
		t.Error("file2 declared access incorrect")
	}
}

func TestGetDeclaredAccessReturnsCopy(t *testing.T) {
	ac := NewAccessController()

	fileID := filestore.GenerateFileID([]byte("file"))

	inputs := map[string]transaction.FileAccess{
		"file": {FileID: fileID, Permission: transaction.Read},
	}
	ac.SetDeclaredAccess(inputs)

	// Get declared and modify it
	declared := ac.GetDeclaredAccess()
	declared[fileID] = transaction.Write

	// Original should be unchanged
	originalDeclared := ac.GetDeclaredAccess()
	if originalDeclared[fileID] != transaction.Read {
		t.Error("GetDeclaredAccess should return a copy, not the original")
	}
}

func TestValidateFinalAccessSuccess(t *testing.T) {
	ac := NewAccessController()

	fileID := filestore.GenerateFileID([]byte("file"))

	inputs := map[string]transaction.FileAccess{
		"file": {FileID: fileID, Permission: transaction.Write},
	}
	ac.SetDeclaredAccess(inputs)
	ac.RecordAccess(fileID, transaction.Write)

	// Should pass validation
	err := ac.ValidateFinalAccess()
	if err != nil {
		t.Errorf("ValidateFinalAccess failed: %v", err)
	}
}

func TestValidateFinalAccessUndeclaredAccess(t *testing.T) {
	ac := NewAccessController()

	fileID1 := filestore.GenerateFileID([]byte("file1"))
	fileID2 := filestore.GenerateFileID([]byte("file2"))

	inputs := map[string]transaction.FileAccess{
		"file1": {FileID: fileID1, Permission: transaction.Write},
	}
	ac.SetDeclaredAccess(inputs)

	// Access file2 which was not declared
	ac.RecordAccess(fileID2, transaction.Read)

	// Should fail validation
	err := ac.ValidateFinalAccess()
	if err == nil {
		t.Error("Expected error for undeclared file access")
	}
}

func TestValidateFinalAccessPermissionViolation(t *testing.T) {
	ac := NewAccessController()

	fileID := filestore.GenerateFileID([]byte("file"))

	inputs := map[string]transaction.FileAccess{
		"file": {FileID: fileID, Permission: transaction.Read},
	}
	ac.SetDeclaredAccess(inputs)

	// Record Write access when only Read was declared
	ac.RecordAccess(fileID, transaction.Write)

	// Should fail validation
	err := ac.ValidateFinalAccess()
	if err == nil {
		t.Error("Expected error for permission violation")
	}
}

func TestValidateFinalAccessReadWithWriteDeclared(t *testing.T) {
	ac := NewAccessController()

	fileID := filestore.GenerateFileID([]byte("file"))

	inputs := map[string]transaction.FileAccess{
		"file": {FileID: fileID, Permission: transaction.Write},
	}
	ac.SetDeclaredAccess(inputs)

	// Record only Read access when Write was declared (should be fine)
	ac.RecordAccess(fileID, transaction.Read)

	// Should pass validation
	err := ac.ValidateFinalAccess()
	if err != nil {
		t.Errorf("ValidateFinalAccess should allow Read when Write declared: %v", err)
	}
}

func TestConcurrentAccess(t *testing.T) {
	ac := NewAccessController()

	fileID1 := filestore.GenerateFileID([]byte("file1"))
	fileID2 := filestore.GenerateFileID([]byte("file2"))

	inputs := map[string]transaction.FileAccess{
		"file1": {FileID: fileID1, Permission: transaction.Write},
		"file2": {FileID: fileID2, Permission: transaction.Read},
	}
	ac.SetDeclaredAccess(inputs)

	// Perform concurrent operations
	done := make(chan bool)
	numGoroutines := 10

	for range numGoroutines {
		go func() {
			defer func() { done <- true }()

			// Concurrent reads should be safe
			_ = ac.ValidateAccess(fileID1, transaction.Write)
			_ = ac.ValidateAccess(fileID2, transaction.Read)
			ac.RecordAccess(fileID1, transaction.Write)
			ac.RecordAccess(fileID2, transaction.Read)
			_ = ac.GetAccessLog()
			_ = ac.GetDeclaredAccess()
		}()
	}

	// Wait for all goroutines
	for range numGoroutines {
		<-done
	}

	// Verify final state is consistent
	accessLog := ac.GetAccessLog()
	if len(accessLog) != 2 {
		t.Errorf("Expected 2 access log entries after concurrent access, got %d", len(accessLog))
	}
}
