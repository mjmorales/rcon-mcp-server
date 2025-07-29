// Package mcp implements the Model Context Protocol server for RCON connections.
// It provides tools for connecting to, managing, and executing commands on RCON servers.
package mcp

import (
	"context"
	"fmt"
	"log"

	"github.com/mjmorales/rcon-mcp-server/internal/rcon"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// sessionManager is a singleton instance that manages all active RCON sessions.
// It provides thread-safe operations for creating, retrieving, and removing sessions.
var sessionManager = rcon.NewSessionManager()

// ConnectParams represents parameters for the connect tool
type ConnectParams struct {
	SessionID string `json:"session_id" jsonschema:"Unique identifier for this RCON session"`
	Name      string `json:"name,omitempty" jsonschema:"Friendly name for this connection (optional)"`
	Address   string `json:"address" jsonschema:"RCON server address (host:port)"`
	Password  string `json:"password" jsonschema:"RCON server password"`
}

// DisconnectParams represents parameters for the disconnect tool
type DisconnectParams struct {
	SessionID string `json:"session_id" jsonschema:"Session ID to disconnect"`
}

// ExecuteParams represents parameters for the execute tool
type ExecuteParams struct {
	SessionID string `json:"session_id" jsonschema:"Session ID to use for execution"`
	Command   string `json:"command" jsonschema:"Command to execute on the RCON server"`
}

// ListSessionsParams represents parameters for the list_sessions tool
type ListSessionsParams struct{}

// Connect establishes a new RCON connection to a server.
// It creates a session, connects to the server, and authenticates using the provided password.
// Returns an error if the session already exists, connection fails, or authentication fails.
func Connect(ctx context.Context, cc *mcp.ServerSession, params *mcp.CallToolParamsFor[ConnectParams]) (*mcp.CallToolResultFor[any], error) {
	// Create a new session
	session, err := sessionManager.CreateSession(params.Arguments.SessionID, params.Arguments.Name, params.Arguments.Address)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	// Connect to the server
	if err := session.Client.Connect(params.Arguments.Address); err != nil {
		_ = sessionManager.RemoveSession(params.Arguments.SessionID)
		return nil, fmt.Errorf("failed to connect: %w", err)
	}

	// Authenticate
	if err := session.Client.Authenticate(params.Arguments.Password); err != nil {
		_ = sessionManager.RemoveSession(params.Arguments.SessionID)
		return nil, fmt.Errorf("failed to authenticate: %w", err)
	}

	return &mcp.CallToolResultFor[any]{
		Content: []mcp.Content{&mcp.TextContent{
			Text: fmt.Sprintf("Connected to RCON server at %s (session: %s)", params.Arguments.Address, params.Arguments.SessionID),
		}},
	}, nil
}

// Disconnect terminates an existing RCON connection and removes the session.
// Returns an error if the session doesn't exist.
func Disconnect(ctx context.Context, cc *mcp.ServerSession, params *mcp.CallToolParamsFor[DisconnectParams]) (*mcp.CallToolResultFor[any], error) {
	if err := sessionManager.RemoveSession(params.Arguments.SessionID); err != nil {
		return nil, fmt.Errorf("failed to disconnect: %w", err)
	}

	return &mcp.CallToolResultFor[any]{
		Content: []mcp.Content{&mcp.TextContent{
			Text: fmt.Sprintf("Disconnected session: %s", params.Arguments.SessionID),
		}},
	}, nil
}

// Execute sends a command to the RCON server and returns the response.
// The session must exist and be authenticated. Returns an error if the session
// is not found or if command execution fails.
func Execute(ctx context.Context, cc *mcp.ServerSession, params *mcp.CallToolParamsFor[ExecuteParams]) (*mcp.CallToolResultFor[any], error) {
	// Get the session
	session, err := sessionManager.GetSession(params.Arguments.SessionID)
	if err != nil {
		return nil, fmt.Errorf("session not found: %w", err)
	}

	// Execute the command
	response, err := session.Client.Execute(params.Arguments.Command)
	if err != nil {
		return nil, fmt.Errorf("failed to execute command: %w", err)
	}

	return &mcp.CallToolResultFor[any]{
		Content: []mcp.Content{&mcp.TextContent{
			Text: response,
		}},
	}, nil
}

// ListSessions retrieves information about all active RCON sessions.
// It returns session IDs, names, addresses, and connection/authentication status.
func ListSessions(ctx context.Context, cc *mcp.ServerSession, params *mcp.CallToolParamsFor[ListSessionsParams]) (*mcp.CallToolResultFor[any], error) {
	sessions := sessionManager.ListSessions()

	if len(sessions) == 0 {
		return &mcp.CallToolResultFor[any]{
			Content: []mcp.Content{&mcp.TextContent{
				Text: "No active RCON sessions",
			}},
		}, nil
	}

	sessionInfo := "Active RCON sessions:\n"
	for _, session := range sessions {
		status := "disconnected"
		if session.Client.IsConnected() {
			if session.Client.IsAuthenticated() {
				status = "connected & authenticated"
			} else {
				status = "connected (not authenticated)"
			}
		}

		name := session.Name
		if name == "" {
			name = "unnamed"
		}

		sessionInfo += fmt.Sprintf("- %s (%s): %s - %s\n", session.ID, name, session.Address, status)
	}

	return &mcp.CallToolResultFor[any]{
		Content: []mcp.Content{&mcp.TextContent{
			Text: sessionInfo,
		}},
	}, nil
}

// Serve initializes and runs the MCP server.
// It registers all RCON tools and starts listening for MCP connections via stdio.
// The function blocks until the server is terminated or encounters a fatal error.
func Serve() {
	// Create a server
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "rcon-mcp-server",
		Version: "v1.0.0",
	}, nil)

	// Register RCON tools
	mcp.AddTool(server, &mcp.Tool{
		Name:        "rcon_connect",
		Description: "Connect to an RCON server and authenticate",
	}, Connect)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "rcon_disconnect",
		Description: "Disconnect from an RCON server",
	}, Disconnect)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "rcon_execute",
		Description: "Execute a command on an RCON server",
	}, Execute)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "rcon_list_sessions",
		Description: "List all active RCON sessions",
	}, ListSessions)

	fmt.Println("RCON MCP server is ready!")
	// Run the server
	if err := server.Run(context.Background(), mcp.NewStdioTransport()); err != nil {
		log.Fatal(err)
	}

	// Cleanup all sessions on exit to ensure graceful shutdown
	if err := sessionManager.DisconnectAll(); err != nil {
		log.Printf("Failed to disconnect all sessions cleanly: %v", err)
	}
}
