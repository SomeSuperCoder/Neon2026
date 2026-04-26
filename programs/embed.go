// Package programs exposes the compiled QuanticScript bytecode for the
// built-in programs as embedded byte slices. Keeping the embed declarations
// here (alongside the .qsb files) avoids duplicating the binaries and works
// around the Go restriction that //go:embed paths cannot contain "..".
package programs

import _ "embed"

// SystemProgram is the compiled bytecode for the System_Program.
//
//go:embed system/system.qsb
var SystemProgram []byte

// TokenProgram is the compiled bytecode for the Token_Program.
//
//go:embed token/token.qsb
var TokenProgram []byte

// StakingProgram is the compiled bytecode for the Staking_Program.
//
//go:embed staking/staking.qsb
var StakingProgram []byte
