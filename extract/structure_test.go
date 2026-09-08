package extract_test

import (
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/extract"
)

func TestMarkdownContextsSurviveReformattingAndUnrelatedLines(t *testing.T) {
	c := qt.New(t)
	one := "# Retries\n\nThe client may retry.\n\n## Limits\n\nUse three attempts.\n\n# Uploads\n\nThe client may retry.\n"
	two := "\ufeff# **Retries**\r\n\r\nUnrelated new prose.\r\n\r\nThe client may retry.\r\n\r\n" +
		"Limits\r\n------\r\n\r\nUse three attempts.\r\n\r\nUploads\r\n=======\r\n\r\nThe client may retry.\r\n"
	before := parseStructure(t, document.Markdown, one)
	after := parseStructure(t, document.Markdown, two)
	c.Assert(before.Blocks[1].Context, qt.DeepEquals, after.Blocks[2].Context)
	c.Assert(before.Blocks[3].Context, qt.DeepEquals, after.Blocks[4].Context)
	c.Assert(before.Blocks[5].Context, qt.DeepEquals, after.Blocks[6].Context)
	c.Assert(before.Blocks[1].Context, qt.Not(qt.DeepEquals), before.Blocks[5].Context)
	c.Assert(before.Blocks[3].Context, qt.HasLen, 2)
}

func TestNamedSourceContextsDistinguishIdenticalStrings(t *testing.T) {
	for _, test := range []struct {
		format document.Format
		source string
	}{
		{document.Go, "package sample\nvar first = \"The client retries.\"\nvar second = \"The client retries.\"\n"},
		{document.JavaScript, "const first = 'The client retries.';\nconst second = 'The client retries.';\n"},
		{document.TypeScript, "const first = 'The client retries.';\nconst second = 'The client retries.';\n"},
		{document.TSX, "const first = 'The client retries.';\nconst second = 'The client retries.';\n"},
		{document.Python, "first = 'The client retries.'\nsecond = 'The client retries.'\n"},
		{document.Rust, "const FIRST: &str = \"The client retries.\";\nconst SECOND: &str = \"The client retries.\";\n"},
		{document.Java, "class Sample { String first = \"The client retries.\"; String second = \"The client retries.\"; }\n"},
		{document.C, "const char *first = \"The client retries.\";\nconst char *second = \"The client retries.\";\n"},
		{document.CPP, "const char *first = \"The client retries.\";\nconst char *second = \"The client retries.\";\n"},
		{document.CSharp, "class Sample { string first = \"The client retries.\"; string second = \"The client retries.\"; }\n"},
		{document.YAML, "first: The client retries.\nsecond: The client retries.\n"},
		{document.Bash, "first='The client retries.'\nsecond='The client retries.'\n"},
		{document.Shell, "first='The client retries.'\nsecond='The client retries.'\n"},
		{document.Zsh, "first='The client retries.'\nsecond='The client retries.'\n"},
		{document.Fish, "function first\n echo 'The client retries.'\nend\nfunction second\n echo 'The client retries.'\nend\n"},
		{document.PowerShell, "$first = 'The client retries.'\n$second = 'The client retries.'\n"},
	} {
		t.Run(string(test.format), func(t *testing.T) {
			c := qt.New(t)
			doc := parseStructure(t, test.format, test.source)
			c.Assert(doc.Blocks, qt.HasLen, 2)
			c.Assert(len(doc.Blocks[0].Context) > 0, qt.IsTrue)
			c.Assert(doc.Blocks[0].Context, qt.Not(qt.DeepEquals), doc.Blocks[1].Context)
			moved := parseStructure(t, test.format, "\n\n"+test.source)
			c.Assert(doc.Blocks[0].Context, qt.DeepEquals, moved.Blocks[0].Context)
			c.Assert(doc.Blocks[1].Context, qt.DeepEquals, moved.Blocks[1].Context)
		})
	}
}

func TestGoCommentContextIncludesOwnerAndReceiver(t *testing.T) {
	c := qt.New(t)
	doc := parseStructure(t, document.Go, "package sample\n"+
		"// The client retries.\nfunc (first *First) Retry() {}\n"+
		"// The client retries.\nfunc (second *Second) Retry() {}\n")
	c.Assert(doc.Blocks, qt.HasLen, 2)
	c.Assert(strings.Join(doc.Blocks[0].Context, "/"), qt.Contains, "First")
	c.Assert(strings.Join(doc.Blocks[1].Context, "/"), qt.Contains, "Second")
	c.Assert(doc.Blocks[0].Context, qt.Not(qt.DeepEquals), doc.Blocks[1].Context)
}

func parseStructure(t *testing.T, format document.Format, text string) document.Document {
	t.Helper()
	c := qt.New(t)
	doc, err := extract.Parse(t.Context(), document.Source{Name: "sample", Format: format, Bytes: []byte(text)},
		extract.Options{IncludeStructure: true})
	c.Assert(err, qt.IsNil)
	return doc
}
