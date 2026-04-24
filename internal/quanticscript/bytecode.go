package quanticscript

import (
	"fmt"
)

// Bytecode format constants
const (
	// Magic number for QuanticScript bytecode files
	BytecodeMagic = 0x5153 // "QS" in hex

	// Version 1.0
	BytecodeVersion = 0x0100

	// Header size: magic (2 bytes) + version (2 bytes) + entry offset (4 bytes)
	BytecodeHeaderSize = 8
)

// BytecodeHeader represents the header of a QuanticScript bytecode file
type BytecodeHeader struct {
	Magic       uint16 // Magic number (0x5153 = "QS")
	Version     uint16 // Bytecode version
	EntryOffset uint32 // Offset to entry function
}

// IsQuanticScriptBytecode checks if the given data is valid QuanticScript bytecode
func IsQuanticScriptBytecode(data []byte) bool {
	if len(data) < BytecodeHeaderSize {
		return false
	}

	// Check magic number
	magic := uint16(data[0]) | uint16(data[1])<<8
	return magic == BytecodeMagic
}

// ParseBytecodeHeader parses the bytecode header
func ParseBytecodeHeader(data []byte) (*BytecodeHeader, error) {
	if len(data) < BytecodeHeaderSize {
		return nil, fmt.Errorf("bytecode too short: expected at least %d bytes, got %d", BytecodeHeaderSize, len(data))
	}

	magic := uint16(data[0]) | uint16(data[1])<<8
	if magic != BytecodeMagic {
		return nil, fmt.Errorf("invalid bytecode magic: expected 0x%04x, got 0x%04x", BytecodeMagic, magic)
	}

	version := uint16(data[2]) | uint16(data[3])<<8
	if version != BytecodeVersion {
		return nil, fmt.Errorf("unsupported bytecode version: expected 0x%04x, got 0x%04x", BytecodeVersion, version)
	}

	entryOffset := uint32(data[4]) | uint32(data[5])<<8 | uint32(data[6])<<16 | uint32(data[7])<<24

	return &BytecodeHeader{
		Magic:       magic,
		Version:     version,
		EntryOffset: entryOffset,
	}, nil
}

// GetBytecodeBody returns the bytecode body (without header)
func GetBytecodeBody(data []byte) ([]byte, error) {
	if len(data) < BytecodeHeaderSize {
		return nil, fmt.Errorf("bytecode too short")
	}

	return data[BytecodeHeaderSize:], nil
}

// CreateBytecode creates a bytecode file with header
func CreateBytecode(body []byte, entryOffset uint32) []byte {
	bytecode := make([]byte, BytecodeHeaderSize+len(body))

	// Write header
	bytecode[0] = byte(BytecodeMagic & 0xFF)
	bytecode[1] = byte((BytecodeMagic >> 8) & 0xFF)
	bytecode[2] = byte(BytecodeVersion & 0xFF)
	bytecode[3] = byte((BytecodeVersion >> 8) & 0xFF)
	bytecode[4] = byte(entryOffset & 0xFF)
	bytecode[5] = byte((entryOffset >> 8) & 0xFF)
	bytecode[6] = byte((entryOffset >> 16) & 0xFF)
	bytecode[7] = byte((entryOffset >> 24) & 0xFF)

	// Copy body
	copy(bytecode[BytecodeHeaderSize:], body)

	return bytecode
}
