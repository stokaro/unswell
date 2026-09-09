package report_test

import (
	"bytes"
	"encoding/json"
	"math"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/report"
)

func preparedReport(t *testing.T) unswell.RunResult {
	t.Helper()
	c := qt.New(t)
	engine, err := unswell.New(unswell.Options{PreparedFeatures: []string{"prose-words", "type-token-ratio"},
		PreparedKinds: []string{"sentence", "paragraph", "fragment"}})
	c.Assert(err, qt.IsNil)
	result, err := engine.Analyze(t.Context(), document.Source{Name: "<img src=x>.md", Format: document.Markdown,
		Bytes: []byte("# Private heading\n\nThe **cache** expires.\n\n...\n")})
	c.Assert(err, qt.IsNil)
	return result
}

func TestPreparedReportsRoundTripWithoutSourceOrModel(t *testing.T) {
	c := qt.New(t)
	result := preparedReport(t)
	var saved bytes.Buffer
	c.Assert(report.Write(&saved, "json", result, report.Options{}), qt.IsNil)
	c.Assert(result.PreparedFeatures.Sources[0].Units, qt.HasLen, 3)
	c.Assert(saved.String(), qt.Not(qt.Contains), "Private heading")
	c.Assert(saved.String(), qt.Not(qt.Contains), "expires")
	loaded, err := report.Read(&saved)
	c.Assert(err, qt.IsNil)
	c.Assert(loaded, qt.DeepEquals, result)
	for _, format := range []string{"text", "markdown", "html", "sarif"} {
		var output bytes.Buffer
		c.Assert(report.Write(&output, format, loaded, report.Options{}), qt.IsNil)
		if format == "markdown" {
			c.Assert(output.String(), qt.Contains, "prose&#45;words")
		} else {
			c.Assert(output.String(), qt.Contains, "prose-words")
		}
		if format == "html" {
			c.Assert(output.String(), qt.Not(qt.Contains), "<img src=x>")
		}
	}
}

func TestPreparedReportsRejectInvalidBindingsAndValues(t *testing.T) {
	c := qt.New(t)
	data, err := json.Marshal(preparedReport(t))
	c.Assert(err, qt.IsNil)
	for _, row := range []struct {
		name string
		edit func(*unswell.PreparedFeatureCollection)
	}{
		{"contract", func(p *unswell.PreparedFeatureCollection) { p.Version = "future" }},
		{"kinds", func(p *unswell.PreparedFeatureCollection) { p.Kinds = []string{"unknown"} }},
		{"duplicate kinds", func(p *unswell.PreparedFeatureCollection) { p.Kinds[1] = p.Kinds[0] }},
		{"unknown feature", func(p *unswell.PreparedFeatureCollection) { p.Requested[0] = "unknown" }},
		{"missing source", func(p *unswell.PreparedFeatureCollection) { p.Sources = nil }},
		{"source identity", func(p *unswell.PreparedFeatureCollection) { p.Sources[0].SourceHash = "wrong" }},
		{"capabilities", func(p *unswell.PreparedFeatureCollection) { p.Sources[0].Capabilities = nil }},
		{"target binding", func(p *unswell.PreparedFeatureCollection) { p.Sources[0].Units[0].Binding.Contract = "future" }},
		{"target scope", func(p *unswell.PreparedFeatureCollection) { p.Sources[0].Units[0].Binding.Kind = "document" }},
		{"target range", func(p *unswell.PreparedFeatureCollection) { p.Sources[0].Units[0].Binding.Segments[0].End = 999 }},
		{"context range", func(p *unswell.PreparedFeatureCollection) { p.Sources[0].Units[1].Binding.ContextSpans = nil }},
		{"counted range", func(p *unswell.PreparedFeatureCollection) { p.Sources[0].Units[1].Segments[0].Start = 0 }},
		{"missing targets", func(p *unswell.PreparedFeatureCollection) { p.Sources[0].Units = nil }},
		{"missing values", func(p *unswell.PreparedFeatureCollection) { p.Sources[0].Units[0].Values = nil }},
		{"missing number", func(p *unswell.PreparedFeatureCollection) { p.Sources[0].Units[0].Values[0].Number = nil }},
		{"absence reason", func(p *unswell.PreparedFeatureCollection) { p.Sources[0].Units[0].Values[0].Reason = "disabled" }},
		{"duplicate target", func(p *unswell.PreparedFeatureCollection) {
			p.Sources[0].Units = append(p.Sources[0].Units, p.Sources[0].Units[0])
		}},
	} {
		t.Run(row.name, func(t *testing.T) {
			c := qt.New(t)
			var result unswell.RunResult
			c.Assert(json.Unmarshal(data, &result), qt.IsNil)
			row.edit(result.PreparedFeatures)
			var output bytes.Buffer
			c.Assert(report.Write(&output, "json", result, report.Options{}), qt.IsNotNil)
			c.Assert(output.Len(), qt.Equals, 0)
			altered, err := json.Marshal(result)
			c.Assert(err, qt.IsNil)
			_, err = report.Read(bytes.NewReader(altered))
			c.Assert(err, qt.IsNotNil)
		})
	}
}

func TestPreparedReportsRejectNonfiniteValues(t *testing.T) {
	c := qt.New(t)
	result := preparedReport(t)
	for _, number := range []float64{math.NaN(), math.Inf(1)} {
		result.PreparedFeatures.Sources[0].Units[0].Values[0].Number = &number
		for _, format := range []string{"json", "sarif", "text", "markdown", "html"} {
			var output bytes.Buffer
			c.Assert(report.Write(&output, format, result, report.Options{}), qt.IsNotNil)
			c.Assert(output.Len(), qt.Equals, 0)
		}
	}
}
