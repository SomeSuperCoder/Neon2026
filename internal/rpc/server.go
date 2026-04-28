package rpc

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/poh-blockchain/internal/filestore"
	"github.com/poh-blockchain/internal/processor"
	"github.com/poh-blockchain/internal/storage"
)

// RPCServer represents the JSON-RPC HTTP server
type RPCServer struct {
	config       *ServerConfig
	httpServer   *http.Server
	handler      *RPCHandler
	logger       *log.Logger
	listener     net.Listener
	shutdownOnce sync.Once
	wg           sync.WaitGroup
}

// NewRPCServer creates a new RPC server instance
func NewRPCServer(
	config *ServerConfig,
	ledger *storage.Ledger,
	fileStore *filestore.FileStore,
	txProcessor *processor.TxProcessor,
	logger *log.Logger,
) (*RPCServer, error) {
	if config == nil {
		config = DefaultServerConfig()
	}

	if logger == nil {
		logger = log.Default()
	}

	// Create query engine
	queryEngine := NewQueryEngine(ledger, fileStore)

	// Create RPC handler
	handler := NewRPCHandler(queryEngine, txProcessor, logger)

	server := &RPCServer{
		config:  config,
		handler: handler,
		logger:  logger,
	}

	// Create HTTP server
	server.httpServer = &http.Server{
		Addr:         fmt.Sprintf("%s:%d", config.BindAddress, config.Port),
		Handler:      server,
		ReadTimeout:  config.ReadTimeout,
		WriteTimeout: config.WriteTimeout,
	}

	return server, nil
}

// Start starts the RPC server
func (s *RPCServer) Start() error {
	var err error
	s.listener, err = net.Listen("tcp", s.httpServer.Addr)
	if err != nil {
		return fmt.Errorf("failed to create listener: %w", err)
	}

	s.logger.Printf("RPC server listening on %s", s.httpServer.Addr)

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		if err := s.httpServer.Serve(s.listener); err != nil && err != http.ErrServerClosed {
			s.logger.Printf("RPC server error: %v", err)
		}
	}()

	return nil
}

// Stop gracefully stops the RPC server
func (s *RPCServer) Stop() error {
	var stopErr error
	s.shutdownOnce.Do(func() {
		s.logger.Println("Shutting down RPC server...")

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := s.httpServer.Shutdown(ctx); err != nil {
			stopErr = fmt.Errorf("server shutdown error: %w", err)
		}

		s.wg.Wait()
		s.logger.Println("RPC server stopped")
	})

	return stopErr
}

// ServeHTTP implements http.Handler interface
func (s *RPCServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers for browser access
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// Handle preflight requests
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Only accept POST requests
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Validate request size (1MB limit)
	const maxRequestSize = 1 << 20 // 1MB
	r.Body = http.MaxBytesReader(w, r.Body, maxRequestSize)

	// Set content type
	w.Header().Set("Content-Type", "application/json")

	// Delegate to handler
	s.handler.ServeHTTP(w, r)
}

// Address returns the server's listening address
func (s *RPCServer) Address() string {
	if s.listener != nil {
		return s.listener.Addr().String()
	}
	return s.httpServer.Addr
}
