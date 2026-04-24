package transaction

import (
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/poh-blockchain/internal/filestore"
)

// Package transaction implements the transaction and instruction data structures
// for the file-based state model, including serialization and access control types.

// TxID is a unique identifier for transactions (32-byte hash)
type TxID [32]byte

// String returns the hex-encoded string representation of TxID
func (tid TxID) String() string {
	return hex.EncodeToString(tid[:])
}

// TxIDFromString creates a TxID from a hex-encoded string
func TxIDFromString(s string) (TxID, error) {
	var tid TxID
	bytes, err := hex.DecodeString(s)
	if err != nil {
		return tid, fmt.Errorf("invalid hex string: %w", err)
	}
	if len(bytes) != 32 {
		return tid, fmt.Errorf("invalid TxID length: expected 32 bytes, got %d", len(bytes))
	}
	copy(tid[:], bytes)
	return tid, nil
}

// TxIDFromBytes creates a TxID from a byte slice
func TxIDFromBytes(b []byte) (TxID, error) {
	var tid TxID
	if len(b) != 32 {
		return tid, fmt.Errorf("invalid TxID length: expected 32 bytes, got %d", len(b))
	}
	copy(tid[:], b)
	return tid, nil
}

// AccessPermission defines the type of access (Read or Write) for file operations
// This implements Requirement 4.2
type AccessPermission uint8

const (
	// Read permission allows reading file data
	Read AccessPermission = 1
	// Write permission allows modifying file data
	Write AccessPermission = 2
)

// String returns the string representation of AccessPermission
func (ap AccessPermission) String() string {
	switch ap {
	case Read:
		return "Read"
	case Write:
		return "Write"
	default:
		return "Unknown"
	}
}

// FileAccess represents a file reference with its access permission
// This implements Requirement 4.1
type FileAccess struct {
	FileID     filestore.FileID
	Permission AccessPermission
}

// fileAccessJSON is used for JSON serialization/deserialization
type fileAccessJSON struct {
	FileID     string `json:"file_id"`
	Permission uint8  `json:"permission"`
}

// MarshalJSON implements custom JSON marshaling for FileAccess
func (fa FileAccess) MarshalJSON() ([]byte, error) {
	faj := fileAccessJSON{
		FileID:     fa.FileID.String(),
		Permission: uint8(fa.Permission),
	}
	return json.Marshal(faj)
}

// UnmarshalJSON implements custom JSON unmarshaling for FileAccess
func (fa *FileAccess) UnmarshalJSON(data []byte) error {
	var faj fileAccessJSON
	if err := json.Unmarshal(data, &faj); err != nil {
		return fmt.Errorf("failed to unmarshal file access: %w", err)
	}

	fileID, err := filestore.FileIDFromString(faj.FileID)
	if err != nil {
		return fmt.Errorf("invalid file ID: %w", err)
	}

	fa.FileID = fileID
	fa.Permission = AccessPermission(faj.Permission)
	return nil
}

// PublicKey represents an Ed25519 public key (32 bytes)
type PublicKey [32]byte

// String returns the hex-encoded string representation of PublicKey
func (pk PublicKey) String() string {
	return hex.EncodeToString(pk[:])
}

// PublicKeyFromString creates a PublicKey from a hex-encoded string
func PublicKeyFromString(s string) (PublicKey, error) {
	var pk PublicKey
	bytes, err := hex.DecodeString(s)
	if err != nil {
		return pk, fmt.Errorf("invalid hex string: %w", err)
	}
	if len(bytes) != 32 {
		return pk, fmt.Errorf("invalid PublicKey length: expected 32 bytes, got %d", len(bytes))
	}
	copy(pk[:], bytes)
	return pk, nil
}

// PublicKeyFromBytes creates a PublicKey from a byte slice
func PublicKeyFromBytes(b []byte) (PublicKey, error) {
	var pk PublicKey
	if len(b) != 32 {
		return pk, fmt.Errorf("invalid PublicKey length: expected 32 bytes, got %d", len(b))
	}
	copy(pk[:], b)
	return pk, nil
}

// ToEd25519 converts PublicKey to ed25519.PublicKey for signature verification
func (pk PublicKey) ToEd25519() ed25519.PublicKey {
	return ed25519.PublicKey(pk[:])
}

// Signature represents a cryptographic signature with its associated public key
// This implements Requirement 3.3
type Signature struct {
	PublicKey PublicKey
	Signature [64]byte
}

// signatureJSON is used for JSON serialization/deserialization
type signatureJSON struct {
	PublicKey string `json:"public_key"`
	Signature string `json:"signature"`
}

// MarshalJSON implements custom JSON marshaling for Signature
func (s Signature) MarshalJSON() ([]byte, error) {
	sj := signatureJSON{
		PublicKey: s.PublicKey.String(),
		Signature: hex.EncodeToString(s.Signature[:]),
	}
	return json.Marshal(sj)
}

// UnmarshalJSON implements custom JSON unmarshaling for Signature
func (s *Signature) UnmarshalJSON(data []byte) error {
	var sj signatureJSON
	if err := json.Unmarshal(data, &sj); err != nil {
		return fmt.Errorf("failed to unmarshal signature: %w", err)
	}

	pk, err := PublicKeyFromString(sj.PublicKey)
	if err != nil {
		return fmt.Errorf("invalid public key: %w", err)
	}
	s.PublicKey = pk

	sigBytes, err := hex.DecodeString(sj.Signature)
	if err != nil {
		return fmt.Errorf("invalid signature hex: %w", err)
	}
	if len(sigBytes) != 64 {
		return fmt.Errorf("invalid signature length: expected 64 bytes, got %d", len(sigBytes))
	}
	copy(s.Signature[:], sigBytes)

	return nil
}

// Verify verifies the signature against the provided message
func (s *Signature) Verify(message []byte) bool {
	return ed25519.Verify(s.PublicKey.ToEd25519(), message, s.Signature[:])
}

// Instruction represents a single operation within a transaction
// This implements Requirements 4.1, 4.2, 4.3, 4.4, 4.5
type Instruction struct {
	ProgramID filestore.FileID
	Inputs    map[string]FileAccess
	Data      []byte
}

// instructionJSON is used for JSON serialization/deserialization
type instructionJSON struct {
	ProgramID string                `json:"program_id"`
	Inputs    map[string]FileAccess `json:"inputs"`
	Data      []byte                `json:"data"`
}

// MarshalJSON implements custom JSON marshaling for Instruction
func (i Instruction) MarshalJSON() ([]byte, error) {
	ij := instructionJSON{
		ProgramID: i.ProgramID.String(),
		Inputs:    i.Inputs,
		Data:      i.Data,
	}
	return json.Marshal(ij)
}

// UnmarshalJSON implements custom JSON unmarshaling for Instruction
func (i *Instruction) UnmarshalJSON(data []byte) error {
	var ij instructionJSON
	if err := json.Unmarshal(data, &ij); err != nil {
		return fmt.Errorf("failed to unmarshal instruction: %w", err)
	}

	programID, err := filestore.FileIDFromString(ij.ProgramID)
	if err != nil {
		return fmt.Errorf("invalid program ID: %w", err)
	}

	i.ProgramID = programID
	i.Inputs = ij.Inputs
	i.Data = ij.Data
	return nil
}

// Marshal serializes an Instruction to JSON bytes
func (i *Instruction) Marshal() ([]byte, error) {
	return json.Marshal(i)
}

// UnmarshalInstruction deserializes an Instruction from JSON bytes
func UnmarshalInstruction(data []byte) (*Instruction, error) {
	i := &Instruction{}
	if err := json.Unmarshal(data, i); err != nil {
		return nil, fmt.Errorf("failed to unmarshal instruction: %w", err)
	}
	return i, nil
}

// Transaction represents a collection of instructions with signatures
// This implements Requirements 3.1, 3.2, 3.3, 3.4, 3.5
type Transaction struct {
	LastSeen     TxID
	Instructions []Instruction
	Signatures   []Signature
}

// transactionJSON is used for JSON serialization/deserialization
type transactionJSON struct {
	LastSeen     string        `json:"last_seen"`
	Instructions []Instruction `json:"instructions"`
	Signatures   []Signature   `json:"signatures"`
}

// MarshalJSON implements custom JSON marshaling for Transaction
func (t Transaction) MarshalJSON() ([]byte, error) {
	tj := transactionJSON{
		LastSeen:     t.LastSeen.String(),
		Instructions: t.Instructions,
		Signatures:   t.Signatures,
	}
	return json.Marshal(tj)
}

// UnmarshalJSON implements custom JSON unmarshaling for Transaction
func (t *Transaction) UnmarshalJSON(data []byte) error {
	var tj transactionJSON
	if err := json.Unmarshal(data, &tj); err != nil {
		return fmt.Errorf("failed to unmarshal transaction: %w", err)
	}

	lastSeen, err := TxIDFromString(tj.LastSeen)
	if err != nil {
		return fmt.Errorf("invalid last seen tx ID: %w", err)
	}

	t.LastSeen = lastSeen
	t.Instructions = tj.Instructions
	t.Signatures = tj.Signatures
	return nil
}

// Marshal serializes a Transaction to JSON bytes
func (t *Transaction) Marshal() ([]byte, error) {
	return json.Marshal(t)
}

// UnmarshalTransaction deserializes a Transaction from JSON bytes
func UnmarshalTransaction(data []byte) (*Transaction, error) {
	t := &Transaction{}
	if err := json.Unmarshal(data, t); err != nil {
		return nil, fmt.Errorf("failed to unmarshal transaction: %w", err)
	}
	return t, nil
}

// GetFeePayer returns the public key of the fee payer (first signature)
// This implements Requirement 3.4
func (t *Transaction) GetFeePayer() (PublicKey, error) {
	if len(t.Signatures) == 0 {
		return PublicKey{}, fmt.Errorf("transaction has no signatures")
	}
	return t.Signatures[0].PublicKey, nil
}

// GetSigners returns all public keys that signed the transaction
func (t *Transaction) GetSigners() []PublicKey {
	signers := make([]PublicKey, len(t.Signatures))
	for i, sig := range t.Signatures {
		signers[i] = sig.PublicKey
	}
	return signers
}

// FeeConfig holds configurable fee parameters for transaction processing
// This implements Requirement 7.2 and 7.3
type FeeConfig struct {
	BaseFee        int64 // Base fee per transaction
	InstructionFee int64 // Fee per instruction
	SignatureFee   int64 // Fee per signature
}

// DefaultFeeConfig returns the default fee configuration
func DefaultFeeConfig() FeeConfig {
	return FeeConfig{
		BaseFee:        5000, // Base fee per transaction
		InstructionFee: 1000, // Fee per instruction
		SignatureFee:   500,  // Fee per signature
	}
}

// CalculateFee computes the total fee for a transaction based on its complexity
// This implements Requirements 7.1, 7.2, 7.3
func CalculateFee(tx *Transaction, config FeeConfig) int64 {
	baseFee := config.BaseFee
	instructionFee := config.InstructionFee * int64(len(tx.Instructions))
	signatureFee := config.SignatureFee * int64(len(tx.Signatures))
	return baseFee + instructionFee + signatureFee
}

// CalculateFeeWithDefaults computes the fee using default configuration
func CalculateFeeWithDefaults(tx *Transaction) int64 {
	return CalculateFee(tx, DefaultFeeConfig())
}
