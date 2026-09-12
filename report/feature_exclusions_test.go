package report_test

import (
	"bytes"
	"encoding/json"
	"html"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/report"
)

func excludedFeatureReport(t *testing.T) unswell.RunResult {
	t.Helper()
	c := qt.New(t)
	engine, err := unswell.New(unswell.Options{Features: []string{
		"activation/readability.long-paragraph", "prose-words", "type-token-ratio"},
		PreparedFeatures: []string{"prose-words"}, PreparedKinds: []string{"sentence", "paragraph"}})
	c.Assert(err, qt.IsNil)
	const prose = "Клиент повторяет запрос после сбоя транспорта и ждет ответа сервера."
	result, err := engine.Analyze(t.Context(), document.Source{Name: "mixed.md", Format: document.Markdown,
		Bytes: []byte("# " + prose + "\n\n" + prose + "\n\nThe client retries.\n")})
	c.Assert(err, qt.IsNil)
	return result
}

func TestExcludedFeaturesRoundTripAcrossReports(t *testing.T) {
	c := qt.New(t)
	result := excludedFeatureReport(t)
	var saved bytes.Buffer
	c.Assert(report.Write(&saved, "json", result, report.Options{}), qt.IsNil)
	c.Assert(saved.String(), qt.Contains, `"excluded": true`)
	c.Assert(saved.String(), qt.Contains, `"number": null`)
	c.Assert(saved.String(), qt.Not(qt.Contains), "Клиент")
	loaded, err := report.Read(&saved)
	c.Assert(err, qt.IsNil)
	c.Assert(loaded, qt.DeepEquals, result)
	for _, format := range []string{"text", "markdown", "html", "sarif"} {
		var output bytes.Buffer
		c.Assert(report.Write(&output, format, loaded, report.Options{}), qt.IsNil)
		c.Assert(html.UnescapeString(output.String()), qt.Contains, "excluded_unit")
	}
}

func TestFeatureReportsRejectFalseExclusionMeasurements(t *testing.T) {
	c := qt.New(t)
	data, err := json.Marshal(excludedFeatureReport(t))
	c.Assert(err, qt.IsNil)
	for name, edit := range map[string]func(*unswell.RunResult){
		"unrecorded exclusion": func(r *unswell.RunResult) { r.Documents[0].Excluded = nil },
		"missing marker":       func(r *unswell.RunResult) { r.Features.Sources[0].Units[1].Excluded = false },
		"false marker":         func(r *unswell.RunResult) { r.Features.Sources[0].Units[2].Excluded = true },
		"counted tokens": func(r *unswell.RunResult) {
			u := &r.Features.Sources[0].Units[1]
			u.Segments = []document.Span{u.Span}
		},
		"activation zero": func(r *unswell.RunResult) { r.Features.Sources[0].Units[1].Values[0].Number = new(float64) },
		"word count zero": func(r *unswell.RunResult) { r.Features.Sources[0].Units[1].Values[1].Number = new(float64) },
		"unrelated reason": func(r *unswell.RunResult) {
			r.Features.Sources[0].Units[1].Values[1].Reason = "no_prose_words"
		},
	} {
		t.Run(name, func(t *testing.T) {
			c := qt.New(t)
			var result unswell.RunResult
			c.Assert(json.Unmarshal(data, &result), qt.IsNil)
			edit(&result)
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
