package server_test

import (
	"testing"

	qt "github.com/frankban/quicktest"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/mcp/internal/server"
)

func TestChecksRetainSuppressionAuditAndRejectUnusedPermissions(t *testing.T) {
	c := qt.New(t)
	policy := []byte("version: 1\nextends: [builtin:custom]\nrules:\n" +
		"  policy.banned-phrases: {enabled: true, parameters: {phrases: [robust]}, score: {weight: 20, cap: 20}}\n")
	session := connect(c, t.Context(), server.Options{Config: policy})
	engine, err := unswell.New(unswell.Options{Config: policy})
	c.Assert(err, qt.IsNil)
	for _, prose := range []string{"The robust client starts.", "The client starts."} {
		text := "<!-- unswell-disable-next-block policy.banned-phrases -- Required contract wording. -->\n\n" + prose
		response, err := session.CallTool(t.Context(), &mcp.CallToolParams{Name: "unswell_check", Arguments: server.CheckInput{
			Sources: []server.Source{{Name: "guide.md", Text: text}},
		}})
		c.Assert(err, qt.IsNil)
		direct, analysisErr := engine.Analyze(t.Context(), document.Source{Name: "guide.md", Format: document.Markdown, Bytes: []byte(text)})
		c.Assert(response.IsError, qt.Equals, analysisErr != nil)
		checked := output[server.CheckOutput](c, response)
		c.Assert(checked.Result, qt.DeepEquals, direct)
		c.Assert(checked.Result.Suppressions, qt.HasLen, 1)
	}
}
