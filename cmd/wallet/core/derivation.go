package core

import (
	"crypto/ed25519"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/binary"
	"fmt"
)

// HDKey represents a hierarchical deterministic key with chain code
type HDKey struct {
	Key       []byte
	ChainCode []byte
}

// DeriveKey derives a key using SLIP-0010 Ed25519 derivation
// Path format: m/44'/501'/0'/0'/index'
// All indices are hardened (indicated by ')
func DeriveKey(seed []byte, path string) (*HDKey, error) {
	// For Ed25519, we use SLIP-0010 which starts with "ed25519 seed"
	h := hmac.New(sha512.New, []byte("ed25519 seed"))
	h.Write(seed)
	I := h.Sum(nil)

	// Split into key and chain code
	masterKey := I[:32]
	masterChainCode := I[32:]

	// Derivation path: m/44'/501'/0'/0'/index'
	// All indices are hardened for Ed25519
	indices := []uint32{
		hardenedIndex(44),  // purpose
		hardenedIndex(501), // coin type (Solana)
		hardenedIndex(0),   // account
		hardenedIndex(0),   // change
	}

	key := masterKey
	chainCode := masterChainCode

	// Derive through the fixed path
	for _, index := range indices {
		key, chainCode = deriveHardenedChild(key, chainCode, index)
	}

	return &HDKey{
		Key:       key,
		ChainCode: chainCode,
	}, nil
}

// DeriveAccount derives a specific account from the master key
// accountIndex is the final index in the path m/44'/501'/0'/0'/index'
func (k *HDKey) DeriveAccount(accountIndex uint32) (*HDKey, error) {
	// Derive the account with hardened index
	key, chainCode := deriveHardenedChild(k.Key, k.ChainCode, hardenedIndex(accountIndex))

	return &HDKey{
		Key:       key,
		ChainCode: chainCode,
	}, nil
}

// ToEd25519 converts the HD key to an Ed25519 keypair
func (k *HDKey) ToEd25519() (ed25519.PublicKey, ed25519.PrivateKey) {
	// For Ed25519, the private key is the 32-byte seed
	// The actual private key in Ed25519 is 64 bytes (32 seed + 32 public)
	privateKey := ed25519.NewKeyFromSeed(k.Key)
	publicKey := privateKey.Public().(ed25519.PublicKey)

	return publicKey, privateKey
}

// deriveHardenedChild derives a hardened child key
func deriveHardenedChild(key, chainCode []byte, index uint32) ([]byte, []byte) {
	// For hardened derivation: HMAC-SHA512(chainCode, 0x00 || key || index)
	h := hmac.New(sha512.New, chainCode)
	h.Write([]byte{0x00})
	h.Write(key)

	indexBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(indexBytes, index)
	h.Write(indexBytes)

	I := h.Sum(nil)

	// Split into new key and chain code
	newKey := I[:32]
	newChainCode := I[32:]

	return newKey, newChainCode
}

// hardenedIndex returns a hardened index (adds 2^31)
func hardenedIndex(index uint32) uint32 {
	return index + 0x80000000
}

// DerivationPath represents a BIP44 derivation path
type DerivationPath struct {
	Purpose  uint32
	CoinType uint32
	Account  uint32
	Change   uint32
	Index    uint32
}

// DefaultDerivationPath returns the default path for PoH blockchain
// m/44'/501'/0'/0'/0'
func DefaultDerivationPath() DerivationPath {
	return DerivationPath{
		Purpose:  44,
		CoinType: 501, // Using Solana's coin type for compatibility
		Account:  0,
		Change:   0,
		Index:    0,
	}
}

// String returns the string representation of the derivation path
func (p DerivationPath) String() string {
	return fmt.Sprintf("m/%d'/%d'/%d'/%d'/%d'",
		p.Purpose, p.CoinType, p.Account, p.Change, p.Index)
}
