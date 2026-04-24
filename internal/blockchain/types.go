package blockchain

import "time"

// Transaction represents a blockchain transaction
type Transaction struct {
	Sender    string                 `json:"sender"`
	Receiver  string                 `json:"receiver"`
	Amount    float64                `json:"amount"`
	Signature []byte                 `json:"signature"`
	Data      map[string]interface{} `json:"data"`
}

// Entry represents a ledger record in the blockchain
type Entry struct {
	Hash             []byte        `json:"hash"`
	NumHashes        int64         `json:"num_hashes"`
	Transactions     []Transaction `json:"transactions"`
	FileTransactions [][]byte      `json:"file_transactions,omitempty"` // Serialized file-based transactions
	PreviousHash     []byte        `json:"previous_hash"`
	Timestamp        time.Time     `json:"timestamp"`
}

// BlockHeader contains metadata for a block
type BlockHeader struct {
	PreviousBlockHash []byte    `json:"previous_block_hash"`
	MerkleRoot        []byte    `json:"merkle_root"`
	StateRoot         []byte    `json:"state_root"` // Root hash of file state tree
	Slot              int64     `json:"slot"`
	Timestamp         time.Time `json:"timestamp"`
	BlockHeight       int64     `json:"block_height"`
}

// Block represents a collection of entries produced during a slot
type Block struct {
	Version uint32      `json:"version"` // Block format version for backwards compatibility
	Header  BlockHeader `json:"header"`
	Entries []Entry     `json:"entries"`
}

const (
	// BlockVersion1 is the initial block format version
	BlockVersion1 uint32 = 1
)

// SetVersion sets a custom version for the block
// This should only be used when explicitly upgrading block format
func (b *Block) SetVersion(version uint32) {
	b.Version = version
}

// GetVersion returns the block version
func (b *Block) GetVersion() uint32 {
	if b.Version == 0 {
		// For backwards compatibility with blocks created before versioning
		return BlockVersion1
	}
	return b.Version
}
