package rcon

import (
	"bytes"
	"encoding/binary"
	"net"
	"testing"
	"time"
)

// mockConn implements net.Conn for testing
type mockConn struct {
	readBuf  *bytes.Buffer
	writeBuf *bytes.Buffer
	closed   bool
	readErr  error
	writeErr error
}

func newMockConn() *mockConn {
	return &mockConn{
		readBuf:  &bytes.Buffer{},
		writeBuf: &bytes.Buffer{},
	}
}

func (m *mockConn) Read(b []byte) (n int, err error) {
	if m.readErr != nil {
		return 0, m.readErr
	}
	return m.readBuf.Read(b)
}

func (m *mockConn) Write(b []byte) (n int, err error) {
	if m.writeErr != nil {
		return 0, m.writeErr
	}
	return m.writeBuf.Write(b)
}

func (m *mockConn) Close() error {
	m.closed = true
	return nil
}

func (m *mockConn) LocalAddr() net.Addr                { return nil }
func (m *mockConn) RemoteAddr() net.Addr               { return nil }
func (m *mockConn) SetDeadline(t time.Time) error      { return nil }
func (m *mockConn) SetReadDeadline(t time.Time) error  { return nil }
func (m *mockConn) SetWriteDeadline(t time.Time) error { return nil }

// writePacketToBuffer writes a packet to the buffer for mock reading
func writePacketToBuffer(buf *bytes.Buffer, packet *Packet) error {
	bodyBytes := []byte(packet.Body)
	packet.Size = int32(len(bodyBytes) + 10)

	if err := binary.Write(buf, binary.LittleEndian, packet.Size); err != nil {
		return err
	}
	if err := binary.Write(buf, binary.LittleEndian, packet.ID); err != nil {
		return err
	}
	if err := binary.Write(buf, binary.LittleEndian, packet.Type); err != nil {
		return err
	}
	buf.Write(bodyBytes)
	buf.WriteByte(0)
	buf.WriteByte(0)
	return nil
}

func TestNewClient(t *testing.T) {
	client := NewClient()
	if client == nil {
		t.Fatal("NewClient returned nil")
	}
	if client.requestID != 1 {
		t.Errorf("Expected requestID to be 1, got %d", client.requestID)
	}
	if client.isConnected {
		t.Error("Expected isConnected to be false")
	}
	if client.isAuthorized {
		t.Error("Expected isAuthorized to be false")
	}
}

func TestClient_Connect(t *testing.T) {
	tests := []struct {
		name          string
		address       string
		alreadyConn   bool
		wantErr       bool
		errContains   string
	}{
		{
			name:    "successful connection",
			address: "localhost:25575",
			wantErr: false,
		},
		{
			name:        "already connected",
			address:     "localhost:25575",
			alreadyConn: true,
			wantErr:     true,
			errContains: "already connected",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClient()
			if tt.alreadyConn {
				client.isConnected = true
			}

			// For this test, we'll just check the state changes
			// In a real scenario, we'd need to mock net.Dial
			err := client.Connect(tt.address)
			
			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got nil")
				} else if tt.errContains != "" && !contains(err.Error(), tt.errContains) {
					t.Errorf("Expected error containing %q, got %q", tt.errContains, err.Error())
				}
			} else {
				// Since we can't actually connect in tests, we expect an error
				if err == nil {
					t.Error("Expected connection error in test environment")
				}
			}
		})
	}
}

func TestClient_Authenticate(t *testing.T) {
	tests := []struct {
		name        string
		password    string
		setup       func(*Client, *mockConn)
		wantErr     bool
		errContains string
	}{
		{
			name:     "successful authentication",
			password: "testpass",
			setup: func(c *Client, mc *mockConn) {
				c.isConnected = true
				c.conn = mc
				// Write auth response with matching ID
				writePacketToBuffer(mc.readBuf, &Packet{
					ID:   2, // Will match the request ID
					Type: PacketTypeAuthResponse,
					Body: "",
				})
			},
			wantErr: false,
		},
		{
			name:     "not connected",
			password: "testpass",
			setup: func(c *Client, mc *mockConn) {
				// Leave disconnected
			},
			wantErr:     true,
			errContains: "not connected",
		},
		{
			name:     "already authenticated",
			password: "testpass",
			setup: func(c *Client, mc *mockConn) {
				c.isConnected = true
				c.isAuthorized = true
			},
			wantErr:     true,
			errContains: "already authenticated",
		},
		{
			name:     "invalid password",
			password: "badpass",
			setup: func(c *Client, mc *mockConn) {
				c.isConnected = true
				c.conn = mc
				// Write auth response with ID -1 (auth failure)
				writePacketToBuffer(mc.readBuf, &Packet{
					ID:   -1,
					Type: PacketTypeAuthResponse,
					Body: "",
				})
			},
			wantErr:     true,
			errContains: "invalid password",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClient()
			client.requestID = 2 // Set known request ID for testing
			mc := newMockConn()
			
			tt.setup(client, mc)
			
			err := client.Authenticate(tt.password)
			
			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got nil")
				} else if tt.errContains != "" && !contains(err.Error(), tt.errContains) {
					t.Errorf("Expected error containing %q, got %q", tt.errContains, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
				if !client.isAuthorized {
					t.Error("Expected client to be authorized")
				}
			}
		})
	}
}

func TestClient_Execute(t *testing.T) {
	tests := []struct {
		name        string
		command     string
		setup       func(*Client, *mockConn)
		want        string
		wantErr     bool
		errContains string
	}{
		{
			name:    "successful command execution",
			command: "list",
			setup: func(c *Client, mc *mockConn) {
				c.isConnected = true
				c.isAuthorized = true
				c.conn = mc
				// Write command response
				writePacketToBuffer(mc.readBuf, &Packet{
					ID:   2, // Will match the request ID
					Type: PacketTypeResponse,
					Body: "Player1\nPlayer2\nPlayer3",
				})
			},
			want:    "Player1\nPlayer2\nPlayer3",
			wantErr: false,
		},
		{
			name:    "not connected",
			command: "list",
			setup: func(c *Client, mc *mockConn) {
				// Leave disconnected
			},
			wantErr:     true,
			errContains: "not connected",
		},
		{
			name:    "not authenticated",
			command: "list",
			setup: func(c *Client, mc *mockConn) {
				c.isConnected = true
				// Leave unauthorized
			},
			wantErr:     true,
			errContains: "not authenticated",
		},
		{
			name:    "response ID mismatch",
			command: "list",
			setup: func(c *Client, mc *mockConn) {
				c.isConnected = true
				c.isAuthorized = true
				c.conn = mc
				// Write response with wrong ID
				writePacketToBuffer(mc.readBuf, &Packet{
					ID:   99, // Wrong ID
					Type: PacketTypeResponse,
					Body: "data",
				})
			},
			wantErr:     true,
			errContains: "response ID mismatch",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClient()
			client.requestID = 2 // Set known request ID for testing
			mc := newMockConn()
			
			tt.setup(client, mc)
			
			got, err := client.Execute(tt.command)
			
			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got nil")
				} else if tt.errContains != "" && !contains(err.Error(), tt.errContains) {
					t.Errorf("Expected error containing %q, got %q", tt.errContains, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
				if got != tt.want {
					t.Errorf("Expected response %q, got %q", tt.want, got)
				}
			}
		})
	}
}

func TestClient_Disconnect(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*Client, *mockConn)
		wantErr bool
	}{
		{
			name: "successful disconnect",
			setup: func(c *Client, mc *mockConn) {
				c.isConnected = true
				c.isAuthorized = true
				c.conn = mc
			},
			wantErr: false,
		},
		{
			name: "already disconnected",
			setup: func(c *Client, mc *mockConn) {
				// Leave disconnected
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClient()
			mc := newMockConn()
			
			tt.setup(client, mc)
			
			err := client.Disconnect()
			
			if tt.wantErr && err == nil {
				t.Error("Expected error but got nil")
			} else if !tt.wantErr && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
			
			if client.isConnected {
				t.Error("Expected client to be disconnected")
			}
			if client.isAuthorized {
				t.Error("Expected client to be unauthorized")
			}
			if client.conn != nil {
				t.Error("Expected connection to be nil")
			}
		})
	}
}

func TestClient_IsConnected(t *testing.T) {
	tests := []struct {
		name  string
		setup func(*Client)
		want  bool
	}{
		{
			name: "connected",
			setup: func(c *Client) {
				c.isConnected = true
			},
			want: true,
		},
		{
			name: "not connected",
			setup: func(c *Client) {
				c.isConnected = false
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClient()
			tt.setup(client)
			
			if got := client.IsConnected(); got != tt.want {
				t.Errorf("IsConnected() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_IsAuthenticated(t *testing.T) {
	tests := []struct {
		name  string
		setup func(*Client)
		want  bool
	}{
		{
			name: "authenticated",
			setup: func(c *Client) {
				c.isAuthorized = true
			},
			want: true,
		},
		{
			name: "not authenticated",
			setup: func(c *Client) {
				c.isAuthorized = false
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClient()
			tt.setup(client)
			
			if got := client.IsAuthenticated(); got != tt.want {
				t.Errorf("IsAuthenticated() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPacketSerialization(t *testing.T) {
	tests := []struct {
		name   string
		packet *Packet
	}{
		{
			name: "auth packet",
			packet: &Packet{
				ID:   1,
				Type: PacketTypeAuth,
				Body: "password123",
			},
		},
		{
			name: "command packet",
			packet: &Packet{
				ID:   2,
				Type: PacketTypeCommand,
				Body: "status",
			},
		},
		{
			name: "empty body packet",
			packet: &Packet{
				ID:   3,
				Type: PacketTypeResponse,
				Body: "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClient()
			mc := newMockConn()
			client.conn = mc
			
			// Send packet
			err := client.sendPacket(tt.packet)
			if err != nil {
				t.Fatalf("sendPacket failed: %v", err)
			}
			
			// Read back what was written
			written := mc.writeBuf.Bytes()
			
			// Verify size field
			var size int32
			r := bytes.NewReader(written)
			if err := binary.Read(r, binary.LittleEndian, &size); err != nil {
				t.Fatalf("Failed to read size: %v", err)
			}
			
			expectedSize := int32(len(tt.packet.Body) + 10)
			if size != expectedSize {
				t.Errorf("Expected size %d, got %d", expectedSize, size)
			}
			
			// Verify the packet can be read back
			mc.readBuf = bytes.NewBuffer(written[4:]) // Skip size field
			client.conn = mc
			
			// Read the packet back (without size prefix)
			var id int32
			var pType int32
			
			if err := binary.Read(mc.readBuf, binary.LittleEndian, &id); err != nil {
				t.Fatalf("Failed to read ID: %v", err)
			}
			if id != tt.packet.ID {
				t.Errorf("Expected ID %d, got %d", tt.packet.ID, id)
			}
			
			if err := binary.Read(mc.readBuf, binary.LittleEndian, &pType); err != nil {
				t.Fatalf("Failed to read type: %v", err)
			}
			if PacketType(pType) != tt.packet.Type {
				t.Errorf("Expected type %d, got %d", tt.packet.Type, pType)
			}
		})
	}
}

// Helper function
func contains(s, substr string) bool {
	return bytes.Contains([]byte(s), []byte(substr))
}