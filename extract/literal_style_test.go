package extract_test

import (
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/extract"
)

// Raw, triple-quoted, heredoc and here-string literals must reach the rules with
// their backslashes intact: decoding them would invent characters the file does
// not contain and shift every following source coordinate.
func TestRawAndHeredocLiteralsKeepTheirSourceText(t *testing.T) {
	for _, row := range []struct {
		name   string
		format document.Format
		source string
		text   string
	}{
		{"python raw prefix", document.Python, "message = r\"Visible raw \\d text.\"\n", "Visible raw \\d text."},
		{"python triple quotes", document.Python, "message = \"\"\"Visible triple text.\"\"\"\n", "Visible triple text."},
		{"java text block", document.Java, "class A { String m = \"\"\"\n    Visible block text.\n    \"\"\"; }\n",
			"\n    Visible block text.\n    "},
		{"cpp raw string", document.CPP, "const char *m = R\"(Visible raw \\d text.)\";\n", "Visible raw \\d text."},
		{"go raw string", document.Go, "package a\nconst m = `Visible raw \\d text.`\n", "Visible raw \\d text."},
		{"bash heredoc", document.Bash, "cat <<EOF\nVisible heredoc text.\nEOF\n", "Visible heredoc text.\n"},
		{"bash quoted heredoc", document.Bash, "cat <<'EOF'\nVisible quoted \\d heredoc text.\nEOF\n",
			"Visible quoted \\d heredoc text.\n"},
		{"powershell here-string", document.PowerShell, "$m = @\"\nVisible here-string text.\n\"@\n",
			"\nVisible here-string text.\n"},
		{"powershell verbatim here-string", document.PowerShell, "$m = @'\nVisible verbatim here-string text.\n'@\n",
			"\nVisible verbatim here-string text.\n"},
	} {
		t.Run(row.name, func(t *testing.T) {
			c := qt.New(t)
			doc, err := extract.Parse(t.Context(),
				document.Source{Name: "sample", Format: row.format, Bytes: []byte(row.source)}, extract.Options{})
			c.Assert(err, qt.IsNil)
			c.Assert(doc.Blocks, qt.HasLen, 1)
			c.Assert(doc.Blocks[0].Text, qt.Equals, row.text)
			c.Assert(doc.Blocks[0].Map, qt.HasLen, len(doc.Blocks[0].Text))
		})
	}
}

// A Python byte literal holds bytes, not English. It is excluded with a reason
// rather than decoded as text.
func TestPythonByteLiteralIsExcluded(t *testing.T) {
	c := qt.New(t)
	source := "message = b\"Visible byte text.\"\n"
	doc, err := extract.Parse(t.Context(),
		document.Source{Name: "sample.py", Format: document.Python, Bytes: []byte(source)}, extract.Options{})
	c.Assert(err, qt.IsNil)
	c.Assert(doc.Blocks, qt.HasLen, 0)
	c.Assert(doc.Excluded, qt.HasLen, 1)
	c.Assert(doc.Excluded[0].Reason, qt.Equals, "byte-literal")
}

// YAML has four scalar styles with different escape rules; each one maps back to
// its own source bytes.
func TestYAMLScalarStylesDecodeToProse(t *testing.T) {
	for _, row := range []struct {
		name   string
		source string
		text   string
	}{
		{"hexadecimal escape", "key: \"Visible caf\\xe9 text.\"\n", "Visible café text."},
		{"long scalar escape", "key: \"Emoji \\U0001F600 visible text.\"\n", "Emoji 😀 visible text."},
		{"control escape", "key: \"Visible \\a control text.\"\n", "Visible  \x00  control text."},
		{"line continuation", "key: \"Visible continued \\\n  text here.\"\n", "Visible continued text here."},
		{"doubled single quotes", "key: 'Visible ''quoted'' text.'\n", "Visible 'quoted' text."},
		{"literal block", "key: |\n  Visible block text.\n  More visible text.\n", "Visible block text.\nMore visible text.\n"},
		{"folded block", "key: >\n  Visible folded text.\n  More visible text.\n", "Visible folded text. More visible text.\n"},
	} {
		t.Run(row.name, func(t *testing.T) {
			c := qt.New(t)
			doc, err := extract.Parse(t.Context(),
				document.Source{Name: "sample.yaml", Format: document.YAML, Bytes: []byte(row.source)}, extract.Options{})
			c.Assert(err, qt.IsNil)
			c.Assert(doc.Blocks, qt.HasLen, 1)
			c.Assert(doc.Blocks[0].Text, qt.Equals, row.text)
			c.Assert(doc.Blocks[0].Map, qt.HasLen, len(doc.Blocks[0].Text))
		})
	}
}

// Chomped trailing line breaks belong to the block scalar's own origins, so a
// following key still starts at its real byte offset.
func TestYAMLBlockScalarKeepsFollowingKeyOrigins(t *testing.T) {
	c := qt.New(t)
	source := "key: |\n  Visible block text.\n\n\nnext: More visible text.\n"
	doc, err := extract.Parse(t.Context(),
		document.Source{Name: "sample.yaml", Format: document.YAML, Bytes: []byte(source)}, extract.Options{})
	c.Assert(err, qt.IsNil)
	c.Assert(doc.Blocks, qt.HasLen, 2)
	c.Assert(doc.Blocks[0].Text, qt.Equals, "Visible block text.\n")
	c.Assert(doc.Blocks[1].Text, qt.Equals, "More visible text.")
	span := document.Bounds(doc.Blocks[1].Spans(0, len(doc.Blocks[1].Text)))
	c.Assert(string([]byte(source)[span.Start:span.End]), qt.Equals, "More visible text.")
}
