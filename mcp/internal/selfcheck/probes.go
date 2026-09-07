package selfcheck

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/stokaro/unswell/mcp/internal/server"
)

func verifyProbes(ctx context.Context, session *mcp.ClientSession) ([]server.CheckOutput, error) {
	probes := []struct{ name, text, outcome string }{
		{"mcp-negative.md", "Certainly! The client opens connections.", "policy_failure"},
		{"mcp-rewritten.md", "The client opens connections.", "pass"},
		{"mcp-negative.go", "package sample\n// Certainly! The client opens connections.\n", "policy_failure"},
		{"mcp-negative.yaml", "message: Certainly! The client opens connections.\n", "policy_failure"},
		{"mcp-negative.cs", "class Sample { string value = \"Certainly! The client opens connections.\"; }", "policy_failure"},
		{"mcp-invalid.cs", "class Sample { string value = \"unfinished", "error"},
	}
	results := make([]server.CheckOutput, 0, len(probes))
	for _, probe := range probes {
		input := server.CheckInput{Sources: []server.Source{{Name: probe.name, Text: probe.text}}}
		checked, err := check(ctx, session, input, probe.outcome)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", probe.name, err)
		}
		if checked.Result.Gate.Passed != (probe.outcome == "pass") {
			return nil, fmt.Errorf("%s: MCP outcome disagrees with the engine gate", probe.name)
		}
		results = append(results, checked)
	}
	return results, nil
}
