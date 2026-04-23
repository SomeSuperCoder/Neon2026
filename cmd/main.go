package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/poh-blockchain/internal/blockchain"
	"github.com/poh-blockchain/internal/consensus"
	"github.com/poh-blockchain/internal/network"
	"github.com/poh-blockchain/internal/poh"
	"github.com/poh-blockchain/internal/storage"
	"github.com/poh-blockchain/internal/verification"
)

func main() {
	// Parse command-line flags
	nodeTypeStr := flag.String("type", "replica", "Node type: leader or replica")
	port := flag.Int("port", 8080, "Port to listen on")
	peersStr := flag.String("peers", "", "Comma-separated list of peer addresses (host:port)")
	dbPath := flag.String("db", "blockchain.db", "Path to the database file")
	flag.Parse()

	// Determine node type
	var nodeType network.NodeType
	switch strings.ToLower(*nodeTypeStr) {
	case "leader":
		nodeType = network.LEADER
		log.Println("Starting node as LEADER")
	case "replica":
		nodeType = network.REPLICA
		log.Println("Starting node as REPLICA")
	default:
		log.Fatalf("Invalid node type: %s (must be 'leader' or 'replica')", *nodeTypeStr)
	}

	// Initialize PoH Clock with genesis seed
	log.Println("Initializing PoH Clock...")
	pohClock := poh.NewPohClock([]byte("genesis-seed"))

	// Initialize Block Producer
	log.Println("Initializing Block Producer...")
	blockProducer := blockchain.NewBlockProducer(pohClock)

	// Initialize Network Node
	log.Printf("Initializing Network Node on port %d...\n", *port)
	networkNode := network.NewNetworkNode("0.0.0.0", *port, nodeType)

	// Start network node
	if err := networkNode.Start(); err != nil {
		log.Fatalf("Failed to start network node: %v", err)
	}
	log.Printf("Network node listening on 0.0.0.0:%d\n", *port)

	// Connect to peers if specified
	if *peersStr != "" {
		peers := strings.Split(*peersStr, ",")
		for _, peer := range peers {
			peer = strings.TrimSpace(peer)
			if peer != "" {
				log.Printf("Connecting to peer: %s\n", peer)
				if err := networkNode.ConnectToPeer(peer); err != nil {
					log.Printf("Warning: Failed to connect to peer %s: %v\n", peer, err)
				} else {
					log.Printf("Successfully connected to peer: %s\n", peer)
				}
			}
		}
	}

	// Initialize Consensus Manager
	log.Println("Initializing Consensus Manager...")
	consensusManager := consensus.NewConsensusManager(nodeType)

	// Initialize Ledger
	log.Printf("Initializing Ledger (database: %s)...\n", *dbPath)
	ledger, err := storage.NewLedger(*dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize ledger: %v", err)
	}
	defer ledger.Close()

	// Load existing blockchain state
	latestBlock, chainHeight, err := ledger.LoadBlockchainState()
	if err != nil {
		log.Fatalf("Failed to load blockchain state: %v", err)
	}

	if chainHeight > 0 {
		log.Printf("Loaded existing blockchain: height=%d, latest block slot=%d\n",
			chainHeight, latestBlock.Header.Slot)
	} else {
		log.Println("No existing blockchain found, starting fresh")
	}

	// Initialize Verifier
	log.Println("Initializing Verifier...")
	verifier := verification.NewVerifier()

	// Verify existing chain if it exists
	if chainHeight > 0 {
		log.Println("Verifying existing blockchain...")
		if err := verifier.VerifyChain(ledger); err != nil {
			log.Fatalf("Blockchain verification failed: %v", err)
		}
		log.Println("Blockchain verification successful")
	}

	log.Println("Node initialization complete")

	// Setup graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start node logic based on type
	stopChan := make(chan struct{})

	if nodeType == network.LEADER {
		log.Println("Starting leader node logic...")
		go runLeaderNode(consensusManager, blockProducer, networkNode, ledger, chainHeight, stopChan)
	} else {
		log.Println("Starting replica node logic...")
		go runReplicaNode(consensusManager, verifier, networkNode, ledger, stopChan)
	}

	// Wait for shutdown signal
	<-sigChan
	log.Println("Shutdown signal received, stopping node...")

	// Signal goroutines to stop
	close(stopChan)

	// Stop network node
	networkNode.Stop()
	log.Println("Network node stopped")

	// Close ledger
	if err := ledger.Close(); err != nil {
		log.Printf("Error closing ledger: %v\n", err)
	}
	log.Println("Ledger closed")

	log.Println("Node shutdown complete")
}

func runLeaderNode(cm *consensus.ConsensusManager, bp *blockchain.BlockProducer, nn *network.NetworkNode, ledger *storage.Ledger, currentHeight int64, stopChan <-chan struct{}) {
	log.Println("Leader node: Starting continuous block production")

	blockHeight := currentHeight

	for {
		select {
		case <-stopChan:
			log.Println("Leader node: Stopping block production")
			return
		default:
			// Get current slot
			currentSlot := cm.GetCurrentSlot()

			// Wait for slot to start
			cm.WaitForSlotStart(currentSlot)

			// Check if we're the leader for this slot
			if !cm.IsLeader(currentSlot) {
				log.Printf("Leader node: Not leader for slot %d, skipping\n", currentSlot)
				continue
			}

			log.Printf("Leader node: Producing block for slot %d\n", currentSlot)

			// Produce block
			block, err := bp.ProduceBlock(currentSlot)
			if err != nil {
				log.Printf("Leader node: Error producing block: %v\n", err)
				continue
			}

			// Set block height and previous block hash
			blockHeight++
			block.Header.BlockHeight = blockHeight

			// Get previous block hash if not genesis
			if blockHeight > 1 {
				prevBlock, err := ledger.GetBlockByHeight(blockHeight - 1)
				if err != nil {
					log.Printf("Leader node: Error getting previous block: %v\n", err)
					blockHeight-- // Rollback height increment
					continue
				}
				block.Header.PreviousBlockHash = prevBlock.Header.MerkleRoot
			}

			log.Printf("Leader node: Block produced - height=%d, slot=%d, entries=%d\n",
				block.Header.BlockHeight, block.Header.Slot, len(block.Entries))

			// Store block to ledger
			if err := ledger.StoreBlock(block); err != nil {
				log.Printf("Leader node: Error storing block: %v\n", err)
				blockHeight-- // Rollback height increment
				continue
			}

			log.Printf("Leader node: Block stored to ledger - height=%d\n", block.Header.BlockHeight)

			// Broadcast block to network
			if err := nn.BroadcastBlock(block); err != nil {
				log.Printf("Leader node: Error broadcasting block: %v\n", err)
				// Continue even if broadcast fails - block is already stored
			} else {
				log.Printf("Leader node: Block broadcasted to peers - height=%d\n", block.Header.BlockHeight)
			}
		}
	}
}

func runReplicaNode(cm *consensus.ConsensusManager, v *verification.Verifier, nn *network.NetworkNode, ledger *storage.Ledger, stopChan <-chan struct{}) {
	log.Println("Replica node: Starting block reception and validation")

	for {
		select {
		case <-stopChan:
			log.Println("Replica node: Stopping block reception")
			return
		default:
			// Receive block from network
			block, err := nn.ReceiveBlock()
			if err != nil {
				// Check if error is due to shutdown
				select {
				case <-stopChan:
					return
				default:
					log.Printf("Replica node: Error receiving block: %v\n", err)
					continue
				}
			}

			log.Printf("Replica node: Received block - height=%d, slot=%d, entries=%d\n",
				block.Header.BlockHeight, block.Header.Slot, len(block.Entries))

			// Validate block with consensus manager
			if err := cm.ValidateBlock(block); err != nil {
				log.Printf("Replica node: Block validation failed (consensus): %v\n", err)
				continue
			}

			log.Printf("Replica node: Block passed consensus validation - height=%d\n", block.Header.BlockHeight)

			// Verify block with verifier
			if err := v.VerifyBlock(block); err != nil {
				log.Printf("Replica node: Block verification failed: %v\n", err)
				continue
			}

			log.Printf("Replica node: Block passed verification - height=%d\n", block.Header.BlockHeight)

			// Verify block linkage if not genesis
			if block.Header.BlockHeight > 1 {
				prevBlock, err := ledger.GetBlockByHeight(block.Header.BlockHeight - 1)
				if err != nil {
					log.Printf("Replica node: Error getting previous block for linkage verification: %v\n", err)
					continue
				}

				if err := v.VerifyBlockLink(block, prevBlock); err != nil {
					log.Printf("Replica node: Block linkage verification failed: %v\n", err)
					continue
				}

				log.Printf("Replica node: Block linkage verified - height=%d\n", block.Header.BlockHeight)
			}

			// Store valid block to ledger
			if err := ledger.StoreBlock(block); err != nil {
				log.Printf("Replica node: Error storing block: %v\n", err)
				continue
			}

			log.Printf("Replica node: Block stored successfully - height=%d, slot=%d\n",
				block.Header.BlockHeight, block.Header.Slot)
		}
	}
}
