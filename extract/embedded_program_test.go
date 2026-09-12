package extract_test

import (
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/extract"
)

// A shell string or heredoc whose lines are mostly code holds a program for
// another interpreter. The whole literal is excluded with its reason, and the
// comment and the message string around it are still checked.
func TestShellProgramLiteralsAreExcluded(t *testing.T) {
	const jqFilter = "'.items[]\n  | select(.enabled == true)\n  | {name: .name, id: .id}\n  | .name'"
	const python = "import sys\nfrom pathlib import Path\n\ndef main(root):\n    for path in Path(root).rglob(\"*.md\"):\n" +
		"        print(path)\n\nmain(sys.argv[1])\n"
	for _, row := range []struct {
		name    string
		format  document.Format
		source  string
		program string
	}{
		{"bash jq filter", document.Bash,
			"# Keep the input.\nnames=$(jq -r " + jqFilter + " input.json)\nmessage='Keep café coordinates.'\n", jqFilter},
		{"bash python heredoc", document.Bash,
			"# Keep the input.\npython3 - <<'PY'\n" + python + "PY\nmessage='Keep café coordinates.'\n", python},
		{"sh awk program", document.Shell,
			"# Keep the input.\nawk '\nBEGIN { FS = \",\" }\n/^#/ { next }\n{ print $2 }\n' input.csv\nmessage='Keep café coordinates.'\n",
			"'\nBEGIN { FS = \",\" }\n/^#/ { next }\n{ print $2 }\n'"},
		{"zsh sql heredoc", document.Zsh,
			"# Keep the input.\npsql <<SQL\nSELECT name\nFROM records\nWHERE enabled = 1;\nSQL\nmessage='Keep café coordinates.'\n",
			"SELECT name\nFROM records\nWHERE enabled = 1;\n"},
		{"bash shebang heredoc", document.Bash,
			"# Keep the input.\ncat <<'PY' >run.py\n#!/usr/bin/env python3\nprint('ready')\nPY\nmessage='Keep café coordinates.'\n",
			"#!/usr/bin/env python3\nprint('ready')\n"},
		{"powershell here-string", document.PowerShell,
			"# Keep the input.\n$filter = @'\n.items[]\n| select(.enabled == true)\n| .name\n'@\n$message = 'Keep café coordinates.'\n",
			"\n.items[]\n| select(.enabled == true)\n| .name\n"},
	} {
		t.Run(row.name, func(t *testing.T) {
			c := qt.New(t)
			doc, err := extract.Parse(t.Context(),
				document.Source{Name: "sample", Format: row.format, Bytes: []byte(row.source)}, extract.Options{})
			c.Assert(err, qt.IsNil)
			c.Assert(blockTexts(doc), qt.DeepEquals, []string{"Keep the input.", "Keep café coordinates."})
			c.Assert(doc.Excluded, qt.HasLen, 1)
			c.Assert(doc.Excluded[0].Reason, qt.Equals, "embedded-program")
			span := doc.Excluded[0].Span
			c.Assert(row.source[span.Start:span.End], qt.Contains, row.program)
			c.Assert(strings.Trim(row.source[span.Start:span.End], "'@\"\n"), qt.Equals, strings.Trim(row.program, "'@\"\n"))
			assertSourceMap(c, doc, row.source)
		})
	}
}

// Prose in shell literals stays checked: short literals never reach the line
// test, English lines end sentences, and other grammars are not tested at all.
func TestShellProseLiteralsStayChecked(t *testing.T) {
	for _, row := range []struct {
		name   string
		format document.Format
		source string
		text   string
	}{
		{"three-line heredoc", document.Bash,
			"cat <<EOF\nThe first line is prose.\nThe second line is prose too.\nThe third line closes the note.\nEOF\n",
			"The first line is prose.\nThe second line is prose too.\nThe third line closes the note.\n"},
		{"one-line heredoc", document.Bash, "cat <<EOF\nKeep café coordinates.\nEOF\n", "Keep café coordinates.\n"},
		{"two-line filter", document.Bash, "jq -r '.items[]\n| .name' input.json\n", ".items[]\n| .name"},
		{"usage message", document.Shell,
			"cat <<USAGE\nUsage: acquire.sh [options] <file>\n  -v  Print progress while cloning.\n" +
				"  -o  Write records to the given directory.\nUSAGE\n",
			"Usage: acquire.sh [options] <file>\n  -v  Print progress while cloning.\n  -o  Write records to the given directory.\n"},
		{"python program string", document.Python, "text = \"\"\"import sys\nimport os\nprint(sys.argv)\n\"\"\"\n",
			"import sys\nimport os\nprint(sys.argv)\n"},
	} {
		t.Run(row.name, func(t *testing.T) {
			c := qt.New(t)
			doc, err := extract.Parse(t.Context(),
				document.Source{Name: "sample", Format: row.format, Bytes: []byte(row.source)}, extract.Options{})
			c.Assert(err, qt.IsNil)
			c.Assert(doc.Blocks, qt.HasLen, 1)
			c.Assert(doc.Blocks[0].Text, qt.Equals, row.text)
			c.Assert(doc.Excluded, qt.HasLen, 0)
		})
	}
}

// Configured policy decides before the line test runs: a reasoned exception
// keeps its ID and reason, and a disabled string context keeps its own reason.
func TestConfiguredPolicyOutranksProgramTest(t *testing.T) {
	source := "# Keep the input.\nfilter='.items[]\n| select(.enabled == true)\n| {name: .name}\n| .name'\n"
	exception := extract.Exception{
		ID: "jq-filter", Paths: []string{"*.sh"}, Kinds: []string{"string"}, Symbols: []string{"filter"}, Reason: "The filter is a jq program.",
	}
	for _, row := range []struct {
		name   string
		policy extract.Policy
		reason string
	}{
		{"exception", extract.Policy{Exceptions: []extract.Exception{exception}}, "config:jq-filter: The filter is a jq program."},
		{"disabled context", extract.Policy{Languages: map[document.Format]extract.LanguagePolicy{
			document.Shell: {Contexts: []string{"comment"}},
		}}, "config:context-disabled:string"},
	} {
		t.Run(row.name, func(t *testing.T) {
			c := qt.New(t)
			doc, err := extract.Parse(t.Context(),
				document.Source{Name: "sample.sh", Format: document.Shell, Bytes: []byte(source)}, extract.Options{Policy: row.policy})
			c.Assert(err, qt.IsNil)
			c.Assert(blockTexts(doc), qt.DeepEquals, []string{"Keep the input."})
			c.Assert(doc.Excluded, qt.HasLen, 1)
			c.Assert(doc.Excluded[0].Reason, qt.Equals, row.reason)
			c.Assert(source[doc.Excluded[0].Span.Start:doc.Excluded[0].Span.End], qt.Equals, source[strings.Index(source, "'"):len(source)-1])
		})
	}
}

func blockTexts(doc document.Document) []string {
	texts := make([]string, 0, len(doc.Blocks))
	for _, block := range doc.Blocks {
		texts = append(texts, strings.TrimSpace(block.Text))
	}
	return texts
}
