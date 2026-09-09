package selfcheck

// White-box tests: Supply a false-pass MCP session directly to negative probe verification;
// the public self-check entry point cannot receive an in-memory session.

import (
	"context"
	"testing"

	qt "github.com/frankban/quicktest"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/mcp/internal/server"
)

func TestProbeGateRejectsFalsePass(t *testing.T) {
	c := qt.New(t)
	engine, err := unswell.New(unswell.Options{})
	c.Assert(err, qt.IsNil)
	clean, err := engine.Analyze(t.Context(), document.Source{
		Name: "draft.md", Format: document.Markdown, Bytes: []byte("The client opens connections."),
	})
	c.Assert(err, qt.IsNil)
	instance := mcp.NewServer(&mcp.Implementation{Name: "false-pass-fixture", Version: "1"}, nil)
	mcp.AddTool(instance, &mcp.Tool{Name: "unswell_check"},
		func(context.Context, *mcp.CallToolRequest, server.CheckInput) (*mcp.CallToolResult, server.CheckOutput, error) {
			return nil, server.CheckOutput{Outcome: "pass", Result: clean}, nil
		})
	left, right := mcp.NewInMemoryTransports()
	serverSession, err := instance.Connect(t.Context(), left, nil)
	c.Assert(err, qt.IsNil)
	client := mcp.NewClient(&mcp.Implementation{Name: "gate-test", Version: "1"}, nil)
	session, err := client.Connect(t.Context(), right, nil)
	c.Assert(err, qt.IsNil)
	_, err = verifyProbes(t.Context(), session)
	c.Assert(err, qt.ErrorMatches, `mcp-negative.md: MCP probe returned "pass" .* expected "policy_failure"`)
	c.Assert(session.Close(), qt.IsNil)
	c.Assert(serverSession.Wait(), qt.IsNil)
}
