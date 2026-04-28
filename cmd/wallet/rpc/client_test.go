package rpc

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/poh-blockchain/internal/rpc"
)

func TestNewRPCClient(t *testing.T) {
	endpoint := "http://localhost:8899"
	client := NewRPCClient(endpoint)

	if client == nil {
		t.Fatal("NewRPCClient returned nil")
	}

	if client.endpoint != endpoint {
		t.Errorf("Expected endpoint %s, got %s", endpoint, client.endpoint)
	}

	if client.httpClient == nil {
		t.Error("HTTP client is nil")
	}

	if client.httpClient.Timeout != 10*time.Second {
		t.Errorf("Expected timeout 10s, got %v", client.httpClient.Timeout)
	}
}

func TestRPCClient_BuildRequest(t *testing.T) {
	client := NewRPCClient("http://localhost:8899")

	tests := []struct {
		name   string
		method string
		params interface{}
	}{
		{
			name:   "simple method with string param",
			method: "getBalance",
			params: map[string]string{"address": "0x1234"},
		},
		{
			name:   "method with multiple params",
			method: "getTransactionHistory",
			params: map[string]interface{}{"address": "0x1234", "limit": 20},
		},
		{
			name:   "method with nil params",
			method: "getBlockHeight",
			params: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := client.buildRequest(tt.method, tt.params)
			if err != nil {
				t.Fatalf("buildRequest failed: %v", err)
			}

			if req.JSONRPC != "2.0" {
				t.Errorf("Expected JSONRPC 2.0, got %s", req.JSONRPC)
			}

			if req.Method != tt.method {
				t.Errorf("Expected method %s, got %s", tt.method, req.Method)
			}

			// Check ID is auto-incrementing
			if req.ID == nil {
				t.Error("Request ID is nil")
			}
		})
	}
}

func TestRPCClient_AutoIncrementingID(t *testing.T) {
	client := NewRPCClient("http://localhost:8899")

	req1, _ := client.buildRequest("method1", nil)
	req2, _ := client.buildRequest("method2", nil)
	req3, _ := client.buildRequest("method3", nil)

	id1, ok1 := req1.ID.(uint64)
	id2, ok2 := req2.ID.(uint64)
	id3, ok3 := req3.ID.(uint64)

	if !ok1 || !ok2 || !ok3 {
		t.Fatal("Request IDs are not uint64")
	}

	if id2 != id1+1 {
		t.Errorf("Expected ID %d, got %d", id1+1, id2)
	}

	if id3 != id2+1 {
		t.Errorf("Expected ID %d, got %d", id2+1, id3)
	}
}

func TestRPCClient_RequestTimeout(t *testing.T) {
	// Create a server that delays response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(15 * time.Second) // Longer than 10s timeout
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewRPCClient(server.URL)

	_, err := client.call("testMethod", nil)
	if err == nil {
		t.Error("Expected timeout error, got nil")
	}
}

func TestRPCClient_Call_Success(t *testing.T) {
	// Create a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
		}

		// Parse request
		var req rpc.JSONRPCRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request: %v", err)
		}

		// Send response
		response := rpc.JSONRPCResponse{
			JSONRPC: "2.0",
			Result:  int64(1000000),
			ID:      req.ID,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewRPCClient(server.URL)

	result, err := client.call("getBalance", map[string]string{"address": "0x1234"})
	if err != nil {
		t.Fatalf("call failed: %v", err)
	}

	balance, ok := result.(float64) // JSON numbers are decoded as float64
	if !ok {
		t.Fatalf("Expected float64 result, got %T", result)
	}

	if int64(balance) != 1000000 {
		t.Errorf("Expected balance 1000000, got %d", int64(balance))
	}
}

func TestRPCClient_Call_RPCError(t *testing.T) {
	// Create a mock server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req rpc.JSONRPCRequest
		json.NewDecoder(r.Body).Decode(&req)

		response := rpc.JSONRPCResponse{
			JSONRPC: "2.0",
			Error: &rpc.RPCError{
				Code:    -32601,
				Message: "Method not found",
			},
			ID: req.ID,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewRPCClient(server.URL)

	_, err := client.call("unknownMethod", nil)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	rpcErr, ok := err.(*RPCError)
	if !ok {
		t.Fatalf("Expected RPCError, got %T", err)
	}

	if rpcErr.Code != -32601 {
		t.Errorf("Expected error code -32601, got %d", rpcErr.Code)
	}

	if rpcErr.Message != "Method not found" {
		t.Errorf("Expected message 'Method not found', got %s", rpcErr.Message)
	}
}

func TestRPCClient_Call_InvalidEndpoint(t *testing.T) {
	client := NewRPCClient("http://invalid-endpoint-that-does-not-exist:9999")

	_, err := client.call("testMethod", nil)
	if err == nil {
		t.Error("Expected error for invalid endpoint, got nil")
	}
}

func TestRPCClient_Call_InvalidJSON(t *testing.T) {
	// Create a server that returns invalid JSON
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("invalid json"))
	}))
	defer server.Close()

	client := NewRPCClient(server.URL)

	_, err := client.call("testMethod", nil)
	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
}

// Tests for RPC client methods

func TestRPCClient_GetBalance(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req rpc.JSONRPCRequest
		json.NewDecoder(r.Body).Decode(&req)

		if req.Method != "getBalance" {
			t.Errorf("Expected method getBalance, got %s", req.Method)
		}

		// Parse params
		var params map[string]string
		json.Unmarshal(req.Params, &params)

		if params["address"] != "test-address" {
			t.Errorf("Expected address test-address, got %s", params["address"])
		}

		response := rpc.JSONRPCResponse{
			JSONRPC: "2.0",
			Result:  int64(1000000),
			ID:      req.ID,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewRPCClient(server.URL)

	balance, err := client.GetBalance("test-address")
	if err != nil {
		t.Fatalf("GetBalance failed: %v", err)
	}

	if balance != 1000000 {
		t.Errorf("Expected balance 1000000, got %d", balance)
	}
}

func TestRPCClient_GetAccountInfo(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req rpc.JSONRPCRequest
		json.NewDecoder(r.Body).Decode(&req)

		if req.Method != "getAccountInfo" {
			t.Errorf("Expected method getAccountInfo, got %s", req.Method)
		}

		accountInfo := rpc.AccountInfo{
			Address:    "test-address",
			Balance:    1000000,
			Owner:      "system-program",
			DataLength: 0,
			Executable: false,
		}

		response := rpc.JSONRPCResponse{
			JSONRPC: "2.0",
			Result:  accountInfo,
			ID:      req.ID,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewRPCClient(server.URL)

	accountInfo, err := client.GetAccountInfo("test-address")
	if err != nil {
		t.Fatalf("GetAccountInfo failed: %v", err)
	}

	if accountInfo.Address != "test-address" {
		t.Errorf("Expected address test-address, got %s", accountInfo.Address)
	}

	if accountInfo.Balance != 1000000 {
		t.Errorf("Expected balance 1000000, got %d", accountInfo.Balance)
	}

	if accountInfo.Owner != "system-program" {
		t.Errorf("Expected owner system-program, got %s", accountInfo.Owner)
	}
}

func TestRPCClient_GetTransactionHistory(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req rpc.JSONRPCRequest
		json.NewDecoder(r.Body).Decode(&req)

		if req.Method != "getTransactionHistory" {
			t.Errorf("Expected method getTransactionHistory, got %s", req.Method)
		}

		// Parse params
		var params map[string]interface{}
		json.Unmarshal(req.Params, &params)

		if params["address"] != "test-address" {
			t.Errorf("Expected address test-address, got %v", params["address"])
		}

		if int(params["limit"].(float64)) != 20 {
			t.Errorf("Expected limit 20, got %v", params["limit"])
		}

		txRecords := []rpc.TransactionRecord{
			{
				Signature:   "tx-sig-1",
				BlockHeight: 100,
				Slot:        100,
				Timestamp:   time.Now(),
				Success:     true,
			},
			{
				Signature:   "tx-sig-2",
				BlockHeight: 99,
				Slot:        99,
				Timestamp:   time.Now().Add(-1 * time.Minute),
				Success:     true,
			},
		}

		response := rpc.JSONRPCResponse{
			JSONRPC: "2.0",
			Result:  txRecords,
			ID:      req.ID,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewRPCClient(server.URL)

	txHistory, err := client.GetTransactionHistory("test-address", 20)
	if err != nil {
		t.Fatalf("GetTransactionHistory failed: %v", err)
	}

	if len(txHistory) != 2 {
		t.Errorf("Expected 2 transactions, got %d", len(txHistory))
	}

	if txHistory[0].Signature != "tx-sig-1" {
		t.Errorf("Expected signature tx-sig-1, got %s", txHistory[0].Signature)
	}
}

func TestRPCClient_SendTransaction(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req rpc.JSONRPCRequest
		json.NewDecoder(r.Body).Decode(&req)

		if req.Method != "sendTransaction" {
			t.Errorf("Expected method sendTransaction, got %s", req.Method)
		}

		// Parse params
		var params map[string]string
		json.Unmarshal(req.Params, &params)

		if params["transaction"] == "" {
			t.Error("Expected transaction parameter")
		}

		response := rpc.JSONRPCResponse{
			JSONRPC: "2.0",
			Result:  "transaction-signature-123",
			ID:      req.ID,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewRPCClient(server.URL)

	// Create a dummy transaction (base64 encoded)
	txData := []byte("dummy-transaction-data")

	signature, err := client.SendTransaction(txData)
	if err != nil {
		t.Fatalf("SendTransaction failed: %v", err)
	}

	if signature != "transaction-signature-123" {
		t.Errorf("Expected signature transaction-signature-123, got %s", signature)
	}
}

func TestRPCClient_GetTransactionStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req rpc.JSONRPCRequest
		json.NewDecoder(r.Body).Decode(&req)

		if req.Method != "getTransactionStatus" {
			t.Errorf("Expected method getTransactionStatus, got %s", req.Method)
		}

		// Parse params
		var params map[string]string
		json.Unmarshal(req.Params, &params)

		if params["signature"] != "test-signature" {
			t.Errorf("Expected signature test-signature, got %s", params["signature"])
		}

		txStatus := rpc.TransactionStatus{
			Signature:   "test-signature",
			Confirmed:   true,
			BlockHeight: 100,
			Slot:        100,
		}

		response := rpc.JSONRPCResponse{
			JSONRPC: "2.0",
			Result:  txStatus,
			ID:      req.ID,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewRPCClient(server.URL)

	status, err := client.GetTransactionStatus("test-signature")
	if err != nil {
		t.Fatalf("GetTransactionStatus failed: %v", err)
	}

	if status.Signature != "test-signature" {
		t.Errorf("Expected signature test-signature, got %s", status.Signature)
	}

	if !status.Confirmed {
		t.Error("Expected transaction to be confirmed")
	}

	if status.BlockHeight != 100 {
		t.Errorf("Expected block height 100, got %d", status.BlockHeight)
	}
}

func TestRPCClient_GetBlockHeight(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req rpc.JSONRPCRequest
		json.NewDecoder(r.Body).Decode(&req)

		if req.Method != "getBlockHeight" {
			t.Errorf("Expected method getBlockHeight, got %s", req.Method)
		}

		response := rpc.JSONRPCResponse{
			JSONRPC: "2.0",
			Result:  int64(12345),
			ID:      req.ID,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewRPCClient(server.URL)

	height, err := client.GetBlockHeight()
	if err != nil {
		t.Fatalf("GetBlockHeight failed: %v", err)
	}

	if height != 12345 {
		t.Errorf("Expected block height 12345, got %d", height)
	}
}
