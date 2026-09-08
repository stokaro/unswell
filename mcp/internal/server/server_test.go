package server_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	qt "github.com/frankban/quicktest"
	"github.com/modelcontextprotocol/go-sdk/jsonrpc"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/mcp/internal/server"
	"github.com/stokaro/unswell/nlp"
	"github.com/stokaro/unswell/nlp/english"
)

func connect(c *qt.C, ctx context.Context, options server.Options) *mcp.ClientSession {
	c.Helper()
	instance, err := server.New(options)
	c.Assert(err, qt.IsNil)
	client := mcp.NewClient(&mcp.Implementation{Name: "unswell-test", Version: "1"}, nil)
	left, right := mcp.NewInMemoryTransports()
	serverSession, err := instance.Connect(ctx, left, nil)
	c.Assert(err, qt.IsNil)
	session, err := client.Connect(ctx, right, nil)
	c.Assert(err, qt.IsNil)
	c.Cleanup(func() {
		// Stop server work while the client can still read. Closing the client
		// first can interrupt a canceled request's pending response write.
		c.Assert(serverSession.Close(), qt.IsNil)
		c.Assert(session.Close(), qt.IsNil)
		if err := serverSession.Wait(); err != nil {
			// The SDK may reject that response after its explicit shutdown starts.
			c.Assert(err, qt.ErrorIs, &jsonrpc.Error{Code: -32004})
		}
	})
	return session
}

func output[T any](c *qt.C, result *mcp.CallToolResult) T {
	c.Helper()
	encoded, err := json.Marshal(result.StructuredContent)
	c.Assert(err, qt.IsNil)
	var value T
	c.Assert(json.Unmarshal(encoded, &value), qt.IsNil)
	return value
}

func TestDiscoveryAndFixedPolicy(t *testing.T) {
	c := qt.New(t)
	policy := []byte("version: 1\nextraction:\n  contexts: [comment]\n  languages: {yaml: {contexts: [string]}}\n")
	session := connect(c, t.Context(), server.Options{Config: policy})
	tools, err := session.ListTools(t.Context(), nil)
	c.Assert(err, qt.IsNil)
	c.Assert(tools.Tools, qt.HasLen, 2)
	for _, tool := range tools.Tools {
		c.Assert(tool.Annotations.ReadOnlyHint, qt.IsTrue)
		c.Assert(*tool.Annotations.OpenWorldHint, qt.IsFalse)
		c.Assert(tool.OutputSchema, qt.IsNotNil)
	}
	result, err := session.CallTool(t.Context(), &mcp.CallToolParams{Name: "unswell_describe", Arguments: server.DescribeInput{}})
	c.Assert(err, qt.IsNil)
	c.Assert(result.IsError, qt.IsFalse)
	description := output[server.Description](c, result)
	c.Assert(description.Formats, qt.HasLen, len(document.Formats()))
	c.Assert(description.Policy.Extraction.Contexts, qt.DeepEquals, []string{"comment"})
	c.Assert(description.Policy.Extraction.Languages[document.YAML].Contexts, qt.DeepEquals, []string{"string"})
	engine, err := unswell.New(unswell.Options{Config: policy})
	c.Assert(err, qt.IsNil)
	c.Assert(description.Rules, qt.DeepEquals, engine.Catalog())
}

func TestChecksMatchEngine(t *testing.T) {
	cases := []struct {
		name, text, outcome string
		format              document.Format
		failed              bool
	}{
		{"clean.md", "The client opens connections.", "pass", document.Markdown, false},
		{"bad.md", "Certainly! The client opens connections.", "policy_failure", document.Markdown, false},
		{
			"message.cs",
			"class Sample { string value = \"Certainly! The client opens connections.\"; }",
			"policy_failure",
			document.CSharp,
			false,
		},
		{"config.yaml", "message: Certainly! The client opens connections.\n", "policy_failure", document.YAML, false},
		{"bad.cs", "class Sample { string value = \"unfinished", "error", document.CSharp, true},
		{"empty.md", "", "error", document.Markdown, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := qt.New(t)
			policy := []byte("version: 1\nextends: [builtin:strict-v1]\n")
			session := connect(c, t.Context(), server.Options{Config: policy})
			input := server.CheckInput{Sources: []server.Source{{Name: tc.name, Text: tc.text}}}
			result, err := session.CallTool(t.Context(), &mcp.CallToolParams{Name: "unswell_check", Arguments: input})
			c.Assert(err, qt.IsNil)
			c.Assert(result.IsError, qt.Equals, tc.failed)
			checked := output[server.CheckOutput](c, result)
			c.Assert(checked.Outcome, qt.Equals, tc.outcome)
			engine, err := unswell.New(unswell.Options{Config: policy})
			c.Assert(err, qt.IsNil)
			direct, directErr := engine.Analyze(t.Context(), document.Source{Name: tc.name, Format: tc.format, Bytes: []byte(tc.text)})
			c.Assert(directErr != nil, qt.Equals, tc.failed)
			c.Assert(checked.Result, qt.DeepEquals, direct)
		})
	}
}

func TestInvalidToolArgumentsCannotWeakenPolicy(t *testing.T) {
	cases := []map[string]any{
		{"sources": []any{}},
		{"sources": "not a list"},
		{"sources": []any{map[string]any{"name": "a.md", "text": "Clean prose."}}, "no_gate": true},
		{"sources": []any{map[string]any{"name": "a.md", "text": "Clean prose.", "contexts": []any{}}}},
	}
	for _, input := range cases {
		t.Run("arguments", func(t *testing.T) {
			c := qt.New(t)
			session := connect(c, t.Context(), server.Options{})
			result, err := session.CallTool(t.Context(), &mcp.CallToolParams{Name: "unswell_check", Arguments: input})
			c.Assert(err, qt.IsNil)
			c.Assert(result.IsError, qt.IsTrue)
		})
	}
}

func TestContextualClustersMatchThePublicEngine(t *testing.T) {
	for _, row := range []struct{ id, text string }{
		{"filler.section-announcement", "In this section, we will describe setup. In this section, we will describe deployment."},
		{"syntax.passive-candidate-density", "The request is carefully validated by the server before execution. " +
			"The response is securely recorded by the client after completion."},
		{"repetition.summary-echo", "The client opens a connection to the server and sends the request with its credentials.\n\n" +
			"## Summary\n\nThe client creates a connection to the server and sends the request with its credentials."},
	} {
		t.Run(row.id, func(t *testing.T) {
			c := qt.New(t)
			policy := []byte("version: 1\nextends: [builtin:custom]\nrules:\n  " + row.id + ": {enabled: true, gate: forbid}\n")
			session := connect(c, t.Context(), server.Options{Config: policy})
			input := server.CheckInput{Sources: []server.Source{{Name: "guide.md", Text: row.text}}}
			response, err := session.CallTool(t.Context(), &mcp.CallToolParams{Name: "unswell_check", Arguments: input})
			c.Assert(err, qt.IsNil)
			c.Assert(response.IsError, qt.IsFalse)
			checked := output[server.CheckOutput](c, response)
			c.Assert(checked.Outcome, qt.Equals, "policy_failure")
			c.Assert(checked.Result.Findings, qt.HasLen, 1)
			c.Assert(checked.Result.Findings[0].Related, qt.HasLen, 1)
			engine, err := unswell.New(unswell.Options{Config: policy})
			c.Assert(err, qt.IsNil)
			direct, err := engine.Analyze(t.Context(), document.Source{Name: "guide.md", Format: document.Markdown, Bytes: []byte(row.text)})
			c.Assert(err, qt.IsNil)
			c.Assert(checked.Result, qt.DeepEquals, direct)
		})
	}
}

type waitingNLP struct {
	nlp.Provider
	started  chan struct{}
	canceled chan struct{}
}

func (p waitingNLP) Analyze(ctx context.Context, _ document.MappedText, _ []nlp.Capability) ([]document.Sentence, error) {
	close(p.started)
	<-ctx.Done()
	close(p.canceled)
	return nil, ctx.Err()
}

func TestClientCancellationReachesEngine(t *testing.T) {
	c := qt.New(t)
	backend, err := english.New()
	c.Assert(err, qt.IsNil)
	provider := waitingNLP{Provider: backend, started: make(chan struct{}), canceled: make(chan struct{})}
	session := connect(c, t.Context(), server.Options{NLP: provider})
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	finished := make(chan error, 1)
	go func() {
		_, callErr := session.CallTool(ctx, &mcp.CallToolParams{Name: "unswell_check", Arguments: server.CheckInput{
			Sources: []server.Source{{Name: "draft.md", Text: "The client opens connections."}},
		}})
		finished <- callErr
	}()
	awaitSignal(c, provider.started)
	cancel()
	awaitSignal(c, provider.canceled)
	c.Assert(<-finished, qt.ErrorIs, context.Canceled)
}

func awaitSignal(c *qt.C, channel <-chan struct{}) {
	c.Helper()
	select {
	case <-channel:
	case <-time.After(5 * time.Second):
		c.Fatal("timed out waiting for MCP cancellation lifecycle")
	}
}

func TestContextOverrideAndReasonedException(t *testing.T) {
	c := qt.New(t)
	policy := "version: 1\nextends: [builtin:strict-v1]\nextraction:\n  contexts: [comment]\n" +
		"  languages: {yaml: {contexts: [comment, string]}}\n  exceptions:\n" +
		"    - {id: fixture, paths: ['*.yaml'], kinds: [string], symbols: [fixture], reason: Deliberate fixture.}\n"
	session := connect(c, t.Context(), server.Options{Config: []byte(policy)})
	input := server.CheckInput{
		Sources: []server.Source{
			{
				Name: "draft.yaml",
				Text: "# The client opens connections.\nfixture: Certainly! The client opens connections.\nmessage: Runtime prose.\n",
			},
		},
	}
	result, err := session.CallTool(t.Context(), &mcp.CallToolParams{Name: "unswell_check", Arguments: input})
	c.Assert(err, qt.IsNil)
	checked := output[server.CheckOutput](c, result)
	c.Assert(checked.Outcome, qt.Equals, "pass")
	var found bool
	for _, excluded := range checked.Result.Documents[0].Excluded {
		found = found || strings.HasPrefix(excluded.Reason, "config:fixture:")
	}
	c.Assert(found, qt.IsTrue)
}
