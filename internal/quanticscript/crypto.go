package quanticscript

import (
	"crypto/ed25519"
	"crypto/sha256"
	"fmt"
)

// sha256Hash computes the SHA-256 hash of the input data
func sha256Hash(data []byte) []byte {
	hash := sha256.Sum256(data)
	return hash[:]
}

// verifySignature verifies an Ed25519 signature
// pubkey: 32-byte Ed25519 public key
// message: arbitrary message bytes
// signature: 64-byte Ed25519 signature
func verifySignature(pubkey, message, signature []byte) bool {
	// Validate input lengths
	if len(pubkey) != ed25519.PublicKeySize {
		return false
	}
	if len(signature) != ed25519.SignatureSize {
		return false
	}

	// Verify signature
	return ed25519.Verify(pubkey, message, signature)
}

// derivePublicKey derives an Ed25519 public key from a seed
// seed: 32-byte seed for key derivation
func derivePublicKey(seed []byte) ([]byte, error) {
	// Validate seed length
	if len(seed) != ed25519.SeedSize {
		return nil, fmt.Errorf("invalid seed length: expected %d bytes, got %d", ed25519.SeedSize, len(seed))
	}

	// Generate private key from seed
	privateKey := ed25519.NewKeyFromSeed(seed)

	// Extract public key
	publicKey := privateKey.Public().(ed25519.PublicKey)

	return publicKey, nil
}
