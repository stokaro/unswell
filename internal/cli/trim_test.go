package cli_test

import (
	"encoding/json"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"
)

const trimDraft = "It is important to note that the client retries a request after a connection failure. " +
	"The client does not retry a request that the server rejected.\n"

// draftTrims proposes one pure deletion, one trim that drops a negation, and
// one for an item that does not exist.
const draftTrims = `{"trims": [
	{"item": "i001", "text": "The client retries a request after a connection failure."},
	{"item": "i002", "text": "The client does retry a request that the server rejected."},
	{"item": "i009", "text": "Nothing."}]}`

type trimReport struct {
	SchemaVersion string `json:"schema_version"`
	Protocol      string `json:"protocol"`
	Summary       struct {
		Documents int `json:"documents"`
		Failed    int `json:"failed"`
		Trims     int `json:"trims"`
		Rejected  int `json:"rejected"`
		Written   int `json:"written"`
	} `json:"summary"`
	Documents []struct {
		Path      string `json:"path"`
		Sentences int    `json:"sentences"`
		Proposed  int    `json:"proposed"`
		Written   bool   `json:"written"`
		Error     string `json:"error"`
		Trims     []struct {
			Original string   `json:"original"`
			Trimmed  string   `json:"trimmed"`
			Removed  []string `json:"removed"`
			Span     struct {
				Start int `json:"start"`
				End   int `json:"end"`
			} `json:"span"`
		} `json:"trims"`
		Rejected []struct {
			Problem string `json:"problem"`
		} `json:"rejected"`
	} `json:"documents"`
}

func trimAnswers(text string) func(chatCall, int) (int, string) {
	return func(chatCall, int) (int, string) { return http.StatusOK, text }
}

func TestTrimKeepsOnlyPureDeletions(t *testing.T) {
	c := qt.New(t)
	model := newFakeModel(t, trimAnswers(draftTrims))
	outcome := runCommand(t, "trim", t.TempDir(), trimDraft, "--stdin", "--filename", "draft.md", "--report", "json:-", "--quiet",
		"--endpoint", model.server.URL, "--model", "first")
	c.Assert(outcome.code, qt.Equals, 0, qt.Commentf("stderr: %s", outcome.stderr))
	var report trimReport
	c.Assert(json.Unmarshal([]byte(outcome.stdout), &report), qt.IsNil)
	c.Assert(report.SchemaVersion, qt.Equals, "unswell-trim-v1")
	c.Assert(report.Protocol, qt.Equals, "unswell-trim-v1")
	c.Assert(report.Summary.Trims, qt.Equals, 1)
	c.Assert(report.Summary.Rejected, qt.Equals, 2)
	document := report.Documents[0]
	c.Assert(document.Sentences, qt.Equals, 2)
	c.Assert(document.Proposed, qt.Equals, 3)
	trim := document.Trims[0]
	c.Assert(trimDraft[trim.Span.Start:trim.Span.End], qt.Equals, trim.Original)
	c.Assert(trim.Trimmed, qt.Equals, "The client retries a request after a connection failure.")
	c.Assert(trim.Removed, qt.DeepEquals, []string{"it", "is", "important", "to", "note", "that"})
	c.Assert(document.Rejected[0].Problem, qt.Equals, "removes a negation or a modal word")
	c.Assert(document.Rejected[1].Problem, qt.Equals, "unknown item")
	calls := model.received()
	c.Assert(calls, qt.HasLen, 1)
	c.Assert(calls[0].User, qt.Contains, "===== SENTENCES =====\ni001: It is important to note")
}

func TestTrimTextReportShowsBeforeAndAfter(t *testing.T) {
	c := qt.New(t)
	model := newFakeModel(t, trimAnswers(draftTrims))
	root := t.TempDir()
	c.Assert(os.WriteFile(filepath.Join(root, "draft.md"), []byte(trimDraft), 0o600), qt.IsNil)
	outcome := runCommand(t, "trim", root, "", "draft.md", "--endpoint", model.server.URL, "--model", "first")
	c.Assert(outcome.code, qt.Equals, 0, qt.Commentf("stderr: %s", outcome.stderr))
	host := strings.TrimPrefix(model.server.URL, "http://")
	c.Assert(outcome.stdout, qt.Equals, ""+
		"draft.md:1:1: trim\n"+
		"    - It is important to note that the client retries a request after a connection failure.\n"+
		"    + The client retries a request after a connection failure.\n"+
		"1 trim in 1 document. 2 proposals did not pass the check.\n"+
		"Model: first at "+host+".\n"+
		"Nothing was written. Use --write to apply these trims.\n"+
		"A trim only deletes words. A deleted word can still carry meaning that the check cannot see.\n")
	c.Assert(outcome.stderr, qt.Contains, "trim: draft.md: 1 trim, 2 proposals rejected\n")
	// Without --write the file is untouched.
	data, err := fs.ReadFile(os.DirFS(root), "draft.md")
	c.Assert(err, qt.IsNil)
	c.Assert(string(data), qt.Equals, trimDraft)
}

func TestTrimWritesAcceptedTrimsOnRequest(t *testing.T) {
	c := qt.New(t)
	model := newFakeModel(t, trimAnswers(draftTrims))
	root := t.TempDir()
	path := filepath.Join(root, "draft.md")
	// A read-only source keeps its mode after it is replaced.
	c.Assert(os.WriteFile(path, []byte(trimDraft), 0o400), qt.IsNil)
	outcome := runCommand(t, "trim", root, "", "draft.md", "--write", "--quiet", "--endpoint", model.server.URL, "--model", "first")
	c.Assert(outcome.code, qt.Equals, 0, qt.Commentf("stderr: %s", outcome.stderr))
	c.Assert(outcome.stdout, qt.Contains, "Written to 1 file. Review the change before you keep it.\n")
	data, err := fs.ReadFile(os.DirFS(root), "draft.md")
	c.Assert(err, qt.IsNil)
	c.Assert(string(data), qt.Equals, "The client retries a request after a connection failure. "+
		"The client does not retry a request that the server rejected.\n")
	if runtime.GOOS != "windows" {
		info, err := os.Stat(path)
		c.Assert(err, qt.IsNil)
		c.Assert(info.Mode().Perm(), qt.Equals, os.FileMode(0o400))
	}
	entries, err := os.ReadDir(root)
	c.Assert(err, qt.IsNil)
	c.Assert(entries, qt.HasLen, 1)
}

func TestTrimDoesNotWriteAFileThatChanged(t *testing.T) {
	c := qt.New(t)
	root := t.TempDir()
	path := filepath.Join(root, "draft.md")
	c.Assert(os.WriteFile(path, []byte(trimDraft), 0o600), qt.IsNil)
	model := newFakeModel(t, func(chatCall, int) (int, string) {
		// Someone edits the file while the model answers.
		if err := os.WriteFile(path, []byte(trimDraft+"A new line.\n"), 0o600); err != nil {
			return http.StatusInternalServerError, err.Error()
		}
		return http.StatusOK, draftTrims
	})
	outcome := runCommand(t, "trim", root, "", "draft.md", "--write", "--quiet", "--report", "json:-",
		"--endpoint", model.server.URL, "--model", "first")
	c.Assert(outcome.code, qt.Equals, 2)
	var report trimReport
	c.Assert(json.Unmarshal([]byte(outcome.stdout), &report), qt.IsNil)
	c.Assert(report.Documents[0].Error, qt.Equals, "the file changed during the run and was not written")
	c.Assert(report.Documents[0].Written, qt.IsFalse)
	data, err := fs.ReadFile(os.DirFS(root), "draft.md")
	c.Assert(err, qt.IsNil)
	c.Assert(string(data), qt.Equals, trimDraft+"A new line.\n")
}

func TestTrimFailsClosed(t *testing.T) {
	cases := []struct {
		name    string
		answer  func(chatCall, int) (int, string)
		args    []string
		calls   int
		problem string
	}{
		{"refusal", trimAnswers("I would not change this text."), []string{"--attempts", "2"}, 2, "the answer contains no JSON object"},
		{"other object", trimAnswers(`{"flags": []}`), []string{"--attempts", "1"}, 1, "the answer has no trims field"},
		{"rejected key", func(chatCall, int) (int, string) { return http.StatusUnauthorized, "invalid key" }, nil, 1, "answered 401"},
		{"document over the limit", trimAnswers(draftTrims), []string{"--max-document-bytes", "20"}, 0, "--max-document-bytes is 20"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := qt.New(t)
			model := newFakeModel(t, tc.answer)
			args := append([]string{"--stdin", "--filename", "draft.md", "--report", "json:-", "--quiet",
				"--endpoint", model.server.URL, "--model", "first"}, tc.args...)
			outcome := runCommand(t, "trim", t.TempDir(), trimDraft, args...)
			c.Assert(outcome.code, qt.Equals, 2)
			c.Assert(outcome.stderr, qt.Contains, "the run is incomplete: 1 of 1 documents failed")
			var report trimReport
			c.Assert(json.Unmarshal([]byte(outcome.stdout), &report), qt.IsNil)
			c.Assert(report.Documents[0].Error, qt.Contains, tc.problem)
			c.Assert(report.Summary.Trims, qt.Equals, 0)
			c.Assert(model.received(), qt.HasLen, tc.calls)
		})
	}
}

func TestTrimRejectsOptionsBeforeAnyCall(t *testing.T) {
	model := newFakeModel(t, trimAnswers(draftTrims))
	cases := []struct {
		name    string
		args    []string
		problem string
	}{
		{"no model", nil, "name a model with --endpoint and --model, or a local program with --command"},
		{"write with stdin", []string{"--write", "--endpoint", model.server.URL, "--model", "first"},
			"--write changes files and cannot be combined with --stdin"},
		{"sarif report", []string{"--report", "sarif:-", "--endpoint", model.server.URL, "--model", "first"},
			`review writes text or json reports, not "sarif"`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := qt.New(t)
			args := append([]string{"--stdin", "--filename", "draft.md"}, tc.args...)
			outcome := runCommand(t, "trim", t.TempDir(), trimDraft, args...)
			c.Assert(outcome.code, qt.Equals, 2)
			c.Assert(outcome.stderr, qt.Contains, tc.problem)
			c.Assert(outcome.stdout, qt.Equals, "")
		})
	}
	qt.New(t).Assert(model.received(), qt.HasLen, 0)
}

func TestTrimFailsWhenNoSentenceCanBeTrimmed(t *testing.T) {
	c := qt.New(t)
	model := newFakeModel(t, trimAnswers(draftTrims))
	outcome := runCommand(t, "trim", t.TempDir(), "```sh\nmake build\n```\n", "--stdin", "--filename", "code.md", "--quiet",
		"--endpoint", model.server.URL, "--model", "first")
	c.Assert(outcome.code, qt.Equals, 2)
	c.Assert(outcome.stderr, qt.Contains, "no selected document has a sentence to trim")
	c.Assert(model.received(), qt.HasLen, 0)
}
