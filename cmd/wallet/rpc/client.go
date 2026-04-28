package rpc

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/poh-blockchain/internal/rpc"
)

// RPCClient is a JSON-RPC 2.0 client for communicating with the RPC node
type RPCClient struct {
	endpoint   string
	httpClient *http.Client
	requestID  uint64
}

// RPCError represents an RPC error returned by the server
type RPCError struct {
	Code    int
	Message string
	Data    interface{}
}

// Error implements the error interface
func (e *RPCError) Error() string {
	if e.Data != nil {
		return fmt.Sprintf("RPC error %d: %s (data: %v)", e.Code, e.Message, e.Data)
	}
	return fmt.Sprintf("RPC error %d: %s", e.Code, e.Message)
}

// NewRPCClient creates a new RPC client with the specified endpoint
func NewRPCClient(endpoint string) *RPCClient {
	return &RPCClient{
		endpoint: endpoint,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		requestID: 0,
	}
}

// buildRequest creates a JSON-RPC request with auto-incrementing ID
func (c *RPCClient) buildRequest(method string, params interface{}) (*rpc.JSONRPCRequest, error) {
	// Increment request ID atomically
	id := atomic.AddUint64(&c.requestID, 1)

	// Marshal params to JSON
	var paramsJSON json.RawMessage
	if params != nil {
		paramBytes, err := json.Marshal(params)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal params: %w", err)
		}
		paramsJSON = paramBytes
	}

	return &rpc.JSONRPCRequest{
		JSONRPC: "2.0",
		Method:  method,
		Params:  paramsJSON,
		ID:      id,
	}, nil
}

// call makes a JSON-RPC call and returns the result
func (c *RPCClient) call(method string, params interface{}) (interface{}, error) {
	// Build request
	req, err := c.buildRequest(method, params)
	if err != nil {
		return nil, fmt.Errorf("failed to build request: %w", err)
	}

	// Marshal request to JSON
	reqBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequest(http.MethodPost, c.endpoint, bytes.NewReader(reqBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	// Send request
	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer httpResp.Body.Close()

	// Parse response
	var resp rpc.JSONRPCResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Check for RPC error
	if resp.Error != nil {
		return nil, &RPCError{
			Code:    resp.Error.Code,
			Message: resp.Error.Message,
			Data:    resp.Error.Data,
		}
	}

	return resp.Result, nil
}

// GetBalance retrieves the balance for the specified address
func (c *RPCClient) GetBalance(address string) (int64, error) {
	params := map[string]string{
		"address": address,
	}

	result, err := c.call("getBalance", params)
	if err != nil {
		return 0, fmt.Errorf("getBalance failed: %w", err)
	}

	// JSON numbers are decoded as float64
	balance, ok := result.(float64)
	if !ok {
		return 0, fmt.Errorf("unexpected result type: %T", result)
	}

	return int64(balance), nil
}

// GetAccountInfo retrieves full account information for the specified address
func (c *RPCClient) GetAccountInfo(address string) (*rpc.AccountInfo, error) {
	params := map[string]string{
		"address": address,
	}

	result, err := c.call("getAccountInfo", params)
	if err != nil {
		return nil, fmt.Errorf("getAccountInfo failed: %w", err)
	}

	// Convert result to AccountInfo
	// The result is a map[string]interface{} from JSON decoding
	resultBytes, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result: %w", err)
	}

	var accountInfo rpc.AccountInfo
	if err := json.Unmarshal(resultBytes, &accountInfo); err != nil {
		return nil, fmt.Errorf("failed to unmarshal account info: %w", err)
	}

	return &accountInfo, nil
}

// GetTransactionHistory retrieves transaction history for the specified address
func (c *RPCClient) GetTransactionHistory(address string, limit int) ([]rpc.TransactionRecord, error) {
	params := map[string]interface{}{
		"address": address,
		"limit":   limit,
	}

	result, err := c.call("getTransactionHistory", params)
	if err != nil {
		return nil, fmt.Errorf("getTransactionHistory failed: %w", err)
	}

	// Convert result to []TransactionRecord
	resultBytes, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result: %w", err)
	}

	var txRecords []rpc.TransactionRecord
	if err := json.Unmarshal(resultBytes, &txRecords); err != nil {
		return nil, fmt.Errorf("failed to unmarshal transaction records: %w", err)
	}

	return txRecords, nil
}

// SendTransaction submits a signed transaction to the blockchain
// txData should be the marshaled transaction bytes
func (c *RPCClient) SendTransaction(txData []byte) (string, error) {
	// Encode transaction as base64
	txBase64 := base64.StdEncoding.EncodeToString(txData)

	params := map[string]string{
		"transaction": txBase64,
	}

	result, err := c.call("sendTransaction", params)
	if err != nil {
		return "", fmt.Errorf("sendTransaction failed: %w", err)
	}

	// Result should be the transaction signature as a string
	signature, ok := result.(string)
	if !ok {
		return "", fmt.Errorf("unexpected result type: %T", result)
	}

	return signature, nil
}

// GetTransactionStatus retrieves the confirmation status of a transaction
func (c *RPCClient) GetTransactionStatus(signature string) (*rpc.TransactionStatus, error) {
	params := map[string]string{
		"signature": signature,
	}

	result, err := c.call("getTransactionStatus", params)
	if err != nil {
		return nil, fmt.Errorf("getTransactionStatus failed: %w", err)
	}

	// Convert result to TransactionStatus
	resultBytes, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result: %w", err)
	}

	var txStatus rpc.TransactionStatus
	if err := json.Unmarshal(resultBytes, &txStatus); err != nil {
		return nil, fmt.Errorf("failed to unmarshal transaction status: %w", err)
	}

	return &txStatus, nil
}

// GetBlockHeight retrieves the current blockchain height
func (c *RPCClient) GetBlockHeight() (int64, error) {
	result, err := c.call("getBlockHeight", nil)
	if err != nil {
		return 0, fmt.Errorf("getBlockHeight failed: %w", err)
	}

	// JSON numbers are decoded as float64
	height, ok := result.(float64)
	if !ok {
		return 0, fmt.Errorf("unexpected result type: %T", result)
	}

	return int64(height), nil
}
