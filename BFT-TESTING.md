# Byzantine Fault Tolerance (BFT) Testing Guide

## Overview

This PoH blockchain implementation includes malicious node capabilities for testing Byzantine Fault Tolerance. This allows you to simulate real-world attack scenarios and verify that honest nodes maintain consensus despite Byzantine faults.

## BFT Theory

### The Byzantine Generals Problem

In distributed systems, some nodes may behave maliciously (Byzantine faults):
- Send conflicting information to different nodes
- Provide incorrect data
- Fail to respond or respond incorrectly
- Attempt to disrupt consensus

### BFT Requirement

For a system to tolerate Byzantine faults:

```
Honest Nodes > 2 × Malicious Nodes
```

Or equivalently:
```
Honest Nodes > (2/3) × Total Nodes
```

**Examples**:
- 4 honest + 1 malicious = **BFT** ✓ (4 > 2×1)
- 3 honest + 1 malicious = **BFT** ✓ (3 > 2×1)
- 2 honest + 1 malicious = **NO BFT** ✗ (2 = 2×1)
- 2 honest + 2 malicious = **NO BFT** ✗ (2 ≤ 2×2)

## Running BFT Tests

### Quick Start

```bash
# Test with BFT (should maintain integrity)
./demo-bft.sh 4 1

# Test without BFT (may accept invalid blocks)
./demo-bft.sh 2 2
```

### Manual Testing

You can also manually start nodes with the `--malicious` flag:

```bash
# Start honest leader
./bin/poh-node --type=leader --port=8000 --db=leader.db

# Start honest replica
./bin/poh-node --type=replica --port=8001 --peers=localhost:8000 --db=replica1.db

# Start malicious replica
./bin/poh-node --type=replica --port=8002 --peers=localhost:8000 --db=malicious1.db --malicious
```

## Malicious Behaviors Implemented

### Malicious Leader

A malicious leader exhibits the following behaviors:

1. **Invalid Hash Counts** (every 3rd block)
   - Sets `NumHashes` to 1 instead of actual count
   - Makes PoH verification fail
   - Tests if replicas properly validate hash chains

2. **Wrong Previous Hash** (every 5th block)
   - Corrupts `PreviousBlockHash` field
   - Breaks chain linkage
   - Tests if replicas verify block continuity

3. **Selective Storage**
   - Doesn't store corrupted blocks locally
   - Creates inconsistent state
   - Tests replica independence

### Malicious Replica

A malicious replica exhibits the following behaviors:

1. **Skip Validation** (every 4th block)
   - Stores blocks without any validation
   - Accepts potentially invalid data
   - Tests if other replicas maintain standards

2. **Ignore Consensus Failures** (every 6th block)
   - Stores blocks that fail consensus validation
   - Tests consensus enforcement

3. **Ignore Verification Failures** (every 6th block)
   - Stores blocks that fail PoH verification
   - Tests cryptographic verification enforcement

4. **Ignore Linkage Failures** (every 6th block)
   - Stores blocks with broken chain linkage
   - Tests chain integrity enforcement

## Test Scenarios

### Scenario 1: Strong BFT (6 honest, 2 malicious)

```bash
./demo-bft.sh 6 2
```

**Expected Behavior**:
- All honest nodes reject invalid blocks
- Malicious nodes have corrupted local state
- Network maintains valid consensus
- Chain integrity preserved across honest nodes

**What to Watch**:
- Honest replicas logging "Block verification failed"
- Malicious replicas logging "MALICIOUS: Stored block without validation"
- Honest nodes maintaining identical chain heights

### Scenario 2: Minimal BFT (3 honest, 1 malicious)

```bash
./demo-bft.sh 3 1
```

**Expected Behavior**:
- Honest majority maintains consensus
- Single malicious node isolated
- Network operates normally
- Demonstrates minimum BFT threshold

**What to Watch**:
- 3 honest nodes with identical state
- 1 malicious node with divergent state
- Corrupted blocks rejected by honest nodes

### Scenario 3: No BFT (2 honest, 2 malicious)

```bash
./demo-bft.sh 2 2
```

**Expected Behavior**:
- Network lacks BFT threshold
- Invalid blocks may be accepted
- Chain integrity compromised
- Demonstrates importance of BFT

**What to Watch**:
- Inconsistent chain heights across nodes
- Some nodes accepting invalid blocks
- Network unable to reach consensus

### Scenario 4: Extreme Byzantine Fault (2 honest, 3 malicious)

```bash
./demo-bft.sh 2 3
```

**Expected Behavior**:
- Malicious nodes outnumber honest nodes
- Network completely compromised
- Demonstrates catastrophic failure without BFT

**What to Watch**:
- Honest nodes overwhelmed by invalid data
- Majority of network in corrupted state
- Complete loss of consensus

## Verification Methods

### Check Chain Heights

After running for 30 seconds, check if honest nodes have consistent heights:

```bash
# Check honest nodes
sqlite3 replica1.db "SELECT COUNT(*) FROM blocks;"
sqlite3 replica2.db "SELECT COUNT(*) FROM blocks;"

# Check malicious nodes (may differ)
sqlite3 malicious1.db "SELECT COUNT(*) FROM blocks;"
```

### Verify Chain Integrity

Run verification on honest nodes:

```bash
# Should pass on honest nodes
sqlite3 replica1.db "SELECT block_height, slot FROM blocks ORDER BY block_height;"

# May show gaps or inconsistencies on malicious nodes
sqlite3 malicious1.db "SELECT block_height, slot FROM blocks ORDER BY block_height;"
```

### Analyze Logs

Look for these patterns in the logs:

**Honest Node Success**:
```
Replica node: Block passed consensus validation
Replica node: Block passed verification
Replica node: Block linkage verified
Replica node: Block stored successfully
```

**Honest Node Rejection**:
```
Replica node: Block verification failed: entry hash does not match
Replica node: Block linkage verification failed
```

**Malicious Node Behavior**:
```
MALICIOUS: Skipping validation for block X
MALICIOUS: Stored block X without validation
MALICIOUS: Ignoring verification failure, storing anyway
```

## Performance Impact

Malicious nodes have minimal performance impact on honest nodes:
- Honest nodes reject invalid blocks quickly
- Network bandwidth slightly increased (broadcasting invalid blocks)
- Honest nodes maintain normal block production rate
- No degradation in consensus speed

## Real-World Implications

This testing demonstrates:

1. **Importance of Validation**: Every node must independently validate
2. **BFT Threshold**: Need >2/3 honest nodes for security
3. **Attack Resistance**: Honest nodes resist various attack vectors
4. **Consensus Preservation**: Valid chain maintained despite faults

## Limitations

Current implementation limitations:

1. **No Leader Election**: Single leader (not Byzantine fault tolerant for leader)
2. **No Slashing**: Malicious nodes not penalized
3. **No Reputation**: All nodes treated equally
4. **Simplified Attacks**: Real attacks may be more sophisticated

## Future Enhancements

Potential improvements for BFT testing:

1. **Multiple Leaders**: Rotate leadership to test leader BFT
2. **Network Partitions**: Simulate network splits
3. **Timing Attacks**: Delay blocks strategically
4. **Sybil Attacks**: Simulate multiple identities
5. **Eclipse Attacks**: Isolate honest nodes
6. **Consensus Voting**: Implement voting mechanism for block acceptance

## Conclusion

The BFT testing capability allows you to:
- Verify system resilience against Byzantine faults
- Understand BFT requirements practically
- Test validation and verification logic
- Demonstrate consensus preservation
- Educate about distributed systems security

Use these tools to build confidence in the blockchain's fault tolerance and understand the critical importance of the BFT threshold in distributed systems.
