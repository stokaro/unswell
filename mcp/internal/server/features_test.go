package server_test

import (
	"testing"

	qt "github.com/frankban/quicktest"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/mcp/internal/server"
)

func TestFeatureCollectionMatchesThePublicEngine(t *testing.T) {
	c := qt.New(t)
	ids := []string{"prose-words", "noun-token-ratio", "activation/readability.long-paragraph"}
	session := connect(c, t.Context(), server.Options{Features: ids})
	engine, err := unswell.New(unswell.Options{Features: ids})
	c.Assert(err, qt.IsNil)
	response, err := session.CallTool(t.Context(), &mcp.CallToolParams{Name: "unswell_describe", Arguments: server.DescribeInput{}})
	c.Assert(err, qt.IsNil)
	c.Assert(output[server.Description](c, response).Features, qt.DeepEquals, engine.FeatureIDs())
	const prose = "# Heading\n\nThe cache expires.\n"
	response, err = session.CallTool(t.Context(), &mcp.CallToolParams{Name: "unswell_check", Arguments: server.CheckInput{
		Sources: []server.Source{{Name: "guide.md", Format: document.Markdown, Text: prose}},
	}})
	c.Assert(err, qt.IsNil)
	c.Assert(response.IsError, qt.IsFalse)
	checked := output[server.CheckOutput](c, response)
	direct, err := engine.Analyze(t.Context(), document.Source{Name: "guide.md", Format: document.Markdown, Bytes: []byte(prose)})
	c.Assert(err, qt.IsNil)
	c.Assert(checked.Result, qt.DeepEquals, direct)
	c.Assert(checked.Result.Features, qt.IsNotNil)
}
