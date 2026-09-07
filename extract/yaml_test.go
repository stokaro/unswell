package extract_test

import (
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/extract"
)

func TestYAMLScalarStyles(t *testing.T) {
	cases := []struct{ name, source, text string }{
		{"plain", "message: Let's read the manual.\n", "Let's read the manual."},
		{"single", "message: 'Let''s read the manual.'\n", "Let's read the manual."},
		{"double", "message: \"Let\\u2019s read the manual.\"\n", "Let’s read the manual."},
		{"folded strip", "message: >-\n  It is important\n  to note that.\n", "It is important to note that."},
		{"folded blank", "message: >\n  First line.\n\n  Last line.\n", "First line.\nLast line.\n"},
		{"literal keep", "message: |+\n  First line.\n  Last line.\n\n", "First line.\nLast line.\n\n"},
		{"indent indicator", "message: |2-\n    First line.\n  Last line.\n", "  First line.\nLast line."},
		{"quoted folding", "message: \"First line.\n  Last line.\"\n", "First line. Last line."},
		{"escaped continuation", "message: \"First\\\n  line.\"\n", "Firstline."},
		{"plain folding", "message: First line.\n  Last line.\n", "First line. Last line."},
		{"CRLF", "message: >-\r\n  It is important\r\n  to note that.\r\n", "It is important to note that."},
		{"Unicode key", "café: Read the manual.\n", "Read the manual."},
		{"BOM", "\ufeffmessage: Read the manual.\n", "Read the manual."},
		{"tagged", "message: !!str 123\n", "123"},
		{"hash inside string", "message: \"Read #the manual.\"\n", "Read #the manual."},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := qt.New(t)
			doc, err := extract.Parse(t.Context(), document.Source{Name: "sample.yaml", Format: document.YAML, Bytes: []byte(tc.source)},
				extract.Options{})
			c.Assert(err, qt.IsNil)
			c.Assert(doc.Blocks, qt.HasLen, 1)
			c.Assert(doc.Blocks[0].Kind, qt.Equals, "string")
			c.Assert(doc.Blocks[0].Text, qt.Equals, tc.text)
			assertSourceMap(c, doc, tc.source)
		})
	}
}

func TestYAMLTypesAliasesAndStreams(t *testing.T) {
	c := qt.New(t)
	source := "# Read the manual.\nvalues: [true, false, null, 12, 1.5, \"123\"]\n" +
		"anchor: &text A useful value.\nalias: *text\nbinary: !!binary YQ==\n---\nmessage: Other prose.\n"
	doc, err := extract.Parse(t.Context(), document.Source{Name: "sample.yml", Format: document.YAML, Bytes: []byte(source)},
		extract.Options{})
	c.Assert(err, qt.IsNil)
	c.Assert(doc.Blocks, qt.HasLen, 4)
	c.Assert(strings.TrimSpace(doc.Blocks[0].Text), qt.Equals, "Read the manual.")
	c.Assert(doc.Blocks[1].Text, qt.Equals, "123")
	c.Assert(doc.Blocks[2].Text, qt.Equals, "A useful value.")
	c.Assert(doc.Blocks[3].Text, qt.Equals, "Other prose.")
	var reasons []string
	for _, excluded := range doc.Excluded {
		reasons = append(reasons, excluded.Reason)
	}
	c.Assert(reasons, qt.Contains, "yaml-alias")
	c.Assert(reasons, qt.Contains, "yaml-non-string:!!binary")
	assertSourceMap(c, doc, source)
}

func TestNewLanguageExceptionSymbols(t *testing.T) {
	cases := []struct {
		name, source string
		format       document.Format
	}{
		{
			"sample.cs",
			"// Check the runtime message.\nclass Sample { const string fixture = \"Fixture prose.\"; string message = \"Runtime prose.\"; }",
			document.CSharp,
		},
		{"sample.yaml", "# Check the runtime message.\nfixture: Fixture prose.\nmessage: Runtime prose.\n", document.YAML},
		{"sample.yml", "# Check the runtime message.\nvalues: {fixture: Fixture prose., message: Runtime prose.}\n", document.YAML},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := qt.New(t)
			policy := extract.Policy{Exceptions: []extract.Exception{{
				ID: "fixture", Paths: []string{"sample.*"}, Formats: []document.Format{tc.format}, Kinds: []string{"string"},
				Symbols: []string{"fixture"}, Reason: "Deliberate fixture prose.",
			}}}
			doc, err := extract.Parse(t.Context(), document.Source{Name: tc.name, Format: tc.format, Bytes: []byte(tc.source)},
				extract.Options{Policy: policy})
			c.Assert(err, qt.IsNil)
			c.Assert(doc.Blocks, qt.HasLen, 2)
			c.Assert(doc.Blocks[0].Kind, qt.Equals, "comment")
			c.Assert(doc.Blocks[1].Text, qt.Equals, "Runtime prose.")
		})
	}
}
