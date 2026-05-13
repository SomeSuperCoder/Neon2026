package main

import (
	"os"
	"testing"
)

func TestAuditCommandLine(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		shouldFail bool
	}{
		{
			name:       "help flag",
			args:       []string{"--help"},
			shouldFail: false,
		},
		{
			name:       "default parameters",
			args:       []string{"--duration", "30", "--validators", "3"},
			shouldFail: false,
		},
		{
			name:       "bft subcommand",
			args:       []string{"bft", "--honest", "4", "--malicious", "1", "--duration", "60"},
			shouldFail: false,
		},
		{
			name:       "dpos subcommand",
			args:       []string{"dpos", "--validators", "5", "--duration", "120"},
			shouldFail: false,
		},
		{
			name:       "report subcommand",
			args:       []string{"report", "--db-dir", "./test-results"},
			shouldFail: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original args
			oldArgs := os.Args
			defer func() { os.Args = oldArgs }()

			// Set up test args
			os.Args = append([]string{"audit"}, tt.args...)

			// Test that main runs without panicking
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("main panicked with args %v: %v", tt.args, r)
				}
			}()

			// Run main
			main()
		})
	}
}

func TestMainFunction(t *testing.T) {
	// Test that main doesn't panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("main panicked: %v", r)
		}
	}()

	// Run with minimal arguments
	oldArgs := os.Args
	os.Args = []string{"audit", "--duration", "1", "--validators", "1"}
	defer func() { os.Args = oldArgs }()

	// This should run without panicking
	main()
}
