package selfcheck

// White-box tests: Exercise private batch planning with small frame budgets and invalid numeric evidence;
// the public process entry point fixes the SDK limit and rejects invalid saved reports before planning.

import (
	"context"
	"encoding/json"
	"math"
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/mcp/internal/server"
)

func TestBatchBudgetsBoundBothEncodedFrames(t *testing.T) {
	for _, largeRequest := range []bool{false, true} {
		t.Run(map[bool]string{false: "reply", true: "request"}[largeRequest], func(t *testing.T) {
			c := qt.New(t)
			doc := unswell.DocumentResult{Name: "first.md", Format: document.Markdown, Source: "A clean sentence."}
			if largeRequest {
				doc.Source = strings.Repeat("\"\\\n<&>", 3000)
			} else {
				doc.SourceHash = strings.Repeat("\"\\\n<&>", 3000)
			}
			expected := unswell.RunResult{Documents: []unswell.DocumentResult{doc, doc, doc}}
			expected.Documents[1].Name = "other.md"
			expected.Documents[2].Name = "third.md"
			requestSize, replySize, err := documentSizes(expected, 0)
			c.Assert(err, qt.IsNil)
			limit := frameEnvelopeBytes + 2*max(requestSize, 3*replySize)
			batches, err := planBatches(t.Context(), expected, limit)
			c.Assert(err, qt.IsNil)
			c.Assert(batches, qt.HasLen, 2)
			c.Assert(batches[0], qt.HasLen, 2)
			c.Assert(batches[1], qt.HasLen, 1)
			var names []string
			for _, batch := range batches {
				assertFramesFit(t, expected, batch, limit)
				for _, item := range batch {
					names = append(names, item.Name)
				}
			}
			c.Assert(names, qt.DeepEquals, []string{"first.md", "other.md", "third.md"})
			batches, err = planBatches(t.Context(), expected, limit-1)
			c.Assert(err, qt.IsNil)
			c.Assert(batches, qt.HasLen, 3)
			batches, err = planBatches(t.Context(), expected, frameEnvelopeBytes+max(requestSize, 3*replySize)-1)
			c.Assert(err, qt.ErrorMatches, `MCP self-check document "first.md" exceeds .* cannot split a document`)
			c.Assert(batches, qt.IsNil)
		})
	}
}

func assertFramesFit(t *testing.T, expected unswell.RunResult, documents []unswell.DocumentResult, limit int) {
	t.Helper()
	c := qt.New(t)
	input := server.CheckInput{}
	for _, doc := range documents {
		input.Sources = append(input.Sources, server.Source{Name: doc.Name, Format: doc.Format, Text: doc.Source})
	}
	request, err := json.Marshal(map[string]any{"jsonrpc": "2.0", "id": 1, "method": "tools/call",
		"params": mcp.CallToolParams{Name: "unswell_check", Arguments: input}})
	c.Assert(err, qt.IsNil)
	c.Assert(len(request)+1 <= limit, qt.IsTrue)
	output, err := json.Marshal(server.CheckOutput{Outcome: "pass", Result: normalize(expectedBatch(expected, documents))})
	c.Assert(err, qt.IsNil)
	result := mcp.CallToolResult{StructuredContent: json.RawMessage(output), Content: []mcp.Content{&mcp.TextContent{Text: string(output)}}}
	reply, err := json.Marshal(map[string]any{"jsonrpc": "2.0", "id": 1, "result": result})
	c.Assert(err, qt.IsNil)
	c.Assert(len(reply)+1 <= limit, qt.IsTrue)
}

func TestBatchPlanningRejectsCanceledOrInvalidEvidence(t *testing.T) {
	c := qt.New(t)
	expected := unswell.RunResult{Documents: []unswell.DocumentResult{{Name: "draft.md"}}}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	batches, err := planBatches(ctx, expected, mcp.DefaultMaxLineLength)
	c.Assert(err, qt.ErrorIs, context.Canceled)
	c.Assert(batches, qt.IsNil)
	expected.Documents[0].Maximum = math.NaN()
	batches, err = planBatches(t.Context(), expected, mcp.DefaultMaxLineLength)
	c.Assert(err, qt.ErrorMatches, `MCP self-check document "draft.md": json: unsupported value: NaN`)
	c.Assert(batches, qt.IsNil)
}

func TestBatchPlanningRetainsSourceCountLimit(t *testing.T) {
	c := qt.New(t)
	expected := unswell.RunResult{Documents: make([]unswell.DocumentResult, server.MaxSources+1)}
	batches, err := planBatches(t.Context(), expected, mcp.DefaultMaxLineLength)
	c.Assert(err, qt.IsNil)
	c.Assert(batches, qt.HasLen, 2)
	c.Assert(batches[0], qt.HasLen, server.MaxSources)
	c.Assert(batches[1], qt.HasLen, 1)
}
