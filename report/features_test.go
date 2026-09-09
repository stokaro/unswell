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

func featureReport(t *testing.T) unswell.RunResult {
	t.Helper()
	c := qt.New(t)
	engine, err := unswell.New(unswell.Options{Features: []string{"prose-words", "type-token-ratio"}})
	c.Assert(err, qt.IsNil)
	result, err := engine.Analyze(t.Context(), document.Source{Name: "<img src=x>.md", Format: document.Markdown,
		Bytes: []byte("# Heading\n\nThe cache expires.\n")})
	c.Assert(err, qt.IsNil)
	return result
}

func TestFeatureReportsRoundTripWithoutSourceOrModel(t *testing.T) {
	c := qt.New(t)
	result := featureReport(t)
	var saved bytes.Buffer
	c.Assert(report.Write(&saved, "json", result, report.Options{}), qt.IsNil)
	c.Assert(saved.String(), qt.Contains, `"number": null`)
	c.Assert(saved.String(), qt.Not(qt.Contains), "The cache expires.")
	c.Assert(saved.String(), qt.Not(qt.Contains), "Heading")
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
		if format != "markdown" {
			c.Assert(output.String(), qt.Contains, "unsupported_unit")
		}
	}
}

func TestFeatureReportsRejectIncompatibleOrMisleadingValues(t *testing.T) {
	c := qt.New(t)
	data, err := json.Marshal(featureReport(t))
	c.Assert(err, qt.IsNil)
	for _, row := range []struct {
		name string
		edit func(*unswell.RunResult)
	}{
		{"contract", func(r *unswell.RunResult) { r.Features.Version = "future" }},
		{"unknown feature", func(r *unswell.RunResult) { r.Features.Requested[0] = "unknown" }},
		{"duplicate request", func(r *unswell.RunResult) { r.Features.Requested[1] = r.Features.Requested[0] }},
		{"uncollected source", func(r *unswell.RunResult) { r.Features.Sources = nil }},
		{"unknown source", func(r *unswell.RunResult) { r.Features.Sources[0].Path = "elsewhere.md" }},
		{"source identity", func(r *unswell.RunResult) { r.Features.Sources[0].SourceHash = "wrong" }},
		{"missing capability", func(r *unswell.RunResult) { r.Features.Sources[0].Capabilities = nil }},
		{"missing block", func(r *unswell.RunResult) { r.Features.Sources[0].Units = nil }},
		{"context text", func(r *unswell.RunResult) { r.Features.Sources[0].Units[0].ContextHash = "Heading" }},
		{"block contract", func(r *unswell.RunResult) { r.Features.Sources[0].Units[0].Binding.Contract = "future" }},
		{"block text hash", func(r *unswell.RunResult) { r.Features.Sources[0].Units[0].Binding.TextSHA256 = "invalid" }},
		{"trimmed text hash", func(r *unswell.RunResult) { r.Features.Sources[0].Units[0].Binding.TrimmedSHA256 = "invalid" }},
		{"block map", func(r *unswell.RunResult) { r.Features.Sources[0].Units[0].Binding.Segments[0].Start = -1 }},
		{"trimmed map", func(r *unswell.RunResult) { r.Features.Sources[0].Units[0].Binding.TrimmedSegments[0].Start = -1 }},
		{"empty block segment", func(r *unswell.RunResult) {
			s := &r.Features.Sources[0].Units[0].Binding.Segments[0]
			s.End = s.Start
		}},
		{"invalid range", func(r *unswell.RunResult) { r.Features.Sources[0].Units[1].Segments[0].End = 999 }},
		{"wrong definition", func(r *unswell.RunResult) { r.Features.Sources[0].Units[1].Values[0].Version = "99" }},
		{"missing value", func(r *unswell.RunResult) { r.Features.Sources[0].Units[1].Values[0].Number = nil }},
		{"number and absence", func(r *unswell.RunResult) { r.Features.Sources[0].Units[1].Values[0].Reason = "unsupported_unit" }},
		{"unsupported zero", func(r *unswell.RunResult) {
			zero := float64(0)
			r.Features.Sources[0].Units[0].Values[0].Number = &zero
		}},
	} {
		t.Run(row.name, func(t *testing.T) {
			c := qt.New(t)
			var result unswell.RunResult
			c.Assert(json.Unmarshal(data, &result), qt.IsNil)
			row.edit(&result)
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

func TestFeatureReportsRejectNonfiniteNumbers(t *testing.T) {
	c := qt.New(t)
	result := featureReport(t)
	for _, number := range []float64{math.NaN(), math.Inf(1)} {
		result.Features.Sources[0].Units[1].Values[0].Number = &number
		for _, format := range []string{"json", "sarif", "text", "markdown", "html"} {
			var output bytes.Buffer
			c.Assert(report.Write(&output, format, result, report.Options{}), qt.IsNotNil)
			c.Assert(output.Len(), qt.Equals, 0)
		}
	}
}
