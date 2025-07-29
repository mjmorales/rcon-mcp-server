package cmd

import (
	"github.com/mjmorales/rcon-mcp-server/internal/mcp"
	"github.com/spf13/cobra"
)

// serveCmd represents the serve command which starts the MCP server.
// This command initializes the RCON MCP server and begins listening for connections.
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the RCON MCP server",
	Long: `Start the RCON Model Context Protocol (MCP) server.

This server provides tools for connecting to and managing RCON servers,
allowing you to execute commands remotely through the MCP interface.

Available tools:
- rcon_connect: Connect to an RCON server
- rcon_disconnect: Disconnect from an RCON server
- rcon_execute: Execute commands on an RCON server
- rcon_list_sessions: List all active RCON sessions`,
	Run: func(cmd *cobra.Command, args []string) {
		// Start the MCP server. This will block until the server is terminated.
		mcp.Serve()
	},
}

// init registers the serve command with the root command during package initialization.
func init() {
	rootCmd.AddCommand(serveCmd)
}
