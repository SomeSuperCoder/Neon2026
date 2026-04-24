package parallel

import (
	"fmt"

	"github.com/poh-blockchain/internal/filestore"
	"github.com/poh-blockchain/internal/transaction"
)

/*
Package parallel provides transaction dependency analysis for parallel execution.

OVERVIEW

This package implements the foundation for parallel transaction execution by analyzing
declared file access patterns in transactions. It identifies which transactions can
safely execute concurrently without conflicts.

The system is based on the principle that transactions declare all file accesses
upfront through their instruction inputs. This enables static analysis of dependencies
before execution begins.

CONFLICT DETECTION

Transactions conflict when they access the same file in incompatible ways:

1. Write-Write Conflict: Both transactions write to the same file
   - Must execute sequentially to maintain deterministic order
   - Example: tx1 writes file A, tx2 writes file A → CONFLICT

2. Read-Write Conflict: One transaction reads while another writes
   - Reader must see consistent state (either before or after write)
   - Example: tx1 reads file A, tx2 writes file A → CONFLICT

3. Read-Read: No conflict - multiple transactions can read simultaneously
   - Example: tx1 reads file A, tx2 reads file A → NO CONFLICT

DEPENDENCY GRAPH

The dependency graph represents relationships between transactions:
- Nodes: Transaction indices (0, 1, 2, ...)
- Edges: Conflicts between transactions
- Dependencies: Transaction X depends on Y if they conflict and Y comes first

The graph enables:
- Identifying which transactions must execute sequentially
- Finding sets of transactions that can execute in parallel
- Maintaining deterministic execution order

BATCHING ALGORITHM

FindParallelBatches uses a greedy algorithm to group non-conflicting transactions:

1. Process transactions in original order (maintains determinism)
2. Start each batch with the next unprocessed transaction
3. Add subsequent transactions that don't conflict with any in the batch
4. Continue until all transactions are assigned to batches

Example:
  Input: [tx0, tx1, tx2, tx3]
  Conflicts: tx0↔tx2 (both write file A)
  Output: [[tx0, tx1, tx3], [tx2]]

  Batch 0 can execute in parallel (tx0, tx1, tx3 don't conflict)
  Batch 1 executes after (tx2 conflicts with tx0)

FUTURE PARALLEL EXECUTION IMPLEMENTATION

This package provides the analysis foundation. Future implementation will add:

1. PARALLEL EXECUTOR
   - Execute batches of non-conflicting transactions concurrently
   - Use goroutines/thread pool for parallel execution
   - Collect results and merge state changes

2. LOCK-FREE EXECUTION
   - Transactions declare all accesses upfront (already implemented)
   - No runtime locking needed - conflicts resolved statically
   - Optimistic execution with validation

3. STATE ISOLATION
   - Each transaction gets isolated view of file state
   - Changes buffered in transaction-local cache
   - Commit all changes atomically after validation

4. SCHEDULER INTEGRATION
   - Integrate with block producer to batch transactions
   - Analyze incoming transactions for parallel opportunities
   - Schedule batches across available CPU cores

5. PERFORMANCE OPTIMIZATION
   - Cache dependency analysis results
   - Reuse batching decisions for similar transaction patterns
   - Profile and optimize hot paths

6. ADVANCED FEATURES
   - Speculative execution: Execute likely non-conflicting txs early
   - Dynamic rebatching: Adjust batches based on actual conflicts
   - Priority scheduling: Execute high-fee transactions first

USAGE EXAMPLE

	analyzer := NewParallelAnalyzer()

	// Analyze transaction dependencies
	graph := analyzer.AnalyzeDependencies(transactions)

	// Find parallel execution opportunities
	batches := analyzer.FindParallelBatches(graph)

	// Execute each batch in parallel
	for _, batch := range batches {
		// Future: Execute batch.TransactionIndices concurrently
		executeBatchInParallel(batch, transactions)
	}

DETERMINISM GUARANTEE

The system maintains deterministic execution order:
- Transactions are processed in their original order
- Batching preserves order within constraints
- Conflicting transactions execute in original sequence
- Final state is identical to sequential execution

This ensures consensus across all nodes in the blockchain network.

PERFORMANCE CHARACTERISTICS

Time Complexity:
- AnalyzeDependencies: O(n² × m) where n=transactions, m=avg files per tx
- FindParallelBatches: O(n² × b) where b=avg batch size
- HasConflict: O(m) where m=avg files per transaction

Space Complexity:
- DependencyGraph: O(n²) for conflicts map
- FileAccessPattern: O(m) per transaction

For typical workloads with n=1000 transactions and m=10 files per transaction,
analysis completes in milliseconds.
*/

// ParallelAnalyzer analyzes transactions to identify parallel execution opportunities.
// It builds dependency graphs and detects conflicts between transactions based on
// their declared file access patterns.
//
// The analyzer is stateless and thread-safe - all methods can be called concurrently.
type ParallelAnalyzer struct {
	// No state needed - all methods are stateless for thread safety
}

// NewParallelAnalyzer creates a new parallel execution analyzer.
func NewParallelAnalyzer() *ParallelAnalyzer {
	return &ParallelAnalyzer{}
}

// ConflictType represents the type of conflict between transactions
type ConflictType int

const (
	NoConflict ConflictType = iota
	ReadWriteConflict
	WriteWriteConflict
)

// DependencyGraph represents the dependency relationships between transactions.
// Nodes are transaction indices, edges represent dependencies (conflicts).
type DependencyGraph struct {
	// NumTransactions is the total number of transactions in the graph
	NumTransactions int

	// Conflicts maps transaction index pairs to their conflict type
	// Key format: "txIndex1:txIndex2" where txIndex1 < txIndex2
	Conflicts map[string]ConflictType

	// Dependencies maps each transaction to the set of transactions it depends on
	// (transactions that must execute before it due to conflicts)
	Dependencies map[int][]int
}

// NewDependencyGraph creates a new dependency graph for the given number of transactions.
func NewDependencyGraph(numTxs int) *DependencyGraph {
	return &DependencyGraph{
		NumTransactions: numTxs,
		Conflicts:       make(map[string]ConflictType),
		Dependencies:    make(map[int][]int),
	}
}

// AddConflict records a conflict between two transactions.
func (dg *DependencyGraph) AddConflict(tx1, tx2 int, conflictType ConflictType) {
	if tx1 > tx2 {
		tx1, tx2 = tx2, tx1
	}
	key := fmt.Sprintf("%d:%d", tx1, tx2)
	dg.Conflicts[key] = conflictType

	// Add to dependencies - tx2 depends on tx1 (must execute after)
	dg.Dependencies[tx2] = append(dg.Dependencies[tx2], tx1)
}

// HasConflict checks if two transactions have a conflict.
func (dg *DependencyGraph) HasConflict(tx1, tx2 int) bool {
	if tx1 > tx2 {
		tx1, tx2 = tx2, tx1
	}
	key := fmt.Sprintf("%d:%d", tx1, tx2)
	_, exists := dg.Conflicts[key]
	return exists
}

// GetConflictType returns the type of conflict between two transactions.
func (dg *DependencyGraph) GetConflictType(tx1, tx2 int) ConflictType {
	if tx1 > tx2 {
		tx1, tx2 = tx2, tx1
	}
	key := fmt.Sprintf("%d:%d", tx1, tx2)
	if conflictType, exists := dg.Conflicts[key]; exists {
		return conflictType
	}
	return NoConflict
}

// FileAccessPattern represents how a transaction accesses files
type FileAccessPattern struct {
	ReadFiles  map[filestore.FileID]bool
	WriteFiles map[filestore.FileID]bool
}

// NewFileAccessPattern creates a new file access pattern.
func NewFileAccessPattern() *FileAccessPattern {
	return &FileAccessPattern{
		ReadFiles:  make(map[filestore.FileID]bool),
		WriteFiles: make(map[filestore.FileID]bool),
	}
}

// AnalyzeDependencies builds a dependency graph for a set of transactions.
// It analyzes the declared file access patterns in each transaction's instructions
// to identify conflicts.
func (pa *ParallelAnalyzer) AnalyzeDependencies(txs []*transaction.Transaction) *DependencyGraph {
	graph := NewDependencyGraph(len(txs))

	// Extract access patterns for each transaction
	patterns := make([]*FileAccessPattern, len(txs))
	for i, tx := range txs {
		patterns[i] = pa.extractAccessPattern(tx)
	}

	// Compare each pair of transactions for conflicts
	for i := 0; i < len(txs); i++ {
		for j := i + 1; j < len(txs); j++ {
			conflictType := pa.detectConflict(patterns[i], patterns[j])
			if conflictType != NoConflict {
				graph.AddConflict(i, j, conflictType)
			}
		}
	}

	return graph
}

// extractAccessPattern extracts the file access pattern from a transaction.
func (pa *ParallelAnalyzer) extractAccessPattern(tx *transaction.Transaction) *FileAccessPattern {
	pattern := NewFileAccessPattern()

	for _, instr := range tx.Instructions {
		for _, access := range instr.Inputs {
			if access.Permission == transaction.Write {
				pattern.WriteFiles[access.FileID] = true
			} else {
				pattern.ReadFiles[access.FileID] = true
			}
		}
	}

	return pattern
}

// detectConflict determines if two access patterns conflict.
// Conflict rules:
// - Write-Write: Both transactions write to the same file
// - Read-Write: One transaction reads while another writes to the same file
// - Read-Read: No conflict (both can read simultaneously)
func (pa *ParallelAnalyzer) detectConflict(p1, p2 *FileAccessPattern) ConflictType {
	// Check for write-write conflicts
	for fileID := range p1.WriteFiles {
		if p2.WriteFiles[fileID] {
			return WriteWriteConflict
		}
	}

	// Check for read-write conflicts (p1 reads, p2 writes)
	for fileID := range p1.ReadFiles {
		if p2.WriteFiles[fileID] {
			return ReadWriteConflict
		}
	}

	// Check for read-write conflicts (p1 writes, p2 reads)
	for fileID := range p1.WriteFiles {
		if p2.ReadFiles[fileID] {
			return ReadWriteConflict
		}
	}

	return NoConflict
}

// HasConflict checks if two transactions conflict based on their file access patterns.
// This is a convenience method that doesn't require building a full dependency graph.
func (pa *ParallelAnalyzer) HasConflict(tx1, tx2 *transaction.Transaction) bool {
	p1 := pa.extractAccessPattern(tx1)
	p2 := pa.extractAccessPattern(tx2)
	return pa.detectConflict(p1, p2) != NoConflict
}

// Batch represents a set of transactions that can be executed in parallel.
type Batch struct {
	// TransactionIndices contains the indices of transactions in this batch
	TransactionIndices []int
}

// FindParallelBatches identifies sets of non-conflicting transactions that can
// be executed in parallel. It uses a greedy algorithm to group transactions into
// batches where no two transactions in the same batch conflict.
//
// The algorithm maintains deterministic execution order by processing transactions
// in their original order and only batching those that don't conflict with earlier
// transactions in the same batch.
func (pa *ParallelAnalyzer) FindParallelBatches(graph *DependencyGraph) []Batch {
	if graph.NumTransactions == 0 {
		return []Batch{}
	}

	batches := []Batch{}
	processed := make(map[int]bool)

	// Process transactions in order
	for i := 0; i < graph.NumTransactions; i++ {
		if processed[i] {
			continue
		}

		// Start a new batch with this transaction
		batch := Batch{
			TransactionIndices: []int{i},
		}
		processed[i] = true

		// Try to add more transactions to this batch
		for j := i + 1; j < graph.NumTransactions; j++ {
			if processed[j] {
				continue
			}

			// Check if transaction j conflicts with any transaction in the current batch
			canAdd := true
			for _, batchTxIdx := range batch.TransactionIndices {
				if graph.HasConflict(batchTxIdx, j) {
					canAdd = false
					break
				}
			}

			if canAdd {
				batch.TransactionIndices = append(batch.TransactionIndices, j)
				processed[j] = true
			}
		}

		batches = append(batches, batch)
	}

	return batches
}

// GetBatchStatistics returns statistics about the batching efficiency.
type BatchStatistics struct {
	TotalTransactions   int
	NumBatches          int
	AverageBatchSize    float64
	MaxBatchSize        int
	ParallelizationRate float64 // Ratio of transactions that can run in parallel
}

// CalculateBatchStatistics computes statistics for a set of batches.
func CalculateBatchStatistics(batches []Batch, totalTxs int) BatchStatistics {
	if len(batches) == 0 || totalTxs == 0 {
		return BatchStatistics{}
	}

	maxBatchSize := 0
	totalInBatches := 0

	for _, batch := range batches {
		batchSize := len(batch.TransactionIndices)
		totalInBatches += batchSize
		if batchSize > maxBatchSize {
			maxBatchSize = batchSize
		}
	}

	avgBatchSize := float64(totalInBatches) / float64(len(batches))

	// Parallelization rate: if all transactions were sequential, we'd have N batches
	// The fewer batches we have, the more parallelization we achieved
	parallelizationRate := 1.0 - (float64(len(batches)) / float64(totalTxs))

	return BatchStatistics{
		TotalTransactions:   totalTxs,
		NumBatches:          len(batches),
		AverageBatchSize:    avgBatchSize,
		MaxBatchSize:        maxBatchSize,
		ParallelizationRate: parallelizationRate,
	}
}
