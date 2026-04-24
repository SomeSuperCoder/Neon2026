package quanticscript

import (
	"fmt"
)

// ExampleAssembler demonstrates assembling QuanticScript assembly to bytecode
func ExampleAssembler() {
	assembly := `
		; Simple program that adds two numbers
		PUSH i64 10
		PUSH i64 20
		ADD
		RET
	`

	bytecode, err := AssembleToBody(assembly)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Assembled %d bytes of bytecode\n", len(bytecode))
	// Output: Assembled 22 bytes of bytecode
}

// ExampleDisassembler demonstrates disassembling bytecode to assembly
func ExampleDisassembler() {
	// First assemble some code
	assembly := `
		PUSH i64 42
		RET
	`

	bytecode, _ := AssembleToBody(assembly)

	// Now disassemble it
	disassembled, err := DisassembleBody(bytecode)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Println(disassembled)
	// Output:
	//     PUSH        i64 42 ; cost: 1
	//     RET          ; cost: 2
}

// ExampleAssembleToFile demonstrates creating a complete bytecode file
func ExampleAssembleToFile() {
	assembly := `
		PUSH i64 100
		PUSH i64 50
		SUB
		RET
	`

	bytecode, err := AssembleToFile(assembly)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Verify it's valid bytecode
	if IsQuanticScriptBytecode(bytecode) {
		fmt.Println("Valid QuanticScript bytecode file created")
	}
	// Output: Valid QuanticScript bytecode file created
}

// ExampleDisassembleFile demonstrates disassembling a complete bytecode file
func ExampleDisassembleFile() {
	// Create a bytecode file
	assembly := `
		PUSH i64 5
		PUSH i64 3
		MUL
		RET
	`

	bytecode, _ := AssembleToFile(assembly)

	// Disassemble it
	disassembled, err := DisassembleFile(bytecode)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Println(disassembled)
	// Output:
	// ; QuanticScript Bytecode
	// ; Version: 0x0100
	// ; Entry Offset: 0
	//
	//     PUSH        i64 5 ; cost: 1
	//     PUSH        i64 3 ; cost: 1
	//     MUL          ; cost: 3
	//     RET          ; cost: 2
}
