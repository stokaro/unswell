package selfcheck

// White-box tests: Inject an in-memory MCP session into batch verification and corrupt later replies;
// the public self-check entry point starts a process and does not accept an existing session.

import (
	"context"
	"fmt"
	"testing"

	qt "github.com/frankban/quicktest"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/mcp/internal/server"
)

func batchFixture(t *testing.T) (*unswell.Engine, unswell.RunResult) {
	t.Helper()
	c := qt.New(t)
	engine, err := unswell.New(unswell.Options{IncludeSource: true, Features: []string{"prose-words", "type-token-ratio"},
		PreparedFeatures: []string{"prose-words"}, PreparedKinds: []string{"sentence", "paragraph"}})
	c.Assert(err, qt.IsNil)
	sources := make([]document.Source, server.MaxSources+1)
	for i := range sources {
		text := "The client opens connections."
		if i == len(sources)-1 {
			text = "<!-- unswell-disable-next-block filler.announced-importance -- Required external wording. -->\n" +
				"It is important to note that the client retries."
		}
		sources[i] = document.Source{Name: fmt.Sprintf("draft-%03d.md", i), Format: document.Markdown, Bytes: []byte(text)}
	}
	expected, err := engine.AnalyzeAll(t.Context(), sources)
	c.Assert(err, qt.IsNil)
	c.Assert(expected.Gate.Passed, qt.IsTrue)
	c.Assert(expected.Findings, qt.HasLen, 1)
	return engine, expected
}

func batchSession(t *testing.T, instance *mcp.Server) *mcp.ClientSession {
	t.Helper()
	c := qt.New(t)
	left, right := mcp.NewInMemoryTransports()
	serverSession, err := instance.Connect(t.Context(), left, nil)
	c.Assert(err, qt.IsNil)
	client := mcp.NewClient(&mcp.Implementation{Name: "batch-test", Version: "1"}, nil)
	session, err := client.Connect(t.Context(), right, nil)
	c.Assert(err, qt.IsNil)
	t.Cleanup(func() {
		c.Assert(session.Close(), qt.IsNil)
		c.Assert(serverSession.Wait(), qt.IsNil)
	})
	return session
}

func TestRepositoryBatchesPreserveEveryDocumentAndSuppression(t *testing.T) {
	c := qt.New(t)
	_, expected := batchFixture(t)
	instance, err := server.New(server.Options{Features: expected.Features.Requested,
		PreparedFeatures: expected.PreparedFeatures.Requested, PreparedKinds: expected.PreparedFeatures.Kinds})
	c.Assert(err, qt.IsNil)
	batches, err := verifyBatches(t.Context(), batchSession(t, instance), expected)
	c.Assert(err, qt.IsNil)
	c.Assert(batches, qt.HasLen, 2)
	c.Assert(batches[0].Result.Documents, qt.HasLen, server.MaxSources)
	c.Assert(batches[1].Result.Documents, qt.HasLen, 1)
	c.Assert(batches[1].Result.Findings, qt.HasLen, 1)
	c.Assert(batches[1].Result.Suppressions, qt.HasLen, 1)
	c.Assert(batches[1].Result.Findings[0].Suppressed, qt.IsTrue)
	c.Assert(batches[0].Result.Features.Sources, qt.HasLen, server.MaxSources)
	c.Assert(batches[1].Result.Features.Sources, qt.HasLen, 1)
	c.Assert(expected.Features.Sources, qt.HasLen, server.MaxSources+1)
	c.Assert(batches[0].Result.PreparedFeatures.Sources, qt.HasLen, server.MaxSources)
	c.Assert(batches[1].Result.PreparedFeatures.Sources, qt.HasLen, 1)
	c.Assert(expected.PreparedFeatures.Sources, qt.HasLen, server.MaxSources+1)
	c.Assert(expected.Documents[0].Source, qt.Equals, "The client opens connections.")
	c.Assert(expected.Findings[0].Primary.Snippet, qt.Not(qt.Equals), "")
	c.Assert(expected.Suppressions[0].Directive.Snippet, qt.Not(qt.Equals), "")
}

func TestRepositoryBatchesRejectLaterMismatch(t *testing.T) {
	for _, defect := range []string{"missing document", "changed manifest", "missing features", "changed feature value",
		"missing prepared features", "changed prepared value"} {
		t.Run(defect, func(t *testing.T) {
			c := qt.New(t)
			engine, expected := batchFixture(t)
			instance := mcp.NewServer(&mcp.Implementation{Name: "false-batch-fixture", Version: "1"}, nil)
			mcp.AddTool(instance, &mcp.Tool{Name: "unswell_check"},
				func(ctx context.Context, _ *mcp.CallToolRequest, input server.CheckInput) (*mcp.CallToolResult, server.CheckOutput, error) {
					var sources []document.Source
					for _, source := range input.Sources {
						sources = append(sources, document.Source{Name: source.Name, Format: source.Format, Bytes: []byte(source.Text)})
					}
					result, err := engine.AnalyzeAll(ctx, sources)
					if len(sources) == 1 {
						corruptBatch(&result, defect)
					}
					return nil, server.CheckOutput{Outcome: "pass", Result: normalize(result)}, err
				})
			batches, err := verifyBatches(t.Context(), batchSession(t, instance), expected)
			c.Assert(err, qt.ErrorMatches, "MCP repository batch 2 differs from normalized CLI evidence")
			c.Assert(batches, qt.HasLen, 0)
		})
	}
}

func corruptBatch(result *unswell.RunResult, defect string) {
	switch defect {
	case "missing document":
		result.Documents = nil
	case "changed manifest":
		result.Manifest.ConfigHash = "wrong-policy"
	case "missing features":
		result.Features = nil
	case "missing prepared features":
		result.PreparedFeatures = nil
	case "changed prepared value":
		*result.PreparedFeatures.Sources[0].Units[0].Values[0].Number += 1
	case "changed feature value":
		*result.Features.Sources[0].Units[0].Values[0].Number += 1
	}
}
