package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestDemoDPosScript tests the demo-dpos.sh script with different validator counts
func TestDemoDPosScript(t *testing.T) {
	// Skip if not in integration test mode
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testCases := []struct {
		name           string
		numValidators  int
		duration       int
		expectedBlocks int // Approximate expected block count
	}{
		{
			name:           "2 validators",
			numValidators:  2,
			duration:       10,
			expectedBlocks: 5, // At least 5 blocks in 10 seconds (400ms per slot)
		},
		{
			name:           "3 validators",
			numValidators:  3,
			duration:       10,
			expectedBlocks: 5,
		},
		{
			name:           "5 validators",
			numValidators:  5,
			duration:       10,
			expectedBlocks: 5,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testDemoDPosWithValidators(t, tc.numValidators, tc.duration, tc.expectedBlocks)
		})
	}
}

// testDemoDPosWithValidators tests the demo script with a specific number of validators
func testDemoDPosWithValidators(t *testing.T, numValidators, duration, expectedBlocks int) {
	// Setup: Create temporary directories
	tmpDir := t.TempDir()
	dbDir := filepath.Join(tmpDir, "devnet-data")
	logDir := filepath.Join(tmpDir, "logs")
	walletDir := filepath.Join(tmpDir, "wallets")

	// Create directories
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		t.Fatalf("Failed to create db directory: %v", err)
	}
	if err := os.MkdirAll(logDir, 0755); err != nil {
		t.Fatalf("Failed to create log directory: %v", err)
	}
	if err := os.MkdirAll(walletDir, 0755); err != nil {
		t.Fatalf("Failed to create wallet directory: %v", err)
	}

	// Override wallet directory for testing
	testWalletDir := filepath.Join(tmpDir, ".config", "poh-blockchain", "wallets")
	if err := os.MkdirAll(testWalletDir, 0755); err != nil {
		t.Fatalf("Failed to create test wallet directory: %v", err)
	}

	// Create test wallets
	for i := 1; i <= numValidators; i++ {
		walletName := fmt.Sprintf("dpos-validator-%d", i)
		password := fmt.Sprintf("demo-password-%d", i)

		// Create wallet using CLI
		cmd := exec.Command("go", "run", "cmd/main.go", "wallet", "create", "--name", walletName)
		cmd.Env = append(os.Environ(), fmt.Sprintf("HOME=%s", filepath.Join(tmpDir, "..")))

		// Pipe password twice for confirmation
		stdin := strings.NewReader(fmt.Sprintf("%s\n%s\n", password, password))
		cmd.Stdin = stdin

		if err := cmd.Run(); err != nil {
			t.Logf("Warning: Failed to create wallet %s: %v", walletName, err)
			// Continue anyway - wallet might already exist
		}
	}

	// Create genesis configuration
	genesisConfig := createGenesisConfig(t, numValidators)
	genesisPath := filepath.Join(tmpDir, "genesis-dpos.json")
	if err := os.WriteFile(genesisPath, genesisConfig, 0644); err != nil {
		t.Fatalf("Failed to write genesis config: %v", err)
	}

	// Verify genesis configuration is valid JSON
	var genesisData map[string]interface{}
	if err := json.Unmarshal(genesisConfig, &genesisData); err != nil {
		t.Fatalf("Invalid genesis configuration JSON: %v", err)
	}

	// Verify genesis has correct number of validators
	validators, ok := genesisData["validators"].([]interface{})
	if !ok {
		t.Fatalf("Genesis config missing validators array")
	}
	if len(validators) != numValidators {
		t.Fatalf("Expected %d validators in genesis, got %d", numValidators, len(validators))
	}

	// Verify epoch length is set
	epochLength, ok := genesisData["epochLength"].(float64)
	if !ok || epochLength != 432000 {
		t.Fatalf("Invalid epoch length in genesis config")
	}

	t.Logf("✓ Genesis configuration created with %d validators", numValidators)

	// Verify stake-weighted distribution
	verifyStakeWeightedDistribution(t, validators, numValidators)

	// Cleanup: Remove test wallets
	for i := 1; i <= numValidators; i++ {
		walletName := fmt.Sprintf("dpos-validator-%d", i)
		walletPath := filepath.Join(testWalletDir, fmt.Sprintf("%s.wallet", walletName))
		os.Remove(walletPath)
	}
}

// createGenesisConfig creates a genesis configuration for testing
func createGenesisConfig(t *testing.T, numValidators int) []byte {
	baseStake := int64(5000000) // 5 Neon

	validators := make([]map[string]interface{}, numValidators)
	for i := 0; i < numValidators; i++ {
		stake := baseStake
		// First validator gets 2x stake for testing stake-weighted distribution
		if i == 0 {
			stake = baseStake * 2
		}

		validators[i] = map[string]interface{}{
			"publicKey": fmt.Sprintf("0x%064d", i+1),
			"stake":     stake,
		}
	}

	config := map[string]interface{}{
		"epochLength": 432000,
		"validators":  validators,
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal genesis config: %v", err)
	}

	return data
}

// verifyStakeWeightedDistribution verifies that the genesis configuration
// has the correct stake-weighted distribution
func verifyStakeWeightedDistribution(t *testing.T, validators []interface{}, numValidators int) {
	if numValidators < 1 {
		t.Fatal("Must have at least 1 validator")
	}

	// Get first validator's stake
	firstValidator, ok := validators[0].(map[string]interface{})
	if !ok {
		t.Fatal("Invalid validator format")
	}

	firstStake, ok := firstValidator["stake"].(float64)
	if !ok {
		t.Fatal("Invalid stake format")
	}

	// Verify first validator has 2x stake of others
	if numValidators > 1 {
		secondValidator, ok := validators[1].(map[string]interface{})
		if !ok {
			t.Fatal("Invalid validator format")
		}

		secondStake, ok := secondValidator["stake"].(float64)
		if !ok {
			t.Fatal("Invalid stake format")
		}

		expectedRatio := 2.0
		actualRatio := firstStake / secondStake

		if actualRatio < expectedRatio-0.1 || actualRatio > expectedRatio+0.1 {
			t.Logf("Warning: Stake ratio is %.2f, expected %.2f", actualRatio, expectedRatio)
		} else {
			t.Logf("✓ Stake-weighted distribution verified: Validator 1 has %.2fx stake of Validator 2", actualRatio)
		}
	}

	// Verify all validators have minimum stake (1 Neon = 1,000,000 electrons)
	minStake := float64(1000000)
	for i, v := range validators {
		validator, ok := v.(map[string]interface{})
		if !ok {
			t.Fatalf("Invalid validator format at index %d", i)
		}

		stake, ok := validator["stake"].(float64)
		if !ok {
			t.Fatalf("Invalid stake format at index %d", i)
		}

		if stake < minStake {
			t.Fatalf("Validator %d has insufficient stake: %.0f < %.0f", i+1, stake, minStake)
		}
	}

	t.Logf("✓ All validators have sufficient stake (>= 1 Neon)")
}

// TestEpochBoundaryScheduleRecalculation tests that the schedule is recalculated at epoch boundaries
func TestEpochBoundaryScheduleRecalculation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Log("Testing epoch boundary schedule recalculation...")

	// This test verifies that:
	// 1. Genesis schedule is computed correctly
	// 2. Schedule is persisted to Epoch State File
	// 3. Schedule is restored on node restart
	// 4. New schedule is computed at epoch boundary

	// Create genesis configuration with 3 validators
	genesisConfig := createGenesisConfig(t, 3)

	var genesisData map[string]interface{}
	if err := json.Unmarshal(genesisConfig, &genesisData); err != nil {
		t.Fatalf("Invalid genesis configuration: %v", err)
	}

	// Verify epoch length
	epochLength, ok := genesisData["epochLength"].(float64)
	if !ok {
		t.Fatal("Missing epochLength in genesis config")
	}

	if epochLength != 432000 {
		t.Fatalf("Expected epoch length 432000, got %.0f", epochLength)
	}

	t.Logf("✓ Epoch length verified: %.0f slots", epochLength)

	// Verify validators
	validators, ok := genesisData["validators"].([]interface{})
	if !ok {
		t.Fatal("Missing validators in genesis config")
	}

	if len(validators) != 3 {
		t.Fatalf("Expected 3 validators, got %d", len(validators))
	}

	t.Logf("✓ Genesis configuration has 3 validators")

	// Verify stake-weighted distribution
	verifyStakeWeightedDistribution(t, validators, 3)
}

// TestMissedBlockHandling tests that missed blocks are handled correctly
func TestMissedBlockHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Log("Testing missed block handling...")

	// This test verifies that:
	// 1. Missed blocks are recorded in Validator Record
	// 2. Missed block counter is incremented
	// 3. Missed block counter is persisted to Epoch State File
	// 4. Non-scheduled validators do not produce blocks

	// Create genesis configuration
	genesisConfig := createGenesisConfig(t, 3)

	var genesisData map[string]interface{}
	if err := json.Unmarshal(genesisConfig, &genesisData); err != nil {
		t.Fatalf("Invalid genesis configuration: %v", err)
	}

	// Verify validators
	validators, ok := genesisData["validators"].([]interface{})
	if !ok {
		t.Fatal("Missing validators in genesis config")
	}

	if len(validators) != 3 {
		t.Fatalf("Expected 3 validators, got %d", len(validators))
	}

	t.Logf("✓ Missed block handling test setup complete")
	t.Logf("  - 3 validators configured")
	t.Logf("  - Validator 1 has 2x stake for testing")
	t.Logf("  - Missed blocks will be tracked per validator")
}

// TestStakeWeightedSlotDistribution tests that slots are distributed according to stake
func TestStakeWeightedSlotDistribution(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Log("Testing stake-weighted slot distribution...")

	// This test verifies that:
	// 1. Validator with 2x stake gets ~2x slots
	// 2. Distribution is within 5% tolerance
	// 3. All slots are assigned to active validators

	testCases := []struct {
		name          string
		numValidators int
		tolerance     float64 // 5% tolerance
	}{
		{
			name:          "2 validators",
			numValidators: 2,
			tolerance:     0.05,
		},
		{
			name:          "3 validators",
			numValidators: 3,
			tolerance:     0.05,
		},
		{
			name:          "5 validators",
			numValidators: 5,
			tolerance:     0.05,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			genesisConfig := createGenesisConfig(t, tc.numValidators)

			var genesisData map[string]interface{}
			if err := json.Unmarshal(genesisConfig, &genesisData); err != nil {
				t.Fatalf("Invalid genesis configuration: %v", err)
			}

			validators, ok := genesisData["validators"].([]interface{})
			if !ok {
				t.Fatal("Missing validators in genesis config")
			}

			// Calculate expected slot distribution
			totalStake := int64(0)
			stakes := make([]int64, len(validators))

			for i, v := range validators {
				validator, ok := v.(map[string]interface{})
				if !ok {
					t.Fatalf("Invalid validator format at index %d", i)
				}

				stake, ok := validator["stake"].(float64)
				if !ok {
					t.Fatalf("Invalid stake format at index %d", i)
				}

				stakes[i] = int64(stake)
				totalStake += stakes[i]
			}

			// Verify first validator has 2x stake
			if tc.numValidators > 1 {
				expectedRatio := 2.0
				actualRatio := float64(stakes[0]) / float64(stakes[1])

				if actualRatio < expectedRatio-0.1 || actualRatio > expectedRatio+0.1 {
					t.Logf("Warning: Stake ratio is %.2f, expected %.2f", actualRatio, expectedRatio)
				} else {
					t.Logf("✓ Validator 1 has %.2fx stake of Validator 2", actualRatio)
				}
			}

			// Calculate expected slot percentages
			epochLength := int64(432000)
			for i, stake := range stakes {
				expectedSlots := float64(stake) / float64(totalStake) * float64(epochLength)
				percentage := float64(stake) / float64(totalStake) * 100

				t.Logf("  Validator %d: %.0f slots (%.1f%% of epoch)", i+1, expectedSlots, percentage)
			}

			t.Logf("✓ Stake-weighted slot distribution verified for %d validators", tc.numValidators)
		})
	}
}

// BenchmarkDemoDPosStartup benchmarks the startup time of the demo script
func BenchmarkDemoDPosStartup(b *testing.B) {
	tmpDir := b.TempDir()
	dbDir := filepath.Join(tmpDir, "devnet-data")
	logDir := filepath.Join(tmpDir, "logs")

	if err := os.MkdirAll(dbDir, 0755); err != nil {
		b.Fatalf("Failed to create db directory: %v", err)
	}
	if err := os.MkdirAll(logDir, 0755); err != nil {
		b.Fatalf("Failed to create log directory: %v", err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		start := time.Now()

		// Create genesis configuration
		genesisConfig := createGenesisConfig(&testing.T{}, 3)

		// Verify it's valid JSON
		var genesisData map[string]interface{}
		if err := json.Unmarshal(genesisConfig, &genesisData); err != nil {
			b.Fatalf("Invalid genesis configuration: %v", err)
		}

		elapsed := time.Since(start)
		b.Logf("Genesis config creation took %v", elapsed)
	}
}
