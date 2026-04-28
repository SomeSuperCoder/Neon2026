package rpc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// TestServerConfig tests the server configuration
func TestServerConfig(t *testing.T) {
	t.Run("default config", func(t *testing.T) {
		config := DefaultServerConfig()

		if config.BindAddress != "127.0.0.1" {
			t.Errorf("Expected bind address 127.0.0.1, got %s", config.BindAddress)
		}

		if config.Port != 8899 {
			t.Errorf("Expected port 8899, got %d", config.Port)
		}

		if config.MaxConns != 100 {
			t.Errorf("Expected max connections 100, got %d", config.MaxConns)
		}

		if config.ReadTimeout != 10*time.Second {
			t.Errorf("Expected read timeout 10s, got %v", config.ReadTimeout)
		}

		if config.WriteTimeout != 10*time.Second {
			t.Errorf("Expected write timeout 10s, got %v", config.WriteTimeout)
		}
	})

	t.Run("custom config", func(t *testing.T) {
		config := &ServerConfig{
			BindAddress:  "0.0.0.0",
			Port:         9000,
			MaxConns:     200,
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 5 * time.Second,
		}

		if config.BindAddress != "0.0.0.0" {
			t.Errorf("Expected bind address 0.0.0.0, got %s", config.BindAddress)
		}

		if config.Port != 9000 {
			t.Errorf("Expected port 9000, got %d", config.Port)
		}
	})
}

// TestNewRPCServer tests server creation
func TestNewRPCServer(t *testing.T) {
	t.Run("with default config", func(t *testing.T) {
		ledger, cleanup := createTestLedger(t)
		defer cleanup()

		fs := createTestFileStore(t)
		defer fs.Close()

		server, err := NewRPCServer(nil, ledger, fs, nil, nil)
		if err != nil {
			t.Fatalf("Failed to create server: %v", err)
		}

		if server == nil {
			t.Fatal("Expected non-nil server")
		}

		if server.config == nil {
			t.Fatal("Expected non-nil config")
		}

		if server.handler == nil {
			t.Fatal("Expected non-nil handler")
		}

		if server.logger == nil {
			t.Fatal("Expected non-nil logger")
		}
	})

	t.Run("with custom config", func(t *testing.T) {
		ledger, cleanup := createTestLedger(t)
		defer cleanup()

		fs := createTestFileStore(t)
		defer fs.Close()

		config := &ServerConfig{
			BindAddress:  "127.0.0.1",
			Port:         9999,
			MaxConns:     50,
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 5 * time.Second,
		}

		server, err := NewRPCServer(config, ledger, fs, nil, nil)
		if err != nil {
			t.Fatalf("Failed to create server: %v", err)
		}

		if server.config.Port != 9999 {
			t.Errorf("Expected port 9999, got %d", server.config.Port)
		}
	})

	t.Run("with custom logger", func(t *testing.T) {
		ledger, cleanup := createTestLedger(t)
		defer cleanup()

		fs := createTestFileStore(t)
		defer fs.Close()

		var buf bytes.Buffer
		logger := log.New(&buf, "TEST: ", log.LstdFlags)

		server, err := NewRPCServer(nil, ledger, fs, nil, logger)
		if err != nil {
			t.Fatalf("Failed to create server: %v", err)
		}

		if server.logger != logger {
			t.Error("Expected custom logger to be used")
		}
	})
}

// TestServerStartStop tests server lifecycle
func TestServerStartStop(t *testing.T) {
	t.Run("start and stop", func(t *testing.T) {
		ledger, cleanup := createTestLedger(t)
		defer cleanup()

		fs := createTestFileStore(t)
		defer fs.Close()

		config := &ServerConfig{
			BindAddress:  "127.0.0.1",
			Port:         0, // Use random port
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 5 * time.Second,
		}

		server, err := NewRPCServer(config, ledger, fs, nil, nil)
		if err != nil {
			t.Fatalf("Failed to create server: %v", err)
		}

		// Start server
		if err := server.Start(); err != nil {
			t.Fatalf("Failed to start server: %v", err)
		}

		// Verify server is listening
		addr := server.Address()
		if addr == "" {
			t.Fatal("Expected non-empty address")
		}

		// Stop server
		if err := server.Stop(); err != nil {
			t.Fatalf("Failed to stop server: %v", err)
		}
	})

	t.Run("stop is idempotent", func(t *testing.T) {
		ledger, cleanup := createTestLedger(t)
		defer cleanup()

		fs := createTestFileStore(t)
		defer fs.Close()

		config := &ServerConfig{
			BindAddress:  "127.0.0.1",
			Port:         0,
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 5 * time.Second,
		}

		server, err := NewRPCServer(config, ledger, fs, nil, nil)
		if err != nil {
			t.Fatalf("Failed to create server: %v", err)
		}

		if err := server.Start(); err != nil {
			t.Fatalf("Failed to start server: %v", err)
		}

		// Stop multiple times
		if err := server.Stop(); err != nil {
			t.Fatalf("First stop failed: %v", err)
		}

		if err := server.Stop(); err != nil {
			t.Fatalf("Second stop failed: %v", err)
		}
	})
}

// TestServerHTTPMethods tests HTTP method handling
func TestServerHTTPMethods(t *testing.T) {
	server := createTestServer(t)
	defer server.Stop()

	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}

	baseURL := fmt.Sprintf("http://%s", server.Address())

	tests := []struct {
		name           string
		method         string
		expectedStatus int
	}{
		{"POST allowed", http.MethodPost, http.StatusOK},
		{"GET not allowed", http.MethodGet, http.StatusMethodNotAllowed},
		{"PUT not allowed", http.MethodPut, http.StatusMethodNotAllowed},
		{"DELETE not allowed", http.MethodDelete, http.StatusMethodNotAllowed},
		{"PATCH not allowed", http.MethodPatch, http.StatusMethodNotAllowed},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(tt.method, baseURL, strings.NewReader("{}"))
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatalf("Request failed: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}
		})
	}
}

// TestServerCORS tests CORS header handling
func TestServerCORS(t *testing.T) {
	server := createTestServer(t)
	defer server.Stop()

	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}

	baseURL := fmt.Sprintf("http://%s", server.Address())

	t.Run("OPTIONS preflight", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodOptions, baseURL, nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		// Check CORS headers
		if origin := resp.Header.Get("Access-Control-Allow-Origin"); origin != "*" {
			t.Errorf("Expected Access-Control-Allow-Origin: *, got %s", origin)
		}

		if methods := resp.Header.Get("Access-Control-Allow-Methods"); methods != "POST, OPTIONS" {
			t.Errorf("Expected Access-Control-Allow-Methods: POST, OPTIONS, got %s", methods)
		}

		if headers := resp.Header.Get("Access-Control-Allow-Headers"); headers != "Content-Type" {
			t.Errorf("Expected Access-Control-Allow-Headers: Content-Type, got %s", headers)
		}
	})

	t.Run("POST with CORS", func(t *testing.T) {
		reqBody := `{"jsonrpc":"2.0","method":"getBlockHeight","params":[],"id":1}`
		req, err := http.NewRequest(http.MethodPost, baseURL, strings.NewReader(reqBody))
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		// Check CORS headers are present
		if origin := resp.Header.Get("Access-Control-Allow-Origin"); origin != "*" {
			t.Errorf("Expected Access-Control-Allow-Origin: *, got %s", origin)
		}
	})
}

// TestServerRequestSizeLimit tests request size validation
func TestServerRequestSizeLimit(t *testing.T) {
	server := createTestServer(t)
	defer server.Stop()

	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}

	baseURL := fmt.Sprintf("http://%s", server.Address())

	t.Run("request within limit", func(t *testing.T) {
		// Small request (well under 1MB)
		reqBody := `{"jsonrpc":"2.0","method":"getBlockHeight","params":[],"id":1}`
		resp, err := http.Post(baseURL, "application/json", strings.NewReader(reqBody))
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}
	})

	t.Run("request exceeds limit", func(t *testing.T) {
		// Create a request larger than 1MB
		largeData := strings.Repeat("x", 2*1024*1024) // 2MB
		reqBody := fmt.Sprintf(`{"jsonrpc":"2.0","method":"test","params":{"data":"%s"},"id":1}`, largeData)

		resp, err := http.Post(baseURL, "application/json", strings.NewReader(reqBody))
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		// Should get an error response
		if resp.StatusCode == http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			var jsonResp JSONRPCResponse
			if err := json.Unmarshal(body, &jsonResp); err == nil {
				if jsonResp.Error == nil {
					t.Error("Expected error response for oversized request")
				}
			}
		}
	})
}

// TestServerConcurrentRequests tests concurrent request handling
func TestServerConcurrentRequests(t *testing.T) {
	server := createTestServer(t)
	defer server.Stop()

	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}

	baseURL := fmt.Sprintf("http://%s", server.Address())

	// Number of concurrent requests
	numRequests := 50
	var wg sync.WaitGroup
	var successCount atomic.Int32
	var errorCount atomic.Int32

	for i := 0; i < numRequests; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			reqBody := fmt.Sprintf(`{"jsonrpc":"2.0","method":"getBlockHeight","params":[],"id":%d}`, id)
			resp, err := http.Post(baseURL, "application/json", strings.NewReader(reqBody))
			if err != nil {
				errorCount.Add(1)
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode == http.StatusOK {
				successCount.Add(1)
			} else {
				errorCount.Add(1)
			}
		}(i)
	}

	wg.Wait()

	if successCount.Load() != int32(numRequests) {
		t.Errorf("Expected %d successful requests, got %d (errors: %d)",
			numRequests, successCount.Load(), errorCount.Load())
	}
}

// TestServerGracefulShutdown tests graceful shutdown
func TestServerGracefulShutdown(t *testing.T) {
	server := createTestServer(t)

	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}

	baseURL := fmt.Sprintf("http://%s", server.Address())

	// Start a long-running request
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		reqBody := `{"jsonrpc":"2.0","method":"getBlockHeight","params":[],"id":1}`
		resp, err := http.Post(baseURL, "application/json", strings.NewReader(reqBody))
		if err != nil {
			// Connection might be closed during shutdown, which is acceptable
			return
		}
		defer resp.Body.Close()
	}()

	// Give the request time to start
	time.Sleep(50 * time.Millisecond)

	// Shutdown server
	shutdownDone := make(chan error, 1)
	go func() {
		shutdownDone <- server.Stop()
	}()

	// Wait for shutdown with timeout
	select {
	case err := <-shutdownDone:
		if err != nil {
			t.Errorf("Shutdown failed: %v", err)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("Shutdown timed out")
	}

	wg.Wait()
}

// TestServerContentType tests content type handling
func TestServerContentType(t *testing.T) {
	server := createTestServer(t)
	defer server.Stop()

	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}

	baseURL := fmt.Sprintf("http://%s", server.Address())

	reqBody := `{"jsonrpc":"2.0","method":"getBlockHeight","params":[],"id":1}`
	resp, err := http.Post(baseURL, "application/json", strings.NewReader(reqBody))
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	contentType := resp.Header.Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type: application/json, got %s", contentType)
	}
}

// TestServerRequestLogging tests request logging
func TestServerRequestLogging(t *testing.T) {
	ledger, cleanup := createTestLedger(t)
	defer cleanup()

	fs := createTestFileStore(t)
	defer fs.Close()

	var logBuf bytes.Buffer
	logger := log.New(&logBuf, "", 0)

	config := &ServerConfig{
		BindAddress:  "127.0.0.1",
		Port:         0,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	server, err := NewRPCServer(config, ledger, fs, nil, logger)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}
	defer server.Stop()

	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}

	baseURL := fmt.Sprintf("http://%s", server.Address())

	// Make a request
	reqBody := `{"jsonrpc":"2.0","method":"getBlockHeight","params":[],"id":1}`
	resp, err := http.Post(baseURL, "application/json", strings.NewReader(reqBody))
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	// Check that request was logged
	logOutput := logBuf.String()
	if !strings.Contains(logOutput, "getBlockHeight") {
		t.Error("Expected method name in log output")
	}

	if !strings.Contains(logOutput, "completed") {
		t.Error("Expected 'completed' in log output")
	}
}

// TestServerInvalidJSON tests handling of invalid JSON
func TestServerInvalidJSON(t *testing.T) {
	server := createTestServer(t)
	defer server.Stop()

	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}

	baseURL := fmt.Sprintf("http://%s", server.Address())

	tests := []struct {
		name    string
		body    string
		wantErr bool
		errCode int
	}{
		{"invalid JSON", `{invalid json}`, true, ParseError},
		{"empty body", ``, true, ParseError},
		{"not an object", `"string"`, true, ParseError},
		{"valid JSON but method not found", `{"jsonrpc":"2.0","method":"test","id":1}`, true, MethodNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := http.Post(baseURL, "application/json", strings.NewReader(tt.body))
			if err != nil {
				t.Fatalf("Request failed: %v", err)
			}
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("Failed to read response: %v", err)
			}

			var jsonResp JSONRPCResponse
			if err := json.Unmarshal(body, &jsonResp); err != nil {
				if !tt.wantErr {
					t.Errorf("Failed to parse response: %v", err)
				}
				return
			}

			if tt.wantErr && jsonResp.Error == nil {
				t.Error("Expected error response")
			}

			if !tt.wantErr && jsonResp.Error != nil {
				t.Errorf("Unexpected error: %v", jsonResp.Error.Message)
			}

			if tt.wantErr && jsonResp.Error != nil && tt.errCode != 0 {
				if jsonResp.Error.Code != tt.errCode {
					t.Errorf("Expected error code %d, got %d", tt.errCode, jsonResp.Error.Code)
				}
			}
		})
	}
}

// Helper functions

func createTestServer(t *testing.T) *RPCServer {
	t.Helper()

	ledger, cleanup := createTestLedger(t)
	t.Cleanup(cleanup)

	fs := createTestFileStore(t)
	t.Cleanup(func() { fs.Close() })

	config := &ServerConfig{
		BindAddress:  "127.0.0.1",
		Port:         0, // Use random port
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	server, err := NewRPCServer(config, ledger, fs, nil, nil)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	return server
}
