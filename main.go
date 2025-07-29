// Package main provides the entry point for the RCON MCP server application.
package main

import "github.com/mjmorales/rcon-mcp-server/cmd"

func main() {
	// Execute runs the root command which starts the CLI application.
	// This delegates all command handling to the cmd package using Cobra.
	cmd.Execute()
}
