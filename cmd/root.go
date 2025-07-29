// Package cmd contains all CLI commands for the RCON MCP server.
// It uses the Cobra library for command-line interface management.
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands.
// It provides the main entry point for the RCON MCP server CLI.
var rootCmd = &cobra.Command{
	Use:   "rcon-mcp-server",
	Short: "RCON Model Context Protocol server",
	Long: `RCON MCP Server is a Model Context Protocol (MCP) server that provides
tools for connecting to and managing RCON (Remote Console) servers.

This server enables AI assistants and other MCP clients to interact with
game servers and other applications that support the RCON protocol.

Features:
- Multiple simultaneous RCON connections
- Session management
- Secure authentication
- Command execution

To start the server, use:
  rcon-mcp-server serve`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
// It returns an error code to the OS on failure.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.rcon-mcp-server.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
}
