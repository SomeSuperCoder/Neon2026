package filestore

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/dgraph-io/badger/v4"
)

// Package filestore manages all on-chain file state with efficient storage and retrieval.
// It provides the core FileStore implementation using BadgerDB for persistence.

// FileID is a unique identifier for each file on the blockchain (32-byte SHA-256 hash)
type FileID [32]byte

// File represents the fundamental unit of on-chain state
type File struct {
	ID         FileID
	Balance    int64
	TxManager  FileID
	Data       []byte
	Executable bool
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// fileJSON is used for JSON serialization/deserialization
type fileJSON struct {
	ID         string `json:"id"`
	Balance    int64  `json:"balance"`
	TxManager  string `json:"tx_manager"`
	Data       []byte `json:"data"`
	Executable bool   `json:"executable"`
	CreatedAt  int64  `json:"created_at"`
	UpdatedAt  int64  `json:"updated_at"`
}

// String returns the hex-encoded string representation of FileID
func (fid FileID) String() string {
	return hex.EncodeToString(fid[:])
}

// FileIDFromString creates a FileID from a hex-encoded string
func FileIDFromString(s string) (FileID, error) {
	var fid FileID
	bytes, err := hex.DecodeString(s)
	if err != nil {
		return fid, fmt.Errorf("invalid hex string: %w", err)
	}
	if len(bytes) != 32 {
		return fid, fmt.Errorf("invalid FileID length: expected 32 bytes, got %d", len(bytes))
	}
	copy(fid[:], bytes)
	return fid, nil
}

// FileIDFromBytes creates a FileID from a byte slice
func FileIDFromBytes(b []byte) (FileID, error) {
	var fid FileID
	if len(b) != 32 {
		return fid, fmt.Errorf("invalid FileID length: expected 32 bytes, got %d", len(b))
	}
	copy(fid[:], b)
	return fid, nil
}

// GenerateFileID generates a new FileID from input data using SHA-256
func GenerateFileID(data []byte) FileID {
	return sha256.Sum256(data)
}

// Marshal serializes a File to JSON bytes
func (f *File) Marshal() ([]byte, error) {
	fj := fileJSON{
		ID:         f.ID.String(),
		Balance:    f.Balance,
		TxManager:  f.TxManager.String(),
		Data:       f.Data,
		Executable: f.Executable,
		CreatedAt:  f.CreatedAt.Unix(),
		UpdatedAt:  f.UpdatedAt.Unix(),
	}
	return json.Marshal(&fj)
}

// Unmarshal deserializes a File from JSON bytes
func (f *File) Unmarshal(data []byte) error {
	var fj fileJSON
	if err := json.Unmarshal(data, &fj); err != nil {
		return fmt.Errorf("failed to unmarshal file: %w", err)
	}

	id, err := FileIDFromString(fj.ID)
	if err != nil {
		return fmt.Errorf("invalid file ID: %w", err)
	}
	f.ID = id

	txManager, err := FileIDFromString(fj.TxManager)
	if err != nil {
		return fmt.Errorf("invalid tx manager ID: %w", err)
	}
	f.TxManager = txManager

	f.Balance = fj.Balance
	f.Data = fj.Data
	f.Executable = fj.Executable
	f.CreatedAt = time.Unix(fj.CreatedAt, 0)
	f.UpdatedAt = time.Unix(fj.UpdatedAt, 0)

	return nil
}

// UnmarshalFile is a convenience function to unmarshal a File from bytes
func UnmarshalFile(data []byte) (*File, error) {
	f := &File{}
	if err := f.Unmarshal(data); err != nil {
		return nil, err
	}
	return f, nil
}

// Storage cost constants
const (
	BaseCostPerKB  = int64(1000) // Base cost: 1000 units per KB
	CostGrowthRate = 1.1         // Exponential growth rate
)

// CalculateStorageCost calculates the storage cost for a given data size
// using an exponential growth formula: cost = base * size_in_kb * (1.1 ^ size_in_mb)
// This enforces economic constraints on file sizes as specified in Requirement 6.2
func CalculateStorageCost(dataSize int64) int64 {
	if dataSize <= 0 {
		return 0
	}

	sizeInKB := BytesToKB(dataSize)
	if sizeInKB == 0 {
		sizeInKB = 1 // Minimum 1 KB for any non-zero data
	}

	sizeInMB := KBToMB(sizeInKB)
	multiplier := math.Pow(CostGrowthRate, sizeInMB)

	cost := int64(float64(BaseCostPerKB*sizeInKB) * multiplier)
	return cost
}

// ValidateStorageCost checks if a file's balance covers the storage cost for its data size
// This implements Requirements 6.1, 6.3, and 6.4
func ValidateStorageCost(file *File) error {
	if file == nil {
		return fmt.Errorf("file cannot be nil")
	}

	requiredCost := CalculateStorageCost(int64(len(file.Data)))

	if file.Balance < requiredCost {
		return fmt.Errorf("insufficient balance for storage: required %d, have %d", requiredCost, file.Balance)
	}

	return nil
}

// ValidateBalanceForDataSize checks if a given balance is sufficient for a specific data size
// This is used to validate operations before they are applied (Requirement 6.3, 6.4)
func ValidateBalanceForDataSize(balance int64, dataSize int64) error {
	requiredCost := CalculateStorageCost(dataSize)

	if balance < requiredCost {
		return fmt.Errorf("insufficient balance for data size %d bytes: required %d, have %d",
			dataSize, requiredCost, balance)
	}

	return nil
}

// BytesToKB converts bytes to kilobytes (rounded up)
func BytesToKB(bytes int64) int64 {
	if bytes <= 0 {
		return 0
	}
	kb := bytes / 1024
	if bytes%1024 != 0 {
		kb++ // Round up
	}
	return kb
}

// KBToMB converts kilobytes to megabytes (as float64 for precise calculation)
func KBToMB(kb int64) float64 {
	return float64(kb) / 1024.0
}

// MBToKB converts megabytes to kilobytes
func MBToKB(mb float64) int64 {
	return int64(mb * 1024.0)
}

// BytesToMB converts bytes to megabytes (as float64 for precise calculation)
func BytesToMB(bytes int64) float64 {
	return float64(bytes) / (1024.0 * 1024.0)
}

// FileStore manages all on-chain file state with efficient storage and retrieval
// It uses BadgerDB for persistent storage and maintains an in-memory cache for hot files
type FileStore struct {
	db    *badger.DB
	cache map[FileID]*File
	mu    sync.RWMutex
}

// NewFileStore initializes a new FileStore with BadgerDB at the specified path
// This implements the initialization requirement for the FileStore component
func NewFileStore(dbPath string) (*FileStore, error) {
	// Open BadgerDB with default options
	opts := badger.DefaultOptions(dbPath)
	opts.Logger = nil // Disable BadgerDB logging to avoid noise

	db, err := badger.Open(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to open BadgerDB at %s: %w", dbPath, err)
	}

	fs := &FileStore{
		db:    db,
		cache: make(map[FileID]*File),
	}

	return fs, nil
}

// Close closes the FileStore and releases all resources
func (fs *FileStore) Close() error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	// Clear cache
	fs.cache = nil

	// Close database
	if fs.db != nil {
		return fs.db.Close()
	}

	return nil
}

// CreateFile creates a new file in the store and returns its ID
// This validates storage cost before creation (Requirements 6.1, 6.2, 6.3, 6.4, 6.5)
func (fs *FileStore) CreateFile(file *File) (FileID, error) {
	if file == nil {
		return FileID{}, fmt.Errorf("file cannot be nil")
	}

	// Validate storage cost before creation (Requirement 6.1, 6.2)
	if err := ValidateStorageCost(file); err != nil {
		return FileID{}, fmt.Errorf("storage cost validation failed: %w", err)
	}

	// Set timestamps
	now := time.Now()
	file.CreatedAt = now
	file.UpdatedAt = now

	// Generate FileID if not set
	if file.ID == (FileID{}) {
		// Generate ID from file data
		idData := fmt.Sprintf("%s-%d-%s", file.TxManager.String(), now.UnixNano(), string(file.Data))
		file.ID = GenerateFileID([]byte(idData))
	}

	// Serialize file
	data, err := file.Marshal()
	if err != nil {
		return FileID{}, fmt.Errorf("failed to marshal file: %w", err)
	}

	// Store in BadgerDB
	err = fs.db.Update(func(txn *badger.Txn) error {
		return txn.Set(file.ID[:], data)
	})
	if err != nil {
		return FileID{}, fmt.Errorf("failed to store file in database: %w", err)
	}

	// Update cache
	fs.mu.Lock()
	fs.cache[file.ID] = file
	fs.mu.Unlock()

	return file.ID, nil
}

// GetFile retrieves a file by ID with caching support
// This implements Requirements 1.1, 1.2, 1.3, 1.4, 1.5
func (fs *FileStore) GetFile(id FileID) (*File, error) {
	// Check cache first
	fs.mu.RLock()
	if cached, ok := fs.cache[id]; ok {
		fs.mu.RUnlock()
		// Return a copy to prevent external modifications
		fileCopy := *cached
		return &fileCopy, nil
	}
	fs.mu.RUnlock()

	// Not in cache, fetch from database
	var fileData []byte
	err := fs.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(id[:])
		if err != nil {
			return err
		}

		fileData, err = item.ValueCopy(nil)
		return err
	})

	if err == badger.ErrKeyNotFound {
		return nil, fmt.Errorf("file not found: %s", id.String())
	}
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve file from database: %w", err)
	}

	// Unmarshal file
	file, err := UnmarshalFile(fileData)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal file: %w", err)
	}

	// Update cache
	fs.mu.Lock()
	fs.cache[id] = file
	fs.mu.Unlock()

	// Return a copy
	fileCopy := *file
	return &fileCopy, nil
}

// UpdateFile updates an existing file with validation
// This validates storage cost constraints (Requirements 6.3, 6.4)
func (fs *FileStore) UpdateFile(id FileID, file *File) error {
	if file == nil {
		return fmt.Errorf("file cannot be nil")
	}

	// Ensure the file ID matches
	if file.ID != id {
		return fmt.Errorf("file ID mismatch: expected %s, got %s", id.String(), file.ID.String())
	}

	// Validate storage cost (Requirement 6.3, 6.4)
	if err := ValidateStorageCost(file); err != nil {
		return fmt.Errorf("storage cost validation failed: %w", err)
	}

	// Update timestamp
	file.UpdatedAt = time.Now()

	// Serialize file
	data, err := file.Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal file: %w", err)
	}

	// Update in BadgerDB
	err = fs.db.Update(func(txn *badger.Txn) error {
		// Check if file exists
		_, err := txn.Get(id[:])
		if err == badger.ErrKeyNotFound {
			return fmt.Errorf("file not found: %s", id.String())
		}
		if err != nil {
			return err
		}

		// Update the file
		return txn.Set(id[:], data)
	})
	if err != nil {
		return fmt.Errorf("failed to update file in database: %w", err)
	}

	// Update cache
	fs.mu.Lock()
	fs.cache[id] = file
	fs.mu.Unlock()

	return nil
}

// DeleteFile removes a file from the store
func (fs *FileStore) DeleteFile(id FileID) error {
	// Delete from BadgerDB
	err := fs.db.Update(func(txn *badger.Txn) error {
		// Check if file exists
		_, err := txn.Get(id[:])
		if err == badger.ErrKeyNotFound {
			return fmt.Errorf("file not found: %s", id.String())
		}
		if err != nil {
			return err
		}

		// Delete the file
		return txn.Delete(id[:])
	})
	if err != nil {
		return fmt.Errorf("failed to delete file from database: %w", err)
	}

	// Remove from cache
	fs.mu.Lock()
	delete(fs.cache, id)
	fs.mu.Unlock()

	return nil
}

// GetFileBalance retrieves the balance of a specific file
func (fs *FileStore) GetFileBalance(id FileID) (int64, error) {
	file, err := fs.GetFile(id)
	if err != nil {
		return 0, err
	}
	return file.Balance, nil
}

// UpdateFileBalance updates the balance of a specific file
// This validates that the new balance still covers storage costs (Requirement 6.3)
func (fs *FileStore) UpdateFileBalance(id FileID, delta int64) error {
	// Get the current file
	file, err := fs.GetFile(id)
	if err != nil {
		return err
	}

	// Calculate new balance
	newBalance := file.Balance + delta

	// Validate that new balance covers storage cost (Requirement 6.3)
	if err := ValidateBalanceForDataSize(newBalance, int64(len(file.Data))); err != nil {
		return fmt.Errorf("balance update would violate storage cost constraint: %w", err)
	}

	// Update balance
	file.Balance = newBalance

	// Update the file
	return fs.UpdateFile(id, file)
}

// GetAllFileIDs returns all file IDs in the store (sorted)
func (fs *FileStore) GetAllFileIDs() ([]FileID, error) {
	var fileIDs []FileID

	err := fs.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false // We only need keys
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			key := item.Key()

			// Convert key to FileID
			var fid FileID
			if len(key) == 32 {
				copy(fid[:], key)
				fileIDs = append(fileIDs, fid)
			}
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to retrieve file IDs: %w", err)
	}

	return fileIDs, nil
}

// CalculateStateRoot computes the state root hash from all files in the store
// State Root = MerkleRoot(sorted_file_ids, file_hashes)
// Where file_hash = SHA-256(file.ID || file.Balance || file.TxManager ||
//
//	file.Data || file.Executable || file.UpdatedAt)
func (fs *FileStore) CalculateStateRoot() ([]byte, error) {
	// Get all file IDs (already sorted by BadgerDB iteration order)
	fileIDs, err := fs.GetAllFileIDs()
	if err != nil {
		return nil, fmt.Errorf("failed to get file IDs: %w", err)
	}

	// If no files exist, return empty hash
	if len(fileIDs) == 0 {
		emptyHash := sha256.Sum256([]byte{})
		return emptyHash[:], nil
	}

	// Calculate hash for each file
	fileHashes := make([][]byte, len(fileIDs))
	for i, fid := range fileIDs {
		file, err := fs.GetFile(fid)
		if err != nil {
			return nil, fmt.Errorf("failed to get file %s: %w", fid.String(), err)
		}

		// Calculate file hash
		fileHash := calculateFileHash(file)
		fileHashes[i] = fileHash
	}

	// Build Merkle tree from file hashes
	stateRoot := buildMerkleTree(fileHashes)
	return stateRoot, nil
}

// calculateFileHash computes the hash of a file's state
// file_hash = SHA-256(file.ID || file.Balance || file.TxManager ||
//
//	file.Data || file.Executable || file.UpdatedAt)
func calculateFileHash(file *File) []byte {
	hasher := sha256.New()

	// Write file ID
	hasher.Write(file.ID[:])

	// Write balance (8 bytes, big-endian)
	balanceBytes := make([]byte, 8)
	for i := 0; i < 8; i++ {
		balanceBytes[7-i] = byte(file.Balance >> (i * 8))
	}
	hasher.Write(balanceBytes)

	// Write tx manager
	hasher.Write(file.TxManager[:])

	// Write data
	hasher.Write(file.Data)

	// Write executable flag
	if file.Executable {
		hasher.Write([]byte{1})
	} else {
		hasher.Write([]byte{0})
	}

	// Write updated timestamp (8 bytes, Unix timestamp)
	timestampBytes := make([]byte, 8)
	timestamp := file.UpdatedAt.Unix()
	for i := 0; i < 8; i++ {
		timestampBytes[7-i] = byte(timestamp >> (i * 8))
	}
	hasher.Write(timestampBytes)

	return hasher.Sum(nil)
}

// buildMerkleTree builds a Merkle tree from a list of hashes
func buildMerkleTree(hashes [][]byte) []byte {
	if len(hashes) == 0 {
		emptyHash := sha256.Sum256([]byte{})
		return emptyHash[:]
	}

	if len(hashes) == 1 {
		return hashes[0]
	}

	// Build tree by repeatedly hashing pairs
	currentLevel := hashes
	for len(currentLevel) > 1 {
		var nextLevel [][]byte

		for i := 0; i < len(currentLevel); i += 2 {
			if i+1 < len(currentLevel) {
				// Hash pair
				combined := append(currentLevel[i], currentLevel[i+1]...)
				hash := sha256.Sum256(combined)
				nextLevel = append(nextLevel, hash[:])
			} else {
				// Odd one out, hash with itself
				combined := append(currentLevel[i], currentLevel[i]...)
				hash := sha256.Sum256(combined)
				nextLevel = append(nextLevel, hash[:])
			}
		}

		currentLevel = nextLevel
	}

	return currentLevel[0]
}
