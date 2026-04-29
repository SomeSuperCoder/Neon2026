package internal

import (
	"os"
	"testing"
	"time"

	"github.com/poh-blockchain/internal/blockchain"
	"github.com/poh-blockchain/internal/consensus"
	"github.com/poh-blockchain/internal/filestore"
	"github.com/poh-blockchain/internal/network"
	"github.com/poh-blockchain/internal/poh"
	"github.com/poh-blockchain/internal/storage"
	"github.com/poh-blockchain/internal/verification"
)

// TestFullNodeBlockProductionAndVerification tests initializing a full node,
// producing blocks, and verifying the chain
func TestFullNodeBlockProductionAndVerification(t *testing.T) {
	// Create temporary database
	dbPath := "test_full_node.db"
	defer os.Remove(dbPath)

	// Initialize PoH Clock
	pohClock := poh.NewPohClock([]byte("test-genesis-seed"))

	// Initialize Block Producer
	blockProducer := blockchain.NewBlockProducer(pohClock)

	// Initialize Ledger
	ledger, err := storage.NewLedger(dbPath)
	if err != nil {
		t.Fatalf("Failed to initialize ledger: %v", err)
	}
	defer ledger.Close()

	// Initialize Verifier
	verifier := verification.NewVerifier()

	// Initialize Consensus Manager
	// Use observer mode (zero validator ID) for this test
	consensusManager := consensus.NewConsensusManager(filestore.FileID{}, nil)

	// Produce first block (genesis)
	currentSlot := consensusManager.GetCurrentSlot()
	block1, err := blockProducer.ProduceBlock(currentSlot, []byte{})
	if err != nil {
		t.Fatalf("Failed to produce block 1: %v", err)
	}
	block1.Header.BlockHeight = 1

	// Store block 1
	if err := ledger.StoreBlock(block1); err != nil {
		t.Fatalf("Failed to store block 1: %v", err)
	}

	// Verify block 1
	if err := verifier.VerifyBlock(block1); err != nil {
		t.Fatalf("Block 1 verification failed: %v", err)
	}

	// Wait for next slot
	time.Sleep(450 * time.Millisecond)

	// Produce second block
	currentSlot = consensusManager.GetCurrentSlot()
	block2, err := blockProducer.ProduceBlock(currentSlot, []byte{})
	if err != nil {
		t.Fatalf("Failed to produce block 2: %v", err)
	}
	block2.Header.BlockHeight = 2
	block2.Header.PreviousBlockHash = block1.Header.MerkleRoot

	// Store block 2
	if err := ledger.StoreBlock(block2); err != nil {
		t.Fatalf("Failed to store block 2: %v", err)
	}

	// Verify block 2
	if err := verifier.VerifyBlock(block2); err != nil {
		t.Fatalf("Block 2 verification failed: %v", err)
	}

	// Verify block linkage
	if err := verifier.VerifyBlockLink(block2, block1); err != nil {
		t.Fatalf("Block 2 linkage verification failed: %v", err)
	}

	// Verify entire chain
	if err := verifier.VerifyChain(ledger); err != nil {
		t.Fatalf("Chain verification failed: %v", err)
	}

	// Verify chain height
	height, err := ledger.GetChainHeight()
	if err != nil {
		t.Fatalf("Failed to get chain height: %v", err)
	}
	if height != 2 {
		t.Errorf("Expected chain height 2, got %d", height)
	}

	t.Log("Full node block production and verification test passed")
}

// TestLeaderReplicaCommunication tests leader-replica communication and block propagation
func TestLeaderReplicaCommunication(t *testing.T) {
	// Create temporary databases
	leaderDBPath := "test_leader.db"
	replicaDBPath := "test_replica.db"
	defer os.Remove(leaderDBPath)
	defer os.Remove(replicaDBPath)

	// Initialize leader components
	leaderPohClock := poh.NewPohClock([]byte("leader-genesis-seed"))
	leaderBlockProducer := blockchain.NewBlockProducer(leaderPohClock)
	leaderLedger, err := storage.NewLedger(leaderDBPath)
	if err != nil {
		t.Fatalf("Failed to initialize leader ledger: %v", err)
	}
	defer leaderLedger.Close()

	leaderConsensus := consensus.NewConsensusManager(filestore.FileID{}, nil)
	leaderNetwork := network.NewNetworkNode("127.0.0.1", 9001)
	if err := leaderNetwork.Start(); err != nil {
		t.Fatalf("Failed to start leader network: %v", err)
	}
	defer leaderNetwork.Stop()

	// Initialize replica components
	replicaLedger, err := storage.NewLedger(replicaDBPath)
	if err != nil {
		t.Fatalf("Failed to initialize replica ledger: %v", err)
	}
	defer replicaLedger.Close()

	replicaConsensus := consensus.NewConsensusManager(filestore.FileID{}, nil)
	replicaVerifier := verification.NewVerifier()
	replicaNetwork := network.NewNetworkNode("127.0.0.1", 9002)
	if err := replicaNetwork.Start(); err != nil {
		t.Fatalf("Failed to start replica network: %v", err)
	}
	defer replicaNetwork.Stop()

	// Connect replica to leader
	time.Sleep(100 * time.Millisecond) // Give leader time to start listening
	if err := replicaNetwork.ConnectToPeer("127.0.0.1:9001"); err != nil {
		t.Fatalf("Failed to connect replica to leader: %v", err)
	}

	// Give connection time to establish
	time.Sleep(100 * time.Millisecond)

	// Leader produces a block
	currentSlot := leaderConsensus.GetCurrentSlot()
	block, err := leaderBlockProducer.ProduceBlock(currentSlot, []byte{})
	if err != nil {
		t.Fatalf("Failed to produce block: %v", err)
	}
	block.Header.BlockHeight = 1

	// Leader stores the block
	if err := leaderLedger.StoreBlock(block); err != nil {
		t.Fatalf("Failed to store block on leader: %v", err)
	}

	// Leader broadcasts the block
	if err := leaderNetwork.BroadcastBlock(block); err != nil {
		t.Fatalf("Failed to broadcast block: %v", err)
	}

	// Replica receives the block
	receivedBlock, err := replicaNetwork.ReceiveBlock()
	if err != nil {
		t.Fatalf("Failed to receive block on replica: %v", err)
	}

	// Validate received block
	if err := replicaConsensus.ValidateBlock(receivedBlock); err != nil {
		t.Fatalf("Block validation failed on replica: %v", err)
	}

	// Verify received block
	if err := replicaVerifier.VerifyBlock(receivedBlock); err != nil {
		t.Fatalf("Block verification failed on replica: %v", err)
	}

	// Store received block on replica
	if err := replicaLedger.StoreBlock(receivedBlock); err != nil {
		t.Fatalf("Failed to store block on replica: %v", err)
	}

	// Verify both nodes have the same block
	leaderStoredBlock, err := leaderLedger.GetBlockByHeight(1)
	if err != nil {
		t.Fatalf("Failed to get block from leader ledger: %v", err)
	}

	replicaStoredBlock, err := replicaLedger.GetBlockByHeight(1)
	if err != nil {
		t.Fatalf("Failed to get block from replica ledger: %v", err)
	}

	if leaderStoredBlock.Header.Slot != replicaStoredBlock.Header.Slot {
		t.Errorf("Block slots don't match: leader=%d, replica=%d",
			leaderStoredBlock.Header.Slot, replicaStoredBlock.Header.Slot)
	}

	if len(leaderStoredBlock.Entries) != len(replicaStoredBlock.Entries) {
		t.Errorf("Block entry counts don't match: leader=%d, replica=%d",
			len(leaderStoredBlock.Entries), len(replicaStoredBlock.Entries))
	}

	t.Log("Leader-replica communication and block propagation test passed")
}

// TestLedgerPersistenceAndRecovery tests ledger persistence and recovery after restart
func TestLedgerPersistenceAndRecovery(t *testing.T) {
	dbPath := "test_persistence.db"
	defer os.Remove(dbPath)

	// Phase 1: Create blockchain and store blocks
	{
		pohClock := poh.NewPohClock([]byte("persistence-test-seed"))
		blockProducer := blockchain.NewBlockProducer(pohClock)
		ledger, err := storage.NewLedger(dbPath)
		if err != nil {
			t.Fatalf("Failed to initialize ledger: %v", err)
		}

		consensusManager := consensus.NewConsensusManager(filestore.FileID{}, nil)

		// Produce and store 3 blocks
		var previousBlock blockchain.Block
		for i := 1; i <= 3; i++ {
			currentSlot := consensusManager.GetCurrentSlot()
			block, err := blockProducer.ProduceBlock(currentSlot, []byte{})
			if err != nil {
				t.Fatalf("Failed to produce block %d: %v", i, err)
			}
			block.Header.BlockHeight = int64(i)

			if i > 1 {
				block.Header.PreviousBlockHash = previousBlock.Header.MerkleRoot
			}

			if err := ledger.StoreBlock(block); err != nil {
				t.Fatalf("Failed to store block %d: %v", i, err)
			}

			previousBlock = block

			// Wait between blocks
			time.Sleep(450 * time.Millisecond)
		}

		// Close the ledger
		if err := ledger.Close(); err != nil {
			t.Fatalf("Failed to close ledger: %v", err)
		}

		t.Log("Phase 1: Created and stored 3 blocks")
	}

	// Phase 2: Restart and recover blockchain state
	{
		ledger, err := storage.NewLedger(dbPath)
		if err != nil {
			t.Fatalf("Failed to reopen ledger: %v", err)
		}
		defer ledger.Close()

		// Load blockchain state
		latestBlock, height, err := ledger.LoadBlockchainState()
		if err != nil {
			t.Fatalf("Failed to load blockchain state: %v", err)
		}

		// Verify chain height
		if height != 3 {
			t.Errorf("Expected chain height 3, got %d", height)
		}

		// Verify latest block
		if latestBlock.Header.BlockHeight != 3 {
			t.Errorf("Expected latest block height 3, got %d", latestBlock.Header.BlockHeight)
		}

		// Verify all blocks can be retrieved
		for i := int64(1); i <= 3; i++ {
			block, err := ledger.GetBlockByHeight(i)
			if err != nil {
				t.Fatalf("Failed to retrieve block %d: %v", i, err)
			}

			if block.Header.BlockHeight != i {
				t.Errorf("Block %d has incorrect height: %d", i, block.Header.BlockHeight)
			}

			if len(block.Entries) == 0 {
				t.Errorf("Block %d has no entries", i)
			}
		}

		// Verify chain integrity
		verifier := verification.NewVerifier()
		if err := verifier.VerifyChain(ledger); err != nil {
			t.Fatalf("Chain verification failed after recovery: %v", err)
		}

		t.Log("Phase 2: Successfully recovered and verified blockchain state")
	}

	t.Log("Ledger persistence and recovery test passed")
}
