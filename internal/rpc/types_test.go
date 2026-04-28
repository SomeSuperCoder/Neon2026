package rpc

import (
	"encoding/json"
	"testing"
	"time"
)

func TestJSONRPCRequest_Marshal(t *testing.T) {
	tests := []struct {
		name    string
		request JSONRPCRequest
		want    string
	}{
		{
			name: "simple request with string params",
			request: JSONRPCRequest{
				JSONRPC: "2.0",
				Method:  "getBalance",
				Params:  json.RawMessage(`{"address":"abc123"}`),
				ID:      1,
			},
			want: `{"jsonrpc":"2.0","method":"getBalance","params":{"address":"abc123"},"id":1}`,
		},
		{
			name: "request with null ID",
			request: JSONRPCRequest{
				JSONRPC: "2.0",
				Method:  "getBlockHeight",
				Params:  json.RawMessage(`{}`),
				ID:      nil,
			},
			want: `{"jsonrpc":"2.0","method":"getBlockHeight","params":{},"id":null}`,
		},
		{
			name: "request with string ID",
			request: JSONRPCRequest{
				JSONRPC: "2.0",
				Method:  "sendTransaction",
				Params:  json.RawMessage(`{"transaction":"0x123"}`),
				ID:      "request-1",
			},
			want: `{"jsonrpc":"2.0","method":"sendTransaction","params":{"transaction":"0x123"},"id":"request-1"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := json.Marshal(tt.request)
			if err != nil {
				t.Fatalf("Marshal() error = %v", err)
			}
			if string(got) != tt.want {
				t.Errorf("Marshal() = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestJSONRPCRequest_Unmarshal(t *testing.T) {
	tests := []struct {
		name    string
		json    string
		want    JSONRPCRequest
		wantErr bool
	}{
		{
			name: "valid request",
			json: `{"jsonrpc":"2.0","method":"getBalance","params":{"address":"abc123"},"id":1}`,
			want: JSONRPCRequest{
				JSONRPC: "2.0",
				Method:  "getBalance",
				Params:  json.RawMessage(`{"address":"abc123"}`),
				ID:      float64(1), // JSON numbers unmarshal to float64
			},
		},
		{
			name: "request with empty params",
			json: `{"jsonrpc":"2.0","method":"getBlockHeight","params":{},"id":2}`,
			want: JSONRPCRequest{
				JSONRPC: "2.0",
				Method:  "getBlockHeight",
				Params:  json.RawMessage(`{}`),
				ID:      float64(2),
			},
		},
		{
			name:    "invalid json",
			json:    `{"jsonrpc":"2.0","method":}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got JSONRPCRequest
			err := json.Unmarshal([]byte(tt.json), &got)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Unmarshal() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if got.JSONRPC != tt.want.JSONRPC {
				t.Errorf("JSONRPC = %v, want %v", got.JSONRPC, tt.want.JSONRPC)
			}
			if got.Method != tt.want.Method {
				t.Errorf("Method = %v, want %v", got.Method, tt.want.Method)
			}
			if string(got.Params) != string(tt.want.Params) {
				t.Errorf("Params = %v, want %v", string(got.Params), string(tt.want.Params))
			}
		})
	}
}

func TestJSONRPCResponse_Marshal(t *testing.T) {
	tests := []struct {
		name     string
		response JSONRPCResponse
		want     string
	}{
		{
			name: "success response",
			response: JSONRPCResponse{
				JSONRPC: "2.0",
				Result:  map[string]interface{}{"balance": 1000},
				ID:      1,
			},
			want: `{"jsonrpc":"2.0","result":{"balance":1000},"id":1}`,
		},
		{
			name: "error response",
			response: JSONRPCResponse{
				JSONRPC: "2.0",
				Error: &RPCError{
					Code:    InvalidParams,
					Message: "Invalid address",
				},
				ID: 1,
			},
			want: `{"jsonrpc":"2.0","error":{"code":-32602,"message":"Invalid address"},"id":1}`,
		},
		{
			name: "error response with data",
			response: JSONRPCResponse{
				JSONRPC: "2.0",
				Error: &RPCError{
					Code:    InternalError,
					Message: "Internal error",
					Data:    "Database connection failed",
				},
				ID: 1,
			},
			want: `{"jsonrpc":"2.0","error":{"code":-32603,"message":"Internal error","data":"Database connection failed"},"id":1}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := json.Marshal(tt.response)
			if err != nil {
				t.Fatalf("Marshal() error = %v", err)
			}
			if string(got) != tt.want {
				t.Errorf("Marshal() = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestRPCError_ErrorCodes(t *testing.T) {
	tests := []struct {
		name string
		code int
		want int
	}{
		{"ParseError", ParseError, -32700},
		{"InvalidRequest", InvalidRequest, -32600},
		{"MethodNotFound", MethodNotFound, -32601},
		{"InvalidParams", InvalidParams, -32602},
		{"InternalError", InternalError, -32603},
		{"InvalidSignature", InvalidSignature, -32001},
		{"MalformedTransaction", MalformedTransaction, -32002},
		{"InsufficientBalance", InsufficientBalance, -32003},
		{"AccountNotFound", AccountNotFound, -32004},
		{"TransactionNotFound", TransactionNotFound, -32005},
		{"NetworkError", NetworkError, -32006},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.code != tt.want {
				t.Errorf("Error code %s = %d, want %d", tt.name, tt.code, tt.want)
			}
		})
	}
}

func TestAccountInfo_Marshal(t *testing.T) {
	info := AccountInfo{
		Address:    "abc123",
		Balance:    1000000000,
		Owner:      "System_Program",
		DataLength: 0,
		Executable: false,
	}

	got, err := json.Marshal(info)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	want := `{"address":"abc123","balance":1000000000,"owner":"System_Program","dataLength":0,"executable":false}`
	if string(got) != want {
		t.Errorf("Marshal() = %s, want %s", got, want)
	}
}

func TestTransactionStatus_Marshal(t *testing.T) {
	tests := []struct {
		name   string
		status TransactionStatus
		want   string
	}{
		{
			name: "confirmed transaction",
			status: TransactionStatus{
				Signature:   "tx123",
				Confirmed:   true,
				BlockHeight: 12345,
				Slot:        30863,
			},
			want: `{"signature":"tx123","confirmed":true,"blockHeight":12345,"slot":30863}`,
		},
		{
			name: "pending transaction",
			status: TransactionStatus{
				Signature: "tx456",
				Confirmed: false,
			},
			want: `{"signature":"tx456","confirmed":false}`,
		},
		{
			name: "failed transaction",
			status: TransactionStatus{
				Signature:   "tx789",
				Confirmed:   true,
				BlockHeight: 12346,
				Slot:        30864,
				Error:       "Insufficient balance",
			},
			want: `{"signature":"tx789","confirmed":true,"blockHeight":12346,"slot":30864,"error":"Insufficient balance"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := json.Marshal(tt.status)
			if err != nil {
				t.Fatalf("Marshal() error = %v", err)
			}
			if string(got) != tt.want {
				t.Errorf("Marshal() = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestTransactionRecord_Marshal(t *testing.T) {
	timestamp := time.Date(2026, 4, 27, 16, 30, 0, 0, time.UTC)

	record := TransactionRecord{
		Signature:   "tx123abc",
		BlockHeight: 12345,
		Slot:        30863,
		Timestamp:   timestamp,
		Success:     true,
		Instructions: []InstructionRecord{
			{
				ProgramID: "System_Program",
				Type:      "Transfer",
				Accounts:  []string{"a1b2c3d4", "e5f6g7h8"},
				Data:      "0100000000000000",
			},
		},
	}

	got, err := json.Marshal(record)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	// Verify it contains expected fields
	var unmarshaled map[string]interface{}
	if err := json.Unmarshal(got, &unmarshaled); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if unmarshaled["signature"] != "tx123abc" {
		t.Errorf("signature = %v, want tx123abc", unmarshaled["signature"])
	}
	if unmarshaled["blockHeight"] != float64(12345) {
		t.Errorf("blockHeight = %v, want 12345", unmarshaled["blockHeight"])
	}
	if unmarshaled["success"] != true {
		t.Errorf("success = %v, want true", unmarshaled["success"])
	}
}

func TestInstructionRecord_Marshal(t *testing.T) {
	instruction := InstructionRecord{
		ProgramID: "System_Program",
		Type:      "Transfer",
		Accounts:  []string{"sender", "recipient"},
		Data:      "0x123456",
	}

	got, err := json.Marshal(instruction)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	want := `{"programId":"System_Program","type":"Transfer","accounts":["sender","recipient"],"data":"0x123456"}`
	if string(got) != want {
		t.Errorf("Marshal() = %s, want %s", got, want)
	}
}
