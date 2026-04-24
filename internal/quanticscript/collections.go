package quanticscript

import (
	"fmt"
	"sort"
)

// Collection operations implementation

// execArrayNew creates a new empty array
func (bi *BytecodeInterpreter) execArrayNew() error {
	// Create empty array
	arr := make([]Value, 0)
	return bi.push(NewArray(arr))
}

// execArrayLen gets the length of an array
func (bi *BytecodeInterpreter) execArrayLen() error {
	// Pop array from stack
	arrValue, err := bi.pop()
	if err != nil {
		return err
	}

	if arrValue.Type != TypeArray {
		return fmt.Errorf("ARRAYLEN requires array, got %v", arrValue.Type)
	}

	arr, _ := arrValue.AsArray()
	length := uint64(len(arr))

	return bi.push(NewU64(length))
}

// execArrayGet gets an element from an array by index
func (bi *BytecodeInterpreter) execArrayGet() error {
	// Pop index from stack
	indexValue, err := bi.pop()
	if err != nil {
		return err
	}

	if indexValue.Type != TypeU64 {
		return fmt.Errorf("ARRAYGET requires u64 for index, got %v", indexValue.Type)
	}

	// Pop array from stack
	arrValue, err := bi.pop()
	if err != nil {
		return err
	}

	if arrValue.Type != TypeArray {
		return fmt.Errorf("ARRAYGET requires array, got %v", arrValue.Type)
	}

	arr, _ := arrValue.AsArray()
	index, _ := indexValue.AsU64()

	// Check bounds
	if int(index) >= len(arr) {
		return fmt.Errorf("array index out of bounds: %d", index)
	}

	return bi.push(arr[index])
}

// execArraySet sets an element in an array by index
func (bi *BytecodeInterpreter) execArraySet() error {
	// Pop value from stack
	value, err := bi.pop()
	if err != nil {
		return err
	}

	// Pop index from stack
	indexValue, err := bi.pop()
	if err != nil {
		return err
	}

	if indexValue.Type != TypeU64 {
		return fmt.Errorf("ARRAYSET requires u64 for index, got %v", indexValue.Type)
	}

	// Pop array from stack
	arrValue, err := bi.pop()
	if err != nil {
		return err
	}

	if arrValue.Type != TypeArray {
		return fmt.Errorf("ARRAYSET requires array, got %v", arrValue.Type)
	}

	arr, _ := arrValue.AsArray()
	index, _ := indexValue.AsU64()

	// Check bounds
	if int(index) >= len(arr) {
		return fmt.Errorf("array index out of bounds: %d", index)
	}

	// Set value (create new array to maintain immutability)
	newArr := make([]Value, len(arr))
	copy(newArr, arr)
	newArr[index] = value

	return bi.push(NewArray(newArr))
}

// execArrayPush pushes an element to the end of an array
func (bi *BytecodeInterpreter) execArrayPush() error {
	// Pop value from stack
	value, err := bi.pop()
	if err != nil {
		return err
	}

	// Pop array from stack
	arrValue, err := bi.pop()
	if err != nil {
		return err
	}

	if arrValue.Type != TypeArray {
		return fmt.Errorf("ARRAYPUSH requires array, got %v", arrValue.Type)
	}

	arr, _ := arrValue.AsArray()

	// Create new array with appended value
	newArr := make([]Value, len(arr)+1)
	copy(newArr, arr)
	newArr[len(arr)] = value

	return bi.push(NewArray(newArr))
}

// execArrayPop pops an element from the end of an array
func (bi *BytecodeInterpreter) execArrayPop() error {
	// Pop array from stack
	arrValue, err := bi.pop()
	if err != nil {
		return err
	}

	if arrValue.Type != TypeArray {
		return fmt.Errorf("ARRAYPOP requires array, got %v", arrValue.Type)
	}

	arr, _ := arrValue.AsArray()

	// Check if array is empty
	if len(arr) == 0 {
		return fmt.Errorf("cannot pop from empty array")
	}

	// Get last element
	lastElement := arr[len(arr)-1]

	// Create new array without last element
	newArr := make([]Value, len(arr)-1)
	copy(newArr, arr[:len(arr)-1])

	// Push new array and popped element
	if err := bi.push(NewArray(newArr)); err != nil {
		return err
	}
	return bi.push(lastElement)
}

// Map operations

// execMapNew creates a new empty map
func (bi *BytecodeInterpreter) execMapNew() error {
	// Create empty map
	m := make(map[string]Value)
	return bi.push(NewMap(m))
}

// execMapGet gets a value from a map by key
func (bi *BytecodeInterpreter) execMapGet() error {
	// Pop key from stack
	keyValue, err := bi.pop()
	if err != nil {
		return err
	}

	if keyValue.Type != TypeString {
		return fmt.Errorf("MAPGET requires string for key, got %v", keyValue.Type)
	}

	// Pop map from stack
	mapValue, err := bi.pop()
	if err != nil {
		return err
	}

	if mapValue.Type != TypeMap {
		return fmt.Errorf("MAPGET requires map, got %v", mapValue.Type)
	}

	m, _ := mapValue.AsMap()
	key, _ := keyValue.AsString()

	// Get value from map (returns zero value if not found)
	value, exists := m[key]
	if !exists {
		// Return null (empty bytes) if key not found
		return bi.push(NewBytes(nil))
	}

	return bi.push(value)
}

// execMapSet sets a value in a map by key
func (bi *BytecodeInterpreter) execMapSet() error {
	// Pop value from stack
	value, err := bi.pop()
	if err != nil {
		return err
	}

	// Pop key from stack
	keyValue, err := bi.pop()
	if err != nil {
		return err
	}

	if keyValue.Type != TypeString {
		return fmt.Errorf("MAPSET requires string for key, got %v", keyValue.Type)
	}

	// Pop map from stack
	mapValue, err := bi.pop()
	if err != nil {
		return err
	}

	if mapValue.Type != TypeMap {
		return fmt.Errorf("MAPSET requires map, got %v", mapValue.Type)
	}

	m, _ := mapValue.AsMap()
	key, _ := keyValue.AsString()

	// Create new map with updated value
	newMap := make(map[string]Value, len(m)+1)
	for k, v := range m {
		newMap[k] = v
	}
	newMap[key] = value

	return bi.push(NewMap(newMap))
}

// execMapHas checks if a map has a key
func (bi *BytecodeInterpreter) execMapHas() error {
	// Pop key from stack
	keyValue, err := bi.pop()
	if err != nil {
		return err
	}

	if keyValue.Type != TypeString {
		return fmt.Errorf("MAPHAS requires string for key, got %v", keyValue.Type)
	}

	// Pop map from stack
	mapValue, err := bi.pop()
	if err != nil {
		return err
	}

	if mapValue.Type != TypeMap {
		return fmt.Errorf("MAPHAS requires map, got %v", mapValue.Type)
	}

	m, _ := mapValue.AsMap()
	key, _ := keyValue.AsString()

	// Check if key exists
	_, exists := m[key]

	return bi.push(NewBool(exists))
}

// execMapDel deletes a key from a map
func (bi *BytecodeInterpreter) execMapDel() error {
	// Pop key from stack
	keyValue, err := bi.pop()
	if err != nil {
		return err
	}

	if keyValue.Type != TypeString {
		return fmt.Errorf("MAPDEL requires string for key, got %v", keyValue.Type)
	}

	// Pop map from stack
	mapValue, err := bi.pop()
	if err != nil {
		return err
	}

	if mapValue.Type != TypeMap {
		return fmt.Errorf("MAPDEL requires map, got %v", mapValue.Type)
	}

	m, _ := mapValue.AsMap()
	key, _ := keyValue.AsString()

	// Create new map without the key
	newMap := make(map[string]Value, len(m))
	for k, v := range m {
		if k != key {
			newMap[k] = v
		}
	}

	return bi.push(NewMap(newMap))
}

// Set operations

// execSetNew creates a new empty set
func (bi *BytecodeInterpreter) execSetNew() error {
	// Create empty set
	s := make(map[string]bool)
	return bi.push(NewSet(s))
}

// execSetAdd adds an element to a set
func (bi *BytecodeInterpreter) execSetAdd() error {
	// Pop element from stack
	elemValue, err := bi.pop()
	if err != nil {
		return err
	}

	if elemValue.Type != TypeString {
		return fmt.Errorf("SETADD requires string for element, got %v", elemValue.Type)
	}

	// Pop set from stack
	setValue, err := bi.pop()
	if err != nil {
		return err
	}

	if setValue.Type != TypeSet {
		return fmt.Errorf("SETADD requires set, got %v", setValue.Type)
	}

	s, _ := setValue.AsSet()
	elem, _ := elemValue.AsString()

	// Create new set with added element
	newSet := make(map[string]bool, len(s)+1)
	for k := range s {
		newSet[k] = true
	}
	newSet[elem] = true

	return bi.push(NewSet(newSet))
}

// execSetHas checks if a set has an element
func (bi *BytecodeInterpreter) execSetHas() error {
	// Pop element from stack
	elemValue, err := bi.pop()
	if err != nil {
		return err
	}

	if elemValue.Type != TypeString {
		return fmt.Errorf("SETHAS requires string for element, got %v", elemValue.Type)
	}

	// Pop set from stack
	setValue, err := bi.pop()
	if err != nil {
		return err
	}

	if setValue.Type != TypeSet {
		return fmt.Errorf("SETHAS requires set, got %v", setValue.Type)
	}

	s, _ := setValue.AsSet()
	elem, _ := elemValue.AsString()

	// Check if element exists
	exists := s[elem]

	return bi.push(NewBool(exists))
}

// execSetDel deletes an element from a set
func (bi *BytecodeInterpreter) execSetDel() error {
	// Pop element from stack
	elemValue, err := bi.pop()
	if err != nil {
		return err
	}

	if elemValue.Type != TypeString {
		return fmt.Errorf("SETDEL requires string for element, got %v", elemValue.Type)
	}

	// Pop set from stack
	setValue, err := bi.pop()
	if err != nil {
		return err
	}

	if setValue.Type != TypeSet {
		return fmt.Errorf("SETDEL requires set, got %v", setValue.Type)
	}

	s, _ := setValue.AsSet()
	elem, _ := elemValue.AsString()

	// Create new set without the element
	newSet := make(map[string]bool, len(s))
	for k := range s {
		if k != elem {
			newSet[k] = true
		}
	}

	return bi.push(NewSet(newSet))
}

// Helper function to convert Value to string for sorting/comparison
func valueToString(v Value) string {
	switch v.Type {
	case TypeString:
		s, _ := v.AsString()
		return s
	case TypeI64:
		i, _ := v.AsI64()
		return fmt.Sprintf("%d", i)
	case TypeU64:
		u, _ := v.AsU64()
		return fmt.Sprintf("%d", u)
	case TypeBool:
		b, _ := v.AsBool()
		return fmt.Sprintf("%t", b)
	default:
		return ""
	}
}

// sortValues sorts an array of values deterministically
func sortValues(arr []Value) []Value {
	// Create a copy to avoid modifying the original
	sorted := make([]Value, len(arr))
	copy(sorted, arr)

	// Sort using deterministic string comparison
	sort.Slice(sorted, func(i, j int) bool {
		return valueToString(sorted[i]) < valueToString(sorted[j])
	})

	return sorted
}
