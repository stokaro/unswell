package report_test

import (
	"bytes"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/baseline"
	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/report"
)

func TestSavedBaselineRendersWithoutReanalysis(t *testing.T) {
	c := qt.New(t)
	source := document.Source{Name: "guide.md", Format: document.Markdown, Bytes: []byte("Certainly! The client retries.")}
	capture, err := unswell.New(unswell.Options{CollectBaseline: true})
	c.Assert(err, qt.IsNil)
	raw, err := capture.Analyze(t.Context(), source)
	c.Assert(err, qt.IsNil)
	file, err := baseline.Create(t.Context(), *raw.BaselineSnapshot)
	c.Assert(err, qt.IsNil)
	data, err := baseline.Encode(t.Context(), file)
	c.Assert(err, qt.IsNil)
	engine, err := unswell.New(unswell.Options{Baseline: data, GateMode: "new"})
	c.Assert(err, qt.IsNil)
	result, err := engine.Analyze(t.Context(), source)
	c.Assert(err, qt.IsNil)
	var saved bytes.Buffer
	c.Assert(report.Write(&saved, "json", result, report.Options{}), qt.IsNil)
	loaded, err := report.Read(&saved)
	c.Assert(err, qt.IsNil)
	c.Assert(loaded, qt.DeepEquals, result)
	for _, format := range []string{"text", "markdown", "html", "sarif"} {
		var output bytes.Buffer
		c.Assert(report.Write(&output, format, loaded, report.Options{}), qt.IsNil)
		if format == "sarif" {
			c.Assert(output.String(), qt.Contains, `"baselineState": "unchanged"`)
			c.Assert(output.String(), qt.Contains, baseline.FingerprintVersion)
		} else {
			c.Assert(output.String(), qt.Contains, "existing")
		}
		c.Assert(output.String(), qt.Not(qt.Contains), "Certainly! The client retries.")
	}
}
