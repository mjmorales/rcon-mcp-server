package rcon

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

// Session represents a managed RCON connection session.
// Each session maintains its own client connection and metadata.
type Session struct {
	ID      string  // Unique identifier for the session
	Client  *Client // RCON client instance for this session
	Address string  // Server address in "host:port" format
	Name    string  // Optional friendly name for the session
	Created int64   // Unix timestamp when the session was created
}

// SessionManager provides thread-safe management of multiple RCON sessions.
// It allows creating, retrieving, listing, and removing sessions.
type SessionManager struct {
	sessions map[string]*Session // Map of session ID to session instance
	mu       sync.RWMutex        // Read-write mutex for thread-safe access
}

// NewSessionManager creates a new instance of SessionManager.
// The manager starts with no active sessions.
func NewSessionManager() *SessionManager {
	return &SessionManager{
		sessions: make(map[string]*Session),
	}
}

// CreateSession creates a new RCON session with the specified parameters.
// Returns an error if a session with the given ID already exists.
// The session is created with a new client but not connected.
func (sm *SessionManager) CreateSession(id, name, address string) (*Session, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if _, exists := sm.sessions[id]; exists {
		return nil, fmt.Errorf("session with ID %s already exists", id)
	}

	session := &Session{
		ID:      id,
		Client:  NewClient(),
		Address: address,
		Name:    name,
		Created: getCurrentTimestamp(),
	}

	sm.sessions[id] = session
	return session, nil
}

// GetSession retrieves an existing session by its ID.
// Returns an error if the session doesn't exist.
func (sm *SessionManager) GetSession(id string) (*Session, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	session, exists := sm.sessions[id]
	if !exists {
		return nil, fmt.Errorf("session with ID %s not found", id)
	}

	return session, nil
}

// ListSessions returns a slice of all active sessions.
// The returned slice is a copy and can be safely modified.
func (sm *SessionManager) ListSessions() []*Session {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	sessions := make([]*Session, 0, len(sm.sessions))
	for _, session := range sm.sessions {
		sessions = append(sessions, session)
	}

	return sessions
}

// RemoveSession removes a session from the manager and disconnects its client.
// Returns an error if the session doesn't exist.
// The client is gracefully disconnected before removal.
func (sm *SessionManager) RemoveSession(id string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	session, exists := sm.sessions[id]
	if !exists {
		return fmt.Errorf("session with ID %s not found", id)
	}

	// Disconnect the client if connected
	if session.Client.IsConnected() {
		if err := session.Client.Disconnect(); err != nil {
			return fmt.Errorf("failed to disconnect client: %w", err)
		}
	}

	delete(sm.sessions, id)
	return nil
}

// DisconnectAll disconnects all active sessions and clears the session map.
// This is typically called during server shutdown.
// Returns an error if any disconnection fails, but attempts to disconnect all sessions.
func (sm *SessionManager) DisconnectAll() error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	var errs []error
	for id, session := range sm.sessions {
		if session.Client.IsConnected() {
			if err := session.Client.Disconnect(); err != nil {
				errs = append(errs, fmt.Errorf("failed to disconnect session %s: %w", id, err))
			}
		}
	}

	// Clear all sessions
	sm.sessions = make(map[string]*Session)

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}

// getCurrentTimestamp returns the current Unix timestamp in seconds.
// Used for tracking session creation time.
func getCurrentTimestamp() int64 {
	return time.Now().Unix()
}
