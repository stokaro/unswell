package report_test

import (
	"bytes"
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/report"
)

func TestNearSentenceSavedVersion(t *testing.T) {
	c := qt.New(t)
	policy := []byte("version: 1\nextends: [builtin:custom]\nrules:\n" +
		"  repetition.near-sentence: {enabled: true}\n")
	engine, err := unswell.New(unswell.Options{Config: policy, CollectBaseline: true})
	c.Assert(err, qt.IsNil)
	first := "The client opens a connection to the server and sends the request with its credentials."
	text := first + "\n\n" + strings.Replace(first, "opens", "creates", 1)
	result, err := engine.Analyze(t.Context(), document.Source{Name: "guide.md", Format: document.Markdown, Bytes: []byte(text)})
	c.Assert(err, qt.IsNil)
	c.Assert(result.Findings, qt.HasLen, 1)
	c.Assert(result.Findings[0].RuleVersion, qt.Equals, "2")
	var encoded bytes.Buffer
	c.Assert(report.Write(&encoded, "json", result, report.Options{}), qt.IsNil)
	saved, err := report.Read(&encoded)
	c.Assert(err, qt.IsNil)
	c.Assert(saved.Findings, qt.DeepEquals, result.Findings)
	c.Assert(saved.BaselineSnapshot, qt.DeepEquals, result.BaselineSnapshot)
}
