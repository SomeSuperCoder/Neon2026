# Testing & Demo Summary

## Available Demo Scripts

### 1. Regular Demo (`demo.sh`)
**Purpose**: Run a basic blockchain network with honest nodes

**Usage**:
```bash
./demo.sh           # 1 leader + 2 replicas (default)
./demo.sh 5         # 1 leader + 5 replicas
```

**Features**:
- All nodes are honest
- Tests basic blockchain functionality
- Verifies block production and propagation
- Demonstrates P2P networking

---

### 2. BFT Testing Demo (`demo-bft.sh`)
**Purpose**: Test Byzantine Fault Tolerance with malicious nodes

**Usage**:
```bash
./demo-bft.sh 3 1   # 3 honest + 1 malicious (has BFT)
./demo-bft.sh 4 2   # 4 honest + 2 malicious (has BFT)
./demo-bft.sh 2 2   # 2 honest + 2 malicious (NO BFT)
```

**Features**:
- Mix of honest and malicious nodes
- Automatic BFT calculation
- Color-coded output (green=honest, red=malicious)
- Tests network resilience

---

### 3. Stop Script (`stop-demo.sh`)
**Purpose**: Stop any running demo and optionally clean up

**Usage**:
```bash
./stop-demo.sh
```

**Features**:
- Stops both regular and BFT demos
- Optional database cleanup
- Graceful shutdown

---

## Command-Line Flags

### Node Configuration
```bash
--type=leader|replica    # Node type (default: replica)
--port=8000             # Port to listen on (default: 8080)
--peers=host:port       # Comma-separated peer list
--db=path/to/db         # Database file path
--malicious             # Enable malicious mode (default: false)
```

### Examples
```bash
# Honest leader
./bin/poh-node --type=leader --port=8000 --db=leader.db

# Honest replica
./bin/poh-node --type=replica --port=8001 --peers=localhost:8000 --db=replica.db

# Malicious replica
./bin/poh-node --type=replica --port=8002 --peers=localhost:8000 --db=malicious.db --malicious
```

---

## Malicious Behaviors

### Leader Malicious Behaviors
| Frequency | Behavior | Impact |
|-----------|----------|--------|
| Every 3rd block | Invalid hash count (NumHashes=1) | PoH verification fails |
| Every 5th block | Wrong previous hash | Chain linkage breaks |
| Corrupted blocks | Skip local storage | Inconsistent state |
| All blocks | Broadcast corrupted data | Tests replica validation |

### Replica Malicious Behaviors
| Frequency | Behavior | Impact |
|-----------|----------|--------|
| Every 4th block | Skip all validation | Accept any block |
| Every 6th block | Ignore consensus failures | Store invalid blocks |
| Every 6th block | Ignore verification failures | Store unverified blocks |
| Every 6th block | Ignore linkage failures | Break chain continuity |

---

## BFT Requirements

### Formula
```
Honest Nodes > 2 × Malicious Nodes
```

### Examples
| Honest | Malicious | Total | BFT Status | Reason |
|--------|-----------|-------|------------|--------|
| 4 | 1 | 5 | ✓ Has BFT | 4 > 2×1 |
| 3 | 1 | 4 | ✓ Has BFT | 3 > 2×1 |
| 5 | 2 | 7 | ✓ Has BFT | 5 > 2×2 |
| 2 | 1 | 3 | ✗ No BFT | 2 = 2×1 |
| 2 | 2 | 4 | ✗ No BFT | 2 ≤ 2×2 |
| 1 | 2 | 3 | ✗ No BFT | 1 < 2×2 |

---

## Test Scenarios

### Scenario 1: Verify Basic Functionality
```bash
./demo.sh 3
```
**Goal**: Ensure blockchain works correctly with honest nodes
**Expected**: All nodes maintain identical chain state

### Scenario 2: Test BFT Resilience
```bash
./demo-bft.sh 4 1
```
**Goal**: Verify network handles Byzantine faults
**Expected**: Honest nodes reject invalid blocks, maintain consensus

### Scenario 3: Demonstrate BFT Failure
```bash
./demo-bft.sh 2 2
```
**Goal**: Show what happens without BFT threshold
**Expected**: Network may accept invalid blocks, lose consensus

### Scenario 4: Stress Test BFT
```bash
./demo-bft.sh 6 2
```
**Goal**: Test strong BFT with multiple malicious nodes
**Expected**: Network easily handles Byzantine faults

---

## Integration Tests

Run comprehensive integration tests:

```bash
# All tests
go test -v ./internal

# Specific test
go test -v ./internal -run TestFullNodeBlockProductionAndVerification
go test -v ./internal -run TestLeaderReplicaCommunication
go test -v ./internal -run TestLedgerPersistenceAndRecovery

# Genesis and DPoS tests
go test -v ./internal/genesis
go test -v ./internal/genesis -run TestInitializeDPoSGenesis
go test -v ./internal/genesis -run TestLoadBuiltinPrograms

# Transaction input validation tests
go test -v ./internal/processor -run TestValidate
go test -v ./internal/transaction -run TestCreateTransferInstruction
go test -v ./internal/transaction -run TestTransactionBuilder

# Validation integration tests
go test -v ./internal -run TestEndToEndTransferWithProperInputDeclarations
go test -v ./internal -run TestTransactionRejection
go test -v ./internal -run TestNoStateChangesOnValidationFailure
go test -v ./internal -run TestMultiInstructionTransactionValidation
```

**Test Coverage**:
1. Full node initialization and block production
2. Leader-replica communication and block propagation
3. Ledger persistence and recovery after restart
4. DPoS genesis initialization with validator records
5. Epoch state and reward pool file creation
6. Builtin program loading (System, Token, Staking)
7. Transaction input validation and permission checking
8. Transaction builder with proper input declarations
9. System Program helper functions for instruction creation
10. End-to-end validation flow with proper input declarations
11. Transaction rejection scenarios (missing program, incorrect permissions, non-existent files, non-executable programs)
12. State isolation on validation failures (no state changes except fee deduction)
13. Multi-instruction transaction validation and atomicity

---

## QuanticScript Tests

Run QuanticScript language and bytecode tests:

```bash
# All QuanticScript tests
go test -v ./internal/quanticscript

# Specific test categories
go test -v ./internal/quanticscript -run TestLexer
go test -v ./internal/quanticscript -run TestParser
go test -v ./internal/quanticscript -run TestTypeChecker
go test -v ./internal/quanticscript -run TestCodeGen
go test -v ./internal/quanticscript -run TestInterpreter
go test -v ./internal/quanticscript -run TestBytecodeHelpers
go test -v ./internal/quanticscript -run TestArithmeticOperations
```

**Test Coverage**:
1. Lexer tokenization and source location tracking
2. Parser AST construction and error handling
3. Type checker inference and validation
4. Code generator bytecode emission
5. Interpreter bytecode execution
6. Bytecode helper functions for test construction
7. Arithmetic operations: ADD, SUB, MUL, DIV, MOD for i64 and u64 (`arithmetic_test.go`)
8. Comparison operations: EQ, LT, GT, LTE, GTE for numbers; logical AND, OR, NOT; operator precedence (`comparison_test.go`)
9. Blockchain state operations: GETFILE, UPDATEFILE, GETBALANCE, GETSIGNER, HASSIGNER, GETINSTRDATA (`comprehensive_test.go`)
10. Cross-program invocation: basic invocation, depth tracking, permission enforcement, compute budget deduction (`cross_program_test.go`)
11. String operations: STRCONCAT with non-empty strings, empty first/second string, and two empty strings (`comprehensive_test.go`)
12. Cryptographic hashing: SHA256 with known input ("hello"), deterministic output verification, and empty input (`comprehensive_test.go`)
13. Function calls: basic CALL/RET with return address handling, and multi-value stack verification after call returns (`comprehensive_test.go`)

### Bytecode Test Helpers

The `comprehensive_test.go` file provides helper functions for building bytecode in tests:

- `buildPushI64(value)` - Create PUSH instruction for i64
- `buildPushU64(value)` - Create PUSH instruction for u64
- `buildPushBool(value)` - Create PUSH instruction for bool
- `buildPushString(value)` - Create PUSH instruction for string
- `buildPushBytes(value)` - Create PUSH instruction for bytes
- `buildJump(offset)` - Create JMP instruction with offset
- `buildJumpIf(offset)` - Create JMPIF instruction with offset
- `buildLoad(offset)` - Create LOAD instruction for memory access
- `buildStore(offset)` - Create STORE instruction for memory access
- `buildCall(offset)` - Create CALL instruction for function calls

These helpers simplify bytecode construction for comprehensive interpreter tests covering control flow, data structures, arithmetic, blockchain state, cross-program invocation, string operations, cryptographic operations, and function calls.

### Transaction Input Validation Tests

The transaction input validation system includes comprehensive tests in `internal/processor/input_validator_test.go`, `internal/transaction/`, and `internal/validation_integration_test.go`:

```bash
# Input validator tests
go test -v ./internal/processor -run TestValidateInstructionInputs
go test -v ./internal/processor -run TestValidateExecutableProgram
go test -v ./internal/processor -run TestValidateFileExists

# Transaction builder tests
go test -v ./internal/transaction -run TestTransactionBuilder
go test -v ./internal/transaction -run TestAddTransferInstruction

# System helpers tests
go test -v ./internal/transaction -run TestCreateTransferInstruction
go test -v ./internal/transaction -run TestEncodeTransferData

# Validation integration tests
go test -v ./internal -run TestEndToEndTransferWithProperInputDeclarations
go test -v ./internal -run TestTransactionRejection
go test -v ./internal -run TestNoStateChangesOnValidationFailure
go test -v ./internal -run TestMultiInstructionTransactionValidation
```

**Input Validator Test Coverage** (`input_validator_test.go`):
1. `TestValidateInstructionInputs_ValidInputs` - Validates properly declared inputs with correct permissions
2. `TestValidateInstructionInputs_MissingProgramDeclaration` - Detects missing program in inputs map
3. `TestValidateInstructionInputs_NonExecutableProgram` - Detects program without Executable flag
4. `TestValidateInstructionInputs_MissingFile` - Detects non-existent files in inputs
5. `TestValidateInstructionInputs_InvalidPermission` - Detects invalid permission values (not Read/Write)
6. `TestValidateInstructionInputs_EmptyInputsMap` - Rejects instructions with empty inputs
7. `TestValidateExecutableProgram` - Validates program executable flag separately
8. `TestValidateFileExists` - Validates file existence separately

**Transaction Builder Test Coverage** (`builder_test.go`):
1. `TestNewTransactionBuilder` - Constructor initialization
2. `TestAddTransferInstruction` - Adding transfer with proper input declarations
3. `TestAddTransferInstructionZeroAmount` - Rejection of zero amounts
4. `TestAddTransferInstructionNegativeAmount` - Rejection of negative amounts
5. `TestAddSignature` - Adding signatures to transactions
6. `TestBuild` - Building complete transactions
7. `TestMultipleInstructions` - Multi-instruction transaction support

**System Helpers Test Coverage** (`system_helpers_test.go`):
1. `TestCreateTransferInstruction` - Creates transfer with proper inputs and permissions
2. `TestCreateTransferInstructionZeroAmount` - Validates amount is positive
3. `TestEncodeTransferData` - Encodes transfer data in correct format (73 bytes)
4. `TestInputKeyConstants` - Verifies standard input key constants

**Validation Integration Test Coverage** (`validation_integration_test.go`):
1. `TestEndToEndTransferWithProperInputDeclarations` - Complete transfer flow with proper input declarations
2. `TestTransactionRejectionMissingProgramDeclaration` - Rejects transactions without program declaration
3. `TestTransactionRejectionIncorrectPermissions` - Rejects transactions with incorrect permissions (e.g., Read instead of Write)
4. `TestTransactionRejectionNonExistentFiles` - Rejects transactions referencing non-existent files
5. `TestTransactionRejectionNonExecutableProgram` - Rejects transactions invoking non-executable programs
6. `TestNoStateChangesOnValidationFailure` - Verifies no state changes occur when validation fails
7. `TestMultiInstructionTransactionValidation` - Validates multi-instruction transactions with proper declarations
8. `TestMultiInstructionTransactionPartialValidationFailure` - Rejects entire transaction if any instruction fails validation

**Error Types**:
- `ErrMissingInput` - Required file not declared in instruction inputs
- `ErrInvalidPermission` - Invalid permission value (must be Read=1 or Write=2)
- `ErrProgramNotExecutable` - Program file lacks Executable flag
- `ErrFileNotFound` - Input file doesn't exist in file store

**Validation Flow**:
1. TransactionBuilder creates instructions with proper input declarations
2. InputValidator checks all inputs before execution
3. Transaction processor rejects invalid instructions before state changes
4. All file accesses must be explicitly declared with correct permissions
5. Failed validation prevents any state modifications (except fee deduction)

### Cross-Program Invocation Tests (`cross_program_test.go`)

`InvokableMockContext` extends `MockExecutionContext` with full CPI support:

- `RegisterProgram(id, handler)` - registers a mock program and declares it as invocable
- `InvokeProgram(id, data, budget, depth)` - dispatches to registered handler, records history
- `GetDeclaredPrograms()` - returns the list of declared program IDs
- `GetInvocationHistory()` - returns all recorded `InvocationRecord` entries

Test functions:
- `TestCrossProgramBasicInvocation` - result passing, data forwarding, invocation history recording
- `TestCrossProgramDepthTracking` - depth increments, nested calls, `MaxInvokeDepth` enforcement
- `TestCrossProgramPermissions` - undeclared program rejection, budget deduction, insufficient budget error, error propagation

### Bytecode Verification Tests (`bytecode_verification_test.go`)

Verifies the compiled bytecode for both builtin programs against requirements 3.3, 3.4, 3.5.

```bash
go test -v ./internal/quanticscript -run TestBytecodeVerification
go test -v ./internal/quanticscript -run TestComputeBudget
go test -v ./internal/quanticscript -run TestDeterminism
go test -v ./internal/quanticscript -run TestPerformance
```

Test functions:
- `TestBytecodeVerification_SystemProgram` - validates header magic/version, disassembly contains CALL/RET/cost annotations
- `TestBytecodeVerification_TokenProgram` - same structural checks for Token_Program
- `TestBytecodeVerification_TokenLargerThanSystem` - asserts Token_Program bytecode is >= System_Program (more handlers)
- `TestBytecodeVerification_DisassemblyRoundTrip` - disassembles then re-assembles both programs and compares byte-for-byte
- `TestBytecodeVerification_InstructionCosts` - asserts every instruction line in the disassembly has a `cost:` annotation
- `TestComputeBudget_SystemProgram` - executes System_Program and verifies non-zero, bounded budget consumption (<10,000 units)
- `TestComputeBudget_TokenProgram` - executes Token_Program and verifies non-zero budget consumption
- `TestDeterminism_SystemProgram` - runs System_Program 10 times and asserts identical budget and stack state each run
- `TestDeterminism_TokenProgram` - same determinism check for Token_Program
- `TestPerformance_ProgramExecution` - runs each program 100 times and asserts average execution under 1ms

---

## DPoS Genesis Tests (`internal/genesis/programs_test.go`)

Tests for Delegated Proof of Stake genesis initialization and builtin program loading.

```bash
# All genesis tests
go test -v ./internal/genesis

# Specific test categories
go test -v ./internal/genesis -run TestInitializeDPoSGenesis
go test -v ./internal/genesis -run TestLoadBuiltinPrograms
go test -v ./internal/genesis -run TestEpochStateFileStructure
go test -v ./internal/genesis -run TestRewardPoolFileStructure
```

**Test Coverage**:
1. DPoS genesis initialization with multiple validators
2. Validator record creation and serialization
3. Epoch state file structure and deserialization
4. Reward pool file structure and deserialization
5. Builtin program loading (System, Token, Staking)
6. Idempotent initialization (skip if already initialized)
7. Error handling (zero validators, serialization failures)
8. Storage cost allocation for genesis files

**Test Functions**:
- `TestInitializeDPoSGenesis` - full genesis initialization with 3 validators, verifies all files created correctly
- `TestInitializeDPoSGenesisZeroValidators` - rejects genesis config with zero validators
- `TestInitializeDPoSGenesisIdempotent` - verifies second initialization call is skipped
- `TestLoadBuiltinProgramsWithStaking` - loads System, Token, and Staking programs
- `TestLoadBuiltinProgramsWithoutStaking` - loads only System and Token programs (nil staking bytecode)
- `TestEpochStateFileStructure` - verifies Epoch State File properties and deserialization
- `TestRewardPoolFileStructure` - verifies Reward Pool File properties and deserialization

---

## Verification Commands

### Check Chain Heights
```bash
# Honest nodes (should match)
sqlite3 replica1.db "SELECT COUNT(*) FROM blocks;"
sqlite3 replica2.db "SELECT COUNT(*) FROM blocks;"

# Malicious nodes (may differ)
sqlite3 malicious1.db "SELECT COUNT(*) FROM blocks;"
```

### View Block Details
```bash
sqlite3 leader.db "SELECT block_height, slot, timestamp FROM blocks ORDER BY block_height;"
```

### Check Chain Integrity
```bash
sqlite3 replica1.db "SELECT block_height, hex(merkle_root), hex(previous_hash) FROM blocks ORDER BY block_height;"
```

### Monitor Block Production Rate
```bash
watch -n 1 'sqlite3 leader.db "SELECT COUNT(*) as total_blocks FROM blocks;"'
```

---

## tmux Navigation

| Action | Command |
|--------|---------|
| Switch panes | `Ctrl+B` then arrow keys |
| Zoom pane | `Ctrl+B` then `Z` |
| Detach session | `Ctrl+B` then `D` |
| Reattach | `tmux attach -t poh-demo` or `tmux attach -t poh-bft-demo` |
| Scroll in pane | `Ctrl+B` then `[`, use arrows, press `q` to exit |
| Kill pane | `Ctrl+B` then `X` |

---

## Log Patterns to Watch

### Honest Node Success
```
✓ Block passed consensus validation
✓ Block passed verification
✓ Block linkage verified
✓ Block stored successfully
```

### Honest Node Rejection
```
✗ Block validation failed (consensus)
✗ Block verification failed
✗ Block linkage verification failed
```

### Malicious Node Activity
```
⚠ MALICIOUS: Corrupted block X
⚠ MALICIOUS: Skipping validation
⚠ MALICIOUS: Ignoring verification failure
⚠ MALICIOUS: Stored block without validation
```

---

## Performance Metrics

### Block Production
- **Rate**: ~2.5 blocks/second (400ms slots)
- **Size**: Minimum 64 ticks per block (800,000 hashes)
- **Latency**: <10ms on localhost

### Network
- **Protocol**: TCP with length-prefixed messages
- **Serialization**: JSON
- **Propagation**: Broadcast to all connected peers

### Storage
- **Database**: SQLite
- **Tables**: blocks, entries, transactions
- **Indexes**: block_height, block_hash, slot

---

## Documentation

- **README.md**: Project overview and quick start
- **docs/guides/demo.md**: Detailed demo guide with tmux commands
- **docs/testing/bft-testing.md**: In-depth BFT theory and testing guide
- **docs/testing/testing-summary.md**: This file - quick reference

---

## Quick Reference

```bash
# Start basic demo
./demo.sh 3

# Start BFT demo with BFT
./demo-bft.sh 4 1

# Start BFT demo without BFT
./demo-bft.sh 2 2

# Stop any demo
./stop-demo.sh

# Run tests
go test -v ./internal

# Build manually
go build -o bin/poh-node cmd/main.go

# Check database
sqlite3 replica1.db "SELECT * FROM blocks;"
```

---

## Troubleshooting

### "tmux: command not found"
```bash
# Ubuntu/Debian
sudo apt-get install tmux

# macOS
brew install tmux
```

### "address already in use"
```bash
./stop-demo.sh
# Or manually kill processes using ports 8000-8009
```

### Nodes not connecting
- Ensure leader starts first (scripts handle this)
- Check firewall settings
- Verify ports 8000-8009 are available

### Database locked
```bash
# Stop all nodes first
./stop-demo.sh
# Then clean up
rm -f *.db
```

---

## Next Steps

1. Run basic demo to understand normal operation
2. Run BFT demo with BFT to see fault tolerance
3. Run BFT demo without BFT to see failure modes
4. Inspect databases to verify chain state
5. Review logs to understand validation process
6. Experiment with different node configurations
7. Read BFT-TESTING.md for deeper understanding
