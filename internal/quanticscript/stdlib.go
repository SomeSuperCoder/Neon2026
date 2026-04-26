package quanticscript

import (
	"fmt"
)

// String operations

// execStrConcat concatenates two strings
func (bi *BytecodeInterpreter) execStrConcat() error {
	// Pop second string from stack
	str2Value, err := bi.pop()
	if err != nil {
		return err
	}

	if str2Value.Type != TypeString {
		return fmt.Errorf("STRCONCAT requires string, got %v", str2Value.Type)
	}

	// Pop first string from stack
	str1Value, err := bi.pop()
	if err != nil {
		return err
	}

	if str1Value.Type != TypeString {
		return fmt.Errorf("STRCONCAT requires string, got %v", str1Value.Type)
	}

	str1, _ := str1Value.AsString()
	str2, _ := str2Value.AsString()

	// Concatenate strings
	result := str1 + str2

	return bi.push(NewString(result))
}

// execStrSubstring extracts a substring
func (bi *BytecodeInterpreter) execStrSubstring() error {
	// Pop end index from stack
	endValue, err := bi.pop()
	if err != nil {
		return err
	}

	if endValue.Type != TypeU64 {
		return fmt.Errorf("STRSUBSTRING requires u64 for end index, got %v", endValue.Type)
	}

	// Pop start index from stack
	startValue, err := bi.pop()
	if err != nil {
		return err
	}

	if startValue.Type != TypeU64 {
		return fmt.Errorf("STRSUBSTRING requires u64 for start index, got %v", startValue.Type)
	}

	// Pop string from stack
	strValue, err := bi.pop()
	if err != nil {
		return err
	}

	if strValue.Type != TypeString {
		return fmt.Errorf("STRSUBSTRING requires string, got %v", strValue.Type)
	}

	str, _ := strValue.AsString()
	start, _ := startValue.AsU64()
	end, _ := endValue.AsU64()

	// Validate indices
	if start > uint64(len(str)) {
		return fmt.Errorf("start index out of bounds: %d", start)
	}
	if end > uint64(len(str)) {
		return fmt.Errorf("end index out of bounds: %d", end)
	}
	if start > end {
		return fmt.Errorf("start index greater than end index")
	}

	// Extract substring
	result := str[start:end]

	return bi.push(NewString(result))
}

// execStrLen gets the length of a string
func (bi *BytecodeInterpreter) execStrLen() error {
	// Pop string from stack
	strValue, err := bi.pop()
	if err != nil {
		return err
	}

	if strValue.Type != TypeString {
		return fmt.Errorf("STRLEN requires string, got %v", strValue.Type)
	}

	str, _ := strValue.AsString()
	length := uint64(len(str))

	return bi.push(NewU64(length))
}

// execStrToBytes converts a string to bytes
func (bi *BytecodeInterpreter) execStrToBytes() error {
	// Pop string from stack
	strValue, err := bi.pop()
	if err != nil {
		return err
	}

	if strValue.Type != TypeString {
		return fmt.Errorf("STRTOBYTES requires string, got %v", strValue.Type)
	}

	str, _ := strValue.AsString()
	bytes := []byte(str)

	return bi.push(NewBytes(bytes))
}

// execStrFromBytes converts bytes to a string
func (bi *BytecodeInterpreter) execStrFromBytes() error {
	// Pop bytes from stack
	bytesValue, err := bi.pop()
	if err != nil {
		return err
	}

	if bytesValue.Type != TypeBytes {
		return fmt.Errorf("STRFROMBYTES requires bytes, got %v", bytesValue.Type)
	}

	bytes, _ := bytesValue.AsBytes()
	str := string(bytes)

	return bi.push(NewString(str))
}

// Math operations

// execMathMin returns the minimum of two values
func (bi *BytecodeInterpreter) execMathMin() error {
	// Pop second value from stack
	b, err := bi.pop()
	if err != nil {
		return err
	}

	// Pop first value from stack
	a, err := bi.pop()
	if err != nil {
		return err
	}

	if a.Type != b.Type {
		return fmt.Errorf("MATHMIN requires same types, got %v and %v", a.Type, b.Type)
	}

	var result Value
	switch a.Type {
	case TypeI64:
		aVal, _ := a.AsI64()
		bVal, _ := b.AsI64()
		if aVal < bVal {
			result = a
		} else {
			result = b
		}
	case TypeU64:
		aVal, _ := a.AsU64()
		bVal, _ := b.AsU64()
		if aVal < bVal {
			result = a
		} else {
			result = b
		}
	default:
		return fmt.Errorf("MATHMIN unsupported type: %v", a.Type)
	}

	return bi.push(result)
}

// execMathMax returns the maximum of two values
func (bi *BytecodeInterpreter) execMathMax() error {
	// Pop second value from stack
	b, err := bi.pop()
	if err != nil {
		return err
	}

	// Pop first value from stack
	a, err := bi.pop()
	if err != nil {
		return err
	}

	if a.Type != b.Type {
		return fmt.Errorf("MATHMAX requires same types, got %v and %v", a.Type, b.Type)
	}

	var result Value
	switch a.Type {
	case TypeI64:
		aVal, _ := a.AsI64()
		bVal, _ := b.AsI64()
		if aVal > bVal {
			result = a
		} else {
			result = b
		}
	case TypeU64:
		aVal, _ := a.AsU64()
		bVal, _ := b.AsU64()
		if aVal > bVal {
			result = a
		} else {
			result = b
		}
	default:
		return fmt.Errorf("MATHMAX unsupported type: %v", a.Type)
	}

	return bi.push(result)
}

// execMathAbs returns the absolute value
func (bi *BytecodeInterpreter) execMathAbs() error {
	// Pop value from stack
	a, err := bi.pop()
	if err != nil {
		return err
	}

	var result Value
	switch a.Type {
	case TypeI64:
		aVal, _ := a.AsI64()
		if aVal < 0 {
			result = NewI64(-aVal)
		} else {
			result = a
		}
	case TypeU64:
		// Unsigned values are always positive
		result = a
	default:
		return fmt.Errorf("MATHABS unsupported type: %v", a.Type)
	}

	return bi.push(result)
}

// execMathPow computes power (deterministic integer only)
func (bi *BytecodeInterpreter) execMathPow() error {
	// Pop exponent from stack
	expValue, err := bi.pop()
	if err != nil {
		return err
	}

	// Pop base from stack
	baseValue, err := bi.pop()
	if err != nil {
		return err
	}

	if baseValue.Type != expValue.Type {
		return fmt.Errorf("MATHPOW requires same types, got %v and %v", baseValue.Type, expValue.Type)
	}

	var result Value
	switch baseValue.Type {
	case TypeI64:
		base, _ := baseValue.AsI64()
		exp, _ := expValue.AsI64()
		if exp < 0 {
			return fmt.Errorf("MATHPOW does not support negative exponents")
		}
		result = NewI64(intPow(base, exp))
	case TypeU64:
		base, _ := baseValue.AsU64()
		exp, _ := expValue.AsU64()
		result = NewU64(uintPow(base, exp))
	default:
		return fmt.Errorf("MATHPOW unsupported type: %v", baseValue.Type)
	}

	return bi.push(result)
}

// intPow computes integer power deterministically
func intPow(base, exp int64) int64 {
	if exp == 0 {
		return 1
	}
	result := int64(1)
	for i := int64(0); i < exp; i++ {
		result *= base
	}
	return result
}

// uintPow computes unsigned integer power deterministically
func uintPow(base, exp uint64) uint64 {
	if exp == 0 {
		return 1
	}
	result := uint64(1)
	for i := uint64(0); i < exp; i++ {
		result *= base
	}
	return result
}

// Cross-program invocation operations

// Invoke is a high-level wrapper for the INVOKE bytecode instruction
// It simplifies cross-program invocation by handling data marshaling
// This implements Requirements 9.1, 9.2, 9.3, 9.5
//
// Usage in QuanticScript:
//   import { invoke } from "std/invoke";
//   let result = invoke(targetProgramID, invokeData, computeBudget);
//
// The function takes:
//   - programID: FileID (32 bytes) - The target program to invoke
//   - invokeData: bytes - The data to pass to the invoked program
//   - computeBudget: i64 - The compute budget to allocate to the invoked program
//
// Returns:
//   - result: bytes - The result data from the invoked program
//
// Errors:
//   - If the target program is not in the declared program list (Requirement 9.2)
//   - If the invocation depth exceeds the maximum (Requirement 9.4)
//   - If the compute budget is insufficient (Requirement 9.5)
//   - If the invoked program execution fails (Requirement 9.5)
//
// Note: This function is typically generated by the compiler when the high-level
// invoke() function is called in QuanticScript code. The compiler will emit:
//   PUSH programID
//   PUSH invokeData
//   PUSH computeBudget
//   INVOKE

// GetInvokeDepth returns the current cross-program invocation depth
// This is useful for programs that need to know how deep they are in the call stack
//
// Usage in QuanticScript:
//   import { getInvokeDepth } from "std/invoke";
//   let depth = getInvokeDepth();
//
// The function returns:
//   - depth: u32 - The current invocation depth (0 for top-level programs)
//
// Note: This is implemented by reading the interpreter's invokeDepth field
// The compiler will need to provide a way to access this, possibly through
// a special builtin function or by storing it in a known memory location

// execBytesToI64LE decodes TypeBytes (8 bytes) as a little-endian int64.
// This enables bytecode programs to decode LE-encoded amounts from instruction data.
func (bi *BytecodeInterpreter) execBytesToI64LE() error {
	bytesValue, err := bi.pop()
	if err != nil {
		return err
	}

	if bytesValue.Type != TypeBytes {
		return fmt.Errorf("BYTESTOI64LE requires bytes, got %v", bytesValue.Type)
	}

	data, _ := bytesValue.AsBytes()
	if len(data) < 8 {
		return fmt.Errorf("BYTESTOI64LE requires at least 8 bytes, got %d", len(data))
	}

	val := int64(data[0]) |
		int64(data[1])<<8 |
		int64(data[2])<<16 |
		int64(data[3])<<24 |
		int64(data[4])<<32 |
		int64(data[5])<<40 |
		int64(data[6])<<48 |
		int64(data[7])<<56

	return bi.push(NewI64(val))
}

// SetRegistry sets the instruction dispatch registry for this interpreter.
// Programs that use the DISPATCH opcode must have a registry set.
func (bi *BytecodeInterpreter) SetRegistry(registry map[int]InstructionDef) {
	bi.registry = registry
}

// execDispatch implements the DISPATCH opcode.
// Stack in:  [instrData (bytes)]
// Stack out: [handler (string), arg0, arg1, ..., argN]  (args in schema order, handler on top)
// On error:  returns a Go error (no state modification).
func (bi *BytecodeInterpreter) execDispatch() error {
	instrDataValue, err := bi.pop()
	if err != nil {
		return err
	}
	if instrDataValue.Type != TypeBytes {
		return fmt.Errorf("DISPATCH requires bytes on stack, got %v", instrDataValue.Type)
	}
	data, _ := instrDataValue.AsBytes()

	if bi.registry == nil {
		return fmt.Errorf("DISPATCH: no instruction registry set for this program")
	}

	def, args, err := Dispatch(data, bi.registry)
	if err != nil {
		return fmt.Errorf("DISPATCH: %w", err)
	}

	// Deduct per-arg cost (2 per arg) in addition to the base cost already charged
	perArgCost := InstructionCost(2 * len(def.Args))
	if err := bi.deductCost(perArgCost); err != nil {
		return err
	}

	// Push args in schema order (first arg deepest, last arg on top before handler)
	for _, argDef := range def.Args {
		if err := bi.push(args[argDef.Name]); err != nil {
			return err
		}
	}

	// Push handler name on top so assembly can branch on it
	return bi.push(NewString(def.Handler))
}

// Helper functions for System Program and smart contracts

// execSlice extracts a byte slice from start (inclusive) to end (exclusive)
// Stack: [data (bytes), start (i64), end (i64)] -> [result (bytes)]
func (bi *BytecodeInterpreter) execSlice() error {
	// Pop end index from stack
	endValue, err := bi.pop()
	if err != nil {
		return err
	}

	if endValue.Type != TypeI64 {
		return fmt.Errorf("SLICE requires i64 for end index, got %v", endValue.Type)
	}

	// Pop start index from stack
	startValue, err := bi.pop()
	if err != nil {
		return err
	}

	if startValue.Type != TypeI64 {
		return fmt.Errorf("SLICE requires i64 for start index, got %v", startValue.Type)
	}

	// Pop data from stack
	dataValue, err := bi.pop()
	if err != nil {
		return err
	}

	if dataValue.Type != TypeBytes {
		return fmt.Errorf("SLICE requires bytes for data, got %v", dataValue.Type)
	}

	data, _ := dataValue.AsBytes()
	start, _ := startValue.AsI64()
	end, _ := endValue.AsI64()

	// Validate indices
	if start < 0 || end < 0 {
		return fmt.Errorf("SLICE: indices must be non-negative, got start=%d end=%d", start, end)
	}
	if start > end {
		return fmt.Errorf("SLICE: start index %d greater than end index %d", start, end)
	}
	if end > int64(len(data)) {
		return fmt.Errorf("SLICE: end index %d out of bounds for length %d", end, len(data))
	}

	// Extract slice
	result := data[start:end]

	return bi.push(NewBytes(result))
}

// execBytesToFileID converts a 32-byte slice to FileID
// Stack: [data (bytes)] -> [fileID (FileID)]
func (bi *BytecodeInterpreter) execBytesToFileID() error {
	// Pop data from stack
	dataValue, err := bi.pop()
	if err != nil {
		return err
	}

	if dataValue.Type != TypeBytes {
		return fmt.Errorf("BYTESTOFILEID requires bytes, got %v", dataValue.Type)
	}

	data, _ := dataValue.AsBytes()

	if len(data) != 32 {
		return fmt.Errorf("BYTESTOFILEID requires exactly 32 bytes, got %d", len(data))
	}

	// Create FileID value
	fileIDCopy := make([]byte, 32)
	copy(fileIDCopy, data)

	return bi.push(NewFileID(fileIDCopy))
}

// execLog logs a value for debugging purposes
// Stack: [value] -> []
// This is a high-cost operation (5000 units) to discourage use in production
// The value is popped from the stack and logged (in a real implementation,
// this would write to a debug log or emit an event)
func (bi *BytecodeInterpreter) execLog() error {
	// Pop value from stack
	value, err := bi.pop()
	if err != nil {
		return err
	}

	// Format the value as a string for logging
	var logMsg string
	switch value.Type {
	case TypeI64:
		v, _ := value.AsI64()
		logMsg = fmt.Sprintf("LOG: i64(%d)", v)
	case TypeU64:
		v, _ := value.AsU64()
		logMsg = fmt.Sprintf("LOG: u64(%d)", v)
	case TypeString:
		v, _ := value.AsString()
		logMsg = fmt.Sprintf("LOG: string(%q)", v)
	case TypeBytes:
		v, _ := value.AsBytes()
		logMsg = fmt.Sprintf("LOG: bytes(len=%d)", len(v))
	case TypeBool:
		v, _ := value.AsBool()
		logMsg = fmt.Sprintf("LOG: bool(%v)", v)
	default:
		logMsg = fmt.Sprintf("LOG: %v", value.Type)
	}

	// In a real implementation, this would write to a debug log
	// For now, we just consume the value (it's already popped)
	_ = logMsg // Suppress unused variable warning

	return nil
}
