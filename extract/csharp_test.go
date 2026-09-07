package extract_test

import (
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/extract"
)

func TestCSharpStrings(t *testing.T) {
	cases := []struct{ name, literal, text string }{
		{"quoted", `"Let\x2019s read \"the manual\"."`, "Let’s read \"the manual\"."},
		{"verbatim", `@"Let's read ""the manual"" at C:\docs."`, `Let's read "the manual" at C:\docs.`},
		{"verbatim leading quote", `@"""Read the manual."""`, `"Read the manual."`},
		{"interpolated", `$"It is {name} to note that."`, "It is  \x00  to note that."},
		{"verbatim interpolated", `$@"Read {name} at C:\docs."`, "Read  \x00  at C:\\docs."},
		{"raw", `"""Let's read "the manual"."""`, `Let's read "the manual".`},
		{"four delimiters", `""""Read """the manual""".""""`, `Read """the manual""".`},
		{"raw interpolated", `$$"""It is {{name}} to note that {literal}."""`, "It is  \x00  to note that {literal}."},
		{
			"raw multiline",
			"\"\"\"\n    First line.\n      Extra indent.\n    Last line.\n    \"\"\"",
			"First line.\n  Extra indent.\nLast line.",
		},
		{"raw blank line", "\"\"\"\n    First line.\n\n    Last line.\n    \"\"\"", "First line.\n\nLast line."},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := qt.New(t)
			source := "class Sample { string message = " + tc.literal + "; }"
			doc, err := extract.Parse(t.Context(), document.Source{Name: "sample.cs", Format: document.CSharp, Bytes: []byte(source)},
				extract.Options{})
			c.Assert(err, qt.IsNil)
			c.Assert(doc.Blocks, qt.HasLen, 1)
			c.Assert(doc.Blocks[0].Text, qt.Equals, tc.text)
			assertSourceMap(c, doc, source)
		})
	}
}

func TestCSharpNestedInterpolationString(t *testing.T) {
	c := qt.New(t)
	source := `class Sample { string message = $"Read {format("manual")}."; }`
	doc, err := extract.Parse(t.Context(), document.Source{Name: "sample.cs", Format: document.CSharp, Bytes: []byte(source)},
		extract.Options{})
	c.Assert(err, qt.IsNil)
	c.Assert(doc.Blocks, qt.HasLen, 2)
	c.Assert(doc.Blocks[0].Text, qt.Equals, "Read  \x00 .")
	c.Assert(doc.Blocks[1].Text, qt.Equals, "manual")
	assertSourceMap(c, doc, source)
}
