package internal

import (
	"os"
	"testing"

	"github.com/poh-blockchain/internal/consensus"
	"github.com/poh-blockchain/internal/filestore"
	"github.com/poh-blockchain/internal/genesis"
	"github.com/poh-blockchain/internal/runtime"
	"github.com/poh-blockchain/programs"
)

// TestDPoSWiringIntegration tests that all DPoS components are properly wired together
func TestDPoSWiringIntegration(t *testing.T) {
	// Create temporary database for file store
	dbPath := "test_dpos_wiring.db"
	defer os.RemoveAll(dbPath)

	// Initialize FileStore
	fs, err := filestore.NewFileStore(dbPath)
	if err != nil {
		t.Fatalf("Failed to create FileStore: %v", err)
	}
	defer fs.Close()

	// Load built-in programs including staking
	if err := genesis.LoadBuiltinPrograms(fs, programs.SystemProgram, programs.TokenProgram, programs.StakingProgram); err != nil {
		t.Fatalf("Failed to load built-in programs: %v", err)
	}

	// Create genesis configuration
	genesisConfig := consensus.GenesisConfig{
		EpochLength: 100, // Short epoch for testing
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

	// Create ConsensusManager with genesis config
	// Use the first genesis validator's ID for this test
	pk := genesisConfig.GenesisValidators[0].PublicKey
	validatorID := filestore.GenerateFileID(append([]byte("validator:"), pk[:]...))
	cm := consensus.NewConsensusManagerWithGenesis(validatorID, nil, genesisConfig)

	// Initialize Runtime
	rt := runtime.NewRuntime()

	// Wire FileStore and Runtime to ConsensusManager
	cm.SetFileStore(fs)
	cm.SetRuntime(rt)

	// Initialize DPoS genesis
	if err := cm.InitializeGenesis(genesisConfig); err != nil {
		t.Fatalf("Failed to initialize DPoS genesis: %v", err)
	}

	// Verify epoch state file was created
	epochStateFile, err := fs.GetFile(genesis.EpochStateFileID)
	if err != nil {
		t.Fatalf("Epoch State File not found: %v", err)
	}
	if epochStateFile == nil {
		t.Fatal("Epoch State File is nil")
	}

	// Verify reward pool file was created
	rewardPoolFile, err := fs.GetFile(genesis.RewardPoolFileID)
	if err != nil {
		t.Fatalf("Reward Pool File not found: %v", err)
	}
	if rewardPoolFile == nil {
		t.Fatal("Reward Pool File is nil")
	}

	// Verify staking program was loaded
	stakingProgramFile, err := fs.GetFile(genesis.StakingProgramID)
	if err != nil {
		t.Fatalf("Staking Program File not found: %v", err)
	}
	if stakingProgramFile == nil {
		t.Fatal("Staking Program File is nil")
	}
	if !stakingProgramFile.Executable {
		t.Error("Staking Program File should be executable")
	}

	// Verify current epoch is 0
	if cm.GetCurrentEpoch() != 0 {
		t.Errorf("Expected current epoch to be 0, got %d", cm.GetCurrentEpoch())
	}

	// Verify epoch boundary detection works
	if !cm.IsEpochBoundary(0) {
		t.Error("Slot 0 should be an epoch boundary")
	}
	if !cm.IsEpochBoundary(100) {
		t.Error("Slot 100 should be an epoch boundary")
	}
	if cm.IsEpochBoundary(50) {
		t.Error("Slot 50 should not be an epoch boundary")
	}

	// Verify scheduled validator can be retrieved
	scheduledValidator := cm.GetScheduledValidator(0)
	if scheduledValidator == (filestore.FileID{}) {
		t.Error("Expected a scheduled validator for slot 0")
	}

	t.Log("DPoS wiring integration test passed")
}
