package filestore

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestFileID(t *testing.T) {
	// Test FileID creation and string conversion
	data := []byte("test data for file ID generation")
	fid := GenerateFileID(data)

	// Test String() method
	fidStr := fid.String()
	if len(fidStr) != 64 { // 32 bytes = 64 hex characters
		t.Errorf("FileID string length incorrect: expected 64, got %d", len(fidStr))
	}

	// Test FileIDFromString
	fid2, err := FileIDFromString(fidStr)
	if err != nil {
		t.Fatalf("FileIDFromString failed: %v", err)
	}
	if fid != fid2 {
		t.Errorf("FileID mismatch after string conversion")
	}

	// Test FileIDFromBytes
	fid3, err := FileIDFromBytes(fid[:])
	if err != nil {
		t.Fatalf("FileIDFromBytes failed: %v", err)
	}
	if fid != fid3 {
		t.Errorf("FileID mismatch after bytes conversion")
	}
}

func TestFileIDFromStringErrors(t *testing.T) {
	// Test invalid hex string
	_, err := FileIDFromString("invalid hex")
	if err == nil {
		t.Error("Expected error for invalid hex string")
	}

	// Test wrong length
	_, err = FileIDFromString("abcd")
	if err == nil {
		t.Error("Expected error for wrong length")
	}
}

func TestFileIDFromBytesErrors(t *testing.T) {
	// Test wrong length
	_, err := FileIDFromBytes([]byte{1, 2, 3})
	if err == nil {
		t.Error("Expected error for wrong length")
	}
}

func TestFileMarshalUnmarshal(t *testing.T) {
	// Create a test file
	now := time.Now().Truncate(time.Second) // Truncate to second precision
	txManagerID := GenerateFileID([]byte("tx_manager"))
	fileID := GenerateFileID([]byte("test_file"))

	original := &File{
		ID:         fileID,
		Balance:    1000000,
		TxManager:  txManagerID,
		Data:       []byte("test file data"),
		Executable: true,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	// Marshal the file
	data, err := original.Marshal()
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	// Unmarshal the file
	restored := &File{}
	err = restored.Unmarshal(data)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	// Verify all fields match
	if restored.ID != original.ID {
		t.Errorf("ID mismatch: expected %v, got %v", original.ID, restored.ID)
	}
	if restored.Balance != original.Balance {
		t.Errorf("Balance mismatch: expected %d, got %d", original.Balance, restored.Balance)
	}
	if restored.TxManager != original.TxManager {
		t.Errorf("TxManager mismatch: expected %v, got %v", original.TxManager, restored.TxManager)
	}
	if !bytes.Equal(restored.Data, original.Data) {
		t.Errorf("Data mismatch: expected %v, got %v", original.Data, restored.Data)
	}
	if restored.Executable != original.Executable {
		t.Errorf("Executable mismatch: expected %v, got %v", original.Executable, restored.Executable)
	}
	if !restored.CreatedAt.Equal(original.CreatedAt) {
		t.Errorf("CreatedAt mismatch: expected %v, got %v", original.CreatedAt, restored.CreatedAt)
	}
	if !restored.UpdatedAt.Equal(original.UpdatedAt) {
		t.Errorf("UpdatedAt mismatch: expected %v, got %v", original.UpdatedAt, restored.UpdatedAt)
	}
}

func TestUnmarshalFile(t *testing.T) {
	// Create and marshal a test file
	fileID := GenerateFileID([]byte("test"))
	txManagerID := GenerateFileID([]byte("manager"))
	now := time.Now().Truncate(time.Second)

	original := &File{
		ID:         fileID,
		Balance:    5000,
		TxManager:  txManagerID,
		Data:       []byte("data"),
		Executable: false,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	data, err := original.Marshal()
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	// Test UnmarshalFile convenience function
	restored, err := UnmarshalFile(data)
	if err != nil {
		t.Fatalf("UnmarshalFile failed: %v", err)
	}

	if restored.Balance != original.Balance {
		t.Errorf("Balance mismatch: expected %d, got %d", original.Balance, restored.Balance)
	}
}

func TestFileUnmarshalErrors(t *testing.T) {
	// Test invalid protobuf data
	f := &File{}
	err := f.Unmarshal([]byte("invalid protobuf data"))
	if err == nil {
		t.Error("Expected error for invalid protobuf data")
	}
}

// Storage Cost Calculation Tests

func TestCalculateStorageCost(t *testing.T) {
	tests := []struct {
		name     string
		dataSize int64
		expected int64
	}{
		{
			name:     "zero size",
			dataSize: 0,
			expected: 0,
		},
		{
			name:     "1 byte (rounds to 1 KB)",
			dataSize: 1,
			expected: 1000, // 1 KB * 1000 * (1.1^0.0009765625) ≈ 1000
		},
		{
			name:     "1 KB exactly",
			dataSize: 1024,
			expected: 1000, // 1 KB * 1000 * (1.1^0.0009765625) ≈ 1000
		},
		{
			name:     "1 MB exactly",
			dataSize: 1024 * 1024,
			expected: 1126400, // 1024 KB * 1000 * (1.1^1) ≈ 1,126,400
		},
		{
			name:     "10 MB",
			dataSize: 10 * 1024 * 1024,
			expected: 26577920, // 10240 KB * 1000 * (1.1^10) ≈ 26,577,920
		},
		{
			name:     "512 bytes (rounds to 1 KB)",
			dataSize: 512,
			expected: 1000,
		},
		{
			name:     "1025 bytes (rounds to 2 KB)",
			dataSize: 1025,
			expected: 2000, // 2 KB * 1000 * (1.1^0.001953125) ≈ 2000
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cost := CalculateStorageCost(tt.dataSize)

			// Allow for small floating point rounding differences (within 1%)
			tolerance := float64(tt.expected) * 0.01
			diff := float64(cost - tt.expected)
			if diff < 0 {
				diff = -diff
			}

			if diff > tolerance {
				t.Errorf("CalculateStorageCost(%d) = %d, expected approximately %d (diff: %f, tolerance: %f)",
					tt.dataSize, cost, tt.expected, diff, tolerance)
			}
		})
	}
}

func TestCalculateStorageCostNegative(t *testing.T) {
	cost := CalculateStorageCost(-100)
	if cost != 0 {
		t.Errorf("Expected 0 cost for negative size, got %d", cost)
	}
}

func TestValidateStorageCost(t *testing.T) {
	tests := []struct {
		name        string
		file        *File
		expectError bool
	}{
		{
			name: "sufficient balance for 1 KB",
			file: &File{
				Balance: 2000,
				Data:    make([]byte, 1024),
			},
			expectError: false,
		},
		{
			name: "insufficient balance",
			file: &File{
				Balance: 500,
				Data:    make([]byte, 1024),
			},
			expectError: true,
		},
		{
			name: "exact balance",
			file: &File{
				Balance: 1000,
				Data:    make([]byte, 1024),
			},
			expectError: false,
		},
		{
			name: "empty data",
			file: &File{
				Balance: 0,
				Data:    []byte{},
			},
			expectError: false,
		},
		{
			name: "large file with sufficient balance",
			file: &File{
				Balance: 30000000,
				Data:    make([]byte, 10*1024*1024), // 10 MB
			},
			expectError: false,
		},
		{
			name: "large file with insufficient balance",
			file: &File{
				Balance: 1000000,
				Data:    make([]byte, 10*1024*1024), // 10 MB
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateStorageCost(tt.file)
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

func TestValidateStorageCostNilFile(t *testing.T) {
	err := ValidateStorageCost(nil)
	if err == nil {
		t.Error("Expected error for nil file")
	}
}

func TestValidateBalanceForDataSize(t *testing.T) {
	tests := []struct {
		name        string
		balance     int64
		dataSize    int64
		expectError bool
	}{
		{
			name:        "sufficient balance",
			balance:     2000,
			dataSize:    1024,
			expectError: false,
		},
		{
			name:        "insufficient balance",
			balance:     500,
			dataSize:    1024,
			expectError: true,
		},
		{
			name:        "exact balance",
			balance:     1000,
			dataSize:    1024,
			expectError: false,
		},
		{
			name:        "zero data size",
			balance:     0,
			dataSize:    0,
			expectError: false,
		},
		{
			name:        "large data size with sufficient balance",
			balance:     30000000,
			dataSize:    10 * 1024 * 1024,
			expectError: false,
		},
		{
			name:        "large data size with insufficient balance",
			balance:     1000000,
			dataSize:    10 * 1024 * 1024,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateBalanceForDataSize(tt.balance, tt.dataSize)
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

func TestBytesToKB(t *testing.T) {
	tests := []struct {
		bytes    int64
		expected int64
	}{
		{0, 0},
		{1, 1},    // Rounds up
		{1023, 1}, // Rounds up
		{1024, 1}, // Exact
		{1025, 2}, // Rounds up
		{2048, 2}, // Exact
		{2049, 3}, // Rounds up
		{-100, 0}, // Negative
	}

	for _, tt := range tests {
		result := BytesToKB(tt.bytes)
		if result != tt.expected {
			t.Errorf("BytesToKB(%d) = %d, expected %d", tt.bytes, result, tt.expected)
		}
	}
}

func TestKBToMB(t *testing.T) {
	tests := []struct {
		kb       int64
		expected float64
	}{
		{0, 0.0},
		{1, 0.0009765625}, // 1/1024
		{1024, 1.0},       // Exact 1 MB
		{2048, 2.0},       // Exact 2 MB
		{512, 0.5},        // Half MB
	}

	for _, tt := range tests {
		result := KBToMB(tt.kb)
		if result != tt.expected {
			t.Errorf("KBToMB(%d) = %f, expected %f", tt.kb, result, tt.expected)
		}
	}
}

func TestMBToKB(t *testing.T) {
	tests := []struct {
		mb       float64
		expected int64
	}{
		{0.0, 0},
		{1.0, 1024},
		{2.0, 2048},
		{0.5, 512},
		{10.0, 10240},
	}

	for _, tt := range tests {
		result := MBToKB(tt.mb)
		if result != tt.expected {
			t.Errorf("MBToKB(%f) = %d, expected %d", tt.mb, result, tt.expected)
		}
	}
}

func TestBytesToMB(t *testing.T) {
	tests := []struct {
		bytes    int64
		expected float64
	}{
		{0, 0.0},
		{1024 * 1024, 1.0},       // 1 MB
		{2 * 1024 * 1024, 2.0},   // 2 MB
		{512 * 1024, 0.5},        // 0.5 MB
		{10 * 1024 * 1024, 10.0}, // 10 MB
	}

	for _, tt := range tests {
		result := BytesToMB(tt.bytes)
		if result != tt.expected {
			t.Errorf("BytesToMB(%d) = %f, expected %f", tt.bytes, result, tt.expected)
		}
	}
}

func TestStorageCostExponentialGrowth(t *testing.T) {
	// Verify that storage cost grows exponentially with size
	cost1MB := CalculateStorageCost(1 * 1024 * 1024)
	cost2MB := CalculateStorageCost(2 * 1024 * 1024)
	cost5MB := CalculateStorageCost(5 * 1024 * 1024)
	cost10MB := CalculateStorageCost(10 * 1024 * 1024)

	// Cost should grow faster than linear
	// For linear growth: cost2MB should be ~2x cost1MB
	// For exponential: cost2MB should be > 2x cost1MB
	linearRatio := float64(cost2MB) / float64(cost1MB)
	if linearRatio <= 2.0 {
		t.Errorf("Expected exponential growth, but 2MB/1MB ratio is %f (should be > 2.0)", linearRatio)
	}

	// Verify costs are increasing
	if cost1MB >= cost2MB || cost2MB >= cost5MB || cost5MB >= cost10MB {
		t.Errorf("Storage costs should increase: 1MB=%d, 2MB=%d, 5MB=%d, 10MB=%d",
			cost1MB, cost2MB, cost5MB, cost10MB)
	}
}

// FileStore Tests

func TestNewFileStore(t *testing.T) {
	// Create a temporary directory for the test database
	dbPath := t.TempDir() + "/testdb"

	fs, err := NewFileStore(dbPath)
	if err != nil {
		t.Fatalf("NewFileStore failed: %v", err)
	}
	defer fs.Close()

	if fs.db == nil {
		t.Error("FileStore database is nil")
	}
	if fs.cache == nil {
		t.Error("FileStore cache is nil")
	}
}

func TestNewReadOnlyFileStore(t *testing.T) {
	// Create a temporary directory for the test database
	dbPath := t.TempDir() + "/testdb"

	// First create a regular FileStore and add some data
	fs, err := NewFileStore(dbPath)
	if err != nil {
		t.Fatalf("NewFileStore failed: %v", err)
	}

	// Create a test file
	txManagerID := GenerateFileID([]byte("tx_manager"))
	file := &File{
		Balance:    10000,
		TxManager:  txManagerID,
		Data:       []byte("test data"),
		Executable: false,
	}

	fileID, err := fs.CreateFile(file)
	if err != nil {
		t.Fatalf("CreateFile failed: %v", err)
	}

	// Close the regular FileStore
	fs.Close()

	// Now open as read-only
	readOnlyFS, err := NewReadOnlyFileStore(dbPath)
	if err != nil {
		t.Fatalf("NewReadOnlyFileStore failed: %v", err)
	}
	defer readOnlyFS.Close()

	if readOnlyFS.db == nil {
		t.Error("ReadOnly FileStore database is nil")
	}
	if readOnlyFS.cache == nil {
		t.Error("ReadOnly FileStore cache is nil")
	}

	// Verify we can read the existing file
	retrievedFile, err := readOnlyFS.GetFile(fileID)
	if err != nil {
		t.Fatalf("GetFile from read-only store failed: %v", err)
	}

	if retrievedFile.Balance != 10000 {
		t.Errorf("Balance mismatch: expected 10000, got %d", retrievedFile.Balance)
	}
	if string(retrievedFile.Data) != "test data" {
		t.Errorf("Data mismatch: expected 'test data', got '%s'", string(retrievedFile.Data))
	}
}

func TestNewReadOnlyFileStore_NonExistentDB(t *testing.T) {
	// Try to open a non-existent database in read-only mode
	dbPath := t.TempDir() + "/nonexistent"

	_, err := NewReadOnlyFileStore(dbPath)
	if err == nil {
		t.Error("Expected error for non-existent database in read-only mode")
	}
}

func TestReadOnlyFileStore_WriteOperationsShouldFail(t *testing.T) {
	// Create a temporary directory for the test database
	dbPath := t.TempDir() + "/testdb"

	// First create a regular FileStore
	fs, err := NewFileStore(dbPath)
	if err != nil {
		t.Fatalf("NewFileStore failed: %v", err)
	}
	fs.Close()

	// Open as read-only
	readOnlyFS, err := NewReadOnlyFileStore(dbPath)
	if err != nil {
		t.Fatalf("NewReadOnlyFileStore failed: %v", err)
	}
	defer readOnlyFS.Close()

	// Try to create a file (should fail)
	txManagerID := GenerateFileID([]byte("tx_manager"))
	file := &File{
		Balance:    10000,
		TxManager:  txManagerID,
		Data:       []byte("test data"),
		Executable: false,
	}

	_, err = readOnlyFS.CreateFile(file)
	if err == nil {
		t.Error("Expected error when trying to create file in read-only store")
	}

	// The error should indicate read-only mode
	if !strings.Contains(err.Error(), "read-only") && !strings.Contains(err.Error(), "readonly") {
		t.Errorf("Expected read-only error, got: %v", err)
	}
}

func TestFileStoreCreateFile(t *testing.T) {
	dbPath := t.TempDir() + "/testdb"
	fs, err := NewFileStore(dbPath)
	if err != nil {
		t.Fatalf("NewFileStore failed: %v", err)
	}
	defer fs.Close()

	txManagerID := GenerateFileID([]byte("tx_manager"))
	file := &File{
		Balance:    10000,
		TxManager:  txManagerID,
		Data:       []byte("test data"),
		Executable: false,
	}

	fileID, err := fs.CreateFile(file)
	if err != nil {
		t.Fatalf("CreateFile failed: %v", err)
	}

	if fileID == (FileID{}) {
		t.Error("CreateFile returned empty FileID")
	}

	// Verify file was created with correct ID
	if file.ID != fileID {
		t.Errorf("File ID mismatch: expected %s, got %s", fileID.String(), file.ID.String())
	}

	// Verify timestamps were set
	if file.CreatedAt.IsZero() {
		t.Error("CreatedAt timestamp not set")
	}
	if file.UpdatedAt.IsZero() {
		t.Error("UpdatedAt timestamp not set")
	}
}

func TestFileStoreCreateFileInsufficientBalance(t *testing.T) {
	dbPath := t.TempDir() + "/testdb"
	fs, err := NewFileStore(dbPath)
	if err != nil {
		t.Fatalf("NewFileStore failed: %v", err)
	}
	defer fs.Close()

	txManagerID := GenerateFileID([]byte("tx_manager"))
	file := &File{
		Balance:    100, // Insufficient for 1 KB of data
		TxManager:  txManagerID,
		Data:       make([]byte, 1024),
		Executable: false,
	}

	_, err = fs.CreateFile(file)
	if err == nil {
		t.Error("Expected error for insufficient balance, got none")
	}
}

func TestFileStoreGetFile(t *testing.T) {
	dbPath := t.TempDir() + "/testdb"
	fs, err := NewFileStore(dbPath)
	if err != nil {
		t.Fatalf("NewFileStore failed: %v", err)
	}
	defer fs.Close()

	// Create a file
	txManagerID := GenerateFileID([]byte("tx_manager"))
	originalFile := &File{
		Balance:    10000,
		TxManager:  txManagerID,
		Data:       []byte("test data"),
		Executable: true,
	}

	fileID, err := fs.CreateFile(originalFile)
	if err != nil {
		t.Fatalf("CreateFile failed: %v", err)
	}

	// Retrieve the file
	retrievedFile, err := fs.GetFile(fileID)
	if err != nil {
		t.Fatalf("GetFile failed: %v", err)
	}

	// Verify all fields match
	if retrievedFile.ID != originalFile.ID {
		t.Errorf("ID mismatch")
	}
	if retrievedFile.Balance != originalFile.Balance {
		t.Errorf("Balance mismatch: expected %d, got %d", originalFile.Balance, retrievedFile.Balance)
	}
	if retrievedFile.TxManager != originalFile.TxManager {
		t.Errorf("TxManager mismatch")
	}
	if !bytes.Equal(retrievedFile.Data, originalFile.Data) {
		t.Errorf("Data mismatch")
	}
	if retrievedFile.Executable != originalFile.Executable {
		t.Errorf("Executable mismatch")
	}
}

func TestFileStoreGetFileCaching(t *testing.T) {
	dbPath := t.TempDir() + "/testdb"
	fs, err := NewFileStore(dbPath)
	if err != nil {
		t.Fatalf("NewFileStore failed: %v", err)
	}
	defer fs.Close()

	// Create a file
	txManagerID := GenerateFileID([]byte("tx_manager"))
	file := &File{
		Balance:    10000,
		TxManager:  txManagerID,
		Data:       []byte("test data"),
		Executable: false,
	}

	fileID, err := fs.CreateFile(file)
	if err != nil {
		t.Fatalf("CreateFile failed: %v", err)
	}

	// First retrieval (from cache after creation)
	file1, err := fs.GetFile(fileID)
	if err != nil {
		t.Fatalf("First GetFile failed: %v", err)
	}

	// Clear cache to test database retrieval
	fs.mu.Lock()
	delete(fs.cache, fileID)
	fs.mu.Unlock()

	// Second retrieval (from database)
	file2, err := fs.GetFile(fileID)
	if err != nil {
		t.Fatalf("Second GetFile failed: %v", err)
	}

	// Verify both retrievals return the same data
	if file1.Balance != file2.Balance {
		t.Errorf("Balance mismatch between cached and database retrieval")
	}
}

func TestFileStoreGetFileNotFound(t *testing.T) {
	dbPath := t.TempDir() + "/testdb"
	fs, err := NewFileStore(dbPath)
	if err != nil {
		t.Fatalf("NewFileStore failed: %v", err)
	}
	defer fs.Close()

	// Try to get a non-existent file
	nonExistentID := GenerateFileID([]byte("non-existent"))
	_, err = fs.GetFile(nonExistentID)
	if err == nil {
		t.Error("Expected error for non-existent file, got none")
	}
}

func TestFileStoreUpdateFile(t *testing.T) {
	dbPath := t.TempDir() + "/testdb"
	fs, err := NewFileStore(dbPath)
	if err != nil {
		t.Fatalf("NewFileStore failed: %v", err)
	}
	defer fs.Close()

	// Create a file
	txManagerID := GenerateFileID([]byte("tx_manager"))
	file := &File{
		Balance:    10000,
		TxManager:  txManagerID,
		Data:       []byte("original data"),
		Executable: false,
	}

	fileID, err := fs.CreateFile(file)
	if err != nil {
		t.Fatalf("CreateFile failed: %v", err)
	}

	// Update the file
	file.Data = []byte("updated data")
	file.Balance = 15000

	err = fs.UpdateFile(fileID, file)
	if err != nil {
		t.Fatalf("UpdateFile failed: %v", err)
	}

	// Retrieve and verify the update
	updatedFile, err := fs.GetFile(fileID)
	if err != nil {
		t.Fatalf("GetFile failed: %v", err)
	}

	if !bytes.Equal(updatedFile.Data, []byte("updated data")) {
		t.Errorf("Data not updated: got %s", string(updatedFile.Data))
	}
	if updatedFile.Balance != 15000 {
		t.Errorf("Balance not updated: expected 15000, got %d", updatedFile.Balance)
	}
}

func TestFileStoreUpdateFileInsufficientBalance(t *testing.T) {
	dbPath := t.TempDir() + "/testdb"
	fs, err := NewFileStore(dbPath)
	if err != nil {
		t.Fatalf("NewFileStore failed: %v", err)
	}
	defer fs.Close()

	// Create a file
	txManagerID := GenerateFileID([]byte("tx_manager"))
	file := &File{
		Balance:    10000,
		TxManager:  txManagerID,
		Data:       []byte("data"),
		Executable: false,
	}

	fileID, err := fs.CreateFile(file)
	if err != nil {
		t.Fatalf("CreateFile failed: %v", err)
	}

	// Try to update with insufficient balance for larger data
	file.Data = make([]byte, 10*1024*1024) // 10 MB
	file.Balance = 1000                    // Insufficient

	err = fs.UpdateFile(fileID, file)
	if err == nil {
		t.Error("Expected error for insufficient balance, got none")
	}
}

func TestFileStoreUpdateFileNotFound(t *testing.T) {
	dbPath := t.TempDir() + "/testdb"
	fs, err := NewFileStore(dbPath)
	if err != nil {
		t.Fatalf("NewFileStore failed: %v", err)
	}
	defer fs.Close()

	// Try to update a non-existent file
	txManagerID := GenerateFileID([]byte("tx_manager"))
	nonExistentID := GenerateFileID([]byte("non-existent"))
	file := &File{
		ID:         nonExistentID,
		Balance:    10000,
		TxManager:  txManagerID,
		Data:       []byte("data"),
		Executable: false,
	}

	err = fs.UpdateFile(nonExistentID, file)
	if err == nil {
		t.Error("Expected error for non-existent file, got none")
	}
}

func TestFileStoreDeleteFile(t *testing.T) {
	dbPath := t.TempDir() + "/testdb"
	fs, err := NewFileStore(dbPath)
	if err != nil {
		t.Fatalf("NewFileStore failed: %v", err)
	}
	defer fs.Close()

	// Create a file
	txManagerID := GenerateFileID([]byte("tx_manager"))
	file := &File{
		Balance:    10000,
		TxManager:  txManagerID,
		Data:       []byte("test data"),
		Executable: false,
	}

	fileID, err := fs.CreateFile(file)
	if err != nil {
		t.Fatalf("CreateFile failed: %v", err)
	}

	// Delete the file
	err = fs.DeleteFile(fileID)
	if err != nil {
		t.Fatalf("DeleteFile failed: %v", err)
	}

	// Verify file is deleted
	_, err = fs.GetFile(fileID)
	if err == nil {
		t.Error("Expected error for deleted file, got none")
	}
}

func TestFileStoreDeleteFileNotFound(t *testing.T) {
	dbPath := t.TempDir() + "/testdb"
	fs, err := NewFileStore(dbPath)
	if err != nil {
		t.Fatalf("NewFileStore failed: %v", err)
	}
	defer fs.Close()

	// Try to delete a non-existent file
	nonExistentID := GenerateFileID([]byte("non-existent"))
	err = fs.DeleteFile(nonExistentID)
	if err == nil {
		t.Error("Expected error for non-existent file, got none")
	}
}

func TestFileStoreGetFileBalance(t *testing.T) {
	dbPath := t.TempDir() + "/testdb"
	fs, err := NewFileStore(dbPath)
	if err != nil {
		t.Fatalf("NewFileStore failed: %v", err)
	}
	defer fs.Close()

	// Create a file
	txManagerID := GenerateFileID([]byte("tx_manager"))
	file := &File{
		Balance:    12345,
		TxManager:  txManagerID,
		Data:       []byte("test data"),
		Executable: false,
	}

	fileID, err := fs.CreateFile(file)
	if err != nil {
		t.Fatalf("CreateFile failed: %v", err)
	}

	// Get balance
	balance, err := fs.GetFileBalance(fileID)
	if err != nil {
		t.Fatalf("GetFileBalance failed: %v", err)
	}

	if balance != 12345 {
		t.Errorf("Balance mismatch: expected 12345, got %d", balance)
	}
}

func TestFileStoreUpdateFileBalance(t *testing.T) {
	dbPath := t.TempDir() + "/testdb"
	fs, err := NewFileStore(dbPath)
	if err != nil {
		t.Fatalf("NewFileStore failed: %v", err)
	}
	defer fs.Close()

	// Create a file
	txManagerID := GenerateFileID([]byte("tx_manager"))
	file := &File{
		Balance:    10000,
		TxManager:  txManagerID,
		Data:       []byte("test data"),
		Executable: false,
	}

	fileID, err := fs.CreateFile(file)
	if err != nil {
		t.Fatalf("CreateFile failed: %v", err)
	}

	// Update balance (increase)
	err = fs.UpdateFileBalance(fileID, 5000)
	if err != nil {
		t.Fatalf("UpdateFileBalance (increase) failed: %v", err)
	}

	balance, err := fs.GetFileBalance(fileID)
	if err != nil {
		t.Fatalf("GetFileBalance failed: %v", err)
	}
	if balance != 15000 {
		t.Errorf("Balance after increase: expected 15000, got %d", balance)
	}

	// Update balance (decrease)
	err = fs.UpdateFileBalance(fileID, -3000)
	if err != nil {
		t.Fatalf("UpdateFileBalance (decrease) failed: %v", err)
	}

	balance, err = fs.GetFileBalance(fileID)
	if err != nil {
		t.Fatalf("GetFileBalance failed: %v", err)
	}
	if balance != 12000 {
		t.Errorf("Balance after decrease: expected 12000, got %d", balance)
	}
}

func TestFileStoreUpdateFileBalanceInsufficientForStorage(t *testing.T) {
	dbPath := t.TempDir() + "/testdb"
	fs, err := NewFileStore(dbPath)
	if err != nil {
		t.Fatalf("NewFileStore failed: %v", err)
	}
	defer fs.Close()

	// Create a file with 1 KB of data
	txManagerID := GenerateFileID([]byte("tx_manager"))
	file := &File{
		Balance:    10000,
		TxManager:  txManagerID,
		Data:       make([]byte, 1024), // 1 KB requires ~1000 units
		Executable: false,
	}

	fileID, err := fs.CreateFile(file)
	if err != nil {
		t.Fatalf("CreateFile failed: %v", err)
	}

	// Try to decrease balance below storage cost
	err = fs.UpdateFileBalance(fileID, -9500) // Would leave 500, but need ~1000
	if err == nil {
		t.Error("Expected error for balance below storage cost, got none")
	}
}

func TestFileStoreThreadSafety(t *testing.T) {
	dbPath := t.TempDir() + "/testdb"
	fs, err := NewFileStore(dbPath)
	if err != nil {
		t.Fatalf("NewFileStore failed: %v", err)
	}
	defer fs.Close()

	// Create multiple files for concurrent access
	txManagerID := GenerateFileID([]byte("tx_manager"))
	numFiles := 10
	fileIDs := make([]FileID, numFiles)

	for i := 0; i < numFiles; i++ {
		file := &File{
			Balance:    100000,
			TxManager:  txManagerID,
			Data:       []byte("test data"),
			Executable: false,
		}
		fileID, err := fs.CreateFile(file)
		if err != nil {
			t.Fatalf("CreateFile failed: %v", err)
		}
		fileIDs[i] = fileID
	}

	// Perform concurrent reads on different files (should not conflict)
	done := make(chan bool)
	numGoroutines := 10

	for i := 0; i < numGoroutines; i++ {
		go func(n int) {
			defer func() { done <- true }()

			// Read different files concurrently
			fileID := fileIDs[n]
			_, err := fs.GetFile(fileID)
			if err != nil {
				t.Errorf("Concurrent GetFile failed: %v", err)
			}

			// Get balance
			_, err = fs.GetFileBalance(fileID)
			if err != nil {
				t.Errorf("Concurrent GetFileBalance failed: %v", err)
			}
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}
}
