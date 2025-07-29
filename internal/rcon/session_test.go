package rcon

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestNewSessionManager(t *testing.T) {
	sm := NewSessionManager()
	if sm == nil {
		t.Fatal("NewSessionManager returned nil")
	}
	if sm.sessions == nil {
		t.Error("Expected sessions map to be initialized")
	}
	if len(sm.sessions) != 0 {
		t.Errorf("Expected empty sessions map, got %d sessions", len(sm.sessions))
	}
}

func TestSessionManager_CreateSession(t *testing.T) {
	tests := []struct {
		name        string
		sessionID   string
		sessionName string
		address     string
		setupFunc   func(*SessionManager)
		wantErr     bool
		errContains string
	}{
		{
			name:        "create new session",
			sessionID:   "test-session-1",
			sessionName: "Test Server",
			address:     "localhost:25575",
			wantErr:     false,
		},
		{
			name:        "create session with empty name",
			sessionID:   "test-session-2",
			sessionName: "",
			address:     "localhost:25575",
			wantErr:     false,
		},
		{
			name:        "create duplicate session",
			sessionID:   "duplicate-id",
			sessionName: "Duplicate",
			address:     "localhost:25575",
			setupFunc: func(sm *SessionManager) {
				// Pre-create a session with the same ID
				sm.sessions["duplicate-id"] = &Session{ID: "duplicate-id"}
			},
			wantErr:     true,
			errContains: "already exists",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sm := NewSessionManager()
			if tt.setupFunc != nil {
				tt.setupFunc(sm)
			}

			session, err := sm.CreateSession(tt.sessionID, tt.sessionName, tt.address)

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
				if session == nil {
					t.Fatal("Expected session but got nil")
				}
				if session.ID != tt.sessionID {
					t.Errorf("Expected session ID %q, got %q", tt.sessionID, session.ID)
				}
				if session.Name != tt.sessionName {
					t.Errorf("Expected session name %q, got %q", tt.sessionName, session.Name)
				}
				if session.Address != tt.address {
					t.Errorf("Expected address %q, got %q", tt.address, session.Address)
				}
				if session.Client == nil {
					t.Error("Expected client to be initialized")
				}
				if session.Created == 0 {
					t.Error("Expected created timestamp to be set")
				}
			}
		})
	}
}

func TestSessionManager_GetSession(t *testing.T) {
	tests := []struct {
		name        string
		sessionID   string
		setupFunc   func(*SessionManager)
		wantErr     bool
		errContains string
		wantSession *Session
	}{
		{
			name:      "get existing session",
			sessionID: "existing-session",
			setupFunc: func(sm *SessionManager) {
				sm.sessions["existing-session"] = &Session{
					ID:      "existing-session",
					Name:    "Test Session",
					Address: "localhost:25575",
				}
			},
			wantErr: false,
			wantSession: &Session{
				ID:      "existing-session",
				Name:    "Test Session",
				Address: "localhost:25575",
			},
		},
		{
			name:        "get non-existent session",
			sessionID:   "non-existent",
			setupFunc:   func(sm *SessionManager) {},
			wantErr:     true,
			errContains: "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sm := NewSessionManager()
			if tt.setupFunc != nil {
				tt.setupFunc(sm)
			}

			session, err := sm.GetSession(tt.sessionID)

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
				if session == nil {
					t.Fatal("Expected session but got nil")
				}
				if tt.wantSession != nil {
					if session.ID != tt.wantSession.ID {
						t.Errorf("Expected session ID %q, got %q", tt.wantSession.ID, session.ID)
					}
					if session.Name != tt.wantSession.Name {
						t.Errorf("Expected session name %q, got %q", tt.wantSession.Name, session.Name)
					}
					if session.Address != tt.wantSession.Address {
						t.Errorf("Expected address %q, got %q", tt.wantSession.Address, session.Address)
					}
				}
			}
		})
	}
}

func TestSessionManager_ListSessions(t *testing.T) {
	tests := []struct {
		name      string
		setupFunc func(*SessionManager)
		wantCount int
		wantIDs   []string
	}{
		{
			name:      "empty list",
			setupFunc: func(sm *SessionManager) {},
			wantCount: 0,
			wantIDs:   []string{},
		},
		{
			name: "multiple sessions",
			setupFunc: func(sm *SessionManager) {
				sm.sessions["session-1"] = &Session{ID: "session-1"}
				sm.sessions["session-2"] = &Session{ID: "session-2"}
				sm.sessions["session-3"] = &Session{ID: "session-3"}
			},
			wantCount: 3,
			wantIDs:   []string{"session-1", "session-2", "session-3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sm := NewSessionManager()
			if tt.setupFunc != nil {
				tt.setupFunc(sm)
			}

			sessions := sm.ListSessions()

			if len(sessions) != tt.wantCount {
				t.Errorf("Expected %d sessions, got %d", tt.wantCount, len(sessions))
			}

			// Check that all expected IDs are present
			sessionMap := make(map[string]bool)
			for _, s := range sessions {
				sessionMap[s.ID] = true
			}

			for _, wantID := range tt.wantIDs {
				if !sessionMap[wantID] {
					t.Errorf("Expected session ID %q not found", wantID)
				}
			}
		})
	}
}

func TestSessionManager_RemoveSession(t *testing.T) {
	tests := []struct {
		name        string
		sessionID   string
		setupFunc   func(*SessionManager)
		wantErr     bool
		errContains string
	}{
		{
			name:      "remove existing disconnected session",
			sessionID: "session-to-remove",
			setupFunc: func(sm *SessionManager) {
				sm.sessions["session-to-remove"] = &Session{
					ID:     "session-to-remove",
					Client: NewClient(),
				}
			},
			wantErr: false,
		},
		{
			name:      "remove connected session",
			sessionID: "connected-session",
			setupFunc: func(sm *SessionManager) {
				client := NewClient()
				client.isConnected = true
				client.conn = newMockConn()
				sm.sessions["connected-session"] = &Session{
					ID:     "connected-session",
					Client: client,
				}
			},
			wantErr: false,
		},
		{
			name:        "remove non-existent session",
			sessionID:   "non-existent",
			setupFunc:   func(sm *SessionManager) {},
			wantErr:     true,
			errContains: "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sm := NewSessionManager()
			if tt.setupFunc != nil {
				tt.setupFunc(sm)
			}

			err := sm.RemoveSession(tt.sessionID)

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
				// Verify session was removed
				if _, exists := sm.sessions[tt.sessionID]; exists {
					t.Error("Expected session to be removed but it still exists")
				}
			}
		})
	}
}

func TestSessionManager_DisconnectAll(t *testing.T) {
	tests := []struct {
		name      string
		setupFunc func(*SessionManager)
		wantErr   bool
	}{
		{
			name:      "disconnect empty sessions",
			setupFunc: func(sm *SessionManager) {},
			wantErr:   false,
		},
		{
			name: "disconnect multiple sessions",
			setupFunc: func(sm *SessionManager) {
				// Add disconnected session
				sm.sessions["session-1"] = &Session{
					ID:     "session-1",
					Client: NewClient(),
				}
				
				// Add connected session
				client2 := NewClient()
				client2.isConnected = true
				client2.conn = newMockConn()
				sm.sessions["session-2"] = &Session{
					ID:     "session-2",
					Client: client2,
				}
				
				// Add another connected session
				client3 := NewClient()
				client3.isConnected = true
				client3.conn = newMockConn()
				sm.sessions["session-3"] = &Session{
					ID:     "session-3",
					Client: client3,
				}
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sm := NewSessionManager()
			if tt.setupFunc != nil {
				tt.setupFunc(sm)
			}

			err := sm.DisconnectAll()

			if tt.wantErr && err == nil {
				t.Error("Expected error but got nil")
			} else if !tt.wantErr && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}

			// Verify all sessions were removed
			if len(sm.sessions) != 0 {
				t.Errorf("Expected all sessions to be removed, but %d remain", len(sm.sessions))
			}
		})
	}
}

func TestSessionManager_ConcurrentAccess(t *testing.T) {
	sm := NewSessionManager()
	done := make(chan bool)
	
	// Number of goroutines to run concurrently
	numGoroutines := 10
	
	// Start goroutines that create sessions
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			sessionID := fmt.Sprintf("session-%d", id)
			_, err := sm.CreateSession(sessionID, "Test", "localhost:25575")
			if err != nil {
				t.Errorf("Failed to create session %s: %v", sessionID, err)
			}
			done <- true
		}(i)
	}
	
	// Start goroutines that list sessions
	for i := 0; i < numGoroutines; i++ {
		go func() {
			sessions := sm.ListSessions()
			// Just accessing the list, actual count may vary due to timing
			_ = len(sessions)
			done <- true
		}()
	}
	
	// Start goroutines that get sessions
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			sessionID := fmt.Sprintf("session-%d", id)
			// May or may not exist yet
			_, _ = sm.GetSession(sessionID)
			done <- true
		}(i)
	}
	
	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines*3; i++ {
		<-done
	}
	
	// Verify final state
	sessions := sm.ListSessions()
	if len(sessions) != numGoroutines {
		t.Errorf("Expected %d sessions, got %d", numGoroutines, len(sessions))
	}
}

func TestGetCurrentTimestamp(t *testing.T) {
	before := time.Now().Unix()
	timestamp := getCurrentTimestamp()
	after := time.Now().Unix()
	
	if timestamp < before || timestamp > after {
		t.Errorf("Timestamp %d is not within expected range [%d, %d]", timestamp, before, after)
	}
}