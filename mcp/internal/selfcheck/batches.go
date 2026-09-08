package selfcheck

import (
	"context"
	"fmt"
	"reflect"
	"slices"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/mcp/internal/server"
)

func verifyBatches(ctx context.Context, session *mcp.ClientSession, expected unswell.RunResult) ([]server.CheckOutput, error) {
	if !batchableEvidence(expected) {
		return nil, fmt.Errorf("MCP self-check batches require a full scan without baseline, changed-unit, or bypass policies")
	}
	var results []server.CheckOutput
	for documents := range slices.Chunk(expected.Documents, server.MaxSources) {
		input := server.CheckInput{Sources: make([]server.Source, 0, len(documents))}
		for _, doc := range documents {
			input.Sources = append(input.Sources, server.Source{Name: doc.Name, Format: doc.Format, Text: doc.Source})
		}
		checked, err := check(ctx, session, input, "pass")
		if err != nil {
			return nil, fmt.Errorf("MCP repository batch %d: %w", len(results)+1, err)
		}
		if !reflect.DeepEqual(normalize(expectedBatch(expected, documents)), checked.Result) {
			return nil, fmt.Errorf("MCP repository batch %d differs from normalized CLI evidence", len(results)+1)
		}
		results = append(results, checked)
	}
	return results, nil
}

func batchableEvidence(result unswell.RunResult) bool {
	return result.Baseline == nil && result.BaselineSnapshot == nil && result.Changes == nil && result.PolicyComparison == nil &&
		!result.Manifest.NoGate && len(result.Gate.Accepted) == 0 && len(result.Gate.Unchanged) == 0
}

func expectedBatch(expected unswell.RunResult, documents []unswell.DocumentResult) unswell.RunResult {
	paths := make(map[string]bool, len(documents))
	for _, doc := range documents {
		paths[doc.Name] = true
	}
	expected.Documents = documents
	expected.Findings = slices.DeleteFunc(slices.Clone(expected.Findings), func(f unswell.Finding) bool { return !paths[f.Primary.Path] })
	expected.Assessments = slices.DeleteFunc(slices.Clone(expected.Assessments), func(a unswell.Assessment) bool { return !paths[a.Path] })
	expected.Suppressions = slices.DeleteFunc(slices.Clone(expected.Suppressions), func(s unswell.Suppression) bool {
		return !paths[s.Directive.Path]
	})
	if len(expected.Suppressions) == 0 {
		expected.Suppressions = nil // Empty suppressions are omitted by the wire schema.
	}
	expected.Features = batchFeatures(expected.Features, paths)
	return expected
}

func batchFeatures(collection *unswell.FeatureCollection, paths map[string]bool) *unswell.FeatureCollection {
	if collection == nil {
		return nil
	}
	batch := *collection
	batch.Sources = slices.DeleteFunc(slices.Clone(collection.Sources), func(source unswell.FeatureSource) bool {
		return !paths[source.Path]
	})
	return &batch
}
