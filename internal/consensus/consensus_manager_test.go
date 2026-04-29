package consensus

import (
	"testing"
	"time"

	"github.com/poh-blockchain/internal/filestore"
	"github.com/poh-blockchain/internal/genesis"
	"github.com/poh-blockchain/internal/quanticscript"
	"github.com/poh-blockchain/internal/runtime"
)

// TestNewConsensusManagerWithGenesisConfig tests that ConsensusManager can be created with a GenesisConfig
func TestNewConsensusManagerWithGenesisConfig(t *testing.T) {
	config := GenesisConfig{
		EpochLength: 432000,
		GenesisValidators: []GenesisValidator{
			{
				PublicKey:   [32]byte{1, 2, 3, 4},
				StakeAmount: 1000000,
			},
			{
				PublicKey:   [32]byte{5, 6, 7, 8},
				StakeAmount: 2000000,
			},
		},
	}

	// Create with a test validator ID
	pk := [32]byte{1, 2, 3, 4}
	validatorID := filestore.GenerateFileID(append([]byte("validator:"), pk[:]...))

	cm := NewConsensusManagerWithGenesis(validatorID, nil, config)

	if cm == nil {
		t.Fatal("NewConsensusManagerWithGenesis returned nil")
	}

	if cm.epochLength != config.EpochLength {
		t.Errorf("expected epochLength %d, got %d", config.EpochLength, cm.epochLength)
	}

	if len(cm.genesisValidators) != len(config.GenesisValidators) {
		t.Errorf("expected %d genesis validators, got %d", len(config.GenesisValidators), len(cm.genesisValidators))
	}

	if cm.currentEpoch != 0 {
		t.Errorf("expected currentEpoch 0, got %d", cm.currentEpoch)
	}
}

// TestGetCurrentEpoch tests epoch calculation from slot
func TestGetCurrentEpoch(t *testing.T) {
	config := GenesisConfig{
		EpochLength: 100,
		GenesisValidators: []GenesisValidator{
			{
				PublicKey:   [32]byte{1},
				StakeAmount: 1000000,
			},
		},
	}

	pk := [32]byte{1}
	validatorID := filestore.GenerateFileID(append([]byte("validator:"), pk[:]...))
	cm := NewConsensusManagerWithGenesis(validatorID, nil, config)

	// At slot 0, should be epoch 0
	if cm.GetCurrentEpoch() != 0 {
		t.Errorf("expected epoch 0 at slot 0, got %d", cm.GetCurrentEpoch())
	}

	// Manually set currentSlot for testing
	cm.currentEpoch = 5
	if cm.GetCurrentEpoch() != 5 {
		t.Errorf("expected epoch 5, got %d", cm.GetCurrentEpoch())
	}
}

// TestIsEpochBoundary tests epoch boundary detection
func TestIsEpochBoundary(t *testing.T) {
	config := GenesisConfig{
		EpochLength: 100,
		GenesisValidators: []GenesisValidator{
			{
				PublicKey:   [32]byte{1},
				StakeAmount: 1000000,
			},
		},
	}

	pk := [32]byte{1}
	validatorID := filestore.GenerateFileID(append([]byte("validator:"), pk[:]...))
	cm := NewConsensusManagerWithGenesis(validatorID, nil, config)

	tests := []struct {
		slot     int64
		expected bool
	}{
		{0, true},    // slot 0 is epoch boundary
		{99, false},  // slot 99 is not
		{100, true},  // slot 100 is epoch boundary
		{199, false}, // slot 199 is not
		{200, true},  // slot 200 is epoch boundary
	}

	for _, tt := range tests {
		result := cm.IsEpochBoundary(tt.slot)
		if result != tt.expected {
			t.Errorf("IsEpochBoundary(%d) = %v, expected %v", tt.slot, result, tt.expected)
		}
	}
}

// TestGetScheduledValidator tests validator schedule lookup
func TestGetScheduledValidator(t *testing.T) {
	config := GenesisConfig{
		EpochLength: 10,
		GenesisValidators: []GenesisValidator{
			{
				PublicKey:   [32]byte{1},
				StakeAmount: 1000000,
			},
			{
				PublicKey:   [32]byte{2},
				StakeAmount: 2000000,
			},
		},
	}

	pk := [32]byte{1}
	validatorID := filestore.GenerateFileID(append([]byte("validator:"), pk[:]...))
	cm := NewConsensusManagerWithGenesis(validatorID, nil, config)

	// Initialize validator schedule
	pk1 := [32]byte{1}
	pk2 := [32]byte{2}
	validator1ID := filestore.GenerateFileID(append([]byte("validator:"), pk1[:]...))
	validator2ID := filestore.GenerateFileID(append([]byte("validator:"), pk2[:]...))

	cm.validatorSchedule = []filestore.FileID{validator1ID, validator2ID, validator1ID, validator2ID}

	// Test slot 0 -> validator1
	result := cm.GetScheduledValidator(0)
	if result != validator1ID {
		t.Errorf("expected validator1 at slot 0, got %s", result.String())
	}

	// Test slot 1 -> validator2
	result = cm.GetScheduledValidator(1)
	if result != validator2ID {
		t.Errorf("expected validator2 at slot 1, got %s", result.String())
	}
}

// TestRecordMissedBlock tests missed block recording
func TestRecordMissedBlock(t *testing.T) {
	config := GenesisConfig{
		EpochLength: 100,
		GenesisValidators: []GenesisValidator{
			{
				PublicKey:   [32]byte{1},
				StakeAmount: 1000000,
			},
		},
	}

	pk := [32]byte{1}
	validatorID := filestore.GenerateFileID(append([]byte("validator:"), pk[:]...))
	cm := NewConsensusManagerWithGenesis(validatorID, nil, config)

	// Initialize with a mock validator ID
	pk2 := [32]byte{1}
	validatorID2 := filestore.GenerateFileID(append([]byte("validator:"), pk2[:]...))

	// Test without fileStore - should fail
	err := cm.RecordMissedBlock(0, validatorID2)
	if err == nil {
		t.Errorf("RecordMissedBlock should fail without fileStore")
	}
}

// TestWaitForSlotStart tests slot timing
func TestWaitForSlotStart(t *testing.T) {
	config := GenesisConfig{
		EpochLength: 100,
		GenesisValidators: []GenesisValidator{
			{
				PublicKey:   [32]byte{1},
				StakeAmount: 1000000,
			},
		},
	}

	pk := [32]byte{1}
	validatorID := filestore.GenerateFileID(append([]byte("validator:"), pk[:]...))
	cm := NewConsensusManagerWithGenesis(validatorID, nil, config)

	// Set genesis timestamp to now
	cm.genesisTimestamp = time.Now()

	// Get current slot
	currentSlot := cm.GetCurrentSlot()

	// Wait for next slot
	nextSlot := currentSlot + 1
	start := time.Now()
	cm.WaitForSlotStart(nextSlot)
	elapsed := time.Since(start)

	// Should have waited approximately 400ms (slot duration)
	// Allow some tolerance for test execution time
	if elapsed < 350*time.Millisecond || elapsed > 500*time.Millisecond {
		t.Logf("WaitForSlotStart took %v (expected ~400ms)", elapsed)
	}
}

// TestComputeValidatorSchedule tests deterministic schedule computation
func TestComputeValidatorSchedule(t *testing.T) {
	config := GenesisConfig{
		EpochLength: 10,
		GenesisValidators: []GenesisValidator{
			{
				PublicKey:   [32]byte{1},
				StakeAmount: 1000000, // 1 Neon
			},
			{
				PublicKey:   [32]byte{2},
				StakeAmount: 2000000, // 2 Neon (2x weight)
			},
		},
	}

	pk := [32]byte{1}
	validatorID := filestore.GenerateFileID(append([]byte("validator:"), pk[:]...))
	cm := NewConsensusManagerWithGenesis(validatorID, nil, config)

	// Create validator entries
	pk1 := [32]byte{1}
	pk2 := [32]byte{2}
	validator1ID := filestore.GenerateFileID(append([]byte("validator:"), pk1[:]...))
	validator2ID := filestore.GenerateFileID(append([]byte("validator:"), pk2[:]...))

	validators := []ValidatorEntry{
		{
			FileID: validator1ID,
			Stake:  1000000,
		},
		{
			FileID: validator2ID,
			Stake:  2000000,
		},
	}

	// Create a seed
	seed := [32]byte{1, 2, 3, 4}

	// Compute schedule
	schedule := cm.ComputeValidatorSchedule(seed[:], validators)

	// Verify schedule length matches epoch length
	if int64(len(schedule)) != cm.epochLength {
		t.Errorf("expected schedule length %d, got %d", cm.epochLength, len(schedule))
	}

	// Verify schedule contains only valid validators
	for _, validatorID := range schedule {
		if validatorID != validator1ID && validatorID != validator2ID {
			t.Errorf("schedule contains invalid validator: %s", validatorID.String())
		}
	}

	// Verify determinism: same seed should produce same schedule
	schedule2 := cm.ComputeValidatorSchedule(seed[:], validators)
	if len(schedule) != len(schedule2) {
		t.Errorf("schedule length mismatch on second call")
	}
	for i := range schedule {
		if schedule[i] != schedule2[i] {
			t.Errorf("schedule mismatch at slot %d", i)
		}
	}

	// Verify different seed produces different schedule
	seed2 := [32]byte{5, 6, 7, 8}
	schedule3 := cm.ComputeValidatorSchedule(seed2[:], validators)
	different := false
	for i := range schedule {
		if schedule[i] != schedule3[i] {
			different = true
			break
		}
	}
	if !different {
		t.Errorf("different seeds should produce different schedules")
	}

	// Verify stake-weighted distribution: validator2 should appear ~2x more often
	count1 := int64(0)
	count2 := int64(0)
	for _, validatorID := range schedule {
		if validatorID == validator1ID {
			count1++
		} else if validatorID == validator2ID {
			count2++
		}
	}

	// Expect roughly 1:2 ratio (with some tolerance)
	ratio := float64(count2) / float64(count1)
	if ratio < 1.5 || ratio > 2.5 {
		t.Logf("stake-weighted ratio: %.2f (expected ~2.0), count1=%d, count2=%d", ratio, count1, count2)
	}
}

// TestIsLeaderWithSchedule tests IsLeader with stake-weighted schedule
func TestIsLeaderWithSchedule(t *testing.T) {
	config := GenesisConfig{
		EpochLength: 10,
		GenesisValidators: []GenesisValidator{
			{
				PublicKey:   [32]byte{1},
				StakeAmount: 1000000,
			},
		},
	}

	// Create validator IDs
	pk := [32]byte{1}
	validatorID := filestore.GenerateFileID(append([]byte("validator:"), pk[:]...))

	pk2 := [32]byte{2}
	otherValidatorID := filestore.GenerateFileID(append([]byte("validator:"), pk2[:]...))

	// Create a node with validator identity
	cmValidator := NewConsensusManagerWithGenesis(validatorID, nil, config)

	// Create an observer node (no validator identity)
	cmObserver := NewConsensusManagerWithGenesis(filestore.FileID{}, nil, config)

	// Set up a simple schedule
	cmValidator.validatorSchedule = []filestore.FileID{validatorID, validatorID}
	cmObserver.validatorSchedule = []filestore.FileID{validatorID, validatorID}

	// For a validator node, IsLeader should return true if it's scheduled
	if cmValidator.IsLeader(0) != true {
		t.Errorf("validator node should be leader for scheduled slot")
	}

	// For an observer node, IsLeader should always return false
	if cmObserver.IsLeader(0) != false {
		t.Errorf("observer node should not be leader")
	}

	// Test with different validator scheduled
	cmValidator.validatorSchedule = []filestore.FileID{otherValidatorID}
	if cmValidator.IsLeader(0) != false {
		t.Errorf("validator node should not be leader when different validator is scheduled")
	}
}

// TestProcessEpochBoundary tests epoch boundary processing
func TestProcessEpochBoundary(t *testing.T) {
	config := GenesisConfig{
		EpochLength: 100,
		GenesisValidators: []GenesisValidator{
			{
				PublicKey:   [32]byte{1},
				StakeAmount: 1000000,
			},
		},
	}

	pk := [32]byte{1}
	validatorID := filestore.GenerateFileID(append([]byte("validator:"), pk[:]...))
	cm := NewConsensusManagerWithGenesis(validatorID, nil, config)

	// Test epoch boundary detection
	if !cm.IsEpochBoundary(0) {
		t.Errorf("slot 0 should be epoch boundary")
	}

	if !cm.IsEpochBoundary(100) {
		t.Errorf("slot 100 should be epoch boundary")
	}

	if cm.IsEpochBoundary(50) {
		t.Errorf("slot 50 should not be epoch boundary")
	}
}

// TestStakeWeightedIsLeaderWhenScheduled tests IsLeader returns true when local validator is scheduled
func TestStakeWeightedIsLeaderWhenScheduled(t *testing.T) {
	config := GenesisConfig{
		EpochLength: 10,
		GenesisValidators: []GenesisValidator{
			{
				PublicKey:   [32]byte{1},
				StakeAmount: 1000000,
			},
		},
	}

	// Create validator ID
	pk := [32]byte{1}
	localValidatorID := filestore.GenerateFileID(append([]byte("validator:"), pk[:]...))

	// Create ConsensusManager with local validator identity
	cm := NewConsensusManagerWithGenesis(localValidatorID, nil, config)

	// Set up schedule where local validator is scheduled for slot 0
	cm.validatorSchedule = []filestore.FileID{localValidatorID}

	// IsLeader should return true for slot 0
	if !cm.IsLeader(0) {
		t.Errorf("IsLeader(0) should return true when local validator is scheduled")
	}
}

// TestStakeWeightedIsLeaderWhenNotScheduled tests IsLeader returns false when local validator is not scheduled
func TestStakeWeightedIsLeaderWhenNotScheduled(t *testing.T) {
	config := GenesisConfig{
		EpochLength: 10,
		GenesisValidators: []GenesisValidator{
			{
				PublicKey:   [32]byte{1},
				StakeAmount: 1000000,
			},
		},
	}

	// Create two different validator IDs
	pk1 := [32]byte{1}
	pk2 := [32]byte{2}
	localValidatorID := filestore.GenerateFileID(append([]byte("validator:"), pk1[:]...))
	otherValidatorID := filestore.GenerateFileID(append([]byte("validator:"), pk2[:]...))

	// Create ConsensusManager with local validator identity
	cm := NewConsensusManagerWithGenesis(localValidatorID, nil, config)

	// Set up schedule where a different validator is scheduled for slot 0
	cm.validatorSchedule = []filestore.FileID{otherValidatorID}

	// IsLeader should return false for slot 0
	if cm.IsLeader(0) {
		t.Errorf("IsLeader(0) should return false when local validator is not scheduled")
	}
}

// TestStakeWeightedIsLeaderObserverMode tests IsLeader returns false in observer mode
func TestStakeWeightedIsLeaderObserverMode(t *testing.T) {
	config := GenesisConfig{
		EpochLength: 10,
		GenesisValidators: []GenesisValidator{
			{
				PublicKey:   [32]byte{1},
				StakeAmount: 1000000,
			},
		},
	}

	// Create ConsensusManager with zero validator ID (observer mode)
	zeroValidatorID := filestore.FileID{}
	cm := NewConsensusManagerWithGenesis(zeroValidatorID, nil, config)

	// Set up schedule with some validator
	pk := [32]byte{1}
	someValidatorID := filestore.GenerateFileID(append([]byte("validator:"), pk[:]...))
	cm.validatorSchedule = []filestore.FileID{someValidatorID}

	// IsLeader should always return false in observer mode
	if cm.IsLeader(0) {
		t.Errorf("IsLeader(0) should return false in observer mode")
	}
}

// TestScheduleComputationDeterminism tests that same inputs produce identical schedules
func TestScheduleComputationDeterminism(t *testing.T) {
	config := GenesisConfig{
		EpochLength: 1000, // Use larger epoch for better statistical testing
		GenesisValidators: []GenesisValidator{
			{
				PublicKey:   [32]byte{1},
				StakeAmount: 1000000,
			},
		},
	}

	pk := [32]byte{1}
	validatorID := filestore.GenerateFileID(append([]byte("validator:"), pk[:]...))
	cm := NewConsensusManagerWithGenesis(validatorID, nil, config)

	// Create validator entries
	pk1 := [32]byte{1}
	pk2 := [32]byte{2}
	pk3 := [32]byte{3}
	validator1ID := filestore.GenerateFileID(append([]byte("validator:"), pk1[:]...))
	validator2ID := filestore.GenerateFileID(append([]byte("validator:"), pk2[:]...))
	validator3ID := filestore.GenerateFileID(append([]byte("validator:"), pk3[:]...))

	validators := []ValidatorEntry{
		{FileID: validator1ID, Stake: 1000000},
		{FileID: validator2ID, Stake: 2000000},
		{FileID: validator3ID, Stake: 3000000},
	}

	// Create a seed
	seed := [32]byte{1, 2, 3, 4, 5, 6, 7, 8}

	// Compute schedule multiple times with same inputs
	schedule1 := cm.ComputeValidatorSchedule(seed[:], validators)
	schedule2 := cm.ComputeValidatorSchedule(seed[:], validators)
	schedule3 := cm.ComputeValidatorSchedule(seed[:], validators)

	// All schedules should be identical
	if len(schedule1) != len(schedule2) || len(schedule1) != len(schedule3) {
		t.Fatalf("schedule lengths differ: %d, %d, %d", len(schedule1), len(schedule2), len(schedule3))
	}

	for i := range schedule1 {
		if schedule1[i] != schedule2[i] {
			t.Errorf("schedule1 and schedule2 differ at slot %d", i)
		}
		if schedule1[i] != schedule3[i] {
			t.Errorf("schedule1 and schedule3 differ at slot %d", i)
		}
	}
}

// TestScheduleComputationDifferentSeeds tests that different seeds produce different schedules
func TestScheduleComputationDifferentSeeds(t *testing.T) {
	config := GenesisConfig{
		EpochLength: 1000,
		GenesisValidators: []GenesisValidator{
			{
				PublicKey:   [32]byte{1},
				StakeAmount: 1000000,
			},
		},
	}

	pk := [32]byte{1}
	validatorID := filestore.GenerateFileID(append([]byte("validator:"), pk[:]...))
	cm := NewConsensusManagerWithGenesis(validatorID, nil, config)

	// Create validator entries
	pk1 := [32]byte{1}
	pk2 := [32]byte{2}
	validator1ID := filestore.GenerateFileID(append([]byte("validator:"), pk1[:]...))
	validator2ID := filestore.GenerateFileID(append([]byte("validator:"), pk2[:]...))

	validators := []ValidatorEntry{
		{FileID: validator1ID, Stake: 1000000},
		{FileID: validator2ID, Stake: 2000000},
	}

	// Create different seeds
	seed1 := [32]byte{1, 2, 3, 4, 5, 6, 7, 8}
	seed2 := [32]byte{9, 10, 11, 12, 13, 14, 15, 16}
	seed3 := [32]byte{17, 18, 19, 20, 21, 22, 23, 24}

	// Compute schedules with different seeds
	schedule1 := cm.ComputeValidatorSchedule(seed1[:], validators)
	schedule2 := cm.ComputeValidatorSchedule(seed2[:], validators)
	schedule3 := cm.ComputeValidatorSchedule(seed3[:], validators)

	// Schedules should be different
	differences12 := 0
	differences13 := 0
	differences23 := 0

	for i := range schedule1 {
		if schedule1[i] != schedule2[i] {
			differences12++
		}
		if schedule1[i] != schedule3[i] {
			differences13++
		}
		if schedule2[i] != schedule3[i] {
			differences23++
		}
	}

	// Expect significant differences (at least 10% of slots should differ)
	minDifferences := len(schedule1) / 10

	if differences12 < minDifferences {
		t.Errorf("schedule1 and schedule2 are too similar: only %d differences out of %d slots", differences12, len(schedule1))
	}
	if differences13 < minDifferences {
		t.Errorf("schedule1 and schedule3 are too similar: only %d differences out of %d slots", differences13, len(schedule1))
	}
	if differences23 < minDifferences {
		t.Errorf("schedule2 and schedule3 are too similar: only %d differences out of %d slots", differences23, len(schedule1))
	}
}

// TestScheduleStakeWeightedDistribution tests that validators get slots proportional to their stake
func TestScheduleStakeWeightedDistribution(t *testing.T) {
	config := GenesisConfig{
		EpochLength: 10000, // Large epoch for statistical accuracy
		GenesisValidators: []GenesisValidator{
			{
				PublicKey:   [32]byte{1},
				StakeAmount: 1000000,
			},
		},
	}

	pk := [32]byte{1}
	validatorID := filestore.GenerateFileID(append([]byte("validator:"), pk[:]...))
	cm := NewConsensusManagerWithGenesis(validatorID, nil, config)

	// Create validator entries with 1:2 stake ratio
	pk1 := [32]byte{1}
	pk2 := [32]byte{2}
	validator1ID := filestore.GenerateFileID(append([]byte("validator:"), pk1[:]...))
	validator2ID := filestore.GenerateFileID(append([]byte("validator:"), pk2[:]...))

	validators := []ValidatorEntry{
		{FileID: validator1ID, Stake: 1000000}, // 1 Neon
		{FileID: validator2ID, Stake: 2000000}, // 2 Neon (2x stake)
	}

	seed := [32]byte{1, 2, 3, 4, 5, 6, 7, 8}

	// Compute schedule
	schedule := cm.ComputeValidatorSchedule(seed[:], validators)

	// Count slots for each validator
	count1 := int64(0)
	count2 := int64(0)

	for _, validatorID := range schedule {
		if validatorID == validator1ID {
			count1++
		} else if validatorID == validator2ID {
			count2++
		}
	}

	// Calculate actual ratio
	ratio := float64(count2) / float64(count1)

	// Expected ratio is 2.0 (validator2 has 2x stake)
	// Allow 5% tolerance as specified in requirements
	expectedRatio := 2.0
	tolerance := 0.05
	minRatio := expectedRatio * (1.0 - tolerance)
	maxRatio := expectedRatio * (1.0 + tolerance)

	if ratio < minRatio || ratio > maxRatio {
		t.Errorf("stake-weighted distribution outside tolerance: ratio=%.4f, expected=%.2f±%.0f%%, count1=%d, count2=%d",
			ratio, expectedRatio, tolerance*100, count1, count2)
	}

	t.Logf("stake-weighted distribution: ratio=%.4f (expected=%.2f), count1=%d, count2=%d", ratio, expectedRatio, count1, count2)
}

// TestScheduleStakeWeightedDistributionThreeValidators tests distribution with three validators
func TestScheduleStakeWeightedDistributionThreeValidators(t *testing.T) {
	config := GenesisConfig{
		EpochLength: 10000,
		GenesisValidators: []GenesisValidator{
			{
				PublicKey:   [32]byte{1},
				StakeAmount: 1000000,
			},
		},
	}

	pk := [32]byte{1}
	validatorID := filestore.GenerateFileID(append([]byte("validator:"), pk[:]...))
	cm := NewConsensusManagerWithGenesis(validatorID, nil, config)

	// Create validator entries with 10:5:2 stake ratio
	pk1 := [32]byte{1}
	pk2 := [32]byte{2}
	pk3 := [32]byte{3}
	validator1ID := filestore.GenerateFileID(append([]byte("validator:"), pk1[:]...))
	validator2ID := filestore.GenerateFileID(append([]byte("validator:"), pk2[:]...))
	validator3ID := filestore.GenerateFileID(append([]byte("validator:"), pk3[:]...))

	validators := []ValidatorEntry{
		{FileID: validator1ID, Stake: 10000000}, // 10 Neon
		{FileID: validator2ID, Stake: 5000000},  // 5 Neon
		{FileID: validator3ID, Stake: 2000000},  // 2 Neon
	}

	totalStake := int64(17000000)
	seed := [32]byte{1, 2, 3, 4, 5, 6, 7, 8}

	// Compute schedule
	schedule := cm.ComputeValidatorSchedule(seed[:], validators)

	// Count slots for each validator
	count1 := int64(0)
	count2 := int64(0)
	count3 := int64(0)

	for _, validatorID := range schedule {
		if validatorID == validator1ID {
			count1++
		} else if validatorID == validator2ID {
			count2++
		} else if validatorID == validator3ID {
			count3++
		}
	}

	// Calculate percentages
	pct1 := float64(count1) / float64(len(schedule)) * 100
	pct2 := float64(count2) / float64(len(schedule)) * 100
	pct3 := float64(count3) / float64(len(schedule)) * 100

	// Expected percentages based on stake
	expectedPct1 := float64(10000000) / float64(totalStake) * 100 // ~58.8%
	expectedPct2 := float64(5000000) / float64(totalStake) * 100  // ~29.4%
	expectedPct3 := float64(2000000) / float64(totalStake) * 100  // ~11.8%

	// Allow 5% tolerance
	tolerance := 5.0

	if pct1 < expectedPct1-tolerance || pct1 > expectedPct1+tolerance {
		t.Errorf("validator1 percentage outside tolerance: %.2f%% (expected %.2f%%±%.0f%%)", pct1, expectedPct1, tolerance)
	}
	if pct2 < expectedPct2-tolerance || pct2 > expectedPct2+tolerance {
		t.Errorf("validator2 percentage outside tolerance: %.2f%% (expected %.2f%%±%.0f%%)", pct2, expectedPct2, tolerance)
	}
	if pct3 < expectedPct3-tolerance || pct3 > expectedPct3+tolerance {
		t.Errorf("validator3 percentage outside tolerance: %.2f%% (expected %.2f%%±%.0f%%)", pct3, expectedPct3, tolerance)
	}

	t.Logf("stake-weighted distribution: v1=%.2f%% (expected %.2f%%), v2=%.2f%% (expected %.2f%%), v3=%.2f%% (expected %.2f%%)",
		pct1, expectedPct1, pct2, expectedPct2, pct3, expectedPct3)
}

// TestScheduleEdgeCaseSingleValidator tests schedule computation with a single validator
func TestScheduleEdgeCaseSingleValidator(t *testing.T) {
	config := GenesisConfig{
		EpochLength: 100,
		GenesisValidators: []GenesisValidator{
			{
				PublicKey:   [32]byte{1},
				StakeAmount: 1000000,
			},
		},
	}

	pk := [32]byte{1}
	validatorID := filestore.GenerateFileID(append([]byte("validator:"), pk[:]...))
	cm := NewConsensusManagerWithGenesis(validatorID, nil, config)

	// Create single validator entry
	pk1 := [32]byte{1}
	validator1ID := filestore.GenerateFileID(append([]byte("validator:"), pk1[:]...))

	validators := []ValidatorEntry{
		{FileID: validator1ID, Stake: 1000000},
	}

	seed := [32]byte{1, 2, 3, 4, 5, 6, 7, 8}

	// Compute schedule
	schedule := cm.ComputeValidatorSchedule(seed[:], validators)

	// Verify schedule length
	if int64(len(schedule)) != cm.epochLength {
		t.Errorf("schedule length mismatch: got %d, expected %d", len(schedule), cm.epochLength)
	}

	// All slots should be assigned to the single validator
	for i, validatorID := range schedule {
		if validatorID != validator1ID {
			t.Errorf("slot %d assigned to wrong validator: got %s, expected %s", i, validatorID.String(), validator1ID.String())
		}
	}
}

// TestScheduleEdgeCaseZeroTotalStake tests schedule computation with zero total stake
func TestScheduleEdgeCaseZeroTotalStake(t *testing.T) {
	config := GenesisConfig{
		EpochLength: 100,
		GenesisValidators: []GenesisValidator{
			{
				PublicKey:   [32]byte{1},
				StakeAmount: 1000000,
			},
		},
	}

	pk := [32]byte{1}
	validatorID := filestore.GenerateFileID(append([]byte("validator:"), pk[:]...))
	cm := NewConsensusManagerWithGenesis(validatorID, nil, config)

	// Create validator entries with zero stake
	pk1 := [32]byte{1}
	validator1ID := filestore.GenerateFileID(append([]byte("validator:"), pk1[:]...))

	validators := []ValidatorEntry{
		{FileID: validator1ID, Stake: 0},
	}

	seed := [32]byte{1, 2, 3, 4, 5, 6, 7, 8}

	// Compute schedule
	schedule := cm.ComputeValidatorSchedule(seed[:], validators)

	// Schedule should be empty or all zero FileIDs
	if int64(len(schedule)) != cm.epochLength {
		t.Errorf("schedule length mismatch: got %d, expected %d", len(schedule), cm.epochLength)
	}

	// All slots should be zero FileID (no validator assigned)
	for i, validatorID := range schedule {
		if validatorID != (filestore.FileID{}) {
			t.Errorf("slot %d should have zero FileID with zero total stake, got %s", i, validatorID.String())
		}
	}
}

// TestScheduleEdgeCaseEmptyValidatorList tests schedule computation with empty validator list
func TestScheduleEdgeCaseEmptyValidatorList(t *testing.T) {
	config := GenesisConfig{
		EpochLength: 100,
		GenesisValidators: []GenesisValidator{
			{
				PublicKey:   [32]byte{1},
				StakeAmount: 1000000,
			},
		},
	}

	pk := [32]byte{1}
	validatorID := filestore.GenerateFileID(append([]byte("validator:"), pk[:]...))
	cm := NewConsensusManagerWithGenesis(validatorID, nil, config)

	// Empty validator list
	validators := []ValidatorEntry{}

	seed := [32]byte{1, 2, 3, 4, 5, 6, 7, 8}

	// Compute schedule
	schedule := cm.ComputeValidatorSchedule(seed[:], validators)

	// Schedule should be empty
	if len(schedule) != 0 {
		t.Errorf("schedule should be empty for empty validator list, got length %d", len(schedule))
	}
}

// TestEnumerateActiveValidators tests enumerating active validators from FileStore
func TestEnumerateActiveValidators(t *testing.T) {
	// Create a temporary FileStore
	tmpDir := t.TempDir()
	fs, err := filestore.NewFileStore(tmpDir)
	if err != nil {
		t.Fatalf("failed to create FileStore: %v", err)
	}
	defer fs.Close()

	// Create ConsensusManager
	config := GenesisConfig{
		EpochLength: 100,
		GenesisValidators: []GenesisValidator{
			{
				PublicKey:   [32]byte{1},
				StakeAmount: 1000000,
			},
		},
	}

	pk := [32]byte{1}
	validatorID := filestore.GenerateFileID(append([]byte("validator:"), pk[:]...))
	cm := NewConsensusManagerWithGenesis(validatorID, nil, config)
	cm.SetFileStore(fs)

	// Create test validator records
	pk1 := [32]byte{1, 2, 3}
	pk2 := [32]byte{4, 5, 6}
	pk3 := [32]byte{7, 8, 9}

	validator1ID := filestore.GenerateFileID(append([]byte("validator:"), pk1[:]...))
	validator2ID := filestore.GenerateFileID(append([]byte("validator:"), pk2[:]...))
	validator3ID := filestore.GenerateFileID(append([]byte("validator:"), pk3[:]...))

	// Create validator 1: active with 2 Neon stake
	validator1Data, err := quanticscript.SerializeValidatorRecord(
		pk1[:],
		0,       // commission
		2000000, // totalStake (2 Neon)
		1,       // status (active)
		0,       // blocksProduced
		0,       // missedBlocks
		0,       // slashedThisEpoch
	)
	if err != nil {
		t.Fatalf("failed to serialize validator1: %v", err)
	}

	storageCost := filestore.CalculateStorageCost(int64(len(validator1Data)))
	validator1File := &filestore.File{
		ID:        validator1ID,
		Balance:   storageCost + 1000,
		TxManager: filestore.FileID{},
		Data:      validator1Data,
	}
	_, err = fs.CreateFile(validator1File)
	if err != nil {
		t.Fatalf("failed to create validator1 file: %v", err)
	}

	// Create validator 2: active with 5 Neon stake
	validator2Data, err := quanticscript.SerializeValidatorRecord(
		pk2[:],
		0,       // commission
		5000000, // totalStake (5 Neon)
		1,       // status (active)
		0,       // blocksProduced
		0,       // missedBlocks
		0,       // slashedThisEpoch
	)
	if err != nil {
		t.Fatalf("failed to serialize validator2: %v", err)
	}

	storageCost = filestore.CalculateStorageCost(int64(len(validator2Data)))
	validator2File := &filestore.File{
		ID:        validator2ID,
		Balance:   storageCost + 1000,
		TxManager: filestore.FileID{},
		Data:      validator2Data,
	}
	_, err = fs.CreateFile(validator2File)
	if err != nil {
		t.Fatalf("failed to create validator2 file: %v", err)
	}

	// Create validator 3: inactive with 3 Neon stake (should be filtered out)
	validator3Data, err := quanticscript.SerializeValidatorRecord(
		pk3[:],
		0,       // commission
		3000000, // totalStake (3 Neon)
		0,       // status (inactive)
		0,       // blocksProduced
		0,       // missedBlocks
		0,       // slashedThisEpoch
	)
	if err != nil {
		t.Fatalf("failed to serialize validator3: %v", err)
	}

	storageCost = filestore.CalculateStorageCost(int64(len(validator3Data)))
	validator3File := &filestore.File{
		ID:        validator3ID,
		Balance:   storageCost + 1000,
		TxManager: filestore.FileID{},
		Data:      validator3Data,
	}
	_, err = fs.CreateFile(validator3File)
	if err != nil {
		t.Fatalf("failed to create validator3 file: %v", err)
	}

	// Create validator 4: active but with insufficient stake (should be filtered out)
	pk4 := [32]byte{10, 11, 12}
	validator4ID := filestore.GenerateFileID(append([]byte("validator:"), pk4[:]...))
	validator4Data, err := quanticscript.SerializeValidatorRecord(
		pk4[:],
		0,      // commission
		500000, // totalStake (0.5 Neon - below minimum)
		1,      // status (active)
		0,      // blocksProduced
		0,      // missedBlocks
		0,      // slashedThisEpoch
	)
	if err != nil {
		t.Fatalf("failed to serialize validator4: %v", err)
	}

	storageCost = filestore.CalculateStorageCost(int64(len(validator4Data)))
	validator4File := &filestore.File{
		ID:        validator4ID,
		Balance:   storageCost + 1000,
		TxManager: filestore.FileID{},
		Data:      validator4Data,
	}
	_, err = fs.CreateFile(validator4File)
	if err != nil {
		t.Fatalf("failed to create validator4 file: %v", err)
	}

	// Enumerate active validators
	activeValidators, err := cm.enumerateActiveValidators()
	if err != nil {
		t.Fatalf("enumerateActiveValidators failed: %v", err)
	}

	// Should only return validator1 and validator2
	if len(activeValidators) != 2 {
		t.Errorf("expected 2 active validators, got %d", len(activeValidators))
	}

	// Verify the validators are correct
	foundValidator1 := false
	foundValidator2 := false

	for _, v := range activeValidators {
		if v.FileID == validator1ID {
			foundValidator1 = true
			if v.Stake != 2000000 {
				t.Errorf("validator1 stake mismatch: got %d, expected 2000000", v.Stake)
			}
		} else if v.FileID == validator2ID {
			foundValidator2 = true
			if v.Stake != 5000000 {
				t.Errorf("validator2 stake mismatch: got %d, expected 5000000", v.Stake)
			}
		} else {
			t.Errorf("unexpected validator in results: %s", v.FileID.String())
		}
	}

	if !foundValidator1 {
		t.Errorf("validator1 not found in active validators")
	}
	if !foundValidator2 {
		t.Errorf("validator2 not found in active validators")
	}
}

// TestProcessEpochBoundaryScheduleRecalculation tests schedule recalculation at epoch boundary
func TestProcessEpochBoundaryScheduleRecalculation(t *testing.T) {
	// Create a temporary FileStore
	tmpDir := t.TempDir()
	fs, err := filestore.NewFileStore(tmpDir)
	if err != nil {
		t.Fatalf("failed to create FileStore: %v", err)
	}
	defer fs.Close()

	// Create ConsensusManager with small epoch for testing
	config := GenesisConfig{
		EpochLength: 10,
		GenesisValidators: []GenesisValidator{
			{
				PublicKey:   [32]byte{1},
				StakeAmount: 1000000,
			},
		},
	}

	pk := [32]byte{1}
	validatorID := filestore.GenerateFileID(append([]byte("validator:"), pk[:]...))
	cm := NewConsensusManagerWithGenesis(validatorID, nil, config)
	cm.SetFileStore(fs)

	// Create mock runtime
	rt := &runtime.Runtime{}
	cm.SetRuntime(rt)

	// Create test validator records
	pk1 := [32]byte{1, 2, 3}
	pk2 := [32]byte{4, 5, 6}

	validator1ID := filestore.GenerateFileID(append([]byte("validator:"), pk1[:]...))
	validator2ID := filestore.GenerateFileID(append([]byte("validator:"), pk2[:]...))

	// Create validator 1: active with 2 Neon stake
	validator1Data, err := quanticscript.SerializeValidatorRecord(
		pk1[:],
		0,       // commission
		2000000, // totalStake (2 Neon)
		1,       // status (active)
		0,       // blocksProduced
		0,       // missedBlocks
		0,       // slashedThisEpoch
	)
	if err != nil {
		t.Fatalf("failed to serialize validator1: %v", err)
	}

	storageCost := filestore.CalculateStorageCost(int64(len(validator1Data)))
	validator1File := &filestore.File{
		ID:        validator1ID,
		Balance:   storageCost + 1000,
		TxManager: filestore.FileID{},
		Data:      validator1Data,
	}
	_, err = fs.CreateFile(validator1File)
	if err != nil {
		t.Fatalf("failed to create validator1 file: %v", err)
	}

	// Create validator 2: active with 3 Neon stake
	validator2Data, err := quanticscript.SerializeValidatorRecord(
		pk2[:],
		0,       // commission
		3000000, // totalStake (3 Neon)
		1,       // status (active)
		0,       // blocksProduced
		0,       // missedBlocks
		0,       // slashedThisEpoch
	)
	if err != nil {
		t.Fatalf("failed to serialize validator2: %v", err)
	}

	storageCost = filestore.CalculateStorageCost(int64(len(validator2Data)))
	validator2File := &filestore.File{
		ID:        validator2ID,
		Balance:   storageCost + 1000,
		TxManager: filestore.FileID{},
		Data:      validator2Data,
	}
	_, err = fs.CreateFile(validator2File)
	if err != nil {
		t.Fatalf("failed to create validator2 file: %v", err)
	}

	// Set initial schedule
	initialSchedule := []filestore.FileID{validator1ID, validator1ID}
	cm.SetValidatorSchedule(initialSchedule)

	// Process epoch boundary at slot 10
	lastBlockHash := [32]byte{1, 2, 3, 4, 5, 6, 7, 8}
	err = cm.ProcessEpochBoundaryWithHash(10, lastBlockHash)
	if err != nil {
		t.Fatalf("ProcessEpochBoundaryWithHash failed: %v", err)
	}

	// Verify epoch was incremented
	if cm.GetCurrentEpoch() != 1 {
		t.Errorf("expected epoch 1, got %d", cm.GetCurrentEpoch())
	}

	// Verify schedule was updated (should have 10 slots)
	if len(cm.validatorSchedule) != 10 {
		t.Errorf("expected schedule length 10, got %d", len(cm.validatorSchedule))
	}

	// Verify schedule contains only valid validators
	for i, vid := range cm.validatorSchedule {
		if vid != validator1ID && vid != validator2ID {
			t.Errorf("slot %d has invalid validator: %s", i, vid.String())
		}
	}
}

// TestProcessEpochBoundarySchedulePersistence tests schedule persistence to Epoch State File
func TestProcessEpochBoundarySchedulePersistence(t *testing.T) {
	// Create a temporary FileStore
	tmpDir := t.TempDir()
	fs, err := filestore.NewFileStore(tmpDir)
	if err != nil {
		t.Fatalf("failed to create FileStore: %v", err)
	}
	defer fs.Close()

	// Create ConsensusManager
	config := GenesisConfig{
		EpochLength: 10,
		GenesisValidators: []GenesisValidator{
			{
				PublicKey:   [32]byte{1},
				StakeAmount: 1000000,
			},
		},
	}

	pk := [32]byte{1}
	validatorID := filestore.GenerateFileID(append([]byte("validator:"), pk[:]...))
	cm := NewConsensusManagerWithGenesis(validatorID, nil, config)
	cm.SetFileStore(fs)

	// Create mock runtime
	rt := &runtime.Runtime{}
	cm.SetRuntime(rt)

	// Create test validator
	pk1 := [32]byte{1, 2, 3}
	validator1ID := filestore.GenerateFileID(append([]byte("validator:"), pk1[:]...))

	validator1Data, err := quanticscript.SerializeValidatorRecord(
		pk1[:],
		0,       // commission
		2000000, // totalStake (2 Neon)
		1,       // status (active)
		0,       // blocksProduced
		0,       // missedBlocks
		0,       // slashedThisEpoch
	)
	if err != nil {
		t.Fatalf("failed to serialize validator1: %v", err)
	}

	storageCost := filestore.CalculateStorageCost(int64(len(validator1Data)))
	validator1File := &filestore.File{
		ID:        validator1ID,
		Balance:   storageCost + 1000,
		TxManager: filestore.FileID{},
		Data:      validator1Data,
	}
	_, err = fs.CreateFile(validator1File)
	if err != nil {
		t.Fatalf("failed to create validator1 file: %v", err)
	}

	// Process epoch boundary
	lastBlockHash := [32]byte{1, 2, 3, 4, 5, 6, 7, 8}
	err = cm.ProcessEpochBoundaryWithHash(10, lastBlockHash)
	if err != nil {
		t.Fatalf("ProcessEpochBoundaryWithHash failed: %v", err)
	}

	// Verify Epoch State File was created/updated
	epochStateFile, err := fs.GetFile(genesis.EpochStateFileID)
	if err != nil {
		t.Fatalf("failed to get Epoch State File: %v", err)
	}

	// Deserialize and verify
	epochNumber, epochStartSlot, totalSlotsInEpoch, validatorSchedule, _, err := quanticscript.DeserializeEpochState(epochStateFile.Data)
	if err != nil {
		t.Fatalf("failed to deserialize epoch state: %v", err)
	}

	if epochNumber != 1 {
		t.Errorf("expected epoch 1, got %d", epochNumber)
	}
	if epochStartSlot != 10 {
		t.Errorf("expected epochStartSlot 10, got %d", epochStartSlot)
	}
	if totalSlotsInEpoch != 10 {
		t.Errorf("expected totalSlotsInEpoch 10, got %d", totalSlotsInEpoch)
	}
	if len(validatorSchedule) != 10 {
		t.Errorf("expected schedule length 10, got %d", len(validatorSchedule))
	}
}

// TestProcessEpochBoundaryScheduleRestoration tests schedule restoration from Epoch State File on node restart
func TestProcessEpochBoundaryScheduleRestoration(t *testing.T) {
	// Create a temporary FileStore
	tmpDir := t.TempDir()
	fs, err := filestore.NewFileStore(tmpDir)
	if err != nil {
		t.Fatalf("failed to create FileStore: %v", err)
	}
	defer fs.Close()

	// Create test validator
	pk1 := [32]byte{1, 2, 3}
	validator1ID := filestore.GenerateFileID(append([]byte("validator:"), pk1[:]...))

	// Create a pre-existing Epoch State File
	validator1FileID := [32]byte{}
	copy(validator1FileID[:], validator1ID[:])

	fullSchedule := make([][32]byte, 10)
	for i := 0; i < 10; i++ {
		fullSchedule[i] = validator1FileID
	}

	missedBlockCounters := []int64{5}

	epochStateData, err := quanticscript.SerializeEpochState(
		3,   // epochNumber
		300, // epochStartSlot
		10,  // totalSlotsInEpoch
		fullSchedule,
		missedBlockCounters,
	)
	if err != nil {
		t.Fatalf("failed to serialize epoch state: %v", err)
	}

	storageCost := filestore.CalculateStorageCost(int64(len(epochStateData)))
	epochStateFile := &filestore.File{
		ID:        genesis.EpochStateFileID,
		Balance:   storageCost + 1000,
		TxManager: filestore.FileID{},
		Data:      epochStateData,
	}
	_, err = fs.CreateFile(epochStateFile)
	if err != nil {
		t.Fatalf("failed to create Epoch State File: %v", err)
	}

	// Create ConsensusManager and initialize genesis (should restore from file)
	config := GenesisConfig{
		EpochLength: 10,
		GenesisValidators: []GenesisValidator{
			{
				PublicKey:   pk1,
				StakeAmount: 1000000,
			},
		},
	}

	pk := [32]byte{1}
	validatorID := filestore.GenerateFileID(append([]byte("validator:"), pk[:]...))
	cm := NewConsensusManagerWithGenesis(validatorID, nil, config)
	cm.SetFileStore(fs)

	err = cm.InitializeGenesis(config)
	if err != nil {
		t.Fatalf("InitializeGenesis failed: %v", err)
	}

	// Verify epoch was restored
	if cm.GetCurrentEpoch() != 3 {
		t.Errorf("expected epoch 3, got %d", cm.GetCurrentEpoch())
	}

	// Verify schedule was restored
	if len(cm.validatorSchedule) != 10 {
		t.Errorf("expected schedule length 10, got %d", len(cm.validatorSchedule))
	}

	for i, vid := range cm.validatorSchedule {
		if vid != validator1ID {
			t.Errorf("slot %d has wrong validator: got %s, expected %s", i, vid.String(), validator1ID.String())
		}
	}
}

// TestCompactScheduleSerializationRoundTrip tests compact schedule serialization and expansion round-trip
func TestCompactScheduleSerializationRoundTrip(t *testing.T) {
	validator1 := [32]byte{1, 2, 3}
	validator2 := [32]byte{4, 5, 6}
	validator3 := [32]byte{7, 8, 9}

	// Create a schedule with repeated validators
	fullSchedule := make([][32]byte, 100)
	for i := 0; i < 100; i++ {
		if i < 50 {
			fullSchedule[i] = validator1
		} else if i < 80 {
			fullSchedule[i] = validator2
		} else {
			fullSchedule[i] = validator3
		}
	}

	missedBlockCounters := []int64{10, 5, 2}

	// Serialize
	data, err := quanticscript.SerializeEpochState(
		5,
		500,
		100,
		fullSchedule,
		missedBlockCounters,
	)
	if err != nil {
		t.Fatalf("SerializeEpochState failed: %v", err)
	}

	// Deserialize
	epochNumber, epochStartSlot, totalSlotsInEpoch, restoredSchedule, restoredCounters, err := quanticscript.DeserializeEpochState(data)
	if err != nil {
		t.Fatalf("DeserializeEpochState failed: %v", err)
	}

	// Verify metadata
	if epochNumber != 5 {
		t.Errorf("epochNumber mismatch: got %d, expected 5", epochNumber)
	}
	if epochStartSlot != 500 {
		t.Errorf("epochStartSlot mismatch: got %d, expected 500", epochStartSlot)
	}
	if totalSlotsInEpoch != 100 {
		t.Errorf("totalSlotsInEpoch mismatch: got %d, expected 100", totalSlotsInEpoch)
	}

	// Verify schedule
	if len(restoredSchedule) != len(fullSchedule) {
		t.Fatalf("schedule length mismatch: got %d, expected %d", len(restoredSchedule), len(fullSchedule))
	}

	for i := range fullSchedule {
		if restoredSchedule[i] != fullSchedule[i] {
			t.Errorf("schedule mismatch at slot %d: got %v, expected %v", i, restoredSchedule[i], fullSchedule[i])
		}
	}

	// Verify counters
	if len(restoredCounters) != len(missedBlockCounters) {
		t.Fatalf("counters length mismatch: got %d, expected %d", len(restoredCounters), len(missedBlockCounters))
	}

	for i := range missedBlockCounters {
		if restoredCounters[i] != missedBlockCounters[i] {
			t.Errorf("counter mismatch at index %d: got %d, expected %d", i, restoredCounters[i], missedBlockCounters[i])
		}
	}
}

// TestRecordMissedBlockIncrementsValidatorRecord tests that missed block counter increments in Validator Record
func TestRecordMissedBlockIncrementsValidatorRecord(t *testing.T) {
	// Create a temporary FileStore
	tmpDir := t.TempDir()
	fs, err := filestore.NewFileStore(tmpDir)
	if err != nil {
		t.Fatalf("failed to create FileStore: %v", err)
	}
	defer fs.Close()

	// Create ConsensusManager
	config := GenesisConfig{
		EpochLength: 100,
		GenesisValidators: []GenesisValidator{
			{
				PublicKey:   [32]byte{1},
				StakeAmount: 1000000,
			},
		},
	}

	pk := [32]byte{1}
	validatorID := filestore.GenerateFileID(append([]byte("validator:"), pk[:]...))
	cm := NewConsensusManagerWithGenesis(validatorID, nil, config)
	cm.SetFileStore(fs)

	// Create test validator
	pk1 := [32]byte{1, 2, 3}
	validator1ID := filestore.GenerateFileID(append([]byte("validator:"), pk1[:]...))

	validator1Data, err := quanticscript.SerializeValidatorRecord(
		pk1[:],
		0,       // commission
		2000000, // totalStake (2 Neon)
		1,       // status (active)
		0,       // blocksProduced
		0,       // missedBlocks (initial)
		0,       // slashedThisEpoch
	)
	if err != nil {
		t.Fatalf("failed to serialize validator1: %v", err)
	}

	storageCost := filestore.CalculateStorageCost(int64(len(validator1Data)))
	validator1File := &filestore.File{
		ID:        validator1ID,
		Balance:   storageCost + 1000,
		TxManager: filestore.FileID{},
		Data:      validator1Data,
	}
	_, err = fs.CreateFile(validator1File)
	if err != nil {
		t.Fatalf("failed to create validator1 file: %v", err)
	}

	// Create Epoch State File with validator in schedule
	validator1FileID := [32]byte{}
	copy(validator1FileID[:], validator1ID[:])

	fullSchedule := make([][32]byte, 10)
	for i := 0; i < 10; i++ {
		fullSchedule[i] = validator1FileID
	}

	missedBlockCounters := []int64{0}

	epochStateData, err := quanticscript.SerializeEpochState(
		0,  // epochNumber
		0,  // epochStartSlot
		10, // totalSlotsInEpoch
		fullSchedule,
		missedBlockCounters,
	)
	if err != nil {
		t.Fatalf("failed to serialize epoch state: %v", err)
	}

	storageCost = filestore.CalculateStorageCost(int64(len(epochStateData)))
	epochStateFile := &filestore.File{
		ID:        genesis.EpochStateFileID,
		Balance:   storageCost + 1000,
		TxManager: filestore.FileID{},
		Data:      epochStateData,
	}
	_, err = fs.CreateFile(epochStateFile)
	if err != nil {
		t.Fatalf("failed to create Epoch State File: %v", err)
	}

	// Record a missed block
	err = cm.RecordMissedBlock(5, validator1ID)
	if err != nil {
		t.Fatalf("RecordMissedBlock failed: %v", err)
	}

	// Verify missed block counter was incremented in Validator Record
	updatedValidatorFile, err := fs.GetFile(validator1ID)
	if err != nil {
		t.Fatalf("failed to get updated validator file: %v", err)
	}

	_, _, _, _, _, missedBlocks, _, err := quanticscript.DeserializeValidatorRecord(updatedValidatorFile.Data)
	if err != nil {
		t.Fatalf("failed to deserialize validator record: %v", err)
	}

	if missedBlocks != 1 {
		t.Errorf("expected missedBlocks=1, got %d", missedBlocks)
	}

	// Record another missed block
	err = cm.RecordMissedBlock(6, validator1ID)
	if err != nil {
		t.Fatalf("RecordMissedBlock failed on second call: %v", err)
	}

	// Verify counter incremented again
	updatedValidatorFile, err = fs.GetFile(validator1ID)
	if err != nil {
		t.Fatalf("failed to get updated validator file: %v", err)
	}

	_, _, _, _, _, missedBlocks, _, err = quanticscript.DeserializeValidatorRecord(updatedValidatorFile.Data)
	if err != nil {
		t.Fatalf("failed to deserialize validator record: %v", err)
	}

	if missedBlocks != 2 {
		t.Errorf("expected missedBlocks=2, got %d", missedBlocks)
	}
}

// TestRecordMissedBlockIncrementsEpochState tests that missed block counter increments in Epoch State File
func TestRecordMissedBlockIncrementsEpochState(t *testing.T) {
	// Create a temporary FileStore
	tmpDir := t.TempDir()
	fs, err := filestore.NewFileStore(tmpDir)
	if err != nil {
		t.Fatalf("failed to create FileStore: %v", err)
	}
	defer fs.Close()

	// Create ConsensusManager
	config := GenesisConfig{
		EpochLength: 100,
		GenesisValidators: []GenesisValidator{
			{
				PublicKey:   [32]byte{1},
				StakeAmount: 1000000,
			},
		},
	}

	pk := [32]byte{1}
	validatorID := filestore.GenerateFileID(append([]byte("validator:"), pk[:]...))
	cm := NewConsensusManagerWithGenesis(validatorID, nil, config)
	cm.SetFileStore(fs)

	// Create test validators
	pk1 := [32]byte{1, 2, 3}
	pk2 := [32]byte{4, 5, 6}
	validator1ID := filestore.GenerateFileID(append([]byte("validator:"), pk1[:]...))
	validator2ID := filestore.GenerateFileID(append([]byte("validator:"), pk2[:]...))

	// Create validator 1
	validator1Data, err := quanticscript.SerializeValidatorRecord(
		pk1[:],
		0,       // commission
		2000000, // totalStake (2 Neon)
		1,       // status (active)
		0,       // blocksProduced
		0,       // missedBlocks
		0,       // slashedThisEpoch
	)
	if err != nil {
		t.Fatalf("failed to serialize validator1: %v", err)
	}

	storageCost := filestore.CalculateStorageCost(int64(len(validator1Data)))
	validator1File := &filestore.File{
		ID:        validator1ID,
		Balance:   storageCost + 1000,
		TxManager: filestore.FileID{},
		Data:      validator1Data,
	}
	_, err = fs.CreateFile(validator1File)
	if err != nil {
		t.Fatalf("failed to create validator1 file: %v", err)
	}

	// Create validator 2
	validator2Data, err := quanticscript.SerializeValidatorRecord(
		pk2[:],
		0,       // commission
		3000000, // totalStake (3 Neon)
		1,       // status (active)
		0,       // blocksProduced
		0,       // missedBlocks
		0,       // slashedThisEpoch
	)
	if err != nil {
		t.Fatalf("failed to serialize validator2: %v", err)
	}

	storageCost = filestore.CalculateStorageCost(int64(len(validator2Data)))
	validator2File := &filestore.File{
		ID:        validator2ID,
		Balance:   storageCost + 1000,
		TxManager: filestore.FileID{},
		Data:      validator2Data,
	}
	_, err = fs.CreateFile(validator2File)
	if err != nil {
		t.Fatalf("failed to create validator2 file: %v", err)
	}

	// Create Epoch State File with both validators in schedule
	validator1FileID := [32]byte{}
	validator2FileID := [32]byte{}
	copy(validator1FileID[:], validator1ID[:])
	copy(validator2FileID[:], validator2ID[:])

	fullSchedule := make([][32]byte, 10)
	for i := 0; i < 10; i++ {
		if i < 5 {
			fullSchedule[i] = validator1FileID
		} else {
			fullSchedule[i] = validator2FileID
		}
	}

	missedBlockCounters := []int64{0, 0}

	epochStateData, err := quanticscript.SerializeEpochState(
		0,  // epochNumber
		0,  // epochStartSlot
		10, // totalSlotsInEpoch
		fullSchedule,
		missedBlockCounters,
	)
	if err != nil {
		t.Fatalf("failed to serialize epoch state: %v", err)
	}

	storageCost = filestore.CalculateStorageCost(int64(len(epochStateData)))
	epochStateFile := &filestore.File{
		ID:        genesis.EpochStateFileID,
		Balance:   storageCost + 1000,
		TxManager: filestore.FileID{},
		Data:      epochStateData,
	}
	_, err = fs.CreateFile(epochStateFile)
	if err != nil {
		t.Fatalf("failed to create Epoch State File: %v", err)
	}

	// Record a missed block for validator1
	err = cm.RecordMissedBlock(2, validator1ID)
	if err != nil {
		t.Fatalf("RecordMissedBlock failed: %v", err)
	}

	// Verify missed block counter was incremented in Epoch State File
	updatedEpochStateFile, err := fs.GetFile(genesis.EpochStateFileID)
	if err != nil {
		t.Fatalf("failed to get updated epoch state file: %v", err)
	}

	_, _, _, _, missedCounters, err := quanticscript.DeserializeEpochState(updatedEpochStateFile.Data)
	if err != nil {
		t.Fatalf("failed to deserialize epoch state: %v", err)
	}

	if len(missedCounters) != 2 {
		t.Fatalf("expected 2 missed block counters, got %d", len(missedCounters))
	}

	if missedCounters[0] != 1 {
		t.Errorf("expected validator1 missed counter=1, got %d", missedCounters[0])
	}
	if missedCounters[1] != 0 {
		t.Errorf("expected validator2 missed counter=0, got %d", missedCounters[1])
	}

	// Record a missed block for validator2
	err = cm.RecordMissedBlock(7, validator2ID)
	if err != nil {
		t.Fatalf("RecordMissedBlock failed for validator2: %v", err)
	}

	// Verify both counters
	updatedEpochStateFile, err = fs.GetFile(genesis.EpochStateFileID)
	if err != nil {
		t.Fatalf("failed to get updated epoch state file: %v", err)
	}

	_, _, _, _, missedCounters, err = quanticscript.DeserializeEpochState(updatedEpochStateFile.Data)
	if err != nil {
		t.Fatalf("failed to deserialize epoch state: %v", err)
	}

	if missedCounters[0] != 1 {
		t.Errorf("expected validator1 missed counter=1, got %d", missedCounters[0])
	}
	if missedCounters[1] != 1 {
		t.Errorf("expected validator2 missed counter=1, got %d", missedCounters[1])
	}
}

// TestSlotSkipDoesNotProduceBlockFromNonScheduledValidator tests that slot skip does not produce block from non-scheduled validator
func TestSlotSkipDoesNotProduceBlockFromNonScheduledValidator(t *testing.T) {
	// Create two validator IDs
	pk1 := [32]byte{1, 2, 3}
	pk2 := [32]byte{4, 5, 6}
	validator1ID := filestore.GenerateFileID(append([]byte("validator:"), pk1[:]...))
	validator2ID := filestore.GenerateFileID(append([]byte("validator:"), pk2[:]...))

	// Create ConsensusManager for validator1
	config := GenesisConfig{
		EpochLength: 10,
		GenesisValidators: []GenesisValidator{
			{
				PublicKey:   pk1,
				StakeAmount: 1000000,
			},
		},
	}

	cm := NewConsensusManagerWithGenesis(validator1ID, nil, config)

	// Set up schedule where validator2 is scheduled for slot 0
	cm.validatorSchedule = []filestore.FileID{validator2ID}

	// Verify that validator1 is NOT the leader for slot 0
	if cm.IsLeader(0) {
		t.Errorf("validator1 should not be leader when validator2 is scheduled")
	}

	// This test verifies the IsLeader logic - the actual block production loop
	// should check IsLeader and skip producing a block if it returns false
	// The block production loop implementation will be tested in integration tests
}

// TestSlotSkipLogic tests the slot skip and missed block detection logic
func TestSlotSkipLogic(t *testing.T) {
	// Create a temporary FileStore
	tmpDir := t.TempDir()
	fs, err := filestore.NewFileStore(tmpDir)
	if err != nil {
		t.Fatalf("failed to create FileStore: %v", err)
	}
	defer fs.Close()

	// Create ConsensusManager
	config := GenesisConfig{
		EpochLength: 10,
		GenesisValidators: []GenesisValidator{
			{
				PublicKey:   [32]byte{1},
				StakeAmount: 1000000,
			},
		},
	}

	pk := [32]byte{1}
	validatorID := filestore.GenerateFileID(append([]byte("validator:"), pk[:]...))
	cm := NewConsensusManagerWithGenesis(validatorID, nil, config)
	cm.SetFileStore(fs)

	// Create test validators
	pk1 := [32]byte{1, 2, 3}
	pk2 := [32]byte{4, 5, 6}
	validator1ID := filestore.GenerateFileID(append([]byte("validator:"), pk1[:]...))
	validator2ID := filestore.GenerateFileID(append([]byte("validator:"), pk2[:]...))

	// Create validator records
	validator1Data, err := quanticscript.SerializeValidatorRecord(
		pk1[:], 0, 2000000, 1, 0, 0, 0,
	)
	if err != nil {
		t.Fatalf("failed to serialize validator1: %v", err)
	}

	storageCost := filestore.CalculateStorageCost(int64(len(validator1Data)))
	validator1File := &filestore.File{
		ID: validator1ID, Balance: storageCost + 1000, TxManager: filestore.FileID{}, Data: validator1Data,
	}
	_, err = fs.CreateFile(validator1File)
	if err != nil {
		t.Fatalf("failed to create validator1 file: %v", err)
	}

	validator2Data, err := quanticscript.SerializeValidatorRecord(
		pk2[:], 0, 3000000, 1, 0, 0, 0,
	)
	if err != nil {
		t.Fatalf("failed to serialize validator2: %v", err)
	}

	storageCost = filestore.CalculateStorageCost(int64(len(validator2Data)))
	validator2File := &filestore.File{
		ID: validator2ID, Balance: storageCost + 1000, TxManager: filestore.FileID{}, Data: validator2Data,
	}
	_, err = fs.CreateFile(validator2File)
	if err != nil {
		t.Fatalf("failed to create validator2 file: %v", err)
	}

	// Set up schedule where validator1 is scheduled for slot 0, validator2 for slot 1
	cm.validatorSchedule = []filestore.FileID{validator1ID, validator2ID}

	// Create Epoch State File
	validator1FileID := [32]byte{}
	validator2FileID := [32]byte{}
	copy(validator1FileID[:], validator1ID[:])
	copy(validator2FileID[:], validator2ID[:])

	fullSchedule := [][32]byte{validator1FileID, validator2FileID}
	missedBlockCounters := []int64{0, 0}

	epochStateData, err := quanticscript.SerializeEpochState(0, 0, 10, fullSchedule, missedBlockCounters)
	if err != nil {
		t.Fatalf("failed to serialize epoch state: %v", err)
	}

	storageCost = filestore.CalculateStorageCost(int64(len(epochStateData)))
	epochStateFile := &filestore.File{
		ID: genesis.EpochStateFileID, Balance: storageCost + 1000, TxManager: filestore.FileID{}, Data: epochStateData,
	}
	_, err = fs.CreateFile(epochStateFile)
	if err != nil {
		t.Fatalf("failed to create Epoch State File: %v", err)
	}

	// Test ProcessSlotSkip function
	err = cm.ProcessSlotSkip(0)
	if err != nil {
		t.Fatalf("ProcessSlotSkip failed: %v", err)
	}

	// Verify missed block was recorded for validator1
	updatedValidatorFile, err := fs.GetFile(validator1ID)
	if err != nil {
		t.Fatalf("failed to get updated validator file: %v", err)
	}

	_, _, _, _, _, missedBlocks, _, err := quanticscript.DeserializeValidatorRecord(updatedValidatorFile.Data)
	if err != nil {
		t.Fatalf("failed to deserialize validator record: %v", err)
	}

	if missedBlocks != 1 {
		t.Errorf("expected validator1 missedBlocks=1, got %d", missedBlocks)
	}

	// Verify epoch state was updated
	updatedEpochStateFile, err := fs.GetFile(genesis.EpochStateFileID)
	if err != nil {
		t.Fatalf("failed to get updated epoch state file: %v", err)
	}

	_, _, _, _, missedCounters, err := quanticscript.DeserializeEpochState(updatedEpochStateFile.Data)
	if err != nil {
		t.Fatalf("failed to deserialize epoch state: %v", err)
	}

	if len(missedCounters) != 2 {
		t.Fatalf("expected 2 missed block counters, got %d", len(missedCounters))
	}

	if missedCounters[0] != 1 {
		t.Errorf("expected validator1 missed counter=1, got %d", missedCounters[0])
	}
	if missedCounters[1] != 0 {
		t.Errorf("expected validator2 missed counter=0, got %d", missedCounters[1])
	}
}

// TestRecordBlockProductionIncrementsValidatorRecord tests that block production counter increments in Validator Record
func TestRecordBlockProductionIncrementsValidatorRecord(t *testing.T) {
	// Create a temporary FileStore
	tmpDir := t.TempDir()
	fs, err := filestore.NewFileStore(tmpDir)
	if err != nil {
		t.Fatalf("failed to create FileStore: %v", err)
	}
	defer fs.Close()

	// Create ConsensusManager
	config := GenesisConfig{
		EpochLength: 100,
		GenesisValidators: []GenesisValidator{
			{
				PublicKey:   [32]byte{1},
				StakeAmount: 1000000,
			},
		},
	}

	pk := [32]byte{1}
	validatorID := filestore.GenerateFileID(append([]byte("validator:"), pk[:]...))
	cm := NewConsensusManagerWithGenesis(validatorID, nil, config)
	cm.SetFileStore(fs)

	// Create test validator record
	pk1 := [32]byte{1, 2, 3}
	validator1ID := filestore.GenerateFileID(append([]byte("validator:"), pk1[:]...))

	validator1Data, err := quanticscript.SerializeValidatorRecord(
		pk1[:],
		0,       // commission
		2000000, // totalStake (2 Neon)
		1,       // status (active)
		0,       // blocksProduced (initial)
		0,       // missedBlocks
		0,       // slashedThisEpoch
	)
	if err != nil {
		t.Fatalf("failed to serialize validator1: %v", err)
	}

	storageCost := filestore.CalculateStorageCost(int64(len(validator1Data)))
	validator1File := &filestore.File{
		ID:        validator1ID,
		Balance:   storageCost + 1000,
		TxManager: filestore.FileID{},
		Data:      validator1Data,
	}
	_, err = fs.CreateFile(validator1File)
	if err != nil {
		t.Fatalf("failed to create validator1 file: %v", err)
	}

	// Record a block production
	err = cm.RecordBlockProduction(validator1ID)
	if err != nil {
		t.Fatalf("RecordBlockProduction failed: %v", err)
	}

	// Verify block production counter was incremented in Validator Record
	updatedValidatorFile, err := fs.GetFile(validator1ID)
	if err != nil {
		t.Fatalf("failed to get updated validator file: %v", err)
	}

	_, _, _, _, blocksProduced, _, _, err := quanticscript.DeserializeValidatorRecord(updatedValidatorFile.Data)
	if err != nil {
		t.Fatalf("failed to deserialize validator record: %v", err)
	}

	if blocksProduced != 1 {
		t.Errorf("expected blocksProduced=1 after first block, got %d", blocksProduced)
	}

	// Record another block production
	err = cm.RecordBlockProduction(validator1ID)
	if err != nil {
		t.Fatalf("RecordBlockProduction failed on second call: %v", err)
	}

	// Verify counter incremented again
	updatedValidatorFile, err = fs.GetFile(validator1ID)
	if err != nil {
		t.Fatalf("failed to get updated validator file: %v", err)
	}

	_, _, _, _, blocksProduced, _, _, err = quanticscript.DeserializeValidatorRecord(updatedValidatorFile.Data)
	if err != nil {
		t.Fatalf("failed to deserialize validator record: %v", err)
	}

	if blocksProduced != 2 {
		t.Errorf("expected blocksProduced=2 after second block, got %d", blocksProduced)
	}
}

// TestBlockProductionCounterResetAtEpochBoundary tests that block production counter resets at epoch boundary
func TestBlockProductionCounterResetAtEpochBoundary(t *testing.T) {
	// Create a temporary FileStore
	tmpDir := t.TempDir()
	fs, err := filestore.NewFileStore(tmpDir)
	if err != nil {
		t.Fatalf("failed to create FileStore: %v", err)
	}
	defer fs.Close()

	// Create ConsensusManager
	config := GenesisConfig{
		EpochLength: 100,
		GenesisValidators: []GenesisValidator{
			{
				PublicKey:   [32]byte{1},
				StakeAmount: 1000000,
			},
		},
	}

	pk := [32]byte{1}
	validatorID := filestore.GenerateFileID(append([]byte("validator:"), pk[:]...))
	cm := NewConsensusManagerWithGenesis(validatorID, nil, config)
	cm.SetFileStore(fs)

	// Create test validator record
	pk1 := [32]byte{1, 2, 3}
	validator1ID := filestore.GenerateFileID(append([]byte("validator:"), pk1[:]...))

	validator1Data, err := quanticscript.SerializeValidatorRecord(
		pk1[:],
		0,       // commission
		2000000, // totalStake (2 Neon)
		1,       // status (active)
		5,       // blocksProduced (from previous epoch)
		0,       // missedBlocks
		0,       // slashedThisEpoch
	)
	if err != nil {
		t.Fatalf("failed to serialize validator1: %v", err)
	}

	storageCost := filestore.CalculateStorageCost(int64(len(validator1Data)))
	validator1File := &filestore.File{
		ID:        validator1ID,
		Balance:   storageCost + 1000,
		TxManager: filestore.FileID{},
		Data:      validator1Data,
	}
	_, err = fs.CreateFile(validator1File)
	if err != nil {
		t.Fatalf("failed to create validator1 file: %v", err)
	}

	// Reset block production counter at epoch boundary
	err = cm.ResetBlockProductionCounter(validator1ID)
	if err != nil {
		t.Fatalf("ResetBlockProductionCounter failed: %v", err)
	}

	// Verify block production counter was reset to zero
	updatedValidatorFile, err := fs.GetFile(validator1ID)
	if err != nil {
		t.Fatalf("failed to get updated validator file: %v", err)
	}

	_, _, _, _, blocksProduced, _, _, err := quanticscript.DeserializeValidatorRecord(updatedValidatorFile.Data)
	if err != nil {
		t.Fatalf("failed to deserialize validator record: %v", err)
	}

	if blocksProduced != 0 {
		t.Errorf("expected blocksProduced=0 after reset, got %d", blocksProduced)
	}
}

// TestResetAllBlockProductionCountersAtEpochBoundary tests resetting all validators' block production counters at epoch boundary
func TestResetAllBlockProductionCountersAtEpochBoundary(t *testing.T) {
	// Create a temporary FileStore
	tmpDir := t.TempDir()
	fs, err := filestore.NewFileStore(tmpDir)
	if err != nil {
		t.Fatalf("failed to create FileStore: %v", err)
	}
	defer fs.Close()

	// Create ConsensusManager
	config := GenesisConfig{
		EpochLength: 100,
		GenesisValidators: []GenesisValidator{
			{
				PublicKey:   [32]byte{1},
				StakeAmount: 1000000,
			},
		},
	}

	pk := [32]byte{1}
	validatorID := filestore.GenerateFileID(append([]byte("validator:"), pk[:]...))
	cm := NewConsensusManagerWithGenesis(validatorID, nil, config)
	cm.SetFileStore(fs)

	// Create test validator records with non-zero block production counters
	pk1 := [32]byte{1, 2, 3}
	pk2 := [32]byte{4, 5, 6}
	pk3 := [32]byte{7, 8, 9}
	validator1ID := filestore.GenerateFileID(append([]byte("validator:"), pk1[:]...))
	validator2ID := filestore.GenerateFileID(append([]byte("validator:"), pk2[:]...))
	validator3ID := filestore.GenerateFileID(append([]byte("validator:"), pk3[:]...))

	// Create validator 1 with 5 blocks produced
	validator1Data, err := quanticscript.SerializeValidatorRecord(
		pk1[:], 0, 2000000, 1, 5, 0, 0,
	)
	if err != nil {
		t.Fatalf("failed to serialize validator1: %v", err)
	}

	storageCost := filestore.CalculateStorageCost(int64(len(validator1Data)))
	validator1File := &filestore.File{
		ID: validator1ID, Balance: storageCost + 1000, TxManager: filestore.FileID{}, Data: validator1Data,
	}
	_, err = fs.CreateFile(validator1File)
	if err != nil {
		t.Fatalf("failed to create validator1 file: %v", err)
	}

	// Create validator 2 with 3 blocks produced
	validator2Data, err := quanticscript.SerializeValidatorRecord(
		pk2[:], 0, 3000000, 1, 3, 0, 0,
	)
	if err != nil {
		t.Fatalf("failed to serialize validator2: %v", err)
	}

	storageCost = filestore.CalculateStorageCost(int64(len(validator2Data)))
	validator2File := &filestore.File{
		ID: validator2ID, Balance: storageCost + 1000, TxManager: filestore.FileID{}, Data: validator2Data,
	}
	_, err = fs.CreateFile(validator2File)
	if err != nil {
		t.Fatalf("failed to create validator2 file: %v", err)
	}

	// Create validator 3 with 7 blocks produced
	validator3Data, err := quanticscript.SerializeValidatorRecord(
		pk3[:], 0, 1000000, 1, 7, 0, 0,
	)
	if err != nil {
		t.Fatalf("failed to serialize validator3: %v", err)
	}

	storageCost = filestore.CalculateStorageCost(int64(len(validator3Data)))
	validator3File := &filestore.File{
		ID: validator3ID, Balance: storageCost + 1000, TxManager: filestore.FileID{}, Data: validator3Data,
	}
	_, err = fs.CreateFile(validator3File)
	if err != nil {
		t.Fatalf("failed to create validator3 file: %v", err)
	}

	// Reset all block production counters
	err = cm.ResetAllBlockProductionCounters()
	if err != nil {
		t.Fatalf("ResetAllBlockProductionCounters failed: %v", err)
	}

	// Verify all counters were reset to zero
	for _, validatorID := range []filestore.FileID{validator1ID, validator2ID, validator3ID} {
		updatedValidatorFile, err := fs.GetFile(validatorID)
		if err != nil {
			t.Fatalf("failed to get updated validator file: %v", err)
		}

		_, _, _, _, blocksProduced, _, _, err := quanticscript.DeserializeValidatorRecord(updatedValidatorFile.Data)
		if err != nil {
			t.Fatalf("failed to deserialize validator record: %v", err)
		}

		if blocksProduced != 0 {
			t.Errorf("expected blocksProduced=0 after reset for validator %s, got %d", validatorID.String(), blocksProduced)
		}
	}
}
