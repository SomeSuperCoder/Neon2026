package rpc

import (
	"encoding/json"
	"time"
)

// JSONRPCRequest represents a JSON-RPC 2.0 request
type JSONRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params"`
	ID      interface{}     `json:"id"`
}

// JSONRPCResponse represents a JSON-RPC 2.0 response
type JSONRPCResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	Result  interface{} `json:"result,omitempty"`
	Error   *RPCError   `json:"error,omitempty"`
	ID      interface{} `json:"id"`
}

// RPCError represents a JSON-RPC 2.0 error
type RPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// JSON-RPC 2.0 standard error codes
const (
	ParseError     = -32700 // Invalid JSON
	InvalidRequest = -32600 // Invalid Request object
	MethodNotFound = -32601 // Method does not exist
	InvalidParams  = -32602 // Invalid method parameters
	InternalError  = -32603 // Internal JSON-RPC error
)

// Custom application error codes
const (
	InvalidSignature     = -32001 // Transaction signature invalid
	MalformedTransaction = -32002 // Transaction format invalid
	InsufficientBalance  = -32003 // Account has insufficient balance
	AccountNotFound      = -32004 // Account does not exist
	TransactionNotFound  = -32005 // Transaction not found
	NetworkError         = -32006 // Network communication error
)

// AccountInfo represents full account details
type AccountInfo struct {
	Address    string `json:"address"`
	Balance    int64  `json:"balance"`
	Owner      string `json:"owner"`
	DataLength int    `json:"dataLength"`
	Executable bool   `json:"executable"`
}

// TransactionStatus represents transaction confirmation status
type TransactionStatus struct {
	Signature   string `json:"signature"`
	Confirmed   bool   `json:"confirmed"`
	BlockHeight int64  `json:"blockHeight,omitempty"`
	Slot        int64  `json:"slot,omitempty"`
	Error       string `json:"error,omitempty"`
}

// TransactionRecord represents a transaction with full details
type TransactionRecord struct {
	Signature    string              `json:"signature"`
	BlockHeight  int64               `json:"blockHeight"`
	Slot         int64               `json:"slot"`
	Timestamp    time.Time           `json:"timestamp"`
	Success      bool                `json:"success"`
	Error        string              `json:"error,omitempty"`
	Instructions []InstructionRecord `json:"instructions"`
}

// InstructionRecord represents an instruction within a transaction
type InstructionRecord struct {
	ProgramID string   `json:"programId"`
	Type      string   `json:"type"`
	Accounts  []string `json:"accounts"`
	Data      string   `json:"data"`
}
