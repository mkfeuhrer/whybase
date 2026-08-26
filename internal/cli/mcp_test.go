package cli

import (
	"strings"
	"testing"
)

func TestMCPCommandRegistered(t *testing.T) {
	found := false
	for _, c := range NewRoot().Commands() {
		if c.Name() == "mcp" {
			found = true
			if !strings.Contains(c.Short, "MCP") {
				t.Fatalf("short should mention MCP: %q", c.Short)
			}
		}
	}
	if !found {
		t.Fatal("mcp subcommand not registered")
	}
}
