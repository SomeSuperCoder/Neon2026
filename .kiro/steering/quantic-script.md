---
inclusion: always
---

# QuanticScript Language

QuanticScript is a TypeScript-like smart contract language that compiles to a custom stack-based bytecode and runs on the PoH blockchain. The language is actively evolving — if a feature is missing, extend the language rather than working around it.

## Core Principle

> The language is not complete. When implementing a smart contract, extend the lexer, parser, type checker, code generator, assembler, or interpreter as needed. The only requirement is a working smart contract.

## Pipeline

Source (`.qs`) → Lexer → Parser → AST → TypeChecker → CodeGen → Bytecode (`.qsb`)
Assembly (`.qsa`) → Assembler → Bytecode (`.qsb`)
Bytecode (`.qsb`) → Disassembler → Assembly (`.qsa`)

All pipeline stages live in `internal/quanticscript/`.

## Execution Context

`ExecutionContext` (in `context.go`) provides the program with access to the FileStore, signers, instruction data, and program ID. Blockchain operations (`OpGetFile`, `OpUpdateBalance`, etc.) go through this context — never bypass it.

## Cross-Program Invocation

Max invocation depth is enforced at runtime. Use `OpInvoke` / `OpInvokeRet`. Track depth via `invokeDepth` on the interpreter.

## Extending the Language

When a language feature is needed:

1. Add token(s) to `token.go` and `lexer.go`
2. Add AST node(s) to `ast.go`
3. Update `parser.go` to produce the new node
4. Update `typechecker.go` for type rules
5. Update `codegen.go` to emit bytecode
6. Add opcode(s) if needed (see above)
7. Handle the opcode in `interpreter.go`
8. Write tests first, then implement

## Testing

Write tests before implementation. Test files mirror source files (`foo.go` → `foo_test.go`). Use table-driven tests. Integration tests go in `internal/quanticscript/*_integration_test.go`. Use `checkParserErrors(t, parser)` for parser tests.

```bash
go test ./internal/quanticscript/...
go test -v ./internal/quanticscript/... -run TestSomething
```

## Smart Contract Conventions

- Programs are compiled to `.qsb` and stored under `programs/<name>/`
- Keep `.qs` source, `.qsa` assembly, and `.qsb` bytecode in sync
- Use `OpDispatch` + a registry for instruction routing in programs
- Instruction data is little-endian encoded; use `ParseInstructionU64` / `ParseInstructionU8` helpers from `stdlib_programs.go`
- FileID and PublicKey must be exactly 32 bytes; use `NewFileIDFromBytes` / `NewPublicKeyFromBytes`

# Compile always
Never every Write bytecode or assembly directly!!! Ever! Unless 1000000% needed
