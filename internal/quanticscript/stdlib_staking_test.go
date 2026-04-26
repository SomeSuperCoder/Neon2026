package quanticscript

import (
	"testing"
)

// TestSerializeDeserializeValidatorRecord tests round-trip serialization of ValidatorRecord
func TestSerializeDeserializeValidatorRecord(t *testing.T) {
	tests := []struct {
		name             string
		pubkey           [32]byte
		commission       int64
		totalStake       int64
		status           uint8
		blocksProduced   int64
		missedBlocks     int64
		slashedThisEpoch uint8
	}{
		{
			name:             "basic validator",
			pubkey:           [32]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32},
			commission:       5,
			totalStake:       1000000,
			status:           1, // active
			blocksProduced:   100,
			missedBlocks:     2,
			slashedThisEpoch: 0,
		},
		{
			name:             "inactive validator",
			pubkey:           [32]byte{32, 31, 30, 29, 28, 27, 26, 25, 24, 23, 22, 21, 20, 19, 18, 17, 16, 15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1},
			commission:       10,
			totalStake:       500000,
			status:           0, // inactive
			blocksProduced:   0,
			missedBlocks:     0,
			slashedThisEpoch: 0,
		},
		{
			name:             "slashed validator",
			pubkey:           [32]byte{255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255},
			commission:       0,
			totalStake:       950000,
			status:           1, // active
			blocksProduced:   50,
			missedBlocks:     5,
			slashedThisEpoch: 1,
		},
		{
			name:             "max commission",
			pubkey:           [32]byte{},
			commission:       100,
			totalStake:       5000000,
			status:           1,
			blocksProduced:   432000,
			missedBlocks:     0,
			slashedThisEpoch: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Serialize
			data, err := SerializeValidatorRecord(
				tt.pubkey[:],
				tt.commission,
				tt.totalStake,
				tt.status,
				tt.blocksProduced,
				tt.missedBlocks,
				tt.slashedThisEpoch,
			)
			if err != nil {
				t.Fatalf("SerializeValidatorRecord failed: %v", err)
			}

			// Verify size
			if len(data) != 66 {
				t.Errorf("expected 66 bytes, got %d", len(data))
			}

			// Deserialize
			pubkey, commission, totalStake, status, blocksProduced, missedBlocks, slashedThisEpoch, err := DeserializeValidatorRecord(data)
			if err != nil {
				t.Fatalf("DeserializeValidatorRecord failed: %v", err)
			}

			// Verify round-trip
			if string(pubkey) != string(tt.pubkey[:]) {
				t.Errorf("pubkey mismatch")
			}
			if commission != tt.commission {
				t.Errorf("commission mismatch: expected %d, got %d", tt.commission, commission)
			}
			if totalStake != tt.totalStake {
				t.Errorf("totalStake mismatch: expected %d, got %d", tt.totalStake, totalStake)
			}
			if status != tt.status {
				t.Errorf("status mismatch: expected %d, got %d", tt.status, status)
			}
			if blocksProduced != tt.blocksProduced {
				t.Errorf("blocksProduced mismatch: expected %d, got %d", tt.blocksProduced, blocksProduced)
			}
			if missedBlocks != tt.missedBlocks {
				t.Errorf("missedBlocks mismatch: expected %d, got %d", tt.missedBlocks, missedBlocks)
			}
			if slashedThisEpoch != tt.slashedThisEpoch {
				t.Errorf("slashedThisEpoch mismatch: expected %d, got %d", tt.slashedThisEpoch, slashedThisEpoch)
			}
		})
	}
}

// TestSerializeDeserializeStakeAccount tests round-trip serialization of StakeAccount
func TestSerializeDeserializeStakeAccount(t *testing.T) {
	tests := []struct {
		name              string
		delegatorPubkey   [32]byte
		validatorFileID   [32]byte
		stakedAmount      int64
		activationEpoch   int64
		status            uint8
		deactivationEpoch int64
	}{
		{
			name:              "active stake account",
			delegatorPubkey:   [32]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32},
			validatorFileID:   [32]byte{32, 31, 30, 29, 28, 27, 26, 25, 24, 23, 22, 21, 20, 19, 18, 17, 16, 15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1},
			stakedAmount:      500000,
			activationEpoch:   0,
			status:            0, // active
			deactivationEpoch: 0,
		},
		{
			name:              "deactivating stake account",
			delegatorPubkey:   [32]byte{10, 20, 30, 40, 50, 60, 70, 80, 90, 100, 110, 120, 130, 140, 150, 160, 170, 180, 190, 200, 210, 220, 230, 240, 250, 1, 2, 3, 4, 5, 6, 7},
			validatorFileID:   [32]byte{7, 6, 5, 4, 3, 2, 1, 250, 240, 230, 220, 210, 200, 190, 180, 170, 160, 150, 140, 130, 120, 110, 100, 90, 80, 70, 60, 50, 40, 30, 20, 10},
			stakedAmount:      1000000,
			activationEpoch:   5,
			status:            1, // deactivating
			deactivationEpoch: 10,
		},
		{
			name:              "withdrawn stake account",
			delegatorPubkey:   [32]byte{},
			validatorFileID:   [32]byte{},
			stakedAmount:      0,
			activationEpoch:   0,
			status:            2, // withdrawn
			deactivationEpoch: 0,
		},
		{
			name:              "large stake amount",
			delegatorPubkey:   [32]byte{255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255},
			validatorFileID:   [32]byte{255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255},
			stakedAmount:      9223372036854775807, // max int64
			activationEpoch:   1000,
			status:            0,
			deactivationEpoch: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Serialize
			data, err := SerializeStakeAccount(
				tt.delegatorPubkey[:],
				tt.validatorFileID[:],
				tt.stakedAmount,
				tt.activationEpoch,
				tt.status,
				tt.deactivationEpoch,
			)
			if err != nil {
				t.Fatalf("SerializeStakeAccount failed: %v", err)
			}

			// Verify size
			if len(data) != 89 {
				t.Errorf("expected 89 bytes, got %d", len(data))
			}

			// Deserialize
			delegatorPubkey, validatorFileID, stakedAmount, activationEpoch, status, deactivationEpoch, err := DeserializeStakeAccount(data)
			if err != nil {
				t.Fatalf("DeserializeStakeAccount failed: %v", err)
			}

			// Verify round-trip
			if string(delegatorPubkey) != string(tt.delegatorPubkey[:]) {
				t.Errorf("delegatorPubkey mismatch")
			}
			if string(validatorFileID) != string(tt.validatorFileID[:]) {
				t.Errorf("validatorFileID mismatch")
			}
			if stakedAmount != tt.stakedAmount {
				t.Errorf("stakedAmount mismatch: expected %d, got %d", tt.stakedAmount, stakedAmount)
			}
			if activationEpoch != tt.activationEpoch {
				t.Errorf("activationEpoch mismatch: expected %d, got %d", tt.activationEpoch, activationEpoch)
			}
			if status != tt.status {
				t.Errorf("status mismatch: expected %d, got %d", tt.status, status)
			}
			if deactivationEpoch != tt.deactivationEpoch {
				t.Errorf("deactivationEpoch mismatch: expected %d, got %d", tt.deactivationEpoch, deactivationEpoch)
			}
		})
	}
}

// TestSerializeDeserializeEpochState tests round-trip serialization of EpochState
func TestSerializeDeserializeEpochState(t *testing.T) {
	tests := []struct {
		name                string
		epochNumber         int64
		epochStartSlot      int64
		totalSlotsInEpoch   int64
		validatorSchedule   [][32]byte
		missedBlockCounters []int64
	}{
		{
			name:                "epoch 0 with 3 validators",
			epochNumber:         0,
			epochStartSlot:      0,
			totalSlotsInEpoch:   432000,
			validatorSchedule:   [][32]byte{{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1}, {2, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 2}, {3, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 3}},
			missedBlockCounters: []int64{0, 0, 0},
		},
		{
			name:                "epoch 1 with missed blocks",
			epochNumber:         1,
			epochStartSlot:      432000,
			totalSlotsInEpoch:   432000,
			validatorSchedule:   [][32]byte{{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1}, {2, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 2}},
			missedBlockCounters: []int64{5, 10},
		},
		{
			name:                "single validator",
			epochNumber:         10,
			epochStartSlot:      4320000,
			totalSlotsInEpoch:   432000,
			validatorSchedule:   [][32]byte{{255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255}},
			missedBlockCounters: []int64{0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Serialize
			data, err := SerializeEpochState(
				tt.epochNumber,
				tt.epochStartSlot,
				tt.totalSlotsInEpoch,
				tt.validatorSchedule,
				tt.missedBlockCounters,
			)
			if err != nil {
				t.Fatalf("SerializeEpochState failed: %v", err)
			}

			// Verify minimum size
			if len(data) < 32 {
				t.Errorf("expected at least 32 bytes, got %d", len(data))
			}

			// Deserialize
			epochNumber, epochStartSlot, totalSlotsInEpoch, validatorSchedule, missedBlockCounters, err := DeserializeEpochState(data)
			if err != nil {
				t.Fatalf("DeserializeEpochState failed: %v", err)
			}

			// Verify round-trip
			if epochNumber != tt.epochNumber {
				t.Errorf("epochNumber mismatch: expected %d, got %d", tt.epochNumber, epochNumber)
			}
			if epochStartSlot != tt.epochStartSlot {
				t.Errorf("epochStartSlot mismatch: expected %d, got %d", tt.epochStartSlot, epochStartSlot)
			}
			if totalSlotsInEpoch != tt.totalSlotsInEpoch {
				t.Errorf("totalSlotsInEpoch mismatch: expected %d, got %d", tt.totalSlotsInEpoch, totalSlotsInEpoch)
			}
			if len(validatorSchedule) != len(tt.validatorSchedule) {
				t.Errorf("validatorSchedule length mismatch: expected %d, got %d", len(tt.validatorSchedule), len(validatorSchedule))
			}
			if len(missedBlockCounters) != len(tt.missedBlockCounters) {
				t.Errorf("missedBlockCounters length mismatch: expected %d, got %d", len(tt.missedBlockCounters), len(missedBlockCounters))
			}

			// Verify schedule entries
			for i, expected := range tt.validatorSchedule {
				if i < len(validatorSchedule) {
					if string(validatorSchedule[i][:]) != string(expected[:]) {
						t.Errorf("validatorSchedule[%d] mismatch", i)
					}
				}
			}

			// Verify missed block counters
			for i, expected := range tt.missedBlockCounters {
				if i < len(missedBlockCounters) {
					if missedBlockCounters[i] != expected {
						t.Errorf("missedBlockCounters[%d] mismatch: expected %d, got %d", i, expected, missedBlockCounters[i])
					}
				}
			}
		})
	}
}

// TestSerializeDeserializeRewardPool tests round-trip serialization of RewardPool
func TestSerializeDeserializeRewardPool(t *testing.T) {
	tests := []struct {
		name                 string
		balance              int64
		lastDistributedEpoch int64
	}{
		{
			name:                 "empty reward pool",
			balance:              0,
			lastDistributedEpoch: 0,
		},
		{
			name:                 "reward pool with balance",
			balance:              1000000,
			lastDistributedEpoch: 5,
		},
		{
			name:                 "large reward pool",
			balance:              9223372036854775807, // max int64
			lastDistributedEpoch: 1000,
		},
		{
			name:                 "reward pool after many epochs",
			balance:              5000000,
			lastDistributedEpoch: 999,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Serialize
			data, err := SerializeRewardPool(tt.balance, tt.lastDistributedEpoch)
			if err != nil {
				t.Fatalf("SerializeRewardPool failed: %v", err)
			}

			// Verify size
			if len(data) != 16 {
				t.Errorf("expected 16 bytes, got %d", len(data))
			}

			// Deserialize
			balance, lastDistributedEpoch, err := DeserializeRewardPool(data)
			if err != nil {
				t.Fatalf("DeserializeRewardPool failed: %v", err)
			}

			// Verify round-trip
			if balance != tt.balance {
				t.Errorf("balance mismatch: expected %d, got %d", tt.balance, balance)
			}
			if lastDistributedEpoch != tt.lastDistributedEpoch {
				t.Errorf("lastDistributedEpoch mismatch: expected %d, got %d", tt.lastDistributedEpoch, lastDistributedEpoch)
			}
		})
	}
}

// TestValidatorRecordInvalidInputs tests error handling for invalid inputs
func TestValidatorRecordInvalidInputs(t *testing.T) {
	tests := []struct {
		name        string
		pubkey      []byte
		shouldError bool
	}{
		{
			name:        "valid pubkey",
			pubkey:      make([]byte, 32),
			shouldError: false,
		},
		{
			name:        "short pubkey",
			pubkey:      make([]byte, 31),
			shouldError: true,
		},
		{
			name:        "long pubkey",
			pubkey:      make([]byte, 33),
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := SerializeValidatorRecord(tt.pubkey, 5, 1000000, 1, 0, 0, 0)
			if (err != nil) != tt.shouldError {
				t.Errorf("expected error=%v, got error=%v", tt.shouldError, err != nil)
			}
		})
	}
}

// TestStakeAccountInvalidInputs tests error handling for invalid inputs
func TestStakeAccountInvalidInputs(t *testing.T) {
	tests := []struct {
		name            string
		delegatorPubkey []byte
		validatorFileID []byte
		shouldError     bool
	}{
		{
			name:            "valid inputs",
			delegatorPubkey: make([]byte, 32),
			validatorFileID: make([]byte, 32),
			shouldError:     false,
		},
		{
			name:            "short delegator pubkey",
			delegatorPubkey: make([]byte, 31),
			validatorFileID: make([]byte, 32),
			shouldError:     true,
		},
		{
			name:            "short validator file id",
			delegatorPubkey: make([]byte, 32),
			validatorFileID: make([]byte, 31),
			shouldError:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := SerializeStakeAccount(tt.delegatorPubkey, tt.validatorFileID, 500000, 0, 0, 0)
			if (err != nil) != tt.shouldError {
				t.Errorf("expected error=%v, got error=%v", tt.shouldError, err != nil)
			}
		})
	}
}

// TestDeserializeInvalidData tests error handling for corrupted data
func TestDeserializeInvalidData(t *testing.T) {
	tests := []struct {
		name        string
		data        []byte
		deserialize func([]byte) error
	}{
		{
			name: "ValidatorRecord too short",
			data: make([]byte, 65),
			deserialize: func(data []byte) error {
				_, _, _, _, _, _, _, err := DeserializeValidatorRecord(data)
				return err
			},
		},
		{
			name: "StakeAccount too short",
			data: make([]byte, 88),
			deserialize: func(data []byte) error {
				_, _, _, _, _, _, err := DeserializeStakeAccount(data)
				return err
			},
		},
		{
			name: "RewardPool too short",
			data: make([]byte, 15),
			deserialize: func(data []byte) error {
				_, _, err := DeserializeRewardPool(data)
				return err
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.deserialize(tt.data)
			if err == nil {
				t.Errorf("expected error for invalid data")
			}
		})
	}
}
