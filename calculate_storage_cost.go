package main

import (
	"fmt"
	"math"
	"os"
)

// CalculateStorageCost calculates the storage cost for a given data size
// using an exponential growth formula: cost = base * size_in_kb * (1.1 ^ size_in_mb)
func CalculateStorageCost(dataSize int64) int64 {
	if dataSize == 0 {
		return 0
	}

	// Round up to nearest KB
	sizeInKB := (dataSize + 1023) / 1024
	if sizeInKB == 0 {
		sizeInKB = 1
	}

	// Calculate size in MB for exponential factor
	sizeInMB := float64(dataSize) / (1024.0 * 1024.0)

	// Apply exponential growth: 1.1 ^ size_in_mb
	multiplier := math.Pow(1.1, sizeInMB)

	cost := int64(float64(1000*sizeInKB) * multiplier)
	return cost
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run calculate_storage_cost.go <file_path>")
		os.Exit(1)
	}

	filePath := os.Args[1]

	// Read file to get size
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		os.Exit(1)
	}

	dataSize := fileInfo.Size()
	storageCost := CalculateStorageCost(dataSize)

	fmt.Printf("File: %s\n", filePath)
	fmt.Printf("Size: %d bytes\n", dataSize)
	fmt.Printf("Storage Cost: %d Neon units\n", storageCost)
}
