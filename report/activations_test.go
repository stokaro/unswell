package report_test

import (
	"bytes"
	"encoding/json"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/report"
)

func activationReport(t *testing.T) unswell.RunResult {
	t.Helper()
	c := qt.New(t)
	engine, err := unswell.New(unswell.Options{Features: []string{"activation/readability.long-paragraph"}, Config: []byte(
		"version: 1\nextends: [builtin:custom]\nrules:\n  readability.long-paragraph: {enabled: true}\n")})
	c.Assert(err, qt.IsNil)
	result, err := engine.Analyze(t.Context(), document.Source{Name: "guide.md", Format: document.Markdown,
		Bytes: []byte("The cache expires.")})
	c.Assert(err, qt.IsNil)
	return result
}

func TestActivationReportsRoundTripAndRenderAllFormats(t *testing.T) {
	c := qt.New(t)
	result := activationReport(t)
	var encoded bytes.Buffer
	c.Assert(report.Write(&encoded, "json", result, report.Options{}), qt.IsNil)
	loaded, err := report.Read(&encoded)
	c.Assert(err, qt.IsNil)
	c.Assert(loaded, qt.DeepEquals, result)
	c.Assert(*loaded.Features.Sources[0].Units[0].Values[0].Number, qt.Equals, float64(0))
	for _, format := range []string{"json", "sarif", "text", "markdown", "html"} {
		var output bytes.Buffer
		c.Assert(report.Write(&output, format, loaded, report.Options{}), qt.IsNil)
		c.Assert(output.String(), qt.Not(qt.Contains), "The cache expires.")
		if format != "markdown" {
			c.Assert(output.String(), qt.Contains, "unswell-rule-activations-v1")
		}
	}
}

func TestActivationReportsRejectContradictoryOrUnboundValues(t *testing.T) {
	c := qt.New(t)
	data, err := json.Marshal(activationReport(t))
	c.Assert(err, qt.IsNil)
	for _, row := range []struct {
		name string
		edit func(*unswell.RunResult)
	}{
		{"contract", func(r *unswell.RunResult) { r.Features.ActivationContract = "" }},
		{"catalog hash", func(r *unswell.RunResult) { r.Features.Sources[0].RulesetHash = "" }},
		{"unknown rule", func(r *unswell.RunResult) { r.Manifest.Rules = nil }},
		{"missing POS", func(r *unswell.RunResult) { r.Features.Sources[0].Capabilities = nil }},
		{"outside range", func(r *unswell.RunResult) { *r.Features.Sources[0].Units[0].Values[0].Number = 2 }},
		{"ambiguous zero", func(r *unswell.RunResult) { r.Features.Sources[0].Units[0].Values[0].Reason = "disabled" }},
		{"unfinished", func(r *unswell.RunResult) {
			v := &r.Features.Sources[0].Units[0].Values[0]
			v.Number, v.Reason = nil, "not_evaluated"
		}},
		{"missing reason", func(r *unswell.RunResult) { r.Features.Sources[0].Units[0].Values[0].Number = nil }},
		{"invalid reason", func(r *unswell.RunResult) {
			v := &r.Features.Sources[0].Units[0].Values[0]
			v.Number, v.Reason = nil, "inapplicable/source prose"
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
