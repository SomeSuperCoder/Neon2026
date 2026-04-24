package blockchain

import (
	"encoding/hex"
	"encoding/json"
	"time"
)

// transactionJSON is used for custom JSON marshaling of Transaction
type transactionJSON struct {
	Sender    string                 `json:"sender"`
	Receiver  string                 `json:"receiver"`
	Amount    float64                `json:"amount"`
	Signature string                 `json:"signature"`
	Data      map[string]interface{} `json:"data"`
}

// MarshalJSON implements custom JSON marshaling for Transaction
func (t Transaction) MarshalJSON() ([]byte, error) {
	return json.Marshal(transactionJSON{
		Sender:    t.Sender,
		Receiver:  t.Receiver,
		Amount:    t.Amount,
		Signature: hex.EncodeToString(t.Signature),
		Data:      t.Data,
	})
}

// UnmarshalJSON implements custom JSON unmarshaling for Transaction
func (t *Transaction) UnmarshalJSON(data []byte) error {
	var tj transactionJSON
	if err := json.Unmarshal(data, &tj); err != nil {
		return err
	}

	sig, err := hex.DecodeString(tj.Signature)
	if err != nil {
		return err
	}

	t.Sender = tj.Sender
	t.Receiver = tj.Receiver
	t.Amount = tj.Amount
	t.Signature = sig
	t.Data = tj.Data

	return nil
}

// entryJSON is used for custom JSON marshaling of Entry
type entryJSON struct {
	Hash         string        `json:"hash"`
	NumHashes    int64         `json:"num_hashes"`
	Transactions []Transaction `json:"transactions"`
	PreviousHash string        `json:"previous_hash"`
	Timestamp    time.Time     `json:"timestamp"`
}

// MarshalJSON implements custom JSON marshaling for Entry
func (e Entry) MarshalJSON() ([]byte, error) {
	return json.Marshal(entryJSON{
		Hash:         hex.EncodeToString(e.Hash),
		NumHashes:    e.NumHashes,
		Transactions: e.Transactions,
		PreviousHash: hex.EncodeToString(e.PreviousHash),
		Timestamp:    e.Timestamp,
	})
}

// UnmarshalJSON implements custom JSON unmarshaling for Entry
func (e *Entry) UnmarshalJSON(data []byte) error {
	var ej entryJSON
	if err := json.Unmarshal(data, &ej); err != nil {
		return err
	}

	hash, err := hex.DecodeString(ej.Hash)
	if err != nil {
		return err
	}

	prevHash, err := hex.DecodeString(ej.PreviousHash)
	if err != nil {
		return err
	}

	e.Hash = hash
	e.NumHashes = ej.NumHashes
	e.Transactions = ej.Transactions
	e.PreviousHash = prevHash
	e.Timestamp = ej.Timestamp

	return nil
}

// blockHeaderJSON is used for custom JSON marshaling of BlockHeader
type blockHeaderJSON struct {
	PreviousBlockHash string    `json:"previous_block_hash"`
	MerkleRoot        string    `json:"merkle_root"`
	Slot              int64     `json:"slot"`
	Timestamp         time.Time `json:"timestamp"`
	BlockHeight       int64     `json:"block_height"`
}

// MarshalJSON implements custom JSON marshaling for BlockHeader
func (bh BlockHeader) MarshalJSON() ([]byte, error) {
	return json.Marshal(blockHeaderJSON{
		PreviousBlockHash: hex.EncodeToString(bh.PreviousBlockHash),
		MerkleRoot:        hex.EncodeToString(bh.MerkleRoot),
		Slot:              bh.Slot,
		Timestamp:         bh.Timestamp,
		BlockHeight:       bh.BlockHeight,
	})
}

// UnmarshalJSON implements custom JSON unmarshaling for BlockHeader
func (bh *BlockHeader) UnmarshalJSON(data []byte) error {
	var bhj blockHeaderJSON
	if err := json.Unmarshal(data, &bhj); err != nil {
		return err
	}

	prevBlockHash, err := hex.DecodeString(bhj.PreviousBlockHash)
	if err != nil {
		return err
	}

	merkleRoot, err := hex.DecodeString(bhj.MerkleRoot)
	if err != nil {
		return err
	}

	bh.PreviousBlockHash = prevBlockHash
	bh.MerkleRoot = merkleRoot
	bh.Slot = bhj.Slot
	bh.Timestamp = bhj.Timestamp
	bh.BlockHeight = bhj.BlockHeight

	return nil
}

// blockJSON is used for custom JSON marshaling of Block
type blockJSON struct {
	Version uint32      `json:"version"`
	Header  BlockHeader `json:"header"`
	Entries []Entry     `json:"entries"`
}

// MarshalJSON implements custom JSON marshaling for Block
func (b Block) MarshalJSON() ([]byte, error) {
	version := b.Version
	if version == 0 {
		version = BlockVersion1
	}

	return json.Marshal(blockJSON{
		Version: version,
		Header:  b.Header,
		Entries: b.Entries,
	})
}

// UnmarshalJSON implements custom JSON unmarshaling for Block
func (b *Block) UnmarshalJSON(data []byte) error {
	var bj blockJSON
	if err := json.Unmarshal(data, &bj); err != nil {
		return err
	}

	// Default to version 1 if not specified (backwards compatibility)
	if bj.Version == 0 {
		bj.Version = BlockVersion1
	}

	b.Version = bj.Version
	b.Header = bj.Header
	b.Entries = bj.Entries

	return nil
}
