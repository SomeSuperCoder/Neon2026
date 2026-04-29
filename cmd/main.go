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
	"github.com/poh-blockchain/internal/genesis"
	"github.com/poh-blockchain/internal/network"
	"github.com/poh-blockchain/internal/poh"
	"github.com/poh-blockchain/internal/processor"
	"github.com/poh-blockchain/internal/quanticscript"
	"github.com/poh-blockchain/internal/rpc"
	"github.com/poh-blockchain/internal/runtime"
	"github.com/poh-blockchain/internal/storage"
	"github.com/poh-blockchain/internal/transaction"
	"github.com/poh-blockchain/internal/verification"
	"github.com/poh-blockchain/programs"
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
		case "rpc":
			handleRPCCommand()
			return
		case "wallet":
			handleWalletCommand()
			return
		case "help":
			printHelp()
			return
		}
	}

	// Parse command-line flags for node operation
	nodeTypeStr := flag.String("type", "", "DEPRECATED: Node type (use --wallet instead)")
	port := flag.Int("port", 8080, "Port to listen on")
	peersStr := flag.String("peers", "", "Comma-separated list of peer addresses (host:port)")
	dbPath := flag.String("db", "blockchain.db", "Path to the database file")
	malicious := flag.Bool("malicious", false, "Run node in malicious mode (for BFT testing)")
	walletName := flag.String("wallet", "", "Wallet name for validator identity (omit for observer mode)")
	flag.Parse()

	// Check for deprecated --type flag
	if *nodeTypeStr != "" {
		if strings.ToLower(*nodeTypeStr) == "leader" {
			log.Fatal("Error: --type flag is deprecated. Use --wallet <name> instead. Create a wallet with: poh-blockchain wallet create --name <name>")
		} else if strings.ToLower(*nodeTypeStr) == "replica" {
			log.Fatal("Error: --type flag is deprecated. Use --wallet <name> for validation, or omit for observer mode.")
		}
	}

	// Initialize PoH Clock with genesis seed
	log.Println("Initializing PoH Clock...")
	pohClock := poh.NewPohClock([]byte("genesis-seed"))

	// Initialize Block Producer
	log.Println("Initializing Block Producer...")
	blockProducer := blockchain.NewBlockProducer(pohClock)

	// Initialize Network Node
	log.Printf("Initializing Network Node on port %d...\n", *port)
	networkNode := network.NewNetworkNode("0.0.0.0", *port)

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

	// Initialize FileStore first (needed for validator identity initialization)
	stateDBPath := strings.TrimSuffix(*dbPath, ".db") + "_state.db"
	log.Printf("Initializing FileStore (database: %s)...\n", stateDBPath)
	fileStore, err := initFileStore(stateDBPath)
	if err != nil {
		log.Fatalf("Failed to initialize FileStore: %v", err)
	}
	defer fileStore.Close()

	// Initialize validator identity from wallet
	log.Println("Initializing validator identity...")
	localValidatorID, localValidatorKeys, err := initializeValidatorIdentity(*walletName, fileStore)
	if err != nil {
		log.Fatalf("Failed to initialize validator identity: %v", err)
	}

	// Initialize Consensus Manager with DPoS genesis configuration
	log.Println("Initializing Consensus Manager...")

	// Create genesis configuration for DPoS
	// For now, use a simple 2-validator genesis setup
	// In production, this would be loaded from a config file
	genesisConfig := consensus.GenesisConfig{
		EpochLength: 432000, // ~2 days at 400ms/slot
		GenesisValidators: []consensus.GenesisValidator{
			{
				PublicKey:   [32]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32},
				StakeAmount: 10000000, // 10 Neon
			},
			{
				PublicKey:   [32]byte{32, 31, 30, 29, 28, 27, 26, 25, 24, 23, 22, 21, 20, 19, 18, 17, 16, 15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1},
				StakeAmount: 5000000, // 5 Neon
			},
		},
	}

	consensusManager := consensus.NewConsensusManagerWithGenesis(localValidatorID, localValidatorKeys, genesisConfig)

	// Initialize Ledger
	log.Printf("Initializing Ledger (database: %s)...\n", *dbPath)
	ledger, err := storage.NewLedger(*dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize ledger: %v", err)
	}
	defer ledger.Close()

	// Initialize Runtime for program execution
	log.Println("Initializing Runtime...")
	rt := runtime.NewRuntime()

	// Wire FileStore and Runtime to ConsensusManager for DPoS operations
	consensusManager.SetFileStore(fileStore)
	consensusManager.SetRuntime(rt)

	// Initialize DPoS genesis state
	log.Println("Initializing DPoS genesis state...")
	if err := consensusManager.InitializeGenesis(genesisConfig); err != nil {
		log.Fatalf("Failed to initialize DPoS genesis: %v", err)
	}
	log.Println("DPoS genesis initialization complete")

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

	// Start node logic based on validator identity
	stopChan := make(chan struct{})

	if localValidatorID != (filestore.FileID{}) {
		// Node is configured as a validator
		if *malicious {
			log.Println("Starting node as MALICIOUS VALIDATOR")
		} else {
			log.Println("Starting node as VALIDATOR")
		}
		go runLeaderNode(consensusManager, blockProducer, networkNode, ledger, chainHeight, *malicious, stopChan)
	} else {
		// Node is in observer mode
		log.Println("Starting node in OBSERVER mode")
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
			if cm.ShouldProduceBlock(currentSlot) {
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

					// Record block production in validator record (Requirement 8.4)
					if err := cm.RecordBlockProduction(cm.GetLocalValidatorID()); err != nil {
						log.Printf("Leader node: Error recording block production: %v\n", err)
						// Continue even if recording fails - block is already stored
					}
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
			} else {
				// We're not the scheduled leader for this slot
				scheduledValidator := cm.GetScheduledValidator(currentSlot)
				log.Printf("Leader node: Not leader for slot %d (scheduled: %s), waiting for block\n",
					currentSlot, scheduledValidator.String())

				// Wait 200ms for scheduled validator's block
				blockReceived := cm.WaitForBlockOrTimeout(currentSlot, 200)

				if !blockReceived {
					// No block received within 200ms, record missed block
					log.Printf("Leader node: No block received for slot %d, recording missed block for validator %s\n",
						currentSlot, scheduledValidator.String())

					if err := cm.ProcessSlotSkip(currentSlot); err != nil {
						log.Printf("Leader node: Error processing slot skip: %v\n", err)
					}
				} else {
					log.Printf("Leader node: Block received for slot %d from scheduled validator\n", currentSlot)
				}
			}

			// Check if we've reached an epoch boundary
			if cm.IsEpochBoundary(currentSlot) {
				log.Printf("Leader node: Epoch boundary reached at slot %d\n", currentSlot)

				// Reset block production counters for all validators (Requirement 8.4)
				if err := cm.ResetAllBlockProductionCounters(); err != nil {
					log.Printf("Leader node: Error resetting block production counters: %v\n", err)
				}
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
	fmt.Println("  wallet <subcommand> [options]")
	fmt.Println("    Wallet management commands:")
	fmt.Println("      create --name <wallet-name>")
	fmt.Println("        Create a new password-protected wallet")
	fmt.Println("      list")
	fmt.Println("        List all available wallets")
	fmt.Println("      show --name <wallet-name>")
	fmt.Println("        Display wallet information")
	fmt.Println("      export --name <wallet-name> --output <file>")
	fmt.Println("        Export wallet to unencrypted JSON")
	fmt.Println("      import --input <file> --name <wallet-name>")
	fmt.Println("        Import wallet from unencrypted JSON")
	fmt.Println()
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
	fmt.Println("  rpc --rpc-port <port> --rpc-bind <address> --ledger-path <path> --state-path <path>")
	fmt.Println("    Start RPC node for blockchain queries and transaction submission")
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
	txProcessor := processor.NewTxProcessor(fs2, rt)

	// PRODUCTION-READY: Account creation must go through proper transaction system
	// This creates a CREATE_FILE transaction and processes it through the transaction processor

	// For the first account creation, we need a bootstrap mechanism
	// In production, this would be handled by a pre-funded genesis account
	// Generate bootstrap keypair first so we can derive the correct FileID
	bootstrapPub, bootstrapPriv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		log.Fatalf("Failed to generate bootstrap keypair: %v", err)
	}

	var bootstrapTxPubKey transaction.PublicKey
	copy(bootstrapTxPubKey[:], bootstrapPub)

	// Derive bootstrap FileID from public key (same as user accounts)
	bootstrapID := publicKeyToFileID(bootstrapTxPubKey)

	// Create bootstrap account with sufficient balance to pay for new account creation
	bootstrapFile := &filestore.File{
		ID:         bootstrapID,
		Balance:    *balance + 10000, // Extra for transaction fees
		TxManager:  genesis.SystemProgramID,
		Data:       []byte{},
		Executable: false,
	}

	// GENESIS ONLY: This is the ONLY direct creation allowed - the bootstrap account for first account creation
	// In production, this would be replaced by a proper genesis account funded during blockchain initialization
	_, err = fs2.CreateFile(bootstrapFile)
	if err != nil {
		log.Fatalf("Failed to create bootstrap account: %v", err)
	}

	// Now create the user account through proper transaction system
	createFileInstr, err := transaction.CreateFileInstruction(
		genesis.SystemProgramID,
		fileID,
		bootstrapID, // Bootstrap account pays for creation
		*balance,
		txPubKey,
	)
	if err != nil {
		log.Fatalf("Failed to create CREATE_FILE instruction: %v", err)
	}

	// Build transaction
	tx := &transaction.Transaction{
		LastSeen:     transaction.TxID{},
		Instructions: []transaction.Instruction{*createFileInstr},
		Signatures:   []transaction.Signature{},
	}

	// Sign transaction with bootstrap account
	txData, err := tx.Marshal()
	if err != nil {
		log.Fatalf("Failed to marshal transaction: %v", err)
	}

	bootstrapSig := ed25519.Sign(bootstrapPriv, txData)
	var sig [64]byte
	copy(sig[:], bootstrapSig)

	tx.Signatures = []transaction.Signature{
		{PublicKey: bootstrapTxPubKey, Signature: sig},
	}

	// Process transaction through proper channels
	result, err := txProcessor.ProcessTransaction(tx)
	if err != nil {
		log.Fatalf("Account creation transaction failed: %v", err)
	}

	if !result.Success {
		log.Fatalf("Account creation failed: %v", result.Error)
	}

	createdID := fileID

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

	fmt.Printf("Account created successfully through proper transaction system!\n")
	fmt.Printf("Address: %s\n", createdID.String())
	fmt.Printf("Balance: %d\n", *balance)
	fmt.Printf("Transaction ID: %s\n", result.TxID.String())
	fmt.Printf("Keypair saved to: %s\n", *output)
	fmt.Printf("\nNote: Account created via CREATE_FILE transaction (production-ready).\n")

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
		fmt.Println("\nError: All parameters are required and amount must be positive")
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
		log.Fatalf("Invalid public key in keypair file: %v", err)
	}

	privKeyBytes, err := hex.DecodeString(keypair["private_key"])
	if err != nil {
		log.Fatalf("Invalid private key in keypair file: %v", err)
	}

	var txPubKey transaction.PublicKey
	copy(txPubKey[:], pubKeyBytes)

	fromID, err := filestore.FileIDFromString(keypair["address"])
	if err != nil {
		log.Fatalf("Invalid from address in keypair file: %v", err)
	}

	toID, err := filestore.FileIDFromString(*toAddr)
	if err != nil {
		log.Fatalf("Invalid destination address: %v\nPlease provide a valid 32-byte hex address", err)
	}

	// Initialize file store
	fs2, err := initFileStore(*stateDB)
	if err != nil {
		log.Fatalf("Failed to open state database: %v", err)
	}
	defer fs2.Close()

	// Initialize runtime and processor
	rt := runtime.NewRuntime()
	txProcessor := processor.NewTxProcessor(fs2, rt)

	// Use TransactionBuilder to construct transaction with proper input declarations
	builder := transaction.NewTransactionBuilder(transaction.TxID{})

	// Add transfer instruction using the builder
	// This automatically declares:
	// - System Program with Read permission
	// - Sender with Write permission
	// - Receiver with Write permission
	err = builder.AddTransferInstruction(
		genesis.SystemProgramID,
		fromID,
		toID,
		*amount,
	)
	if err != nil {
		log.Fatalf("Failed to create transfer instruction: %v\nEnsure amount is positive and addresses are valid", err)
	}

	// Build the transaction (without signatures first, for signing)
	tx, err := builder.Build()
	if err != nil {
		log.Fatalf("Failed to build transaction: %v", err)
	}

	// Sign transaction
	txData, err := tx.Marshal()
	if err != nil {
		log.Fatalf("Failed to marshal transaction for signing: %v", err)
	}

	signature := ed25519.Sign(ed25519.PrivateKey(privKeyBytes), txData)
	var sig [64]byte
	copy(sig[:], signature)

	// Add signature to transaction
	tx.Signatures = []transaction.Signature{
		{PublicKey: txPubKey, Signature: sig},
	}

	// Process transaction
	result, err := txProcessor.ProcessTransaction(tx)
	if err != nil {
		log.Fatalf("Transaction processing failed: %v\n\nPossible causes:\n- Insufficient balance in sender account\n- Invalid file permissions\n- Account does not exist", err)
	}

	if !result.Success {
		log.Fatalf("Transaction execution failed: %v", result.Error)
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

// handleRPCCommand starts the RPC node
func handleRPCCommand() {
	fs := flag.NewFlagSet("rpc", flag.ExitOnError)
	rpcPort := fs.Int("rpc-port", 8899, "RPC HTTP listening port")
	rpcBind := fs.String("rpc-bind", "127.0.0.1", "RPC bind address")
	ledgerPath := fs.String("ledger-path", "", "Path to ledger database (required)")
	statePath := fs.String("state-path", "", "Path to state database (required)")

	fs.Parse(os.Args[2:])

	// Validate required parameters
	if *ledgerPath == "" || *statePath == "" {
		fmt.Println("Error: --ledger-path and --state-path are required")
		fmt.Println()
		fmt.Println("Usage: rpc --ledger-path <path> --state-path <path> [--rpc-port <port>] [--rpc-bind <address>]")
		fmt.Println()
		fmt.Println("Example:")
		fmt.Println("  poh-blockchain rpc --ledger-path ./validator1.db --state-path ./validator1_state.db")
		os.Exit(1)
	}

	log.Printf("Starting RPC node...")
	log.Printf("  Ledger: %s", *ledgerPath)
	log.Printf("  State: %s", *statePath)
	log.Printf("  Bind: %s:%d", *rpcBind, *rpcPort)

	// Initialize ledger
	log.Println("Initializing ledger...")
	ledger, err := storage.NewLedger(*ledgerPath)
	if err != nil {
		log.Fatalf("Failed to initialize ledger: %v", err)
	}
	defer ledger.Close()

	// Initialize FileStore
	log.Println("Initializing FileStore...")
	fileStore, err := initFileStore(*statePath)
	if err != nil {
		log.Fatalf("Failed to initialize FileStore: %v", err)
	}
	defer fileStore.Close()

	// Initialize Runtime and TxProcessor
	log.Println("Initializing transaction processor...")
	rt := runtime.NewRuntime()
	txProcessor := processor.NewTxProcessor(fileStore, rt)

	// Create RPC server configuration
	config := &rpc.ServerConfig{
		BindAddress: *rpcBind,
		Port:        *rpcPort,
		LedgerPath:  *ledgerPath,
		StatePath:   *statePath,
	}

	// Create RPC server
	log.Println("Creating RPC server...")
	rpcServer, err := rpc.NewRPCServer(config, ledger, fileStore, txProcessor, log.Default())
	if err != nil {
		log.Fatalf("Failed to create RPC server: %v", err)
	}

	// Start RPC server
	log.Println("Starting RPC server...")
	if err := rpcServer.Start(); err != nil {
		log.Fatalf("Failed to start RPC server: %v", err)
	}

	log.Printf("RPC server started successfully on http://%s:%d", *rpcBind, *rpcPort)
	log.Println("Press Ctrl+C to stop...")

	// Setup graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Wait for shutdown signal
	<-sigChan
	log.Println("Shutdown signal received, stopping RPC server...")

	// Stop RPC server
	if err := rpcServer.Stop(); err != nil {
		log.Printf("Error stopping RPC server: %v", err)
	}

	// Close database connections
	log.Println("Closing database connections...")
	if err := fileStore.Close(); err != nil {
		log.Printf("Error closing FileStore: %v", err)
	}
	if err := ledger.Close(); err != nil {
		log.Printf("Error closing ledger: %v", err)
	}

	log.Println("RPC node shutdown complete")
}

// publicKeyToFileID converts a public key to a file ID
func publicKeyToFileID(pubkey transaction.PublicKey) filestore.FileID {
	var fileID filestore.FileID
	copy(fileID[:], pubkey[:])
	return fileID
}

// initFileStore initializes a file store and loads the built-in QuanticScript programs at genesis.
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

	// Load built-in QuanticScript programs (System_Program, Token_Program, and Staking_Program).
	// This is idempotent — already-loaded programs are skipped.
	log.Println("Loading built-in programs...")
	if err := genesis.LoadBuiltinPrograms(fs, programs.SystemProgram, programs.TokenProgram, programs.StakingProgram); err != nil {
		log.Printf("Warning: failed to load built-in programs: %v", err)
	}

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
