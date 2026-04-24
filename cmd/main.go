package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/poh-blockchain/internal/blockchain"
	"github.com/poh-blockchain/internal/consensus"
	"github.com/poh-blockchain/internal/filestore"
	"github.com/poh-blockchain/internal/network"
	"github.com/poh-blockchain/internal/poh"
	"github.com/poh-blockchain/internal/processor"
	"github.com/poh-blockchain/internal/quanticscript"
	"github.com/poh-blockchain/internal/runtime"
	"github.com/poh-blockchain/internal/storage"
	"github.com/poh-blockchain/internal/system"
	"github.com/poh-blockchain/internal/transaction"
	"github.com/poh-blockchain/internal/verification"
)

func main() {
	// Check if a subcommand is provided
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "account":
			handleAccountCommand()
			return
		case "transfer":
			handleTransferCommand()
			return
		case "query":
			handleQueryCommand()
			return
		case "submit":
			handleSubmitCommand()
			return
		case "status":
			handleStatusCommand()
			return
		case "qsc":
			handleQuanticScriptCommand()
			return
		case "help":
			printHelp()
			return
		}
	}

	// Parse command-line flags for node operation
	nodeTypeStr := flag.String("type", "replica", "Node type: leader or replica")
	port := flag.Int("port", 8080, "Port to listen on")
	peersStr := flag.String("peers", "", "Comma-separated list of peer addresses (host:port)")
	dbPath := flag.String("db", "blockchain.db", "Path to the database file")
	malicious := flag.Bool("malicious", false, "Run node in malicious mode (for BFT testing)")
	flag.Parse()

	// Determine node type
	var nodeType network.NodeType
	switch strings.ToLower(*nodeTypeStr) {
	case "leader":
		nodeType = network.LEADER
		if *malicious {
			log.Println("Starting node as MALICIOUS LEADER")
		} else {
			log.Println("Starting node as LEADER")
		}
	case "replica":
		nodeType = network.REPLICA
		if *malicious {
			log.Println("Starting node as MALICIOUS REPLICA")
		} else {
			log.Println("Starting node as REPLICA")
		}
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
		go runLeaderNode(consensusManager, blockProducer, networkNode, ledger, chainHeight, *malicious, stopChan)
	} else {
		log.Println("Starting replica node logic...")
		go runReplicaNode(consensusManager, verifier, networkNode, ledger, *malicious, stopChan)
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

func runLeaderNode(cm *consensus.ConsensusManager, bp *blockchain.BlockProducer, nn *network.NetworkNode, ledger *storage.Ledger, currentHeight int64, malicious bool, stopChan <-chan struct{}) {
	log.Println("Leader node: Starting continuous block production")

	blockHeight := currentHeight
	maliciousCounter := 0

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

			// Produce block (with empty state root for now)
			block, err := bp.ProduceBlock(currentSlot, []byte{})
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

			// MALICIOUS BEHAVIOR: Corrupt blocks periodically
			if malicious {
				maliciousCounter++
				if maliciousCounter%3 == 0 {
					// Corrupt the block by modifying entries
					if len(block.Entries) > 0 {
						block.Entries[0].NumHashes = 1 // Invalid hash count
						log.Printf("MALICIOUS: Corrupted block %d by setting invalid hash count\n", blockHeight)
					}
				}
				if maliciousCounter%5 == 0 {
					// Send block with wrong previous hash
					block.Header.PreviousBlockHash = []byte("corrupted-hash")
					log.Printf("MALICIOUS: Corrupted block %d with wrong previous hash\n", blockHeight)
				}
			}

			log.Printf("Leader node: Block produced - height=%d, slot=%d, entries=%d\n",
				block.Header.BlockHeight, block.Header.Slot, len(block.Entries))

			// Store block to ledger (only if not malicious or not corrupted)
			if !malicious || maliciousCounter%3 != 0 {
				if err := ledger.StoreBlock(block); err != nil {
					log.Printf("Leader node: Error storing block: %v\n", err)
					blockHeight-- // Rollback height increment
					continue
				}
				log.Printf("Leader node: Block stored to ledger - height=%d\n", block.Header.BlockHeight)
			} else {
				log.Printf("MALICIOUS: Skipping storage of corrupted block %d\n", blockHeight)
			}

			// Broadcast block to network (including corrupted ones)
			if err := nn.BroadcastBlock(block); err != nil {
				log.Printf("Leader node: Error broadcasting block: %v\n", err)
				// Continue even if broadcast fails - block is already stored
			} else {
				log.Printf("Leader node: Block broadcasted to peers - height=%d\n", block.Header.BlockHeight)
			}
		}
	}
}

func runReplicaNode(cm *consensus.ConsensusManager, v *verification.Verifier, nn *network.NetworkNode, ledger *storage.Ledger, malicious bool, stopChan <-chan struct{}) {
	log.Println("Replica node: Starting block reception and validation")

	maliciousCounter := 0

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

			// MALICIOUS BEHAVIOR: Skip validation periodically
			if malicious {
				maliciousCounter++
				if maliciousCounter%4 == 0 {
					log.Printf("MALICIOUS: Skipping validation for block %d\n", block.Header.BlockHeight)
					// Store block without validation
					if err := ledger.StoreBlock(block); err != nil {
						log.Printf("MALICIOUS: Error storing unvalidated block: %v\n", err)
					} else {
						log.Printf("MALICIOUS: Stored block %d without validation\n", block.Header.BlockHeight)
					}
					continue
				}
			}

			// Validate block with consensus manager
			if err := cm.ValidateBlock(block); err != nil {
				log.Printf("Replica node: Block validation failed (consensus): %v\n", err)
				if malicious && maliciousCounter%6 == 0 {
					log.Printf("MALICIOUS: Ignoring validation failure, storing anyway\n")
					ledger.StoreBlock(block)
				}
				continue
			}

			log.Printf("Replica node: Block passed consensus validation - height=%d\n", block.Header.BlockHeight)

			// Verify block with verifier
			if err := v.VerifyBlock(block); err != nil {
				log.Printf("Replica node: Block verification failed: %v\n", err)
				if malicious && maliciousCounter%6 == 0 {
					log.Printf("MALICIOUS: Ignoring verification failure, storing anyway\n")
					ledger.StoreBlock(block)
				}
				continue
			}

			log.Printf("Replica node: Block passed verification - height=%d\n", block.Header.BlockHeight)

			// Verify block linkage if not genesis
			linkageVerified := true
			if block.Header.BlockHeight > 1 {
				prevBlock, err := ledger.GetBlockByHeight(block.Header.BlockHeight - 1)
				if err != nil {
					// Previous block not found - this can happen if blocks arrive out of order
					// Store the block anyway and linkage will be verified when previous blocks arrive
					log.Printf("Replica node: Warning - previous block not found, storing anyway (height=%d)\n", block.Header.BlockHeight)
					linkageVerified = false
				} else {
					if err := v.VerifyBlockLink(block, prevBlock); err != nil {
						log.Printf("Replica node: Block linkage verification failed: %v\n", err)
						if malicious && maliciousCounter%6 == 0 {
							log.Printf("MALICIOUS: Ignoring linkage failure, storing anyway\n")
							ledger.StoreBlock(block)
						}
						continue
					}
					log.Printf("Replica node: Block linkage verified - height=%d\n", block.Header.BlockHeight)
				}
			}

			// Store valid block to ledger
			if err := ledger.StoreBlock(block); err != nil {
				log.Printf("Replica node: Error storing block: %v\n", err)
				continue
			}

			if linkageVerified || block.Header.BlockHeight == 1 {
				log.Printf("Replica node: Block stored successfully - height=%d, slot=%d\n",
					block.Header.BlockHeight, block.Header.Slot)
			} else {
				log.Printf("Replica node: Block stored (linkage pending) - height=%d, slot=%d\n",
					block.Header.BlockHeight, block.Header.Slot)
			}
		}
	}
}

// CLI Command Handlers

func printHelp() {
	fmt.Println("PoH Blockchain CLI")
	fmt.Println("\nUsage:")
	fmt.Println("  poh-blockchain [command] [options]")
	fmt.Println("\nCommands:")
	fmt.Println("  account create --balance <amount> --output <file>")
	fmt.Println("    Create a new account with initial balance")
	fmt.Println("    Generates a new keypair and saves it to the output file")
	fmt.Println()
	fmt.Println("  transfer --from <keypair-file> --to <address> --amount <amount> --state <db-path>")
	fmt.Println("    Transfer balance from one account to another")
	fmt.Println()
	fmt.Println("  query --address <address> --state <db-path>")
	fmt.Println("    Query account information")
	fmt.Println()
	fmt.Println("  submit --tx <tx-file> --state <db-path>")
	fmt.Println("    Submit a transaction from a JSON file")
	fmt.Println()
	fmt.Println("  status --tx <tx-id> --state <db-path>")
	fmt.Println("    Check transaction status")
	fmt.Println()
	fmt.Println("  qsc <subcommand> [options]")
	fmt.Println("    QuanticScript compiler commands:")
	fmt.Println("      compile --input <source.qs> --output <bytecode.qsb>")
	fmt.Println("        Compile QuanticScript source to bytecode")
	fmt.Println("      assemble --input <assembly.qsa> --output <bytecode.qsb>")
	fmt.Println("        Assemble QuanticScript assembly to bytecode")
	fmt.Println("      disassemble --input <bytecode.qsb> --output <assembly.qsa>")
	fmt.Println("        Disassemble bytecode to QuanticScript assembly")
	fmt.Println()
	fmt.Println("  help")
	fmt.Println("    Show this help message")
	fmt.Println()
	fmt.Println("Node Operation:")
	fmt.Println("  poh-blockchain --type <leader|replica> --port <port> [options]")
	fmt.Println("    Run as a blockchain node")
}

// handleAccountCommand creates a new account
func handleAccountCommand() {
	fs := flag.NewFlagSet("account", flag.ExitOnError)
	balance := fs.Int64("balance", 1000000, "Initial balance for the account")
	output := fs.String("output", "data/keypair.json", "Output file for keypair")
	stateDB := fs.String("state", "data/state.db", "Path to state database")

	if len(os.Args) < 3 || os.Args[2] != "create" {
		fmt.Println("Usage: account create --balance <amount> --output <file> --state <db-path>")
		os.Exit(1)
	}

	fs.Parse(os.Args[3:])

	// Generate new keypair
	pubKey, privKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		log.Fatalf("Failed to generate keypair: %v", err)
	}

	// Convert to our PublicKey type
	var txPubKey transaction.PublicKey
	copy(txPubKey[:], pubKey)

	// Derive file ID from public key
	fileID := publicKeyToFileID(txPubKey)

	// Initialize file store
	fs2, err := initFileStore(*stateDB)
	if err != nil {
		log.Fatalf("Failed to open state database: %v", err)
	}
	defer fs2.Close()

	// Initialize runtime and processor
	rt := runtime.NewRuntime()
	rt.RegisterBuiltinProgram(system.NewSystemProgram())
	txProcessor := processor.NewTxProcessor(fs2, rt)

	// Create account file directly in the store (genesis account)
	accountFile := &filestore.File{
		ID:         fileID,
		Balance:    *balance,
		TxManager:  system.SystemProgramID,
		Data:       []byte{},
		Executable: false,
	}

	createdID, err := fs2.CreateFile(accountFile)
	if err != nil {
		log.Fatalf("Failed to create account: %v", err)
	}

	// Save keypair to file
	keypair := map[string]string{
		"public_key":  hex.EncodeToString(pubKey),
		"private_key": hex.EncodeToString(privKey),
		"address":     createdID.String(),
	}

	data, err := json.MarshalIndent(keypair, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal keypair: %v", err)
	}

	// Ensure output directory exists
	if idx := strings.LastIndex(*output, "/"); idx != -1 {
		dir := (*output)[:idx]
		if err := os.MkdirAll(dir, 0755); err != nil {
			log.Fatalf("Failed to create output directory: %v", err)
		}
	}

	if err := os.WriteFile(*output, data, 0600); err != nil {
		log.Fatalf("Failed to write keypair file: %v", err)
	}

	fmt.Printf("Account created successfully!\n")
	fmt.Printf("Address: %s\n", createdID.String())
	fmt.Printf("Balance: %d\n", *balance)
	fmt.Printf("Keypair saved to: %s\n", *output)

	_ = txProcessor // Suppress unused warning
}

// handleTransferCommand transfers balance between accounts
func handleTransferCommand() {
	fs := flag.NewFlagSet("transfer", flag.ExitOnError)
	fromFile := fs.String("from", "", "Keypair file for source account")
	toAddr := fs.String("to", "", "Destination account address")
	amount := fs.Int64("amount", 0, "Amount to transfer")
	stateDB := fs.String("state", "data/state.db", "Path to state database")

	fs.Parse(os.Args[2:])

	if *fromFile == "" || *toAddr == "" || *amount <= 0 {
		fmt.Println("Usage: transfer --from <keypair-file> --to <address> --amount <amount> --state <db-path>")
		os.Exit(1)
	}

	// Load keypair
	keypairData, err := os.ReadFile(*fromFile)
	if err != nil {
		log.Fatalf("Failed to read keypair file: %v", err)
	}

	var keypair map[string]string
	if err := json.Unmarshal(keypairData, &keypair); err != nil {
		log.Fatalf("Failed to parse keypair file: %v", err)
	}

	pubKeyBytes, err := hex.DecodeString(keypair["public_key"])
	if err != nil {
		log.Fatalf("Invalid public key: %v", err)
	}

	privKeyBytes, err := hex.DecodeString(keypair["private_key"])
	if err != nil {
		log.Fatalf("Invalid private key: %v", err)
	}

	var txPubKey transaction.PublicKey
	copy(txPubKey[:], pubKeyBytes)

	fromID, err := filestore.FileIDFromString(keypair["address"])
	if err != nil {
		log.Fatalf("Invalid from address: %v", err)
	}

	toID, err := filestore.FileIDFromString(*toAddr)
	if err != nil {
		log.Fatalf("Invalid to address: %v", err)
	}

	// Initialize file store
	fs2, err := initFileStore(*stateDB)
	if err != nil {
		log.Fatalf("Failed to open state database: %v", err)
	}
	defer fs2.Close()

	// Initialize runtime and processor
	rt := runtime.NewRuntime()
	rt.RegisterBuiltinProgram(system.NewSystemProgram())
	txProcessor := processor.NewTxProcessor(fs2, rt)

	// Create transfer instruction
	transferData := system.EncodeTransferInstruction(*amount)
	instruction := transaction.Instruction{
		ProgramID: system.SystemProgramID,
		Inputs: map[string]transaction.FileAccess{
			"from": {FileID: fromID, Permission: transaction.Write},
			"to":   {FileID: toID, Permission: transaction.Write},
		},
		Data: transferData,
	}

	// Create transaction
	tx := &transaction.Transaction{
		LastSeen:     transaction.TxID{},
		Instructions: []transaction.Instruction{instruction},
		Signatures:   []transaction.Signature{},
	}

	// Sign transaction
	txData, err := tx.Marshal()
	if err != nil {
		log.Fatalf("Failed to marshal transaction: %v", err)
	}

	signature := ed25519.Sign(ed25519.PrivateKey(privKeyBytes), txData)
	var sig [64]byte
	copy(sig[:], signature)

	tx.Signatures = []transaction.Signature{
		{PublicKey: txPubKey, Signature: sig},
	}

	// Process transaction
	result, err := txProcessor.ProcessTransaction(tx)
	if err != nil {
		log.Fatalf("Transaction failed: %v", err)
	}

	if !result.Success {
		log.Fatalf("Transaction failed: %v", result.Error)
	}

	fmt.Printf("Transfer successful!\n")
	fmt.Printf("Transaction ID: %s\n", result.TxID.String())
	fmt.Printf("Amount: %d\n", *amount)
	fmt.Printf("From: %s\n", fromID.String())
	fmt.Printf("To: %s\n", toID.String())
}

// handleQueryCommand queries account information
func handleQueryCommand() {
	fs := flag.NewFlagSet("query", flag.ExitOnError)
	address := fs.String("address", "", "Account address to query")
	stateDB := fs.String("state", "data/state.db", "Path to state database")

	fs.Parse(os.Args[2:])

	if *address == "" {
		fmt.Println("Usage: query --address <address> --state <db-path>")
		os.Exit(1)
	}

	fileID, err := filestore.FileIDFromString(*address)
	if err != nil {
		log.Fatalf("Invalid address: %v", err)
	}

	// Initialize file store
	fs2, err := initFileStore(*stateDB)
	if err != nil {
		log.Fatalf("Failed to open state database: %v", err)
	}
	defer fs2.Close()

	// Query file
	file, err := fs2.GetFile(fileID)
	if err != nil {
		log.Fatalf("Failed to query account: %v", err)
	}

	fmt.Printf("Account Information:\n")
	fmt.Printf("  Address: %s\n", file.ID.String())
	fmt.Printf("  Balance: %d\n", file.Balance)
	fmt.Printf("  TxManager: %s\n", file.TxManager.String())
	fmt.Printf("  Executable: %v\n", file.Executable)
	fmt.Printf("  Data Size: %d bytes\n", len(file.Data))
	fmt.Printf("  Created: %s\n", file.CreatedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("  Updated: %s\n", file.UpdatedAt.Format("2006-01-02 15:04:05"))

	// Calculate storage cost
	storageCost := filestore.CalculateStorageCost(int64(len(file.Data)))
	fmt.Printf("  Storage Cost: %d\n", storageCost)
}

// handleSubmitCommand submits a transaction from a JSON file
func handleSubmitCommand() {
	fs := flag.NewFlagSet("submit", flag.ExitOnError)
	txFile := fs.String("tx", "", "Transaction JSON file")
	stateDB := fs.String("state", "data/state.db", "Path to state database")

	fs.Parse(os.Args[2:])

	if *txFile == "" {
		fmt.Println("Usage: submit --tx <tx-file> --state <db-path>")
		os.Exit(1)
	}

	// Load transaction from file
	txData, err := os.ReadFile(*txFile)
	if err != nil {
		log.Fatalf("Failed to read transaction file: %v", err)
	}

	tx, err := transaction.UnmarshalTransaction(txData)
	if err != nil {
		log.Fatalf("Failed to parse transaction: %v", err)
	}

	// Initialize file store
	fs2, err := initFileStore(*stateDB)
	if err != nil {
		log.Fatalf("Failed to open state database: %v", err)
	}
	defer fs2.Close()

	// Initialize runtime and processor
	rt := runtime.NewRuntime()
	rt.RegisterBuiltinProgram(system.NewSystemProgram())
	txProcessor := processor.NewTxProcessor(fs2, rt)

	// Process transaction
	result, err := txProcessor.ProcessTransaction(tx)
	if err != nil {
		log.Fatalf("Transaction failed: %v", err)
	}

	if !result.Success {
		log.Fatalf("Transaction failed: %v", result.Error)
	}

	fmt.Printf("Transaction submitted successfully!\n")
	fmt.Printf("Transaction ID: %s\n", result.TxID.String())
	fmt.Printf("Gas Used: %d\n", result.GasUsed)
}

// handleStatusCommand checks transaction status
func handleStatusCommand() {
	fs := flag.NewFlagSet("status", flag.ExitOnError)
	txID := fs.String("tx", "", "Transaction ID")
	stateDB := fs.String("state", "data/state.db", "Path to state database")

	fs.Parse(os.Args[2:])

	if *txID == "" {
		fmt.Println("Usage: status --tx <tx-id> --state <db-path>")
		os.Exit(1)
	}

	// Initialize file store
	fs2, err := initFileStore(*stateDB)
	if err != nil {
		log.Fatalf("Failed to open state database: %v", err)
	}
	defer fs2.Close()

	// For now, we just confirm the state database is accessible
	// A full implementation would track transaction history
	fmt.Printf("Transaction Status Check:\n")
	fmt.Printf("  Transaction ID: %s\n", *txID)
	fmt.Printf("  Note: Transaction history tracking not yet implemented\n")
	fmt.Printf("  State database: %s (accessible)\n", *stateDB)
}

// publicKeyToFileID converts a public key to a file ID
func publicKeyToFileID(pubkey transaction.PublicKey) filestore.FileID {
	var fileID filestore.FileID
	copy(fileID[:], pubkey[:])
	return fileID
}

// ensureSystemProgram ensures the system program file exists in the state
func ensureSystemProgram(fs *filestore.FileStore) {
	// Check if system program already exists
	_, err := fs.GetFile(system.SystemProgramID)
	if err == nil {
		return // Already exists
	}

	// Create system program file with sufficient balance for storage
	systemProgram := &filestore.File{
		ID:         system.SystemProgramID,
		Balance:    1000000,                // Sufficient balance for storage
		TxManager:  system.SystemProgramID, // Self-managed
		Data:       []byte("builtin-system-program"),
		Executable: true,
	}

	_, err = fs.CreateFile(systemProgram)
	if err != nil {
		log.Printf("Warning: Failed to create system program file: %v", err)
	}
}

// initFileStore initializes a file store and ensures system program exists
func initFileStore(dbPath string) (*filestore.FileStore, error) {
	// Ensure the directory exists
	dir := "."
	if idx := strings.LastIndex(dbPath, "/"); idx != -1 {
		dir = dbPath[:idx]
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	fs, err := filestore.NewFileStore(dbPath)
	if err != nil {
		return nil, err
	}
	ensureSystemProgram(fs)
	return fs, nil
}

// handleQuanticScriptCommand handles QuanticScript compiler commands
func handleQuanticScriptCommand() {
	if len(os.Args) < 3 {
		printQuanticScriptHelp()
		os.Exit(1)
	}

	subcommand := os.Args[2]

	switch subcommand {
	case "compile":
		handleCompileCommand()
	case "assemble":
		handleAssembleCommand()
	case "disassemble":
		handleDisassembleCommand()
	case "help":
		printQuanticScriptHelp()
	default:
		fmt.Printf("Unknown QuanticScript subcommand: %s\n", subcommand)
		printQuanticScriptHelp()
		os.Exit(1)
	}
}

// printQuanticScriptHelp prints help for QuanticScript compiler commands
func printQuanticScriptHelp() {
	fmt.Println("QuanticScript Compiler")
	fmt.Println("\nUsage:")
	fmt.Println("  poh-blockchain qsc <subcommand> [options]")
	fmt.Println("\nSubcommands:")
	fmt.Println("  compile --input <source.qs> --output <bytecode.qsb> [--verbose]")
	fmt.Println("    Compile QuanticScript source code to bytecode")
	fmt.Println("    Options:")
	fmt.Println("      --input, -i    Input source file (.qs)")
	fmt.Println("      --output, -o   Output bytecode file (.qsb)")
	fmt.Println("      --verbose, -v  Enable verbose output with diagnostics")
	fmt.Println()
	fmt.Println("  assemble --input <assembly.qsa> --output <bytecode.qsb> [--verbose]")
	fmt.Println("    Assemble QuanticScript assembly to bytecode")
	fmt.Println("    Options:")
	fmt.Println("      --input, -i    Input assembly file (.qsa)")
	fmt.Println("      --output, -o   Output bytecode file (.qsb)")
	fmt.Println("      --verbose, -v  Enable verbose output with diagnostics")
	fmt.Println()
	fmt.Println("  disassemble --input <bytecode.qsb> --output <assembly.qsa> [--verbose]")
	fmt.Println("    Disassemble bytecode to QuanticScript assembly")
	fmt.Println("    Options:")
	fmt.Println("      --input, -i    Input bytecode file (.qsb)")
	fmt.Println("      --output, -o   Output assembly file (.qsa)")
	fmt.Println("      --verbose, -v  Enable verbose output with diagnostics")
	fmt.Println()
	fmt.Println("  help")
	fmt.Println("    Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  # Compile source to bytecode")
	fmt.Println("  poh-blockchain qsc compile -i program.qs -o program.qsb")
	fmt.Println()
	fmt.Println("  # Assemble assembly to bytecode")
	fmt.Println("  poh-blockchain qsc assemble -i program.qsa -o program.qsb")
	fmt.Println()
	fmt.Println("  # Disassemble bytecode to assembly")
	fmt.Println("  poh-blockchain qsc disassemble -i program.qsb -o program.qsa")
}

// handleCompileCommand compiles QuanticScript source to bytecode
func handleCompileCommand() {
	fs := flag.NewFlagSet("compile", flag.ExitOnError)
	input := fs.String("input", "", "Input source file (.qs)")
	inputShort := fs.String("i", "", "Input source file (.qs) (short)")
	output := fs.String("output", "", "Output bytecode file (.qsb)")
	outputShort := fs.String("o", "", "Output bytecode file (.qsb) (short)")
	verbose := fs.Bool("verbose", false, "Enable verbose output")
	verboseShort := fs.Bool("v", false, "Enable verbose output (short)")

	fs.Parse(os.Args[3:])

	// Use short flags if long flags are empty
	if *input == "" {
		input = inputShort
	}
	if *output == "" {
		output = outputShort
	}
	if !*verbose {
		verbose = verboseShort
	}

	if *input == "" || *output == "" {
		fmt.Println("Error: --input and --output are required")
		fmt.Println()
		printQuanticScriptHelp()
		os.Exit(1)
	}

	if *verbose {
		fmt.Printf("Compiling %s to %s...\n", *input, *output)
	}

	// Read source file
	sourceCode, err := os.ReadFile(*input)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading source file: %v\n", err)
		os.Exit(1)
	}

	// Split source into lines for error reporting
	sourceLines := strings.Split(string(sourceCode), "\n")

	// Compile source to bytecode
	bytecode, errors := compileSource(string(sourceCode), *verbose)
	if len(errors) > 0 {
		fmt.Fprintf(os.Stderr, "\nCompilation failed with %d error(s):\n\n", len(errors))
		for i, err := range errors {
			if *verbose {
				// Verbose mode: show detailed error with suggestions
				fmt.Fprintf(os.Stderr, "%s\n", formatCompilationError(err, sourceLines))
			} else {
				// Normal mode: show error message only
				fmt.Fprintf(os.Stderr, "%d. %v\n", i+1, err)
			}
		}
		os.Exit(1)
	}

	// Write bytecode to output file
	if err := os.WriteFile(*output, bytecode, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing output file: %v\n", err)
		os.Exit(1)
	}

	if *verbose {
		fmt.Printf("Successfully compiled to %s (%d bytes)\n", *output, len(bytecode))
	} else {
		fmt.Printf("Compiled successfully: %s\n", *output)
	}
}

// handleAssembleCommand assembles QuanticScript assembly to bytecode
func handleAssembleCommand() {
	fs := flag.NewFlagSet("assemble", flag.ExitOnError)
	input := fs.String("input", "", "Input assembly file (.qsa)")
	inputShort := fs.String("i", "", "Input assembly file (.qsa) (short)")
	output := fs.String("output", "", "Output bytecode file (.qsb)")
	outputShort := fs.String("o", "", "Output bytecode file (.qsb) (short)")
	verbose := fs.Bool("verbose", false, "Enable verbose output")
	verboseShort := fs.Bool("v", false, "Enable verbose output (short)")

	fs.Parse(os.Args[3:])

	// Use short flags if long flags are empty
	if *input == "" {
		input = inputShort
	}
	if *output == "" {
		output = outputShort
	}
	if !*verbose {
		verbose = verboseShort
	}

	if *input == "" || *output == "" {
		fmt.Println("Error: --input and --output are required")
		fmt.Println()
		printQuanticScriptHelp()
		os.Exit(1)
	}

	if *verbose {
		fmt.Printf("Assembling %s to %s...\n", *input, *output)
	}

	// Read assembly file
	assemblyCode, err := os.ReadFile(*input)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading assembly file: %v\n", err)
		os.Exit(1)
	}

	// Assemble to bytecode
	bytecode, err := quanticscript.AssembleToFile(string(assemblyCode))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Assembly failed: %v\n", err)
		os.Exit(1)
	}

	// Write bytecode to output file
	if err := os.WriteFile(*output, bytecode, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing output file: %v\n", err)
		os.Exit(1)
	}

	if *verbose {
		fmt.Printf("Successfully assembled to %s (%d bytes)\n", *output, len(bytecode))
	} else {
		fmt.Printf("Assembled successfully: %s\n", *output)
	}
}

// handleDisassembleCommand disassembles bytecode to QuanticScript assembly
func handleDisassembleCommand() {
	fs := flag.NewFlagSet("disassemble", flag.ExitOnError)
	input := fs.String("input", "", "Input bytecode file (.qsb)")
	inputShort := fs.String("i", "", "Input bytecode file (.qsb) (short)")
	output := fs.String("output", "", "Output assembly file (.qsa)")
	outputShort := fs.String("o", "", "Output assembly file (.qsa) (short)")
	verbose := fs.Bool("verbose", false, "Enable verbose output")
	verboseShort := fs.Bool("v", false, "Enable verbose output (short)")

	fs.Parse(os.Args[3:])

	// Use short flags if long flags are empty
	if *input == "" {
		input = inputShort
	}
	if *output == "" {
		output = outputShort
	}
	if !*verbose {
		verbose = verboseShort
	}

	if *input == "" || *output == "" {
		fmt.Println("Error: --input and --output are required")
		fmt.Println()
		printQuanticScriptHelp()
		os.Exit(1)
	}

	if *verbose {
		fmt.Printf("Disassembling %s to %s...\n", *input, *output)
	}

	// Read bytecode file
	bytecode, err := os.ReadFile(*input)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading bytecode file: %v\n", err)
		os.Exit(1)
	}

	// Disassemble bytecode
	assembly, err := quanticscript.DisassembleFile(bytecode)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Disassembly failed: %v\n", err)
		os.Exit(1)
	}

	// Write assembly to output file
	if err := os.WriteFile(*output, []byte(assembly), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing output file: %v\n", err)
		os.Exit(1)
	}

	if *verbose {
		fmt.Printf("Successfully disassembled to %s (%d bytes)\n", *output, len(assembly))
	} else {
		fmt.Printf("Disassembled successfully: %s\n", *output)
	}
}

// compileSource compiles QuanticScript source code to bytecode
func compileSource(source string, verbose bool) ([]byte, []error) {
	// Lexical analysis
	if verbose {
		fmt.Println("Phase 1: Lexical analysis...")
	}
	lexer := quanticscript.NewLexer(source, "<input>")

	// Parsing
	if verbose {
		fmt.Println("Phase 2: Parsing...")
	}
	parser := quanticscript.NewParser(lexer)
	program := parser.ParseProgram()

	if len(parser.Errors()) > 0 {
		if verbose {
			fmt.Println("  Parsing failed")
		}
		return nil, parser.Errors()
	}

	if verbose {
		fmt.Printf("  Parsed %d declarations\n", len(program.Declarations))
	}

	// Type checking
	if verbose {
		fmt.Println("Phase 3: Type checking...")
	}
	typeChecker := quanticscript.NewTypeChecker()
	typeChecker.CheckProgram(program)

	if len(typeChecker.Errors()) > 0 {
		if verbose {
			fmt.Println("  Type checking failed")
		}
		return nil, typeChecker.Errors()
	}

	if verbose {
		fmt.Println("  Type checking passed")
	}

	// Code generation
	if verbose {
		fmt.Println("Phase 4: Code generation...")
	}
	codeGen := quanticscript.NewCodeGenerator()
	bytecodeBody, err := codeGen.Generate(program)

	if err != nil {
		if verbose {
			fmt.Println("  Code generation failed")
		}
		return nil, []error{err}
	}

	if len(codeGen.Errors()) > 0 {
		if verbose {
			fmt.Println("  Code generation failed")
		}
		return nil, codeGen.Errors()
	}

	if verbose {
		fmt.Printf("  Generated %d bytes of bytecode\n", len(bytecodeBody))
	}

	// Create bytecode file with header
	bytecode := quanticscript.CreateBytecode(bytecodeBody, 0)

	if verbose {
		fmt.Println("Compilation complete!")
	}

	return bytecode, nil
}

// formatCompilationError formats a compilation error with helpful context
func formatCompilationError(err error, sourceLines []string) string {
	var result strings.Builder

	// Try to extract location information from error
	errStr := err.Error()

	// Write error message
	result.WriteString(fmt.Sprintf("Error: %s\n", errStr))

	// Add helpful suggestions based on error type
	if strings.Contains(errStr, "undefined variable") {
		result.WriteString("\nHelp: Make sure the variable is declared before use.\n")
		result.WriteString("      Use 'let' or 'const' to declare variables.\n")
	} else if strings.Contains(errStr, "expected") {
		result.WriteString("\nHelp: Check for missing or misplaced syntax elements.\n")
	} else if strings.Contains(errStr, "type mismatch") {
		result.WriteString("\nHelp: Ensure operands have compatible types.\n")
		result.WriteString("      You may need to add explicit type conversions.\n")
	} else if strings.Contains(errStr, "undefined function") {
		result.WriteString("\nHelp: Make sure the function is declared or imported.\n")
	} else if strings.Contains(errStr, "non-deterministic") {
		result.WriteString("\nHelp: QuanticScript requires deterministic execution.\n")
		result.WriteString("      Avoid random numbers, system time, and I/O operations.\n")
	}

	return result.String()
}
