package extract_test

import (
	"context"
	"strings"
	"testing"
	"time"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/extract"
)

func TestSourceLanguageGrammars(t *testing.T) {
	cases := []struct {
		name   string
		format document.Format
		source string
	}{
		{"Go", document.Go, "// A useful comment.\r\npackage sample\nconst message = \"Let\\u2019s dive into details.\"\n"},
		{"JavaScript", document.JavaScript, "// A useful comment.\nconst message = \"Let\\u2019s dive into details.\";"},
		{"TypeScript", document.TypeScript, "// A useful comment.\nconst message: string = \"Let\\u2019s dive into details.\";"},
		{"TSX", document.TSX, "// A useful comment.\nconst view = <p title=\"Let’s dive into details.\" />;"},
		{"Python", document.Python, "# A useful comment.\nmessage = \"Let\\u2019s dive into details.\"\n"},
		{"Rust", document.Rust, "// A useful comment.\nconst MESSAGE: &str = \"Let\\u{2019}s dive into details.\";"},
		{"Java", document.Java, "// A useful comment.\nclass Sample { String message = \"Let\\u2019s dive into details.\"; }"},
		{"C", document.C, "// A useful comment.\nconst char *message = \"Let\\u2019s dive into details.\";"},
		{"C++", document.CPP, "// A useful comment.\nconst char *message = R\"tag(Let’s dive into details.)tag\";"},
		{"C#", document.CSharp, "// A useful comment.\nclass Sample { const string message = \"Let\\u2019s dive into details.\"; }"},
		{"YAML", document.YAML, "# A useful comment.\nmessage: \"Let\\u2019s dive into details.\"\n"},
		{"Bash", document.Bash, "#!/usr/bin/env bash\n# A useful comment.\necho $'Let\\u2019s dive into details.'\n"},
		{"POSIX shell", document.Shell, "#!/bin/sh\n# A useful comment.\necho 'Let’s dive into details.'\n"},
		{"Zsh common syntax", document.Zsh, "#!/bin/zsh\n# A useful comment.\nprint 'Let’s dive into details.'\n"},
		{"Fish", document.Fish, "# A useful comment.\necho 'Let’s dive into details.'\n"},
		{"PowerShell", document.PowerShell, "# A useful comment.\n$message = \"Let`u{2019}s dive into details.\"\n"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := qt.New(t)
			doc, err := extract.Parse(
				t.Context(),
				document.Source{Name: "source", Format: tc.format, Bytes: []byte(tc.source)},
				extract.Options{},
			)
			c.Assert(err, qt.IsNil)
			c.Assert(doc.Blocks, qt.HasLen, 2)
			c.Assert(doc.Blocks[0].Kind, qt.Equals, "comment")
			c.Assert(strings.TrimSpace(doc.Blocks[0].Text), qt.Equals, "A useful comment.")
			c.Assert(doc.Blocks[1].Kind, qt.Equals, "string")
			c.Assert(doc.Blocks[1].Text, qt.Equals, "Let’s dive into details.")
			assertSourceMap(c, doc, tc.source)
		})
	}
}

func TestBlockCommentMarkers(t *testing.T) {
	cases := []struct {
		name   string
		format document.Format
		source string
	}{
		{"Go", document.Go, "package sample\n/*\r\n * It is important\r\n * to note that.\r\n */"},
		{"JavaScript", document.JavaScript, "/**\n * It is important\n * to note that.\n */"},
		{"Java", document.Java, "/**\n * It is important\n * to note that.\n */ class Sample {}"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := qt.New(t)
			doc, err := extract.Parse(
				t.Context(),
				document.Source{Name: "sample", Format: tc.format, Bytes: []byte(tc.source)},
				extract.Options{},
			)
			c.Assert(err, qt.IsNil)
			c.Assert(doc.Blocks, qt.HasLen, 1)
			c.Assert(strings.Join(strings.Fields(doc.Blocks[0].Text), " "), qt.Equals, "It is important to note that.")
			assertSourceMap(c, doc, tc.source)
		})
	}
}

func TestEmptyRawStrings(t *testing.T) {
	cases := []struct {
		name   string
		format document.Format
		source string
	}{
		{"Rust", document.Rust, "const VALUE: &str = r#\"\"#;"},
		{"C++", document.CPP, "const char *value = R\"tag()tag\";"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := qt.New(t)
			doc, err := extract.Parse(
				t.Context(),
				document.Source{Name: "sample", Format: tc.format, Bytes: []byte(tc.source)},
				extract.Options{},
			)
			c.Assert(err, qt.IsNil)
			c.Assert(doc.Blocks, qt.HasLen, 0)
		})
	}
}

func TestSourceDeadline(t *testing.T) {
	c := qt.New(t)
	ctx, cancel := context.WithTimeout(t.Context(), time.Millisecond)
	defer cancel()
	source := strings.Repeat("const value = 'A useful string.';\n", 30000)
	_, err := extract.Parse(ctx, document.Source{Name: "large.js", Format: document.JavaScript, Bytes: []byte(source)}, extract.Options{})
	c.Assert(err, qt.ErrorIs, context.DeadlineExceeded)
}

func assertSourceMap(c *qt.C, doc document.Document, source string) {
	c.Helper()
	for _, block := range doc.Blocks {
		c.Assert(block.Map, qt.HasLen, len(block.Text))
		for _, span := range block.Map {
			c.Assert(span.Valid(len(source)), qt.IsTrue)
			_, err := document.Locate([]byte(source), span.Start)
			c.Assert(err, qt.IsNil)
			_, err = document.Locate([]byte(source), span.End)
			c.Assert(err, qt.IsNil)
		}
	}
}

func TestLiteralBoundaries(t *testing.T) {
	cases := []struct {
		name   string
		format document.Format
		source string
		want   string
	}{
		{"JS interpolation", document.JavaScript, "const value = `It is ${important} to note that.`;", "It is  \x00  to note that."},
		{"Python interpolation", document.Python, "value = f'It is {important} to note that.'", "It is  \x00  to note that."},
		{"Bash substitutions", document.Bash, "echo \"It is ${important} to $(printf 'note') that.\"", "It is  \x00  to  \x00  that."},
		{"Fish variable", document.Fish, "echo \"It is $important to note that.\"", "It is  \x00  to note that."},
		{"JavaScript identity escapes", document.JavaScript, `const value = "\a useful \Uppercase word.";`, "a useful Uppercase word."},
		{"PowerShell variable", document.PowerShell, "$value = \"It is $important to note that.\"", "It is  \x00  to note that."},
		{"heredoc expansion", document.Bash, "cat <<EOF\nIt is $important to note that.\nEOF\n", "It is  \x00  to note that.\n"},
		{"literal heredoc", document.Bash, "cat <<'EOF'\nIt is $important to note that.\nEOF\n", "It is $important to note that.\n"},
		{
			"PowerShell here-string",
			document.PowerShell,
			"$value = @\"\nIt is $important to note that.\n\"@",
			"\nIt is  \x00  to note that.\n",
		},
		{"Go raw", document.Go, "package sample\nconst value = `Keep \\n literal.`", "Keep \\n literal."},
		{"Python raw", document.Python, "value = r'Keep \\n literal.'", "Keep \\n literal."},
		{"Rust raw", document.Rust, "const VALUE: &str = r#\"Keep \\n literal.\"#;", "Keep \\n literal."},
		{"Java text block", document.Java, "class Sample { String value = \"\"\"\nA useful document.\n\"\"\"; }", "\nA useful document.\n"},
		{"PowerShell quote", document.PowerShell, "$value = 'Let''s read the manual.'", "Let's read the manual."},
		{"Unicode pair", document.JavaScript, "const value = '\\uD83D\\uDE00 A useful string.';", "😀 A useful string."},
		{"UTF-8 byte escapes", document.Go, "package sample\nconst value = \"Let\\xe2\\x80\\x99s read.\"", "Let’s read."},
		{"comment markers in string", document.Python, "value = 'Use // or # in the example.'", "Use // or # in the example."},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := qt.New(t)
			doc, err := extract.Parse(
				t.Context(),
				document.Source{Name: "source", Format: tc.format, Bytes: []byte(tc.source)},
				extract.Options{},
			)
			c.Assert(err, qt.IsNil)
			c.Assert(len(doc.Blocks) > 0, qt.IsTrue)
			c.Assert(doc.Blocks[0].Text, qt.Equals, tc.want)
			assertSourceMap(c, doc, tc.source)
		})
	}
}

func TestEscapedEvidenceCoordinates(t *testing.T) {
	c := qt.New(t)
	source := "// 😀 header\r\nconst value = \"Let\\u2019s read.\";\r\n"
	doc, err := extract.Parse(
		t.Context(),
		document.Source{Name: "a.js", Format: document.JavaScript, Bytes: []byte(source)},
		extract.Options{},
	)
	c.Assert(err, qt.IsNil)
	block := doc.Blocks[1]
	spans := block.Spans(3, 6)
	c.Assert(spans, qt.HasLen, 1)
	c.Assert(source[spans[0].Start:spans[0].End], qt.Equals, "\\u2019")
	position, err := document.Locate([]byte(source), spans[0].Start)
	c.Assert(err, qt.IsNil)
	c.Assert(position, qt.Equals, document.Position{Line: 2, Column: 19})
}

func TestIncompleteSourceIsRejected(t *testing.T) {
	cases := []struct {
		name   string
		format document.Format
		source string
	}{
		{"JavaScript", document.JavaScript, "const value = 'unfinished"},
		{"Python", document.Python, "value = 'unfinished"},
		{"Bash", document.Bash, "echo \"unfinished"},
		{"PowerShell", document.PowerShell, "$value = \"unfinished"},
		{"C#", document.CSharp, "class Sample { string value = \"unfinished"},
		{"C# escaped interpolation braces", document.CSharp, "class Sample { string value = $\"Read {{literal}}.\"; }"},
		{"YAML quote", document.YAML, "message: \"unfinished\n"},
		{"YAML mapping", document.YAML, "message: text: trailing\n"},
		{"Zsh extended syntax", document.Zsh, "print ${(@f)value}"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := qt.New(t)
			_, err := extract.Parse(
				t.Context(),
				document.Source{Name: "bad", Format: tc.format, Bytes: []byte(tc.source)},
				extract.Options{},
			)
			c.Assert(err, qt.IsNotNil)
		})
	}
}
