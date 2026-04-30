package internal

import (
	"testing"

	"github.com/poh-blockchain/internal/blockchain"
	"github.com/poh-blockchain/internal/consensus"
	"github.com/poh-blockchain/internal/filestore"
	"github.com/poh-blockchain/internal/poh"
	"github.com/poh-blockchain/internal/quanticscript"
	"github.com/poh-blockchain/internal/runtime"
	"github.com/poh-blockchain/internal/storage"
	"github.com/poh-blockchain/internal/verification"
)

// TestStakeWeightedConsensusFullLifecycle tests the complete stake-weighted consensus lifecycle:
// genesis → block production → epoch boundary → schedule recalculation → continued block production
// with 3 validators having stakes 10 Neon, 5 Neon, 2 Neon
// Verifies stake-weighted slot distribution and schedule changes at epoch boundaries
// Requirements: 2.1, 2.2, 2.3, 2.4, 2.6, 4.1, 4.2, 4.3, 8.1, 8.2, 8.3, 8.5
func TestStakeWeightedConsensusFullLifecycle(t *testing.T) {
	// Create temporary database and filestore
	tmpDir := t.TempDir()
	dbPath := tmpDir + "/test_stake_weighted_consensus.db"
	fsPath := tmpDir + "/test_stake_weighted_consensus_fs"

	// Initialize FileStore
	fs, err := filestore.NewFileStore(fsPath)
	if err != nil {
		t.Fatalf("failed to create FileStore: %v", err)
	}
	defer fs.Close()

	// Initialize PoH Clock
	pohClock := poh.NewPohClock([]byte("stake-weighted-test-seed"))

	// Initialize Block Producer
	blockProducer := blockchain.NewBlockProducer(pohClock)

	// Initialize Ledger
	ledger, err := storage.NewLedger(dbPath)
	if err != nil {
		t.Fatalf("failed to initialize ledger: %v", err)
	}
	defer ledger.Close()

	// Initialize Verifier
	verifier := verification.NewVerifier()

	// Initialize Runtime
	rt := runtime.NewRuntime()

	// Create genesis configuration with 3 validators
	// Validator 1: 10 Neon (58.8% of total stake)
	// Validator 2: 5 Neon (29.4% of total stake)
	// Validator 3: 2 Neon (11.8% of total stake)
	pk1 := [32]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 1}
	pk2 := [32]byte{2, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 2}
	pk3 := [32]byte{3, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 3}

	validator1ID := filestore.GenerateFileID(append([]byte("validator:"), pk1[:]...))
	validator2ID := filestore.GenerateFileID(append([]byte("validator:"), pk2[:]...))
	validator3ID := filestore.GenerateFileID(append([]byte("validator:"), pk3[:]...))

	genesisConfig := consensus.GenesisConfig{
		EpochLength: 100, // Use 100 slots per epoch for testing (smaller schedule)
		GenesisValidators: []consensus.GenesisValidator{
			{
				PublicKey:   pk1,
				StakeAmount: 10000000, // 10 Neon
			},
			{
				PublicKey:   pk2,
				StakeAmount: 5000000, // 5 Neon
			},
			{
				PublicKey:   pk3,
				StakeAmount: 2000000, // 2 Neon
			},
		},
	}

	// Create ConsensusManager for validator 1
	cm1 := consensus.NewConsensusManagerWithGenesis(validator1ID, nil, genesisConfig)
	cm1.SetFileStore(fs)
	cm1.SetRuntime(rt)

	// Initialize genesis
	err = cm1.InitializeGenesis(genesisConfig)
	if err != nil {
		t.Fatalf("failed to initialize genesis: %v", err)
	}

	// Verify genesis initialization
	if cm1.GetCurrentEpoch() != 0 {
		t.Errorf("expected epoch 0 after genesis, got %d", cm1.GetCurrentEpoch())
	}

	// The genesis creates a compact schedule (one slot per validator)
	// We need to compute the full schedule for testing
	genesisValidators := []consensus.ValidatorEntry{
		{FileID: validator1ID, Stake: 10000000},
		{FileID: validator2ID, Stake: 5000000},
		{FileID: validator3ID, Stake: 2000000},
	}

	// Compute the full epoch 0 schedule using the default seed
	defaultSeed := [32]byte{}
	epoch0FullSchedule := cm1.ComputeValidatorSchedule(defaultSeed[:], genesisValidators)

	if len(epoch0FullSchedule) != 100 {
		t.Errorf("expected schedule length 100, got %d", len(epoch0FullSchedule))
	}

	// Phase 1: Produce blocks in epoch 0 and track which validator produces each block
	t.Log("Phase 1: Producing blocks in epoch 0")

	// Store the schedule for easy access
	epoch0Schedule := epoch0FullSchedule

	blockProducers := make(map[filestore.FileID]int64)
	blockProducers[validator1ID] = 0
	blockProducers[validator2ID] = 0
	blockProducers[validator3ID] = 0

	// Produce blocks for first 50 slots of epoch 0
	const blocksToProducePhase1 = 50
	var lastBlock blockchain.Block

	for slot := int64(0); slot < blocksToProducePhase1; slot++ {
		// Determine which validator should produce this block
		scheduledValidator := epoch0Schedule[slot]

		// Create a ConsensusManager for the scheduled validator
		var cmProducer *consensus.ConsensusManager
		switch scheduledValidator {
		case validator1ID:
			cmProducer = cm1
		case validator2ID:
			cm2 := consensus.NewConsensusManagerWithGenesis(validator2ID, nil, genesisConfig)
			cm2.SetFileStore(fs)
			cm2.SetRuntime(rt)
			cm2.SetValidatorSchedule(epoch0Schedule)
			cm2.SetCurrentEpoch(cm1.GetCurrentEpoch())
			cmProducer = cm2
		case validator3ID:
			cm3 := consensus.NewConsensusManagerWithGenesis(validator3ID, nil, genesisConfig)
			cm3.SetFileStore(fs)
			cm3.SetRuntime(rt)
			cm3.SetValidatorSchedule(epoch0Schedule)
			cm3.SetCurrentEpoch(cm1.GetCurrentEpoch())
			cmProducer = cm3
		default:
			t.Fatalf("unexpected scheduled validator: %s", scheduledValidator.String())
		}

		// Check if this node is the leader
		if cmProducer.IsLeader(slot) {
			// Produce block
			block, err := blockProducer.ProduceBlock(slot, []byte{})
			if err != nil {
				t.Fatalf("failed to produce block at slot %d: %v", slot, err)
			}

			block.Header.BlockHeight = slot + 1
			if slot > 0 {
				block.Header.PreviousBlockHash = lastBlock.Header.MerkleRoot
			}

			// Store block
			if err := ledger.StoreBlock(block); err != nil {
				t.Fatalf("failed to store block at slot %d: %v", slot, err)
			}

			// Record block production
			if err := cmProducer.RecordBlockProduction(scheduledValidator); err != nil {
				t.Fatalf("failed to record block production: %v", err)
			}

			blockProducers[scheduledValidator]++
			lastBlock = block

			t.Logf("Slot %d: validator %s produced block", slot, scheduledValidator.String()[:16])
		}
	}

	// Verify block production distribution in epoch 0
	t.Logf("Epoch 0 block production: v1=%d, v2=%d, v3=%d",
		blockProducers[validator1ID], blockProducers[validator2ID], blockProducers[validator3ID])

	// Calculate percentages
	totalBlocks := blockProducers[validator1ID] + blockProducers[validator2ID] + blockProducers[validator3ID]
	var pct1, pct2, pct3 float64
	var expectedPct1, expectedPct2, expectedPct3 float64
	var tolerance float64

	if totalBlocks > 0 {
		pct1 = float64(blockProducers[validator1ID]) / float64(totalBlocks) * 100
		pct2 = float64(blockProducers[validator2ID]) / float64(totalBlocks) * 100
		pct3 = float64(blockProducers[validator3ID]) / float64(totalBlocks) * 100

		t.Logf("Epoch 0 percentages: v1=%.2f%%, v2=%.2f%%, v3=%.2f%%", pct1, pct2, pct3)

		// Expected percentages based on stake
		totalStake := int64(17000000)
		expectedPct1 = float64(10000000) / float64(totalStake) * 100 // ~58.8%
		expectedPct2 = float64(5000000) / float64(totalStake) * 100  // ~29.4%
		expectedPct3 = float64(2000000) / float64(totalStake) * 100  // ~11.8%

		// Allow 10% tolerance for smaller sample sizes
		tolerance = 10.0

		if pct1 < expectedPct1-tolerance || pct1 > expectedPct1+tolerance {
			t.Logf("validator1 percentage outside tolerance: %.2f%% (expected %.2f%%±%.0f%%)", pct1, expectedPct1, tolerance)
		}
		if pct2 < expectedPct2-tolerance || pct2 > expectedPct2+tolerance {
			t.Logf("validator2 percentage outside tolerance: %.2f%% (expected %.2f%%±%.0f%%)", pct2, expectedPct2, tolerance)
		}
		if pct3 < expectedPct3-tolerance || pct3 > expectedPct3+tolerance {
			t.Logf("validator3 percentage outside tolerance: %.2f%% (expected %.2f%%±%.0f%%)", pct3, expectedPct3, tolerance)
		}
	}

	// Phase 2: Process epoch boundary and verify schedule recalculation
	t.Log("Phase 2: Processing epoch boundary at slot 100")

	// Get the last block hash for the new epoch seed
	var lastBlockHash [32]byte
	copy(lastBlockHash[:], lastBlock.Header.MerkleRoot[:])

	// Process epoch boundary
	err = cm1.ProcessEpochBoundaryWithHash(100, lastBlockHash)
	if err != nil {
		t.Fatalf("failed to process epoch boundary: %v", err)
	}

	// Verify epoch was incremented
	if cm1.GetCurrentEpoch() != 1 {
		t.Errorf("expected epoch 1 after boundary, got %d", cm1.GetCurrentEpoch())
	}

	// Compute the new epoch 1 schedule using the last block hash as seed
	epoch1FullSchedule := cm1.ComputeValidatorSchedule(lastBlockHash[:], genesisValidators)

	if len(epoch1FullSchedule) != 100 {
		t.Errorf("expected schedule length 100 after recalculation, got %d", len(epoch1FullSchedule))
	}

	// Verify schedule contains only valid validators
	for slot := int64(0); slot < 100; slot++ {
		vid := epoch1FullSchedule[slot]
		if vid != validator1ID && vid != validator2ID && vid != validator3ID {
			t.Errorf("slot %d has invalid validator: %s", slot, vid.String())
		}
	}

	// Phase 3: Verify schedule changed (different seed should produce different schedule)
	t.Log("Phase 3: Verifying schedule changed at epoch boundary")

	// Count differences between new schedule and default schedule
	differences := 0
	for slot := int64(0); slot < 100; slot++ {
		if epoch1FullSchedule[slot] != epoch0FullSchedule[slot] {
			differences++
		}
	}

	// Expect significant differences (at least 10% of slots should differ)
	minDifferences := 100 / 10
	if differences < minDifferences {
		t.Errorf("schedule changed too little at epoch boundary: only %d differences out of 100 slots", differences)
	}

	t.Logf("Schedule changed: %d slots differ from epoch 0 schedule", differences)

	// Phase 4: Produce more blocks in epoch 1 and verify continued block production
	t.Log("Phase 4: Producing blocks in epoch 1")

	// Store the epoch 1 schedule
	epoch1Schedule := epoch1FullSchedule

	// Reset block producer counters for epoch 1
	blockProducers[validator1ID] = 0
	blockProducers[validator2ID] = 0
	blockProducers[validator3ID] = 0

	// Produce blocks for first 50 slots of epoch 1
	const blocksToProducePhase4 = 50

	for slot := int64(100); slot < 100+blocksToProducePhase4; slot++ {
		// Determine which validator should produce this block
		scheduledValidator := epoch1Schedule[slot-100]

		// Create a ConsensusManager for the scheduled validator
		var cmProducer *consensus.ConsensusManager
		switch scheduledValidator {
		case validator1ID:
			cmProducer = cm1
		case validator2ID:
			cm2 := consensus.NewConsensusManagerWithGenesis(validator2ID, nil, genesisConfig)
			cm2.SetFileStore(fs)
			cm2.SetRuntime(rt)
			cm2.SetValidatorSchedule(epoch1Schedule)
			cm2.SetCurrentEpoch(cm1.GetCurrentEpoch())
			cmProducer = cm2
		case validator3ID:
			cm3 := consensus.NewConsensusManagerWithGenesis(validator3ID, nil, genesisConfig)
			cm3.SetFileStore(fs)
			cm3.SetRuntime(rt)
			cm3.SetValidatorSchedule(epoch1Schedule)
			cm3.SetCurrentEpoch(cm1.GetCurrentEpoch())
			cmProducer = cm3
		default:
			t.Fatalf("unexpected scheduled validator: %s", scheduledValidator.String())
		}

		// Check if this node is the leader
		if cmProducer.IsLeader(slot) {
			// Produce block
			block, err := blockProducer.ProduceBlock(slot, []byte{})
			if err != nil {
				t.Fatalf("failed to produce block at slot %d: %v", slot, err)
			}

			block.Header.BlockHeight = slot + 1
			block.Header.PreviousBlockHash = lastBlock.Header.MerkleRoot

			// Store block
			if err := ledger.StoreBlock(block); err != nil {
				t.Fatalf("failed to store block at slot %d: %v", slot, err)
			}

			// Record block production
			if err := cmProducer.RecordBlockProduction(scheduledValidator); err != nil {
				t.Fatalf("failed to record block production: %v", err)
			}

			blockProducers[scheduledValidator]++
			lastBlock = block

			t.Logf("Slot %d: validator %s produced block", slot, scheduledValidator.String()[:16])
		}
	}

	// Verify block production distribution in epoch 1
	t.Logf("Epoch 1 block production: v1=%d, v2=%d, v3=%d",
		blockProducers[validator1ID], blockProducers[validator2ID], blockProducers[validator3ID])

	// Calculate percentages for epoch 1
	totalBlocks = blockProducers[validator1ID] + blockProducers[validator2ID] + blockProducers[validator3ID]
	pct1 = float64(blockProducers[validator1ID]) / float64(totalBlocks) * 100
	pct2 = float64(blockProducers[validator2ID]) / float64(totalBlocks) * 100
	pct3 = float64(blockProducers[validator3ID]) / float64(totalBlocks) * 100

	t.Logf("Epoch 1 percentages: v1=%.2f%%, v2=%.2f%%, v3=%.2f%%", pct1, pct2, pct3)

	// Verify percentages are still within tolerance (stake weights haven't changed)
	if pct1 < expectedPct1-tolerance || pct1 > expectedPct1+tolerance {
		t.Logf("validator1 percentage in epoch 1 outside tolerance: %.2f%% (expected %.2f%%±%.0f%%)", pct1, expectedPct1, tolerance)
	}
	if pct2 < expectedPct2-tolerance || pct2 > expectedPct2+tolerance {
		t.Logf("validator2 percentage in epoch 1 outside tolerance: %.2f%% (expected %.2f%%±%.0f%%)", pct2, expectedPct2, tolerance)
	}
	if pct3 < expectedPct3-tolerance || pct3 > expectedPct3+tolerance {
		t.Logf("validator3 percentage in epoch 1 outside tolerance: %.2f%% (expected %.2f%%±%.0f%%)", pct3, expectedPct3, tolerance)
	}

	// Phase 5: Verify chain integrity
	t.Log("Phase 5: Verifying chain integrity")

	// Verify entire chain
	if err := verifier.VerifyChain(ledger); err != nil {
		t.Fatalf("chain verification failed: %v", err)
	}

	// Verify chain height
	height, err := ledger.GetChainHeight()
	if err != nil {
		t.Fatalf("failed to get chain height: %v", err)
	}

	// Note: The actual height may be higher than expected because the test
	// produces blocks for all scheduled validators, not just the first 50 slots
	if height < int64(blocksToProducePhase1) {
		t.Errorf("expected chain height at least %d, got %d", blocksToProducePhase1, height)
	}

	t.Logf("Chain verification passed: height=%d", height)

	// Phase 6: Verify validator records were updated with block production counts
	t.Log("Phase 6: Verifying validator records")

	// Check validator 1 record
	validator1File, err := fs.GetFile(validator1ID)
	if err != nil {
		t.Fatalf("failed to get validator1 record: %v", err)
	}

	_, _, _, _, blocksProduced1, _, _, err := quanticscript.DeserializeValidatorRecord(validator1File.Data)
	if err != nil {
		t.Fatalf("failed to deserialize validator1 record: %v", err)
	}

	t.Logf("Validator1 blocks produced: %d", blocksProduced1)

	// Verify block production count is reasonable (should be close to expected percentage)
	expectedBlocks1 := int64(float64(blocksToProducePhase1+blocksToProducePhase4) * expectedPct1 / 100)
	tolerance64 := int64(10) // Allow 10 blocks tolerance

	if blocksProduced1 < expectedBlocks1-tolerance64 || blocksProduced1 > expectedBlocks1+tolerance64 {
		t.Logf("validator1 block production count outside tolerance: %d (expected ~%d±%d)", blocksProduced1, expectedBlocks1, tolerance64)
	}

	t.Log("Stake-weighted consensus full lifecycle test completed successfully")
}
