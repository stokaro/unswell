package report_test

import (
	"bytes"
	"html"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/config"
	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/report"
)

func TestSavedReportRetainsInheritedFilePolicies(t *testing.T) {
	c := qt.New(t)
	bundle := config.Bundle{Root: "policy.yaml", Files: map[string][]byte{
		"policy.yaml": []byte("version: 1\nextends: [base.yaml]\noverrides:\n" +
			"  - files: [reference.txt]\n    rules: {policy.banned-phrases: {enabled: false}}\n"),
		"base.yaml": []byte("version: 1\nextends: [builtin:custom]\n" +
			"rules: {policy.banned-phrases: {enabled: true, parameters: {phrases: [robust]}}}\n"),
	}}
	engine, err := unswell.New(unswell.Options{ConfigBundle: &bundle})
	c.Assert(err, qt.IsNil)
	result, err := engine.AnalyzeAll(t.Context(), []document.Source{
		{Name: "guide.txt", Format: document.Plain, Bytes: []byte("The robust client starts.")},
		{Name: "reference.txt", Format: document.Plain, Bytes: []byte("The robust client starts.")},
	})
	c.Assert(err, qt.IsNil)
	c.Assert(result.Findings, qt.HasLen, 1)
	c.Assert(result.Documents[0].ConfigHash, qt.Not(qt.Equals), result.Documents[1].ConfigHash)
	var saved bytes.Buffer
	c.Assert(report.Write(&saved, "json", result, report.Options{}), qt.IsNil)
	// Rendering must depend only on the saved result, even when its inputs are gone.
	clear(bundle.Files)
	restored, err := report.Read(&saved)
	c.Assert(err, qt.IsNil)
	c.Assert(restored, qt.DeepEquals, result)
	for _, format := range []string{"text", "json", "sarif", "html", "markdown"} {
		t.Run(format, func(t *testing.T) {
			c := qt.New(t)
			var output bytes.Buffer
			c.Assert(report.Write(&output, format, restored, report.Options{}), qt.IsNil)
			c.Assert(output.String(), qt.Contains, result.Manifest.ConfigHash)
			c.Assert(html.UnescapeString(output.String()), qt.Contains, "unswell-config-bundle-v1")
			if format == "json" || format == "sarif" || format == "html" {
				for _, doc := range result.Documents {
					c.Assert(output.String(), qt.Contains, doc.ConfigHash)
				}
			}
		})
	}
}
