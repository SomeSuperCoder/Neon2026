package blockchain

import (
	"crypto/sha256"
	"sync"
	"time"

	"github.com/poh-blockchain/internal/poh"
)

// BlockProducer integrates transactions into the PoH clock and produces blocks
type BlockProducer struct {
	pohClock                *poh.PohClock
	pendingTransactions     []Transaction
	pendingFileTransactions [][]byte // Serialized file-based transactions
	currentSlot             int64
	entriesBuffer           []Entry
	previousEntryHash       []byte
	hashCountSinceEntry     int64
	mu                      sync.Mutex
}

// NewBlockProducer initializes a new BlockProducer with a PoH clock instance
func NewBlockProducer(pohClock *poh.PohClock) *BlockProducer {
	return &BlockProducer{
		pohClock:                pohClock,
		pendingTransactions:     make([]Transaction, 0),
		pendingFileTransactions: make([][]byte, 0),
		currentSlot:             0,
		entriesBuffer:           make([]Entry, 0),
		previousEntryHash:       pohClock.GetCurrentHash(),
		hashCountSinceEntry:     0,
	}
}

// AddTransaction queues a transaction for inclusion in the next entry
func (bp *BlockProducer) AddTransaction(tx Transaction) {
	bp.mu.Lock()
	defer bp.mu.Unlock()

	bp.pendingTransactions = append(bp.pendingTransactions, tx)
}

// AddFileTransaction queues a serialized file-based transaction for inclusion in the next entry
func (bp *BlockProducer) AddFileTransaction(txData []byte) {
	bp.mu.Lock()
	defer bp.mu.Unlock()

	bp.pendingFileTransactions = append(bp.pendingFileTransactions, txData)
}

// MixTransactionHash mixes a transaction hash into the PoH chain
func (bp *BlockProducer) MixTransactionHash(txHash []byte) {
	bp.mu.Lock()
	defer bp.mu.Unlock()

	// Get current hash from PoH clock
	currentHash := bp.pohClock.GetCurrentHash()

	// Combine current hash with transaction hash
	combined := append(currentHash, txHash...)

	// Hash the combination to create the mix
	_ = sha256.Sum256(combined)

	// Update the PoH clock by hashing once to incorporate the mix
	bp.pohClock.HashOnce()
	bp.hashCountSinceEntry++
}

// ProduceEntry creates an entry with pending transactions
func (bp *BlockProducer) ProduceEntry() (Entry, error) {
	bp.mu.Lock()
	defer bp.mu.Unlock()

	// Mix transaction hashes into the PoH chain if there are pending transactions
	for _, tx := range bp.pendingTransactions {
		// Create a hash of the transaction for mixing
		txData := append([]byte(tx.Sender), []byte(tx.Receiver)...)
		txHash := sha256.Sum256(txData)

		// Get current hash and mix with transaction hash
		currentHash := bp.pohClock.GetCurrentHash()
		combined := append(currentHash, txHash[:]...)
		sha256.Sum256(combined)

		// Hash once to incorporate the mix
		bp.pohClock.HashOnce()
		bp.hashCountSinceEntry++
	}

	// Mix file transaction hashes into the PoH chain
	for _, ftx := range bp.pendingFileTransactions {
		ftxHash := sha256.Sum256(ftx)

		// Get current hash and mix with transaction hash
		currentHash := bp.pohClock.GetCurrentHash()
		combined := append(currentHash, ftxHash[:]...)
		sha256.Sum256(combined)

		// Hash once to incorporate the mix
		bp.pohClock.HashOnce()
		bp.hashCountSinceEntry++
	}

	// Get the current hash state from PoH clock
	currentHash := bp.pohClock.GetCurrentHash()

	// Create the entry
	entry := Entry{
		Hash:             currentHash,
		NumHashes:        bp.hashCountSinceEntry,
		Transactions:     bp.pendingTransactions,
		FileTransactions: bp.pendingFileTransactions,
		PreviousHash:     bp.previousEntryHash,
		Timestamp:        time.Now(),
	}

	// Add entry to buffer
	bp.entriesBuffer = append(bp.entriesBuffer, entry)

	// Update state for next entry
	bp.previousEntryHash = currentHash
	bp.hashCountSinceEntry = 0

	// Clear pending transactions
	bp.pendingTransactions = make([]Transaction, 0)
	bp.pendingFileTransactions = make([][]byte, 0)

	return entry, nil
}

// ProduceBlock produces a complete block for the given slot with the provided state root
func (bp *BlockProducer) ProduceBlock(slot int64, stateRoot []byte) (Block, error) {
	bp.mu.Lock()
	defer bp.mu.Unlock()

	const minTicksPerBlock = 64
	const hashesPerTick = 12500
	minHashesRequired := int64(minTicksPerBlock * hashesPerTick)

	// Generate enough hashes to meet the minimum tick requirement
	totalHashes := int64(0)
	for _, entry := range bp.entriesBuffer {
		totalHashes += entry.NumHashes
	}

	// Continue generating ticks until we have enough hashes
	for totalHashes < minHashesRequired {
		// Generate a tick (12,500 hashes)
		bp.pohClock.Tick()
		bp.hashCountSinceEntry += hashesPerTick

		// Create an entry for this tick
		currentHash := bp.pohClock.GetCurrentHash()
		entry := Entry{
			Hash:             currentHash,
			NumHashes:        bp.hashCountSinceEntry,
			Transactions:     []Transaction{}, // No transactions for tick-only entries
			FileTransactions: [][]byte{},      // No file transactions for tick-only entries
			PreviousHash:     bp.previousEntryHash,
			Timestamp:        time.Now(),
		}

		bp.entriesBuffer = append(bp.entriesBuffer, entry)
		bp.previousEntryHash = currentHash
		bp.hashCountSinceEntry = 0

		totalHashes += hashesPerTick
	}

	// Calculate Merkle root from entries
	merkleRoot := calculateMerkleRoot(bp.entriesBuffer)

	// Create block header with state root
	header := BlockHeader{
		PreviousBlockHash: []byte{}, // Will be set by caller if needed
		MerkleRoot:        merkleRoot,
		StateRoot:         stateRoot,
		Slot:              slot,
		Timestamp:         time.Now(),
		BlockHeight:       0, // Will be set by caller
	}

	// Create the block with default version
	block := Block{
		Version: BlockVersion1,
		Header:  header,
		Entries: bp.entriesBuffer,
	}

	// Update current slot
	bp.currentSlot = slot

	// Clear entries buffer for next block
	bp.entriesBuffer = make([]Entry, 0)

	return block, nil
}

// calculateMerkleRoot computes a simple Merkle root from entries
func calculateMerkleRoot(entries []Entry) []byte {
	if len(entries) == 0 {
		return []byte{}
	}

	// Collect all entry hashes
	hashes := make([][]byte, len(entries))
	for i, entry := range entries {
		hashes[i] = entry.Hash
	}

	// Build Merkle tree by repeatedly hashing pairs
	for len(hashes) > 1 {
		var nextLevel [][]byte

		for i := 0; i < len(hashes); i += 2 {
			if i+1 < len(hashes) {
				// Hash pair
				combined := append(hashes[i], hashes[i+1]...)
				hash := sha256.Sum256(combined)
				nextLevel = append(nextLevel, hash[:])
			} else {
				// Odd one out, hash with itself
				combined := append(hashes[i], hashes[i]...)
				hash := sha256.Sum256(combined)
				nextLevel = append(nextLevel, hash[:])
			}
		}

		hashes = nextLevel
	}

	return hashes[0]
}
