package genesis

import (
	"os"
	"testing"

	"github.com/poh-blockchain/internal/filestore"
	"github.com/poh-blockchain/internal/quanticscript"
)

// TestInitializeDPoSGenesis verifies that InitializeDPoSGenesis creates the correct files
// with proper data and balances for epoch 0 bootstrap.
// Requirements: 9.1, 9.2, 9.3, 9.4, 9.6
func TestInitializeDPoSGenesis(t *testing.T) {
	// Create a temporary FileStore
	fs, err := filestore.NewFileStore(t.TempDir())
	if err != nil {
		t.Fatalf("failed to create FileStore: %v", err)
	}
	defer fs.Close()

	// Create genesis config with 3 validators
	genesisValidators := []GenesisValidator{
		{
			PublicKey:   [32]byte{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1},
			StakeAmount: 1000000,
		},
		{
			PublicKey:   [32]byte{2, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 2},
			StakeAmount: 2000000,
		},
		{
			PublicKey:   [32]byte{3, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 3},
			StakeAmount: 1500000,
		},
	}

	config := GenesisConfig{
		EpochLength:       432000,
		GenesisValidators: genesisValidators,
	}

	// Initialize DPoS genesis
	err = InitializeDPoSGenesis(fs, config)
	if err != nil {
		t.Fatalf("InitializeDPoSGenesis failed: %v", err)
	}

	// Verify Epoch State File was created
	epochStateFile, err := fs.GetFile(EpochStateFileID)
	if err != nil {
		t.Fatalf("Epoch State File not found: %v", err)
	}

	// Verify Epoch State File data
	epochNumber, epochStartSlot, totalSlotsInEpoch, validatorSchedule, missedBlockCounters, err := quanticscript.DeserializeEpochState(epochStateFile.Data)
	if err != nil {
		t.Fatalf("failed to deserialize Epoch State: %v", err)
	}

	if epochNumber != 0 {
		t.Errorf("expected epoch 0, got %d", epochNumber)
	}
	if epochStartSlot != 0 {
		t.Errorf("expected epoch start slot 0, got %d", epochStartSlot)
	}
	if totalSlotsInEpoch != 432000 {
		t.Errorf("expected total slots 432000, got %d", totalSlotsInEpoch)
	}
	if int64(len(validatorSchedule)) != 3 {
		t.Errorf("expected 3 validators in schedule, got %d", len(validatorSchedule))
	}
	if int64(len(missedBlockCounters)) != 3 {
		t.Errorf("expected 3 missed block counters, got %d", len(missedBlockCounters))
	}

	// Verify Reward Pool File was created
	rewardPoolFile, err := fs.GetFile(RewardPoolFileID)
	if err != nil {
		t.Fatalf("Reward Pool File not found: %v", err)
	}

	// Verify Reward Pool File data
	balance, lastDistributedEpoch, err := quanticscript.DeserializeRewardPool(rewardPoolFile.Data)
	if err != nil {
		t.Fatalf("failed to deserialize Reward Pool: %v", err)
	}

	if balance != 0 {
		t.Errorf("expected Reward Pool balance 0, got %d", balance)
	}
	if lastDistributedEpoch != 0 {
		t.Errorf("expected last distributed epoch 0, got %d", lastDistributedEpoch)
	}

	// Verify Validator Record Files were created
	for i, genValidator := range genesisValidators {
		// Generate validator FileID from pubkey
		validatorFileID := validatorSchedule[i]

		validatorFile, err := fs.GetFile(validatorFileID)
		if err != nil {
			t.Fatalf("Validator Record File %d not found: %v", i, err)
		}

		// Verify Validator Record data
		pubkey, commission, totalStake, status, blocksProduced, missedBlocks, slashedThisEpoch, err := quanticscript.DeserializeValidatorRecord(validatorFile.Data)
		if err != nil {
			t.Fatalf("failed to deserialize Validator Record %d: %v", i, err)
		}

		pubkeyArray := [32]byte{}
		copy(pubkeyArray[:], pubkey)
		if pubkeyArray != genValidator.PublicKey {
			t.Errorf("validator %d: pubkey mismatch", i)
		}
		if commission != 0 {
			t.Errorf("validator %d: expected commission 0, got %d", i, commission)
		}
		if totalStake != genValidator.StakeAmount {
			t.Errorf("validator %d: expected stake %d, got %d", i, genValidator.StakeAmount, totalStake)
		}
		if status != 1 { // 1 = active
			t.Errorf("validator %d: expected status 1 (active), got %d", i, status)
		}
		if blocksProduced != 0 {
			t.Errorf("validator %d: expected blocks produced 0, got %d", i, blocksProduced)
		}
		if missedBlocks != 0 {
			t.Errorf("validator %d: expected missed blocks 0, got %d", i, missedBlocks)
		}
		if slashedThisEpoch != 0 {
			t.Errorf("validator %d: expected slashed false, got %d", i, slashedThisEpoch)
		}

		// Verify File balance covers storage cost
		expectedStorageCost := filestore.CalculateStorageCost(int64(len(validatorFile.Data)))
		if validatorFile.Balance < expectedStorageCost {
			t.Errorf("validator %d: balance %d is less than storage cost %d", i, validatorFile.Balance, expectedStorageCost)
		}

		// Verify TxManager is StakingProgramID
		if validatorFile.TxManager != StakingProgramID {
			t.Errorf("validator %d: expected TxManager %s, got %s", i, StakingProgramID.String(), validatorFile.TxManager.String())
		}
	}

	// Verify Epoch State File balance covers storage cost
	expectedEpochStorageCost := filestore.CalculateStorageCost(int64(len(epochStateFile.Data)))
	if epochStateFile.Balance < expectedEpochStorageCost {
		t.Errorf("Epoch State File: balance %d is less than storage cost %d", epochStateFile.Balance, expectedEpochStorageCost)
	}

	// Verify Reward Pool File balance covers storage cost
	expectedRewardPoolStorageCost := filestore.CalculateStorageCost(int64(len(rewardPoolFile.Data)))
	if rewardPoolFile.Balance < expectedRewardPoolStorageCost {
		t.Errorf("Reward Pool File: balance %d is less than storage cost %d", rewardPoolFile.Balance, expectedRewardPoolStorageCost)
	}
}

// TestInitializeDPoSGenesisZeroValidators verifies that InitializeDPoSGenesis rejects
// a genesis config with zero validators.
// Requirement: 9.6
func TestInitializeDPoSGenesisZeroValidators(t *testing.T) {
	fs, err := filestore.NewFileStore(t.TempDir())
	if err != nil {
		t.Fatalf("failed to create FileStore: %v", err)
	}
	defer fs.Close()

	config := GenesisConfig{
		EpochLength:       432000,
		GenesisValidators: []GenesisValidator{},
	}

	err = InitializeDPoSGenesis(fs, config)
	if err == nil {
		t.Fatalf("expected error for zero validators, got nil")
	}
}

// TestInitializeDPoSGenesisIdempotent verifies that InitializeDPoSGenesis is idempotent
// and skips creation if files already exist.
func TestInitializeDPoSGenesisIdempotent(t *testing.T) {
	fs, err := filestore.NewFileStore(t.TempDir())
	if err != nil {
		t.Fatalf("failed to create FileStore: %v", err)
	}
	defer fs.Close()

	genesisValidators := []GenesisValidator{
		{
			PublicKey:   [32]byte{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1},
			StakeAmount: 1000000,
		},
	}

	config := GenesisConfig{
		EpochLength:       432000,
		GenesisValidators: genesisValidators,
	}

	// First initialization
	err = InitializeDPoSGenesis(fs, config)
	if err != nil {
		t.Fatalf("first InitializeDPoSGenesis failed: %v", err)
	}

	// Get the first Epoch State File
	epochStateFile1, err := fs.GetFile(EpochStateFileID)
	if err != nil {
		t.Fatalf("failed to get Epoch State File after first init: %v", err)
	}

	// Second initialization (should be idempotent)
	err = InitializeDPoSGenesis(fs, config)
	if err != nil {
		t.Fatalf("second InitializeDPoSGenesis failed: %v", err)
	}

	// Get the Epoch State File again
	epochStateFile2, err := fs.GetFile(EpochStateFileID)
	if err != nil {
		t.Fatalf("failed to get Epoch State File after second init: %v", err)
	}

	// Verify they are the same
	if epochStateFile1.ID != epochStateFile2.ID {
		t.Errorf("Epoch State File ID changed after second init")
	}
}

// TestLoadBuiltinProgramsWithStaking verifies that LoadBuiltinPrograms correctly loads
// the Staking Program when bytecode is provided.
// Requirement: 6.5
func TestLoadBuiltinProgramsWithStaking(t *testing.T) {
	fs, err := filestore.NewFileStore(t.TempDir())
	if err != nil {
		t.Fatalf("failed to create FileStore: %v", err)
	}
	defer fs.Close()

	// Create dummy bytecode for testing
	systemBytecode := []byte("system_program_bytecode")
	tokenBytecode := []byte("token_program_bytecode")
	stakingBytecode := []byte("staking_program_bytecode")

	// Load all three programs
	err = LoadBuiltinPrograms(fs, systemBytecode, tokenBytecode, stakingBytecode)
	if err != nil {
		t.Fatalf("LoadBuiltinPrograms failed: %v", err)
	}

	// Verify System Program was loaded
	systemFile, err := fs.GetFile(SystemProgramID)
	if err != nil {
		t.Fatalf("System Program not found: %v", err)
	}
	if len(systemFile.Data) != len(systemBytecode) {
		t.Errorf("System Program bytecode mismatch")
	}

	// Verify Token Program was loaded
	tokenFile, err := fs.GetFile(TokenProgramID)
	if err != nil {
		t.Fatalf("Token Program not found: %v", err)
	}
	if len(tokenFile.Data) != len(tokenBytecode) {
		t.Errorf("Token Program bytecode mismatch")
	}

	// Verify Staking Program was loaded
	stakingFile, err := fs.GetFile(StakingProgramID)
	if err != nil {
		t.Fatalf("Staking Program not found: %v", err)
	}
	if len(stakingFile.Data) != len(stakingBytecode) {
		t.Errorf("Staking Program bytecode mismatch")
	}

	// Verify all programs are executable
	if !systemFile.Executable {
		t.Errorf("System Program should be executable")
	}
	if !tokenFile.Executable {
		t.Errorf("Token Program should be executable")
	}
	if !stakingFile.Executable {
		t.Errorf("Staking Program should be executable")
	}
}

// TestLoadBuiltinProgramsWithoutStaking verifies that LoadBuiltinPrograms skips
// the Staking Program when bytecode is nil.
// Requirement: 6.5
func TestLoadBuiltinProgramsWithoutStaking(t *testing.T) {
	fs, err := filestore.NewFileStore(t.TempDir())
	if err != nil {
		t.Fatalf("failed to create FileStore: %v", err)
	}
	defer fs.Close()

	// Create dummy bytecode for testing
	systemBytecode := []byte("system_program_bytecode")
	tokenBytecode := []byte("token_program_bytecode")

	// Load only System and Token programs (no Staking)
	err = LoadBuiltinPrograms(fs, systemBytecode, tokenBytecode, nil)
	if err != nil {
		t.Fatalf("LoadBuiltinPrograms failed: %v", err)
	}

	// Verify System Program was loaded
	_, err = fs.GetFile(SystemProgramID)
	if err != nil {
		t.Fatalf("System Program not found: %v", err)
	}

	// Verify Token Program was loaded
	_, err = fs.GetFile(TokenProgramID)
	if err != nil {
		t.Fatalf("Token Program not found: %v", err)
	}

	// Verify Staking Program was NOT loaded
	_, err = fs.GetFile(StakingProgramID)
	if err == nil {
		t.Fatalf("Staking Program should not be loaded when bytecode is nil")
	}
}

// TestEpochStateFileStructure verifies that the Epoch State File has the correct structure
// and can be properly deserialized.
// Requirements: 9.2, 10.2
func TestEpochStateFileStructure(t *testing.T) {
	fs, err := filestore.NewFileStore(t.TempDir())
	if err != nil {
		t.Fatalf("failed to create FileStore: %v", err)
	}
	defer fs.Close()

	genesisValidators := []GenesisValidator{
		{
			PublicKey:   [32]byte{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1},
			StakeAmount: 1000000,
		},
		{
			PublicKey:   [32]byte{2, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 2},
			StakeAmount: 2000000,
		},
	}

	config := GenesisConfig{
		EpochLength:       432000,
		GenesisValidators: genesisValidators,
	}

	err = InitializeDPoSGenesis(fs, config)
	if err != nil {
		t.Fatalf("InitializeDPoSGenesis failed: %v", err)
	}

	epochStateFile, err := fs.GetFile(EpochStateFileID)
	if err != nil {
		t.Fatalf("Epoch State File not found: %v", err)
	}

	// Verify File properties
	if epochStateFile.ID != EpochStateFileID {
		t.Errorf("Epoch State File ID mismatch")
	}
	if epochStateFile.TxManager != StakingProgramID {
		t.Errorf("Epoch State File TxManager should be StakingProgramID")
	}
	if epochStateFile.Executable {
		t.Errorf("Epoch State File should not be executable")
	}

	// Verify data can be deserialized
	epochNumber, epochStartSlot, totalSlotsInEpoch, validatorSchedule, missedBlockCounters, err := quanticscript.DeserializeEpochState(epochStateFile.Data)
	if err != nil {
		t.Fatalf("failed to deserialize Epoch State: %v", err)
	}

	if epochNumber != 0 {
		t.Errorf("expected epoch 0, got %d", epochNumber)
	}
	if epochStartSlot != 0 {
		t.Errorf("expected epoch start slot 0, got %d", epochStartSlot)
	}
	if totalSlotsInEpoch != 432000 {
		t.Errorf("expected total slots 432000, got %d", totalSlotsInEpoch)
	}
	if len(validatorSchedule) != 2 {
		t.Errorf("expected 2 validators, got %d", len(validatorSchedule))
	}
	if len(missedBlockCounters) != 2 {
		t.Errorf("expected 2 missed block counters, got %d", len(missedBlockCounters))
	}

	// Verify all missed block counters are initialized to 0
	for i, counter := range missedBlockCounters {
		if counter != 0 {
			t.Errorf("validator %d: expected missed block counter 0, got %d", i, counter)
		}
	}
}

// TestRewardPoolFileStructure verifies that the Reward Pool File has the correct structure
// and can be properly deserialized.
// Requirements: 9.3, 10.3
func TestRewardPoolFileStructure(t *testing.T) {
	fs, err := filestore.NewFileStore(t.TempDir())
	if err != nil {
		t.Fatalf("failed to create FileStore: %v", err)
	}
	defer fs.Close()

	genesisValidators := []GenesisValidator{
		{
			PublicKey:   [32]byte{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1},
			StakeAmount: 1000000,
		},
	}

	config := GenesisConfig{
		EpochLength:       432000,
		GenesisValidators: genesisValidators,
	}

	err = InitializeDPoSGenesis(fs, config)
	if err != nil {
		t.Fatalf("InitializeDPoSGenesis failed: %v", err)
	}

	rewardPoolFile, err := fs.GetFile(RewardPoolFileID)
	if err != nil {
		t.Fatalf("Reward Pool File not found: %v", err)
	}

	// Verify File properties
	if rewardPoolFile.ID != RewardPoolFileID {
		t.Errorf("Reward Pool File ID mismatch")
	}
	if rewardPoolFile.TxManager != StakingProgramID {
		t.Errorf("Reward Pool File TxManager should be StakingProgramID")
	}
	if rewardPoolFile.Executable {
		t.Errorf("Reward Pool File should not be executable")
	}

	// Verify data can be deserialized
	balance, lastDistributedEpoch, err := quanticscript.DeserializeRewardPool(rewardPoolFile.Data)
	if err != nil {
		t.Fatalf("failed to deserialize Reward Pool: %v", err)
	}

	if balance != 0 {
		t.Errorf("expected balance 0, got %d", balance)
	}
	if lastDistributedEpoch != 0 {
		t.Errorf("expected last distributed epoch 0, got %d", lastDistributedEpoch)
	}
}

// TestLoadGenesisConfig verifies that LoadGenesisConfig correctly parses a genesis configuration JSON file
// Requirements: 6.1, 6.2
func TestLoadGenesisConfig(t *testing.T) {
	// Create a temporary directory for the test file
	tmpDir := t.TempDir()
	configPath := tmpDir + "/genesis.json"

	// Create a genesis configuration JSON file
	configJSON := `{
  "epochLength": 432000,
  "validators": [
    {
      "publicKey": "0100000000000000000000000000000000000000000000000000000000000001",
      "stake": 10000000
    },
    {
      "publicKey": "0200000000000000000000000000000000000000000000000000000000000002",
      "stake": 5000000
    },
    {
      "publicKey": "0300000000000000000000000000000000000000000000000000000000000003",
      "stake": 2000000
    }
  ]
}`

	// Write the JSON file
	err := writeFile(configPath, []byte(configJSON))
	if err != nil {
		t.Fatalf("failed to write genesis config file: %v", err)
	}

	// Load the genesis configuration
	config, err := LoadGenesisConfig(configPath)
	if err != nil {
		t.Fatalf("LoadGenesisConfig failed: %v", err)
	}

	// Verify the configuration
	if config.EpochLength != 432000 {
		t.Errorf("expected epoch length 432000, got %d", config.EpochLength)
	}

	if len(config.GenesisValidators) != 3 {
		t.Errorf("expected 3 validators, got %d", len(config.GenesisValidators))
	}

	// Verify first validator
	if config.GenesisValidators[0].StakeAmount != 10000000 {
		t.Errorf("validator 0: expected stake 10000000, got %d", config.GenesisValidators[0].StakeAmount)
	}

	// Verify second validator
	if config.GenesisValidators[1].StakeAmount != 5000000 {
		t.Errorf("validator 1: expected stake 5000000, got %d", config.GenesisValidators[1].StakeAmount)
	}

	// Verify third validator
	if config.GenesisValidators[2].StakeAmount != 2000000 {
		t.Errorf("validator 2: expected stake 2000000, got %d", config.GenesisValidators[2].StakeAmount)
	}
}

// TestLoadGenesisConfigZeroValidators verifies that LoadGenesisConfig rejects a config with zero validators
// Requirement: 6.2
func TestLoadGenesisConfigZeroValidators(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := tmpDir + "/genesis.json"

	configJSON := `{
  "epochLength": 432000,
  "validators": []
}`

	err := writeFile(configPath, []byte(configJSON))
	if err != nil {
		t.Fatalf("failed to write genesis config file: %v", err)
	}

	_, err = LoadGenesisConfig(configPath)
	if err == nil {
		t.Fatalf("expected error for zero validators, got nil")
	}
}

// TestLoadGenesisConfigInvalidJSON verifies that LoadGenesisConfig rejects invalid JSON
// Requirement: 6.1
func TestLoadGenesisConfigInvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := tmpDir + "/genesis.json"

	invalidJSON := `{
  "epochLength": 432000,
  "validators": [
    {
      "publicKey": "0x01...",
      "stake": 10000000
    }
  ]
  // Missing closing brace
`

	err := writeFile(configPath, []byte(invalidJSON))
	if err != nil {
		t.Fatalf("failed to write genesis config file: %v", err)
	}

	_, err = LoadGenesisConfig(configPath)
	if err == nil {
		t.Fatalf("expected error for invalid JSON, got nil")
	}
}

// TestLoadGenesisConfigInvalidPublicKey verifies that LoadGenesisConfig rejects invalid public keys
// Requirement: 6.1
func TestLoadGenesisConfigInvalidPublicKey(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := tmpDir + "/genesis.json"

	configJSON := `{
  "epochLength": 432000,
  "validators": [
    {
      "publicKey": "invalid_hex_string",
      "stake": 10000000
    }
  ]
}`

	err := writeFile(configPath, []byte(configJSON))
	if err != nil {
		t.Fatalf("failed to write genesis config file: %v", err)
	}

	_, err = LoadGenesisConfig(configPath)
	if err == nil {
		t.Fatalf("expected error for invalid public key, got nil")
	}
}

// TestLoadGenesisConfigInvalidStake verifies that LoadGenesisConfig rejects invalid stakes
// Requirement: 6.1
func TestLoadGenesisConfigInvalidStake(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := tmpDir + "/genesis.json"

	configJSON := `{
  "epochLength": 432000,
  "validators": [
    {
      "publicKey": "0100000000000000000000000000000000000000000000000000000000000001",
      "stake": 0
    }
  ]
}`

	err := writeFile(configPath, []byte(configJSON))
	if err != nil {
		t.Fatalf("failed to write genesis config file: %v", err)
	}

	_, err = LoadGenesisConfig(configPath)
	if err == nil {
		t.Fatalf("expected error for zero stake, got nil")
	}
}

// TestLoadGenesisConfigFileNotFound verifies that LoadGenesisConfig handles missing files
// Requirement: 6.1
func TestLoadGenesisConfigFileNotFound(t *testing.T) {
	_, err := LoadGenesisConfig("/nonexistent/path/genesis.json")
	if err == nil {
		t.Fatalf("expected error for missing file, got nil")
	}
}

// TestLoadGenesisConfigInvalidEpochLength verifies that LoadGenesisConfig rejects invalid epoch lengths
// Requirement: 6.1
func TestLoadGenesisConfigInvalidEpochLength(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := tmpDir + "/genesis.json"

	configJSON := `{
  "epochLength": 0,
  "validators": [
    {
      "publicKey": "0100000000000000000000000000000000000000000000000000000000000001",
      "stake": 10000000
    }
  ]
}`

	err := writeFile(configPath, []byte(configJSON))
	if err != nil {
		t.Fatalf("failed to write genesis config file: %v", err)
	}

	_, err = LoadGenesisConfig(configPath)
	if err == nil {
		t.Fatalf("expected error for zero epoch length, got nil")
	}
}

// Helper function to write a file
func writeFile(path string, data []byte) error {
	return os.WriteFile(path, data, 0644)
}
