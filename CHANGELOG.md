# Changelog

All notable changes to the PoH Blockchain project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Security restriction on `UPDATEBALANCE` instruction - can only be called by system program (2026-04-24)
- Comprehensive security model documentation in `docs/reference/security-model.md`
- System program privilege level with FileID `0x01`
- Debug logging infrastructure for bytecode interpreter (2026-04-25)
- Safety limit of 1000 execution steps to prevent infinite loops (2026-04-25)
- Parser infinite loop fix documentation in `docs/reference/parser-infinite-loop-fix.md` (2026-04-25)
- Support for i64 file identifiers in `UPDATEBALANCE` instruction (2026-04-25)
- `len()` builtin function for getting byte array length in QuanticScript (2026-04-25)

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
