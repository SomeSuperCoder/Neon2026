package verification

import (
	"bytes"
	"crypto/sha256"
	"fmt"

	"github.com/poh-blockchain/internal/blockchain"
	"github.com/poh-blockchain/internal/storage"
)

// Verifier verifies blockchain integrity and validity
type Verifier struct{}

// NewVerifier creates a new verifier instance
func NewVerifier() *Verifier {
	return &Verifier{}
}

// VerifyEntry verifies an entry's hash chain
func (v *Verifier) VerifyEntry(entry blockchain.Entry) error {
	// Verify that the entry has a valid hash
	if len(entry.Hash) == 0 {
		return fmt.Errorf("entry hash is empty")
	}

	// Verify that the entry has a previous hash (except for genesis entry)
	if len(entry.PreviousHash) == 0 {
		// This could be a genesis entry, which is acceptable
		return nil
	}

	// Verify that the entry hash correctly links to the previous hash
	// by recreating the hash chain
	currentHash := make([]byte, len(entry.PreviousHash))
	copy(currentHash, entry.PreviousHash)

	// Perform num_hashes iterations to reach the entry hash
	for i := int64(0); i < entry.NumHashes; i++ {
		hash := sha256.Sum256(currentHash)
		currentHash = hash[:]
	}

	// Compare the final hash with the entry hash
	if !bytes.Equal(currentHash, entry.Hash) {
		return fmt.Errorf("entry hash does not match: expected %x, got %x after %d hashes",
			entry.Hash, currentHash, entry.NumHashes)
	}

	return nil
}

// VerifyHashCount verifies that the num_hashes count is accurate
func (v *Verifier) VerifyHashCount(entry blockchain.Entry) error {
	// Verify that num_hashes is positive
	if entry.NumHashes <= 0 {
		return fmt.Errorf("invalid num_hashes: %d (must be positive)", entry.NumHashes)
	}

	// Verify that the entry has a previous hash
	if len(entry.PreviousHash) == 0 {
		// Genesis entry, skip hash count verification
		return nil
	}

	// Recreate the hash chain from previous hash for num_hashes iterations
	currentHash := make([]byte, len(entry.PreviousHash))
	copy(currentHash, entry.PreviousHash)

	for i := int64(0); i < entry.NumHashes; i++ {
		hash := sha256.Sum256(currentHash)
		currentHash = hash[:]
	}

	// Compare the final hash with the entry hash
	if !bytes.Equal(currentHash, entry.Hash) {
		return fmt.Errorf("hash count verification failed: expected %x, got %x after %d hashes",
			entry.Hash, currentHash, entry.NumHashes)
	}

	return nil
}

// VerifyBlock verifies a single block's validity
func (v *Verifier) VerifyBlock(block blockchain.Block) error {
	// Verify that the block has entries
	if len(block.Entries) == 0 {
		return fmt.Errorf("block has no entries")
	}

	// Verify all entries in the block
	for i, entry := range block.Entries {
		if err := v.VerifyEntry(entry); err != nil {
			return fmt.Errorf("entry %d verification failed: %w", i, err)
		}
	}

	// Verify that entries are properly linked within the block
	for i := 1; i < len(block.Entries); i++ {
		if !bytes.Equal(block.Entries[i].PreviousHash, block.Entries[i-1].Hash) {
			return fmt.Errorf("entry %d is not properly linked to entry %d", i, i-1)
		}
	}

	// Calculate the Merkle root from entries and verify it matches the header
	calculatedMerkleRoot := calculateMerkleRoot(block.Entries)
	if !bytes.Equal(calculatedMerkleRoot, block.Header.MerkleRoot) {
		return fmt.Errorf("merkle root mismatch: expected %x, got %x",
			block.Header.MerkleRoot, calculatedMerkleRoot)
	}

	// Verify that the block contains at least 64 ticks worth of hash operations
	const minTicksPerBlock = 64
	const hashesPerTick = 12500
	minHashesRequired := int64(minTicksPerBlock * hashesPerTick)

	totalHashes := int64(0)
	for _, entry := range block.Entries {
		totalHashes += entry.NumHashes
	}

	if totalHashes < minHashesRequired {
		return fmt.Errorf("block does not contain enough hashes: %d (minimum %d required)",
			totalHashes, minHashesRequired)
	}

	// Note: State root verification requires access to FileStore and is performed
	// separately via VerifyStateTransition method

	return nil
}

// calculateMerkleRoot computes a simple Merkle root from entries
func calculateMerkleRoot(entries []blockchain.Entry) []byte {
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

// VerifyBlockLink verifies that a block is properly linked to the previous block
func (v *Verifier) VerifyBlockLink(block blockchain.Block, previousBlock blockchain.Block) error {
	// Verify that the block header previous hash matches the previous block's merkle root
	// (In this implementation, we use merkle root as the block identifier)
	if !bytes.Equal(block.Header.PreviousBlockHash, previousBlock.Header.MerkleRoot) {
		return fmt.Errorf("block header previous hash does not match previous block merkle root: expected %x, got %x",
			previousBlock.Header.MerkleRoot, block.Header.PreviousBlockHash)
	}

	// Verify that the block starts with hash from previous block's final entry
	if len(previousBlock.Entries) > 0 && len(block.Entries) > 0 {
		lastEntryOfPreviousBlock := previousBlock.Entries[len(previousBlock.Entries)-1]
		firstEntryOfCurrentBlock := block.Entries[0]

		if !bytes.Equal(firstEntryOfCurrentBlock.PreviousHash, lastEntryOfPreviousBlock.Hash) {
			return fmt.Errorf("block does not start with hash from previous block's final entry: expected %x, got %x",
				lastEntryOfPreviousBlock.Hash, firstEntryOfCurrentBlock.PreviousHash)
		}
	}

	// Verify that block height is sequential
	if block.Header.BlockHeight != previousBlock.Header.BlockHeight+1 {
		return fmt.Errorf("block height is not sequential: expected %d, got %d",
			previousBlock.Header.BlockHeight+1, block.Header.BlockHeight)
	}

	return nil
}

// VerifyChain verifies the entire blockchain from genesis to current height
func (v *Verifier) VerifyChain(ledger *storage.Ledger) error {
	// Get the current chain height
	height, err := ledger.GetChainHeight()
	if err != nil {
		return fmt.Errorf("failed to get chain height: %w", err)
	}

	// If no blocks exist, nothing to verify
	if height == 0 {
		return nil
	}

	// Verify genesis block (height 0 or 1, depending on implementation)
	genesisBlock, err := ledger.GetBlockByHeight(0)
	if err != nil {
		// Try height 1 if height 0 doesn't exist
		genesisBlock, err = ledger.GetBlockByHeight(1)
		if err != nil {
			return fmt.Errorf("failed to get genesis block: %w", err)
		}
		height = 1 // Adjust starting height
	}

	// Verify the genesis block
	if err := v.VerifyBlock(genesisBlock); err != nil {
		return fmt.Errorf("genesis block verification failed: %w", err)
	}

	// Iterate through all blocks from genesis to current height
	previousBlock := genesisBlock
	startHeight := genesisBlock.Header.BlockHeight + 1

	for i := startHeight; i <= height; i++ {
		currentBlock, err := ledger.GetBlockByHeight(i)
		if err != nil {
			return fmt.Errorf("failed to get block at height %d: %w", i, err)
		}

		// Verify the current block
		if err := v.VerifyBlock(currentBlock); err != nil {
			return fmt.Errorf("block %d verification failed: %w", i, err)
		}

		// Verify the linkage to the previous block
		if err := v.VerifyBlockLink(currentBlock, previousBlock); err != nil {
			return fmt.Errorf("block %d linkage verification failed: %w", i, err)
		}

		previousBlock = currentBlock
	}

	return nil
}

// StateRootVerifier is an interface for verifying state roots
// This allows the verifier to work with different state implementations
type StateRootVerifier interface {
	CalculateStateRoot() ([]byte, error)
}

// VerifyStateRoot verifies that a block's state root matches the actual state
func (v *Verifier) VerifyStateRoot(block blockchain.Block, stateVerifier StateRootVerifier) error {
	// Calculate the actual state root
	actualStateRoot, err := stateVerifier.CalculateStateRoot()
	if err != nil {
		return fmt.Errorf("failed to calculate state root: %w", err)
	}

	// Compare with the block's state root
	if !bytes.Equal(block.Header.StateRoot, actualStateRoot) {
		return fmt.Errorf("state root mismatch: expected %x, got %x",
			block.Header.StateRoot, actualStateRoot)
	}

	return nil
}

// VerifyStateTransition verifies that state transitions in a block are valid
// This should be called after processing all transactions in the block
func (v *Verifier) VerifyStateTransition(block blockchain.Block, stateVerifier StateRootVerifier) error {
	// Verify that the state root in the block header matches the actual state
	return v.VerifyStateRoot(block, stateVerifier)
}
