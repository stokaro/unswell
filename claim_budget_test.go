package unswell_test

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/document"
)

func TestRepeatedClaimBudgetOnLongReference(t *testing.T) {
	c := qt.New(t)
	source := retainedMigrationReference(t)
	analyze := func(limit int) unswell.RunResult {
		policy := fmt.Sprintf("version: 1\nextends: [builtin:custom]\nanalysis: {max_candidates: %d}\n"+
			"rules:\n  repetition.repeated-claim: {enabled: true}\n", limit)
		engine, err := unswell.New(unswell.Options{Config: []byte(policy), Features: []string{"activation/repetition.repeated-claim"}})
		c.Assert(err, qt.IsNil)
		result, err := engine.Analyze(t.Context(), document.Source{Name: "migration.md", Format: document.Markdown, Bytes: source})
		c.Assert(err, qt.IsNil)
		c.Assert(result.Manifest.Complete, qt.IsTrue)
		c.Assert(result.Abstentions, qt.HasLen, 0)
		return result
	}
	bounded, reference := analyze(100000), analyze(1000000)
	c.Assert(bounded.Findings, qt.DeepEquals, reference.Findings)
	c.Assert(bounded.Features.Sources, qt.HasLen, 1)
	c.Assert(len(bounded.Features.Sources[0].Units) > 0, qt.IsTrue)
	c.Assert(reference.Features.Sources, qt.HasLen, 1)
	c.Assert(bounded.Features.Sources[0].Units, qt.HasLen, len(reference.Features.Sources[0].Units))
	measured := 0
	for i, unit := range bounded.Features.Sources[0].Units {
		// Input identities include the budget policy. Compare the observations and
		// source ownership, not hashes that intentionally describe different inputs.
		referenceUnit := reference.Features.Sources[0].Units[i]
		unit.InputHash = referenceUnit.InputHash
		c.Assert(unit, qt.DeepEquals, referenceUnit)
		for _, value := range unit.Values {
			c.Assert(value.Reason, qt.Not(qt.Contains), "budget_exhausted")
			if value.Number != nil {
				measured++
			}
		}
	}
	c.Assert(measured > 0, qt.IsTrue)
}

func retainedMigrationReference(t *testing.T) []byte {
	t.Helper()
	c := qt.New(t)
	file, err := os.Open("research/reviews/2026-09-17-instruction-recall/confirmation/inputs.tar.gz")
	c.Assert(err, qt.IsNil)
	t.Cleanup(func() { c.Check(file.Close(), qt.IsNil) })
	compressed, err := gzip.NewReader(file)
	c.Assert(err, qt.IsNil)
	t.Cleanup(func() { c.Check(compressed.Close(), qt.IsNil) })
	archive := tar.NewReader(compressed)
	for {
		header, err := archive.Next()
		c.Assert(err, qt.IsNil)
		if header.Name != "sources/c05.md" {
			continue
		}
		c.Assert(header.Size, qt.Equals, int64(97753))
		source, err := io.ReadAll(archive)
		c.Assert(err, qt.IsNil)
		hash := sha256.Sum256(source)
		c.Assert(hex.EncodeToString(hash[:]), qt.Equals, "330fb59a5fe90f3659d64d53cff20c403bdfe5fb29325c53c7a6f4d67905ed1a")
		return source
	}
}

func TestRepeatedClaimIdentityPreservesMappedText(t *testing.T) {
	for _, row := range []struct {
		name, left, right string
		want              int
	}{
		{"emphasis", "The **client** retries the request.", "The client retries the request.", 1},
		{"whitespace", "The client retries  the request.", "The client retries the request.", 1},
		{"unicode", "The café serves the visitors.", "The café serves the visitors.", 1},
		{"case", "The café serves the visitors.", "The Café serves the visitors.", 0},
		{"entity", "The client retrieves the café&#39;s value.", "The client retrieves the café's value.", 1},
		{"hidden link", "The client retrieves [the value](one).", "The client retrieves [the value](two).", 0},
		{"opaque", "The client retrieves the `value`.", "The client retrieves the value.", 0},
		{"operand", "The client retrieves the `a-b` value.", "The client retrieves the `a.b` value.", 0},
	} {
		t.Run(row.name, func(t *testing.T) {
			c := qt.New(t)
			text := "\ufeff" + row.left + "\r\n\r\n" + row.right
			result := singleRuleResult(t, "repetition.repeated-claim", text, "", "")
			c.Assert(result.Abstentions, qt.HasLen, 0)
			c.Assert(result.Findings, qt.HasLen, row.want)
			for _, finding := range result.Findings {
				c.Assert(text[finding.Primary.Span.Start:finding.Primary.Span.End], qt.Equals, finding.Primary.Snippet)
				c.Assert(text, qt.Contains, finding.Related[0].Snippet)
			}
		})
	}
}
