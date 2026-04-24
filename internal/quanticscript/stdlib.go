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
