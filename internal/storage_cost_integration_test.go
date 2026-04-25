package internal

import (
	"crypto/ed25519"
	"encoding/binary"
	"os"
	"testing"

	"github.com/poh-blockchain/internal/filestore"
	"github.com/poh-blockchain/internal/genesis"
	"github.com/poh-blockchain/internal/processor"
	"github.com/poh-blockchain/internal/runtime"
	"github.com/poh-blockchain/internal/transaction"
	"github.com/poh-blockchain/programs"
)

// encodeTransferInstructionSC encodes a transfer instruction payload.
// Format: [type(1), amount_le(8), from_fileID(32), to_fileID(32)] = 73 bytes
func encodeTransferInstructionSC(amount int64, from, to filestore.FileID) []byte {
	data := make([]byte, 73)
	data[0] = 1 // Transfer opcode
	binary.LittleEndian.PutUint64(data[1:9], uint64(amount))
	copy(data[9:41], from[:])
	copy(data[41:73], to[:])
	return data
}

// encodeAllocateDataInstruction encodes an AllocateData instruction payload.
func encodeAllocateDataInstruction(size int64) []byte {
	data := make([]byte, 9)
	data[0] = 3 // AllocateData opcode
	binary.LittleEndian.PutUint64(data[1:], uint64(size))
	return data
}

// TestFileCreationWithInsufficientBalance tests that file creation fails when balance is insufficient for storage cost
// This validates Requirements 6.1, 6.2, 6.5
func TestFileCreationWithInsufficientBalance(t *testing.T) {
	// Create temporary database
	dbPath := t.TempDir() + "/test_insufficient_balance.db"
	defer os.RemoveAll(dbPath)

	fs, err := filestore.NewFileStore(dbPath)
	if err != nil {
		t.Fatalf("Failed to create FileStore: %v", err)
	}
	defer fs.Close()

	// Test 1: Create file with 1 KB data but insufficient balance
	t.Run("1KB data with 500 balance", func(t *testing.T) {
		file := &filestore.File{
			Balance:    500,                // Insufficient (need ~1000)
			TxManager:  filestore.FileID{}, // Empty for test
			Data:       make([]byte, 1024), // 1 KB
			Executable: false,
		}

		_, err := fs.CreateFile(file)
		if err == nil {
			t.Error("Expected error for insufficient balance, got none")
		}
		t.Logf("Correctly rejected file creation with insufficient balance: %v", err)
	})

	// Test 2: Create file with 10 KB data but insufficient balance
	t.Run("10KB data with 5000 balance", func(t *testing.T) {
		file := &filestore.File{
			Balance:    5000,                  // Insufficient (need ~10000)
			TxManager:  filestore.FileID{},    // Empty for test
			Data:       make([]byte, 10*1024), // 10 KB
			Executable: false,
		}

		_, err := fs.CreateFile(file)
		if err == nil {
			t.Error("Expected error for insufficient balance, got none")
		}
		t.Logf("Correctly rejected 10KB file creation with insufficient balance: %v", err)
	})

	// Test 3: Create file with exact balance (should succeed)
	t.Run("exact balance for 1KB", func(t *testing.T) {
		dataSize := int64(1024)
		requiredCost := filestore.CalculateStorageCost(dataSize)

		file := &filestore.File{
			Balance:    requiredCost,
			TxManager:  filestore.FileID{},
			Data:       make([]byte, dataSize),
			Executable: false,
		}

		fileID, err := fs.CreateFile(file)
		if err != nil {
			t.Errorf("File creation should succeed with exact balance: %v", err)
		} else {
			t.Logf("Successfully created file %s with exact balance %d", fileID.String(), requiredCost)
		}
	})

	// Test 4: Create file with more than sufficient balance (should succeed)
	t.Run("sufficient balance for 1KB", func(t *testing.T) {
		file := &filestore.File{
			Balance:    10000, // More than enough
			TxManager:  filestore.FileID{},
			Data:       make([]byte, 1024), // 1 KB
			Executable: false,
		}

		fileID, err := fs.CreateFile(file)
		if err != nil {
			t.Errorf("File creation should succeed with sufficient balance: %v", err)
		} else {
			t.Logf("Successfully created file %s with sufficient balance", fileID.String())
		}
	})
}

// TestDataAllocationExceedingBalance tests that data allocation fails when it would exceed available balance
// This validates Requirements 6.2, 6.4, 6.5
func TestDataAllocationExceedingBalance(t *testing.T) {
	// Create temporary database
	dbPath := t.TempDir() + "/test_data_allocation.db"
	defer os.RemoveAll(dbPath)

	fs, err := filestore.NewFileStore(dbPath)
	if err != nil {
		t.Fatalf("Failed to create FileStore: %v", err)
	}
	defer fs.Close()

	// Create a file with sufficient balance for small data
	t.Run("allocate data exceeding balance", func(t *testing.T) {
		initialData := make([]byte, 1024) // 1 KB
		initialCost := filestore.CalculateStorageCost(int64(len(initialData)))

		file := &filestore.File{
			Balance:    initialCost + 5000, // Enough for 1 KB + some extra
			TxManager:  filestore.FileID{},
			Data:       initialData,
			Executable: false,
		}

		fileID, err := fs.CreateFile(file)
		if err != nil {
			t.Fatalf("Failed to create initial file: %v", err)
		}
		t.Logf("Created file %s with balance %d and 1KB data", fileID.String(), file.Balance)

		// Try to allocate 1 MB of data (would require much more balance)
		retrievedFile, err := fs.GetFile(fileID)
		if err != nil {
			t.Fatalf("Failed to retrieve file: %v", err)
		}

		retrievedFile.Data = make([]byte, 1024*1024) // 1 MB
		requiredCost := filestore.CalculateStorageCost(int64(len(retrievedFile.Data)))
		t.Logf("Attempting to allocate 1MB data (requires %d, have %d)", requiredCost, retrievedFile.Balance)

		err = fs.UpdateFile(fileID, retrievedFile)
		if err == nil {
			t.Error("Expected error when allocating data exceeding balance, got none")
		} else {
			t.Logf("Correctly rejected data allocation: %v", err)
		}
	})

	// Test with System Program AllocateData instruction
	t.Run("system program allocate data exceeding balance", func(t *testing.T) {
		// Initialize runtime and processor
		rt := runtime.NewRuntime()
		if err := genesis.LoadBuiltinPrograms(fs, programs.SystemProgram, programs.TokenProgram); err != nil {
			t.Fatalf("Failed to load builtin programs: %v", err)
		}

		txProc := processor.NewTxProcessor(fs, rt)

		// Create an account with limited balance
		ownerPubKey, _, err := ed25519.GenerateKey(nil)
		if err != nil {
			t.Fatalf("Failed to generate keypair: %v", err)
		}

		var ownerPK transaction.PublicKey
		copy(ownerPK[:], ownerPubKey)

		accountID := publicKeyToFileID(ownerPK)
		initialBalance := int64(100000) // 100k units

		accountFile := &filestore.File{
			ID:         accountID,
			Balance:    initialBalance,
			TxManager:  genesis.SystemProgramID,
			Data:       []byte{},
			Executable: false,
		}
		_, err = fs.CreateFile(accountFile)
		if err != nil {
			t.Fatalf("Failed to create account: %v", err)
		}

		t.Logf("Created account %s with balance %d", accountID.String(), initialBalance)

		// Try to allocate 10 MB of data (requires ~26M units, but only have 100k)
		allocateSize := int64(10 * 1024 * 1024) // 10 MB
		requiredCost := filestore.CalculateStorageCost(allocateSize)
		t.Logf("Attempting to allocate %d bytes (requires %d, have %d)", allocateSize, requiredCost, initialBalance)

		allocateInstr := encodeAllocateDataInstruction(allocateSize)
		allocateTx := &transaction.Transaction{
			LastSeen: transaction.TxID{},
			Instructions: []transaction.Instruction{
				{
					ProgramID: genesis.SystemProgramID,
					Inputs: map[string]transaction.FileAccess{
						"account": {FileID: accountID, Permission: transaction.Write},
					},
					Data: allocateInstr,
				},
			},
			Signatures: []transaction.Signature{},
		}

		// Sign transaction
		txData, _ := allocateTx.Marshal()
		sig := ed25519.Sign(ed25519.PrivateKey(append(ownerPubKey, make([]byte, 32)...)), txData)
		var sigBytes [64]byte
		copy(sigBytes[:], sig)

		allocateTx.Signatures = []transaction.Signature{
			{PublicKey: ownerPK, Signature: sigBytes},
		}

		_, err = txProc.ProcessTransaction(allocateTx)
		if err == nil {
			t.Error("Expected error when allocating data exceeding balance, got none")
		} else {
			t.Logf("Correctly rejected allocation: %v", err)
		}
	})
}

// TestBalanceReductionBelowStorageCost tests that balance reduction fails when it would go below storage cost requirement
// This validates Requirements 6.3, 6.5
func TestBalanceReductionBelowStorageCost(t *testing.T) {
	// Create temporary database
	dbPath := t.TempDir() + "/test_balance_reduction.db"
	defer os.RemoveAll(dbPath)

	fs, err := filestore.NewFileStore(dbPath)
	if err != nil {
		t.Fatalf("Failed to create FileStore: %v", err)
	}
	defer fs.Close()

	t.Run("reduce balance below storage cost", func(t *testing.T) {
		// Create file with 10 KB data and sufficient balance
		dataSize := int64(10 * 1024) // 10 KB
		requiredCost := filestore.CalculateStorageCost(dataSize)
		initialBalance := requiredCost + 50000 // Extra balance

		file := &filestore.File{
			Balance:    initialBalance,
			TxManager:  filestore.FileID{},
			Data:       make([]byte, dataSize),
			Executable: false,
		}

		fileID, err := fs.CreateFile(file)
		if err != nil {
			t.Fatalf("Failed to create file: %v", err)
		}
		t.Logf("Created file %s with balance %d (required: %d)", fileID.String(), initialBalance, requiredCost)

		// Try to reduce balance below storage cost
		reductionAmount := initialBalance - requiredCost + 1000 // Would leave less than required
		t.Logf("Attempting to reduce balance by %d (would leave %d, need %d)",
			reductionAmount, initialBalance-reductionAmount, requiredCost)

		err = fs.UpdateFileBalance(fileID, -reductionAmount)
		if err == nil {
			t.Error("Expected error when reducing balance below storage cost, got none")
		} else {
			t.Logf("Correctly rejected balance reduction: %v", err)
		}

		// Verify balance unchanged
		currentBalance, err := fs.GetFileBalance(fileID)
		if err != nil {
			t.Fatalf("Failed to get file balance: %v", err)
		}
		if currentBalance != initialBalance {
			t.Errorf("Balance should be unchanged: expected %d, got %d", initialBalance, currentBalance)
		}
	})

	t.Run("reduce balance to exact storage cost", func(t *testing.T) {
		// Create file with extra balance
		dataSize := int64(5 * 1024) // 5 KB
		requiredCost := filestore.CalculateStorageCost(dataSize)
		initialBalance := requiredCost + 10000

		file := &filestore.File{
			Balance:    initialBalance,
			TxManager:  filestore.FileID{},
			Data:       make([]byte, dataSize),
			Executable: false,
		}

		fileID, err := fs.CreateFile(file)
		if err != nil {
			t.Fatalf("Failed to create file: %v", err)
		}

		// Reduce balance to exactly the storage cost (should succeed)
		reductionAmount := initialBalance - requiredCost
		t.Logf("Reducing balance by %d to exact storage cost %d", reductionAmount, requiredCost)

		err = fs.UpdateFileBalance(fileID, -reductionAmount)
		if err != nil {
			t.Errorf("Should allow balance reduction to exact storage cost: %v", err)
		}

		// Verify balance is now exactly the required cost
		currentBalance, err := fs.GetFileBalance(fileID)
		if err != nil {
			t.Fatalf("Failed to get file balance: %v", err)
		}
		if currentBalance != requiredCost {
			t.Errorf("Balance should be %d, got %d", requiredCost, currentBalance)
		}
	})

	// Test with System Program Transfer instruction
	t.Run("transfer causing balance below storage cost", func(t *testing.T) {
		// Initialize runtime and processor
		rt := runtime.NewRuntime()
		if err := genesis.LoadBuiltinPrograms(fs, programs.SystemProgram, programs.TokenProgram); err != nil {
			t.Fatalf("Failed to load builtin programs: %v", err)
		}

		txProc := processor.NewTxProcessor(fs, rt)

		// Create two accounts
		alicePubKey, alicePrivKey, err := ed25519.GenerateKey(nil)
		if err != nil {
			t.Fatalf("Failed to generate Alice's keypair: %v", err)
		}

		bobPubKey, _, err := ed25519.GenerateKey(nil)
		if err != nil {
			t.Fatalf("Failed to generate Bob's keypair: %v", err)
		}

		var alicePK, bobPK transaction.PublicKey
		copy(alicePK[:], alicePubKey)
		copy(bobPK[:], bobPubKey)

		// Create Alice's account with data
		dataSize := int64(20 * 1024) // 20 KB
		requiredCost := filestore.CalculateStorageCost(dataSize)
		aliceBalance := requiredCost + 5000 // Just slightly more than required

		aliceID := publicKeyToFileID(alicePK)
		aliceFile := &filestore.File{
			ID:         aliceID,
			Balance:    aliceBalance,
			TxManager:  genesis.SystemProgramID,
			Data:       make([]byte, dataSize),
			Executable: false,
		}
		_, err = fs.CreateFile(aliceFile)
		if err != nil {
			t.Fatalf("Failed to create Alice's account: %v", err)
		}

		// Create Bob's account
		bobID := publicKeyToFileID(bobPK)
		bobFile := &filestore.File{
			ID:         bobID,
			Balance:    10000,
			TxManager:  genesis.SystemProgramID,
			Data:       []byte{},
			Executable: false,
		}
		_, err = fs.CreateFile(bobFile)
		if err != nil {
			t.Fatalf("Failed to create Bob's account: %v", err)
		}

		t.Logf("Alice account %s: balance=%d, required=%d, data=%d bytes",
			aliceID.String(), aliceBalance, requiredCost, dataSize)
		t.Logf("Bob account %s: balance=%d", bobID.String(), 10000)

		// Try to transfer amount that would leave Alice below storage cost
		transferAmount := int64(6000) // Would leave Alice with less than required
		t.Logf("Attempting to transfer %d (would leave Alice with %d, needs %d)",
			transferAmount, aliceBalance-transferAmount, requiredCost)

		transferInstr := encodeTransferInstructionSC(transferAmount, aliceID, bobID)
		transferTx := &transaction.Transaction{
			LastSeen: transaction.TxID{},
			Instructions: []transaction.Instruction{
				{
					ProgramID: genesis.SystemProgramID,
					Inputs: map[string]transaction.FileAccess{
						"from": {FileID: aliceID, Permission: transaction.Write},
						"to":   {FileID: bobID, Permission: transaction.Write},
					},
					Data: transferInstr,
				},
			},
			Signatures: []transaction.Signature{},
		}

		// Sign transaction
		txData, _ := transferTx.Marshal()
		aliceSig := ed25519.Sign(alicePrivKey, txData)
		var aliceSigBytes [64]byte
		copy(aliceSigBytes[:], aliceSig)

		transferTx.Signatures = []transaction.Signature{
			{PublicKey: alicePK, Signature: aliceSigBytes},
		}

		_, err = txProc.ProcessTransaction(transferTx)
		if err == nil {
			t.Error("Expected error when transfer would leave balance below storage cost, got none")
		} else {
			t.Logf("Correctly rejected transfer: %v", err)
		}

		// Verify Alice's balance unchanged
		aliceBalance2, err := fs.GetFileBalance(aliceID)
		if err != nil {
			t.Fatalf("Failed to get Alice's balance: %v", err)
		}
		if aliceBalance2 != aliceBalance {
			t.Errorf("Alice's balance should be unchanged: expected %d, got %d", aliceBalance, aliceBalance2)
		}
	})
}

// TestExponentialCostGrowthForLargeFiles tests that storage cost grows exponentially with file size
// This validates Requirement 6.2
func TestExponentialCostGrowthForLargeFiles(t *testing.T) {
	// Calculate costs for various file sizes
	sizes := []struct {
		name string
		size int64
	}{
		{"1 KB", 1 * 1024},
		{"10 KB", 10 * 1024},
		{"100 KB", 100 * 1024},
		{"1 MB", 1 * 1024 * 1024},
		{"5 MB", 5 * 1024 * 1024},
		{"10 MB", 10 * 1024 * 1024},
		{"50 MB", 50 * 1024 * 1024},
		{"100 MB", 100 * 1024 * 1024},
	}

	costs := make([]int64, len(sizes))
	ratios := make([]float64, len(sizes)-1)

	t.Log("Storage cost growth analysis:")
	t.Log("Size\t\tCost\t\tRatio to Previous")
	t.Log("----\t\t----\t\t-----------------")

	for i, s := range sizes {
		costs[i] = filestore.CalculateStorageCost(s.size)

		if i > 0 {
			ratios[i-1] = float64(costs[i]) / float64(costs[i-1])
			t.Logf("%s\t\t%d\t\t%.2fx", s.name, costs[i], ratios[i-1])
		} else {
			t.Logf("%s\t\t%d\t\t-", s.name, costs[i])
		}
	}

	// Verify exponential growth: each ratio should be greater than linear growth
	// For exponential growth with base 1.1, we expect ratios to increase
	t.Run("verify exponential growth", func(t *testing.T) {
		// Cost should grow faster than linear
		// Compare 1MB to 1KB: should be much more than 1024x
		linearRatio := float64(sizes[3].size) / float64(sizes[0].size) // 1MB / 1KB = 1024
		actualRatio := float64(costs[3]) / float64(costs[0])

		t.Logf("Linear ratio (1MB/1KB): %.2fx", linearRatio)
		t.Logf("Actual cost ratio: %.2fx", actualRatio)

		if actualRatio <= linearRatio {
			t.Errorf("Expected exponential growth: actual ratio %.2f should be > linear ratio %.2f",
				actualRatio, linearRatio)
		}

		// Verify that larger files have increasingly higher cost ratios
		// Compare 10MB to 1MB vs 1MB to 100KB
		ratio_1MB_to_10MB := float64(costs[5]) / float64(costs[3])
		ratio_100KB_to_1MB := float64(costs[3]) / float64(costs[2])

		t.Logf("Cost ratio (10MB/1MB): %.2fx", ratio_1MB_to_10MB)
		t.Logf("Cost ratio (1MB/100KB): %.2fx", ratio_100KB_to_1MB)

		if ratio_1MB_to_10MB <= ratio_100KB_to_1MB {
			t.Errorf("Expected accelerating growth: 10MB/1MB ratio (%.2f) should be > 1MB/100KB ratio (%.2f)",
				ratio_1MB_to_10MB, ratio_100KB_to_1MB)
		}
	})

	// Test that files with exponentially growing sizes require exponentially growing balances
	t.Run("create files with exponential sizes", func(t *testing.T) {
		dbPath := t.TempDir() + "/test_exponential_growth.db"
		defer os.RemoveAll(dbPath)

		fs, err := filestore.NewFileStore(dbPath)
		if err != nil {
			t.Fatalf("Failed to create FileStore: %v", err)
		}
		defer fs.Close()

		// Try to create files with increasing sizes
		testSizes := []int64{
			1 * 1024,        // 1 KB
			10 * 1024,       // 10 KB
			100 * 1024,      // 100 KB
			1024 * 1024,     // 1 MB
			5 * 1024 * 1024, // 5 MB
		}

		for _, size := range testSizes {
			requiredCost := filestore.CalculateStorageCost(size)

			// Test with insufficient balance (90% of required)
			insufficientBalance := int64(float64(requiredCost) * 0.9)
			file := &filestore.File{
				Balance:    insufficientBalance,
				TxManager:  filestore.FileID{},
				Data:       make([]byte, size),
				Executable: false,
			}

			_, err := fs.CreateFile(file)
			if err == nil {
				t.Errorf("Should reject file of size %d with insufficient balance %d (need %d)",
					size, insufficientBalance, requiredCost)
			}

			// Test with sufficient balance
			file.Balance = requiredCost + 1000
			fileID, err := fs.CreateFile(file)
			if err != nil {
				t.Errorf("Should accept file of size %d with sufficient balance %d: %v",
					size, file.Balance, err)
			} else {
				t.Logf("Created file %s with size %d bytes, cost %d", fileID.String(), size, requiredCost)
			}
		}
	})
}
