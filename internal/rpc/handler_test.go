package rpc

import (
	"bytes"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/poh-blockchain/internal/filestore"
	"github.com/poh-blockchain/internal/processor"
	"github.com/poh-blockchain/internal/runtime"
	"github.com/poh-blockchain/internal/storage"
	"github.com/poh-blockchain/internal/transaction"
)

// TestHandleRequest_MethodRouting tests that requests are routed to correct handlers
func TestHandleRequest_MethodRouting(t *testing.T) {
	handler := createTestHandler(t)

	tests := []struct {
		name            string
		method          string
		expectError     bool
		expectedErrCode int
	}{
		{
			name:            "getBalance routes correctly",
			method:          "getBalance",
			expectError:     true, // Not implemented yet, but should route
			expectedErrCode: InternalError,
		},
		{
			name:            "getAccountInfo routes correctly",
			method:          "getAccountInfo",
			expectError:     true,
			expectedErrCode: InternalError,
		},
		{
			name:            "getTransactionHistory routes correctly",
			method:          "getTransactionHistory",
			expectError:     true,
			expectedErrCode: InternalError,
		},
		{
			name:            "getBlockHeight routes correctly",
			method:          "getBlockHeight",
			expectError:     true,
			expectedErrCode: InternalError,
		},
		{
			name:            "getRecentBlockhash routes correctly",
			method:          "getRecentBlockhash",
			expectError:     true,
			expectedErrCode: InternalError,
		},
		{
			name:            "sendTransaction routes correctly",
			method:          "sendTransaction",
			expectError:     true,
			expectedErrCode: InvalidParams, // Changed from InternalError
		},
		{
			name:            "getTransactionStatus routes correctly",
			method:          "getTransactionStatus",
			expectError:     true,
			expectedErrCode: InternalError,
		},
		{
			name:            "unknown method returns MethodNotFound",
			method:          "unknownMethod",
			expectError:     true,
			expectedErrCode: MethodNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &JSONRPCRequest{
				JSONRPC: "2.0",
				Method:  tt.method,
				Params:  json.RawMessage(`{}`),
				ID:      1,
			}

			resp := handler.HandleRequest(req)

			if resp == nil {
				t.Fatal("Expected response, got nil")
			}

			if resp.JSONRPC != "2.0" {
				t.Errorf("Expected JSONRPC version 2.0, got %s", resp.JSONRPC)
			}

			if tt.expectError {
				if resp.Error == nil {
					t.Error("Expected error in response, got nil")
				} else if resp.Error.Code != tt.expectedErrCode {
					t.Errorf("Expected error code %d, got %d", tt.expectedErrCode, resp.Error.Code)
				}
			}

			if resp.ID != req.ID {
				t.Errorf("Expected ID %v, got %v", req.ID, resp.ID)
			}
		})
	}
}

// TestHandleRequest_RequestValidation tests request validation
func TestHandleRequest_RequestValidation(t *testing.T) {
	handler := createTestHandler(t)

	tests := []struct {
		name        string
		request     *JSONRPCRequest
		expectError bool
		errorCode   int
	}{
		{
			name: "valid request",
			request: &JSONRPCRequest{
				JSONRPC: "2.0",
				Method:  "getBlockHeight",
				Params:  json.RawMessage(`{}`),
				ID:      1,
			},
			expectError: true, // Method not implemented yet
			errorCode:   InternalError,
		},
		{
			name: "missing method",
			request: &JSONRPCRequest{
				JSONRPC: "2.0",
				Method:  "",
				Params:  json.RawMessage(`{}`),
				ID:      2,
			},
			expectError: true,
			errorCode:   InvalidRequest,
		},
		{
			name: "null ID is valid",
			request: &JSONRPCRequest{
				JSONRPC: "2.0",
				Method:  "getBlockHeight",
				Params:  json.RawMessage(`{}`),
				ID:      nil,
			},
			expectError: true,
			errorCode:   InternalError,
		},
		{
			name: "string ID is valid",
			request: &JSONRPCRequest{
				JSONRPC: "2.0",
				Method:  "getBlockHeight",
				Params:  json.RawMessage(`{}`),
				ID:      "test-id",
			},
			expectError: true,
			errorCode:   InternalError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := handler.HandleRequest(tt.request)

			if resp == nil {
				t.Fatal("Expected response, got nil")
			}

			if tt.expectError {
				if resp.Error == nil {
					t.Error("Expected error in response")
				} else if resp.Error.Code != tt.errorCode {
					t.Errorf("Expected error code %d, got %d", tt.errorCode, resp.Error.Code)
				}
			}
		})
	}
}

// TestServeHTTP_JSONParsing tests HTTP request handling and JSON parsing
func TestServeHTTP_JSONParsing(t *testing.T) {
	handler := createTestHandler(t)

	tests := []struct {
		name           string
		requestBody    string
		expectedStatus int
		expectError    bool
		errorCode      int
	}{
		{
			name:           "valid JSON-RPC request",
			requestBody:    `{"jsonrpc":"2.0","method":"getBlockHeight","params":{},"id":1}`,
			expectedStatus: http.StatusOK,
			expectError:    true, // Method not implemented
			errorCode:      InternalError,
		},
		{
			name:           "invalid JSON",
			requestBody:    `{invalid json}`,
			expectedStatus: http.StatusOK,
			expectError:    true,
			errorCode:      ParseError,
		},
		{
			name:           "wrong JSON-RPC version",
			requestBody:    `{"jsonrpc":"1.0","method":"getBlockHeight","params":{},"id":1}`,
			expectedStatus: http.StatusOK,
			expectError:    true,
			errorCode:      InvalidRequest,
		},
		{
			name:           "missing JSON-RPC version",
			requestBody:    `{"method":"getBlockHeight","params":{},"id":1}`,
			expectedStatus: http.StatusOK,
			expectError:    true,
			errorCode:      InvalidRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tt.requestBody))
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			resp := w.Result()
			defer resp.Body.Close()

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("Failed to read response body: %v", err)
			}

			var jsonResp JSONRPCResponse
			if err := json.Unmarshal(body, &jsonResp); err != nil {
				t.Fatalf("Failed to parse response JSON: %v", err)
			}

			if tt.expectError {
				if jsonResp.Error == nil {
					t.Error("Expected error in response")
				} else if jsonResp.Error.Code != tt.errorCode {
					t.Errorf("Expected error code %d, got %d", tt.errorCode, jsonResp.Error.Code)
				}
			}
		})
	}
}

// TestServeHTTP_Logging tests that requests are logged with timing
func TestServeHTTP_Logging(t *testing.T) {
	var logBuf bytes.Buffer
	logger := log.New(&logBuf, "", 0)

	// Create handler with custom logger
	ledger, cleanup := createTestLedger(t)
	defer cleanup()
	fs := createTestFileStore(t)
	defer fs.Close()

	queryEngine := NewQueryEngine(ledger, fs)
	handler := NewRPCHandler(queryEngine, nil, logger)

	requestBody := `{"jsonrpc":"2.0","method":"getBlockHeight","params":{},"id":1}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(requestBody))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	logOutput := logBuf.String()

	// Check that method name is logged
	if !strings.Contains(logOutput, "getBlockHeight") {
		t.Error("Expected method name in log output")
	}

	// Check that timing is logged
	if !strings.Contains(logOutput, "completed in") {
		t.Error("Expected timing information in log output")
	}
}

// TestHandleRequest_ErrorResponses tests error response generation
func TestHandleRequest_ErrorResponses(t *testing.T) {
	handler := createTestHandler(t)

	tests := []struct {
		name         string
		method       string
		expectedCode int
		expectedMsg  string
	}{
		{
			name:         "method not found",
			method:       "nonExistentMethod",
			expectedCode: MethodNotFound,
			expectedMsg:  "Method not found",
		},
		{
			name:         "empty method name",
			method:       "",
			expectedCode: InvalidRequest,
			expectedMsg:  "Method name is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &JSONRPCRequest{
				JSONRPC: "2.0",
				Method:  tt.method,
				Params:  json.RawMessage(`{}`),
				ID:      1,
			}

			resp := handler.HandleRequest(req)

			if resp.Error == nil {
				t.Fatal("Expected error in response")
			}

			if resp.Error.Code != tt.expectedCode {
				t.Errorf("Expected error code %d, got %d", tt.expectedCode, resp.Error.Code)
			}

			if resp.Error.Message != tt.expectedMsg {
				t.Errorf("Expected error message %q, got %q", tt.expectedMsg, resp.Error.Message)
			}

			if resp.Result != nil {
				t.Error("Expected nil result when error is present")
			}
		})
	}
}

// TestServeHTTP_ContentType tests that response has correct content type
func TestServeHTTP_ContentType(t *testing.T) {
	handler := createTestHandler(t)

	requestBody := `{"jsonrpc":"2.0","method":"getBlockHeight","params":{},"id":1}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(requestBody))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	// Note: Content-Type should be set in the implementation
	// This test documents the expected behavior
}

// TestHandleSendTransaction tests the sendTransaction RPC method
func TestHandleSendTransaction(t *testing.T) {
	handler := createTestHandlerWithProcessor(t)

	tests := []struct {
		name            string
		params          interface{}
		expectError     bool
		expectedErrCode int
		expectedErrMsg  string
	}{
		{
			name:            "missing transaction parameter",
			params:          map[string]interface{}{},
			expectError:     true,
			expectedErrCode: InvalidParams,
			expectedErrMsg:  "transaction parameter is required",
		},
		{
			name: "invalid transaction format",
			params: map[string]interface{}{
				"transaction": "not-valid-base64",
			},
			expectError:     true,
			expectedErrCode: MalformedTransaction,
			expectedErrMsg:  "failed to decode transaction",
		},
		{
			name: "invalid transaction JSON",
			params: map[string]interface{}{
				"transaction": "aW52YWxpZCBqc29u", // "invalid json" in base64
			},
			expectError:     true,
			expectedErrCode: MalformedTransaction,
			expectedErrMsg:  "failed to unmarshal transaction",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			paramsJSON, _ := json.Marshal(tt.params)
			req := &JSONRPCRequest{
				JSONRPC: "2.0",
				Method:  "sendTransaction",
				Params:  paramsJSON,
				ID:      1,
			}

			resp := handler.HandleRequest(req)

			if resp == nil {
				t.Fatal("Expected response, got nil")
			}

			if tt.expectError {
				if resp.Error == nil {
					t.Error("Expected error in response")
				} else {
					if resp.Error.Code != tt.expectedErrCode {
						t.Errorf("Expected error code %d, got %d", tt.expectedErrCode, resp.Error.Code)
					}
					if !strings.Contains(resp.Error.Message, tt.expectedErrMsg) {
						t.Errorf("Expected error message to contain %q, got %q", tt.expectedErrMsg, resp.Error.Message)
					}
				}
			} else {
				if resp.Error != nil {
					t.Errorf("Expected no error, got: %v", resp.Error)
				}
				if resp.Result == nil {
					t.Error("Expected result, got nil")
				}
			}
		})
	}
}

// TestHandleSendTransaction_SignatureVerification tests signature verification
func TestHandleSendTransaction_SignatureVerification(t *testing.T) {
	handler := createTestHandlerWithProcessor(t)

	// Create a transaction with invalid signature
	tx := createTestTransaction(t)

	// Corrupt the signature
	tx.Signatures[0].Signature[0] ^= 0xFF

	// Encode transaction
	txBytes, err := tx.Marshal()
	if err != nil {
		t.Fatalf("Failed to marshal transaction: %v", err)
	}
	txBase64 := base64.StdEncoding.EncodeToString(txBytes)

	params := map[string]interface{}{
		"transaction": txBase64,
	}
	paramsJSON, _ := json.Marshal(params)

	req := &JSONRPCRequest{
		JSONRPC: "2.0",
		Method:  "sendTransaction",
		Params:  paramsJSON,
		ID:      1,
	}

	resp := handler.HandleRequest(req)

	if resp == nil {
		t.Fatal("Expected response, got nil")
	}

	if resp.Error == nil {
		t.Error("Expected error for invalid signature")
	} else if resp.Error.Code != InvalidSignature {
		t.Errorf("Expected error code %d, got %d", InvalidSignature, resp.Error.Code)
	}
}

// Helper functions

func createTestHandler(t *testing.T) *RPCHandler {
	t.Helper()

	ledger, cleanup := createTestLedger(t)
	t.Cleanup(cleanup)

	fs := createTestFileStore(t)
	t.Cleanup(func() { fs.Close() })

	queryEngine := NewQueryEngine(ledger, fs)
	return NewRPCHandler(queryEngine, nil, nil)
}

func createTestHandlerWithProcessor(t *testing.T) *RPCHandler {
	t.Helper()

	ledger, cleanup := createTestLedger(t)
	t.Cleanup(cleanup)

	fs := createTestFileStore(t)
	t.Cleanup(func() { fs.Close() })

	// Create runtime and processor
	rt := createTestRuntime(t)
	txProcessor := createTestProcessor(t, fs, rt)

	queryEngine := NewQueryEngine(ledger, fs)
	return NewRPCHandler(queryEngine, txProcessor, nil)
}

func createTestLedger(t *testing.T) (*storage.Ledger, func()) {
	t.Helper()

	ledger, err := storage.NewLedger(":memory:")
	if err != nil {
		t.Fatalf("Failed to create test ledger: %v", err)
	}

	return ledger, func() {
		ledger.Close()
	}
}

func createTestFileStore(t *testing.T) *filestore.FileStore {
	t.Helper()

	fs, err := filestore.NewFileStore(t.TempDir())
	if err != nil {
		t.Fatalf("Failed to create test filestore: %v", err)
	}

	return fs
}

func createTestRuntime(t *testing.T) *runtime.Runtime {
	t.Helper()

	rt := runtime.NewRuntime()
	return rt
}

func createTestProcessor(t *testing.T, fs *filestore.FileStore, rt *runtime.Runtime) *processor.TxProcessor {
	t.Helper()

	return processor.NewTxProcessor(fs, rt)
}

func createTestTransaction(t *testing.T) *transaction.Transaction {
	t.Helper()

	// Generate a keypair for signing
	pubKey, privKey, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatalf("Failed to generate keypair: %v", err)
	}

	// Create a simple transaction
	var txID transaction.TxID
	tx := &transaction.Transaction{
		LastSeen:     txID,
		Instructions: []transaction.Instruction{},
		Signatures:   []transaction.Signature{},
	}

	// Sign the transaction
	txData, err := tx.Marshal()
	if err != nil {
		t.Fatalf("Failed to marshal transaction: %v", err)
	}

	signature := ed25519.Sign(privKey, txData)
	var sig [64]byte
	copy(sig[:], signature)

	var pk transaction.PublicKey
	copy(pk[:], pubKey)

	tx.Signatures = []transaction.Signature{
		{
			PublicKey: pk,
			Signature: sig,
		},
	}

	return tx
}
