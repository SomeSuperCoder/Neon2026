package main

import (
	"os"
	"path/filepath"
	"testing"
)

// TestExpandPath tests the expandPath function
func TestExpandPath(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
		checkFn func(string) bool
	}{
		{
			name:    "empty path",
			path:    "",
			wantErr: false,
			checkFn: func(result string) bool { return result == "" },
		},
		{
			name:    "path without tilde",
			path:    "/absolute/path",
			wantErr: false,
			checkFn: func(result string) bool { return result == "/absolute/path" },
		},
		{
			name:    "tilde only",
			path:    "~",
			wantErr: false,
			checkFn: func(result string) bool {
				homeDir, _ := os.UserHomeDir()
				return result == homeDir
			},
		},
		{
			name:    "tilde with path",
			path:    "~/.poh-wallet/wallet.dat",
			wantErr: false,
			checkFn: func(result string) bool {
				homeDir, _ := os.UserHomeDir()
				expected := filepath.Join(homeDir, ".poh-wallet/wallet.dat")
				return result == expected
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := expandPath(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("expandPath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && !tt.checkFn(result) {
				t.Errorf("expandPath() = %v, validation failed", result)
			}
		})
	}
}

// TestFileExists tests the fileExists function
func TestFileExists(t *testing.T) {
	// Create a temporary file
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.txt")

	// File doesn't exist yet
	if fileExists(tmpFile) {
		t.Error("fileExists() returned true for non-existent file")
	}

	// Create the file
	if err := os.WriteFile(tmpFile, []byte("test"), 0600); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// File should exist now
	if !fileExists(tmpFile) {
		t.Error("fileExists() returned false for existing file")
	}

	// Non-existent path
	if fileExists("/nonexistent/path/file.txt") {
		t.Error("fileExists() returned true for non-existent path")
	}
}

// TestDefaultConstants tests that default constants are set correctly
func TestDefaultConstants(t *testing.T) {
	if defaultWalletPath != "~/.poh-wallet/wallet.dat" {
		t.Errorf("defaultWalletPath = %s, want ~/.poh-wallet/wallet.dat", defaultWalletPath)
	}

	if defaultRPCURL != "http://localhost:8899" {
		t.Errorf("defaultRPCURL = %s, want http://localhost:8899", defaultRPCURL)
	}
}
