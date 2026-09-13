package extract_test

import (
	"context"
	"strconv"
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/extract"
)

const structuredGoComment = `package example
// Options applies these limits:
//   - Retry: The client may retry only when enabled
//   - Timeout: The café closes after 30 seconds
//
// The server must not discard the configured limit.
//
//	limit := "It is important to note that this is code."
//
// The error message remains eligible.
func Options() {}
`

func sourceProse(t *testing.T, name string, format document.Format, text string, policy extract.Policy) document.Document {
	t.Helper()
	doc, err := extract.Parse(t.Context(), document.Source{Name: name, Format: format, Bytes: []byte(text)},
		extract.Options{Policy: policy, IncludeStructure: true})
	c := qt.New(t)
	c.Assert(err, qt.IsNil)
	return doc
}

func proseTexts(doc document.Document) []string {
	texts := make([]string, len(doc.Blocks))
	for i, block := range doc.Blocks {
		texts[i] = strings.TrimSpace(block.Text)
	}
	return texts
}

func TestGoDocBlocksPreserveOriginalBytes(t *testing.T) {
	c := qt.New(t)
	text := "\ufeff" + strings.ReplaceAll(structuredGoComment, "\n", "\r\n")
	doc := sourceProse(t, "options.go", document.Go, text, extract.Policy{})
	c.Assert(string(doc.Source), qt.Equals, text)
	c.Assert(proseTexts(doc), qt.DeepEquals, []string{
		"Options applies these limits:", "Retry: The client may retry only when enabled",
		"Timeout: The café closes after 30 seconds", "The server must not discard the configured limit.",
		"The error message remains eligible.",
	})
	c.Assert(doc.Blocks[1].Context, qt.Contains, "embedded:godoc:list-item")
	c.Assert(doc.Blocks[1].List.Items, qt.Equals, 2)
	c.Assert(doc.Blocks[2].List.Item, qt.Equals, 1)
	for _, block := range doc.Blocks {
		c.Assert(block.Kind, qt.Equals, "comment")
		for _, span := range block.Map {
			c.Assert(span.Valid(len(text)), qt.IsTrue)
		}
	}
	start := strings.Index(doc.Blocks[2].Text, "café")
	spans := doc.Blocks[2].Spans(start, start+len("café"))
	c.Assert(spans, qt.HasLen, 1)
	c.Assert(text[spans[0].Start:spans[0].End], qt.Equals, "café")
	c.Assert(strings.Join(excludedTexts(doc, text, "comment-code"), ""), qt.Contains, "limit :=")
}

func TestGoDocGrammarAndPlainOverride(t *testing.T) {
	for _, row := range []struct {
		name, input string
		want        []string
	}{
		{"block comment", "/*\n * Limits:\n *   - Keep the first condition\n *   - Keep the second condition\n */",
			[]string{"Limits:", "Keep the first condition", "Keep the second condition"}},
		{"wrapped item", "// Limits:\n//   1. Keep this condition\n//      and its continuation\n//\n" +
			"//      Keep its second paragraph.\n//   2. Keep the other condition",
			[]string{"Limits:", "Keep this condition and its continuation", "Keep its second paragraph.", "Keep the other condition"}},
		{"misindented code", "// Before.\n// func main() {\n//\trun()\n// }\n// After.", []string{"Before.", "After."}},
		{"Unicode list whitespace", "// Limits:\n//   \u2022 \u00a0Keep the café open.\u00a0", []string{"Limits:", "Keep the café open."}},
		{"links and heading", "// Before.\n//\n// # Limits\n//\n// Read [manual].\n//\n" +
			"// [manual]: https://example.com\n// [other]: https://example.org",
			[]string{"Before.", "Limits", "Read [manual]."}},
	} {
		t.Run(row.name, func(t *testing.T) {
			doc := sourceProse(t, "doc.go", document.Go, "package example\n"+row.input+"\nfunc Example() {}\n", extract.Policy{})
			c := qt.New(t)
			c.Assert(proseTexts(doc), qt.DeepEquals, row.want)
			for _, block := range doc.Blocks {
				for offset, char := range block.Text {
					if char == ' ' || char == 0 {
						continue
					}
					span := block.Map[offset]
					c.Assert(string(doc.Source[span.Start:span.End]), qt.Equals, string(char))
				}
			}
		})
	}
	doc := sourceProse(t, "options.go", document.Go, structuredGoComment, extract.Policy{GoComments: "plain"})
	c := qt.New(t)
	c.Assert(doc.Blocks[0].Text, qt.Contains, "- Retry:")
	c.Assert(doc.Blocks[0].Text, qt.Contains, "- Timeout:")
}

func markdownPolicy() extract.Policy {
	return extract.Policy{MarkdownStrings: []extract.MarkdownString{{ID: "help", Paths: []string{"**"}, Symbols: []string{"Long"}}}}
}

func TestSelectedMarkdownStringsUseExistingGrammar(t *testing.T) {
	text := "The server must not retry.\n\n- Keep café and 30 seconds.\n  - Keep `tenant_id` unchanged.\n\n" +
		"```yaml\nlimit: It is important to note that.\n```\n\nThe client may retry only when enabled."
	for _, row := range []struct {
		name   string
		format document.Format
		source string
	}{
		{"help.go", document.Go, "package example\nvar Long = " + strconv.Quote(text) + "\n"},
		{"help.py", document.Python, "Long = \"\"\"" + text + "\"\"\"\n"},
		{"help.yaml", document.YAML, "Long: |\n  " + strings.ReplaceAll(text, "\n", "\n  ") + "\n"},
	} {
		t.Run(row.name, func(t *testing.T) {
			c := qt.New(t)
			doc := sourceProse(t, row.name, row.format, row.source, markdownPolicy())
			c.Assert(doc.Blocks, qt.HasLen, 4)
			c.Assert(doc.Blocks[0].Text, qt.Equals, "The server must not retry.")
			c.Assert(doc.Blocks[1].Text, qt.Contains, "café and 30 seconds")
			c.Assert(doc.Blocks[2].Text, qt.Not(qt.Contains), "tenant_id")
			c.Assert(doc.Blocks[2].List.Depth, qt.Equals, 2)
			c.Assert(doc.Blocks[3].Text, qt.Equals, "The client may retry only when enabled.")
			for _, block := range doc.Blocks {
				c.Assert(block.Kind, qt.Equals, "string")
				for _, span := range block.Map {
					c.Assert(span.Valid(len(row.source)), qt.IsTrue)
				}
			}
		})
	}
}

func TestMarkdownSelectionAndPolicyPrecedence(t *testing.T) {
	c := qt.New(t)
	text := "package example\nvar Long = `First.\n\nSecond.`\nvar Short = `Third.\n\nFourth.`\n"
	doc := sourceProse(t, "help.go", document.Go, text, markdownPolicy())
	c.Assert(proseTexts(doc), qt.DeepEquals, []string{"First.", "Second.", "Third.\n\nFourth."})
	policy := markdownPolicy()
	policy.MarkdownStrings[0].Formats = []document.Format{document.YAML}
	c.Assert(sourceProse(t, "help.go", document.Go, text, policy).Blocks, qt.HasLen, 2)
	policy = markdownPolicy()
	policy.MarkdownStrings[0].Paths = []string{"cli/**"}
	c.Assert(sourceProse(t, "help.go", document.Go, text, policy).Blocks, qt.HasLen, 2)
	policy = markdownPolicy()
	policy.Languages = map[document.Format]extract.LanguagePolicy{document.Go: {Contexts: []string{"comment"}}}
	c.Assert(sourceProse(t, "help.go", document.Go, text, policy).Blocks, qt.HasLen, 0)
	policy = markdownPolicy()
	policy.Exceptions = []extract.Exception{{ID: "external", Paths: []string{"**"}, Kinds: []string{"string"}, Reason: "External contract."}}
	c.Assert(sourceProse(t, "help.go", document.Go, text, policy).Blocks, qt.HasLen, 0)
}

func TestExplicitMarkdownFailuresAndLimits(t *testing.T) {
	c := qt.New(t)
	source := document.Source{Name: "help.js", Format: document.JavaScript, Bytes: []byte("const Long = `Before ${value} after.`;")}
	_, err := extract.Parse(t.Context(), source, extract.Options{Policy: markdownPolicy()})
	c.Assert(err, qt.ErrorMatches, `.*requires a static string.*`)
	policy := markdownPolicy()
	policy.Contexts = []string{"comment"}
	_, err = extract.Parse(t.Context(), source, extract.Options{Policy: policy})
	c.Assert(err, qt.IsNil)
	source = document.Source{Name: "help.go", Format: document.Go, Bytes: []byte("package example\nvar Long = `One.\n\nTwo.`")}
	_, err = extract.Parse(t.Context(), source, extract.Options{Policy: markdownPolicy(), MaxBlocks: 1})
	c.Assert(err, qt.ErrorMatches, `.*exceeds.*`)
	source.Bytes = []byte(structuredGoComment)
	_, err = extract.Parse(t.Context(), source, extract.Options{MaxBlocks: 1})
	c.Assert(err, qt.ErrorMatches, `.*exceeds.*`)
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	_, err = extract.Parse(ctx, source, extract.Options{})
	c.Assert(err, qt.ErrorIs, context.Canceled)
}

func TestMarkdownStringMapsEscapesAndDistinctLists(t *testing.T) {
	c := qt.New(t)
	text := "package example\nvar Long = \"Caf\\u00e9 &amp; tea.\\n\\n- First item\\n- Second item\"\n" +
		"func Other() { Long := `- Third item\n- Fourth item`; _ = Long }\n"
	doc := sourceProse(t, "help.go", document.Go, text, markdownPolicy())
	c.Assert(doc.Blocks, qt.HasLen, 5)
	first := doc.Blocks[0]
	c.Assert(first.Text, qt.Equals, "Café & tea.")
	for _, row := range []struct{ decoded, original string }{{"é", `\u00e9`}, {"&", "&amp;"}} {
		start := strings.Index(first.Text, row.decoded)
		spans := first.Spans(start, start+len(row.decoded))
		c.Assert(spans, qt.HasLen, 1)
		c.Assert(text[spans[0].Start:spans[0].End], qt.Equals, row.original)
	}
	c.Assert(doc.Blocks[1].List.ID, qt.Not(qt.Equals), doc.Blocks[3].List.ID)
	policy := markdownPolicy()
	policy.Languages = map[document.Format]extract.LanguagePolicy{document.Markdown: {Contexts: []string{"paragraph"}}}
	doc = sourceProse(t, "help.go", document.Go, text, policy)
	c.Assert(doc.Blocks, qt.HasLen, 1)
}

func TestGoDocControlsAndExceptionsPrecedeStructure(t *testing.T) {
	c := qt.New(t)
	text := "package example\n// unswell-disable-next-block filler.announced-importance -- Required contract.\n" +
		"// It is important to note that the limit is 30 seconds.\nfunc Retry() {}\n"
	policy := extract.Policy{Exceptions: []extract.Exception{{ID: "external", Paths: []string{"**"},
		Kinds: []string{"comment"}, Reason: "External contract."}}}
	doc := sourceProse(t, "help.go", document.Go, text, policy)
	c.Assert(doc.Blocks, qt.HasLen, 0)
	c.Assert(doc.Directives, qt.HasLen, 1)
	c.Assert(doc.Excluded, qt.Not(qt.HasLen), 0)
	doc = sourceProse(t, "help.go", document.Go, text, extract.Policy{})
	c.Assert(doc.Blocks, qt.HasLen, 1)
	c.Assert(doc.Directives, qt.HasLen, 1)
	c.Assert(doc.Blocks[0].Text, qt.Contains, "limit is 30 seconds")
	text = "package example\n// Before.\n//\n//   unswell-disable-next-block filler.announced-importance -- Example.\n" +
		"//\n// It is important to note that the limit is 30 seconds.\nfunc Retry() {}\n"
	doc = sourceProse(t, "help.go", document.Go, text, extract.Policy{})
	c.Assert(doc.Directives, qt.HasLen, 0)
	c.Assert(doc.Blocks, qt.HasLen, 2)
	c.Assert(excludedTexts(doc, text, "comment-code"), qt.HasLen, 1)
}
