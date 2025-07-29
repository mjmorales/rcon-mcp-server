package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestServeCommand(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		wantOutput []string
		wantErr    bool
	}{
		{
			name:       "serve command help",
			args:       []string{"serve", "--help"},
			wantOutput: []string{"Start the RCON Model Context Protocol", "Available tools:", "rcon_connect", "rcon_execute"},
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset command for each test
			rootCmd.SetArgs(tt.args)
			
			// Capture output
			var buf bytes.Buffer
			rootCmd.SetOut(&buf)
			rootCmd.SetErr(&buf)
			
			// Execute command
			err := rootCmd.Execute()
			
			// Check error
			if tt.wantErr && err == nil {
				t.Error("Expected error but got nil")
			} else if !tt.wantErr && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
			
			// Check output
			output := buf.String()
			for _, expected := range tt.wantOutput {
				if !strings.Contains(output, expected) {
					t.Errorf("Expected output to contain %q, got:\n%s", expected, output)
				}
			}
		})
	}
}

func TestServeCommandStructure(t *testing.T) {
	// Test that serve command is properly registered
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Use == "serve" {
			found = true
			
			// Verify command properties
			if cmd.Short == "" {
				t.Error("Expected serve command to have a short description")
			}
			if cmd.Long == "" {
				t.Error("Expected serve command to have a long description")
			}
			if cmd.Run == nil {
				t.Error("Expected serve command to have a Run function")
			}
			
			break
		}
	}
	
	if !found {
		t.Error("serve command not found in root command")
	}
}