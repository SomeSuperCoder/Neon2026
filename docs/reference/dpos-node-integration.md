# DPoS Node Integration

## Overview

The PoH blockchain node now includes full Delegated Proof of Stake (DPoS) integration. Every node automatically initializes with a genesis validator set and epoch configuration on first startup.

## Node Startup Sequence

When a node starts, the following DPoS initialization occurs:

1. **Genesis Configuration Creation**
   - Defines epoch length (432,000 slots = ~2 days at 400ms/slot)
   - Specifies initial validator set with public keys and stake amounts

2. **ConsensusManager Initialization**
   - Creates ConsensusManager with genesis configuration
   - Wires FileStore for state persistence
   - Wires Runtime for program execution

3. **DPoS Genesis State Initialization**
   - Creates Epoch State File at well-known FileID `0x0000...0004`
   - Creates Reward Pool File at well-known FileID `0x0000...0005`
   - Creates Validator Record Files for each genesis validator
   - Idempotent: skips if already initialized

4. **Blockchain State Loading**
   - Loads existing blockchain from ledger (if present)
   - Verifies chain integrity

5. **Block Production/Reception**
   - Leader nodes produce blocks using DPoS schedule
   - Replica nodes validate and store blocks

## Default Genesis Configuration

The current implementation uses a hardcoded 2-validator setup for development:

```go
genesisConfig := consensus.GenesisConfig{
    EpochLength: 432000, // ~2 days at 400ms/slot
    GenesisValidators: []consensus.GenesisValidator{
        {
            PublicKey:   [32]byte{1, 2, 3, ...}, // Validator 1
            StakeAmount: 10000000,                // 10 Neon
        },
        {
            PublicKey:   [32]byte{32, 31, 30, ...}, // Validator 2
            StakeAmount: 5000000,                    // 5 Neon
        },
    },
}
```

### Validator Details

| Validator | Stake | Status |
|-----------|-------|--------|
| Validator 1 | 10 Neon (10,000,000 electrons) | Active |
| Validator 2 | 5 Neon (5,000,000 electrons) | Active |

### Epoch Configuration

- **Epoch Length**: 432,000 slots
- **Slot Duration**: 400ms
- **Epoch Duration**: ~2 days (172,800 seconds)

## Code Location

The DPoS integration is implemented in `cmd/main.go`:

```go
// Initialize Consensus Manager with DPoS genesis configuration
log.Println("Initializing Consensus Manager...")

// Create genesis configuration for DPoS
genesisConfig := consensus.GenesisConfig{
    EpochLength: 432000,
    GenesisValidators: []consensus.GenesisValidator{
        // ... validators ...
    },
}

consensusManager := consensus.NewConsensusManagerWithGenesis(nodeType, genesisConfig)

// Wire FileStore and Runtime to ConsensusManager for DPoS operations
consensusManager.SetFileStore(fileStore)
consensusManager.SetRuntime(rt)

// Initialize DPoS genesis state
log.Println("Initializing DPoS genesis state...")
if err := consensusManager.InitializeGenesis(genesisConfig); err != nil {
    log.Fatalf("Failed to initialize DPoS genesis: %v", err)
}
log.Println("DPoS genesis initialization complete")
```

## Production Deployment

For production deployments, the genesis configuration should be externalized:

### Recommended Approach

1. **Create Genesis Config File** (`genesis.json`):
```json
{
  "epoch_length": 432000,
  "validators": [
    {
      "public_key": "0102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f20",
      "stake_amount": 10000000
    },
    {
      "public_key": "201f1e1d1c1b1a191817161514131211100f0e0d0c0b0a090807060504030201",
      "stake_amount": 5000000
    }
  ]
}
```

2. **Load Config at Startup**:
```go
// Add flag for genesis config file
genesisConfigPath := flag.String("genesis", "genesis.json", "Path to genesis config file")

// Load and parse config
genesisConfig, err := loadGenesisConfig(*genesisConfigPath)
if err != nil {
    log.Fatalf("Failed to load genesis config: %v", err)
}
```

3. **Validate Config**:
- Ensure at least one validator
- Verify public key format (32 bytes)
- Check stake amounts are positive
- Validate epoch length is reasonable

## Monitoring DPoS State

Use the Validator TUI to monitor DPoS state in real-time:

```bash
# Build validator TUI
go build -o validator-tui ./cmd/validator-tui/main.go

# Monitor state
./validator-tui --state ./blockchain_state.db
```

The TUI displays:
- Current epoch and slot
- Active validator count
- Validator records with stakes and performance
- Reward pool balance
- Estimated APY

## Troubleshooting

### "genesis config must have at least one validator"

The genesis configuration is empty or invalid. Ensure at least one validator is specified with a valid public key and positive stake amount.

### "failed to initialize DPoS genesis"

Check the logs for specific error details. Common causes:
- FileStore initialization failed
- Insufficient permissions to create files
- Corrupted state database

### State Already Initialized

If the node has already been started once, the DPoS genesis state exists. The initialization is idempotent and will skip creation on subsequent startups.

To reset the state:
```bash
rm -f blockchain_state.db
```

## Related Documentation

- [DPoS Genesis Reference](dpos-genesis.md) - Detailed genesis initialization process
- [Validator TUI Guide](../guides/validator-tui.md) - Monitoring validator state
- [DPoS Demo Guide](../guides/dpos-demo.md) - Running DPoS demonstrations
- [DPoS Requirements](.kiro/specs/delegated-proof-of-stake/requirements.md) - Full requirements specification

## Future Enhancements

1. **External Genesis Config** - Load from JSON/TOML file instead of hardcoded values
2. **Genesis Ceremony** - Multi-party genesis generation for decentralized launch
3. **Genesis Validation** - Cryptographic verification of genesis state
4. **Hot Reload** - Update validator set without node restart
5. **Genesis Snapshots** - Export/import genesis state for network forks
