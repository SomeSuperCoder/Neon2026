package consensus

import (
	"fmt"
	"time"

	"github.com/poh-blockchain/internal/blockchain"
	"github.com/poh-blockchain/internal/network"
)

// ConsensusManager handles consensus protocol and slot timing
type ConsensusManager struct {
	nodeType         network.NodeType
	slotDurationMs   int64
	genesisTimestamp time.Time
}

// NewConsensusManager creates a new consensus manager with 400ms slot duration
func NewConsensusManager(nodeType network.NodeType) *ConsensusManager {
	return &ConsensusManager{
		nodeType:         nodeType,
		slotDurationMs:   400,
		genesisTimestamp: time.Now(),
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
// Simplified implementation: always true for LEADER type nodes
func (cm *ConsensusManager) IsLeader(slot int64) bool {
	return cm.nodeType == network.LEADER
}

// ValidateBlock performs basic block validation
func (cm *ConsensusManager) ValidateBlock(block blockchain.Block) error {
	// Verify block slot matches expected slot
	currentSlot := cm.GetCurrentSlot()
	if block.Header.Slot > currentSlot {
		return fmt.Errorf("block slot %d is in the future (current slot: %d)", block.Header.Slot, currentSlot)
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
