package selfcheck

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/mcp/internal/server"
)

// Reserve space for JSON-RPC IDs, tool names, content wrappers, and the newline.
const frameEnvelopeBytes = 4096

func planBatches(ctx context.Context, expected unswell.RunResult, frameLimit int) ([][]unswell.DocumentResult, error) {
	requestLimit := frameLimit - frameEnvelopeBytes
	// The SDK sends structured JSON and the same JSON as text. Escaping that text
	// can double its size, so three copies bound the combined response payload.
	replyLimit := requestLimit / 3
	var batches [][]unswell.DocumentResult
	start, requestBytes, replyBytes := 0, 0, 0
	for i, doc := range expected.Documents {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		requestSize, replySize, err := documentSizes(expected, i)
		if err != nil {
			return nil, fmt.Errorf("MCP self-check document %q: %w", doc.Name, err)
		}
		if requestSize > requestLimit || replySize > replyLimit {
			return nil, fmt.Errorf("MCP self-check document %q exceeds the %d-byte frame budget; cannot split a document", doc.Name, frameLimit)
		}
		if i-start == server.MaxSources || requestBytes+requestSize > requestLimit || replyBytes+replySize > replyLimit {
			batches = append(batches, expected.Documents[start:i])
			start, requestBytes, replyBytes = i, 0, 0
		}
		requestBytes += requestSize
		replyBytes += replySize
	}
	if start < len(expected.Documents) {
		batches = append(batches, expected.Documents[start:])
	}
	return batches, nil
}

func documentSizes(expected unswell.RunResult, index int) (requestSize, replySize int, err error) {
	doc := expected.Documents[index]
	request, err := json.Marshal(server.Source{Name: doc.Name, Format: doc.Format, Text: doc.Source})
	if err != nil {
		return 0, 0, err
	}
	// Count shared metadata once per document as a conservative upper bound.
	// Serialize individual results to avoid allocating an oversized batch first.
	reply, err := json.Marshal(server.CheckOutput{Outcome: "pass", Result: normalize(expectedBatch(expected,
		expected.Documents[index:index+1]))})
	if err != nil {
		return 0, 0, err
	}
	return len(request) + 1, len(reply), nil // Include the separator between sources.
}
