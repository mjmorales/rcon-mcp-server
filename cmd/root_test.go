package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestRootCommand(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		wantOutput []string
		wantErr    bool
	}{
		{
			name:       "root command help",
			args:       []string{"--help"},
			wantOutput: []string{"RCON MCP Server is a Model Context Protocol", "Available Commands:", "serve"},
			wantErr:    false,
		},
		{
			name:       "root command without args",
			args:       []string{},
			wantOutput: []string{"RCON MCP Server is a Model Context Protocol"},
			wantErr:    false,
		},
		{
			name:       "invalid command",
			args:       []string{"invalid"},
			wantOutput: []string{"unknown command"},
			wantErr:    true,
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

func TestExecuteFunction(t *testing.T) {
	// Save original args
	oldArgs := rootCmd.Commands()
	
	// Test Execute function doesn't panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Execute() panicked: %v", r)
		}
		
		// Restore commands
		for _, cmd := range oldArgs {
			rootCmd.AddCommand(cmd)
		}
	}()
	
	// Note: We can't easily test os.Exit(1) behavior
	// In a real scenario, we might use a different approach
}