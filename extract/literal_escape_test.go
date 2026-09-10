package extract_test

import (
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/extract"
)

// Escape decoding decides which bytes reach the rules, so every language mode
// needs its own case: a shared decoder with per-language tables fails quietly.
func TestLanguageEscapeSequencesDecodeToProse(t *testing.T) {
	for _, row := range []struct {
		name   string
		format document.Format
		source string
		text   string
	}{
		{"bash ANSI-C bytes", document.Bash, "message=$'Caf\\xc3\\xa9 visible \\u00e9 text.'\n", "Café visible é text."},
		{"bash ANSI-C escape", document.Bash, "message=$'Escaped \\e visible text.'\n", "Escaped  \x00  visible text."},
		{"bash quoted", document.Bash, "message=\"Escaped \\$ visible \\d text.\"\n", "Escaped $ visible \\d text."},
		{"bash continuation", document.Bash, "message=\"Continued \\\nvisible text.\"\n", "Continued visible text."},
		{"fish single quotes", document.Fish, "set message 'Escaped \\' visible text.'\n", "Escaped ' visible text."},
		{"fish double quotes", document.Fish, "set message \"Escaped \\$ visible \\d text.\"\n", "Escaped $ visible \\d text."},
		{"javascript identity", document.JavaScript, "const message = \"Escaped \\q visible text.\";\n", "Escaped q visible text."},
		{"javascript surrogate pair", document.JavaScript,
			"const message = \"Emoji \\ud83d\\ude00 visible text.\";\n", "Emoji 😀 visible text."},
		{"javascript braced scalar", document.JavaScript,
			"const message = \"Emoji \\u{1f600} visible text.\";\n", "Emoji 😀 visible text."},
		{"python unknown escape", document.Python, "message = \"Escaped \\d visible text.\"\n", "Escaped \\d visible text."},
		{"java space escape", document.Java, "class A { String m = \"Escaped\\svisible text.\"; }\n", "Escaped visible text."},
		{"csharp null and hexadecimal", document.CSharp,
			"class A { string m = \"Escaped \\0 visible \\x41 text.\"; }\n", "Escaped  \x00  visible A text."},
		{"csharp verbatim quotes", document.CSharp, "class A { string m = @\"Visible \"\"quoted\"\" text.\"; }\n", "Visible \"quoted\" text."},
		{"rust braced scalar", document.Rust, "fn main() { let m = \"Caf\\u{e9} visible text.\"; }\n", "Café visible text."},
		{"rust continuation trims indentation", document.Rust,
			"fn main() { let m = \"Continued \\\n    visible text.\"; }\n", "Continued visible text."},
		{"c continuation across CRLF", document.C, "const char *m = \"Visible \\\r\nmore text.\";\n", "Visible more text."},
		{"go long scalar", document.Go, "package a\nconst m = \"Emoji \\U0001F600 visible text.\"\n", "Emoji 😀 visible text."},
		{"go carriage return", document.Go, "package a\nconst m = \"Visible \r more text.\"\n", "Visible   more text."},
		{"powershell escapes", document.PowerShell, "$message = \"Escaped `t visible `u{e9} text.\"\n", "Escaped \t visible é text."},
		{"powershell continuation", document.PowerShell, "$message = \"Continued `\nvisible text.\"\n", "Continued visible text."},
		{"powershell identity", document.PowerShell, "$message = \"Identity `q visible text.\"\n", "Identity q visible text."},
		{"powershell single quotes", document.PowerShell, "$m = 'Visible ''quoted'' text.'\n", "Visible 'quoted' text."},
	} {
		t.Run(row.name, func(t *testing.T) {
			c := qt.New(t)
			source := []byte(row.source)
			doc, err := extract.Parse(t.Context(),
				document.Source{Name: "sample", Format: row.format, Bytes: source}, extract.Options{})
			c.Assert(err, qt.IsNil)
			c.Assert(string(source), qt.Equals, row.source)
			c.Assert(doc.Blocks, qt.HasLen, 1)
			c.Assert(doc.Blocks[0].Text, qt.Equals, row.text)
			c.Assert(doc.Blocks[0].Map, qt.HasLen, len(doc.Blocks[0].Text))
		})
	}
}

// A literal whose escape cannot be decoded is a defect in the input, not prose
// to guess at. Extraction reports it instead of analyzing partial text.
func TestUndecodableEscapesFailExtraction(t *testing.T) {
	for _, row := range []struct {
		name   string
		format document.Format
		source string
		want   string
	}{
		{"lone surrogate", document.JavaScript, "const m = \"Visible \\ud83d text.\";\n",
			"string escape at byte 19: invalid Unicode scalar value"},
		{"mismatched surrogate pair", document.JavaScript, "const m = \"Visible \\ud83d\\u0041 text.\";\n",
			"string escape at byte 19: invalid Unicode surrogate pair"},
		{"byte escape above 255", document.CPP, "const char *m = \"Visible \\x1ff text.\";\n",
			"string escape at byte 25: byte escape exceeds 255"},
		{"numeric surrogate", document.CSharp, "class A { string m = \"Visible \\xd800 text.\"; }\n",
			"string escape at byte 30: numeric escape is not a Unicode scalar value"},
		{"unsupported python escape", document.Python, "m = \"Visible \\N{BULLET} text.\"\n",
			"string escape at byte 13: unsupported python escape \"\\\\\\\\N\""},
		{"unsupported java escape", document.Java, "class A { String m = \"Visible \\q text.\"; }\n",
			"string escape at byte 30: unsupported java escape \"\\\\\\\\q\""},
		{"unsupported rust escape", document.Rust, "fn main() { let m = \"Visible \\q text.\"; }\n",
			"string escape at byte 29: unsupported rust escape \"\\\\\\\\q\""},
	} {
		t.Run(row.name, func(t *testing.T) {
			c := qt.New(t)
			doc, err := extract.Parse(t.Context(),
				document.Source{Name: "sample", Format: row.format, Bytes: []byte(row.source)}, extract.Options{})
			c.Assert(err, qt.ErrorMatches, row.want)
			c.Assert(doc.Blocks, qt.HasLen, 0)
		})
	}
}
