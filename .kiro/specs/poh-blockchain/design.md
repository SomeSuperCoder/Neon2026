# Design Document: Proof of History Blockchain

## Overview

This design implements a simplified Proof of History (PoH) blockchain inspired by Solana's architecture. The system uses sequential SHA-256 hashing as a verifiable delay function to create a cryptographic clock that timestamps transactions without requiring traditional consensus mechanisms for ordering.

The implementation will be built in phases, starting with the core PoH clock generator, then adding data structures, transaction integration, networking, basic consensus, persistence, and verification logic.

## Architecture

The system follows a layered architecture:

```
┌─────────────────────────────────────────┐
│         Application Layer               │
│  (CLI, API, Node Management)            │
└─────────────────────────────────────────┘
                  │
┌─────────────────────────────────────────┐
│         Consensus Layer                 │
│  (Leader Selection, Block Production)   │
└─────────────────────────────────────────┘
                  │
┌─────────────────────────────────────────┐
│         Network Layer                   │
│  (P2P Communication, Serialization)     │
└─────────────────────────────────────────┘
                  │
┌─────────────────────────────────────────┐
│         Core Blockchain Layer           │
│  (PoH Clock, Entries, Blocks)           │
└─────────────────────────────────────────┘
                  │
┌─────────────────────────────────────────┐
│         Storage Layer                   │
│  (Ledger, Persistent Storage)           │
└─────────────────────────────────────────┘
```

### Technology Stack

- **Language**: Go (Golang)
- **Hashing**: crypto/sha256 from standard library
- **Networking**: net package for TCP-based P2P communication
- **Storage**: SQLite with mattn/go-sqlite3 driver for persistent ledger storage
- **Serialization**: encoding/json for data interchange
- **Concurrency**: Goroutines and channels for concurrent operations

## Components and Interfaces

### 1. PoH Clock Generator

**Purpose**: Generate the continuous hash chain that serves as the cryptographic clock.

**Struct**: `PohClock`

**Key Methods**:
- `NewPohClock(seed []byte) *PohClock`: Initialize with a seed hash
- `Tick() Tick`: Perform 12,500 hash operations and return a Tick
- `GetCurrentHash() []byte`: Return the current hash state
- `GetTickCount() int64`: Return the total number of ticks generated
- `HashOnce() []byte`: Perform a single SHA-256 hash operation

**Fields**:
- `currentHash []byte`: The latest hash in the chain
- `tickCount int64`: Total ticks generated
- `hashCount int64`: Total hash operations performed
- `mu sync.RWMutex`: Mutex for thread-safe operations

### 2. Data Structures

**Tick**:
```go
type Tick struct {
    HashValue   []byte
    Timestamp   time.Time
    TickNumber  int64
}
```

**Entry**:
```go
type Entry struct {
    Hash          []byte
    NumHashes     int64
    Transactions  []Transaction
    PreviousHash  []byte
    Timestamp     time.Time
}
```

**Transaction**:
```go
type Transaction struct {
    Sender    string
    Receiver  string
    Amount    float64
    Signature []byte
    Data      map[string]interface{}
}
```

**Block**:
```go
type Block struct {
    Header  BlockHeader
    Entries []Entry
}

type BlockHeader struct {
    PreviousBlockHash []byte
    MerkleRoot        []byte
    Slot              int64
    Timestamp         time.Time
    BlockHeight       int64
}
```

### 3. Block Producer

**Purpose**: Integrate transactions into the PoH clock and produce blocks.

**Struct**: `BlockProducer`

**Key Methods**:
- `NewBlockProducer(pohClock *PohClock) *BlockProducer`: Initialize with a PoH clock instance
- `AddTransaction(tx Transaction)`: Queue a transaction for inclusion
- `ProduceEntry() (Entry, error)`: Create an entry with pending transactions
- `ProduceBlock(slot int64) (Block, error)`: Produce a complete block for a slot
- `MixTransactionHash(txHash []byte)`: Mix transaction hash into PoH chain

**Fields**:
- `pohClock *PohClock`: Reference to the PoH clock
- `pendingTransactions []Transaction`: Queue of transactions awaiting inclusion
- `currentSlot int64`: The current slot number
- `entriesBuffer []Entry`: Entries accumulated for the current block
- `mu sync.Mutex`: Mutex for thread-safe operations

### 4. Network Layer

**Purpose**: Enable communication between nodes for block distribution.

**Struct**: `NetworkNode`

**Key Methods**:
- `NewNetworkNode(host string, port int, nodeType NodeType) *NetworkNode`: Initialize network node
- `ConnectToPeer(peerAddress string) error`: Establish connection to another node
- `BroadcastBlock(block Block) error`: Send block to all connected peers
- `ReceiveBlock() (Block, error)`: Listen for and receive blocks from peers
- `SerializeBlock(block Block) ([]byte, error)`: Convert block to bytes for transmission
- `DeserializeBlock(data []byte) (Block, error)`: Reconstruct block from bytes
- `Start() error`: Start listening for incoming connections
- `Stop()`: Gracefully shutdown network node

**Fields**:
- `connections []*net.Conn`: List of active peer connections
- `nodeType NodeType`: Either LEADER or REPLICA
- `messageQueue chan []byte`: Channel for incoming messages
- `listener net.Listener`: TCP listener for incoming connections
- `mu sync.RWMutex`: Mutex for thread-safe operations

### 5. Consensus Manager

**Purpose**: Implement basic leader-based consensus protocol.

**Struct**: `ConsensusManager`

**Key Methods**:
- `NewConsensusManager(nodeType NodeType) *ConsensusManager`: Initialize consensus manager
- `IsLeader(slot int64) bool`: Determine if this node is leader for the slot
- `ValidateBlock(block Block) error`: Validate a received block
- `GetCurrentSlot() int64`: Calculate current slot based on time
- `WaitForSlotStart(slot int64)`: Block until slot begins

**Fields**:
- `nodeType NodeType`: LEADER or REPLICA
- `slotDurationMs int64`: Duration of each slot (400ms)
- `genesisTimestamp time.Time`: Starting time for slot calculations

### 6. Ledger Storage

**Purpose**: Persist blockchain data and provide query interface.

**Struct**: `Ledger`

**Key Methods**:
- `NewLedger(dbPath string) (*Ledger, error)`: Initialize database connection
- `StoreBlock(block Block) error`: Persist a block to storage
- `StoreEntry(entry Entry) error`: Persist an entry to storage
- `GetBlockByHeight(height int64) (Block, error)`: Retrieve block by height
- `GetBlockByHash(hash []byte) (Block, error)`: Retrieve block by hash
- `GetLatestBlock() (Block, error)`: Get the most recent block
- `GetChainHeight() (int64, error)`: Return current blockchain height
- `Close() error`: Close database connection

**Fields**:
- `db *sql.DB`: SQLite database connection
- `mu sync.RWMutex`: Mutex for thread-safe database operations

**Storage Schema**:
- `blocks` table: block_height, block_hash, previous_hash, merkle_root, slot, timestamp, data
- `entries` table: entry_id, block_height, hash, num_hashes, previous_hash, timestamp, transactions
- `transactions` table: tx_id, entry_id, sender, receiver, amount, signature, data

### 7. Verification Engine

**Purpose**: Verify blockchain integrity and validity.

**Struct**: `Verifier`

**Key Methods**:
- `NewVerifier() *Verifier`: Create a new verifier instance
- `VerifyChain(ledger *Ledger) error`: Verify entire blockchain from genesis
- `VerifyBlock(block Block) error`: Verify a single block's validity
- `VerifyEntry(entry Entry) error`: Verify an entry's hash chain
- `VerifyHashCount(entry Entry) error`: Verify num_hashes is accurate
- `VerifyBlockLink(block Block, previousBlock Block) error`: Verify block linkage

**Verification Steps**:
1. Verify hash chain continuity (each hash correctly derives from previous)
2. Verify entry linkage within blocks
3. Verify block header linkage between blocks
4. Verify num_hashes counts match actual operations
5. Verify Merkle roots match entry data

## Data Models

### Hash Chain Flow

```
Genesis Hash
    │
    ├─> Hash 1 (SHA256)
    ├─> Hash 2 (SHA256)
    ├─> ...
    ├─> Hash 12,500 (SHA256) = Tick 1
    │
    ├─> Hash 12,501
    ├─> ...
    ├─> Hash 25,000 = Tick 2
    │
    └─> Continue...
```

### Entry Creation with Transactions

```
Previous Hash: 0xABC...
    │
    ├─> Hash 100 times (num_hashes = 100)
    ├─> Mix in Transaction Hash
    ├─> Hash 50 more times (num_hashes = 150 total)
    │
    └─> Create Entry {
          hash: current_hash,
          num_hashes: 150,
          transactions: [tx],
          previous_hash: 0xABC...
        }
```

### Block Structure

```
Block {
  header: {
    previous_block_hash: 0x123...,
    merkle_root: 0x456...,
    slot: 42,
    timestamp: 1234567890.123,
    block_height: 100
  },
  entries: [
    Entry { hash: 0xAAA..., num_hashes: 150, ... },
    Entry { hash: 0xBBB..., num_hashes: 200, ... },
    Entry { hash: 0xCCC..., num_hashes: 175, ... },
    ...
  ]
}
```

## Error Handling

### PoH Clock Errors
- **Hash Computation Failure**: Log error and retry with exponential backoff
- **State Corruption**: Detect via checksum, restore from last known good state

### Network Errors
- **Connection Failure**: Retry connection with exponential backoff, max 5 attempts
- **Deserialization Error**: Log malformed message, discard and continue
- **Timeout**: Close stale connections after 30 seconds of inactivity

### Consensus Errors
- **Invalid Block Received**: Reject block, log validation failure details
- **Slot Timing Mismatch**: Discard blocks from wrong slots
- **Missing Previous Block**: Request missing blocks from peers

### Storage Errors
- **Database Write Failure**: Retry transaction up to 3 times, then abort
- **Corruption Detected**: Attempt recovery from backup or resync from peers
- **Disk Full**: Alert operator, pause block production

### Verification Errors
- **Hash Chain Break**: Identify break point, reject all subsequent blocks
- **Invalid num_hashes**: Reject entry and containing block
- **Merkle Root Mismatch**: Reject block as corrupted

## Testing Strategy

### Unit Tests

1. **PoH Clock Tests**
   - Test single hash operation produces correct SHA-256 output
   - Test tick generation after exactly 12,500 hashes
   - Test hash chain continuity over 1000 ticks
   - Test state persistence and restoration

2. **Data Structure Tests**
   - Test Tick, Entry, Block serialization and deserialization
   - Test Entry hash linkage
   - Test Block header validation

3. **Block Producer Tests**
   - Test transaction queuing and inclusion
   - Test entry creation with transactions
   - Test block production with minimum 64 ticks
   - Test transaction hash mixing into PoH chain

4. **Verification Tests**
   - Test detection of modified hashes
   - Test detection of incorrect num_hashes
   - Test detection of broken entry links
   - Test detection of broken block links

### Integration Tests

1. **End-to-End Block Production**
   - Start PoH clock, add transactions, produce block, verify block

2. **Network Communication**
   - Start leader and replica nodes, produce block on leader, receive on replica

3. **Ledger Persistence**
   - Produce blocks, store to ledger, restart system, verify chain loads correctly

4. **Chain Verification**
   - Build chain of 100 blocks, verify entire chain, introduce corruption, detect failure

### Performance Tests

1. **Hash Rate Measurement**
   - Measure hashes per second on target hardware
   - Verify tick interval is consistent

2. **Block Production Throughput**
   - Measure blocks per second
   - Measure transactions per second

3. **Network Latency**
   - Measure block propagation time between nodes

## Implementation Phases

The implementation will follow the 7 phases outlined in the requirements:

1. **Phase 1**: PoH Clock - Core hash chain generator
2. **Phase 2**: Data Structures - Tick, Entry, Block definitions
3. **Phase 3**: Transaction Integration - Mix transactions into PoH
4. **Phase 4**: Network Layer - P2P communication
5. **Phase 5**: Consensus Protocol - Leader-based block production
6. **Phase 6**: Ledger Storage - Persistent storage with SQLite
7. **Phase 7**: Verification Logic - Chain integrity verification

Each phase builds incrementally on the previous phase, ensuring a working system at each step.
