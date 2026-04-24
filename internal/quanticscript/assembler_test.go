package quanticscript

import (
	"testing"
)

func TestAssemblerBasicInstructions(t *testing.T) {
	tests := []struct {
		name     string
		assembly string
		wantErr  bool
	}{
		{
			name: "simple push and add",
			assembly: `
				PUSH i64 10
				PUSH i64 20
				ADD
				RET
			`,
			wantErr: false,
		},
		{
			name: "push bool",
			assembly: `
				PUSH bool true
				PUSH bool false
				AND
				RET
			`,
			wantErr: false,
		},
		{
			name: "memory operations",
			assembly: `
				PUSH i64 42
				STORE 0
				LOAD 0
				RET
			`,
			wantErr: false,
		},
		{
			name: "stack operations",
			assembly: `
				PUSH i64 1
				DUP
				SWAP
				POP
				RET
			`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := AssembleToBody(tt.assembly)
			if (err != nil) != tt.wantErr {
				t.Errorf("Assemble() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAssemblerLabels(t *testing.T) {
	assembly := `
		PUSH i64 10
		PUSH i64 0
		EQ
		JMPIF skip
		PUSH i64 99
	skip:
		RET
	`

	bytecode, err := AssembleToBody(assembly)
	if err != nil {
		t.Fatalf("Assemble() error = %v", err)
	}

	if len(bytecode) == 0 {
		t.Error("Expected non-empty bytecode")
	}
}

func TestAssemblerFunctionCall(t *testing.T) {
	assembly := `
		CALL add_func
		RET
	
	add_func:
		PUSH i64 1
		PUSH i64 2
		ADD
		RET
	`

	bytecode, err := AssembleToBody(assembly)
	if err != nil {
		t.Fatalf("Assemble() error = %v", err)
	}

	if len(bytecode) == 0 {
		t.Error("Expected non-empty bytecode")
	}
}

func TestAssemblerErrors(t *testing.T) {
	tests := []struct {
		name     string
		assembly string
	}{
		{
			name:     "undefined label",
			assembly: "JMP undefined_label",
		},
		{
			name:     "invalid opcode",
			assembly: "INVALID_OP",
		},
		{
			name:     "wrong operand count",
			assembly: "PUSH i64",
		},
		{
			name:     "invalid value type",
			assembly: "PUSH invalid 42",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := AssembleToBody(tt.assembly)
			if err == nil {
				t.Error("Expected error but got none")
			}
		})
	}
}

func TestAssemblerComments(t *testing.T) {
	assembly := `
		; This is a comment
		PUSH i64 42  ; inline comment
		RET
	`

	bytecode, err := AssembleToBody(assembly)
	if err != nil {
		t.Fatalf("Assemble() error = %v", err)
	}

	if len(bytecode) == 0 {
		t.Error("Expected non-empty bytecode")
	}
}

func TestAssembleToFile(t *testing.T) {
	assembly := `
		PUSH i64 100
		RET
	`

	bytecode, err := AssembleToFile(assembly)
	if err != nil {
		t.Fatalf("AssembleToFile() error = %v", err)
	}

	// Verify header
	if !IsQuanticScriptBytecode(bytecode) {
		t.Error("Expected valid QuanticScript bytecode")
	}

	header, err := ParseBytecodeHeader(bytecode)
	if err != nil {
		t.Fatalf("ParseBytecodeHeader() error = %v", err)
	}

	if header.Magic != BytecodeMagic {
		t.Errorf("Expected magic 0x%04x, got 0x%04x", BytecodeMagic, header.Magic)
	}
}
