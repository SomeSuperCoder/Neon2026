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
	Hash         []byte        `json:"hash"`
	NumHashes    int64         `json:"num_hashes"`
	Transactions []Transaction `json:"transactions"`
	PreviousHash []byte        `json:"previous_hash"`
	Timestamp    time.Time     `json:"timestamp"`
}

// BlockHeader contains metadata for a block
type BlockHeader struct {
	PreviousBlockHash []byte    `json:"previous_block_hash"`
	MerkleRoot        []byte    `json:"merkle_root"`
	Slot              int64     `json:"slot"`
	Timestamp         time.Time `json:"timestamp"`
	BlockHeight       int64     `json:"block_height"`
}

// Block represents a collection of entries produced during a slot
type Block struct {
	Header  BlockHeader `json:"header"`
	Entries []Entry     `json:"entries"`
}
