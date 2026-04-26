# Strings and Logging in QuanticScript

## Overview

QuanticScript now fully supports string literals and the `log()` function for debugging. This enables developers to write meaningful debug output and work with text data.

## String Literals

String literals are created using double quotes:

```quanticscript
let greeting: string = "Hello, World!";
let name: string = "Alice";
```

## String Operations

The following string operations are available:

### stringConcat(str1: string, str2: string) -> string
Concatenates two strings:
```quanticscript
let result: string = stringConcat("Hello", " World");
// result = "Hello World"
```

### stringLength(str: string) -> i64
Returns the length of a string:
```quanticscript
let len: i64 = stringLength("Hello");
// len = 5
```

### stringSubstring(str: string, start: i64, end: i64) -> string
Extracts a substring:
```quanticscript
let sub: string = stringSubstring("Hello", 1, 4);
// sub = "ell"
```

### stringToBytes(str: string) -> bytes
Converts a string to bytes:
```quanticscript
let data: bytes = stringToBytes("Hello");
```

### bytesToString(data: bytes) -> string
Converts bytes to a string:
```quanticscript
let str: string = bytesToString(data);
```

## Logging for Debugging

The `log()` function is a high-cost operation (5000 compute units) designed for debugging:

```quanticscript
log(value);  // Logs any value type
```

### Supported Types
- `i64` - Signed 64-bit integers
- `u64` - Unsigned 64-bit integers
- `string` - Text strings
- `bytes` - Byte arrays
- `bool` - Boolean values

### Example
```quanticscript
export function entry(): i64 {
    let x: i64 = 42;
    let msg: string = "The answer is";
    
    log(msg);
    log(x);
    
    return 0;
}
```

## Examples

### Hello, World!
See `14_hello_world.qs` for a simple "Hello, World!" program.

### String Operations
See `15_string_operations.qs` for comprehensive string manipulation examples.

### Debug Logging
See `13_log_debug.qs` for logging examples.

## Performance Considerations

- **log() cost**: 5000 compute units per call
- **String operations**: Relatively low cost (2-3 units each)
- **String concatenation**: O(n) where n is the total length of both strings

## Type Safety

All string operations are type-checked at compile time. The type checker ensures:
- String literals are properly typed as `string`
- String operations receive correct argument types
- The `log()` function accepts any type (via the `any` type)

## Implementation Details

### Bytecode Format
String literals are encoded in bytecode as:
1. OpPush opcode
2. TypeString type byte
3. 8-byte little-endian length
4. String bytes

### Cost Metering
All string operations are metered for compute cost:
- STRLEN: 2 units
- STRCONCAT: 3 units
- STRSUBSTRING: 3 units
- STRTOBYTES: 2 units
- STRFROMBYTES: 2 units
- LOG: 5000 units (high cost for debugging)
