//go:build ignore

package main

import (
	"fmt"

	"github.com/poh-blockchain/internal/quanticscript"
)

func main() {
	source := `
		function transferTokens(fromAccount: i64, toAccount: i64, amount: i64): i64 {
			updateBalance(fromAccount, -amount);
			updateBalance(toAccount, amount);
			return 0;
		}
		
		export function entry(): i64 {
			return transferTokens(100, 200, 1000);
		}
	`

	fmt.Println("Starting lexer...")
	lexer := quanticscript.NewLexer(source, "test.qs")

	fmt.Println("Starting parser...")
	parser := quanticscript.NewParser(lexer)

	fmt.Println("Parsing program...")
	program := parser.ParseProgram()

	fmt.Printf("Parser complete, errors: %d\n", len(parser.Errors()))
	if len(parser.Errors()) > 0 {
		for _, err := range parser.Errors() {
			fmt.Printf("  Error: %s\n", err)
		}
	}

	fmt.Printf("Program has %d declarations\n", len(program.Declarations))
}
