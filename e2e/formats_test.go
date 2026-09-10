package e2e_test

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/extract"
	"github.com/stokaro/unswell/report"
)

// The native job runs this package on Linux, macOS and Windows, so a format
// covered here is covered on every supported platform. A new input format that
// nobody scanned end to end fails this test instead of shipping unverified.
func TestEverySupportedInputFormatHasAScenario(t *testing.T) {
	c := qt.New(t)
	entries, err := os.ReadDir("testdata")
	c.Assert(err, qt.IsNil)
	scanned := map[document.Format][]string{}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		var spec scenario
		path := filepath.Join("testdata", entry.Name(), "case.json")
		data, err := fs.ReadFile(os.DirFS(filepath.Dir(path)), filepath.Base(path))
		c.Assert(err, qt.IsNil)
		c.Assert(json.Unmarshal(data, &spec), qt.IsNil)
		for _, name := range spec.Files {
			format, ok := explicitFormat(spec.Args)
			if !ok {
				format, ok = extract.Detect(name, nil)
			}
			if ok {
				scanned[format] = append(scanned[format], entry.Name())
			}
		}
	}
	for _, format := range document.Formats() {
		c.Assert(scanned[format], qt.Not(qt.HasLen), 0, qt.Commentf("no scenario scans %s", format))
	}
}

func explicitFormat(args []string) (document.Format, bool) {
	index := slices.Index(args, "--format")
	if index < 0 || index+1 >= len(args) {
		return "", false
	}
	return document.Format(args[index+1]), true
}

// One analysis produces every report format from the same result. A writer that
// only works when it is the single requested output would pass the per-format
// tests and fail a real run.
func TestOneAnalysisWritesEveryReportFormat(t *testing.T) {
	c := qt.New(t)
	binary := buildCLI(t)
	workspace := t.TempDir()
	source, err := os.ReadFile(filepath.Join("testdata", "markdown", "sample.md.txt"))
	c.Assert(err, qt.IsNil)
	policy, err := os.ReadFile(filepath.Join("testdata", "policy.yaml"))
	c.Assert(err, qt.IsNil)
	for name, data := range map[string][]byte{"sample.md": source, "policy.yaml": policy} {
		// #nosec G703 -- Each name is a fixed literal inside t.TempDir; fixture bytes never enter the path.
		c.Assert(os.WriteFile(filepath.Join(workspace, name), data, 0o600), qt.IsNil)
	}
	args := []string{"check", "--config", "policy.yaml", "--include-source"}
	for _, format := range report.Formats() {
		args = append(args, "--report", format+":result."+format)
	}
	stdout, stderr, code := invoke(t, binary, workspace, append(args, "sample.md"), nil)
	c.Assert(code, qt.Equals, 1, qt.Commentf("%s\n%s", stdout, stderr))

	written := os.DirFS(workspace)
	saved, err := written.Open("result.json")
	c.Assert(err, qt.IsNil)
	result, err := report.Read(saved)
	c.Assert(errors.Join(err, saved.Close()), qt.IsNil)
	c.Assert(result.Findings, qt.Not(qt.HasLen), 0)
	identifier := result.Findings[0].RuleID

	sarif, err := fs.ReadFile(written, "result.sarif")
	c.Assert(err, qt.IsNil)
	c.Assert(report.ValidateSARIF(sarif), qt.IsNil)
	// Markdown escapes punctuation, so each format is matched on a marker it
	// actually writes rather than on the raw rule identifier.
	for _, row := range []struct{ format, contains string }{
		{"text", identifier},
		{"markdown", "| Location | Rule | Finding |"},
		{"html", "<html"},
	} {
		data, err := fs.ReadFile(written, "result."+row.format)
		c.Assert(err, qt.IsNil)
		c.Assert(string(data), qt.Contains, row.contains, qt.Commentf("%s report", row.format))
	}
}
