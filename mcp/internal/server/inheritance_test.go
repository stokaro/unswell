package server_test

import (
	"testing"

	qt "github.com/frankban/quicktest"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/config"
	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/mcp/internal/server"
)

func TestDescriptionAndChecksShareInheritedFilePolicy(t *testing.T) {
	c := qt.New(t)
	bundle := config.Bundle{Root: "policy.yaml", Files: map[string][]byte{
		"policy.yaml": []byte("version: 1\nextends: [base.yaml]\noverrides:\n" +
			"  - files: [reference/**]\n    rules: {policy.banned-phrases: {enabled: false}}\n"),
		"base.yaml": []byte("version: 1\nextends: [builtin:custom]\n" +
			"rules: {policy.banned-phrases: {enabled: true, parameters: {phrases: [robust]}}}\n" +
			"vocabulary: {dictionaries: [terms.yaml], term_exemptions: [policy.banned-phrases]}\n"),
		"terms.yaml": []byte("version: 1\nterms: [robust estimator]\n"),
	}}
	session := connect(c, t.Context(), server.Options{ConfigBundle: &bundle})
	engine, err := unswell.New(unswell.Options{ConfigBundle: &bundle})
	c.Assert(err, qt.IsNil)
	for _, name := range []string{"guide.md", "reference/api.md"} {
		response, err := session.CallTool(t.Context(), &mcp.CallToolParams{Name: "unswell_describe", Arguments: server.DescribeInput{File: name}})
		c.Assert(err, qt.IsNil)
		c.Assert(response.IsError, qt.IsFalse)
		description := output[server.Description](c, response)
		text := "The robust estimator uses robust methods."
		response, err = session.CallTool(t.Context(), &mcp.CallToolParams{Name: "unswell_check", Arguments: server.CheckInput{
			Sources: []server.Source{{Name: name, Text: text}},
		}})
		c.Assert(err, qt.IsNil)
		c.Assert(response.IsError, qt.IsFalse)
		checked := output[server.CheckOutput](c, response)
		direct, err := engine.Analyze(t.Context(), document.Source{Name: name, Format: document.Markdown, Bytes: []byte(text)})
		c.Assert(err, qt.IsNil)
		c.Assert(checked.Result, qt.DeepEquals, direct)
		c.Assert(checked.Result.Documents[0].ConfigHash, qt.Equals, description.Policy.Hash)
	}
}
