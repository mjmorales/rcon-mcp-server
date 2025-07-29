// Package rcon implements the RCON (Remote Console) protocol client.
// It provides functionality to connect to RCON servers, authenticate, and execute commands.
package rcon

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

// PacketType represents the type of RCON packet as defined by the Source RCON protocol.
type PacketType int32

// RCON packet types as defined by the Source RCON protocol specification.
const (
	PacketTypeAuth         PacketType = 3 // Authentication request
	PacketTypeAuthResponse PacketType = 2 // Authentication response
	PacketTypeCommand      PacketType = 2 // Command execution request
	PacketTypeResponse     PacketType = 0 // Command response
)

// Protocol constants for RCON communication.
const (
	maxPacketSize = 4096             // Maximum allowed packet size in bytes
	headerSize    = 12               // Packet header size: size(4) + id(4) + type(4)
	timeout       = 10 * time.Second // Default timeout for network operations
)

// Packet represents an RCON protocol packet.
// Each packet contains a size, request ID, type, and body payload.
type Packet struct {
	Size int32      // Total packet size in bytes (excluding the size field itself)
	ID   int32      // Request ID for matching responses to requests
	Type PacketType // Type of packet (auth, command, response)
	Body string     // Packet payload (password, command, or response text)
}

// Client manages an RCON connection to a server.
// It handles connection state, authentication, and command execution.
// All operations are thread-safe.
type Client struct {
	conn         net.Conn   // TCP connection to the RCON server
	mu           sync.Mutex // Mutex for thread-safe operations
	requestID    int32      // Counter for generating unique request IDs
	isConnected  bool       // Connection state flag
	isAuthorized bool       // Authentication state flag
}

// NewClient creates a new RCON client instance.
// The client is created in a disconnected state.
func NewClient() *Client {
	return &Client{
		requestID: 1,
	}
}

// Connect establishes a TCP connection to an RCON server.
// The address should be in the format "host:port".
// Returns an error if already connected or if the connection fails.
func (c *Client) Connect(address string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.isConnected {
		return errors.New("already connected")
	}

	conn, err := net.DialTimeout("tcp", address, timeout)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}

	c.conn = conn
	c.isConnected = true
	return nil
}

// Authenticate performs RCON authentication using the provided password.
// Must be called after Connect and before Execute.
// Returns an error if not connected, already authenticated, or if authentication fails.
func (c *Client) Authenticate(password string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.isConnected {
		return errors.New("not connected")
	}

	if c.isAuthorized {
		return errors.New("already authenticated")
	}

	// Send auth packet
	authPacket := &Packet{
		ID:   c.getNextRequestID(),
		Type: PacketTypeAuth,
		Body: password,
	}

	if err := c.sendPacket(authPacket); err != nil {
		return fmt.Errorf("failed to send auth packet: %w", err)
	}

	// Read auth response
	response, err := c.readPacket()
	if err != nil {
		return fmt.Errorf("failed to read auth response: %w", err)
	}

	// Check auth response
	if response.ID == -1 {
		return errors.New("authentication failed: invalid password")
	}

	if response.ID != authPacket.ID {
		return errors.New("authentication failed: unexpected response ID")
	}

	c.isAuthorized = true
	return nil
}

// Execute sends a command to the RCON server and returns the response.
// The client must be connected and authenticated before executing commands.
// Returns the server's response as a string, or an error if execution fails.
func (c *Client) Execute(command string) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.isConnected {
		return "", errors.New("not connected")
	}

	if !c.isAuthorized {
		return "", errors.New("not authenticated")
	}

	// Send command packet
	cmdPacket := &Packet{
		ID:   c.getNextRequestID(),
		Type: PacketTypeCommand,
		Body: command,
	}

	if err := c.sendPacket(cmdPacket); err != nil {
		return "", fmt.Errorf("failed to send command: %w", err)
	}

	// Read response
	response, err := c.readPacket()
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	// Verify response ID matches request
	if response.ID != cmdPacket.ID {
		return "", errors.New("response ID mismatch")
	}

	return response.Body, nil
}

// Disconnect closes the TCP connection to the RCON server.
// It's safe to call Disconnect multiple times or on an already disconnected client.
func (c *Client) Disconnect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.isConnected {
		return nil
	}

	if err := c.conn.Close(); err != nil {
		return fmt.Errorf("failed to close connection: %w", err)
	}

	c.conn = nil
	c.isConnected = false
	c.isAuthorized = false
	return nil
}

// IsConnected returns true if the client has an active connection to the server.
func (c *Client) IsConnected() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.isConnected
}

// IsAuthenticated returns true if the client has successfully authenticated with the server.
func (c *Client) IsAuthenticated() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.isAuthorized
}

// sendPacket encodes and sends a packet to the RCON server.
// It automatically calculates the packet size and adds null terminators.
func (c *Client) sendPacket(packet *Packet) error {
	// Calculate packet size
	bodyBytes := []byte(packet.Body)
	packet.Size = int32(len(bodyBytes) + 10) // body + ID(4) + Type(4) + null terminators(2)

	// Create packet buffer
	buf := new(bytes.Buffer)

	// Write packet fields
	if err := binary.Write(buf, binary.LittleEndian, packet.Size); err != nil {
		return fmt.Errorf("failed to write packet size: %w", err)
	}
	if err := binary.Write(buf, binary.LittleEndian, packet.ID); err != nil {
		return fmt.Errorf("failed to write packet ID: %w", err)
	}
	if err := binary.Write(buf, binary.LittleEndian, packet.Type); err != nil {
		return fmt.Errorf("failed to write packet type: %w", err)
	}
	buf.Write(bodyBytes)
	buf.WriteByte(0) // Body null terminator
	buf.WriteByte(0) // Packet null terminator

	// Send packet
	if err := c.conn.SetWriteDeadline(time.Now().Add(timeout)); err != nil {
		return fmt.Errorf("failed to set write deadline: %w", err)
	}
	if _, err := c.conn.Write(buf.Bytes()); err != nil {
		return err
	}

	return nil
}

// readPacket reads and decodes a packet from the RCON server.
// It validates packet size and parses the packet structure.
func (c *Client) readPacket() (*Packet, error) {
	if err := c.conn.SetReadDeadline(time.Now().Add(timeout)); err != nil {
		return nil, fmt.Errorf("failed to set read deadline: %w", err)
	}

	// Read packet size
	sizeBuf := make([]byte, 4)
	if _, err := io.ReadFull(c.conn, sizeBuf); err != nil {
		return nil, err
	}

	var size int32
	if err := binary.Read(bytes.NewReader(sizeBuf), binary.LittleEndian, &size); err != nil {
		return nil, err
	}

	if size < 10 || size > maxPacketSize {
		return nil, fmt.Errorf("invalid packet size: %d", size)
	}

	// Read rest of packet
	packetBuf := make([]byte, size)
	if _, err := io.ReadFull(c.conn, packetBuf); err != nil {
		return nil, err
	}

	// Parse packet
	packet := &Packet{}
	reader := bytes.NewReader(packetBuf)

	if err := binary.Read(reader, binary.LittleEndian, &packet.ID); err != nil {
		return nil, err
	}

	var pType int32
	if err := binary.Read(reader, binary.LittleEndian, &pType); err != nil {
		return nil, err
	}
	packet.Type = PacketType(pType)

	// Read body (everything except the last 2 null bytes)
	bodyBytes := packetBuf[8 : len(packetBuf)-2]
	packet.Body = string(bodyBytes)

	return packet, nil
}

// getNextRequestID generates a unique request ID for packet tracking.
// IDs are incremented sequentially for each request.
func (c *Client) getNextRequestID() int32 {
	id := c.requestID
	c.requestID++
	return id
}
