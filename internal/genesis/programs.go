package genesis

import (
	"fmt"
	"log"

	"github.com/poh-blockchain/internal/filestore"
	"github.com/poh-blockchain/internal/quanticscript"
)

// Well-known program IDs as defined in the design document.
var (
	// RuntimeProgramID is the runtime itself (0x00...00)
	RuntimeProgramID = filestore.FileID{}

	// SystemProgramID is the built-in system program (0x00...01)
	SystemProgramID = filestore.FileID{
		0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 1,
	}

	// TokenProgramID is the built-in token program (0x00...02)
	TokenProgramID = filestore.FileID{
		0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 2,
	}

	// StakingProgramID is the built-in staking program (0x00...03)
	StakingProgramID = filestore.FileID{
		0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 3,
	}

	// EpochStateFileID is the well-known FileID for epoch state (0x00...04)
	EpochStateFileID = filestore.FileID{
		0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 4,
	}

	// RewardPoolFileID is the well-known FileID for reward pool (0x00...05)
	RewardPoolFileID = filestore.FileID{
		0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 5,
	}
)

// GenesisValidator represents a validator in the genesis configuration.
type GenesisValidator struct {
	PublicKey   [32]byte
	StakeAmount int64
}

// GenesisConfig holds the configuration for DPoS genesis bootstrap.
type GenesisConfig struct {
	EpochLength       int64
	GenesisValidators []GenesisValidator
}

// LoadBuiltinPrograms creates the System_Program, Token_Program, and optionally Staking_Program files in the
// FileStore if they do not already exist. systemBytecode and tokenBytecode are required;
// stakingBytecode is optional (pass nil to skip loading the Staking_Program).
// These are the compiled .qsb contents, typically embedded by the caller via //go:embed.
//
// SECURITY: This function bypasses normal transaction processing and is ONLY allowed during
// genesis bootstrap before the blockchain becomes operational. After genesis, all file creation
// must go through proper transaction processing.
//
// The function is idempotent — programs that are already present are skipped.
//
// Requirements: 3.6, 3.7, 6.5
func LoadBuiltinPrograms(fs *filestore.FileStore, systemBytecode, tokenBytecode, stakingBytecode []byte) error {
	if err := loadProgram(fs, SystemProgramID, systemBytecode, "System_Program"); err != nil {
		return fmt.Errorf("failed to load System_Program: %w", err)
	}

	if err := loadProgram(fs, TokenProgramID, tokenBytecode, "Token_Program"); err != nil {
		return fmt.Errorf("failed to load Token_Program: %w", err)
	}

	if stakingBytecode != nil && len(stakingBytecode) > 0 {
		if err := loadProgram(fs, StakingProgramID, stakingBytecode, "Staking_Program"); err != nil {
			return fmt.Errorf("failed to load Staking_Program: %w", err)
		}
	}

	return nil
}

// loadProgram creates a single built-in program file in the FileStore.
// This is ONLY allowed during genesis bootstrap before the blockchain is operational.
// After genesis, all file creation must go through proper transaction processing.
func loadProgram(fs *filestore.FileStore, id filestore.FileID, bytecode []byte, name string) error {
	// Idempotent: skip if already present.
	if _, err := fs.GetFile(id); err == nil {
		log.Printf("genesis: %s already loaded at %s", name, id.String())
		return nil
	}

	// Calculate the minimum balance required to cover storage rent.
	storageCost := filestore.CalculateStorageCost(int64(len(bytecode)))
	balance := storageCost + 1000 // small buffer above the minimum

	file := &filestore.File{
		ID:         id,
		Balance:    balance,
		TxManager:  RuntimeProgramID,
		Data:       bytecode,
		Executable: true,
	}

	// GENESIS ONLY: This direct creation is only allowed during genesis bootstrap
	// After genesis, all file creation must go through proper transaction processing
	if _, err := fs.CreateFile(file); err != nil {
		return fmt.Errorf("CreateFile failed: %w", err)
	}

	log.Printf("genesis: loaded %s at %s (bytecode=%d bytes, balance=%d)",
		name, id.String(), len(bytecode), balance)
	return nil
}

// InitializeDPoSGenesis initializes the DPoS genesis state by creating:
// - Epoch State File at EpochStateFileID with epoch 0 data
// - Reward Pool File at RewardPoolFileID with zero balance
// - One Validator Record File per genesis validator (status=active, pre-assigned stake, commission=0)
//
// SECURITY: This function bypasses normal transaction processing and is ONLY allowed during
// genesis bootstrap before the blockchain becomes operational. After genesis, all file creation
// must go through proper transaction processing.
//
// This function is idempotent — if the Epoch State File already exists, it returns early.
// If genesis config has zero validators, it returns an error.
//
// Requirements: 9.1, 9.2, 9.3, 9.4, 9.6
func InitializeDPoSGenesis(fs *filestore.FileStore, config GenesisConfig) error {
	// Requirement 9.6: Reject startup if genesis config has zero validators
	if len(config.GenesisValidators) == 0 {
		return fmt.Errorf("genesis config must have at least one validator")
	}

	// Idempotent: skip if Epoch State File already exists
	if _, err := fs.GetFile(EpochStateFileID); err == nil {
		log.Printf("genesis: DPoS state already initialized")
		return nil
	}

	// Create Validator Record Files and build the validator schedule
	validatorSchedule := make([][32]byte, len(config.GenesisValidators))
	missedBlockCounters := make([]int64, len(config.GenesisValidators))

	for i, genValidator := range config.GenesisValidators {
		// Serialize Validator Record
		// status=1 (active), commission=0, blocksProduced=0, missedBlocks=0, slashedThisEpoch=0
		validatorData, err := quanticscript.SerializeValidatorRecord(
			genValidator.PublicKey[:],
			0,                        // commission
			genValidator.StakeAmount, // totalStake
			1,                        // status (active)
			0,                        // blocksProduced
			0,                        // missedBlocks
			0,                        // slashedThisEpoch
		)
		if err != nil {
			return fmt.Errorf("failed to serialize validator record %d: %w", i, err)
		}

		// Calculate storage cost for validator record
		storageCost := filestore.CalculateStorageCost(int64(len(validatorData)))

		// Create Validator Record File
		// Use a deterministic FileID based on the validator's public key
		validatorFileID := filestore.GenerateFileID(append([]byte("validator:"), genValidator.PublicKey[:]...))

		validatorFile := &filestore.File{
			ID:        validatorFileID,
			Balance:   storageCost + 1000, // storage cost + small buffer
			TxManager: StakingProgramID,
			Data:      validatorData,
		}

		// GENESIS ONLY: This direct creation is only allowed during genesis bootstrap
		// After genesis, all file creation must go through proper transaction processing
		_, err = fs.CreateFile(validatorFile)
		if err != nil {
			return fmt.Errorf("failed to create validator record file %d: %w", i, err)
		}

		validatorSchedule[i] = validatorFileID
		missedBlockCounters[i] = 0

		log.Printf("genesis: created validator record %d at %s (stake=%d)",
			i, validatorFileID.String(), genValidator.StakeAmount)
	}

	// Create Epoch State File
	epochStateData, err := quanticscript.SerializeEpochState(
		0,                   // epochNumber
		0,                   // epochStartSlot
		config.EpochLength,  // totalSlotsInEpoch
		validatorSchedule,   // validatorSchedule
		missedBlockCounters, // missedBlockCounters
	)
	if err != nil {
		return fmt.Errorf("failed to serialize epoch state: %w", err)
	}

	epochStorageCost := filestore.CalculateStorageCost(int64(len(epochStateData)))

	epochStateFile := &filestore.File{
		ID:        EpochStateFileID,
		Balance:   epochStorageCost + 1000, // storage cost + small buffer
		TxManager: StakingProgramID,
		Data:      epochStateData,
	}

	// GENESIS ONLY: This direct creation is only allowed during genesis bootstrap
	// After genesis, all file creation must go through proper transaction processing
	_, err = fs.CreateFile(epochStateFile)
	if err != nil {
		return fmt.Errorf("failed to create epoch state file: %w", err)
	}

	log.Printf("genesis: created epoch state file at %s (epoch=0, validators=%d)",
		EpochStateFileID.String(), len(config.GenesisValidators))

	// Create Reward Pool File
	rewardPoolData, err := quanticscript.SerializeRewardPool(
		0, // balance
		0, // lastDistributedEpoch
	)
	if err != nil {
		return fmt.Errorf("failed to serialize reward pool: %w", err)
	}

	rewardPoolStorageCost := filestore.CalculateStorageCost(int64(len(rewardPoolData)))

	rewardPoolFile := &filestore.File{
		ID:        RewardPoolFileID,
		Balance:   rewardPoolStorageCost + 1000, // storage cost + small buffer
		TxManager: StakingProgramID,
		Data:      rewardPoolData,
	}

	// GENESIS ONLY: This direct creation is only allowed during genesis bootstrap
	// After genesis, all file creation must go through proper transaction processing
	_, err = fs.CreateFile(rewardPoolFile)
	if err != nil {
		return fmt.Errorf("failed to create reward pool file: %w", err)
	}

	log.Printf("genesis: created reward pool file at %s (balance=0)",
		RewardPoolFileID.String())

	return nil
}
