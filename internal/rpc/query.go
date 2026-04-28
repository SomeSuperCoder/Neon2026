package rpc

import (
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"github.com/poh-blockchain/internal/blockchain"
	"github.com/poh-blockchain/internal/filestore"
	"github.com/poh-blockchain/internal/storage"
	"github.com/poh-blockchain/internal/transaction"
)

// QueryEngine provides methods to query blockchain data from ledger and filestore
type QueryEngine struct {
	ledger    *storage.Ledger
	fileStore *filestore.FileStore
	cache     *QueryCache
	mu        sync.RWMutex
}

// QueryCache stores frequently accessed data
type QueryCache struct {
	blockHeight     *int64
	blockHeightTime time.Time
	recentBlockhash *string
	blockhashTime   time.Time
	cacheDuration   time.Duration
	mu              sync.RWMutex
}

// NewQueryEngine creates a new query engine with ledger and filestore access
func NewQueryEngine(ledger *storage.Ledger, fileStore *filestore.FileStore) *QueryEngine {
	return &QueryEngine{
		ledger:    ledger,
		fileStore: fileStore,
		cache: &QueryCache{
			cacheDuration: 2 * time.Second, // Cache for 2 seconds
		},
	}
}

// GetBalance retrieves the balance for a given address
// Returns error if address is invalid or account doesn't exist
func (q *QueryEngine) GetBalance(address string) (int64, error) {
	// Validate and parse address
	fileID, err := filestore.FileIDFromString(address)
	if err != nil {
		return 0, fmt.Errorf("invalid address: %w", err)
	}

	// Get file from filestore
	file, err := q.fileStore.GetFile(fileID)
	if err != nil {
		return 0, fmt.Errorf("account not found: %w", err)
	}

	return file.Balance, nil
}

// GetAccountInfo retrieves full account information for a given address
// Returns nil if account doesn't exist (per requirement 2.6)
func (q *QueryEngine) GetAccountInfo(address string) (*AccountInfo, error) {
	// Validate and parse address
	fileID, err := filestore.FileIDFromString(address)
	if err != nil {
		return nil, fmt.Errorf("invalid address: %w", err)
	}

	// Get file from filestore
	file, err := q.fileStore.GetFile(fileID)
	if err != nil {
		// Return nil for non-existent accounts (requirement 2.6)
		return nil, nil
	}

	// Convert to AccountInfo
	info := &AccountInfo{
		Address:    file.ID.String(),
		Balance:    file.Balance,
		Owner:      file.TxManager.String(),
		DataLength: len(file.Data),
		Executable: file.Executable,
	}

	return info, nil
}

// GetBlockHeight retrieves the current blockchain height
func (q *QueryEngine) GetBlockHeight() (int64, error) {
	// Check cache first
	q.cache.mu.RLock()
	if q.cache.blockHeight != nil && time.Since(q.cache.blockHeightTime) < q.cache.cacheDuration {
		height := *q.cache.blockHeight
		q.cache.mu.RUnlock()
		return height, nil
	}
	q.cache.mu.RUnlock()

	// Query from ledger
	height, err := q.ledger.GetChainHeight()
	if err != nil {
		return 0, fmt.Errorf("failed to get chain height: %w", err)
	}

	// Update cache
	q.cache.mu.Lock()
	q.cache.blockHeight = &height
	q.cache.blockHeightTime = time.Now()
	q.cache.mu.Unlock()

	return height, nil
}

// GetRecentBlockhash retrieves the hash of the most recent block
func (q *QueryEngine) GetRecentBlockhash() (string, error) {
	// Check cache first
	q.cache.mu.RLock()
	if q.cache.recentBlockhash != nil && time.Since(q.cache.blockhashTime) < q.cache.cacheDuration {
		blockhash := *q.cache.recentBlockhash
		q.cache.mu.RUnlock()
		return blockhash, nil
	}
	q.cache.mu.RUnlock()

	// Query from ledger
	block, err := q.ledger.GetLatestBlock()
	if err != nil {
		return "", fmt.Errorf("failed to get latest block: %w", err)
	}

	// Use MerkleRoot as the block hash
	blockhash := hex.EncodeToString(block.Header.MerkleRoot)

	// Update cache
	q.cache.mu.Lock()
	q.cache.recentBlockhash = &blockhash
	q.cache.blockhashTime = time.Now()
	q.cache.mu.Unlock()

	return blockhash, nil
}

// GetTransactionHistory retrieves transaction history for an address with pagination
// Returns transactions in reverse chronological order
func (q *QueryEngine) GetTransactionHistory(address string, limit int) ([]TransactionRecord, error) {
	// Validate address
	fileID, err := filestore.FileIDFromString(address)
	if err != nil {
		return nil, fmt.Errorf("invalid address: %w", err)
	}

	// Default limit
	if limit <= 0 {
		limit = 20
	}

	// Get chain height to iterate backwards
	height, err := q.ledger.GetChainHeight()
	if err != nil {
		return nil, fmt.Errorf("failed to get chain height: %w", err)
	}

	var records []TransactionRecord

	// Iterate through blocks in reverse order
	for blockHeight := height; blockHeight > 0 && len(records) < limit; blockHeight-- {
		block, err := q.ledger.GetBlockByHeight(blockHeight)
		if err != nil {
			// Skip blocks that can't be retrieved
			continue
		}

		// Process entries in reverse order
		for i := len(block.Entries) - 1; i >= 0 && len(records) < limit; i-- {
			entry := block.Entries[i]

			// Check FileTransactions first (new format)
			if len(entry.FileTransactions) > 0 {
				for j := len(entry.FileTransactions) - 1; j >= 0 && len(records) < limit; j-- {
					txBytes := entry.FileTransactions[j]
					tx, err := transaction.UnmarshalTransaction(txBytes)
					if err != nil {
						continue
					}

					// Check if transaction involves this address
					if q.transactionInvolvesAddress(tx, fileID) {
						record := q.convertFileTransactionToRecord(tx, block.Header, entry)
						records = append(records, record)
					}
				}
			}

			// Also check legacy Transactions format for backwards compatibility
			for j := len(entry.Transactions) - 1; j >= 0 && len(records) < limit; j-- {
				legacyTx := entry.Transactions[j]

				// Check if legacy transaction involves this address
				if legacyTx.Sender == address || legacyTx.Receiver == address {
					record := q.convertLegacyTransactionToRecord(legacyTx, block.Header, entry)
					records = append(records, record)
				}
			}
		}
	}

	return records, nil
}

// transactionInvolvesAddress checks if a transaction involves the given address
func (q *QueryEngine) transactionInvolvesAddress(tx *transaction.Transaction, address filestore.FileID) bool {
	// Check signers
	for _, sig := range tx.Signatures {
		// Convert PublicKey to FileID for comparison
		pkFileID := filestore.GenerateFileID(sig.PublicKey[:])
		if pkFileID == address {
			return true
		}
	}

	// Check instruction inputs
	for _, instr := range tx.Instructions {
		for _, access := range instr.Inputs {
			if access.FileID == address {
				return true
			}
		}
	}

	return false
}

// convertFileTransactionToRecord converts a file-based transaction to a TransactionRecord
func (q *QueryEngine) convertFileTransactionToRecord(tx *transaction.Transaction, header blockchain.BlockHeader, entry blockchain.Entry) TransactionRecord {
	// Generate transaction signature from first signature
	var sigStr string
	if len(tx.Signatures) > 0 {
		sigStr = hex.EncodeToString(tx.Signatures[0].Signature[:])
	}

	// Convert instructions
	var instructions []InstructionRecord
	for _, instr := range tx.Instructions {
		instrRecord := InstructionRecord{
			ProgramID: instr.ProgramID.String(),
			Type:      q.inferInstructionType(instr),
			Accounts:  q.extractAccountsFromInstruction(instr),
			Data:      hex.EncodeToString(instr.Data),
		}
		instructions = append(instructions, instrRecord)
	}

	return TransactionRecord{
		Signature:    sigStr,
		BlockHeight:  header.BlockHeight,
		Slot:         header.Slot,
		Timestamp:    entry.Timestamp,
		Success:      true, // Assume success if in ledger
		Instructions: instructions,
	}
}

// convertLegacyTransactionToRecord converts a legacy transaction to a TransactionRecord
func (q *QueryEngine) convertLegacyTransactionToRecord(tx blockchain.Transaction, header blockchain.BlockHeader, entry blockchain.Entry) TransactionRecord {
	sigStr := hex.EncodeToString(tx.Signature)

	// Create a simple instruction record for legacy format
	instrRecord := InstructionRecord{
		ProgramID: "system",
		Type:      "transfer",
		Accounts:  []string{tx.Sender, tx.Receiver},
		Data:      fmt.Sprintf("amount:%f", tx.Amount),
	}

	return TransactionRecord{
		Signature:    sigStr,
		BlockHeight:  header.BlockHeight,
		Slot:         header.Slot,
		Timestamp:    entry.Timestamp,
		Success:      true,
		Instructions: []InstructionRecord{instrRecord},
	}
}

// inferInstructionType attempts to determine the instruction type from data
func (q *QueryEngine) inferInstructionType(instr transaction.Instruction) string {
	// Simple heuristic based on data length and program
	if len(instr.Data) == 0 {
		return "unknown"
	}

	// First byte often indicates instruction type
	if len(instr.Data) > 0 {
		switch instr.Data[0] {
		case 0:
			return "transfer"
		case 1:
			return "create_account"
		case 2:
			return "allocate"
		default:
			return "invoke"
		}
	}

	return "unknown"
}

// extractAccountsFromInstruction extracts account addresses from instruction inputs
func (q *QueryEngine) extractAccountsFromInstruction(instr transaction.Instruction) []string {
	var accounts []string
	for _, access := range instr.Inputs {
		accounts = append(accounts, access.FileID.String())
	}
	return accounts
}

// GetTransactionStatus retrieves the confirmation status of a transaction by signature
func (q *QueryEngine) GetTransactionStatus(signature string) (*TransactionStatus, error) {
	// Decode signature
	sigBytes, err := hex.DecodeString(signature)
	if err != nil {
		return nil, fmt.Errorf("invalid signature format: %w", err)
	}

	// Get chain height
	height, err := q.ledger.GetChainHeight()
	if err != nil {
		return nil, fmt.Errorf("failed to get chain height: %w", err)
	}

	// Search through blocks for the transaction
	for blockHeight := height; blockHeight > 0; blockHeight-- {
		block, err := q.ledger.GetBlockByHeight(blockHeight)
		if err != nil {
			continue
		}

		// Check entries for matching transaction
		for _, entry := range block.Entries {
			// Check FileTransactions
			for _, txBytes := range entry.FileTransactions {
				tx, err := transaction.UnmarshalTransaction(txBytes)
				if err != nil {
					continue
				}

				// Check if signature matches
				if len(tx.Signatures) > 0 {
					if string(tx.Signatures[0].Signature[:]) == string(sigBytes) {
						return &TransactionStatus{
							Signature:   signature,
							Confirmed:   true,
							BlockHeight: block.Header.BlockHeight,
							Slot:        block.Header.Slot,
						}, nil
					}
				}
			}

			// Check legacy transactions
			for _, legacyTx := range entry.Transactions {
				if string(legacyTx.Signature) == string(sigBytes) {
					return &TransactionStatus{
						Signature:   signature,
						Confirmed:   true,
						BlockHeight: block.Header.BlockHeight,
						Slot:        block.Header.Slot,
					}, nil
				}
			}
		}
	}

	// Transaction not found
	return &TransactionStatus{
		Signature: signature,
		Confirmed: false,
	}, nil
}
