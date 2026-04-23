package poh

import (
	"crypto/sha256"
	"sync"
	"time"
)

// Tick represents a single tick in the PoH sequence
type Tick struct {
	HashValue  []byte
	Timestamp  time.Time
	TickNumber int64
}

// PohClock generates a verifiable sequence of SHA-256 hashes
type PohClock struct {
	currentHash []byte
	tickCount   int64
	hashCount   int64
	mu          sync.RWMutex
}

// NewPohClock initializes a new PoH clock with a seed hash
func NewPohClock(seed []byte) *PohClock {
	// If no seed provided, use a default seed
	if len(seed) == 0 {
		seed = []byte("genesis")
	}

	// Hash the seed to get the initial hash
	hash := sha256.Sum256(seed)

	return &PohClock{
		currentHash: hash[:],
		tickCount:   0,
		hashCount:   0,
	}
}

// HashOnce performs a single SHA-256 hash operation on currentHash
func (p *PohClock) HashOnce() []byte {
	p.mu.Lock()
	defer p.mu.Unlock()

	hash := sha256.Sum256(p.currentHash)
	p.currentHash = hash[:]
	p.hashCount++

	return p.currentHash
}

// GetCurrentHash returns the current hash state with thread-safe access
func (p *PohClock) GetCurrentHash() []byte {
	p.mu.RLock()
	defer p.mu.RUnlock()

	// Return a copy to prevent external modification
	hashCopy := make([]byte, len(p.currentHash))
	copy(hashCopy, p.currentHash)
	return hashCopy
}

// GetTickCount returns the total number of ticks generated
func (p *PohClock) GetTickCount() int64 {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return p.tickCount
}

// Tick performs 12,500 hash operations and returns a Tick
func (p *PohClock) Tick() Tick {
	const hashesPerTick = 12500

	// Perform 12,500 hash operations
	for i := 0; i < hashesPerTick; i++ {
		p.mu.Lock()
		hash := sha256.Sum256(p.currentHash)
		p.currentHash = hash[:]
		p.hashCount++
		p.mu.Unlock()
	}

	// Atomically update tick count and capture state
	p.mu.Lock()
	p.tickCount++
	tickNumber := p.tickCount
	hashCopy := make([]byte, len(p.currentHash))
	copy(hashCopy, p.currentHash)
	p.mu.Unlock()

	return Tick{
		HashValue:  hashCopy,
		Timestamp:  time.Now(),
		TickNumber: tickNumber,
	}
}
