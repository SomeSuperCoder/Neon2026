# Changelog

All notable changes to the PoH Blockchain project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- DPoS integration in `cmd/main.go` — wired ConsensusManager, FileStore, and Runtime for full DPoS support
  - Node startup now uses `NewConsensusManagerWithGenesis` with genesis validator configuration
  - Default 2-validator genesis setup (10 Neon and 5 Neon stakes) for development
  - Epoch length configured to 432,000 slots (~2 days at 400ms/slot)
  - Automatic DPoS genesis initialization before slot processing
  - FileStore and Runtime properly wired to ConsensusManager for epoch processing
  - Staking bytecode loaded via `LoadBuiltinPrograms` at node startup
  - Integration test (`internal/dpos_wiring_integration_test.go`) validates complete DPoS lifecycle
- Validator TUI dashboard (`cmd/validator-tui/main.go`) — real-time monitoring of validator and staking state
  - Built with Bubble Tea framework and lipgloss for rich terminal UI
  - Read-only FileStore access with BadgerDB
  - File classification for validator records, stake accounts, epoch state, and reward pool
  - Terminal dashboard with styled components and 1-second refresh interval
  - Keyboard controls (q/Ctrl+C for exit)
  - Displays epoch, slot, validator status, active validator count, and delegated stake
  - Shows validator table with commission, blocks produced, missed blocks, and slashing status
  - Summary footer with total staked electrons, reward pool balance, and estimated APY
  - Automatic state directory creation and genesis initialization
  - Comprehensive guide in `docs/guides/validator-tui.md`
- Security restriction on `UPDATEBALANCE` instruction - can only be called by system program (2026-04-24)
- Comprehensive security model documentation in `docs/reference/security-model.md`
- System program privilege level with FileID `0x01`
- Debug logging infrastructure for bytecode interpreter (2026-04-25)
- Safety limit of 1000 execution steps to prevent infinite loops (2026-04-25)
- Parser infinite loop fix documentation in `docs/reference/parser-infinite-loop-fix.md` (2026-04-25)
- Support for i64 file identifiers in `UPDATEBALANCE` instruction (2026-04-25)
- `len()` builtin function for getting byte array length in QuanticScript (2026-04-25)
- Genesis loader (`internal/genesis`) — idempotent bootstrap of System_Program and Token_Program at node startup
- Parallel execution analyzer (`internal/parallel`) — conflict detection for transaction scheduling
- QuanticScript standard library complete: string, math, crypto, blockchain, collections, invoke modules
- `DISPATCH` opcode and instruction registry for smart contract routing
- Privileged opcodes: `OpCreateFile`, `OpCreateFileWithID`, `OpDeleteFile`, `OpTransferBalance`
- `OpSlice`, `OpBytesLen`, `OpBytesAppend`, `OpBytesToFileID` byte manipulation opcodes
- `OpBytesToI64LE` for little-endian decoding of instruction data
- System_Program written in QuanticScript (`programs/system/system.qs`)
- Token_Program written in QuanticScript (`programs/token/token.qs`) with 11 instruction handlers
- Serialization helpers for MintAccount and TokenAccount in `stdlib_programs.go`
- Instruction data parsing utilities: `ParseInstructionU8/U64/I64/FileID/PublicKey/Bool`
- Delegated Proof of Stake (DPoS) spec added to `.kiro/specs/delegated-proof-of-stake/`

### Changed
- `execUpdateBalance()` now validates caller is system program before execution
- `UPDATEBALANCE` instruction now accepts both FileID and i64 types for file identifier parameter (2026-04-25)
- Updated bytecode reference documentation with security notes
- Updated standard library reference with security warnings
- Enhanced example code comments to clarify security restrictions

### Fixed
- Parser infinite loop bug in `parseBlockStmt()` when statement parsing fails (2026-04-25)
- Parser now advances tokens when stuck to prevent infinite loops
- Interpreter execution now has safety limit to prevent runaway execution

### Security
- **BREAKING**: Regular programs can no longer call `UPDATEBALANCE` directly
- Programs must use cross-program invocation to call system program for balance transfers
- Prevents unauthorized balance manipulation and minting attacks

## [1.0.0] - 2026-04-24

### Added
- Complete PoH blockchain implementation
- QuanticScript language with TypeScript-like syntax
- Bytecode interpreter with cost metering
- Cross-program invocation with depth tracking
- Assembler and disassembler for bytecode
- Lexer, parser, type checker, and code generator
- Byzantine Fault Tolerance testing capabilities
- Automated testing tools
- Comprehensive documentation

### Features
- Proof of History clock generator
- Leader-based consensus protocol
- P2P network communication
- SQLite-based persistent ledger
- File-based state model
- Transaction processing with atomic execution
- System program for account management
- CLI tools for account operations

[Unreleased]: https://github.com/poh-blockchain/compare/v1.0.0...HEAD
[1.0.0]: https://github.com/poh-blockchain/releases/tag/v1.0.0
