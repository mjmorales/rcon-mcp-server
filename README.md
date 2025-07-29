# RCON MCP Server

A Model Context Protocol (MCP) server that provides tools for connecting to and managing RCON (Remote Console) servers. This server enables AI assistants and other MCP clients to interact with game servers and other applications that support the RCON protocol.

## Features

- **Multiple Simultaneous Connections**: Manage multiple RCON sessions concurrently
- **Session Management**: Create, list, and remove RCON sessions with friendly names
- **Secure Authentication**: Password-based authentication for RCON servers
- **Command Execution**: Execute commands on connected RCON servers
- **Thread-Safe Operations**: All operations are thread-safe for concurrent use
- **Clean Architecture**: Well-structured Go code with comprehensive tests

## Installation

### Prerequisites

- Go 1.24.5 or higher
- Git

### Building from Source

```bash
# Clone the repository
git clone https://github.com/mjmorales/rcon-mcp-server.git
cd rcon-mcp-server

# Build the binary
go build -o rcon-mcp-server

# Or install globally
go install
```

## Usage

### Starting the Server

```bash
# Start the MCP server
./rcon-mcp-server serve

# Or if installed globally
rcon-mcp-server serve
```

The server will start and listen for MCP connections via stdio.

### Available MCP Tools

The server provides the following tools:

1. **rcon_connect** - Connect to an RCON server
   - `session_id` (required): Unique identifier for this session
   - `name` (optional): Friendly name for the connection
   - `address` (required): RCON server address (host:port)
   - `password` (required): RCON server password

2. **rcon_disconnect** - Disconnect from an RCON server
   - `session_id` (required): Session ID to disconnect

3. **rcon_execute** - Execute a command on an RCON server
   - `session_id` (required): Session ID to use
   - `command` (required): Command to execute

4. **rcon_list_sessions** - List all active RCON sessions
   - No parameters required

### Example Configuration

For Claude Desktop or other MCP clients, add this to your configuration:

```json
{
  "mcpServers": {
    "rcon": {
      "command": "/path/to/rcon-mcp-server",
      "args": ["serve"]
    }
  }
}
```

## Development

### Project Structure

```
rcon-mcp-server/
├── cmd/                    # CLI commands
│   ├── root.go            # Root command setup
│   └── serve.go           # Serve command implementation
├── internal/              # Internal packages
│   ├── mcp/              # MCP server implementation
│   │   └── server.go     # MCP tool handlers
│   └── rcon/             # RCON protocol implementation
│       ├── client.go     # RCON client
│       └── session.go    # Session management
├── main.go               # Entry point
└── go.mod               # Go module definition
```

### Running Tests

```bash
# Run all tests
go test ./... -v

# Run tests with coverage
go test ./... -v -cover

# Generate coverage report
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html
```

### Code Quality

The codebase follows Go best practices:
- Comprehensive documentation for all exported types and functions
- Thread-safe operations with proper mutex usage
- Error handling with descriptive error messages
- Table-driven tests with good coverage
- Clean separation of concerns

## RCON Protocol

This server implements the Source RCON Protocol, which is widely supported by game servers including:
- Minecraft
- Counter-Strike
- Team Fortress 2
- Rust
- ARK: Survival Evolved
- And many others

### Protocol Details

The RCON protocol uses TCP connections with packet-based communication:
- Packet structure: Size (4 bytes) + ID (4 bytes) + Type (4 bytes) + Body + Null terminators
- Authentication required before command execution
- Request/response matching using packet IDs

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

### Guidelines

1. Follow Go conventions and best practices
2. Add tests for new functionality
3. Update documentation as needed
4. Ensure all tests pass before submitting

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Security

- Never commit RCON passwords or sensitive information
- Use environment variables or secure configuration for credentials
- Be cautious when exposing RCON access through MCP

## Troubleshooting

### Connection Issues

- Ensure the RCON server is running and accessible
- Verify the server address format is `host:port`
- Check that RCON is enabled on the target server
- Confirm the password is correct

### Session Management

- Each session requires a unique ID
- Sessions persist until explicitly disconnected
- Use `rcon_list_sessions` to see all active sessions

## Support

For issues, questions, or contributions, please visit the [GitHub repository](https://github.com/mjmorales/rcon-mcp-server).