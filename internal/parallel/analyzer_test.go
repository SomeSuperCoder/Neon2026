package parallel

import (
	"testing"

	"github.com/poh-blockchain/internal/filestore"
	"github.com/poh-blockchain/internal/transaction"
)

// Helper function to create a test FileID
func testFileID(id byte) filestore.FileID {
	var fileID filestore.FileID
	fileID[0] = id
	return fileID
}

// Helper function to create a test transaction with specified file accesses
func createTestTransaction(reads []byte, writes []byte) *transaction.Transaction {
	tx := &transaction.Transaction{
		Instructions: []transaction.Instruction{},
	}

	instr := transaction.Instruction{
		Inputs: make(map[string]transaction.FileAccess),
	}

	for i, r := range reads {
		instr.Inputs[string(rune('r'))+string(rune(i))] = transaction.FileAccess{
			FileID:     testFileID(r),
			Permission: transaction.Read,
		}
	}

	for i, w := range writes {
		instr.Inputs[string(rune('w'))+string(rune(i))] = transaction.FileAccess{
			FileID:     testFileID(w),
			Permission: transaction.Write,
		}
	}

	tx.Instructions = append(tx.Instructions, instr)
	return tx
}

func TestNewParallelAnalyzer(t *testing.T) {
	analyzer := NewParallelAnalyzer()
	if analyzer == nil {
		t.Fatal("NewParallelAnalyzer returned nil")
	}
}

func TestHasConflict_NoConflict_ReadRead(t *testing.T) {
	analyzer := NewParallelAnalyzer()

	// Both transactions read the same file - no conflict
	tx1 := createTestTransaction([]byte{1}, []byte{})
	tx2 := createTestTransaction([]byte{1}, []byte{})

	if analyzer.HasConflict(tx1, tx2) {
		t.Error("Expected no conflict for read-read access")
	}
}

func TestHasConflict_Conflict_WriteWrite(t *testing.T) {
	analyzer := NewParallelAnalyzer()

	// Both transactions write to the same file - conflict
	tx1 := createTestTransaction([]byte{}, []byte{1})
	tx2 := createTestTransaction([]byte{}, []byte{1})

	if !analyzer.HasConflict(tx1, tx2) {
		t.Error("Expected conflict for write-write access")
	}
}

func TestHasConflict_Conflict_ReadWrite(t *testing.T) {
	analyzer := NewParallelAnalyzer()

	// tx1 reads, tx2 writes to the same file - conflict
	tx1 := createTestTransaction([]byte{1}, []byte{})
	tx2 := createTestTransaction([]byte{}, []byte{1})

	if !analyzer.HasConflict(tx1, tx2) {
		t.Error("Expected conflict for read-write access")
	}
}

func TestHasConflict_Conflict_WriteRead(t *testing.T) {
	analyzer := NewParallelAnalyzer()

	// tx1 writes, tx2 reads the same file - conflict
	tx1 := createTestTransaction([]byte{}, []byte{1})
	tx2 := createTestTransaction([]byte{1}, []byte{})

	if !analyzer.HasConflict(tx1, tx2) {
		t.Error("Expected conflict for write-read access")
	}
}

func TestHasConflict_NoConflict_DifferentFiles(t *testing.T) {
	analyzer := NewParallelAnalyzer()

	// Transactions access different files - no conflict
	tx1 := createTestTransaction([]byte{}, []byte{1})
	tx2 := createTestTransaction([]byte{}, []byte{2})

	if analyzer.HasConflict(tx1, tx2) {
		t.Error("Expected no conflict for different file access")
	}
}

func TestAnalyzeDependencies_NoConflicts(t *testing.T) {
	analyzer := NewParallelAnalyzer()

	// Three transactions accessing different files
	txs := []*transaction.Transaction{
		createTestTransaction([]byte{}, []byte{1}),
		createTestTransaction([]byte{}, []byte{2}),
		createTestTransaction([]byte{}, []byte{3}),
	}

	graph := analyzer.AnalyzeDependencies(txs)

	if graph.NumTransactions != 3 {
		t.Errorf("Expected 3 transactions, got %d", graph.NumTransactions)
	}

	if len(graph.Conflicts) != 0 {
		t.Errorf("Expected no conflicts, got %d", len(graph.Conflicts))
	}
}

func TestAnalyzeDependencies_WithConflicts(t *testing.T) {
	analyzer := NewParallelAnalyzer()

	// tx0 and tx1 conflict (both write to file 1)
	// tx1 and tx2 conflict (both write to file 2)
	txs := []*transaction.Transaction{
		createTestTransaction([]byte{}, []byte{1}),
		createTestTransaction([]byte{}, []byte{1, 2}),
		createTestTransaction([]byte{}, []byte{2}),
	}

	graph := analyzer.AnalyzeDependencies(txs)

	if graph.NumTransactions != 3 {
		t.Errorf("Expected 3 transactions, got %d", graph.NumTransactions)
	}

	// Should have 2 conflicts: (0,1) and (1,2)
	if len(graph.Conflicts) != 2 {
		t.Errorf("Expected 2 conflicts, got %d", len(graph.Conflicts))
	}

	if !graph.HasConflict(0, 1) {
		t.Error("Expected conflict between tx0 and tx1")
	}

	if !graph.HasConflict(1, 2) {
		t.Error("Expected conflict between tx1 and tx2")
	}

	if graph.HasConflict(0, 2) {
		t.Error("Expected no conflict between tx0 and tx2")
	}
}

func TestDependencyGraph_ConflictTypes(t *testing.T) {
	analyzer := NewParallelAnalyzer()

	txs := []*transaction.Transaction{
		createTestTransaction([]byte{}, []byte{1}), // tx0: write file 1
		createTestTransaction([]byte{}, []byte{1}), // tx1: write file 1 (write-write)
		createTestTransaction([]byte{1}, []byte{}), // tx2: read file 1 (read-write)
	}

	graph := analyzer.AnalyzeDependencies(txs)

	// tx0 vs tx1: write-write conflict
	if graph.GetConflictType(0, 1) != WriteWriteConflict {
		t.Error("Expected write-write conflict between tx0 and tx1")
	}

	// tx0 vs tx2: read-write conflict
	if graph.GetConflictType(0, 2) != ReadWriteConflict {
		t.Error("Expected read-write conflict between tx0 and tx2")
	}

	// tx1 vs tx2: read-write conflict
	if graph.GetConflictType(1, 2) != ReadWriteConflict {
		t.Error("Expected read-write conflict between tx1 and tx2")
	}
}

func TestFindParallelBatches_NoConflicts(t *testing.T) {
	analyzer := NewParallelAnalyzer()

	// Three transactions with no conflicts
	txs := []*transaction.Transaction{
		createTestTransaction([]byte{}, []byte{1}),
		createTestTransaction([]byte{}, []byte{2}),
		createTestTransaction([]byte{}, []byte{3}),
	}

	graph := analyzer.AnalyzeDependencies(txs)
	batches := analyzer.FindParallelBatches(graph)

	// All transactions should be in one batch
	if len(batches) != 1 {
		t.Errorf("Expected 1 batch, got %d", len(batches))
	}

	if len(batches[0].TransactionIndices) != 3 {
		t.Errorf("Expected 3 transactions in batch, got %d", len(batches[0].TransactionIndices))
	}
}

func TestFindParallelBatches_AllConflict(t *testing.T) {
	analyzer := NewParallelAnalyzer()

	// Three transactions all writing to the same file
	txs := []*transaction.Transaction{
		createTestTransaction([]byte{}, []byte{1}),
		createTestTransaction([]byte{}, []byte{1}),
		createTestTransaction([]byte{}, []byte{1}),
	}

	graph := analyzer.AnalyzeDependencies(txs)
	batches := analyzer.FindParallelBatches(graph)

	// Each transaction should be in its own batch
	if len(batches) != 3 {
		t.Errorf("Expected 3 batches, got %d", len(batches))
	}

	for i, batch := range batches {
		if len(batch.TransactionIndices) != 1 {
			t.Errorf("Expected 1 transaction in batch %d, got %d", i, len(batch.TransactionIndices))
		}
	}
}

func TestFindParallelBatches_PartialConflicts(t *testing.T) {
	analyzer := NewParallelAnalyzer()

	// tx0: write file 1
	// tx1: write file 2
	// tx2: write file 1 (conflicts with tx0)
	// tx3: write file 3
	txs := []*transaction.Transaction{
		createTestTransaction([]byte{}, []byte{1}),
		createTestTransaction([]byte{}, []byte{2}),
		createTestTransaction([]byte{}, []byte{1}),
		createTestTransaction([]byte{}, []byte{3}),
	}

	graph := analyzer.AnalyzeDependencies(txs)
	batches := analyzer.FindParallelBatches(graph)

	// Expected batches:
	// Batch 0: tx0, tx1, tx3 (no conflicts)
	// Batch 1: tx2 (conflicts with tx0)
	if len(batches) != 2 {
		t.Errorf("Expected 2 batches, got %d", len(batches))
	}

	// First batch should have 3 transactions
	if len(batches[0].TransactionIndices) != 3 {
		t.Errorf("Expected 3 transactions in first batch, got %d", len(batches[0].TransactionIndices))
	}

	// Second batch should have 1 transaction
	if len(batches[1].TransactionIndices) != 1 {
		t.Errorf("Expected 1 transaction in second batch, got %d", len(batches[1].TransactionIndices))
	}
}

func TestFindParallelBatches_EmptyGraph(t *testing.T) {
	analyzer := NewParallelAnalyzer()

	graph := NewDependencyGraph(0)
	batches := analyzer.FindParallelBatches(graph)

	if len(batches) != 0 {
		t.Errorf("Expected 0 batches for empty graph, got %d", len(batches))
	}
}

func TestCalculateBatchStatistics(t *testing.T) {
	// Create batches manually for testing
	batches := []Batch{
		{TransactionIndices: []int{0, 1, 2}},
		{TransactionIndices: []int{3}},
	}

	stats := CalculateBatchStatistics(batches, 4)

	if stats.TotalTransactions != 4 {
		t.Errorf("Expected 4 total transactions, got %d", stats.TotalTransactions)
	}

	if stats.NumBatches != 2 {
		t.Errorf("Expected 2 batches, got %d", stats.NumBatches)
	}

	expectedAvg := 2.0 // (3 + 1) / 2
	if stats.AverageBatchSize != expectedAvg {
		t.Errorf("Expected average batch size %.2f, got %.2f", expectedAvg, stats.AverageBatchSize)
	}

	if stats.MaxBatchSize != 3 {
		t.Errorf("Expected max batch size 3, got %d", stats.MaxBatchSize)
	}

	// Parallelization rate: 1 - (2 batches / 4 transactions) = 0.5
	expectedRate := 0.5
	if stats.ParallelizationRate != expectedRate {
		t.Errorf("Expected parallelization rate %.2f, got %.2f", expectedRate, stats.ParallelizationRate)
	}
}

func TestCalculateBatchStatistics_Empty(t *testing.T) {
	stats := CalculateBatchStatistics([]Batch{}, 0)

	if stats.TotalTransactions != 0 {
		t.Errorf("Expected 0 total transactions, got %d", stats.TotalTransactions)
	}

	if stats.NumBatches != 0 {
		t.Errorf("Expected 0 batches, got %d", stats.NumBatches)
	}
}

func TestMultiInstructionTransaction(t *testing.T) {
	analyzer := NewParallelAnalyzer()

	// Create a transaction with multiple instructions
	tx := &transaction.Transaction{
		Instructions: []transaction.Instruction{
			{
				Inputs: map[string]transaction.FileAccess{
					"file1": {FileID: testFileID(1), Permission: transaction.Read},
				},
			},
			{
				Inputs: map[string]transaction.FileAccess{
					"file2": {FileID: testFileID(2), Permission: transaction.Write},
				},
			},
		},
	}

	// Another transaction that writes to file 2
	tx2 := createTestTransaction([]byte{}, []byte{2})

	// Should conflict because both write to file 2
	if !analyzer.HasConflict(tx, tx2) {
		t.Error("Expected conflict for multi-instruction transaction")
	}
}

func TestComplexDependencyChain(t *testing.T) {
	analyzer := NewParallelAnalyzer()

	// Create a chain of dependencies:
	// tx0 writes file 1
	// tx1 reads file 1, writes file 2
	// tx2 reads file 2, writes file 3
	// tx3 writes file 4 (independent)
	txs := []*transaction.Transaction{
		createTestTransaction([]byte{}, []byte{1}),
		createTestTransaction([]byte{1}, []byte{2}),
		createTestTransaction([]byte{2}, []byte{3}),
		createTestTransaction([]byte{}, []byte{4}),
	}

	graph := analyzer.AnalyzeDependencies(txs)

	// tx0 and tx1 conflict (tx1 reads what tx0 writes)
	if !graph.HasConflict(0, 1) {
		t.Error("Expected conflict between tx0 and tx1")
	}

	// tx1 and tx2 conflict (tx2 reads what tx1 writes)
	if !graph.HasConflict(1, 2) {
		t.Error("Expected conflict between tx1 and tx2")
	}

	// tx3 is independent
	if graph.HasConflict(0, 3) || graph.HasConflict(1, 3) || graph.HasConflict(2, 3) {
		t.Error("Expected no conflicts with tx3")
	}

	batches := analyzer.FindParallelBatches(graph)

	// Expected batches with greedy algorithm:
	// The algorithm processes in order and adds non-conflicting txs to each batch
	// Batch 0: tx0 (start), tx2 (no conflict with tx0), tx3 (no conflict with tx0 or tx2)
	// Batch 1: tx1 (conflicts with tx0 and tx2)
	//
	// tx0 writes file 1, tx2 reads file 2 (no conflict)
	// tx0 writes file 1, tx3 writes file 4 (no conflict)
	// tx2 reads file 2, tx3 writes file 4 (no conflict)
	// tx1 reads file 1 (conflicts with tx0 which writes file 1)
	// tx1 writes file 2 (conflicts with tx2 which reads file 2)
	if len(batches) != 2 {
		t.Errorf("Expected 2 batches, got %d", len(batches))
		for i, batch := range batches {
			t.Logf("Batch %d: %v", i, batch.TransactionIndices)
		}
	}
}
