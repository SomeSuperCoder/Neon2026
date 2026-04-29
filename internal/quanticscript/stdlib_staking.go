package quanticscript

import (
	"encoding/binary"
	"fmt"
)

// stdlib_staking.go provides serialization helpers for DPoS data models.
// This implements Requirements 6.1, 10.1, 10.2, 10.3.

// ValidatorRecord serialization functions

// SerializeValidatorRecord serializes a ValidatorRecord to bytes
// Format:
//   - Bytes 0-31: pubkey (32 bytes)
//   - Bytes 32-39: commission (i64 LE)
//   - Bytes 40-47: totalDelegatedStake (i64 LE)
//   - Byte 48: status (u8: 0=inactive, 1=active, 2=deregistered)
//   - Bytes 49-56: blocksProducedThisEpoch (i64 LE)
//   - Bytes 57-64: missedBlocksThisEpoch (i64 LE)
//   - Byte 65: slashedThisEpoch (u8: 0=false, 1=true)
//
// Total: 66 bytes
func SerializeValidatorRecord(
	pubkey []byte,
	commission int64,
	totalStake int64,
	status uint8,
	blocksProduced int64,
	missedBlocks int64,
	slashedThisEpoch uint8,
) ([]byte, error) {
	if len(pubkey) != 32 {
		return nil, fmt.Errorf("pubkey must be 32 bytes, got %d", len(pubkey))
	}

	data := make([]byte, 66)
	offset := 0

	// Write pubkey
	copy(data[offset:offset+32], pubkey)
	offset += 32

	// Write commission
	binary.LittleEndian.PutUint64(data[offset:offset+8], uint64(commission))
	offset += 8

	// Write totalStake
	binary.LittleEndian.PutUint64(data[offset:offset+8], uint64(totalStake))
	offset += 8

	// Write status
	data[offset] = status
	offset++

	// Write blocksProduced
	binary.LittleEndian.PutUint64(data[offset:offset+8], uint64(blocksProduced))
	offset += 8

	// Write missedBlocks
	binary.LittleEndian.PutUint64(data[offset:offset+8], uint64(missedBlocks))
	offset += 8

	// Write slashedThisEpoch
	data[offset] = slashedThisEpoch

	return data, nil
}

// DeserializeValidatorRecord deserializes a ValidatorRecord from bytes
// Returns: pubkey, commission, totalStake, status, blocksProduced, missedBlocks, slashedThisEpoch, error
func DeserializeValidatorRecord(data []byte) ([]byte, int64, int64, uint8, int64, int64, uint8, error) {
	if len(data) != 66 {
		return nil, 0, 0, 0, 0, 0, 0, fmt.Errorf("ValidatorRecord data must be 66 bytes, got %d", len(data))
	}

	offset := 0

	// Read pubkey
	pubkey := make([]byte, 32)
	copy(pubkey, data[offset:offset+32])
	offset += 32

	// Read commission
	commission := int64(binary.LittleEndian.Uint64(data[offset : offset+8]))
	offset += 8

	// Read totalStake
	totalStake := int64(binary.LittleEndian.Uint64(data[offset : offset+8]))
	offset += 8

	// Read status
	status := data[offset]
	offset++

	// Read blocksProduced
	blocksProduced := int64(binary.LittleEndian.Uint64(data[offset : offset+8]))
	offset += 8

	// Read missedBlocks
	missedBlocks := int64(binary.LittleEndian.Uint64(data[offset : offset+8]))
	offset += 8

	// Read slashedThisEpoch
	slashedThisEpoch := data[offset]

	return pubkey, commission, totalStake, status, blocksProduced, missedBlocks, slashedThisEpoch, nil
}

// StakeAccount serialization functions

// SerializeStakeAccount serializes a StakeAccount to bytes
// Format:
//   - Bytes 0-31: delegatorPubkey (32 bytes)
//   - Bytes 32-63: validatorFileID (32 bytes)
//   - Bytes 64-71: stakedAmount (i64 LE)
//   - Bytes 72-79: activationEpoch (i64 LE)
//   - Byte 80: status (u8: 0=active, 1=deactivating, 2=withdrawn)
//   - Bytes 81-88: deactivationEpoch (i64 LE)
//
// Total: 89 bytes
func SerializeStakeAccount(
	delegatorPubkey []byte,
	validatorFileID []byte,
	stakedAmount int64,
	activationEpoch int64,
	status uint8,
	deactivationEpoch int64,
) ([]byte, error) {
	if len(delegatorPubkey) != 32 {
		return nil, fmt.Errorf("delegatorPubkey must be 32 bytes, got %d", len(delegatorPubkey))
	}
	if len(validatorFileID) != 32 {
		return nil, fmt.Errorf("validatorFileID must be 32 bytes, got %d", len(validatorFileID))
	}

	data := make([]byte, 89)
	offset := 0

	// Write delegatorPubkey
	copy(data[offset:offset+32], delegatorPubkey)
	offset += 32

	// Write validatorFileID
	copy(data[offset:offset+32], validatorFileID)
	offset += 32

	// Write stakedAmount
	binary.LittleEndian.PutUint64(data[offset:offset+8], uint64(stakedAmount))
	offset += 8

	// Write activationEpoch
	binary.LittleEndian.PutUint64(data[offset:offset+8], uint64(activationEpoch))
	offset += 8

	// Write status
	data[offset] = status
	offset++

	// Write deactivationEpoch
	binary.LittleEndian.PutUint64(data[offset:offset+8], uint64(deactivationEpoch))

	return data, nil
}

// DeserializeStakeAccount deserializes a StakeAccount from bytes
// Returns: delegatorPubkey, validatorFileID, stakedAmount, activationEpoch, status, deactivationEpoch, error
func DeserializeStakeAccount(data []byte) ([]byte, []byte, int64, int64, uint8, int64, error) {
	if len(data) != 89 {
		return nil, nil, 0, 0, 0, 0, fmt.Errorf("StakeAccount data must be 89 bytes, got %d", len(data))
	}

	offset := 0

	// Read delegatorPubkey
	delegatorPubkey := make([]byte, 32)
	copy(delegatorPubkey, data[offset:offset+32])
	offset += 32

	// Read validatorFileID
	validatorFileID := make([]byte, 32)
	copy(validatorFileID, data[offset:offset+32])
	offset += 32

	// Read stakedAmount
	stakedAmount := int64(binary.LittleEndian.Uint64(data[offset : offset+8]))
	offset += 8

	// Read activationEpoch
	activationEpoch := int64(binary.LittleEndian.Uint64(data[offset : offset+8]))
	offset += 8

	// Read status
	status := data[offset]
	offset++

	// Read deactivationEpoch
	deactivationEpoch := int64(binary.LittleEndian.Uint64(data[offset : offset+8]))

	return delegatorPubkey, validatorFileID, stakedAmount, activationEpoch, status, deactivationEpoch, nil
}

// EpochState serialization functions

// ScheduleEntry represents a compact schedule entry (validator FileID + slot count)
type ScheduleEntry struct {
	ValidatorFileID [32]byte
	AssignedSlots   int64
}

// compactSchedule converts a full slot-indexed schedule to compact format
// Compact format: validator FileID + count of consecutive slots
func compactSchedule(fullSchedule [][32]byte) []ScheduleEntry {
	if len(fullSchedule) == 0 {
		return []ScheduleEntry{}
	}

	var compact []ScheduleEntry
	currentValidator := fullSchedule[0]
	currentCount := int64(1)

	for i := 1; i < len(fullSchedule); i++ {
		if fullSchedule[i] == currentValidator {
			currentCount++
		} else {
			// Save current run
			compact = append(compact, ScheduleEntry{
				ValidatorFileID: currentValidator,
				AssignedSlots:   currentCount,
			})
			// Start new run
			currentValidator = fullSchedule[i]
			currentCount = 1
		}
	}

	// Save final run
	compact = append(compact, ScheduleEntry{
		ValidatorFileID: currentValidator,
		AssignedSlots:   currentCount,
	})

	return compact
}

// expandSchedule converts compact format to full slot-indexed array
func expandSchedule(compactSchedule []ScheduleEntry) [][32]byte {
	// Calculate total slots
	totalSlots := int64(0)
	for _, entry := range compactSchedule {
		totalSlots += entry.AssignedSlots
	}

	if totalSlots == 0 {
		return [][32]byte{}
	}

	fullSchedule := make([][32]byte, totalSlots)
	slotIndex := 0

	for _, entry := range compactSchedule {
		for i := int64(0); i < entry.AssignedSlots; i++ {
			fullSchedule[slotIndex] = entry.ValidatorFileID
			slotIndex++
		}
	}

	return fullSchedule
}

// SerializeEpochState serializes an EpochState to bytes using compact format
// Format (compact):
//   - Bytes 0-7: epochNumber (i64 LE)
//   - Bytes 8-15: epochStartSlot (i64 LE)
//   - Bytes 16-23: totalSlotsInEpoch (i64 LE)
//   - Bytes 24-31: compactEntryCount (i64 LE, C)
//   - Bytes 32+: compact schedule entries (C * 40 bytes each: validatorFileID(32) + assignedSlots(8))
//   - After schedule: missedBlockCounters for unique validators (V * 40 bytes: validatorFileID(32) + counter(8))
func SerializeEpochState(
	epochNumber int64,
	epochStartSlot int64,
	totalSlotsInEpoch int64,
	validatorSchedule [][32]byte,
	missedBlockCounters []int64,
) ([]byte, error) {
	// Convert full schedule to compact format
	compactSched := compactSchedule(validatorSchedule)

	// Build unique validator list for missed block counters
	uniqueValidators := make(map[[32]byte]int)
	var orderedValidators [][32]byte

	for _, entry := range compactSched {
		if _, exists := uniqueValidators[entry.ValidatorFileID]; !exists {
			uniqueValidators[entry.ValidatorFileID] = len(orderedValidators)
			orderedValidators = append(orderedValidators, entry.ValidatorFileID)
		}
	}

	uniqueValidatorCount := int64(len(orderedValidators))

	// Ensure missedBlockCounters matches unique validators
	if len(missedBlockCounters) != len(orderedValidators) {
		// Pad or truncate missedBlockCounters to match
		newCounters := make([]int64, uniqueValidatorCount)
		for i := int64(0); i < uniqueValidatorCount && i < int64(len(missedBlockCounters)); i++ {
			newCounters[i] = missedBlockCounters[i]
		}
		missedBlockCounters = newCounters
	}

	compactEntryCount := int64(len(compactSched))

	// Calculate total size: header(32) + compact entries(C*40) + missed block counters(V*40)
	size := 32 + (compactEntryCount * 40) + (uniqueValidatorCount * 40)
	data := make([]byte, size)
	offset := 0

	// Write epochNumber
	binary.LittleEndian.PutUint64(data[offset:offset+8], uint64(epochNumber))
	offset += 8

	// Write epochStartSlot
	binary.LittleEndian.PutUint64(data[offset:offset+8], uint64(epochStartSlot))
	offset += 8

	// Write totalSlotsInEpoch
	binary.LittleEndian.PutUint64(data[offset:offset+8], uint64(totalSlotsInEpoch))
	offset += 8

	// Write compactEntryCount
	binary.LittleEndian.PutUint64(data[offset:offset+8], uint64(compactEntryCount))
	offset += 8

	// Write compact schedule entries (preserving order)
	for _, entry := range compactSched {
		copy(data[offset:offset+32], entry.ValidatorFileID[:])
		offset += 32
		binary.LittleEndian.PutUint64(data[offset:offset+8], uint64(entry.AssignedSlots))
		offset += 8
	}

	// Write missed block counters for unique validators
	for i, validatorID := range orderedValidators {
		copy(data[offset:offset+32], validatorID[:])
		offset += 32
		binary.LittleEndian.PutUint64(data[offset:offset+8], uint64(missedBlockCounters[i]))
		offset += 8
	}

	return data, nil
}

// DeserializeEpochState deserializes an EpochState from bytes
// Returns: epochNumber, epochStartSlot, totalSlotsInEpoch, validatorSchedule (expanded), missedBlockCounters, error
func DeserializeEpochState(data []byte) (int64, int64, int64, [][32]byte, []int64, error) {
	if len(data) < 32 {
		return 0, 0, 0, nil, nil, fmt.Errorf("EpochState data too short: need at least 32 bytes, got %d", len(data))
	}

	offset := 0

	// Read epochNumber
	epochNumber := int64(binary.LittleEndian.Uint64(data[offset : offset+8]))
	offset += 8

	// Read epochStartSlot
	epochStartSlot := int64(binary.LittleEndian.Uint64(data[offset : offset+8]))
	offset += 8

	// Read totalSlotsInEpoch
	totalSlotsInEpoch := int64(binary.LittleEndian.Uint64(data[offset : offset+8]))
	offset += 8

	// Read compactEntryCount
	compactEntryCount := int64(binary.LittleEndian.Uint64(data[offset : offset+8]))
	offset += 8

	// Read compact schedule entries
	compactSched := make([]ScheduleEntry, compactEntryCount)
	for i := int64(0); i < compactEntryCount; i++ {
		copy(compactSched[i].ValidatorFileID[:], data[offset:offset+32])
		offset += 32
		compactSched[i].AssignedSlots = int64(binary.LittleEndian.Uint64(data[offset : offset+8]))
		offset += 8
	}

	// Calculate unique validator count from remaining data
	remainingBytes := len(data) - offset
	if remainingBytes%40 != 0 {
		return 0, 0, 0, nil, nil, fmt.Errorf("invalid remaining data size for missed block counters: %d bytes", remainingBytes)
	}
	uniqueValidatorCount := int64(remainingBytes / 40)

	// Read missed block counters
	missedBlockCounters := make([]int64, uniqueValidatorCount)
	for i := int64(0); i < uniqueValidatorCount; i++ {
		// Skip validator ID (already in compact schedule)
		offset += 32
		missedBlockCounters[i] = int64(binary.LittleEndian.Uint64(data[offset : offset+8]))
		offset += 8
	}

	// Expand schedule to full slot-indexed array
	validatorSchedule := expandSchedule(compactSched)

	return epochNumber, epochStartSlot, totalSlotsInEpoch, validatorSchedule, missedBlockCounters, nil
}

// RewardPool serialization functions

// SerializeRewardPool serializes a RewardPool to bytes
// Format:
//   - Bytes 0-7: balance (i64 LE)
//   - Bytes 8-15: lastDistributedEpoch (i64 LE)
//
// Total: 16 bytes
func SerializeRewardPool(balance int64, lastDistributedEpoch int64) ([]byte, error) {
	data := make([]byte, 16)
	offset := 0

	// Write balance
	binary.LittleEndian.PutUint64(data[offset:offset+8], uint64(balance))
	offset += 8

	// Write lastDistributedEpoch
	binary.LittleEndian.PutUint64(data[offset:offset+8], uint64(lastDistributedEpoch))

	return data, nil
}

// DeserializeRewardPool deserializes a RewardPool from bytes
// Returns: balance, lastDistributedEpoch, error
func DeserializeRewardPool(data []byte) (int64, int64, error) {
	if len(data) != 16 {
		return 0, 0, fmt.Errorf("RewardPool data must be 16 bytes, got %d", len(data))
	}

	offset := 0

	// Read balance
	balance := int64(binary.LittleEndian.Uint64(data[offset : offset+8]))
	offset += 8

	// Read lastDistributedEpoch
	lastDistributedEpoch := int64(binary.LittleEndian.Uint64(data[offset : offset+8]))

	return balance, lastDistributedEpoch, nil
}
