---
inclusion: always
---

# Product Overview

PoH Blockchain is a Proof of History blockchain implementation inspired by Solana's architecture, built in Go. It uses sequential SHA-256 hashing as a verifiable delay function to create a cryptographic clock, enabling high-throughput transaction ordering without traditional consensus overhead.

## Core Components

- **PoH Clock**: Sequential hash chain serving as a cryptographic timeline
- **Leader-based Consensus**: 400ms slots with Byzantine Fault Tolerance
- **File-Based State Model**: Uniform abstraction for accounts, programs, and data (inspired by Solana's account model)
- **QuanticScript Language**: TypeScript-like smart contract language with deterministic execution
- **P2P Network**: TCP-based block distribution and validation
- **SQLite Ledger**: Persistent blockchain storage with full verification

## Key Features

- Verifiable sequential hashing with tick-based timestamping
- Byzantine Fault Tolerance with malicious node testing
- Cost-metered bytecode execution with exponential storage costs
- Cross-program invocation with depth tracking
- Inline assembly support for performance optimization
- Rich standard library for crypto, blockchain, and query operations

## Development Status

Production-ready: PoH clock, consensus, networking, ledger, file-based state, transaction processing, System Program

In development: QuanticScript compiler (lexer, parser, type checker, code generator complete; standard library in progress)
