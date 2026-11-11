package unswell_test

import (
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/document"
)

// A block the Markdown grammar rejects is excluded. The rest of the file keeps
// its findings, and the run is incomplete so the gate cannot pass.
func TestUnparsedBlockKeepsFindingsAndFailsTheGate(t *testing.T) {
	c := qt.New(t)
	engine, err := unswell.New(unswell.Options{})
	c.Assert(err, qt.IsNil)
	source := "It is important to note that the client retries.\n\n" +
		"| Target rule           | `<rule-name>`                             |\n" +
		"| Diagnostics           | `<total target-rule diagnostics>`         |\n"
	result, err := engine.AnalyzeAll(t.Context(), []document.Source{
		{Name: "guide.md", Format: document.Markdown, Bytes: []byte(source)},
	})
	c.Assert(err, qt.ErrorMatches, `.*1 block\(s\) could not be parsed and were excluded; the first starts at line 3`)
	c.Assert(result.Status, qt.Equals, "incomplete")
	c.Assert(result.Gate.Passed, qt.IsFalse)
	c.Assert(result.Findings, qt.HasLen, 1)
	c.Assert(result.Findings[0].RuleID, qt.Equals, "filler.announced-importance")
	c.Assert(result.Errors, qt.HasLen, 2)
	c.Assert(result.Errors[0].Path, qt.Equals, "guide.md")
}

// A document without a defect stays complete.
func TestParsedDocumentStaysComplete(t *testing.T) {
	c := qt.New(t)
	engine, err := unswell.New(unswell.Options{})
	c.Assert(err, qt.IsNil)
	result, err := engine.AnalyzeAll(t.Context(), []document.Source{
		{Name: "guide.md", Format: document.Markdown, Bytes: []byte("The client retries after a failure.\n")},
	})
	c.Assert(err, qt.IsNil)
	c.Assert(result.Status, qt.Equals, "complete")
}
