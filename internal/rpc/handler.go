package rpc

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/poh-blockchain/internal/processor"
	"github.com/poh-blockchain/internal/transaction"
)

// MethodHandler is a function that handles a specific RPC method
type MethodHandler func(*RPCHandler, *JSONRPCRequest) *JSONRPCResponse

// RPCHandler handles JSON-RPC requests
type RPCHandler struct {
	queryEngine    *QueryEngine
	txProcessor    *processor.TxProcessor
	logger         *log.Logger
	methodRegistry map[string]MethodHandler
}

// NewRPCHandler creates a new RPC handler
func NewRPCHandler(
	queryEngine *QueryEngine,
	txProcessor *processor.TxProcessor,
	logger *log.Logger,
) *RPCHandler {
	if logger == nil {
		logger = log.Default()
	}

	handler := &RPCHandler{
		queryEngine:    queryEngine,
		txProcessor:    txProcessor,
		logger:         logger,
		methodRegistry: make(map[string]MethodHandler),
	}

	// Register all RPC methods
	handler.registerMethods()

	return handler
}

// registerMethods registers all available RPC methods
func (h *RPCHandler) registerMethods() {
	h.methodRegistry["getBalance"] = (*RPCHandler).handleGetBalance
	h.methodRegistry["getAccountInfo"] = (*RPCHandler).handleGetAccountInfo
	h.methodRegistry["getTransactionHistory"] = (*RPCHandler).handleGetTransactionHistory
	h.methodRegistry["getBlockHeight"] = (*RPCHandler).handleGetBlockHeight
	h.methodRegistry["getRecentBlockhash"] = (*RPCHandler).handleGetRecentBlockhash
	h.methodRegistry["sendTransaction"] = (*RPCHandler).handleSendTransaction
	h.methodRegistry["getTransactionStatus"] = (*RPCHandler).handleGetTransactionStatus
}

// ServeHTTP handles HTTP requests
func (h *RPCHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	// Set content type for JSON-RPC
	w.Header().Set("Content-Type", "application/json")

	// Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.logger.Printf("Error reading request body: %v", err)
		h.writeError(w, nil, ParseError, "Failed to read request body")
		return
	}
	defer r.Body.Close()

	// Parse JSON-RPC request
	var req JSONRPCRequest
	if err := json.Unmarshal(body, &req); err != nil {
		h.logger.Printf("Error parsing JSON-RPC request: %v", err)
		h.writeError(w, nil, ParseError, "Invalid JSON")
		return
	}

	// Validate JSON-RPC version
	if req.JSONRPC != "2.0" {
		h.logger.Printf("Invalid JSON-RPC version: %s", req.JSONRPC)
		h.writeError(w, req.ID, InvalidRequest, "Invalid JSON-RPC version")
		return
	}

	// Validate method is not empty
	if req.Method == "" {
		h.logger.Printf("Empty method name in request")
		h.writeError(w, req.ID, InvalidRequest, "Method name is required")
		return
	}

	// Handle request
	response := h.HandleRequest(&req)

	// Log request with timing
	duration := time.Since(startTime)
	status := "success"
	if response.Error != nil {
		status = "error"
	}
	h.logger.Printf("RPC %s completed in %v (status: %s)", req.Method, duration, status)

	// Write response
	responseBytes, err := json.Marshal(response)
	if err != nil {
		h.logger.Printf("Error marshaling response: %v", err)
		h.writeError(w, req.ID, InternalError, "Failed to marshal response")
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(responseBytes)
}

// HandleRequest processes a JSON-RPC request and returns a response
func (h *RPCHandler) HandleRequest(req *JSONRPCRequest) *JSONRPCResponse {
	// Validate request
	if err := h.validateRequest(req); err != nil {
		return h.errorResponse(req.ID, err.Code, err.Message)
	}

	// Look up method handler in registry
	handler, exists := h.methodRegistry[req.Method]
	if !exists {
		h.logger.Printf("Method not found: %s", req.Method)
		return h.errorResponse(req.ID, MethodNotFound, "Method not found")
	}

	// Execute method handler
	return handler(h, req)
}

// validateRequest validates a JSON-RPC request
func (h *RPCHandler) validateRequest(req *JSONRPCRequest) *RPCError {
	// Check JSON-RPC version
	if req.JSONRPC != "2.0" {
		return &RPCError{
			Code:    InvalidRequest,
			Message: "Invalid JSON-RPC version",
		}
	}

	// Check method name
	if req.Method == "" {
		return &RPCError{
			Code:    InvalidRequest,
			Message: "Method name is required",
		}
	}

	return nil
}

// errorResponse creates an error response
func (h *RPCHandler) errorResponse(id interface{}, code int, message string) *JSONRPCResponse {
	return &JSONRPCResponse{
		JSONRPC: "2.0",
		Error: &RPCError{
			Code:    code,
			Message: message,
		},
		ID: id,
	}
}

// successResponse creates a success response
func (h *RPCHandler) successResponse(id interface{}, result interface{}) *JSONRPCResponse {
	return &JSONRPCResponse{
		JSONRPC: "2.0",
		Result:  result,
		ID:      id,
	}
}

// Method handlers (stubs for now, will be implemented in subsequent tasks)

func (h *RPCHandler) handleGetBalance(req *JSONRPCRequest) *JSONRPCResponse {
	// TODO: Implement in task 3.2
	return h.errorResponse(req.ID, InternalError, "Not implemented")
}

func (h *RPCHandler) handleGetAccountInfo(req *JSONRPCRequest) *JSONRPCResponse {
	// TODO: Implement in task 3.2
	return h.errorResponse(req.ID, InternalError, "Not implemented")
}

func (h *RPCHandler) handleGetTransactionHistory(req *JSONRPCRequest) *JSONRPCResponse {
	// TODO: Implement in task 3.4
	return h.errorResponse(req.ID, InternalError, "Not implemented")
}

func (h *RPCHandler) handleGetBlockHeight(req *JSONRPCRequest) *JSONRPCResponse {
	// TODO: Implement in task 3.3
	return h.errorResponse(req.ID, InternalError, "Not implemented")
}

func (h *RPCHandler) handleGetRecentBlockhash(req *JSONRPCRequest) *JSONRPCResponse {
	// TODO: Implement in task 3.3
	return h.errorResponse(req.ID, InternalError, "Not implemented")
}

func (h *RPCHandler) handleSendTransaction(req *JSONRPCRequest) *JSONRPCResponse {
	// Parse parameters
	var params struct {
		Transaction string `json:"transaction"`
	}

	if err := json.Unmarshal(req.Params, &params); err != nil {
		return h.errorResponse(req.ID, InvalidParams, "Invalid parameters")
	}

	// Validate transaction parameter is provided
	if params.Transaction == "" {
		return h.errorResponse(req.ID, InvalidParams, "transaction parameter is required")
	}

	// Decode base64-encoded transaction
	txBytes, err := base64.StdEncoding.DecodeString(params.Transaction)
	if err != nil {
		return h.errorResponse(req.ID, MalformedTransaction, "failed to decode transaction: "+err.Error())
	}

	// Unmarshal transaction
	tx, err := transaction.UnmarshalTransaction(txBytes)
	if err != nil {
		return h.errorResponse(req.ID, MalformedTransaction, "failed to unmarshal transaction: "+err.Error())
	}

	// Verify transaction signatures
	if err := h.verifyTransactionSignatures(tx); err != nil {
		return h.errorResponse(req.ID, InvalidSignature, "signature verification failed: "+err.Error())
	}

	// Check if processor is available
	if h.txProcessor == nil {
		return h.errorResponse(req.ID, InternalError, "transaction processor not available")
	}

	// Submit transaction to processor
	result, err := h.txProcessor.ProcessTransaction(tx)
	if err != nil {
		h.logger.Printf("Transaction processing failed: %v", err)
		return h.errorResponse(req.ID, InternalError, "transaction processing failed: "+err.Error())
	}

	// Check if transaction was successful
	if !result.Success {
		errMsg := "transaction failed"
		if result.Error != nil {
			errMsg = result.Error.Error()
		}
		return h.errorResponse(req.ID, InternalError, errMsg)
	}

	// Return transaction signature
	return h.successResponse(req.ID, result.TxID.String())
}

// verifyTransactionSignatures verifies all signatures in a transaction
func (h *RPCHandler) verifyTransactionSignatures(tx *transaction.Transaction) error {
	if len(tx.Signatures) == 0 {
		return fmt.Errorf("transaction must have at least one signature")
	}

	// Temporarily remove signatures for verification
	savedSignatures := tx.Signatures
	tx.Signatures = []transaction.Signature{}

	txData, err := tx.Marshal()
	if err != nil {
		tx.Signatures = savedSignatures
		return fmt.Errorf("failed to marshal transaction for verification: %w", err)
	}

	// Restore signatures
	tx.Signatures = savedSignatures

	// Verify each signature
	for i, sig := range tx.Signatures {
		if !sig.Verify(txData) {
			return fmt.Errorf("invalid signature at index %d", i)
		}
	}

	return nil
}

func (h *RPCHandler) handleGetTransactionStatus(req *JSONRPCRequest) *JSONRPCResponse {
	// TODO: Implement in task 4.2
	return h.errorResponse(req.ID, InternalError, "Not implemented")
}

// writeError writes a JSON-RPC error response
func (h *RPCHandler) writeError(w http.ResponseWriter, id interface{}, code int, message string) {
	response := &JSONRPCResponse{
		JSONRPC: "2.0",
		Error: &RPCError{
			Code:    code,
			Message: message,
		},
		ID: id,
	}

	responseBytes, _ := json.Marshal(response)
	w.WriteHeader(http.StatusOK)
	w.Write(responseBytes)
}
