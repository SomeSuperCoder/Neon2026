package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"sync"

	_ "github.com/mattn/go-sqlite3"
	"github.com/poh-blockchain/internal/blockchain"
)

// Ledger manages persistent storage of blockchain data
type Ledger struct {
	db *sql.DB
	mu sync.RWMutex
}

// NewLedger creates a new Ledger instance and initializes the database
func NewLedger(dbPath string) (*Ledger, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	ledger := &Ledger{
		db: db,
	}

	// Initialize database schema
	if err := ledger.initSchema(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return ledger, nil
}

// initSchema creates the database tables if they don't exist
func (l *Ledger) initSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS blocks (
		block_height INTEGER PRIMARY KEY,
		block_hash BLOB NOT NULL,
		previous_hash BLOB NOT NULL,
		merkle_root BLOB NOT NULL,
		slot INTEGER NOT NULL,
		timestamp INTEGER NOT NULL,
		data TEXT NOT NULL
	);

	CREATE INDEX IF NOT EXISTS idx_blocks_hash ON blocks(block_hash);
	CREATE INDEX IF NOT EXISTS idx_blocks_slot ON blocks(slot);

	CREATE TABLE IF NOT EXISTS entries (
		entry_id INTEGER PRIMARY KEY AUTOINCREMENT,
		block_height INTEGER NOT NULL,
		hash BLOB NOT NULL,
		num_hashes INTEGER NOT NULL,
		previous_hash BLOB NOT NULL,
		timestamp INTEGER NOT NULL,
		transactions TEXT NOT NULL,
		FOREIGN KEY (block_height) REFERENCES blocks(block_height)
	);

	CREATE INDEX IF NOT EXISTS idx_entries_block_height ON entries(block_height);
	CREATE INDEX IF NOT EXISTS idx_entries_hash ON entries(hash);

	CREATE TABLE IF NOT EXISTS transactions (
		tx_id INTEGER PRIMARY KEY AUTOINCREMENT,
		entry_id INTEGER NOT NULL,
		sender TEXT NOT NULL,
		receiver TEXT NOT NULL,
		amount REAL NOT NULL,
		signature BLOB,
		data TEXT,
		FOREIGN KEY (entry_id) REFERENCES entries(entry_id)
	);

	CREATE INDEX IF NOT EXISTS idx_transactions_entry_id ON transactions(entry_id);
	CREATE INDEX IF NOT EXISTS idx_transactions_sender ON transactions(sender);
	CREATE INDEX IF NOT EXISTS idx_transactions_receiver ON transactions(receiver);
	`

	_, err := l.db.Exec(schema)
	return err
}

// StoreBlock persists a block and all its entries and transactions to the database
func (l *Ledger) StoreBlock(block blockchain.Block) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Start a transaction for atomicity
	tx, err := l.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Serialize the entire block to JSON for the data field
	blockJSON, err := json.Marshal(block)
	if err != nil {
		return fmt.Errorf("failed to serialize block: %w", err)
	}

	// Insert block record
	_, err = tx.Exec(`
		INSERT INTO blocks (block_height, block_hash, previous_hash, merkle_root, slot, timestamp, data)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, block.Header.BlockHeight, block.Header.MerkleRoot, block.Header.PreviousBlockHash,
		block.Header.MerkleRoot, block.Header.Slot, block.Header.Timestamp.Unix(), blockJSON)

	if err != nil {
		return fmt.Errorf("failed to insert block: %w", err)
	}

	// Insert entries
	for _, entry := range block.Entries {
		// Serialize transactions for this entry
		txJSON, err := json.Marshal(entry.Transactions)
		if err != nil {
			return fmt.Errorf("failed to serialize transactions: %w", err)
		}

		result, err := tx.Exec(`
			INSERT INTO entries (block_height, hash, num_hashes, previous_hash, timestamp, transactions)
			VALUES (?, ?, ?, ?, ?, ?)
		`, block.Header.BlockHeight, entry.Hash, entry.NumHashes, entry.PreviousHash,
			entry.Timestamp.Unix(), txJSON)

		if err != nil {
			return fmt.Errorf("failed to insert entry: %w", err)
		}

		entryID, err := result.LastInsertId()
		if err != nil {
			return fmt.Errorf("failed to get entry ID: %w", err)
		}

		// Insert individual transactions
		for _, transaction := range entry.Transactions {
			txDataJSON, err := json.Marshal(transaction.Data)
			if err != nil {
				return fmt.Errorf("failed to serialize transaction data: %w", err)
			}

			_, err = tx.Exec(`
				INSERT INTO transactions (entry_id, sender, receiver, amount, signature, data)
				VALUES (?, ?, ?, ?, ?, ?)
			`, entryID, transaction.Sender, transaction.Receiver, transaction.Amount,
				transaction.Signature, txDataJSON)

			if err != nil {
				return fmt.Errorf("failed to insert transaction: %w", err)
			}
		}
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetBlockByHeight retrieves a block by its height
func (l *Ledger) GetBlockByHeight(height int64) (blockchain.Block, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	var blockJSON string
	err := l.db.QueryRow(`
		SELECT data FROM blocks WHERE block_height = ?
	`, height).Scan(&blockJSON)

	if err != nil {
		if err == sql.ErrNoRows {
			return blockchain.Block{}, fmt.Errorf("block not found at height %d", height)
		}
		return blockchain.Block{}, fmt.Errorf("failed to query block: %w", err)
	}

	var block blockchain.Block
	if err := json.Unmarshal([]byte(blockJSON), &block); err != nil {
		return blockchain.Block{}, fmt.Errorf("failed to deserialize block: %w", err)
	}

	return block, nil
}

// GetBlockByHash retrieves a block by its hash
func (l *Ledger) GetBlockByHash(hash []byte) (blockchain.Block, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	var blockJSON string
	err := l.db.QueryRow(`
		SELECT data FROM blocks WHERE block_hash = ?
	`, hash).Scan(&blockJSON)

	if err != nil {
		if err == sql.ErrNoRows {
			return blockchain.Block{}, fmt.Errorf("block not found with hash %x", hash)
		}
		return blockchain.Block{}, fmt.Errorf("failed to query block: %w", err)
	}

	var block blockchain.Block
	if err := json.Unmarshal([]byte(blockJSON), &block); err != nil {
		return blockchain.Block{}, fmt.Errorf("failed to deserialize block: %w", err)
	}

	return block, nil
}

// GetLatestBlock returns the most recent block
func (l *Ledger) GetLatestBlock() (blockchain.Block, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	var blockJSON string
	err := l.db.QueryRow(`
		SELECT data FROM blocks ORDER BY block_height DESC LIMIT 1
	`).Scan(&blockJSON)

	if err != nil {
		if err == sql.ErrNoRows {
			return blockchain.Block{}, fmt.Errorf("no blocks found in ledger")
		}
		return blockchain.Block{}, fmt.Errorf("failed to query latest block: %w", err)
	}

	var block blockchain.Block
	if err := json.Unmarshal([]byte(blockJSON), &block); err != nil {
		return blockchain.Block{}, fmt.Errorf("failed to deserialize block: %w", err)
	}

	return block, nil
}

// GetChainHeight returns the current blockchain height
func (l *Ledger) GetChainHeight() (int64, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	var height int64
	err := l.db.QueryRow(`
		SELECT COALESCE(MAX(block_height), 0) FROM blocks
	`).Scan(&height)

	if err != nil {
		return 0, fmt.Errorf("failed to query chain height: %w", err)
	}

	return height, nil
}

// Close gracefully shuts down the database connection
func (l *Ledger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.db != nil {
		return l.db.Close()
	}
	return nil
}

// LoadBlockchainState loads the existing blockchain state on startup
// Returns the latest block and chain height, or error if no blocks exist
func (l *Ledger) LoadBlockchainState() (latestBlock blockchain.Block, height int64, err error) {
	// Get chain height
	height, err = l.GetChainHeight()
	if err != nil {
		return blockchain.Block{}, 0, fmt.Errorf("failed to get chain height: %w", err)
	}

	// If no blocks exist, return empty state
	if height == 0 {
		return blockchain.Block{}, 0, nil
	}

	// Get latest block
	latestBlock, err = l.GetLatestBlock()
	if err != nil {
		return blockchain.Block{}, 0, fmt.Errorf("failed to get latest block: %w", err)
	}

	return latestBlock, height, nil
}
