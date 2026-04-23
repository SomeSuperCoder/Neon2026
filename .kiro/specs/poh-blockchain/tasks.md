# Implementation Plan

- [x] 1. Initialize Go project structure and dependencies
  - [x] Create go.mod file with module name and Go version
  - [x] Add required dependencies (mattn/go-sqlite3 for SQLite)
  - [x] Create directory structure: cmd/, internal/poh/, internal/blockchain/, internal/network/, internal/storage/, internal/consensus/
  - [x] Create README.md with project documentation
  - _Requirements: All requirements depend on proper project setup_

- [x] 2. Implement the PoH Clock Generator ✓ COMPLETE
  - [x] 2.1 Create PohClock struct and initialization
    - Define PohClock struct with currentHash, tickCount, hashCount fields and sync.RWMutex
    - Implement NewPohClock(seed []byte) constructor that initializes the clock with a seed hash
    - _Requirements: 1.1, 1.4_
  
  - [x] 2.2 Implement core hashing functionality
    - Implement HashOnce() method that performs single SHA-256 hash on currentHash
    - Implement GetCurrentHash() method with read lock for thread-safe access
    - Implement GetTickCount() method with read lock
    - _Requirements: 1.1, 1.5_
  
  - [x] 2.3 Implement tick generation
    - Implement Tick() method that performs 12,500 hash operations and returns a Tick struct
    - Create Tick struct with HashValue, Timestamp, and TickNumber fields
    - Ensure tick generation updates tickCount and hashCount atomically
    - _Requirements: 1.2, 1.3, 1.4_

- [x] 3. Define core blockchain data structures ✓ COMPLETE
  - [x] 3.1 Create data structure definitions
    - Define Entry struct with Hash, NumHashes, Transactions, PreviousHash, Timestamp fields
    - Define Transaction struct with Sender, Receiver, Amount, Signature, Data fields
    - Define Block struct with Header and Entries fields
    - Define BlockHeader struct with PreviousBlockHash, MerkleRoot, Slot, Timestamp, BlockHeight fields
    - _Requirements: 2.1, 2.2, 2.3, 2.4, 2.5_
  
  - [x] 3.2 Implement JSON serialization methods
    - Add JSON tags to all struct fields for proper serialization
    - Implement helper functions for encoding/decoding []byte fields to/from hex strings in JSON
    - _Requirements: 2.1, 2.2, 2.3, 2.4_

- [x] 4. Implement Block Producer for transaction integration ✓ COMPLETE
  - [x] 4.1 Create BlockProducer struct and initialization
    - Define BlockProducer struct with pohClock, pendingTransactions, currentSlot, entriesBuffer, and sync.Mutex
    - Implement NewBlockProducer(pohClock *PohClock) constructor
    - _Requirements: 3.1, 3.4_
  
  - [x] 4.2 Implement transaction handling
    - Implement AddTransaction(tx Transaction) method with mutex protection
    - Implement MixTransactionHash(txHash []byte) that mixes transaction hash into PoH chain
    - _Requirements: 3.1, 3.2_
  
  - [x] 4.3 Implement entry production
    - Implement ProduceEntry() method that creates Entry with pending transactions
    - Ensure entry records num_hashes since previous entry
    - Ensure entry maintains hash link to previous entry
    - Clear pending transactions after entry creation
    - _Requirements: 3.2, 3.3, 3.5_
  
  - [x] 4.4 Implement block production
    - Implement ProduceBlock(slot int64) method that produces complete block
    - Ensure block contains at least 64 ticks worth of entries
    - Calculate Merkle root from entries
    - Create BlockHeader with previous block hash, merkle root, slot, timestamp, and height
    - _Requirements: 5.2, 5.5_

- [x] 5. Implement network layer for node communication ✓ COMPLETE
  - [x] 5.1 Create NetworkNode struct and initialization
    - Define NetworkNode struct with connections, nodeType, messageQueue, listener, and sync.RWMutex
    - Define NodeType enum (LEADER, REPLICA)
    - Implement NewNetworkNode(host string, port int, nodeType NodeType) constructor
    - _Requirements: 4.3_
  
  - [x] 5.2 Implement connection management
    - Implement Start() method that starts TCP listener for incoming connections
    - Implement ConnectToPeer(peerAddress string) that establishes connection to another node
    - Implement Stop() method for graceful shutdown
    - Use goroutines to handle each connection concurrently
    - _Requirements: 4.3_
  
  - [x] 5.3 Implement block serialization
    - Implement SerializeBlock(block Block) that converts block to JSON bytes
    - Implement DeserializeBlock(data []byte) that reconstructs block from JSON bytes
    - Handle byte slice encoding/decoding for hash fields
    - _Requirements: 4.4, 4.5_
  
  - [x] 5.4 Implement block broadcasting and receiving
    - Implement BroadcastBlock(block Block) that sends block to all connected peers
    - Implement ReceiveBlock() that listens on messageQueue and returns received blocks
    - Add message framing (length prefix) for TCP stream handling
    - _Requirements: 4.1, 4.2_

- [x] 6. Implement consensus manager
  - [x] 6.1 Create ConsensusManager struct
    - Define ConsensusManager struct with nodeType, slotDurationMs, genesisTimestamp
    - Implement NewConsensusManager(nodeType NodeType) constructor with 400ms slot duration
    - _Requirements: 5.1_
  
  - [x] 6.2 Implement slot timing logic
    - Implement GetCurrentSlot() that calculates current slot from time elapsed since genesis
    - Implement WaitForSlotStart(slot int64) that blocks until specified slot begins
    - _Requirements: 5.2, 5.4_
  
  - [x] 6.3 Implement leader determination and block validation
    - Implement IsLeader(slot int64) that returns true if this node is leader (simplified: always true for LEADER type)
    - Implement ValidateBlock(block Block) that performs basic block validation
    - Verify block slot matches expected slot
    - Verify block contains at least 64 ticks worth of hash operations
    - _Requirements: 5.1, 5.3, 5.5_

- [x] 7. Implement ledger storage with SQLite ✓ COMPLETE
  - [x] 7.1 Create Ledger struct and database schema
    - Define Ledger struct with db *sql.DB and sync.RWMutex
    - Implement NewLedger(dbPath string) that opens SQLite connection
    - Create blocks table schema (block_height, block_hash, previous_hash, merkle_root, slot, timestamp, data)
    - Create entries table schema (entry_id, block_height, hash, num_hashes, previous_hash, timestamp, transactions)
    - Create transactions table schema (tx_id, entry_id, sender, receiver, amount, signature, data)
    - _Requirements: 6.1, 6.2_
  
  - [x] 7.2 Implement block storage operations
    - Implement StoreBlock(block Block) that persists block and all entries/transactions
    - Use SQL transactions to ensure atomicity
    - Serialize block data to JSON for storage
    - _Requirements: 6.1_
  
  - [x] 7.3 Implement block retrieval operations
    - Implement GetBlockByHeight(height int64) that retrieves block by height
    - Implement GetBlockByHash(hash []byte) that retrieves block by hash
    - Implement GetLatestBlock() that returns most recent block
    - Implement GetChainHeight() that returns current blockchain height
    - Deserialize JSON data back to Block structs
    - _Requirements: 6.3, 6.4_
  
  - [x] 7.4 Implement ledger initialization and recovery
    - Implement Close() method for graceful database shutdown
    - Add logic to load existing blockchain state on startup
    - _Requirements: 6.5_

- [x] 8. Implement verification engine ✓ COMPLETE
  - [x] 8.1 Create Verifier struct and entry verification
    - Define Verifier struct
    - Implement NewVerifier() constructor
    - Implement VerifyEntry(entry Entry) that verifies entry hash chain
    - Verify entry hash correctly links to previous hash
    - _Requirements: 7.2_
  
  - [x] 8.2 Implement hash count verification
    - Implement VerifyHashCount(entry Entry) that verifies num_hashes accuracy
    - Recreate hash chain from previous hash for num_hashes iterations
    - Compare final hash with entry hash
    - _Requirements: 7.4_
  
  - [x] 8.3 Implement block verification
    - Implement VerifyBlock(block Block) that verifies single block validity
    - Verify all entries in block using VerifyEntry
    - Verify block header merkle root matches calculated merkle root from entries
    - _Requirements: 7.3, 7.5_
  
  - [x] 8.4 Implement block linkage verification
    - Implement VerifyBlockLink(block Block, previousBlock Block) that verifies block linkage
    - Verify block header previous hash matches previous block hash
    - Verify block starts with hash from previous block's final entry
    - _Requirements: 7.3_
  
  - [x] 8.5 Implement full chain verification
    - Implement VerifyChain(ledger *Ledger) that verifies entire blockchain
    - Iterate through all blocks from genesis to current height
    - Verify each block and its linkage to previous block
    - Return error with details if any verification fails
    - _Requirements: 7.1, 7.5_

- [x] 9. Create main application and CLI ✓ COMPLETE
  - [x] 9.1 Implement node initialization
    - Create main.go in cmd/ directory
    - Parse command-line flags for node type (leader/replica), port, peer addresses, database path
    - Initialize all components: PohClock, BlockProducer, NetworkNode, ConsensusManager, Ledger, Verifier
    - _Requirements: All requirements_
  
  - [x] 9.2 Implement leader node logic
    - Create goroutine for continuous block production when node is leader
    - Wait for slot start, produce block with BlockProducer, broadcast via NetworkNode
    - Store produced blocks to Ledger
    - _Requirements: 5.2, 5.3_
  
  - [x] 9.3 Implement replica node logic
    - Create goroutine for receiving blocks from network
    - Validate received blocks with ConsensusManager and Verifier
    - Store valid blocks to Ledger
    - _Requirements: 5.3_
  
  - [x] 9.4 Add graceful shutdown handling
    - Handle OS signals (SIGINT, SIGTERM)
    - Stop network node, close ledger database
    - Log shutdown completion
    - _Requirements: All requirements_

- [x] 10. Write integration tests ✓ COMPLETE
  - [x] Create test that initializes full node, produces blocks, and verifies chain
  - [x] Create test for leader-replica communication and block propagation
  - [x] Create test for ledger persistence and recovery after restart
  - _Requirements: All requirements_
