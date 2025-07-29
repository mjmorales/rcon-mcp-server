package mcp

import (
	"context"
	"strings"
	"testing"

	"github.com/mjmorales/rcon-mcp-server/internal/rcon"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// resetSessionManager resets the global session manager for testing
func resetSessionManager() {
	sessionManager = rcon.NewSessionManager()
}

func TestConnect(t *testing.T) {
	tests := []struct {
		name        string
		params      ConnectParams
		setupFunc   func()
		wantErr     bool
		errContains string
		wantSuccess string
	}{
		{
			name: "successful connection",
			params: ConnectParams{
				SessionID: "test-session",
				Name:      "Test Server",
				Address:   "localhost:25575",
				Password:  "testpass",
			},
			setupFunc:   func() { resetSessionManager() },
			wantErr:     true, // Will fail to connect in test environment
			errContains: "failed to connect",
		},
		{
			name: "duplicate session ID",
			params: ConnectParams{
				SessionID: "duplicate-id",
				Name:      "Test Server",
				Address:   "localhost:25575",
				Password:  "testpass",
			},
			setupFunc: func() {
				resetSessionManager()
				// Pre-create a session
				sessionManager.CreateSession("duplicate-id", "Existing", "localhost:25575")
			},
			wantErr:     true,
			errContains: "already exists",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupFunc != nil {
				tt.setupFunc()
			}

			ctx := context.Background()
			params := &mcp.CallToolParamsFor[ConnectParams]{
				Arguments: tt.params,
			}

			result, err := Connect(ctx, nil, params)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got nil")
				} else if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("Expected error containing %q, got %q", tt.errContains, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
				if result == nil {
					t.Fatal("Expected result but got nil")
				}
				if len(result.Content) == 0 {
					t.Error("Expected content in result")
				}
			}
		})
	}
}

func TestDisconnect(t *testing.T) {
	tests := []struct {
		name        string
		params      DisconnectParams
		setupFunc   func()
		wantErr     bool
		errContains string
	}{
		{
			name: "disconnect existing session",
			params: DisconnectParams{
				SessionID: "test-session",
			},
			setupFunc: func() {
				resetSessionManager()
				sessionManager.CreateSession("test-session", "Test", "localhost:25575")
			},
			wantErr: false,
		},
		{
			name: "disconnect non-existent session",
			params: DisconnectParams{
				SessionID: "non-existent",
			},
			setupFunc: func() {
				resetSessionManager()
			},
			wantErr:     true,
			errContains: "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupFunc != nil {
				tt.setupFunc()
			}

			ctx := context.Background()
			params := &mcp.CallToolParamsFor[DisconnectParams]{
				Arguments: tt.params,
			}

			result, err := Disconnect(ctx, nil, params)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got nil")
				} else if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("Expected error containing %q, got %q", tt.errContains, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
				if result == nil {
					t.Fatal("Expected result but got nil")
				}
				// Verify session was removed
				if _, err := sessionManager.GetSession(tt.params.SessionID); err == nil {
					t.Error("Expected session to be removed")
				}
			}
		})
	}
}

func TestExecute(t *testing.T) {
	tests := []struct {
		name        string
		params      ExecuteParams
		setupFunc   func()
		wantErr     bool
		errContains string
	}{
		{
			name: "execute on non-existent session",
			params: ExecuteParams{
				SessionID: "non-existent",
				Command:   "status",
			},
			setupFunc: func() {
				resetSessionManager()
			},
			wantErr:     true,
			errContains: "not found",
		},
		{
			name: "execute on disconnected session",
			params: ExecuteParams{
				SessionID: "disconnected-session",
				Command:   "status",
			},
			setupFunc: func() {
				resetSessionManager()
				session, _ := sessionManager.CreateSession("disconnected-session", "Test", "localhost:25575")
				// Session exists but client is not connected
				session.Client = rcon.NewClient()
			},
			wantErr:     true,
			errContains: "not connected",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupFunc != nil {
				tt.setupFunc()
			}

			ctx := context.Background()
			params := &mcp.CallToolParamsFor[ExecuteParams]{
				Arguments: tt.params,
			}

			result, err := Execute(ctx, nil, params)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got nil")
				} else if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("Expected error containing %q, got %q", tt.errContains, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
				if result == nil {
					t.Fatal("Expected result but got nil")
				}
			}
		})
	}
}

func TestListSessions(t *testing.T) {
	tests := []struct {
		name       string
		setupFunc  func()
		wantOutput []string
	}{
		{
			name: "no active sessions",
			setupFunc: func() {
				resetSessionManager()
			},
			wantOutput: []string{"No active RCON sessions"},
		},
		{
			name: "multiple sessions with different states",
			setupFunc: func() {
				resetSessionManager()
				
				// Create disconnected session
				session1, _ := sessionManager.CreateSession("session-1", "Server 1", "localhost:25575")
				session1.Client = rcon.NewClient()
				
				// Create connected but not authenticated session
				session2, _ := sessionManager.CreateSession("session-2", "Server 2", "localhost:25576")
				session2.Client = rcon.NewClient()
				// Note: In a real test, we'd mock the connection state
				
				// Create another disconnected session
				session3, _ := sessionManager.CreateSession("session-3", "", "localhost:25577")
				session3.Client = rcon.NewClient()
			},
			wantOutput: []string{
				"Active RCON sessions:",
				"session-1 (Server 1): localhost:25575 - disconnected",
				"session-2 (Server 2): localhost:25576 - disconnected",
				"session-3 (unnamed): localhost:25577 - disconnected",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupFunc != nil {
				tt.setupFunc()
			}

			ctx := context.Background()
			params := &mcp.CallToolParamsFor[ListSessionsParams]{
				Arguments: ListSessionsParams{},
			}

			result, err := ListSessions(ctx, nil, params)

			if err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
			if result == nil {
				t.Fatal("Expected result but got nil")
			}
			if len(result.Content) == 0 {
				t.Fatal("Expected content in result")
			}

			// Check output content
			textContent, ok := result.Content[0].(*mcp.TextContent)
			if !ok {
				t.Fatal("Expected TextContent type")
			}

			output := textContent.Text
			for _, expected := range tt.wantOutput {
				if !strings.Contains(output, expected) {
					t.Errorf("Expected output to contain %q, got:\n%s", expected, output)
				}
			}
		})
	}
}

