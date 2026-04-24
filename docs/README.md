# PoH Blockchain Documentation

Complete documentation for the Proof of History blockchain implementation.

## Getting Started

- **[Quick Start Guide](guides/quickstart.md)** - Get up and running in 30 seconds
- **[Demo Guide](guides/demo.md)** - Interactive demos with tmux
- **[CLI Usage Guide](guides/cli-usage.md)** - Command-line interface reference
- **[QuanticScript Guide](guides/quanticscript.md)** - Smart contract language overview

## Testing

- **[BFT Testing Guide](testing/bft-testing.md)** - Byzantine Fault Tolerance testing
- **[Automated Testing Guide](testing/automated-testing.md)** - Testing without tmux
- **[Testing Summary](testing/testing-summary.md)** - Quick reference for all testing features

## Reference

### QuanticScript Language
- **[Language Reference](reference/language-reference.md)** - Complete syntax and semantics
- **[Standard Library Reference](reference/stdlib-reference.md)** - Built-in functions and modules
- **[Inline Assembly Guide](reference/inline-assembly.md)** - Low-level assembly programming
- **[Bytecode Reference](reference/bytecode-reference.md)** - Bytecode format and opcodes
- **[Cost Model Guide](reference/cost-model.md)** - Understanding computational costs

### Implementation
- **[Implementation Summary](reference/implementation-summary.md)** - Architecture and features
- **[BFT Fix Summary](reference/bft-fix-summary.md)** - Technical details of BFT fixes

## Documentation Structure

```
docs/
├── README.md                          # This file
├── guides/                            # User guides
│   ├── quickstart.md                  # 30-second quick start
│   ├── demo.md                        # Interactive demo guide
│   ├── cli-usage.md                   # CLI reference
│   └── quanticscript.md               # QuanticScript overview
├── testing/                           # Testing documentation
│   ├── bft-testing.md                 # BFT testing guide
│   ├── automated-testing.md           # Automated testing
│   └── testing-summary.md             # Testing quick reference
└── reference/                         # Technical reference
    ├── language-reference.md          # QuanticScript language spec
    ├── stdlib-reference.md            # Standard library API
    ├── inline-assembly.md             # Assembly programming
    ├── bytecode-reference.md          # Bytecode specification
    ├── cost-model.md                  # Cost model details
    ├── implementation-summary.md      # Implementation overview
    └── bft-fix-summary.md             # BFT fixes technical details
```

## Quick Links

### For New Users
1. Start with the [Quick Start Guide](guides/quickstart.md)
2. Run the [Demo](guides/demo.md)
3. Try [BFT Testing](testing/bft-testing.md)

### For Developers
1. Read the [Implementation Summary](reference/implementation-summary.md)
2. Explore the [QuanticScript Language](reference/language-reference.md)
3. Check out the [Standard Library](reference/stdlib-reference.md)

### For Testers
1. Review the [Testing Summary](testing/testing-summary.md)
2. Use [Automated Testing](testing/automated-testing.md)
3. Understand [BFT Testing](testing/bft-testing.md)

## Examples

See the [examples/](../examples/) directory for:
- Simple transfers
- Token contracts
- Advanced smart contracts
- Assembly examples
- Control flow demonstrations

## Contributing

See the main [README.md](../README.md) for contribution guidelines.

## License

See the main project LICENSE file.
