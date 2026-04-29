package consensus

import (
	"fmt"
	"log"
	"time"

	"github.com/poh-blockchain/internal/blockchain"
	"github.com/poh-blockchain/internal/filestore"
	"github.com/poh-blockchain/internal/genesis"
	"github.com/poh-blockchain/internal/quanticscript"
	"github.com/poh-blockchain/internal/runtime"
	"github.com/poh-blockchain/internal/wallet"
)

// GenesisValidator represents a validator in the genesis configuration.
type GenesisValidator struct {
	PublicKey   [32]byte
	StakeAmount int64
}

// ValidatorEntry represents a validator with its stake for schedule computation
type ValidatorEntry struct {
	FileID filestore.FileID
	Stake  int64
}

// GenesisConfig holds the configuration for DPoS genesis bootstrap.
type GenesisConfig struct {
	EpochLength       int64
	GenesisValidators []GenesisValidator
}

// ConsensusManager handles consensus protocol and slot timing
type ConsensusManager struct {
	localValidatorID   filestore.FileID
	localValidatorKeys *wallet.Keypair
	slotDurationMs     int64
	genesisTimestamp   time.Time

	// DPoS fields
	epochLength       int64
	fileStore         *filestore.FileStore
	runtime           *runtime.Runtime
	currentEpoch      int64
	validatorSchedule []filestore.FileID
	genesisValidators []GenesisValidator
}

// NewConsensusManager creates a new consensus manager with 400ms slot duration
func NewConsensusManager(localValidatorID filestore.FileID, localKeys *wallet.Keypair) *ConsensusManager {
	return &ConsensusManager{
		localValidatorID:   localValidatorID,
		localValidatorKeys: localKeys,
		slotDurationMs:     400,
		genesisTimestamp:   time.Now(),
		epochLength:        432000, // default epoch length
		currentEpoch:       0,
	}
}

// NewConsensusManagerWithGenesis creates a new consensus manager with DPoS genesis configuration
func NewConsensusManagerWithGenesis(localValidatorID filestore.FileID, localKeys *wallet.Keypair, config GenesisConfig) *ConsensusManager {
	return &ConsensusManager{
		localValidatorID:   localValidatorID,
		localValidatorKeys: localKeys,
		slotDurationMs:     400,
		genesisTimestamp:   time.Now(),
		epochLength:        config.EpochLength,
		genesisValidators:  config.GenesisValidators,
		currentEpoch:       0,
		validatorSchedule:  make([]filestore.FileID, 0),
	}
}

// GetCurrentSlot calculates the current slot from time elapsed since genesis
func (cm *ConsensusManager) GetCurrentSlot() int64 {
	elapsed := time.Since(cm.genesisTimestamp).Milliseconds()
	return elapsed / cm.slotDurationMs
}

// WaitForSlotStart blocks until the specified slot begins
func (cm *ConsensusManager) WaitForSlotStart(slot int64) {
	targetTime := cm.genesisTimestamp.Add(time.Duration(slot*cm.slotDurationMs) * time.Millisecond)
	now := time.Now()

	if targetTime.After(now) {
		time.Sleep(targetTime.Sub(now))
	}
}

// IsLeader returns true if this node is the leader for the given slot
// With stake-weighted scheduling, checks if the local validator is scheduled for this slot
// Requirements: 3.1, 3.2, 3.3, 3.4, 3.5
func (cm *ConsensusManager) IsLeader(slot int64) bool {
	// Observer mode: no validator identity configured
	if cm.localValidatorID == (filestore.FileID{}) {
		return false
	}

	// Get the scheduled validator for this slot
	scheduledValidator := cm.GetScheduledValidator(slot)

	// Return true if the scheduled validator matches the local validator
	return scheduledValidator == cm.localValidatorID
}

// ValidateBlock performs basic block validation
func (cm *ConsensusManager) ValidateBlock(block blockchain.Block) error {
	// Verify block slot is within acceptable range
	// Allow blocks from recent past and near future to account for network delays
	// and clock skew between nodes
	currentSlot := cm.GetCurrentSlot()
	const slotTolerance = int64(100) // Allow 100 slots (~40 seconds) of tolerance

	if block.Header.Slot > currentSlot+slotTolerance {
		return fmt.Errorf("block slot %d is too far in the future (current slot: %d, max allowed: %d)",
			block.Header.Slot, currentSlot, currentSlot+slotTolerance)
	}

	// Verify block contains at least 64 ticks worth of hash operations
	// Each tick is 12,500 hashes, so minimum is 64 * 12,500 = 800,000 hashes
	const minHashesPerBlock = 64 * 12500

	totalHashes := int64(0)
	for _, entry := range block.Entries {
		totalHashes += entry.NumHashes
	}

	if totalHashes < minHashesPerBlock {
		return fmt.Errorf("block contains insufficient hash operations: %d (minimum: %d)", totalHashes, minHashesPerBlock)
	}

	return nil
}

// GetCurrentEpoch returns the current epoch number
// Requirement 3.3, 9.1
func (cm *ConsensusManager) GetCurrentEpoch() int64 {
	return cm.currentEpoch
}

// IsEpochBoundary returns true if the given slot is an epoch boundary
// Requirement 3.3, 9.4
func (cm *ConsensusManager) IsEpochBoundary(slot int64) bool {
	return slot%cm.epochLength == 0
}

// GetScheduledValidator returns the FileID of the validator scheduled for the given slot
// Requirement 3.4, 3.5
func (cm *ConsensusManager) GetScheduledValidator(slot int64) filestore.FileID {
	if len(cm.validatorSchedule) == 0 {
		return filestore.FileID{}
	}

	// Get position within current epoch
	slotInEpoch := slot % cm.epochLength
	if slotInEpoch >= int64(len(cm.validatorSchedule)) {
		return filestore.FileID{}
	}

	return cm.validatorSchedule[slotInEpoch]
}

// RecordMissedBlock updates the missed-block counter for a validator
// Requirement 3.6, 9.4
func (cm *ConsensusManager) RecordMissedBlock(slot int64, validatorID filestore.FileID) error {
	if cm.fileStore == nil {
		return fmt.Errorf("fileStore not initialized")
	}

	// Get the validator record
	validatorFile, err := cm.fileStore.GetFile(validatorID)
	if err != nil {
		return fmt.Errorf("failed to get validator record: %w", err)
	}

	// Deserialize validator record
	pubkey, commission, totalStake, status, blocksProduced, missedBlocks, slashedThisEpoch, err := quanticscript.DeserializeValidatorRecord(validatorFile.Data)
	if err != nil {
		return fmt.Errorf("failed to deserialize validator record: %w", err)
	}

	// Increment missed blocks counter
	missedBlocks++

	// Serialize updated validator record
	updatedData, err := quanticscript.SerializeValidatorRecord(
		pubkey,
		commission,
		totalStake,
		status,
		blocksProduced,
		missedBlocks,
		slashedThisEpoch,
	)
	if err != nil {
		return fmt.Errorf("failed to serialize validator record: %w", err)
	}

	// Update the file
	validatorFile.Data = updatedData
	if err := cm.fileStore.UpdateFile(validatorID, validatorFile); err != nil {
		return fmt.Errorf("failed to update validator record: %w", err)
	}

	// Also update the Epoch State File's missed-block counters
	epochStateFile, err := cm.fileStore.GetFile(genesis.EpochStateFileID)
	if err != nil {
		return fmt.Errorf("failed to get epoch state file: %w", err)
	}

	epochNumber, epochStartSlot, totalSlotsInEpoch, validatorSchedule, missedBlockCounters, err := quanticscript.DeserializeEpochState(epochStateFile.Data)
	if err != nil {
		return fmt.Errorf("failed to deserialize epoch state: %w", err)
	}

	// Build unique validator list to find the correct index for missed block counters
	uniqueValidators := make(map[filestore.FileID]int)
	var orderedValidators []filestore.FileID

	for _, schedValidatorBytes := range validatorSchedule {
		schedValidatorID := filestore.FileID(schedValidatorBytes)
		if _, exists := uniqueValidators[schedValidatorID]; !exists {
			uniqueValidators[schedValidatorID] = len(orderedValidators)
			orderedValidators = append(orderedValidators, schedValidatorID)
		}
	}

	// Find the validator index in the unique validator list
	validatorIndex, exists := uniqueValidators[validatorID]
	if exists && validatorIndex < len(missedBlockCounters) {
		missedBlockCounters[validatorIndex]++
	}

	// Serialize updated epoch state
	updatedEpochData, err := quanticscript.SerializeEpochState(
		epochNumber,
		epochStartSlot,
		totalSlotsInEpoch,
		validatorSchedule,
		missedBlockCounters,
	)
	if err != nil {
		return fmt.Errorf("failed to serialize epoch state: %w", err)
	}

	// Update the epoch state file
	epochStateFile.Data = updatedEpochData
	if err := cm.fileStore.UpdateFile(genesis.EpochStateFileID, epochStateFile); err != nil {
		return fmt.Errorf("failed to update epoch state file: %w", err)
	}

	log.Printf("consensus: recorded missed block for validator %s (total missed: %d)",
		validatorID.String(), missedBlocks)

	return nil
}

// SetFileStore sets the FileStore reference for DPoS operations
func (cm *ConsensusManager) SetFileStore(fs *filestore.FileStore) {
	cm.fileStore = fs
}

// SetRuntime sets the Runtime reference for DPoS operations
func (cm *ConsensusManager) SetRuntime(rt *runtime.Runtime) {
	cm.runtime = rt
}

// SetValidatorSchedule sets the validator schedule for the current epoch
func (cm *ConsensusManager) SetValidatorSchedule(schedule []filestore.FileID) {
	cm.validatorSchedule = schedule
}

// SetCurrentEpoch sets the current epoch number
func (cm *ConsensusManager) SetCurrentEpoch(epoch int64) {
	cm.currentEpoch = epoch
}

// enumerateActiveValidators reads all Validator Records from FileStore and filters by status=active and stake>=1,000,000 electrons
// Returns a list of ValidatorEntry structs containing FileID and stake for each active validator
// Requirements: 2.1, 2.2
func (cm *ConsensusManager) enumerateActiveValidators() ([]ValidatorEntry, error) {
	if cm.fileStore == nil {
		return nil, fmt.Errorf("fileStore not initialized")
	}

	// Get all file IDs from the FileStore
	allFileIDs, err := cm.fileStore.GetAllFileIDs()
	if err != nil {
		return nil, fmt.Errorf("failed to get all file IDs: %w", err)
	}

	var activeValidators []ValidatorEntry
	const minStake = int64(1000000) // 1 Neon = 1,000,000 electrons

	// Iterate through all files and identify Validator Records
	for _, fileID := range allFileIDs {
		// Try to get the file
		file, err := cm.fileStore.GetFile(fileID)
		if err != nil {
			continue // Skip files that can't be read
		}

		// Check if this is a Validator Record by attempting to deserialize it
		// Validator Records are exactly 66 bytes
		if len(file.Data) != 66 {
			continue
		}

		// Try to deserialize as a Validator Record
		_, _, totalStake, status, _, _, _, err := quanticscript.DeserializeValidatorRecord(file.Data)
		if err != nil {
			continue // Not a valid Validator Record
		}

		// Filter by status=active (1) and stake >= 1,000,000 electrons
		if status == 1 && totalStake >= minStake {
			activeValidators = append(activeValidators, ValidatorEntry{
				FileID: fileID,
				Stake:  totalStake,
			})
		}
	}

	return activeValidators, nil
}

// ComputeValidatorSchedule computes a deterministic stake-weighted validator schedule
// Uses a weighted-random algorithm seeded with the provided seed (typically the last block hash)
// Requirements: 3.3, 9.4
func (cm *ConsensusManager) ComputeValidatorSchedule(epochSeed []byte, validators []ValidatorEntry) []filestore.FileID {
	if len(validators) == 0 {
		return []filestore.FileID{}
	}

	schedule := make([]filestore.FileID, cm.epochLength)

	// Calculate total stake
	totalStake := int64(0)
	for _, v := range validators {
		totalStake += v.Stake
	}

	if totalStake == 0 {
		return schedule
	}

	// Use a deterministic PRNG seeded with epochSeed
	// We'll use a simple linear congruential generator for determinism
	seed := uint64(0)
	if len(epochSeed) >= 8 {
		// Use first 8 bytes of seed as initial value
		for i := 0; i < 8; i++ {
			seed = (seed << 8) | uint64(epochSeed[i])
		}
	}

	// Fill schedule with stake-weighted random selection
	for slot := int64(0); slot < cm.epochLength; slot++ {
		// Generate next random number
		seed = seed*6364136223846793005 + 1442695040888963407 // LCG constants
		randomValue := seed % uint64(totalStake)

		// Find validator corresponding to this random value
		accumulated := int64(0)
		for _, v := range validators {
			accumulated += v.Stake
			if int64(randomValue) < accumulated {
				schedule[slot] = v.FileID
				break
			}
		}
	}

	return schedule
}

// ProcessEpochBoundary processes the epoch boundary
// Updates validator statuses based on stake threshold and triggers reward distribution
// Requirements: 3.1, 3.2, 4.1, 4.2
func (cm *ConsensusManager) ProcessEpochBoundary(slot int64) error {
	if !cm.IsEpochBoundary(slot) {
		return fmt.Errorf("slot %d is not an epoch boundary", slot)
	}

	if cm.fileStore == nil {
		return fmt.Errorf("fileStore not initialized")
	}

	if cm.runtime == nil {
		return fmt.Errorf("runtime not initialized")
	}

	// This will be implemented in task 4.5 with full validator status updates
	// and reward distribution triggering
	log.Printf("consensus: processing epoch boundary at slot %d", slot)

	// Increment epoch
	cm.currentEpoch++

	return nil
}

// ProcessEpochBoundaryWithHash processes the epoch boundary with schedule recalculation
// Enumerates active validators, computes new leader schedule, and persists to Epoch State File
// Requirements: 4.1, 4.2, 4.3, 4.4
func (cm *ConsensusManager) ProcessEpochBoundaryWithHash(slot int64, lastBlockHash [32]byte) error {
	if !cm.IsEpochBoundary(slot) {
		return fmt.Errorf("slot %d is not an epoch boundary", slot)
	}

	if cm.fileStore == nil {
		return fmt.Errorf("fileStore not initialized")
	}

	log.Printf("consensus: processing epoch boundary at slot %d", slot)

	// 1. Enumerate all active validators with stake >= 1 Neon
	validators, err := cm.enumerateActiveValidators()
	if err != nil {
		return fmt.Errorf("failed to enumerate active validators: %w", err)
	}

	if len(validators) == 0 {
		log.Printf("consensus: warning - no active validators found at epoch boundary")
		// Keep existing schedule if no validators found
		cm.currentEpoch++
		return nil
	}

	// Calculate total stake for logging
	totalStake := int64(0)
	for _, v := range validators {
		totalStake += v.Stake
	}

	log.Printf("consensus: epoch %d -> %d: %d active validators, total stake: %d electrons",
		cm.currentEpoch, cm.currentEpoch+1, len(validators), totalStake)

	// 2. Compute new leader schedule using last block hash as seed
	newSchedule := cm.ComputeValidatorSchedule(lastBlockHash[:], validators)

	// 3. Update validatorSchedule and currentEpoch fields
	cm.validatorSchedule = newSchedule
	cm.currentEpoch++

	// 4. Persist new schedule to Epoch State File using compact format
	if err := cm.persistEpochState(slot); err != nil {
		return fmt.Errorf("failed to persist epoch state: %w", err)
	}

	log.Printf("consensus: computed new schedule for epoch %d (%d slots)",
		cm.currentEpoch, len(newSchedule))

	return nil
}

// persistEpochState persists the current epoch state to the Epoch State File
func (cm *ConsensusManager) persistEpochState(epochStartSlot int64) error {
	if cm.fileStore == nil {
		return fmt.Errorf("fileStore not initialized")
	}

	// Convert []filestore.FileID to [][32]byte for serialization
	validatorSchedule := make([][32]byte, len(cm.validatorSchedule))
	for i, fileID := range cm.validatorSchedule {
		copy(validatorSchedule[i][:], fileID[:])
	}

	// Build unique validator list for missed block counters
	uniqueValidators := make(map[filestore.FileID]int)
	var orderedValidators []filestore.FileID

	for _, fileID := range cm.validatorSchedule {
		if _, exists := uniqueValidators[fileID]; !exists {
			uniqueValidators[fileID] = len(orderedValidators)
			orderedValidators = append(orderedValidators, fileID)
		}
	}

	// Initialize missed block counters (all zeros for new epoch)
	missedBlockCounters := make([]int64, len(orderedValidators))

	// Serialize epoch state
	epochStateData, err := quanticscript.SerializeEpochState(
		cm.currentEpoch,
		epochStartSlot,
		cm.epochLength,
		validatorSchedule,
		missedBlockCounters,
	)
	if err != nil {
		return fmt.Errorf("failed to serialize epoch state: %w", err)
	}

	// Check if Epoch State File exists
	epochStateFile, err := cm.fileStore.GetFile(genesis.EpochStateFileID)
	if err != nil {
		// File doesn't exist, create it
		storageCost := filestore.CalculateStorageCost(int64(len(epochStateData)))
		epochStateFile = &filestore.File{
			ID:        genesis.EpochStateFileID,
			Balance:   storageCost + 1000, // Add some buffer
			TxManager: filestore.FileID{},
			Data:      epochStateData,
		}
		_, err = cm.fileStore.CreateFile(epochStateFile)
		if err != nil {
			return fmt.Errorf("failed to create Epoch State File: %w", err)
		}
	} else {
		// File exists, update it
		epochStateFile.Data = epochStateData
		if err := cm.fileStore.UpdateFile(genesis.EpochStateFileID, epochStateFile); err != nil {
			return fmt.Errorf("failed to update Epoch State File: %w", err)
		}
	}

	return nil
}

// InitializeGenesis initializes the DPoS genesis state
// Calls InitializeDPoSGenesis if Epoch State File is absent; restores from file if present
// Logs warning and re-initializes from slot 0 if Epoch State File is corrupted
// Requirements: 9.2, 9.3, 9.4
func (cm *ConsensusManager) InitializeGenesis(config GenesisConfig) error {
	if cm.fileStore == nil {
		return fmt.Errorf("fileStore not initialized")
	}

	// Check if Epoch State File already exists
	epochStateFile, err := cm.fileStore.GetFile(genesis.EpochStateFileID)
	if err == nil && epochStateFile != nil {
		// Epoch State File exists, try to restore from it
		epochNumber, _, totalSlotsInEpoch, validatorSchedule, _, err := quanticscript.DeserializeEpochState(epochStateFile.Data)
		if err != nil {
			// Epoch State File is corrupted, log warning and re-initialize
			log.Printf("consensus: Epoch State File corrupted: %v, re-initializing from slot 0", err)
			return cm.reinitializeGenesisFromConfig(config)
		}

		// Convert [][32]byte to []filestore.FileID
		schedule := make([]filestore.FileID, len(validatorSchedule))
		for i, fileIDBytes := range validatorSchedule {
			copy(schedule[i][:], fileIDBytes[:])
		}

		// Successfully restored epoch state
		cm.currentEpoch = epochNumber
		cm.validatorSchedule = schedule
		cm.epochLength = totalSlotsInEpoch

		log.Printf("consensus: restored epoch state from file (epoch=%d, validators=%d)",
			epochNumber, len(schedule))
		return nil
	}

	// Epoch State File does not exist, initialize from config
	return cm.reinitializeGenesisFromConfig(config)
}

// reinitializeGenesisFromConfig initializes DPoS genesis from the provided config
func (cm *ConsensusManager) reinitializeGenesisFromConfig(config GenesisConfig) error {
	// Call the genesis package's InitializeDPoSGenesis
	genesisConfig := genesis.GenesisConfig{
		EpochLength:       config.EpochLength,
		GenesisValidators: make([]genesis.GenesisValidator, len(config.GenesisValidators)),
	}

	for i, gv := range config.GenesisValidators {
		genesisConfig.GenesisValidators[i] = genesis.GenesisValidator{
			PublicKey:   gv.PublicKey,
			StakeAmount: gv.StakeAmount,
		}
	}

	if err := genesis.InitializeDPoSGenesis(cm.fileStore, genesisConfig); err != nil {
		return fmt.Errorf("failed to initialize DPoS genesis: %w", err)
	}

	// After initialization, restore the epoch state
	epochStateFile, err := cm.fileStore.GetFile(genesis.EpochStateFileID)
	if err != nil {
		return fmt.Errorf("failed to read epoch state file after initialization: %w", err)
	}

	epochNumber, _, totalSlotsInEpoch, validatorSchedule, _, err := quanticscript.DeserializeEpochState(epochStateFile.Data)
	if err != nil {
		return fmt.Errorf("failed to deserialize epoch state: %w", err)
	}

	// Convert [][32]byte to []filestore.FileID
	schedule := make([]filestore.FileID, len(validatorSchedule))
	for i, fileIDBytes := range validatorSchedule {
		copy(schedule[i][:], fileIDBytes[:])
	}

	cm.currentEpoch = epochNumber
	cm.validatorSchedule = schedule
	cm.epochLength = totalSlotsInEpoch

	log.Printf("consensus: initialized DPoS genesis (epoch=%d, validators=%d)",
		epochNumber, len(schedule))
	return nil
}

// ProcessSlotSkip processes a slot skip when the scheduled validator misses their slot
// Records the missed block for the scheduled validator
// Requirements: 5.1, 5.2, 5.5
func (cm *ConsensusManager) ProcessSlotSkip(slot int64) error {
	// Get the scheduled validator for this slot
	scheduledValidator := cm.GetScheduledValidator(slot)

	// If no validator is scheduled, nothing to do
	if scheduledValidator == (filestore.FileID{}) {
		return nil
	}

	// Record the missed block for the scheduled validator
	return cm.RecordMissedBlock(slot, scheduledValidator)
}

// ShouldProduceBlock returns true if the local node should produce a block for the given slot
// This ensures local node does NOT produce a block if not the scheduled leader
// Requirements: 5.1, 5.2, 5.5
func (cm *ConsensusManager) ShouldProduceBlock(slot int64) bool {
	return cm.IsLeader(slot)
}

// GetLocalValidatorID returns the local validator's FileID
func (cm *ConsensusManager) GetLocalValidatorID() filestore.FileID {
	return cm.localValidatorID
}

// WaitForBlockOrTimeout waits for a block to be received within the specified timeout
// Returns true if a block was received, false if timeout occurred
// This is used to implement the 200ms wait logic for scheduled validator blocks
func (cm *ConsensusManager) WaitForBlockOrTimeout(slot int64, timeoutMs int64) bool {
	// This is a placeholder implementation - in a real system, this would
	// integrate with the network layer to wait for block reception
	// For now, we'll simulate that no block is received (timeout)
	return false
}

// RecordBlockProduction increments the blocksProducedThisEpoch counter in a validator's Validator Record
// Requirement 8.4
func (cm *ConsensusManager) RecordBlockProduction(validatorID filestore.FileID) error {
	if cm.fileStore == nil {
		return fmt.Errorf("fileStore not initialized")
	}

	// Get the validator record
	validatorFile, err := cm.fileStore.GetFile(validatorID)
	if err != nil {
		return fmt.Errorf("failed to get validator record: %w", err)
	}

	// Deserialize validator record
	pubkey, commission, totalStake, status, blocksProduced, missedBlocks, slashedThisEpoch, err := quanticscript.DeserializeValidatorRecord(validatorFile.Data)
	if err != nil {
		return fmt.Errorf("failed to deserialize validator record: %w", err)
	}

	// Increment blocks produced counter
	blocksProduced++

	// Serialize updated validator record
	updatedData, err := quanticscript.SerializeValidatorRecord(
		pubkey,
		commission,
		totalStake,
		status,
		blocksProduced,
		missedBlocks,
		slashedThisEpoch,
	)
	if err != nil {
		return fmt.Errorf("failed to serialize validator record: %w", err)
	}

	// Update the file
	validatorFile.Data = updatedData
	if err := cm.fileStore.UpdateFile(validatorID, validatorFile); err != nil {
		return fmt.Errorf("failed to update validator record: %w", err)
	}

	log.Printf("consensus: recorded block production for validator %s (total blocks: %d)",
		validatorID.String(), blocksProduced)

	return nil
}

// ResetBlockProductionCounter resets the blocksProducedThisEpoch counter to zero for a single validator
// Called at epoch boundaries
// Requirement 8.4
func (cm *ConsensusManager) ResetBlockProductionCounter(validatorID filestore.FileID) error {
	if cm.fileStore == nil {
		return fmt.Errorf("fileStore not initialized")
	}

	// Get the validator record
	validatorFile, err := cm.fileStore.GetFile(validatorID)
	if err != nil {
		return fmt.Errorf("failed to get validator record: %w", err)
	}

	// Deserialize validator record
	pubkey, commission, totalStake, status, _, missedBlocks, slashedThisEpoch, err := quanticscript.DeserializeValidatorRecord(validatorFile.Data)
	if err != nil {
		return fmt.Errorf("failed to deserialize validator record: %w", err)
	}

	// Reset blocks produced counter to zero
	blocksProduced := int64(0)

	// Serialize updated validator record
	updatedData, err := quanticscript.SerializeValidatorRecord(
		pubkey,
		commission,
		totalStake,
		status,
		blocksProduced,
		missedBlocks,
		slashedThisEpoch,
	)
	if err != nil {
		return fmt.Errorf("failed to serialize validator record: %w", err)
	}

	// Update the file
	validatorFile.Data = updatedData
	if err := cm.fileStore.UpdateFile(validatorID, validatorFile); err != nil {
		return fmt.Errorf("failed to update validator record: %w", err)
	}

	return nil
}

// ResetAllBlockProductionCounters resets the blocksProducedThisEpoch counter to zero for all validators
// Called at epoch boundaries
// Requirement 8.4
func (cm *ConsensusManager) ResetAllBlockProductionCounters() error {
	if cm.fileStore == nil {
		return fmt.Errorf("fileStore not initialized")
	}

	// Get all file IDs from the FileStore
	allFileIDs, err := cm.fileStore.GetAllFileIDs()
	if err != nil {
		return fmt.Errorf("failed to get all file IDs: %w", err)
	}

	// Iterate through all files and reset block production counters for Validator Records
	for _, fileID := range allFileIDs {
		// Try to get the file
		file, err := cm.fileStore.GetFile(fileID)
		if err != nil {
			continue // Skip files that can't be read
		}

		// Check if this is a Validator Record by attempting to deserialize it
		// Validator Records are exactly 66 bytes
		if len(file.Data) != 66 {
			continue
		}

		// Try to deserialize as a Validator Record
		pubkey, commission, totalStake, status, _, missedBlocks, slashedThisEpoch, err := quanticscript.DeserializeValidatorRecord(file.Data)
		if err != nil {
			continue // Not a valid Validator Record
		}

		// Reset blocks produced counter to zero
		blocksProduced := int64(0)

		// Serialize updated validator record
		updatedData, err := quanticscript.SerializeValidatorRecord(
			pubkey,
			commission,
			totalStake,
			status,
			blocksProduced,
			missedBlocks,
			slashedThisEpoch,
		)
		if err != nil {
			log.Printf("consensus: warning - failed to serialize validator record %s: %v", fileID.String(), err)
			continue
		}

		// Update the file
		file.Data = updatedData
		if err := cm.fileStore.UpdateFile(fileID, file); err != nil {
			log.Printf("consensus: warning - failed to update validator record %s: %v", fileID.String(), err)
			continue
		}
	}

	log.Printf("consensus: reset block production counters for all validators at epoch boundary")

	return nil
}
