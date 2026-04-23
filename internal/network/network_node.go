package network

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"sync"

	"github.com/poh-blockchain/internal/blockchain"
)

// NodeType represents the type of node in the network
type NodeType int

const (
	LEADER NodeType = iota
	REPLICA
)

// NetworkNode handles peer-to-peer communication between nodes
type NetworkNode struct {
	host         string
	port         int
	nodeType     NodeType
	connections  []*net.Conn
	listener     net.Listener
	messageQueue chan []byte
	stopChan     chan struct{}
	mu           sync.RWMutex
}

// NewNetworkNode creates a new network node instance
func NewNetworkNode(host string, port int, nodeType NodeType) *NetworkNode {
	return &NetworkNode{
		host:         host,
		port:         port,
		nodeType:     nodeType,
		connections:  make([]*net.Conn, 0),
		messageQueue: make(chan []byte, 100),
		stopChan:     make(chan struct{}),
	}
}

// Start begins listening for incoming connections
func (nn *NetworkNode) Start() error {
	address := fmt.Sprintf("%s:%d", nn.host, nn.port)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return fmt.Errorf("failed to start listener: %w", err)
	}

	nn.mu.Lock()
	nn.listener = listener
	nn.mu.Unlock()

	// Start accepting connections in a goroutine
	go nn.acceptConnections()

	return nil
}

// acceptConnections handles incoming connection requests
func (nn *NetworkNode) acceptConnections() {
	for {
		select {
		case <-nn.stopChan:
			return
		default:
			conn, err := nn.listener.Accept()
			if err != nil {
				select {
				case <-nn.stopChan:
					return
				default:
					continue
				}
			}

			nn.mu.Lock()
			nn.connections = append(nn.connections, &conn)
			nn.mu.Unlock()

			// Handle this connection in a separate goroutine
			go nn.handleConnection(&conn)
		}
	}
}

// handleConnection processes messages from a single connection
func (nn *NetworkNode) handleConnection(conn *net.Conn) {
	defer func() {
		(*conn).Close()
		nn.removeConnection(conn)
	}()

	for {
		select {
		case <-nn.stopChan:
			return
		default:
			// Read message length prefix (4 bytes)
			lengthBuf := make([]byte, 4)
			_, err := io.ReadFull(*conn, lengthBuf)
			if err != nil {
				return
			}

			messageLength := binary.BigEndian.Uint32(lengthBuf)

			// Read message data
			messageBuf := make([]byte, messageLength)
			_, err = io.ReadFull(*conn, messageBuf)
			if err != nil {
				return
			}

			// Send to message queue
			select {
			case nn.messageQueue <- messageBuf:
			case <-nn.stopChan:
				return
			}
		}
	}
}

// ConnectToPeer establishes a connection to another node
func (nn *NetworkNode) ConnectToPeer(peerAddress string) error {
	conn, err := net.Dial("tcp", peerAddress)
	if err != nil {
		return fmt.Errorf("failed to connect to peer %s: %w", peerAddress, err)
	}

	nn.mu.Lock()
	nn.connections = append(nn.connections, &conn)
	nn.mu.Unlock()

	// Handle incoming messages from this peer
	go nn.handleConnection(&conn)

	return nil
}

// removeConnection removes a connection from the connections list
func (nn *NetworkNode) removeConnection(conn *net.Conn) {
	nn.mu.Lock()
	defer nn.mu.Unlock()

	for i, c := range nn.connections {
		if c == conn {
			nn.connections = append(nn.connections[:i], nn.connections[i+1:]...)
			break
		}
	}
}

// Stop gracefully shuts down the network node
func (nn *NetworkNode) Stop() {
	close(nn.stopChan)

	nn.mu.Lock()
	defer nn.mu.Unlock()

	// Close listener
	if nn.listener != nil {
		nn.listener.Close()
	}

	// Close all connections
	for _, conn := range nn.connections {
		(*conn).Close()
	}

	nn.connections = nil
}

// SerializeBlock converts a block to JSON bytes for network transmission
func (nn *NetworkNode) SerializeBlock(block blockchain.Block) ([]byte, error) {
	data, err := json.Marshal(block)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize block: %w", err)
	}
	return data, nil
}

// DeserializeBlock reconstructs a block from JSON bytes
func (nn *NetworkNode) DeserializeBlock(data []byte) (blockchain.Block, error) {
	var block blockchain.Block
	err := json.Unmarshal(data, &block)
	if err != nil {
		return blockchain.Block{}, fmt.Errorf("failed to deserialize block: %w", err)
	}
	return block, nil
}

// BroadcastBlock sends a block to all connected peers
func (nn *NetworkNode) BroadcastBlock(block blockchain.Block) error {
	// Serialize the block
	data, err := nn.SerializeBlock(block)
	if err != nil {
		return err
	}

	// Create message with length prefix
	messageLength := uint32(len(data))
	lengthBuf := make([]byte, 4)
	binary.BigEndian.PutUint32(lengthBuf, messageLength)

	// Combine length prefix and data
	message := append(lengthBuf, data...)

	nn.mu.RLock()
	connections := make([]*net.Conn, len(nn.connections))
	copy(connections, nn.connections)
	nn.mu.RUnlock()

	// Send to all connected peers
	var lastErr error
	for _, conn := range connections {
		_, err := (*conn).Write(message)
		if err != nil {
			lastErr = err
			// Continue trying to send to other peers even if one fails
		}
	}

	return lastErr
}

// ReceiveBlock listens on the message queue and returns received blocks
func (nn *NetworkNode) ReceiveBlock() (blockchain.Block, error) {
	select {
	case data := <-nn.messageQueue:
		return nn.DeserializeBlock(data)
	case <-nn.stopChan:
		return blockchain.Block{}, fmt.Errorf("network node stopped")
	}
}
