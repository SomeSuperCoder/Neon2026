// System_Program - Built-in program for managing Neon accounts
// Source of truth: compile with `go run cmd/main.go qsc compile -i programs/system/system.qs -o programs/system/system.qsa`
// Then assemble: `go run cmd/main.go qsc assemble -i programs/system/system.qsa -o programs/system/system.qsb`
//
// Instruction dispatch is handled by the DISPATCH opcode (0xF0), which reads the
// SystemProgramRegistry in Go, parses all args, and pushes them onto the stack.
// The .qs source contains zero manual byte arithmetic.
//
// Requirements: 2.1, 2.2, 2.3, 2.4

// Error codes
const ERROR_INSUFFICIENT_BALANCE: i64 = 0x1000;
const ERROR_INVALID_ACCOUNT: i64 = 0x1001;
const ERROR_BALANCE_OVERFLOW: i64 = 0x1002;
const ERROR_STORAGE_RENT_VIOLATION: i64 = 0x1003;
const ERROR_UNAUTHORIZED_SIGNER: i64 = 0x1004;
const ERROR_INVALID_INSTRUCTION: i64 = 0x1FFF;

// Instruction type codes (must match SystemProgramRegistry in instructions.go)
const INSTR_CREATE_ACCOUNT: i64 = 0;
const INSTR_TRANSFER: i64 = 1;
const INSTR_ALLOCATE_SPACE: i64 = 2;

// Entry point: get instruction data, dispatch, branch to handler
export function entry(): i64 {
    // Get raw instruction data and call DISPATCH opcode via inline assembly.
    // DISPATCH reads SystemProgramRegistry, parses args, and pushes:
    //   stack (bottom to top): arg0, arg1, ..., argN, handler_name (string)
    // We pop the handler name (unused in qs — we re-read instrType for branching),
    // then pop the args into local variables.
    //
    // For CREATE_ACCOUNT: args = owner(bytes), balance(i64)
    // For TRANSFER:       args = from(bytes), to(bytes), amount(i64)
    // For ALLOCATE_SPACE: args = account(bytes), extra_balance(i64)
    //
    // Since string comparison is not supported by the compiler, we branch on the
    // raw instruction type byte, which we read via STRFROMBYTES + STRSUBSTRING.

    let instrType: i64 = 0;
    __asm__ {
        GETINSTRDATA
        STRFROMBYTES
        PUSH u64 0
        PUSH u64 1
        STRSUBSTRING
        STRTOBYTES
        BYTESTOI64LE
        STORE instrType
    }

    if (instrType == INSTR_CREATE_ACCOUNT) {
        return dispatchCreateAccount();
    } else if (instrType == INSTR_TRANSFER) {
        return dispatchTransfer();
    } else if (instrType == INSTR_ALLOCATE_SPACE) {
        return dispatchAllocateSpace();
    }

    return ERROR_INVALID_INSTRUCTION;
}

// dispatchCreateAccount: calls DISPATCH to parse CREATE_ACCOUNT args, then delegates
// Stack after DISPATCH: owner(bytes) [slot0], balance(i64) [slot1], handler_name(string) [top]
function dispatchCreateAccount(): i64 {
    let owner: i64 = 0;
    let balance: i64 = 0;
    __asm__ {
        GETINSTRDATA
        DISPATCH
        POP
        STORE balance
        STORE owner
    }
    return handleCreateAccount(owner, balance);
}

// dispatchTransfer: calls DISPATCH to parse TRANSFER args, then delegates
// Stack after DISPATCH: from(bytes) [slot0], to(bytes) [slot1], amount(i64) [slot2], handler_name(string) [top]
function dispatchTransfer(): i64 {
    let from: i64 = 0;
    let to: i64 = 0;
    let amount: i64 = 0;
    __asm__ {
        GETINSTRDATA
        DISPATCH
        POP
        STORE amount
        STORE to
        STORE from
    }
    return handleTransfer(from, to, amount);
}

// dispatchAllocateSpace: calls DISPATCH to parse ALLOCATE_SPACE args, then delegates
// Stack after DISPATCH: account(bytes) [slot0], extra_balance(i64) [slot1], handler_name(string) [top]
function dispatchAllocateSpace(): i64 {
    let account: i64 = 0;
    let extraBalance: i64 = 0;
    __asm__ {
        GETINSTRDATA
        DISPATCH
        POP
        STORE extraBalance
        STORE account
    }
    return handleAllocateSpace(account, extraBalance);
}

// handleCreateAccount: validate signer and create account file
// Requirements: 2.1, 2.4
function handleCreateAccount(owner: i64, balance: i64): i64 {
    if (balance < 0) {
        return ERROR_INSUFFICIENT_BALANCE;
    }
    if (!hasSigner(owner)) {
        return ERROR_UNAUTHORIZED_SIGNER;
    }
    // Account creation is handled by the runtime via the transaction processor.
    // The System_Program validates authorization; the runtime creates the file.
    return 0;
}

// handleTransfer: validate signer, check balances, update both accounts
// Requirements: 2.1, 2.2, 2.4
function handleTransfer(from: i64, to: i64, amount: i64): i64 {
    if (amount < 0) {
        return ERROR_INSUFFICIENT_BALANCE;
    }
    if (!hasSigner(from)) {
        return ERROR_UNAUTHORIZED_SIGNER;
    }
    let sourceBalance: i64 = getBalance(from);
    if (sourceBalance < amount) {
        return ERROR_INSUFFICIENT_BALANCE;
    }
    let destBalance: i64 = getBalance(to);
    let newDest: i64 = destBalance + amount;
    if (newDest < destBalance) {
        return ERROR_BALANCE_OVERFLOW;
    }
    let newSource: i64 = sourceBalance - amount;
    if (newSource < 0) {
        return ERROR_INSUFFICIENT_BALANCE;
    }
    updateBalance(from, 0 - amount);
    updateBalance(to, amount);
    return 0;
}

// handleAllocateSpace: validate signer and update balance for storage rent
// Requirements: 2.1, 2.3, 2.4
function handleAllocateSpace(account: i64, extraBalance: i64): i64 {
    if (extraBalance < 0) {
        return ERROR_INSUFFICIENT_BALANCE;
    }
    if (!hasSigner(account)) {
        return ERROR_UNAUTHORIZED_SIGNER;
    }
    let currentBalance: i64 = getBalance(account);
    let newBalance: i64 = currentBalance + extraBalance;
    if (newBalance < currentBalance) {
        return ERROR_BALANCE_OVERFLOW;
    }
    updateBalance(account, extraBalance);
    return 0;
}
