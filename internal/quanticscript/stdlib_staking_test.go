package quanticscript

import (
	"testing"
)

// TestCompactScheduleRoundTrip tests that compacting and expanding a schedule produces the original schedule
func TestCompactScheduleRoundTrip(t *testing.T) {
	// Create a test schedule with repeated validators
	validator1 := [32]byte{1, 2, 3}
	validator2 := [32]byte{4, 5, 6}
	validator3 := [32]byte{7, 8, 9}

	fullSchedule := [][32]byte{
		validator1, validator1, validator1, // 3 slots for validator1
		validator2, validator2, // 2 slots for validator2
		validator1,                                     // 1 slot for validator1
		validator3, validator3, validator3, validator3, // 4 slots for validator3
	}

	// Compact the schedule
	compact := compactSchedule(fullSchedule)

	// Verify compact format has correct number of entries
	if len(compact) != 4 {
		t.Errorf("expected 4 compact entries, got %d", len(compact))
	}

	// Verify compact entries
	if compact[0].ValidatorFileID != validator1 || compact[0].AssignedSlots != 3 {
		t.Errorf("compact[0] mismatch: got %v with %d slots", compact[0].ValidatorFileID, compact[0].AssignedSlots)
	}
	if compact[1].ValidatorFileID != validator2 || compact[1].AssignedSlots != 2 {
		t.Errorf("compact[1] mismatch: got %v with %d slots", compact[1].ValidatorFileID, compact[1].AssignedSlots)
	}
	if compact[2].ValidatorFileID != validator1 || compact[2].AssignedSlots != 1 {
		t.Errorf("compact[2] mismatch: got %v with %d slots", compact[2].ValidatorFileID, compact[2].AssignedSlots)
	}
	if compact[3].ValidatorFileID != validator3 || compact[3].AssignedSlots != 4 {
		t.Errorf("compact[3] mismatch: got %v with %d slots", compact[3].ValidatorFileID, compact[3].AssignedSlots)
	}

	// Expand the schedule
	expanded := expandSchedule(compact)

	// Verify expanded schedule matches original
	if len(expanded) != len(fullSchedule) {
		t.Fatalf("expanded schedule length mismatch: got %d, expected %d", len(expanded), len(fullSchedule))
	}

	for i := range fullSchedule {
		if expanded[i] != fullSchedule[i] {
			t.Errorf("expanded schedule mismatch at slot %d: got %v, expected %v", i, expanded[i], fullSchedule[i])
		}
	}
}

// TestCompactScheduleEmptySchedule tests compacting an empty schedule
func TestCompactScheduleEmptySchedule(t *testing.T) {
	fullSchedule := [][32]byte{}

	compact := compactSchedule(fullSchedule)

	if len(compact) != 0 {
		t.Errorf("expected empty compact schedule, got %d entries", len(compact))
	}

	expanded := expandSchedule(compact)

	if len(expanded) != 0 {
		t.Errorf("expected empty expanded schedule, got %d entries", len(expanded))
	}
}

// TestCompactScheduleSingleValidator tests compacting a schedule with a single validator
func TestCompactScheduleSingleValidator(t *testing.T) {
	validator1 := [32]byte{1, 2, 3}

	fullSchedule := [][32]byte{
		validator1, validator1, validator1, validator1, validator1,
	}

	compact := compactSchedule(fullSchedule)

	if len(compact) != 1 {
		t.Errorf("expected 1 compact entry, got %d", len(compact))
	}

	if compact[0].ValidatorFileID != validator1 || compact[0].AssignedSlots != 5 {
		t.Errorf("compact[0] mismatch: got %v with %d slots", compact[0].ValidatorFileID, compact[0].AssignedSlots)
	}

	expanded := expandSchedule(compact)

	if len(expanded) != len(fullSchedule) {
		t.Fatalf("expanded schedule length mismatch: got %d, expected %d", len(expanded), len(fullSchedule))
	}

	for i := range fullSchedule {
		if expanded[i] != fullSchedule[i] {
			t.Errorf("expanded schedule mismatch at slot %d", i)
		}
	}
}

// TestCompactScheduleAlternatingValidators tests compacting a schedule with alternating validators
func TestCompactScheduleAlternatingValidators(t *testing.T) {
	validator1 := [32]byte{1, 2, 3}
	validator2 := [32]byte{4, 5, 6}

	fullSchedule := [][32]byte{
		validator1, validator2, validator1, validator2, validator1,
	}

	compact := compactSchedule(fullSchedule)

	// Should have 5 entries (one for each slot since they alternate)
	if len(compact) != 5 {
		t.Errorf("expected 5 compact entries, got %d", len(compact))
	}

	expanded := expandSchedule(compact)

	if len(expanded) != len(fullSchedule) {
		t.Fatalf("expanded schedule length mismatch: got %d, expected %d", len(expanded), len(fullSchedule))
	}

	for i := range fullSchedule {
		if expanded[i] != fullSchedule[i] {
			t.Errorf("expanded schedule mismatch at slot %d", i)
		}
	}
}

// TestSerializeEpochStateCompactFormat tests epoch state serialization with compact format
func TestSerializeEpochStateCompactFormat(t *testing.T) {
	validator1 := [32]byte{1, 2, 3}
	validator2 := [32]byte{4, 5, 6}

	// Create a schedule with repeated validators
	fullSchedule := [][32]byte{
		validator1, validator1, validator1,
		validator2, validator2,
		validator1, validator1,
	}

	missedBlockCounters := []int64{5, 3} // Two unique validators

	// Serialize
	data, err := SerializeEpochState(
		10,   // epochNumber
		1000, // epochStartSlot
		7,    // totalSlotsInEpoch
		fullSchedule,
		missedBlockCounters,
	)
	if err != nil {
		t.Fatalf("SerializeEpochState failed: %v", err)
	}

	// Verify data size: header(32) + 3 compact entries * 40 + 2 unique validators * 40
	// Compact entries: [v1:3], [v2:2], [v1:2]
	// Unique validators: v1, v2
	expectedSize := 32 + (3 * 40) + (2 * 40)
	if len(data) != expectedSize {
		t.Errorf("serialized data size mismatch: got %d, expected %d", len(data), expectedSize)
	}

	// Deserialize
	epochNumber, epochStartSlot, totalSlotsInEpoch, validatorSchedule, counters, err := DeserializeEpochState(data)
	if err != nil {
		t.Fatalf("DeserializeEpochState failed: %v", err)
	}

	// Verify epoch metadata
	if epochNumber != 10 {
		t.Errorf("epochNumber mismatch: got %d, expected 10", epochNumber)
	}
	if epochStartSlot != 1000 {
		t.Errorf("epochStartSlot mismatch: got %d, expected 1000", epochStartSlot)
	}
	if totalSlotsInEpoch != 7 {
		t.Errorf("totalSlotsInEpoch mismatch: got %d, expected 7", totalSlotsInEpoch)
	}

	// Verify expanded schedule matches original
	if len(validatorSchedule) != len(fullSchedule) {
		t.Fatalf("validator schedule length mismatch: got %d, expected %d", len(validatorSchedule), len(fullSchedule))
	}

	for i := range fullSchedule {
		if validatorSchedule[i] != fullSchedule[i] {
			t.Errorf("validator schedule mismatch at slot %d: got %v, expected %v", i, validatorSchedule[i], fullSchedule[i])
		}
	}

	// Verify missed block counters
	if len(counters) != len(missedBlockCounters) {
		t.Fatalf("missed block counters length mismatch: got %d, expected %d", len(counters), len(missedBlockCounters))
	}

	for i := range missedBlockCounters {
		if counters[i] != missedBlockCounters[i] {
			t.Errorf("missed block counter mismatch at index %d: got %d, expected %d", i, counters[i], missedBlockCounters[i])
		}
	}
}

// TestSerializeEpochStateEmptySchedule tests serialization with empty schedule
func TestSerializeEpochStateEmptySchedule(t *testing.T) {
	fullSchedule := [][32]byte{}
	missedBlockCounters := []int64{}

	data, err := SerializeEpochState(
		0,
		0,
		0,
		fullSchedule,
		missedBlockCounters,
	)
	if err != nil {
		t.Fatalf("SerializeEpochState failed: %v", err)
	}

	// Deserialize
	epochNumber, epochStartSlot, totalSlotsInEpoch, validatorSchedule, counters, err := DeserializeEpochState(data)
	if err != nil {
		t.Fatalf("DeserializeEpochState failed: %v", err)
	}

	if epochNumber != 0 {
		t.Errorf("epochNumber mismatch: got %d, expected 0", epochNumber)
	}
	if epochStartSlot != 0 {
		t.Errorf("epochStartSlot mismatch: got %d, expected 0", epochStartSlot)
	}
	if totalSlotsInEpoch != 0 {
		t.Errorf("totalSlotsInEpoch mismatch: got %d, expected 0", totalSlotsInEpoch)
	}
	if len(validatorSchedule) != 0 {
		t.Errorf("expected empty validator schedule, got %d entries", len(validatorSchedule))
	}
	if len(counters) != 0 {
		t.Errorf("expected empty counters, got %d entries", len(counters))
	}
}

// TestSerializeEpochStateLargeSchedule tests serialization with a large schedule
func TestSerializeEpochStateLargeSchedule(t *testing.T) {
	validator1 := [32]byte{1}
	validator2 := [32]byte{2}
	validator3 := [32]byte{3}

	// Create a large schedule (1000 slots)
	fullSchedule := make([][32]byte, 1000)
	for i := 0; i < 1000; i++ {
		if i%3 == 0 {
			fullSchedule[i] = validator1
		} else if i%3 == 1 {
			fullSchedule[i] = validator2
		} else {
			fullSchedule[i] = validator3
		}
	}

	missedBlockCounters := []int64{10, 20, 30}

	// Serialize
	data, err := SerializeEpochState(
		5,
		5000,
		1000,
		fullSchedule,
		missedBlockCounters,
	)
	if err != nil {
		t.Fatalf("SerializeEpochState failed: %v", err)
	}

	// Deserialize
	epochNumber, epochStartSlot, totalSlotsInEpoch, validatorSchedule, counters, err := DeserializeEpochState(data)
	if err != nil {
		t.Fatalf("DeserializeEpochState failed: %v", err)
	}

	// Verify metadata
	if epochNumber != 5 {
		t.Errorf("epochNumber mismatch: got %d, expected 5", epochNumber)
	}
	if epochStartSlot != 5000 {
		t.Errorf("epochStartSlot mismatch: got %d, expected 5000", epochStartSlot)
	}
	if totalSlotsInEpoch != 1000 {
		t.Errorf("totalSlotsInEpoch mismatch: got %d, expected 1000", totalSlotsInEpoch)
	}

	// Verify schedule
	if len(validatorSchedule) != len(fullSchedule) {
		t.Fatalf("validator schedule length mismatch: got %d, expected %d", len(validatorSchedule), len(fullSchedule))
	}

	for i := range fullSchedule {
		if validatorSchedule[i] != fullSchedule[i] {
			t.Errorf("validator schedule mismatch at slot %d", i)
		}
	}

	// Verify counters
	if len(counters) != len(missedBlockCounters) {
		t.Fatalf("missed block counters length mismatch: got %d, expected %d", len(counters), len(missedBlockCounters))
	}

	for i := range missedBlockCounters {
		if counters[i] != missedBlockCounters[i] {
			t.Errorf("missed block counter mismatch at index %d: got %d, expected %d", i, counters[i], missedBlockCounters[i])
		}
	}
}
