package server_test

import (
	"testing"

	qt "github.com/frankban/quicktest"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/baseline"
	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/mcp/internal/server"
)

func TestBaselineUsesSharedEngineAndFixedStartupPolicy(t *testing.T) {
	c := qt.New(t)
	source := document.Source{Name: "guide.md", Format: document.Markdown, Bytes: []byte("Certainly! The client retries.")}
	capture, err := unswell.New(unswell.Options{CollectBaseline: true})
	c.Assert(err, qt.IsNil)
	result, err := capture.Analyze(t.Context(), source)
	c.Assert(err, qt.IsNil)
	file, err := baseline.Create(t.Context(), *result.BaselineSnapshot)
	c.Assert(err, qt.IsNil)
	data, err := baseline.Encode(t.Context(), file)
	c.Assert(err, qt.IsNil)
	session := connect(c, t.Context(), server.Options{Baseline: data, GateMode: "new"})
	engine, err := unswell.New(unswell.Options{Baseline: data, GateMode: "new"})
	c.Assert(err, qt.IsNil)
	for _, text := range []string{string(source.Bytes), "Certainly! The client must never retry."} {
		response, err := session.CallTool(t.Context(), &mcp.CallToolParams{Name: "unswell_check", Arguments: server.CheckInput{
			Sources: []server.Source{{Name: source.Name, Text: text}},
		}})
		c.Assert(err, qt.IsNil)
		source.Bytes = []byte(text)
		direct, err := engine.Analyze(t.Context(), source)
		c.Assert(err, qt.IsNil)
		checked := output[server.CheckOutput](c, response)
		c.Assert(checked.Result, qt.DeepEquals, direct)
	}
	description, err := session.CallTool(t.Context(), &mcp.CallToolParams{Name: "unswell_describe", Arguments: server.DescribeInput{}})
	c.Assert(err, qt.IsNil)
	described := output[server.Description](c, description)
	c.Assert(described.BaselineLoaded, qt.IsTrue)
	c.Assert(described.GateMode, qt.Equals, "new")
}
