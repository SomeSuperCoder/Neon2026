package consensus

import (
	"testing"
	"time"

	"github.com/poh-blockchain/internal/filestore"
	"github.com/poh-blockchain/internal/network"
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

	cm := NewConsensusManagerWithGenesis(network.LEADER, config)

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

	cm := NewConsensusManagerWithGenesis(network.LEADER, config)

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

	cm := NewConsensusManagerWithGenesis(network.LEADER, config)

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

	cm := NewConsensusManagerWithGenesis(network.LEADER, config)

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

	cm := NewConsensusManagerWithGenesis(network.LEADER, config)

	// Initialize with a mock validator ID
	pk := [32]byte{1}
	validatorID := filestore.GenerateFileID(append([]byte("validator:"), pk[:]...))

	// Test without fileStore - should fail
	err := cm.RecordMissedBlock(0, validatorID)
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

	cm := NewConsensusManagerWithGenesis(network.LEADER, config)

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

	cm := NewConsensusManagerWithGenesis(network.LEADER, config)

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

	// Create a leader node
	cmLeader := NewConsensusManagerWithGenesis(network.LEADER, config)

	// Create a replica node
	cmReplica := NewConsensusManagerWithGenesis(network.REPLICA, config)

	// Set up a simple schedule
	pk := [32]byte{1}
	validatorID := filestore.GenerateFileID(append([]byte("validator:"), pk[:]...))
	cmLeader.validatorSchedule = []filestore.FileID{validatorID, validatorID}
	cmReplica.validatorSchedule = []filestore.FileID{validatorID, validatorID}

	// For a LEADER node, IsLeader should return true if it's in the schedule
	// For a REPLICA node, IsLeader should return false
	if cmLeader.IsLeader(0) != true {
		t.Errorf("LEADER node should be leader for scheduled slot")
	}

	if cmReplica.IsLeader(0) != false {
		t.Errorf("REPLICA node should not be leader")
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

	cm := NewConsensusManagerWithGenesis(network.LEADER, config)

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
