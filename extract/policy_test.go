package extract_test

import (
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/extract"
)

func TestSourceExceptionSelectors(t *testing.T) {
	source := "// Check the runtime message.\npackage sample\nconst fixture = \"Fixture prose.\"\nconst message = \"Runtime prose.\"\n"
	cases := []struct {
		name    string
		paths   []string
		formats []document.Format
		kinds   []string
		symbols []string
		texts   []string
	}{
		{"symbol", []string{"src/*.go"}, []document.Format{document.Go}, []string{"string"}, []string{"fixture"},
			[]string{"Check the runtime message.", "Runtime prose."}},
		{"other path", []string{"testdata/*.go"}, nil, []string{"string"}, []string{"fixture"},
			[]string{"Check the runtime message.", "Fixture prose.", "Runtime prose."}},
		{"other format", []string{"src/*.go"}, []document.Format{document.JavaScript}, []string{"string"}, nil,
			[]string{"Check the runtime message.", "Fixture prose.", "Runtime prose."}},
		{"other symbol", []string{"src/*.go"}, nil, []string{"string"}, []string{"absent"},
			[]string{"Check the runtime message.", "Fixture prose.", "Runtime prose."}},
		{"comments only", []string{"src/*.go"}, nil, []string{"comment"}, nil,
			[]string{"Fixture prose.", "Runtime prose."}},
		{"strings only", []string{"src/*.go"}, nil, []string{"string"}, nil,
			[]string{"Check the runtime message."}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := qt.New(t)
			policy := extract.Policy{Exceptions: []extract.Exception{{
				ID: "fixture", Paths: tc.paths, Formats: tc.formats, Kinds: tc.kinds, Symbols: tc.symbols, Reason: "Deliberate fixture prose.",
			}}}
			doc, err := extract.Parse(t.Context(), document.Source{Name: "src/sample.go", Format: document.Go, Bytes: []byte(source)},
				extract.Options{Policy: policy})
			c.Assert(err, qt.IsNil)
			var texts []string
			for _, block := range doc.Blocks {
				texts = append(texts, strings.TrimSpace(block.Text))
			}
			c.Assert(texts, qt.DeepEquals, tc.texts)
			for _, exclusion := range doc.Excluded {
				c.Assert(exclusion.Reason, qt.Equals, "config:fixture: Deliberate fixture prose.")
				c.Assert(exclusion.Span.Valid(len(source)), qt.IsTrue)
			}
		})
	}
}

func TestShellExceptionDoesNotHideOtherStrings(t *testing.T) {
	c := qt.New(t)
	source := "# Check the runtime message.\nfixture='Deliberate fixture.'\nmessage='Runtime prose.'\n"
	policy := extract.Policy{Exceptions: []extract.Exception{{
		ID: "fixture", Paths: []string{"*.sh"}, Kinds: []string{"string"}, Symbols: []string{"fixture"}, Reason: "Deliberate fixture prose.",
	}}}
	doc, err := extract.Parse(t.Context(), document.Source{Name: "sample.sh", Format: document.Shell, Bytes: []byte(source)},
		extract.Options{Policy: policy})
	c.Assert(err, qt.IsNil)
	c.Assert(doc.Blocks, qt.HasLen, 2)
	c.Assert(doc.Blocks[0].Kind, qt.Equals, "comment")
	c.Assert(doc.Blocks[1].Text, qt.Equals, "Runtime prose.")
	c.Assert(doc.Excluded, qt.HasLen, 1)
}

func TestExceptionDoesNotAcceptInvalidSyntax(t *testing.T) {
	c := qt.New(t)
	policy := extract.Policy{Exceptions: []extract.Exception{{
		ID: "fixture", Paths: []string{"*.py"}, Kinds: []string{"string"}, Reason: "Deliberate fixture prose.",
	}}}
	_, err := extract.Parse(t.Context(), document.Source{Name: "bad.py", Format: document.Python, Bytes: []byte("value = 'unfinished")},
		extract.Options{Policy: policy})
	c.Assert(err, qt.IsNotNil)
}
