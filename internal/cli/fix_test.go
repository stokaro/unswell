package cli_test

import (
	"bytes"
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/internal/cli"
)

// fixPlain is ordinary documentation without any mark that the index counts:
// its sentences differ in length and in how they start.
var fixPlain = "When a request arrives, the server checks the token before anything else, because an expired token is the most common " +
	"reason for a rejection and the check costs almost nothing. Why first? Everything after it needs a read from the database. " +
	"If the token is valid, the handler loads the account, compares its plan with the limits of the endpoint, and only then " +
	"reads the body, which can be several megabytes for an upload. Small bodies stay in memory. A larger one goes to a " +
	"temporary file (the handler removes it once the response is sent), so that a slow client cannot hold the memory of a " +
	"worker for minutes. Does the order matter? It does: on a busy day about one request in fifty fails the first check, and " +
	"none of those reach the database or the disk. After the body is read, the handler validates each field, writes the row " +
	"inside one transaction, and answers with the identifier of the new record together with the time the server stored it. "

// fixDraft has marks with a mechanical replacement and one stock word.
var fixDraft = "# \U0001F680 Client\n\nA seamless client — it retries → then it waits.\n\n" +
	"- **Fast**: it starts quickly.\n- ⚡ One connection.\n\n" + fixPlain + "\n"

var fixRevised = "# Client\n\nA seamless client, it retries -> then it waits.\n\n- Fast: it starts quickly.\n- One connection.\n\n" +
	fixPlain + "\n"

type fixReport struct {
	SchemaVersion string `json:"schema_version"`
	Summary       struct {
		Documents int `json:"documents"`
		Failed    int `json:"failed"`
		Unindexed int `json:"without_index"`
		Low       int `json:"left_at_low"`
		Changed   int `json:"changed"`
		Removed   int `json:"marks_removed"`
		Left      int `json:"marks_left"`
		Written   int `json:"written"`
	} `json:"summary"`
	Documents []struct {
		Path    string         `json:"path"`
		Indexed bool           `json:"indexed"`
		LeftLow bool           `json:"left_at_low"`
		Before  int            `json:"index_before"`
		After   int            `json:"index_after"`
		Removed map[string]int `json:"marks_removed"`
		Left    map[string]int `json:"marks_left"`
		Written bool           `json:"written"`
		Error   string         `json:"error"`
	} `json:"documents"`
}

func runFix(t *testing.T, dir string, args ...string) reviewOutcome {
	t.Helper()
	var output, stderr bytes.Buffer
	environment := cli.Environment{Dir: dir, In: strings.NewReader(""), Out: &output, Err: &stderr,
		Getenv: func(string) string { return "" }}
	code := cli.Run(t.Context(), append([]string{"fix"}, args...), environment)
	return reviewOutcome{code: code, stdout: output.String(), stderr: stderr.String()}
}

func writeDraft(c *qt.C, root, name, text string, mode os.FileMode) string {
	path := filepath.Join(root, name)
	c.Assert(os.WriteFile(path, []byte(text), mode), qt.IsNil)
	return path
}

func TestFixProposesWithoutWriting(t *testing.T) {
	c := qt.New(t)
	root := t.TempDir()
	writeDraft(c, root, "draft.md", fixDraft, 0o600)
	outcome := runFix(t, root, "draft.md")
	c.Assert(outcome.code, qt.Equals, 0, qt.Commentf("stderr: %s", outcome.stderr))
	c.Assert(outcome.stdout, qt.Matches, `draft\.md: AI-sounding index \d+ -> \d+; removed 1 em dash, 1 emoji in a heading, `+
		`1 emoji at the start of a list item, 1 arrow, 1 bold label\. Left for you: 1 stock word\.\n`+
		`5 marks removed in 1 page of 1 document\.\n`+
		`Nothing was written\. Use --write to apply these replacements\.\n`+
		`A replacement is mechanical\. Read each one: a comma or a period is not always what the sentence needs, `+
		`and a stock word needs your own words\.\n`)
	c.Assert(outcome.stderr, qt.Matches, `fix: draft\.md: index \d+ -> \d+, 5 marks removed\n`)
	data, err := fs.ReadFile(os.DirFS(root), "draft.md")
	c.Assert(err, qt.IsNil)
	c.Assert(string(data), qt.Equals, fixDraft)
}

func TestFixWritesOnRequest(t *testing.T) {
	c := qt.New(t)
	root := t.TempDir()
	// #nosec G306 -- The test needs a mode other than the default of a new file to show that the mode is kept.
	path := writeDraft(c, root, "draft.md", fixDraft, 0o640)
	writeDraft(c, root, "short.md", "# Client\n\nIt connects — then it waits.\n", 0o600)
	writeDraft(c, root, "client.go", "package client\n\n// Start opens the connection — then it waits.\nfunc Start() {}\n", 0o600)
	outcome := runFix(t, root, "draft.md", "short.md", "client.go", "--write", "--quiet", "--report", "json:-")
	c.Assert(outcome.code, qt.Equals, 0, qt.Commentf("stderr: %s", outcome.stderr))
	c.Assert(outcome.stderr, qt.Equals, "")
	var report fixReport
	c.Assert(json.Unmarshal([]byte(outcome.stdout), &report), qt.IsNil)
	c.Assert(report.SchemaVersion, qt.Equals, "unswell-fix-v1")
	c.Assert(report.Summary.Documents, qt.Equals, 3)
	c.Assert(report.Summary.Unindexed, qt.Equals, 2)
	c.Assert(report.Summary.Changed, qt.Equals, 1)
	c.Assert(report.Summary.Removed, qt.Equals, 5)
	c.Assert(report.Summary.Left, qt.Equals, 1)
	c.Assert(report.Summary.Written, qt.Equals, 1)
	for _, entry := range report.Documents {
		c.Assert(entry.Written, qt.Equals, entry.Path == "draft.md")
		c.Assert(entry.Indexed, qt.Equals, entry.Path == "draft.md")
		if entry.Path != "draft.md" {
			continue
		}
		c.Assert(entry.Before > entry.After, qt.IsTrue, qt.Commentf("%d and %d", entry.Before, entry.After))
		c.Assert(entry.Removed, qt.DeepEquals, map[string]int{"em-dash": 1, "emoji-in-heading": 1, "emoji-at-list-item": 1,
			"arrow": 1, "bold-label-in-list": 1})
		c.Assert(entry.Left, qt.DeepEquals, map[string]int{"stock-word": 1})
	}
	data, err := fs.ReadFile(os.DirFS(root), "draft.md")
	c.Assert(err, qt.IsNil)
	c.Assert(string(data), qt.Equals, fixRevised)
	// A page without an index and a file that is not Markdown stay as they were.
	short, err := fs.ReadFile(os.DirFS(root), "short.md")
	c.Assert(err, qt.IsNil)
	c.Assert(string(short), qt.Contains, "—")
	source, err := fs.ReadFile(os.DirFS(root), "client.go")
	c.Assert(err, qt.IsNil)
	c.Assert(string(source), qt.Contains, "—")
	if runtime.GOOS != "windows" {
		info, err := os.Stat(path)
		c.Assert(err, qt.IsNil)
		c.Assert(info.Mode().Perm(), qt.Equals, os.FileMode(0o640))
	}
	entries, err := os.ReadDir(root)
	c.Assert(err, qt.IsNil)
	c.Assert(entries, qt.HasLen, 3)

	// A second run finds nothing to replace and names what is left.
	again := runFix(t, root, "draft.md", "--write", "--quiet")
	c.Assert(again.code, qt.Equals, 0)
	c.Assert(again.stdout, qt.Matches, `draft\.md: AI-sounding index \d+; nothing to replace\. Left for you: 1 stock word\.\n`+
		`0 marks removed in 0 pages of 1 document\.\nWritten to 0 files\. Review the change before you keep it\.\n(?s:.*)`)
}

func TestFixChangesTheProseOfAnMDXPage(t *testing.T) {
	c := qt.New(t)
	root := t.TempDir()
	component := "<Note title=\"A dash — in an attribute\">\n  The note warns — then it stops.\n</Note>\n\n"
	page := "import Note from './Note.astro';\n\n# Client\n\nA seamless client — it retries → then it waits.\n\n" +
		component + "The total is {items.map((item) => item.name + ' — ' + item.size)} items.\n\n" + fixPlain + "\n"
	writeDraft(c, root, "guide.mdx", page, 0o600)
	outcome := runFix(t, root, "guide.mdx", "--write", "--quiet")
	c.Assert(outcome.code, qt.Equals, 0, qt.Commentf("stderr: %s", outcome.stderr))
	c.Assert(outcome.stdout, qt.Matches, `guide\.mdx: AI-sounding index \d+ -> \d+; removed 2 em dashes, 1 arrow\. `+
		`Left for you: 1 stock word\.\n3 marks removed in 1 page of 1 document\.\n(?s:.*)`)
	data, err := fs.ReadFile(os.DirFS(root), "guide.mdx")
	c.Assert(err, qt.IsNil)
	// The prose changes, in the page and inside the component. The import,
	// the attribute, and the expression stay byte for byte.
	want := strings.NewReplacer(
		"A seamless client — it retries → then it waits.", "A seamless client, it retries -> then it waits.",
		"The note warns — then it stops.", "The note warns, then it stops.",
	).Replace(page)
	c.Assert(string(data), qt.Equals, want)
}

func TestFixLeavesAPageAtTheLevelLow(t *testing.T) {
	c := qt.New(t)
	root := t.TempDir()
	// One dash in a long page: the index is low, and the check lists no marks there.
	low := "# Client\n\nIt connects — then it waits. " + strings.Repeat(fixPlain, 8) + "\n"
	writeDraft(c, root, "low.md", low, 0o600)
	outcome := runFix(t, root, "low.md", "--write", "--report", "json:-")
	c.Assert(outcome.code, qt.Equals, 0, qt.Commentf("stderr: %s", outcome.stderr))
	c.Assert(outcome.stderr, qt.Matches, `fix: low\.md: index \d+, marks low: left as it is\n`)
	var report fixReport
	c.Assert(json.Unmarshal([]byte(outcome.stdout), &report), qt.IsNil)
	c.Assert(report.Summary.Low, qt.Equals, 1)
	c.Assert(report.Summary.Changed, qt.Equals, 0)
	c.Assert(report.Summary.Removed, qt.Equals, 0)
	c.Assert(report.Documents[0].LeftLow, qt.IsTrue)
	c.Assert(report.Documents[0].Before, qt.Equals, report.Documents[0].After)
	text := runFix(t, root, "low.md", "--quiet")
	c.Assert(text.stdout, qt.Matches, `0 marks removed in 0 pages of 1 document\.\n`+
		`Left as they are, with their marks at the level low: 1 page\. Use --include-low to replace those marks too\.\n(?s:.*)`)
	data, err := fs.ReadFile(os.DirFS(root), "low.md")
	c.Assert(err, qt.IsNil)
	c.Assert(string(data), qt.Equals, low)

	// With --include-low the same page is changed.
	included := runFix(t, root, "low.md", "--write", "--quiet", "--include-low", "--report", "json:-")
	c.Assert(included.code, qt.Equals, 0)
	c.Assert(json.Unmarshal([]byte(included.stdout), &report), qt.IsNil)
	c.Assert(report.Summary.Low, qt.Equals, 0)
	c.Assert(report.Summary.Removed, qt.Equals, 1)
	c.Assert(report.Documents[0].LeftLow, qt.IsFalse)
	data, err = fs.ReadFile(os.DirFS(root), "low.md")
	c.Assert(err, qt.IsNil)
	c.Assert(string(data), qt.Equals, strings.Replace(low, " — then", ", then", 1))
}

func TestFixLeavesTheRunsOfSentencesToAPerson(t *testing.T) {
	c := qt.New(t)
	root := t.TempDir()
	even := strings.Repeat("The server reads the request and sends the response to the client. ", 16)
	// High by its marks and by its rhythm: the dashes are replaced, and the runs stay for a person.
	page := "# Client\n\nIt connects — then it waits — and it closes — on a timeout.\n\n" + even + "\n"
	writeDraft(c, root, "draft.md", page, 0o600)
	outcome := runFix(t, root, "draft.md", "--write", "--quiet")
	c.Assert(outcome.code, qt.Equals, 0, qt.Commentf("stderr: %s", outcome.stderr))
	c.Assert(outcome.stdout, qt.Matches, `draft\.md: AI-sounding index \d+ -> \d+; removed 3 em dashes\. `+
		`Left for you: 1 run of plain sentences, 1 run of sentences that open with their subject\.\n(?s:.*)`)
	data, err := fs.ReadFile(os.DirFS(root), "draft.md")
	c.Assert(err, qt.IsNil)
	c.Assert(string(data), qt.Equals, strings.ReplaceAll(page, " —", ","))

	// High by its rhythm alone: the one arrow of a long page is not listed by the check, so the page is left as it is.
	rhythm := "# Client\n\n" + strings.Repeat(even, 5) + "\n\nThe request goes to the proxy → the server.\n"
	writeDraft(c, root, "even.md", rhythm, 0o600)
	left := runFix(t, root, "even.md", "--write", "--report", "json:-")
	c.Assert(left.code, qt.Equals, 0, qt.Commentf("stderr: %s", left.stderr))
	c.Assert(left.stderr, qt.Matches, `fix: even\.md: index \d+, marks low: left as it is\n`)
	var report fixReport
	c.Assert(json.Unmarshal([]byte(left.stdout), &report), qt.IsNil)
	c.Assert(report.Documents[0].LeftLow, qt.IsTrue)
	c.Assert(report.Documents[0].Before >= 70, qt.IsTrue, qt.Commentf("%d", report.Documents[0].Before))
	data, err = fs.ReadFile(os.DirFS(root), "even.md")
	c.Assert(err, qt.IsNil)
	c.Assert(string(data), qt.Equals, rhythm)
}

func TestFixDoesNotWriteAReadOnlyFile(t *testing.T) {
	c := qt.New(t)
	root := t.TempDir()
	path := writeDraft(c, root, "draft.md", fixDraft, 0o400)
	// The temporary directory is removed after the test on every platform.
	t.Cleanup(func() { _ = os.Chmod(path, 0o600) })
	outcome := runFix(t, root, "draft.md", "--write", "--quiet", "--report", "json:-")
	c.Assert(outcome.code, qt.Equals, 2)
	var report fixReport
	c.Assert(json.Unmarshal([]byte(outcome.stdout), &report), qt.IsNil)
	c.Assert(report.Documents[0].Error, qt.Equals, "the file is read-only and was not written")
	c.Assert(report.Documents[0].Written, qt.IsFalse)
	c.Assert(report.Summary.Failed, qt.Equals, 1)
	c.Assert(report.Summary.Written, qt.Equals, 0)
	data, err := fs.ReadFile(os.DirFS(root), "draft.md")
	c.Assert(err, qt.IsNil)
	c.Assert(string(data), qt.Equals, fixDraft)
}

func TestFixFailsClosed(t *testing.T) {
	cases := []struct {
		name    string
		args    []string
		message string
	}{
		{"too many jobs", []string{"draft.md", "--jobs", "40"}, "--jobs must be between 1 and 16"},
		{"no time", []string{"draft.md", "--timeout", "0s"}, "--timeout must be positive"},
		{"another report format", []string{"draft.md", "--report", "sarif:-"}, `text or json reports, not "sarif"`},
		{"a file that is not there", []string{"missing.md"}, "missing.md"},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			c := qt.New(t)
			root := t.TempDir()
			writeDraft(c, root, "draft.md", fixDraft, 0o600)
			outcome := runFix(t, root, test.args...)
			c.Assert(outcome.code, qt.Equals, 2)
			c.Assert(outcome.stderr, qt.Contains, test.message)
			data, err := fs.ReadFile(os.DirFS(root), "draft.md")
			c.Assert(err, qt.IsNil)
			c.Assert(string(data), qt.Equals, fixDraft)
		})
	}
}
