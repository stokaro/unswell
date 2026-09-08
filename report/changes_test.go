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

func TestSavedChangesRenderWithoutGitOrReanalysis(t *testing.T) {
	c := qt.New(t)
	engine, err := unswell.New(unswell.Options{})
	c.Assert(err, qt.IsNil)
	source := []document.Source{{Name: "guide.md", Format: document.Markdown, Bytes: []byte("Certainly! The client retries.")}}
	result, err := engine.AnalyzeChanged(t.Context(), source, source)
	c.Assert(err, qt.IsNil)
	result.Manifest.Git = &unswell.GitSelection{RequestedRef: "main", BaseCommit: strings.Repeat("a", 40),
		HeadCommit: strings.Repeat("b", 40), Clean: true}
	var saved bytes.Buffer
	c.Assert(report.Write(&saved, "json", result, report.Options{}), qt.IsNil)
	loaded, err := report.Read(bytes.NewReader(saved.Bytes()))
	c.Assert(err, qt.IsNil)
	c.Assert(loaded, qt.DeepEquals, result)
	for _, format := range []string{"text", "markdown", "html", "sarif"} {
		var output bytes.Buffer
		c.Assert(report.Write(&output, format, loaded, report.Options{}), qt.IsNil)
		c.Assert(output.String(), qt.Contains, "unchanged")
		c.Assert(output.String(), qt.Contains, result.Manifest.Git.BaseCommit)
		c.Assert(output.String(), qt.Not(qt.Contains), "Certainly! The client retries.")
	}
	invalid := bytes.Replace(saved.Bytes(), []byte(`"clean": true`), []byte(`"clean": false`), 1)
	c.Assert(bytes.Equal(invalid, saved.Bytes()), qt.IsFalse)
	_, err = report.Read(bytes.NewReader(invalid))
	c.Assert(err, qt.IsNotNil)
}
