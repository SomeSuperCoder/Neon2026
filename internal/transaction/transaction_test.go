package transaction

import (
	"crypto/ed25519"
	"crypto/rand"
	"testing"

	"github.com/poh-blockchain/internal/filestore"
)

// TestAccessPermissionString tests the String method of AccessPermission
func TestAccessPermissionString(t *testing.T) {
	tests := []struct {
		perm     AccessPermission
		expected string
	}{
		{Read, "Read"},
		{Write, "Write"},
		{AccessPermission(99), "Unknown"},
	}

	for _, tt := range tests {
		if got := tt.perm.String(); got != tt.expected {
			t.Errorf("AccessPermission.String() = %v, want %v", got, tt.expected)
		}
	}
}

// TestFileAccessSerialization tests FileAccess JSON marshaling/unmarshaling
func TestFileAccessSerialization(t *testing.T) {
	fileID := filestore.GenerateFileID([]byte("test-file"))

	fa := FileAccess{
		FileID:     fileID,
		Permission: Write,
	}

	// Marshal
	data, err := fa.MarshalJSON()
	if err != nil {
		t.Fatalf("Failed to marshal FileAccess: %v", err)
	}

	// Unmarshal
	var fa2 FileAccess
	if err := fa2.UnmarshalJSON(data); err != nil {
		t.Fatalf("Failed to unmarshal FileAccess: %v", err)
	}

	// Verify
	if fa2.FileID != fa.FileID {
		t.Errorf("FileID mismatch: got %v, want %v", fa2.FileID, fa.FileID)
	}
	if fa2.Permission != fa.Permission {
		t.Errorf("Permission mismatch: got %v, want %v", fa2.Permission, fa.Permission)
	}
}

// TestSignatureSerialization tests Signature JSON marshaling/unmarshaling
func TestSignatureSerialization(t *testing.T) {
	// Generate a test key pair
	pubKey, privKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	var pk PublicKey
	copy(pk[:], pubKey)

	message := []byte("test message")
	sigBytes := ed25519.Sign(privKey, message)

	var sig Signature
	sig.PublicKey = pk
	copy(sig.Signature[:], sigBytes)

	// Marshal
	data, err := sig.MarshalJSON()
	if err != nil {
		t.Fatalf("Failed to marshal Signature: %v", err)
	}

	// Unmarshal
	var sig2 Signature
	if err := sig2.UnmarshalJSON(data); err != nil {
		t.Fatalf("Failed to unmarshal Signature: %v", err)
	}

	// Verify
	if sig2.PublicKey != sig.PublicKey {
		t.Errorf("PublicKey mismatch")
	}
	if sig2.Signature != sig.Signature {
		t.Errorf("Signature mismatch")
	}

	// Verify signature
	if !sig2.Verify(message) {
		t.Errorf("Signature verification failed")
	}
}

// TestInstructionSerialization tests Instruction JSON marshaling/unmarshaling
func TestInstructionSerialization(t *testing.T) {
	programID := filestore.GenerateFileID([]byte("system-program"))
	fileID1 := filestore.GenerateFileID([]byte("file1"))
	fileID2 := filestore.GenerateFileID([]byte("file2"))

	instr := Instruction{
		ProgramID: programID,
		Inputs: map[string]FileAccess{
			"from": {FileID: fileID1, Permission: Write},
			"to":   {FileID: fileID2, Permission: Write},
		},
		Data: []byte("transfer 1000"),
	}

	// Marshal
	data, err := instr.Marshal()
	if err != nil {
		t.Fatalf("Failed to marshal Instruction: %v", err)
	}

	// Unmarshal
	instr2, err := UnmarshalInstruction(data)
	if err != nil {
		t.Fatalf("Failed to unmarshal Instruction: %v", err)
	}

	// Verify
	if instr2.ProgramID != instr.ProgramID {
		t.Errorf("ProgramID mismatch")
	}
	if len(instr2.Inputs) != len(instr.Inputs) {
		t.Errorf("Inputs length mismatch: got %d, want %d", len(instr2.Inputs), len(instr.Inputs))
	}
	if string(instr2.Data) != string(instr.Data) {
		t.Errorf("Data mismatch: got %s, want %s", instr2.Data, instr.Data)
	}
}

// TestTransactionSerialization tests Transaction JSON marshaling/unmarshaling
func TestTransactionSerialization(t *testing.T) {
	// Generate test key pair
	pubKey, privKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	var pk PublicKey
	copy(pk[:], pubKey)

	// Create transaction
	programID := filestore.GenerateFileID([]byte("system-program"))
	fileID1 := filestore.GenerateFileID([]byte("file1"))
	fileID2 := filestore.GenerateFileID([]byte("file2"))
	lastSeenID := filestore.GenerateFileID([]byte("last-tx"))

	var lastSeen TxID
	copy(lastSeen[:], lastSeenID[:])

	tx := Transaction{
		LastSeen: lastSeen,
		Instructions: []Instruction{
			{
				ProgramID: programID,
				Inputs: map[string]FileAccess{
					"from": {FileID: fileID1, Permission: Write},
					"to":   {FileID: fileID2, Permission: Write},
				},
				Data: []byte("transfer 1000"),
			},
		},
		Signatures: []Signature{},
	}

	// Sign the transaction
	message := []byte("tx-message")
	sigBytes := ed25519.Sign(privKey, message)
	var sig Signature
	sig.PublicKey = pk
	copy(sig.Signature[:], sigBytes)
	tx.Signatures = append(tx.Signatures, sig)

	// Marshal
	data, err := tx.Marshal()
	if err != nil {
		t.Fatalf("Failed to marshal Transaction: %v", err)
	}

	// Unmarshal
	tx2, err := UnmarshalTransaction(data)
	if err != nil {
		t.Fatalf("Failed to unmarshal Transaction: %v", err)
	}

	// Verify
	if tx2.LastSeen != tx.LastSeen {
		t.Errorf("LastSeen mismatch")
	}
	if len(tx2.Instructions) != len(tx.Instructions) {
		t.Errorf("Instructions length mismatch: got %d, want %d", len(tx2.Instructions), len(tx.Instructions))
	}
	if len(tx2.Signatures) != len(tx.Signatures) {
		t.Errorf("Signatures length mismatch: got %d, want %d", len(tx2.Signatures), len(tx.Signatures))
	}
}

// TestTransactionGetFeePayer tests the GetFeePayer method
func TestTransactionGetFeePayer(t *testing.T) {
	// Generate test key pairs
	pubKey1, _, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}
	pubKey2, _, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	var pk1, pk2 PublicKey
	copy(pk1[:], pubKey1)
	copy(pk2[:], pubKey2)

	// Transaction with no signatures
	tx1 := Transaction{
		Signatures: []Signature{},
	}
	_, err = tx1.GetFeePayer()
	if err == nil {
		t.Errorf("Expected error for transaction with no signatures")
	}

	// Transaction with signatures
	tx2 := Transaction{
		Signatures: []Signature{
			{PublicKey: pk1},
			{PublicKey: pk2},
		},
	}
	feePayer, err := tx2.GetFeePayer()
	if err != nil {
		t.Fatalf("Failed to get fee payer: %v", err)
	}
	if feePayer != pk1 {
		t.Errorf("Fee payer mismatch: expected first signature's public key")
	}
}

// TestTransactionGetSigners tests the GetSigners method
func TestTransactionGetSigners(t *testing.T) {
	// Generate test key pairs
	pubKey1, _, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}
	pubKey2, _, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	var pk1, pk2 PublicKey
	copy(pk1[:], pubKey1)
	copy(pk2[:], pubKey2)

	tx := Transaction{
		Signatures: []Signature{
			{PublicKey: pk1},
			{PublicKey: pk2},
		},
	}

	signers := tx.GetSigners()
	if len(signers) != 2 {
		t.Errorf("Expected 2 signers, got %d", len(signers))
	}
	if signers[0] != pk1 || signers[1] != pk2 {
		t.Errorf("Signers mismatch")
	}
}

// TestDefaultFeeConfig tests the default fee configuration
func TestDefaultFeeConfig(t *testing.T) {
	config := DefaultFeeConfig()

	if config.BaseFee != 5000 {
		t.Errorf("Expected BaseFee 5000, got %d", config.BaseFee)
	}
	if config.InstructionFee != 1000 {
		t.Errorf("Expected InstructionFee 1000, got %d", config.InstructionFee)
	}
	if config.SignatureFee != 500 {
		t.Errorf("Expected SignatureFee 500, got %d", config.SignatureFee)
	}
}

// TestCalculateFee tests the fee calculation function
func TestCalculateFee(t *testing.T) {
	programID := filestore.GenerateFileID([]byte("system-program"))
	fileID1 := filestore.GenerateFileID([]byte("file1"))
	fileID2 := filestore.GenerateFileID([]byte("file2"))

	tests := []struct {
		name        string
		tx          *Transaction
		config      FeeConfig
		expectedFee int64
	}{
		{
			name: "Single instruction, single signature",
			tx: &Transaction{
				Instructions: []Instruction{
					{
						ProgramID: programID,
						Inputs: map[string]FileAccess{
							"from": {FileID: fileID1, Permission: Write},
						},
						Data: []byte("test"),
					},
				},
				Signatures: []Signature{
					{PublicKey: PublicKey{}},
				},
			},
			config:      DefaultFeeConfig(),
			expectedFee: 5000 + 1000 + 500, // base + 1 instruction + 1 signature = 6500
		},
		{
			name: "Multiple instructions, multiple signatures",
			tx: &Transaction{
				Instructions: []Instruction{
					{
						ProgramID: programID,
						Inputs: map[string]FileAccess{
							"from": {FileID: fileID1, Permission: Write},
						},
						Data: []byte("test1"),
					},
					{
						ProgramID: programID,
						Inputs: map[string]FileAccess{
							"to": {FileID: fileID2, Permission: Write},
						},
						Data: []byte("test2"),
					},
					{
						ProgramID: programID,
						Inputs: map[string]FileAccess{
							"account": {FileID: fileID1, Permission: Read},
						},
						Data: []byte("test3"),
					},
				},
				Signatures: []Signature{
					{PublicKey: PublicKey{}},
					{PublicKey: PublicKey{}},
				},
			},
			config:      DefaultFeeConfig(),
			expectedFee: 5000 + 3000 + 1000, // base + 3 instructions + 2 signatures = 9000
		},
		{
			name: "Custom fee configuration",
			tx: &Transaction{
				Instructions: []Instruction{
					{
						ProgramID: programID,
						Inputs:    map[string]FileAccess{},
						Data:      []byte("test"),
					},
				},
				Signatures: []Signature{
					{PublicKey: PublicKey{}},
				},
			},
			config: FeeConfig{
				BaseFee:        10000,
				InstructionFee: 2000,
				SignatureFee:   1000,
			},
			expectedFee: 10000 + 2000 + 1000, // custom base + 1 instruction + 1 signature = 13000
		},
		{
			name: "Empty transaction",
			tx: &Transaction{
				Instructions: []Instruction{},
				Signatures:   []Signature{},
			},
			config:      DefaultFeeConfig(),
			expectedFee: 5000, // only base fee
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fee := CalculateFee(tt.tx, tt.config)
			if fee != tt.expectedFee {
				t.Errorf("CalculateFee() = %d, want %d", fee, tt.expectedFee)
			}
		})
	}
}

// TestCalculateFeeWithDefaults tests the convenience function with default config
func TestCalculateFeeWithDefaults(t *testing.T) {
	programID := filestore.GenerateFileID([]byte("system-program"))
	fileID1 := filestore.GenerateFileID([]byte("file1"))

	tx := &Transaction{
		Instructions: []Instruction{
			{
				ProgramID: programID,
				Inputs: map[string]FileAccess{
					"from": {FileID: fileID1, Permission: Write},
				},
				Data: []byte("test"),
			},
		},
		Signatures: []Signature{
			{PublicKey: PublicKey{}},
		},
	}

	fee := CalculateFeeWithDefaults(tx)
	expectedFee := int64(5000 + 1000 + 500) // base + 1 instruction + 1 signature

	if fee != expectedFee {
		t.Errorf("CalculateFeeWithDefaults() = %d, want %d", fee, expectedFee)
	}
}
