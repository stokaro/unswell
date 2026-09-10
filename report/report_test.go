package report_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/report"
	"github.com/stokaro/unswell/rule"
)

func TestReportEscapingAndRoundTrip(t *testing.T) {
	c := qt.New(t)
	engine, err := unswell.New(unswell.Options{IncludeSource: true})
	c.Assert(err, qt.IsNil)
	result, err := engine.Analyze(
		t.Context(),
		document.Source{Name: "a<&>.txt", Format: document.Plain, Bytes: []byte("Certainly! <script>alert('source text');</script>")},
	)
	c.Assert(err, qt.IsNil)
	var html bytes.Buffer
	c.Assert(report.Write(&html, "html", result, report.Options{}), qt.IsNil)
	c.Assert(html.String(), qt.Not(qt.Contains), "<script>alert('source text');</script>")
	c.Assert(html.String(), qt.Contains, "&lt;script&gt;")
	var saved bytes.Buffer
	c.Assert(report.Write(&saved, "json", result, report.Options{}), qt.IsNil)
	restored, err := report.Read(&saved)
	c.Assert(err, qt.IsNil)
	c.Assert(restored, qt.DeepEquals, result)
}

type failedWriter struct{}

func (failedWriter) Write(_ []byte) (int, error) { return 0, errors.New("injected write failure") }

func TestEveryWriterReturnsErrors(t *testing.T) {
	c := qt.New(t)
	engine, err := unswell.New(unswell.Options{})
	c.Assert(err, qt.IsNil)
	result, err := engine.Analyze(
		t.Context(),
		document.Source{Name: "a.txt", Format: document.Plain, Bytes: []byte("A clean sentence.")},
	)
	c.Assert(err, qt.IsNil)
	for _, format := range report.Formats() {
		c.Check(report.Write(failedWriter{}, format, result, report.Options{}), qt.IsNotNil, qt.Commentf("%s", format))
	}
}

// Formats is the list callers iterate, so it must match what Write accepts. A
// writer added to the switch but not to the list would never be offered, and a
// name in the list without a writer would fail only at run time.
func TestFormatsMatchTheAcceptedWriters(t *testing.T) {
	c := qt.New(t)
	engine, err := unswell.New(unswell.Options{})
	c.Assert(err, qt.IsNil)
	result, err := engine.Analyze(
		t.Context(),
		document.Source{Name: "a.txt", Format: document.Plain, Bytes: []byte("Certainly! Let us dive into the details.")},
	)
	c.Assert(err, qt.IsNil)
	c.Assert(report.Formats(), qt.HasLen, 5)
	for _, format := range report.Formats() {
		var written bytes.Buffer
		c.Assert(report.Write(&written, format, result, report.Options{}), qt.IsNil, qt.Commentf("%s", format))
		c.Assert(written.Len() > 0, qt.IsTrue, qt.Commentf("%s", format))
	}
	c.Assert(report.Write(io.Discard, "yaml", result, report.Options{}), qt.ErrorMatches, `unknown report format "yaml"`)
}

func FuzzSavedResult(f *testing.F) {
	f.Add(`{"schema_version":"unknown"}`)
	engine, err := unswell.New(unswell.Options{Rules: []rule.Rule{}, IncludeSource: true})
	if err != nil {
		f.Fatal(err)
	}
	result, err := engine.Analyze(f.Context(), document.Source{
		Name: "fixture.txt", Format: document.Plain, Bytes: []byte("The client retries after a connection failure."),
	})
	if err != nil {
		f.Fatal(err)
	}
	seed, err := json.Marshal(result)
	if err != nil {
		f.Fatal(err)
	}
	f.Add(string(seed))
	f.Fuzz(func(t *testing.T, text string) {
		if len(text) > 256<<10 {
			t.Skip()
		}
		result, err := report.Read(strings.NewReader(text))
		if err != nil {
			return
		}
		var buffer bytes.Buffer
		c := qt.New(t)
		c.Check(report.Write(&buffer, "html", result, report.Options{}), qt.IsNil)
	})
}
