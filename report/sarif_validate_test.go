package report_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/report"
)

func sarifOf(c *qt.C, source document.Source, options unswell.Options) []byte {
	c.Helper()
	engine, err := unswell.New(options)
	c.Assert(err, qt.IsNil)
	result, err := engine.Analyze(c.Context(), source)
	c.Assert(err, qt.IsNil)
	var output bytes.Buffer
	c.Assert(report.Write(&output, "sarif", result, report.Options{}), qt.IsNil)
	return output.Bytes()
}

func TestSARIFReportsMatchThePublishedSchema(t *testing.T) {
	for _, row := range []struct {
		name string
		text string
	}{
		{"clean prose", "# Notes\n\nThe client retries after a transport failure.\n"},
		{"policy failures", "# Notes\n\nCertainly! It is important to note that this is not just fast, but also simple, " +
			"reliable, and powerful. Certainly! It is important to note that this is not just fast, but also simple, " +
			"reliable, and powerful.\n"},
		{"unicode and structure", "# Заголовок\n\nThe café client retries.\n\n- one item\n- two item\n\n| a | b |\n| --- | --- |\n| c | d |\n"},
	} {
		t.Run(row.name, func(t *testing.T) {
			c := qt.New(t)
			data := sarifOf(c, document.Source{Name: "guide.md", Format: document.Markdown, Bytes: []byte(row.text)},
				unswell.Options{IncludeSource: true})
			c.Assert(report.ValidateSARIF(data), qt.IsNil)
			var document struct {
				Schema  string `json:"$schema"`
				Version string `json:"version"`
			}
			c.Assert(json.Unmarshal(data, &document), qt.IsNil)
			c.Assert(document.Version, qt.Equals, report.SARIFSchemaVersion)
			c.Assert(document.Schema, qt.Equals, "https://json.schemastore.org/sarif-2.1.0.json")
		})
	}
}

func TestSARIFValidationRejectsBrokenDocuments(t *testing.T) {
	c := qt.New(t)
	valid := string(sarifOf(c, document.Source{Name: "guide.md", Format: document.Markdown,
		Bytes: []byte("# Notes\n\nCertainly! The client retries.\n")}, unswell.Options{}))
	for _, row := range []struct{ name, data, message string }{
		{"not json", "{", "(?s).*not valid JSON.*"},
		{"wrong version", strings.Replace(valid, `"version": "2.1.0"`, `"version": "2.0.0"`, 1),
			"(?s).*does not match the 2.1.0 schema.*value must be .2.1.0..*"},
		{"missing runs", strings.Replace(valid, `"runs"`, `"executions"`, 1),
			"(?s).*does not match the 2.1.0 schema.*missing property .runs..*"},
		{"invalid column kind", strings.Replace(valid, `"columnKind": "unicodeCodePoints"`, `"columnKind": "bytes"`, 1),
			"(?s).*does not match the 2.1.0 schema.*columnKind.*"},
	} {
		t.Run(row.name, func(t *testing.T) {
			c := qt.New(t)
			c.Assert(row.data, qt.Not(qt.Equals), valid)
			c.Assert(report.ValidateSARIF([]byte(row.data)), qt.ErrorMatches, row.message)
		})
	}
}

// The schema is validated without any network access, so a build without
// external resources still checks the published contract.
func TestSARIFValidationReadsNoExternalResource(t *testing.T) {
	c := qt.New(t)
	data := sarifOf(c, document.Source{Name: "guide.md", Format: document.Markdown,
		Bytes: []byte("# Notes\n\nThe client retries.\n")}, unswell.Options{})
	c.Assert(report.ValidateSARIF(data), qt.IsNil)
	c.Assert(string(data), qt.Contains, "https://json.schemastore.org/sarif-2.1.0.json")
}
