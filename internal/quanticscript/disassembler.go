package quanticscript

import (
	"fmt"
	"strings"
)

// Disassembler converts bytecode to human-readable assembly text
type Disassembler struct {
	bytecode []byte
	pc       int // program counter
	output   strings.Builder
	labels   map[int]string // bytecode offset -> label name
}

// NewDisassembler creates a new disassembler instance
func NewDisassembler(bytecode []byte) *Disassembler {
	return &Disassembler{
		bytecode: bytecode,
		pc:       0,
		labels:   make(map[int]string),
	}
}

// Disassemble converts bytecode to assembly text with cost annotations
func (d *Disassembler) Disassemble() (string, error) {
	// First pass: identify jump targets and create labels
	if err := d.identifyLabels(); err != nil {
		return "", err
	}

	// Reset for second pass
	d.pc = 0
	d.output.Reset()

	// Second pass: generate assembly text
	for d.pc < len(d.bytecode) {
		// Check if there's a label at this position
		if label, ok := d.labels[d.pc]; ok {
			d.output.WriteString(fmt.Sprintf("\n%s:\n", label))
		}

		offset := d.pc
		if err := d.disassembleInstruction(); err != nil {
			return "", fmt.Errorf("offset %d: %w", offset, err)
		}
	}

	return d.output.String(), nil
}

// identifyLabels scans bytecode to find jump targets and create labels
func (d *Disassembler) identifyLabels() error {
	labelCounter := 0
	pc := 0

	for pc < len(d.bytecode) {
		if pc >= len(d.bytecode) {
			break
		}

		opcode := Opcode(d.bytecode[pc])
		pc++

		// Check for jump instructions
		switch opcode {
		case OpPush:
			// Skip type and value
			if pc >= len(d.bytecode) {
				return fmt.Errorf("unexpected end in PUSH")
			}
			valueType := ValueType(d.bytecode[pc])
			pc++
			switch valueType {
			case TypeI64, TypeU64:
				pc += 8
			case TypeBool:
				pc += 1
			default:
				return fmt.Errorf("unsupported value type: %v", valueType)
			}

		case OpLoad, OpStore:
			pc += 2 // Skip offset

		case OpJmp, OpJmpIf:
			if pc+4 > len(d.bytecode) {
				return fmt.Errorf("unexpected end in jump instruction")
			}
			offset := int32(d.bytecode[pc]) |
				int32(d.bytecode[pc+1])<<8 |
				int32(d.bytecode[pc+2])<<16 |
				int32(d.bytecode[pc+3])<<24
			pc += 4

			// Calculate target address (relative to position after offset)
			targetAddr := pc + int(offset)
			if targetAddr >= 0 && targetAddr < len(d.bytecode) {
				if _, exists := d.labels[targetAddr]; !exists {
					d.labels[targetAddr] = fmt.Sprintf("label_%d", labelCounter)
					labelCounter++
				}
			}

		case OpCall:
			if pc+4 > len(d.bytecode) {
				return fmt.Errorf("unexpected end in CALL")
			}
			offset := int32(d.bytecode[pc]) |
				int32(d.bytecode[pc+1])<<8 |
				int32(d.bytecode[pc+2])<<16 |
				int32(d.bytecode[pc+3])<<24
			pc += 4

			// CALL uses absolute offset
			targetAddr := int(offset)
			if targetAddr >= 0 && targetAddr < len(d.bytecode) {
				if _, exists := d.labels[targetAddr]; !exists {
					d.labels[targetAddr] = fmt.Sprintf("func_%d", labelCounter)
					labelCounter++
				}
			}
		}
	}

	return nil
}

// disassembleInstruction disassembles a single instruction
func (d *Disassembler) disassembleInstruction() error {
	if d.pc >= len(d.bytecode) {
		return fmt.Errorf("program counter out of bounds")
	}

	opcode := Opcode(d.bytecode[d.pc])
	d.pc++

	// Get instruction name
	name, ok := OpcodeNames[opcode]
	if !ok {
		return fmt.Errorf("unknown opcode: 0x%02x", opcode)
	}

	// Get instruction cost
	cost := GetInstructionCost(opcode)

	// Start building instruction line
	var instrLine strings.Builder
	instrLine.WriteString(fmt.Sprintf("    %-12s", name))

	// Add operands based on instruction type
	switch opcode {
	case OpPush:
		operand, err := d.disassemblePushOperand()
		if err != nil {
			return err
		}
		instrLine.WriteString(operand)

	case OpLoad, OpStore:
		operand, err := d.disassembleMemoryOperand()
		if err != nil {
			return err
		}
		instrLine.WriteString(operand)

	case OpJmp, OpJmpIf:
		operand, err := d.disassembleJumpOperand(true)
		if err != nil {
			return err
		}
		instrLine.WriteString(operand)

	case OpCall:
		operand, err := d.disassembleJumpOperand(false)
		if err != nil {
			return err
		}
		instrLine.WriteString(operand)

		// All other instructions have no operands
	}

	// Add cost annotation as comment
	instrLine.WriteString(fmt.Sprintf(" ; cost: %d", cost))

	d.output.WriteString(instrLine.String())
	d.output.WriteString("\n")

	return nil
}

// disassemblePushOperand disassembles PUSH operand
func (d *Disassembler) disassemblePushOperand() (string, error) {
	if d.pc >= len(d.bytecode) {
		return "", fmt.Errorf("unexpected end in PUSH operand")
	}

	valueType := ValueType(d.bytecode[d.pc])
	d.pc++

	switch valueType {
	case TypeI64:
		if d.pc+8 > len(d.bytecode) {
			return "", fmt.Errorf("unexpected end reading i64")
		}
		val := int64(d.bytecode[d.pc]) |
			int64(d.bytecode[d.pc+1])<<8 |
			int64(d.bytecode[d.pc+2])<<16 |
			int64(d.bytecode[d.pc+3])<<24 |
			int64(d.bytecode[d.pc+4])<<32 |
			int64(d.bytecode[d.pc+5])<<40 |
			int64(d.bytecode[d.pc+6])<<48 |
			int64(d.bytecode[d.pc+7])<<56
		d.pc += 8
		return fmt.Sprintf("i64 %d", val), nil

	case TypeU64:
		if d.pc+8 > len(d.bytecode) {
			return "", fmt.Errorf("unexpected end reading u64")
		}
		val := uint64(d.bytecode[d.pc]) |
			uint64(d.bytecode[d.pc+1])<<8 |
			uint64(d.bytecode[d.pc+2])<<16 |
			uint64(d.bytecode[d.pc+3])<<24 |
			uint64(d.bytecode[d.pc+4])<<32 |
			uint64(d.bytecode[d.pc+5])<<40 |
			uint64(d.bytecode[d.pc+6])<<48 |
			uint64(d.bytecode[d.pc+7])<<56
		d.pc += 8
		return fmt.Sprintf("u64 %d", val), nil

	case TypeBool:
		if d.pc >= len(d.bytecode) {
			return "", fmt.Errorf("unexpected end reading bool")
		}
		val := d.bytecode[d.pc] != 0
		d.pc++
		return fmt.Sprintf("bool %v", val), nil

	default:
		return "", fmt.Errorf("unsupported value type: %v", valueType)
	}
}

// disassembleMemoryOperand disassembles LOAD/STORE operand
func (d *Disassembler) disassembleMemoryOperand() (string, error) {
	if d.pc+2 > len(d.bytecode) {
		return "", fmt.Errorf("unexpected end in memory operand")
	}

	offset := uint16(d.bytecode[d.pc]) |
		uint16(d.bytecode[d.pc+1])<<8
	d.pc += 2

	return fmt.Sprintf("%d", offset), nil
}

// disassembleJumpOperand disassembles JMP/JMPIF/CALL operand
func (d *Disassembler) disassembleJumpOperand(isRelative bool) (string, error) {
	if d.pc+4 > len(d.bytecode) {
		return "", fmt.Errorf("unexpected end in jump operand")
	}

	offset := int32(d.bytecode[d.pc]) |
		int32(d.bytecode[d.pc+1])<<8 |
		int32(d.bytecode[d.pc+2])<<16 |
		int32(d.bytecode[d.pc+3])<<24
	d.pc += 4

	var targetAddr int
	if isRelative {
		// For JMP/JMPIF: offset is relative to position after reading offset
		targetAddr = d.pc + int(offset)
	} else {
		// For CALL: offset is absolute
		targetAddr = int(offset)
	}

	// Check if we have a label for this target
	if label, ok := d.labels[targetAddr]; ok {
		return label, nil
	}

	// Otherwise, show numeric offset
	return fmt.Sprintf("%d", offset), nil
}

// DisassembleFile disassembles a complete bytecode file (with header)
func DisassembleFile(data []byte) (string, error) {
	// Parse header
	header, err := ParseBytecodeHeader(data)
	if err != nil {
		return "", err
	}

	// Get bytecode body
	body, err := GetBytecodeBody(data)
	if err != nil {
		return "", err
	}

	// Disassemble body
	disasm := NewDisassembler(body)
	assembly, err := disasm.Disassemble()
	if err != nil {
		return "", err
	}

	// Add header information as comments
	var output strings.Builder
	output.WriteString("; QuanticScript Bytecode\n")
	output.WriteString(fmt.Sprintf("; Version: 0x%04x\n", header.Version))
	output.WriteString(fmt.Sprintf("; Entry Offset: %d\n", header.EntryOffset))
	output.WriteString("\n")
	output.WriteString(assembly)

	return output.String(), nil
}

// DisassembleBody disassembles bytecode body (without header)
func DisassembleBody(body []byte) (string, error) {
	disasm := NewDisassembler(body)
	return disasm.Disassemble()
}
